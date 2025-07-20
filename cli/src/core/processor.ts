import chalk from "chalk";
import { StreamingSpinner } from "../ui/spinner";
import { createGitIgnoreChecker, validateAndRunToolCall } from "./tools";
import { LLM } from "./llm";
import { ContextManager } from "./contextManager";
import dotenv from "dotenv";
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
        model: "gemini-2.5-flash",
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
      typeof response === "object" &&
      "text" in response
    );
  }

  async processQuery(query: string) {
    // 1. Add user message to context
    this.contextManager.addUserMessage(query);

    const streamingSpinner = new StreamingSpinner();
    streamingSpinner.start("🔄 Generating response...");
    while (true) {
      // 2. Build prompt and get LLM response
      const prompt = this.contextManager.buildPrompt();
      const response = await this.LLM.StreamResponse(prompt);

      // 3. Try to parse as JSON array of tool calls
      let toolCalls;
      try {
        toolCalls = JSON.parse(response);
      } catch {
        // If not JSON, treat as final message
        streamingSpinner.succeed("Response completed!");
        console.log(response);
        break;
      }

      // 4. If tool calls, execute them
      for (const toolCall of toolCalls) {
        if (toolCall.tool) {
          const result = await validateAndRunToolCall(
            toolCall,
            this.config,
            this.config.rootDir
          );
          // Add tool result to context
          this.contextManager.addToolResult(toolCall, result);
          // Update spinner text to LLMresult if available
          if (result.result?.LLMresult) {
            streamingSpinner.updateText(result.result.LLMresult);
          } else if (result.error) {
            streamingSpinner.updateText(result.error);
          }
          // Print/display result
          console.log(result.result?.DisplayResult || result.error);
        } else if (toolCall.text) {
          // Final message
          streamingSpinner.succeed("Response completed!");
          console.log(toolCall.text);
          return;
        }
      }
    }
  }
}
