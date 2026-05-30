import fs from 'fs';
import { ToolResult } from '../../types';
import { getErrorMessage } from '../utils';
import { ensureAbsoluteWithinRoot } from './toolUtils';

interface NewFileOptions {
  filePath: string;
  content: string;
}

export async function newFile(
  options: NewFileOptions,
  rootPath: string
): Promise<ToolResult> {
  const { filePath, content } = options;

  const validation = ensureAbsoluteWithinRoot(filePath, rootPath);
  if (validation) {
    return { DisplayResult: 'Fixing Issues', LLMresult: validation };
  }

  try {
    await fs.promises.writeFile(filePath, content);
    return {
      LLMresult: `File ${filePath} created successfully`,
      DisplayResult: `File ${filePath} created successfully`,
    };
  } catch (error) {
    return {
      LLMresult: `Error creating file ${filePath}: ${getErrorMessage(error)}`,
      DisplayResult: `Error creating file ${filePath}`,
    };
  }
}
