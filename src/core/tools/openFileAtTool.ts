import type { ToolResult } from '../../types';
import { readFile } from './readFileTool';

export interface OpenFileAtOptions {
  absolutePath: string;
  startLine?: number;
  endLine?: number;
}

export function openFileAt(
  options: OpenFileAtOptions,
  rootPath: string
): ToolResult {
  return readFile(
    {
      absolutePath: options.absolutePath,
      startLine: options.startLine,
      endLine: options.endLine,
    },
    rootPath
  );
}
