import * as path from 'path';
import type { ToolResult } from '../../types';
import { execCommand } from './shellTool';
import { ensureAbsoluteWithinRoot, shellEscapeSingleQuotes } from './toolUtils';

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

  const searchPath = options.path
    ? path.resolve(rootPath, options.path)
    : rootPath;
  const outsideRoot = ensureAbsoluteWithinRoot(searchPath, rootPath);
  if (outsideRoot) {
    return { DisplayResult: 'Invalid path', LLMresult: outsideRoot };
  }
  const includeFlag = options.include
    ? ` -g '${shellEscapeSingleQuotes(options.include)}'`
    : '';
  const command = `rg -n --hidden --glob '!.git/*' --glob '!node_modules/*'${includeFlag} '${shellEscapeSingleQuotes(
    options.pattern
  )}' '${shellEscapeSingleQuotes(searchPath)}'`;
  const result = await execCommand({ command, directory: rootPath });

  return {
    DisplayResult: result.success ? 'Symbol search results' : 'Symbol search failed',
    LLMresult:
      result.stdout +
      (result.stderr ? `\nSTDERR:\n${result.stderr}` : ''),
  };
}
