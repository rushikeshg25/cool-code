// import { executeToolFromJson, formatToolResult } from './toolExecutor';
// import { parseToolCall } from './toolValidator';

// /**
//  * Example usage of the tool validation and execution system
//  */

// // Example 1: Valid read_file tool call
// const validReadFileExample = `{
//   "tool": "read_file",
//   "parameters": {
//     "absolute_path": "/home/user/project/config.json",
//     "startLine": 10,
//     "endLine": 50
//   }
// }`;

// // Example 2: Invalid read_file tool call (missing absolute_path)
// const invalidReadFileExample = `{
//   "tool": "read_file",
//   "parameters": {
//     "startLine": 10,
//     "endLine": 50
//   }
// }`;

// // Example 3: Valid edit_file tool call
// const validEditFileExample = `{
//   "tool": "edit_file",
//   "parameters": {
//     "absolute_path": "/home/user/project/config.json",
//     "content": "{\\"updated\\": true}",
//     "backup": true
//   }
// }`;

// // Example 4: Valid shell_command tool call
// const validShellExample = `{
//   "tool": "shell_command",
//   "parameters": {
//     "command": "ls -la",
//     "cwd": "/home/user/project"
//   }
// }`;

// // Example 5: Valid glob tool call
// const validGlobExample = `{
//   "tool": "glob",
//   "parameters": {
//     "pattern": "*.ts",
//     "rootDir": "/home/user/project",
//     "recursive": true
//   }
// }`;

// /**
//  * Test function to demonstrate validation
//  */
// export async function testToolValidation() {
//   console.log('🧪 Testing Tool Validation System\n');

//   const examples = [
//     { name: 'Valid Read File', json: validReadFileExample },
//     { name: 'Invalid Read File', json: invalidReadFileExample },
//     { name: 'Valid Edit File', json: validEditFileExample },
//     { name: 'Valid Shell Command', json: validShellExample },
//     { name: 'Valid Glob', json: validGlobExample }
//   ];

//   for (const example of examples) {
//     console.log(`\n📋 Testing: ${example.name}`);
//     console.log(`📄 JSON: ${example.json}`);

//     // Test validation only
//     const validation = parseToolCall(example.json);
//     if (validation.success) {
//       console.log('✅ Validation: PASSED');
//       console.log(`🔧 Tool: ${validation.data?.tool}`);
//     } else {
//       console.log('❌ Validation: FAILED');
//       console.log(`🚨 Error: ${validation.error}`);
//     }

//     // Test execution (only for valid ones)
//     if (validation.success) {
//       try {
//         const result = await executeToolFromJson(example.json);
//         console.log(`⚡ Execution: ${result.success ? 'SUCCESS' : 'FAILED'}`);
//         if (!result.success) {
//           console.log(`🚨 Execution Error: ${result.error}`);
//         }
//       } catch (error) {
//         console.log(`💥 Execution Exception: ${error}`);
//       }
//     }

//     console.log('─'.repeat(50));
//   }
// }

// /**
//  * Example of how to handle tool calls in your LLM response processing
//  */
// export async function processLLMResponseWithTools(llmResponse: string): Promise<string> {
//   // Look for JSON tool calls in the LLM response
//   const toolCallRegex = /\{"tool":\s*"[^"]+",\s*"parameters":\s*\{[^}]+\}\}/g;
//   const toolCalls = llmResponse.match(toolCallRegex);

//   if (!toolCalls || toolCalls.length === 0) {
//     return llmResponse; // No tool calls found
//   }

//   let processedResponse = llmResponse;

//   for (const toolCallJson of toolCalls) {
//     try {
//       console.log(`🔧 Executing tool call: ${toolCallJson}`);

//       const result = await executeToolFromJson(toolCallJson);
//       const formattedResult = formatToolResult(result);

//       // Replace the tool call JSON with the formatted result
//       processedResponse = processedResponse.replace(toolCallJson, formattedResult);

//     } catch (error) {
//       console.error(`❌ Tool execution failed: ${error}`);
//       processedResponse = processedResponse.replace(
//         toolCallJson,
//         `❌ Tool execution failed: ${error}`
//       );
//     }
//   }

//   return processedResponse;
// }

// // Export examples for testing
// export const examples = {
//   validReadFile: validReadFileExample,
//   invalidReadFile: invalidReadFileExample,
//   validEditFile: validEditFileExample,
//   validShell: validShellExample,
//   validGlob: validGlobExample
// };
