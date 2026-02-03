import type { ToolResult } from '../../types';
import { execCommand } from './shellTool';
import { ensureAbsoluteWithinRoot, toRelativePath } from './toolUtils';

export interface GitDiffOptions {
  filePath?: string;
  staged?: boolean;
}

export async function gitDiff(
  options: GitDiffOptions,
  rootPath: string
): Promise<ToolResult> {
  let fileArg = '';
  if (options.filePath) {
    const validation = ensureAbsoluteWithinRoot(options.filePath, rootPath);
    if (validation) {
      return { DisplayResult: 'Fixing Issues', LLMresult: validation };
    }
    fileArg = ` -- "${toRelativePath(options.filePath, rootPath)}"`;
  }

  const command = `git diff${options.staged ? ' --staged' : ''}${fileArg}`;
  const result = await execCommand({ command, directory: rootPath });
  return {
    DisplayResult: result.success ? 'Git diff' : 'Git diff failed',
    LLMresult:
      result.stdout +
      (result.stderr ? `\nSTDERR:\n${result.stderr}` : ''),
  };
}
