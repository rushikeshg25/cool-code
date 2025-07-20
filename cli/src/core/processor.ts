import { StreamingSpinner } from '../ui/spinner';
import { LLM } from './llm';
import { ContextManager } from './contextManager';
import dotenv from 'dotenv';
import { createGitIgnoreChecker } from './tools/ignoreGitIgnoreFileTool';
import { validateAndRunToolCall } from './tools/toolValidator';
export interface QueryResult {
  query: string;
  response: string;
  suggestions?: string[];
  timestamp: Date;
}

export interface LLMConfig {
  model: string;
  temperature?: number;
  maxTokens?: number;
}

export interface configType {
  LLMConfig: LLMConfig;
  rootDir: string;
  doesExistInGitIgnore: (rootDir: string) => boolean | null;
}

export type SpinnerUpdateCallback = (text: string) => void;

export class Processor {
  public config: configType;
  private LLM: LLM;
  private contextManager: ContextManager;

  constructor(rootDir: string) {
    dotenv.config();
    this.config = {
      LLMConfig: {
        model: 'gemini-2.5-flash',
      },
      rootDir,
      doesExistInGitIgnore: createGitIgnoreChecker(rootDir),
    };
    this.LLM = new LLM(this.config.LLMConfig.model);
    this.contextManager = new ContextManager(
      rootDir,
      this.config.doesExistInGitIgnore
    );
  }

  isFinalMessage(response: any): boolean {
    return (
      !Array.isArray(response) &&
      typeof response === 'object' &&
      'text' in response
    );
  }

  async processQuery(query: string) {
    this.contextManager.addUserMessage(query);

    const streamingSpinner = new StreamingSpinner();
    streamingSpinner.start('🔄 Generating response...');

    while (true) {
      const prompt = this.contextManager.buildPrompt();
      console.log(prompt);
      const response = await this.LLM.StreamResponse(prompt);
      // if (toolCall && typeof toolCall === 'object' && 'text' in toolCall) {
      //   streamingSpinner.succeed('Response completed!');
      //   console.log('\n[FINAL MESSAGE]\n' + toolCall.text + '\n');
      //   return;
      // }

      let toolCalls;
      try {
        let cleanResponse = response.trim();
        if (cleanResponse.startsWith('```')) {
          cleanResponse = cleanResponse
            .replace(/^```[a-zA-Z]*\n?/, '')
            .replace(/```$/, '')
            .trim();
        }
        toolCalls = JSON.parse(cleanResponse);
        if (this.isFinalMessage(toolCalls)) {
          streamingSpinner.succeed('Response completed!');
          console.log('\n[FINAL MESSAGE]\n' + response + '\n');
          break;
        }
        console.log('{TOOLCALLS}', toolCalls);
        if (typeof toolCalls === 'string') {
          toolCalls = JSON.parse(toolCalls);
        }
      } catch {
        streamingSpinner.succeed('Response completed!');
        console.log('\n[FINAL MESSAGE]\n' + response + '\n');
        break;
      }

      for (const toolCall of toolCalls) {
        console.log('[DEBUG] toolCall:', toolCall, typeof toolCall);
        if (toolCall && typeof toolCall === 'object' && 'tool' in toolCall) {
          console.log('\n[TOOL CALL]', JSON.stringify(toolCall, null, 2));
          try {
            const result = await validateAndRunToolCall(
              toolCall,
              this.config,
              this.config.rootDir
            );
            this.contextManager.addResponse(
              result.result?.LLMresult as string,
              toolCall
            );
            if (result.result?.LLMresult) {
              streamingSpinner.updateText(result.result.LLMresult);
            } else if (result.error) {
              streamingSpinner.updateText(result.error);
            }
            if (result.success) {
              console.log(
                '[TOOL RESULT]',
                result.result?.DisplayResult ||
                  result.result?.LLMresult ||
                  'No display result.'
              );
            } else {
              console.error('[TOOL ERROR]', result.error);
            }
          } catch (err) {
            streamingSpinner.updateText(
              '[AGENT ERROR] ' +
                (err instanceof Error ? err.message : String(err))
            );
            console.error('[AGENT ERROR]', err);
          }
        } else {
          console.log(
            '[DEBUG] Unhandled toolCall structure:',
            toolCall,
            typeof toolCall
          );
        }
      }
    }
  }
}
