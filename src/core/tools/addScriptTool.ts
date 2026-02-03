import type { ToolResult } from '../../types';
import { readPackageJson, writePackageJson } from './toolUtils';

export interface AddScriptOptions {
  name: string;
  command: string;
  overwrite?: boolean;
}

export function addScript(
  options: AddScriptOptions,
  rootPath: string
): ToolResult {
  if (!options.name || !options.command) {
    return {
      DisplayResult: 'Fixing Issues',
      LLMresult: 'name and command are required.',
    };
  }

  const pkg = readPackageJson(rootPath);
  if (!pkg) {
    return {
      DisplayResult: 'Fixing Issues',
      LLMresult: 'package.json not found or invalid.',
    };
  }

  pkg.scripts = pkg.scripts || {};
  if (pkg.scripts[options.name] && !options.overwrite) {
    return {
      DisplayResult: 'Fixing Issues',
      LLMresult: `Script "${options.name}" already exists.`,
    };
  }

  pkg.scripts[options.name] = options.command;
  writePackageJson(rootPath, pkg);

  return {
    DisplayResult: 'Script added',
    LLMresult: `Added script "${options.name}".`,
  };
}
