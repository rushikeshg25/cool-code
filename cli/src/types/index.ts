export type DatabaseType =
  | "mysql"
  | "postgresql"
  | "sqlite"
  | "mongodb"
  | "redis";

export interface DatabaseConfig {
  type: DatabaseType;
  host?: string;
  port?: number;
  database?: string;
  username?: string;
  password?: string;
}

export interface CLIOptions {
  verbose?: boolean;
  output?: "json" | "table" | "raw";
  config?: string;
}

export interface QueryContext {
  database?: DatabaseConfig;
  options: CLIOptions;
  history: string[];
}

export interface ToolResult {
  LLMresult: string;
  DisplayResult: string;
}

export interface FinalMessageType {
  text: string;
}
