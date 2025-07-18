import path from "path";
import { DynamicSpinner } from "../ui/spinner";
import { createGitIgnoreChecker } from "./tools";
import { LLM } from "./llm";

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

interface configType {
  LLMConfig: LLMConfig;
  rootDir: string;
  doesExistInGitIgnore: (rootDir: string) => boolean | null;
}

export type SpinnerUpdateCallback = (text: string) => void;

export class Processor {
  private config: configType;
  private currentDir: string | null;
  private LLM: LLM;

  constructor(rootDir: string) {
    this.config = {
      LLMConfig: {
        model: "gemini-2.5-flash",
      },
      rootDir,
      doesExistInGitIgnore: createGitIgnoreChecker(rootDir),
    };
    this.currentDir = rootDir;
    this.LLM = new LLM(this.config.LLMConfig.model);
  }

  async processQuery(query: string, spinner: DynamicSpinner) {
    spinner.updateText("Processing query...");
    console.log("Processing query: ", query);
    try {
      await this.LLM.StreamResponse(query);
    } catch (error) {
      throw Error("Error processing query: " + error);
    }
  }

  private async getEnvironment() {}
}

// Export tools
export * from "./tools";
