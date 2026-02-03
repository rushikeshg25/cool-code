import { z } from "zod";
import type { configType } from "../processor";
import type { ToolResult } from "../../types";
import { readFile } from "./readFileTool";
import { editFile } from "./editTool";
import { execCommand } from "./shellTool";
import { globFiles } from "./globTool";
import { grepTool } from "./grepTool";
import { newFile } from "./newFileTool";
import { listRecentFiles } from "./listRecentFilesTool";
import { projectSummary } from "./projectSummaryTool";
import { openFileAt } from "./openFileAtTool";
import { runTests } from "./runTestsTool";
import { lintFix } from "./lintFixTool";
import { formatFile } from "./formatFileTool";
import { gitStatus } from "./gitStatusTool";
import { gitDiff } from "./gitDiffTool";
import { gitCommit } from "./gitCommitTool";
import { findSymbol } from "./findSymbolTool";
import { replaceInFiles } from "./replaceInFilesTool";
import { renameFile } from "./renameFileTool";
import { newModule } from "./newModuleTool";
import { addScript } from "./addScriptTool";
import { generateReadmeSection } from "./generateReadmeSectionTool";

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

export const ListRecentFilesSchema = z.object({
  tool: z.literal("list_recent_files"),
  toolOptions: z.object({
    limit: z.number().int().positive().optional(),
    include: z.string().optional(),
    exclude: z.string().optional(),
  }),
});

export const ProjectSummarySchema = z.object({
  tool: z.literal("project_summary"),
  toolOptions: z.object({}).optional(),
});

export const OpenFileAtSchema = z.object({
  tool: z.literal("open_file_at"),
  toolOptions: z.object({
    absolutePath: z.string().min(1, "File path cannot be empty"),
    startLine: z.number().int().positive().optional(),
    endLine: z.number().int().positive().optional(),
  }),
});

export const RunTestsSchema = z.object({
  tool: z.literal("run_tests"),
  toolOptions: z.object({
    command: z.string().optional(),
  }),
});

export const LintFixSchema = z.object({
  tool: z.literal("lint_fix"),
  toolOptions: z.object({
    command: z.string().optional(),
  }),
});

export const FormatFileSchema = z.object({
  tool: z.literal("format_file"),
  toolOptions: z.object({
    absolutePath: z.string().min(1, "File path cannot be empty"),
  }),
});

export const GitStatusSchema = z.object({
  tool: z.literal("git_status"),
  toolOptions: z.object({}).optional(),
});

export const GitDiffSchema = z.object({
  tool: z.literal("git_diff"),
  toolOptions: z.object({
    filePath: z.string().optional(),
    staged: z.boolean().optional(),
  }),
});

export const GitCommitSchema = z.object({
  tool: z.literal("git_commit"),
  toolOptions: z.object({
    message: z.string().min(1, "Commit message cannot be empty"),
    all: z.boolean().optional(),
    files: z.array(z.string()).optional(),
  }),
});

export const FindSymbolSchema = z.object({
  tool: z.literal("find_symbol"),
  toolOptions: z.object({
    pattern: z.string().min(1, "Pattern cannot be empty"),
    include: z.string().optional(),
    path: z.string().optional(),
  }),
});

export const ReplaceInFilesSchema = z.object({
  tool: z.literal("replace_in_files"),
  toolOptions: z.object({
    pattern: z.string().min(1, "Pattern cannot be empty"),
    replacement: z.string(),
    include: z.string().optional(),
    exclude: z.string().optional(),
    useRegex: z.boolean().optional(),
    dryRun: z.boolean().optional(),
  }),
});

export const RenameFileSchema = z.object({
  tool: z.literal("rename_file"),
  toolOptions: z.object({
    fromPath: z.string().min(1, "From path cannot be empty"),
    toPath: z.string().min(1, "To path cannot be empty"),
    overwrite: z.boolean().optional(),
  }),
});

