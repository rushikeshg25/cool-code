import { z } from "zod";
import type { configType } from "../processor";
import type { ToolResult } from "../../types";
import { readFile } from "./readFileTool";
import { editFile } from "./editTool";
import { execCommand } from "./shellTool";
import { globFiles } from "./globTool";
import { grepTool } from "./grepTool";
import { newFile } from "./newFileTool";

export const ReadFileSchema = z.object({
  tool: z.literal("read_file"),
  toolOptions: z.object({
    absolutePath: z.string().min(1, "File path cannot be empty"),
    startLine: z.number().int().nonnegative().optional(),
    endLine: z.number().int().nonnegative().optional(),
  }),
});

export const EditFileSchema = z.object({
  tool: z.literal("edit_file"),
  toolOptions: z.object({
    filePath: z.string().min(1, "File path cannot be empty"),
    oldString: z.string(),
    newString: z.string(),
    expected_replacements: z.number().int().positive().optional(),
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

export const NewFileSchema = z.object({
  tool: z.literal("new_file"),
  toolOptions: z.object({
    filePath: z.string().min(1, "File path cannot be empty"),
    content: z.string(),
  }),
});

export type ToolCall =
  | z.infer<typeof ReadFileSchema>
  | z.infer<typeof EditFileSchema>
  | z.infer<typeof ShellCommandSchema>
  | z.infer<typeof GlobSchema>
  | z.infer<typeof GrepSchema>
  | z.infer<typeof NewFileSchema>;

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

        const { startLine, endLine } = readFileResult.data.toolOptions;
        if (
          startLine !== undefined &&
          endLine !== undefined &&
          startLine > endLine
        ) {
          return {
            success: false,
            error: "startLine cannot be greater than endLine",
          };
        }

        const result = readFile(readFileResult.data.toolOptions, rootPath);
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

        const shellExecResult = await execCommand(shellResult.data.toolOptions);

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
            error: `Invalid glob toolOptions: ${globResult.error.message}`,
          };
        }

        const result = await globFiles(globResult.data.toolOptions, config);
        return { success: true, data: globResult.data, result };
      }

      case "grep": {
        const grepResult = GrepSchema.safeParse(data);
        if (!grepResult.success) {
          return {
            success: false,
            error: `Invalid grep toolOptions: ${grepResult.error.message}`,
          };
        }

        const result = await grepTool(grepResult.data.toolOptions);
        return { success: true, data: grepResult.data, result };
      }

      case "new_file": {
        const newFileResult = NewFileSchema.safeParse(data);
        if (!newFileResult.success) {
          return {
            success: false,
            error: `Invalid new_file toolOptions: ${newFileResult.error.message}`,
          };
        }

        const result = await newFile({
          filePath: newFileResult.data.toolOptions.filePath,
          content: newFileResult.data.toolOptions.content,
        });
        return { success: true, data: newFileResult.data, result };
      }

      default:
        return {
          success: false,
          error: `Unknown tool: ${data.tool}. Supported tools: read_file, edit_file, shell_command, glob, grep, new_file`,
        };
    }
  } catch (error) {
    return {
      success: false,
      error: `Failed to parse tool call: ${error instanceof Error ? error.message : String(error)}`,
    };
  }
}
