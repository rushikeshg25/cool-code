import path from "path";
import { DynamicSpinner } from "../ui/spinner";
import { createGitIgnoreChecker } from "./tools";

export interface QueryResult {
  query: string;
  response: string;
  suggestions?: string[];
  timestamp: Date;
}

export interface AIConfig {
  model?: string;
  temperature?: number;
  maxTokens?: number;
}

interface configType {
  rootDir: string;
  doesExistInGitIgnore: (rootDir: string) => boolean | null;
}

export type SpinnerUpdateCallback = (text: string) => void;

export class Processor {
  private config: configType;
  private currentDir: string | null;
  
  constructor(rootDir: string) {
    this.config = {
      rootDir,
      doesExistInGitIgnore: createGitIgnoreChecker(rootDir),
    };
    this.currentDir = rootDir;
  }

   async processQuery(query: string, spinner:DynamicSpinner)  {
    spinner.updateText("Processing query...");
    // Simulate processing time
    await new Promise((resolve) => setTimeout(resolve, 1000));
  }

  private async getEnvironment(){
    
  }
}

// Export tools
export * from './tools';
