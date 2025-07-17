export interface CLIOptions {
  query: string;
  verbose?: boolean;
  dryRun?: boolean;
  apiKey?: string;
}

export interface ProgressDisplay {
  showThinking(message: string): void;
  showFileOperation(file: string, operation: string): void;
  showError(error: string): void;
  showSuccess(message: string): void;
}

export interface ParsedQuery {
  intent: DatabaseIntent;
  entities: string[];
  operations: DatabaseOperation[];
  context: QueryContext;
}

export enum DatabaseIntent {
  CREATE_TABLE = 'create_table',
  STORE_DATA = 'store_data',
  CREATE_API = 'create_api',
  INTEGRATE_FRONTEND = 'integrate_frontend'
}

export enum DatabaseOperation {
  CREATE = 'create',
  READ = 'read',
  UPDATE = 'update',
  DELETE = 'delete'
}

export interface QueryContext {
  entities: string[];
  relationships: string[];
  requirements: string[];
}

export interface ProjectContext {
  structure: ProjectStructure;
  existingComponents: ComponentInfo[];
  databaseConfig: DatabaseConfig | null;
  dependencies: PackageInfo[];
}

export interface ProjectStructure {
  srcDir: string;
  apiDir: string;
  componentsDir: string;
  libDir: string;
  hasDatabase: boolean;
  rootDir: string;
}

export interface ComponentInfo {
  filePath: string;
  componentName: string;
  mockDataUsage: MockDataReference[];
  propsInterface?: TypeDefinition;
}

export interface MockDataReference {
  variableName: string;
  dataType: string;
  usage: string;
}

export interface TypeDefinition {
  name: string;
  properties: Record<string, string>;
}

export interface DatabaseConfig {
  provider: string;
  connectionString?: string;
  schemaPath?: string;
}

export interface PackageInfo {
  name: string;
  version: string;
  isDev: boolean;
}

export interface AgentResult {
  success: boolean;
  filesModified: string[];
  filesCreated: string[];
  errors: string[];
  summary: string;
}

export interface DrizzleSchema {
  tableName: string;
  fields: DrizzleField[];
  relationships: Relationship[];
  indexes: Index[];
}

export interface DrizzleField {
  name: string;
  type: DrizzleFieldType;
  constraints: FieldConstraint[];
}

export enum DrizzleFieldType {
  TEXT = 'text',
  INTEGER = 'integer',
  REAL = 'real',
  BLOB = 'blob',
  TIMESTAMP = 'timestamp',
  BOOLEAN = 'boolean'
}

export interface FieldConstraint {
  type: 'primary_key' | 'not_null' | 'unique' | 'foreign_key' | 'default';
  value?: any;
  references?: {
    table: string;
    column: string;
  };
}

export interface Relationship {
  type: 'one_to_one' | 'one_to_many' | 'many_to_many';
  targetTable: string;
  foreignKey: string;
  targetKey?: string;
}

export interface Index {
  name: string;
  columns: string[];
  unique: boolean;
}

export interface APIRoute {
  path: string;
  method: HTTPMethod;
  handler: string;
  validation: ValidationSchema;
  responseType: TypeDefinition;
}

export enum HTTPMethod {
  GET = 'GET',
  POST = 'POST',
  PUT = 'PUT',
  DELETE = 'DELETE',
  PATCH = 'PATCH'
}

export interface ValidationSchema {
  body?: Record<string, any>;
  query?: Record<string, any>;
  params?: Record<string, any>;
}

export interface ComponentUpdate {
  filePath: string;
  changes: CodeChange[];
}

export interface CodeChange {
  type: 'replace' | 'insert' | 'delete';
  startLine: number;
  endLine?: number;
  content: string;
}