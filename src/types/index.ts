export type DatabaseType =
  | 'mysql'
  | 'postgresql'
  | 'sqlite'
  | 'mongodb'
  | 'redis';

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
  output?: 'json' | 'table' | 'raw';
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

export interface LLMConfig {
  model: string;
  temperature?: number;
  maxTokens?: number;
}

export interface TaskItem {
  id: string;
  title: string;
  status: 'todo' | 'in-progress' | 'done' | 'failed';
  detail?: string;
}

export interface TaskList {
  goal: string;
  items: TaskItem[];
}

export type AgentMode = 'planning' | 'agent' | 'ask';
