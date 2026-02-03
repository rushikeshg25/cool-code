import type { ToolResult } from '../../types';
import { scanProject } from '../utils';

export function projectSummary(rootPath: string): ToolResult {
  const scan = scanProject(rootPath);
  return {
    DisplayResult: 'Project summary generated',
    LLMresult: JSON.stringify(scan, null, 2),
  };
}
