import { StreamingSpinner } from '../ui/spinner';
import { LLM } from './llm';
import { ContextManager } from './contextManager';
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
    streamingSpinner.start('Thinking out loud...');

    while (true) {
      const prompt = this.contextManager.buildPrompt();
      const response = await this.LLM.StreamResponse(prompt, () => {});
      // console.log('[RAW RESPONSE]', response);
      let toolCalls;
      try {
        let cleanResponse = response.trim();
        if (cleanResponse.startsWith('```')) {
          cleanResponse = cleanResponse
            .replace(/^```[a-zA-Z]*\n?/, '')
            .replace(/```$/, '')
            .trim();
        }
        if (cleanResponse.startsWith('```json')) {
          cleanResponse = cleanResponse
            .replace(/^```json\n?/, '')
            .replace(/```$/, '')
            .trim();
        }
        toolCalls = JSON.parse(cleanResponse);
        if (this.isFinalMessage(toolCalls)) {
          streamingSpinner.succeed(toolCalls.text);
          break;
        }
        if (typeof toolCalls === 'string') {
          toolCalls = JSON.parse(toolCalls);
        }
      } catch (error) {
        console.log(error);
        streamingSpinner.succeed('Response completed!');
        break;
      }

      for (const toolCall of toolCalls) {
        // console.log('[DEBUG] toolCall:', toolCall, typeof toolCall);
        if (toolCall && typeof toolCall === 'object' && 'tool' in toolCall) {
          streamingSpinner.updateText(
            toolCall.description || 'Thinking out loud...'
          );
          try {
            // console.log(toolCall);
            const result = await validateAndRunToolCall(
              toolCall,
              this.config,
              this.config.rootDir
            );
            this.contextManager.addResponse(
              result.result?.LLMresult as string,
              toolCall
            );
          } catch (err) {
            streamingSpinner.updateText(
              '[AGENT ERROR] ' +
                (err instanceof Error ? err.message : String(err))
            );
            console.error('[AGENT ERROR]', err);
          }
        }
      }
    }
  }
}
