import { z } from "zod";
import { editFile, readFile, execCommand, globFiles } from "./index";
import type { configType } from "../processor";
import type { ToolResult } from "../../types";

export const ReadFileSchema = z.object({
  tool: z.literal("read_file"),
  parameters: z.object({
    absolutePath: z.string().min(1, "File path cannot be empty"),
    startLine: z.number().int().positive().optional(),
    endLine: z.number().int().positive().optional(),
  }),
});

export const EditFileSchema = z.object({
  tool: z.literal("edit_file"),
  toolOptions: z.object({
    filePath: z.string().min(1, "File path cannot be empty"),
    oldString: z.string(),
    newString: z.string(),
    expected_replacements: z.number().optional(),
  }),
});

export const ShellCommandSchema = z.object({
  tool: z.literal("shell_command"),
  toolOptions: z.object({
    command: z.string().min(1, "Command cannot be empty"),
    description: z.string().optional(),
    directory: z.string().optional(),
  }),
});

export const GlobSchema = z.object({
  tool: z.literal("glob"),
  toolOptions: z.object({
    pattern: z.string().min(1, "Pattern cannot be empty"),
  }),
});

export const GrepSchema = z.object({
  tool: z.literal("grep"),
  toolOptions: z.object({
    pattern: z.string().min(1, "Pattern cannot be empty"),
    path: z.string().optional(),
    include: z.string().optional(),
  }),
});

export type ToolCall =
  | z.infer<typeof ReadFileSchema>
  | z.infer<typeof EditFileSchema>
  | z.infer<typeof ShellCommandSchema>
  | z.infer<typeof GlobSchema>;

export interface FileValidationResult {
  exists: boolean;
  isFile: boolean;
  isReadable: boolean;
  error?: string;
}

export async function validateAndRunToolCall(
  jsonData: unknown,
  config: configType,
  rootPath: string
): Promise<{
  success: boolean;
  data?: any;
  error?: string;
  result?: ToolResult;
}> {
  try {
    if (!jsonData || typeof jsonData !== "object" || !("tool" in jsonData)) {
      return {
        success: false,
        error: 'Invalid tool call format. Expected object with "tool" property',
      };
    }

    const data = jsonData as any;

    switch (data.tool) {
      case "read_file": {
        const readFileResult = ReadFileSchema.safeParse(data);
        if (!readFileResult.success) {
          return {
            success: false,
            error: `Invalid read_file toolOptions: ${readFileResult.error.message}`,
          };
        }
        // Call the tool
        const result = readFile(readFileResult.data.parameters, rootPath);
        return { success: true, data: readFileResult.data, result };
      }
      case "edit_file": {
        const editFileResult = EditFileSchema.safeParse(data);
        if (!editFileResult.success) {
          return {
            success: false,
            error: `Invalid edit_file toolOptions: ${editFileResult.error.message}`,
          };
        }
        // Ensure expected_replacements is always a number
        const toolOptions = {
          ...editFileResult.data.toolOptions,
          expected_replacements:
            editFileResult.data.toolOptions.expected_replacements ?? 1,
        };
        const result = editFile(toolOptions);
        return { success: true, data: editFileResult.data, result };
      }
      case "shell_command": {
        const shellResult = ShellCommandSchema.safeParse(data);
        if (!shellResult.success) {
          return {
            success: false,
            error: `Invalid shell_command toolOptions: ${shellResult.error.message}`,
          };
        }
        // Call the tool (async)
        const shellExecResult = await execCommand(shellResult.data.toolOptions);
        // Wrap ShellResult into ToolResult
        const result: ToolResult = {
          LLMresult:
            shellExecResult.stdout +
            (shellExecResult.stderr
              ? `\nSTDERR:\n${shellExecResult.stderr}`
              : ""),
          DisplayResult: shellExecResult.success
            ? `Command executed successfully${shellExecResult.exitCode !== null ? ` (exit code: ${shellExecResult.exitCode})` : ""}`
            : `Command failed${shellExecResult.exitCode !== null ? ` (exit code: ${shellExecResult.exitCode})` : ""}${shellExecResult.error ? `: ${shellExecResult.error}` : ""}`,
        };
        return { success: true, data: shellResult.data, result };
      }
      case "glob": {
        const globResult = GlobSchema.safeParse(data);
        if (!globResult.success) {
          return {
            success: false,
            error: `Invalid glob parameters: ${globResult.error.message}`,
          };
        }
        // Call the tool (async)
        const result = await globFiles(globResult.data.toolOptions, config);
        return { success: true, data: globResult.data, result };
      }
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
