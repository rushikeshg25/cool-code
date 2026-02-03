import type { ToolResult } from '../../types';
import { execCommand } from './shellTool';

export async function gitStatus(rootPath: string): Promise<ToolResult> {
  const result = await execCommand({
    command: 'git status --short -b',
    directory: rootPath,
  });
  return {
    DisplayResult: result.success ? 'Git status' : 'Git status failed',
    LLMresult:
      result.stdout +
      (result.stderr ? `\nSTDERR:\n${result.stderr}` : ''),
  };
}
