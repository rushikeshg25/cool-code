import type { ToolResult } from '../../types';
import { execCommand } from './shellTool';

export interface FindSymbolOptions {
  pattern: string;
  include?: string;
  path?: string;
}

export async function findSymbol(
  options: FindSymbolOptions,
  rootPath: string
): Promise<ToolResult> {
  if (!options.pattern || options.pattern.trim() === '') {
    return {
      DisplayResult: 'Fixing Issues',
      LLMresult: 'pattern is required.',
    };
  }

  const searchPath = options.path || rootPath;
  const includeFlag = options.include ? ` -g "${options.include}"` : '';
  const command = `rg -n --hidden --glob '!.git/*' --glob '!node_modules/*'${includeFlag} "${options.pattern}" "${searchPath}"`;
  const result = await execCommand({ command, directory: rootPath });

  return {
    DisplayResult: result.success ? 'Symbol search results' : 'Symbol search failed',
    LLMresult:
      result.stdout +
      (result.stderr ? `\nSTDERR:\n${result.stderr}` : ''),
  };
}
