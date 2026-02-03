import * as fs from 'fs';
import type { ToolResult } from '../../types';
import { ensureAbsoluteWithinRoot } from './toolUtils';

export interface RenameFileOptions {
  fromPath: string;
  toPath: string;
  overwrite?: boolean;
}

export function renameFile(
  options: RenameFileOptions,
  rootPath: string
): ToolResult {
  const fromValidation = ensureAbsoluteWithinRoot(options.fromPath, rootPath);
  if (fromValidation) {
    return { DisplayResult: 'Fixing Issues', LLMresult: fromValidation };
  }
  const toValidation = ensureAbsoluteWithinRoot(options.toPath, rootPath);
  if (toValidation) {
    return { DisplayResult: 'Fixing Issues', LLMresult: toValidation };
  }

  if (!fs.existsSync(options.fromPath)) {
    return {
      DisplayResult: 'Fixing Issues',
      LLMresult: `Source file does not exist: ${options.fromPath}`,
    };
  }
  if (fs.existsSync(options.toPath) && !options.overwrite) {
    return {
      DisplayResult: 'Fixing Issues',
      LLMresult: `Target already exists: ${options.toPath}`,
    };
  }

  fs.renameSync(options.fromPath, options.toPath);
  return {
    DisplayResult: 'File renamed',
    LLMresult: `Renamed ${options.fromPath} to ${options.toPath}`,
  };
}
