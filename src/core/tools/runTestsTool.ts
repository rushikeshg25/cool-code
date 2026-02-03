import type { ToolResult } from '../../types';
import { execCommand } from './shellTool';
import { readPackageJson } from './toolUtils';

export interface RunTestsOptions {
  command?: string;
}

export async function runTests(
  options: RunTestsOptions,
  rootPath: string
): Promise<ToolResult> {
  let command = options.command;
  if (!command) {
    const pkg = readPackageJson(rootPath);
    if (pkg?.scripts?.test) {
      command = 'npm run test';
    }
  }
  if (!command) {
    return {
      DisplayResult: 'No test command found',
      LLMresult: 'No test command found. Provide a command explicitly.',
    };
  }

  const result = await execCommand({ command, directory: rootPath });
  return {
    DisplayResult: result.success
      ? 'Tests completed successfully'
      : 'Tests failed',
    LLMresult:
      result.stdout +
      (result.stderr ? `\nSTDERR:\n${result.stderr}` : ''),
  };
}
