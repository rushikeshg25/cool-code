import { z } from 'zod';

export const ReadFileSchema = z.object({
  tool: z.literal('read_file'),
  parameters: z.object({
    absolute_path: z.string().min(1, 'File path cannot be empty'),
    startLine: z.number().int().positive().optional(),
    endLine: z.number().int().positive().optional(),
  }),
});

export const EditFileSchema = z.object({
  tool: z.literal('edit_file'),
  parameters: z.object({
    absolute_path: z.string().min(1, 'File path cannot be empty'),
    content: z.string(),
    backup: z.boolean().optional(),
  }),
});

export const ShellCommandSchema = z.object({
  tool: z.literal('shell_command'),
  parameters: z.object({
    command: z.string().min(1, 'Command cannot be empty'),
    cwd: z.string().optional(),
    timeout: z.number().positive().optional(),
  }),
});

export const GlobSchema = z.object({
  tool: z.literal('glob'),
  parameters: z.object({
    pattern: z.string().min(1, 'Pattern cannot be empty'),
    rootDir: z.string().optional(),
    recursive: z.boolean().optional(),
    includeDirectories: z.boolean().optional(),
  }),
});

export const GrepSchema = z.object({
  tool: z.literal('grep'),
  parameters: z.object({
    pattern: z.string().min(1, 'Pattern cannot be empty'),
    path: z.string().optional(),
    include: z.string().optional(),
  }),
});

export type ToolCall =
  | z.infer<typeof ReadFileSchema>
  | z.infer<typeof EditFileSchema>
  | z.infer<typeof ShellCommandSchema>
  | z.infer<typeof GlobSchema>;

export interface ValidationResult {
  success: boolean;
  error?: string;
  data?: ToolCall;
}

export interface FileValidationResult {
  exists: boolean;
  isFile: boolean;
  isReadable: boolean;
  error?: string;
}

/**
 * Validate JSON tool call structure
 */
export function validateToolCall(jsonData: unknown): ValidationResult {
  try {
    if (!jsonData || typeof jsonData !== 'object' || !('tool' in jsonData)) {
      return {
        success: false,
        error: 'Invalid tool call format. Expected object with "tool" property',
      };
    }

    const data = jsonData as any;

    switch (data.tool) {
      case 'read_file':
        const readFileResult = ReadFileSchema.safeParse(data);
        if (!readFileResult.success) {
          return {
            success: false,
            error: `Invalid read_file parameters: ${readFileResult.error.message}`,
          };
        }
        return { success: true, data: readFileResult.data };

      case 'edit_file':
        const editFileResult = EditFileSchema.safeParse(data);
        if (!editFileResult.success) {
          return {
            success: false,
            error: `Invalid edit_file parameters: ${editFileResult.error.message}`,
          };
        }
        return { success: true, data: editFileResult.data };

      case 'shell_command':
        const shellResult = ShellCommandSchema.safeParse(data);
        if (!shellResult.success) {
          return {
            success: false,
            error: `Invalid shell_command parameters: ${shellResult.error.message}`,
          };
        }
        return { success: true, data: shellResult.data };

      case 'glob':
        const globResult = GlobSchema.safeParse(data);
        if (!globResult.success) {
          return {
            success: false,
            error: `Invalid glob parameters: ${globResult.error.message}`,
          };
        }
        return { success: true, data: globResult.data };

      default:
        return {
          success: false,
          error: `Unknown tool: ${data.tool}. Supported tools: read_file, edit_file, shell_command, glob`,
        };
    }
  } catch (error) {
    return {
      success: false,
      error: `Failed to parse tool call: ${error instanceof Error ? error.message : String(error)}`,
    };
  }
}

/**
 * Parse JSON tool call from string
 */
export function parseToolCall(jsonString: string): ValidationResult {
  try {
    const parsed = JSON.parse(jsonString);
    return validateToolCall(parsed);
  } catch (error) {
    return {
      success: false,
      error: `Invalid JSON: ${error instanceof Error ? error.message : String(error)}`,
    };
  }
}
