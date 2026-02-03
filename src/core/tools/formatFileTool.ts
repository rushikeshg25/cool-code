import type { ToolResult } from '../../types';
import { execCommand } from './shellTool';
import { ensureAbsoluteWithinRoot, toRelativePath } from './toolUtils';

export interface FormatFileOptions {
  absolutePath: string;
}

export async function formatFile(
  options: FormatFileOptions,
  rootPath: string
): Promise<ToolResult> {
  const validation = ensureAbsoluteWithinRoot(options.absolutePath, rootPath);
  if (validation) {
    return { DisplayResult: 'Fixing Issues', LLMresult: validation };
  }
  const rel = toRelativePath(options.absolutePath, rootPath);
  const command = `npx prettier --write "${rel}"`;
  const result = await execCommand({ command, directory: rootPath });
  return {
    DisplayResult: result.success ? 'File formatted' : 'Formatting failed',
    LLMresult:
      result.stdout +
      (result.stderr ? `\nSTDERR:\n${result.stderr}` : ''),
  };
}
