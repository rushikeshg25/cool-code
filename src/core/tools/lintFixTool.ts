import type { ToolResult } from '../../types';
import { execCommand } from './shellTool';
import { readPackageJson } from './toolUtils';

export interface LintFixOptions {
  command?: string;
}

export async function lintFix(
  options: LintFixOptions,
  rootPath: string
): Promise<ToolResult> {
  let command = options.command;
  if (!command) {
    const pkg = readPackageJson(rootPath);
    if (pkg?.scripts?.['lint:fix']) {
      command = 'npm run lint:fix';
    } else if (pkg?.scripts?.lint) {
      command = 'npm run lint -- --fix';
    } else if (pkg?.scripts?.format) {
      command = 'npm run format';
    }
  }

  if (!command) {
    return {
      DisplayResult: 'No lint/format command found',
      LLMresult:
        'No lint/format command found. Provide a command explicitly.',
    };
  }

  const result = await execCommand({ command, directory: rootPath });
  return {
    DisplayResult: result.success
      ? 'Lint/format completed successfully'
      : 'Lint/format failed',
    LLMresult:
      result.stdout +
      (result.stderr ? `\nSTDERR:\n${result.stderr}` : ''),
  };
}
