import * as fs from 'fs';
import { globSync } from 'glob';
import * as path from 'path';
import type { ToolResult } from '../../types';

export interface ListRecentFilesOptions {
  limit?: number;
  include?: string;
  exclude?: string;
}

export function listRecentFiles(
  options: ListRecentFilesOptions,
  rootPath: string
): ToolResult {
  const limit = options.limit ?? 20;
  const include = options.include ?? '**/*';
  const ignore = [
    '**/node_modules/**',
    '**/.git/**',
    '**/dist/**',
    '**/build/**',
  ];
  if (options.exclude) {
    ignore.push(options.exclude);
  }

  const files = globSync(include, {
    cwd: rootPath,
    nodir: true,
    dot: false,
    ignore,
  }).map((rel) => path.join(rootPath, rel));

  const withStats = files
    .map((filePath) => {
      try {
        const stat = fs.statSync(filePath);
        return {
          path: filePath,
          mtimeMs: stat.mtimeMs,
        };
      } catch {
        return null;
      }
    })
    .filter(Boolean) as { path: string; mtimeMs: number }[];

  withStats.sort((a, b) => b.mtimeMs - a.mtimeMs);
  const result = withStats.slice(0, limit).map((item) => ({
    path: path.relative(rootPath, item.path),
    modifiedAt: new Date(item.mtimeMs).toISOString(),
  }));

  return {
    DisplayResult: `Found ${result.length} recent files`,
    LLMresult: JSON.stringify(result, null, 2),
  };
}
