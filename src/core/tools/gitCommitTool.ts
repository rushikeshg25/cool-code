import type { ToolResult } from '../../types';
import { execCommand } from './shellTool';
import { ensureAbsoluteWithinRoot, shellEscapeSingleQuotes, toRelativePath } from './toolUtils';

export interface GitCommitOptions {
  message: string;
  all?: boolean;
  files?: string[];
}

export async function gitCommit(
  options: GitCommitOptions,
  rootPath: string
): Promise<ToolResult> {
  if (!options.message || options.message.trim() === '') {
    return {
      DisplayResult: 'Fixing Issues',
      LLMresult: 'Commit message is required.',
    };
  }

  if (options.all) {
    const addAll = await execCommand({
      command: 'git add -A',
      directory: rootPath,
    });
    if (!addAll.success) {
      return {
        DisplayResult: 'Git add failed',
        LLMresult:
          addAll.stdout +
          (addAll.stderr ? `\nSTDERR:\n${addAll.stderr}` : ''),
      };
    }
  } else if (options.files && options.files.length > 0) {
    const relativeFiles: string[] = [];
    for (const filePath of options.files) {
      const validation = ensureAbsoluteWithinRoot(filePath, rootPath);
      if (validation) {
        return { DisplayResult: 'Fixing Issues', LLMresult: validation };
      }
      relativeFiles.push(toRelativePath(filePath, rootPath));
    }
    const addCmd = `git add -- ${relativeFiles.map((f) => `"${f}"`).join(' ')}`;
    const addResult = await execCommand({ command: addCmd, directory: rootPath });
    if (!addResult.success) {
      return {
        DisplayResult: 'Git add failed',
        LLMresult:
          addResult.stdout +
          (addResult.stderr ? `\nSTDERR:\n${addResult.stderr}` : ''),
      };
    }
  }

  const msg = shellEscapeSingleQuotes(options.message);
  const commitCmd = `git commit -m '${msg}'`;
  const commitResult = await execCommand({
    command: commitCmd,
    directory: rootPath,
  });

  return {
    DisplayResult: commitResult.success ? 'Commit created' : 'Commit failed',
    LLMresult:
      commitResult.stdout +
      (commitResult.stderr ? `\nSTDERR:\n${commitResult.stderr}` : ''),
  };
}
