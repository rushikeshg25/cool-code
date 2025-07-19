// import {
//   validateToolCall,
//   parseToolCall,
//   ValidationResult,
// } from './toolValidator';
// import { readFileFromToolCall, ReadFileResult } from './readFileTool';
// import { editFile, appendToFile, replaceInFile, EditResult } from './editTool';
// import { executeCommand, ShellResult } from './shellTool';
// import { glob, GlobResult } from './globTool';

// export interface ToolExecutionResult {
//   success: boolean;
//   toolName: string;
//   result?: any;
//   error?: string;
//   executionTime: number;
// }

// /**
//  * Execute a tool call from JSON string
//  */
// export async function executeToolFromJson(
//   jsonString: string
// ): Promise<ToolExecutionResult> {
//   const startTime = Date.now();

//   // Parse and validate the JSON tool call
//   const validation = parseToolCall(jsonString);
//   if (!validation.success || !validation.data) {
//     return {
//       success: false,
//       toolName: 'unknown',
//       error: validation.error,
//       executionTime: Date.now() - startTime,
//     };
//   }

//   return executeToolCall(validation.data);
// }

// /**
//  * Execute a validated tool call
//  */
// export async function executeToolCall(
//   toolCall: any
// ): Promise<ToolExecutionResult> {
//   const startTime = Date.now();

//   try {
//     switch (toolCall.tool) {
//       case 'read_file':
//         return {
//           success: true,
//           toolName: 'read_file',
//           result: readFileFromToolCall(toolCall),
//           executionTime: Date.now() - startTime,
//         };

//       case 'edit_file':
//         const editResult = editFile(
//           toolCall.parameters.absolute_path,
//           toolCall.parameters.content,
//           { backup: toolCall.parameters.backup }
//         );
//         return {
//           success: editResult.success,
//           toolName: 'edit_file',
//           result: editResult,
//           error: editResult.success ? undefined : editResult.message,
//           executionTime: Date.now() - startTime,
//         };

//       case 'shell_command':
//         const shellResult = await executeCommand(toolCall.parameters.command, {
//           cwd: toolCall.parameters.cwd,
//           timeout: toolCall.parameters.timeout,
//         });
//         return {
//           success: shellResult.success,
//           toolName: 'shell_command',
//           result: shellResult,
//           error: shellResult.success ? undefined : shellResult.stderr,
//           executionTime: Date.now() - startTime,
//         };

//       case 'glob':
//         const globResult = glob(
//           toolCall.parameters.pattern,
//           toolCall.parameters.rootDir,
//           {
//             recursive: toolCall.parameters.recursive,
//             includeDirectories: toolCall.parameters.includeDirectories,
//           }
//         );
//         return {
//           success: true,
//           toolName: 'glob',
//           result: globResult,
//           executionTime: Date.now() - startTime,
//         };

//       default:
//         return {
//           success: false,
//           toolName: toolCall.tool || 'unknown',
//           error: `Unsupported tool: ${toolCall.tool}`,
//           executionTime: Date.now() - startTime,
//         };
//     }
//   } catch (error) {
//     return {
//       success: false,
//       toolName: toolCall.tool || 'unknown',
//       error: error instanceof Error ? error.message : String(error),
//       executionTime: Date.now() - startTime,
//     };
//   }
// }

// /**
//  * Execute multiple tool calls in sequence
//  */
// export async function executeMultipleTools(
//   toolCalls: string[]
// ): Promise<ToolExecutionResult[]> {
//   const results: ToolExecutionResult[] = [];

//   for (const toolCall of toolCalls) {
//     const result = await executeToolFromJson(toolCall);
//     results.push(result);

//     // Stop execution if a critical tool fails
//     if (!result.success && isCriticalTool(result.toolName)) {
//       break;
//     }
//   }

//   return results;
// }

// /**
//  * Check if a tool is critical (failure should stop execution)
//  */
// function isCriticalTool(toolName: string): boolean {
//   const criticalTools = ['read_file', 'edit_file'];
//   return criticalTools.includes(toolName);
// }

// /**
//  * Format tool execution result for display
//  */
// export function formatToolResult(result: ToolExecutionResult): string {
//   const {
//     success,
//     toolName,
//     result: toolResult,
//     error,
//     executionTime,
//   } = result;

//   let output = `🔧 Tool: ${toolName}\n`;
//   output += `⏱️  Execution time: ${executionTime}ms\n`;

//   if (success) {
//     output += `✅ Status: Success\n`;

//     switch (toolName) {
//       case 'read_file':
//         const readResult = toolResult as ReadFileResult;
//         output += `📄 File: ${readResult.path}\n`;
//         output += `📊 Size: ${readResult.size} bytes\n`;
//         if (readResult.lineRange) {
//           output += `📝 Lines: ${readResult.lineRange.actual.start}-${readResult.lineRange.actual.end} of ${readResult.lineRange.totalLines}\n`;
//         }
//         output += `📋 Content:\n${readResult.content}\n`;
//         break;

//       case 'edit_file':
//         const editResult = toolResult as EditResult;
//         output += `📝 ${editResult.message}\n`;
//         break;

//       case 'shell_command':
//         const shellResult = toolResult as ShellResult;
//         output += `💻 Command: ${shellResult.command}\n`;
//         output += `📤 Exit code: ${shellResult.exitCode}\n`;
//         if (shellResult.stdout) {
//           output += `📋 Output:\n${shellResult.stdout}\n`;
//         }
//         if (shellResult.stderr) {
//           output += `⚠️  Errors:\n${shellResult.stderr}\n`;
//         }
//         break;

//       case 'glob':
//         const globResult = toolResult as GlobResult;
//         output += `📁 Found ${globResult.total} items\n`;
//         output += `📄 Files: ${globResult.files.length}\n`;
//         output += `📂 Directories: ${globResult.directories.length}\n`;
//         if (globResult.files.length > 0) {
//           output += `📋 Files:\n${globResult.files.slice(0, 10).join('\n')}\n`;
//           if (globResult.files.length > 10) {
//             output += `... and ${globResult.files.length - 10} more\n`;
//           }
//         }
//         break;
//     }
//   } else {
//     output += `❌ Status: Failed\n`;
//     output += `🚨 Error: ${error}\n`;
//   }

//   return output;
// }
