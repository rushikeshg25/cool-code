import fs from 'fs';
import { ToolResult } from '../../types';
import { getErrorMessage } from '../utils';

interface NewFileOptions {
  filePath: string;
  content: string;
}

export async function newFile(options: NewFileOptions): Promise<ToolResult> {
  const { filePath, content } = options;

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
