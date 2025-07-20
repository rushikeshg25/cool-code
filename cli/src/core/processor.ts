import chalk from "chalk";
import { StreamingSpinner } from "../ui/spinner";
import { createGitIgnoreChecker } from "./tools";
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

  async processQuery(query: string) {
    try {
      // Display query and response header
      console.log(chalk.blue("\n📝 Query:"), query);
      console.log(chalk.green("✨ Response:"));
      console.log();

      // Create streaming spinner that stays at bottom
      const streamingSpinner = new StreamingSpinner();
      streamingSpinner.start("🔄 Generating response...");

      const prompt =
        (await this.contextManager.buildPrompt()) +
        "User's current prompt: " +
        query;
      const response = await this.LLM.StreamResponse(
        prompt,
        (chunk: string) => {
          // Clear the spinner line before writing content
          process.stdout.write("\r\x1b[K");

          // Write the chunk
          process.stdout.write(chunk);

          // If chunk doesn't end with newline, add one for spinner
          if (!chunk.endsWith("\n")) {
            process.stdout.write("\n");
          }
        }
      );

      streamingSpinner.succeed("Response completed!");

      console.log(
        chalk.gray(`Completed at: ${new Date().toLocaleTimeString()}`)
      );

      return response;
    } catch (error) {
      throw new Error("Error processing query: " + error);
    }
  }
}