export const NewModuleSchema = z.object({
  tool: z.literal("new_module"),
  toolOptions: z.object({
    moduleName: z.string().min(1, "Module name cannot be empty"),
    baseDir: z.string().optional(),
    exportFromRootIndex: z.boolean().optional(),
  }),
});

export const AddScriptSchema = z.object({
  tool: z.literal("add_script"),
  toolOptions: z.object({
    name: z.string().min(1, "Script name cannot be empty"),
    command: z.string().min(1, "Script command cannot be empty"),
    overwrite: z.boolean().optional(),
  }),
});

export const GenerateReadmeSectionSchema = z.object({
  tool: z.literal("generate_readme_section"),
  toolOptions: z.object({
    title: z.string().min(1, "Title cannot be empty"),
    bullets: z.array(z.string()).optional(),
    content: z.string().optional(),
  }),
});

export type ToolCall =
  | z.infer<typeof ReadFileSchema>
  | z.infer<typeof EditFileSchema>
  | z.infer<typeof ShellCommandSchema>
  | z.infer<typeof GlobSchema>
  | z.infer<typeof GrepSchema>
  | z.infer<typeof NewFileSchema>
  | z.infer<typeof ListRecentFilesSchema>
  | z.infer<typeof ProjectSummarySchema>
  | z.infer<typeof OpenFileAtSchema>
  | z.infer<typeof RunTestsSchema>
  | z.infer<typeof LintFixSchema>
  | z.infer<typeof FormatFileSchema>
  | z.infer<typeof GitStatusSchema>
  | z.infer<typeof GitDiffSchema>
  | z.infer<typeof GitCommitSchema>
  | z.infer<typeof FindSymbolSchema>
  | z.infer<typeof ReplaceInFilesSchema>
  | z.infer<typeof RenameFileSchema>
  | z.infer<typeof NewModuleSchema>
  | z.infer<typeof AddScriptSchema>
  | z.infer<typeof GenerateReadmeSectionSchema>;

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

      case "list_recent_files": {
        const listResult = ListRecentFilesSchema.safeParse(data);
        if (!listResult.success) {
          return {
            success: false,
            error: `Invalid list_recent_files toolOptions: ${listResult.error.message}`,
          };
        }
        const result = listRecentFiles(
          listResult.data.toolOptions || {},
          rootPath
        );
        return { success: true, data: listResult.data, result };
      }

      case "project_summary": {
        const summaryResult = ProjectSummarySchema.safeParse(data);
        if (!summaryResult.success) {
          return {
            success: false,
            error: `Invalid project_summary toolOptions: ${summaryResult.error.message}`,
          };
        }
        const result = projectSummary(rootPath);
        return { success: true, data: summaryResult.data, result };
      }

      case "open_file_at": {
        const openResult = OpenFileAtSchema.safeParse(data);
        if (!openResult.success) {
          return {
            success: false,
            error: `Invalid open_file_at toolOptions: ${openResult.error.message}`,
          };
        }
        const result = openFileAt(openResult.data.toolOptions, rootPath);
        return { success: true, data: openResult.data, result };
      }

      case "run_tests": {
        const runTestsResult = RunTestsSchema.safeParse(data);
        if (!runTestsResult.success) {
          return {
            success: false,
            error: `Invalid run_tests toolOptions: ${runTestsResult.error.message}`,
          };
        }
        const result = await runTests(runTestsResult.data.toolOptions, rootPath);
        return { success: true, data: runTestsResult.data, result };
      }

      case "lint_fix": {
        const lintFixResult = LintFixSchema.safeParse(data);
        if (!lintFixResult.success) {
          return {
            success: false,
            error: `Invalid lint_fix toolOptions: ${lintFixResult.error.message}`,
          };
        }
        const result = await lintFix(lintFixResult.data.toolOptions, rootPath);
        return { success: true, data: lintFixResult.data, result };
      }

      case "format_file": {
        const formatResult = FormatFileSchema.safeParse(data);
        if (!formatResult.success) {
          return {
            success: false,
            error: `Invalid format_file toolOptions: ${formatResult.error.message}`,
          };
        }
        const result = await formatFile(formatResult.data.toolOptions, rootPath);
        return { success: true, data: formatResult.data, result };
      }

      case "git_status": {
        const statusResult = GitStatusSchema.safeParse(data);
        if (!statusResult.success) {
          return {
            success: false,
            error: `Invalid git_status toolOptions: ${statusResult.error.message}`,
          };
        }
        const result = await gitStatus(rootPath);
        return { success: true, data: statusResult.data, result };
      }

      case "git_diff": {
        const diffResult = GitDiffSchema.safeParse(data);
        if (!diffResult.success) {
          return {
            success: false,
            error: `Invalid git_diff toolOptions: ${diffResult.error.message}`,
          };
        }
        const result = await gitDiff(diffResult.data.toolOptions, rootPath);
        return { success: true, data: diffResult.data, result };
      }

      case "git_commit": {
        const commitResult = GitCommitSchema.safeParse(data);
        if (!commitResult.success) {
          return {
            success: false,
            error: `Invalid git_commit toolOptions: ${commitResult.error.message}`,
          };
        }
        const result = await gitCommit(commitResult.data.toolOptions, rootPath);
        return { success: true, data: commitResult.data, result };
      }

      case "find_symbol": {
        const findResult = FindSymbolSchema.safeParse(data);
        if (!findResult.success) {
          return {
            success: false,
            error: `Invalid find_symbol toolOptions: ${findResult.error.message}`,
          };
        }
        const result = await findSymbol(findResult.data.toolOptions, rootPath);
        return { success: true, data: findResult.data, result };
      }

      case "replace_in_files": {
        const replaceResult = ReplaceInFilesSchema.safeParse(data);
        if (!replaceResult.success) {
          return {
            success: false,
            error: `Invalid replace_in_files toolOptions: ${replaceResult.error.message}`,
          };
        }
        const result = replaceInFiles(
          replaceResult.data.toolOptions,
          rootPath
        );
        return { success: true, data: replaceResult.data, result };
      }

      case "rename_file": {
        const renameResult = RenameFileSchema.safeParse(data);
        if (!renameResult.success) {
          return {
            success: false,
            error: `Invalid rename_file toolOptions: ${renameResult.error.message}`,
          };
        }
        const result = renameFile(renameResult.data.toolOptions, rootPath);
        return { success: true, data: renameResult.data, result };
      }

      case "new_module": {
        const moduleResult = NewModuleSchema.safeParse(data);
        if (!moduleResult.success) {
          return {
            success: false,
            error: `Invalid new_module toolOptions: ${moduleResult.error.message}`,
          };
        }
        const result = newModule(moduleResult.data.toolOptions, rootPath);
        return { success: true, data: moduleResult.data, result };
      }

      case "add_script": {
        const scriptResult = AddScriptSchema.safeParse(data);
        if (!scriptResult.success) {
          return {
            success: false,
            error: `Invalid add_script toolOptions: ${scriptResult.error.message}`,
          };
        }
        const result = addScript(scriptResult.data.toolOptions, rootPath);
        return { success: true, data: scriptResult.data, result };
      }

      case "generate_readme_section": {
        const readmeResult = GenerateReadmeSectionSchema.safeParse(data);
        if (!readmeResult.success) {
          return {
            success: false,
            error: `Invalid generate_readme_section toolOptions: ${readmeResult.error.message}`,
          };
        }
        const result = generateReadmeSection(
          readmeResult.data.toolOptions,
          rootPath
        );
        return { success: true, data: readmeResult.data, result };
      }

      default:
        return {
          success: false,
          error: `Unknown tool: ${data.tool}. Supported tools: read_file, edit_file, shell_command, glob, grep, new_file, list_recent_files, project_summary, open_file_at, run_tests, lint_fix, format_file, git_status, git_diff, git_commit, find_symbol, replace_in_files, rename_file, new_module, add_script, generate_readme_section`,
        };
    }
  } catch (error) {
    return {
      success: false,
      error: `Failed to parse tool call: ${error instanceof Error ? error.message : String(error)}`,
    };
  }
}
