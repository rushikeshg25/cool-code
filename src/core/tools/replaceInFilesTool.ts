import * as fs from 'fs';
import * as path from 'path';
import { globSync } from 'glob';
import type { ToolResult } from '../../types';
import { debugLog } from '../utils';

export interface ReplaceInFilesOptions {
  pattern: string;
  replacement: string;
  include?: string;
  exclude?: string;
  useRegex?: boolean;
  dryRun?: boolean;
}

export function replaceInFiles(
  options: ReplaceInFilesOptions,
  rootPath: string
): ToolResult {
  if (!options.pattern) {
    return { DisplayResult: 'Fixing Issues', LLMresult: 'pattern is required.' };
  }

  let compiledRegex: RegExp | null = null;
  if (options.useRegex) {
    try {
      compiledRegex = new RegExp(options.pattern, 'g');
    } catch (error) {
      return {
        DisplayResult: 'Invalid pattern',
        LLMresult: `Invalid regex pattern: ${
          error instanceof Error ? error.message : String(error)
        }`,
      };
    }
  }

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

  const results: { file: string; replacements: number }[] = [];
  let total = 0;
  const dryRun = options.dryRun !== false;

  for (const filePath of files) {
    let content: string;
    try {
      content = fs.readFileSync(filePath, 'utf-8');
    } catch (error) {
      debugLog(`replaceInFiles: skipped unreadable file ${filePath}`, error);
      continue;
    }

    let replaced = content;
    let count = 0;

    if (options.useRegex && compiledRegex) {
      compiledRegex.lastIndex = 0;
      replaced = content.replace(compiledRegex, (match) => {
        count += 1;
        return options.replacement;
      });
    } else {
      const parts = content.split(options.pattern);
      if (parts.length > 1) {
        count = parts.length - 1;
        replaced = parts.join(options.replacement);
      }
    }

    if (count > 0) {
      total += count;
      results.push({
        file: path.relative(rootPath, filePath),
        replacements: count,
      });
      if (!dryRun) {
        fs.writeFileSync(filePath, replaced);
      }
    }
  }

  const summary = {
    dryRun,
    filesChanged: results.length,
    totalReplacements: total,
    files: results.slice(0, 50),
  };

  return {
    DisplayResult: dryRun
      ? `Dry run complete: ${total} replacements in ${results.length} files`
      : `Replaced ${total} occurrences in ${results.length} files`,
    LLMresult: JSON.stringify(summary, null, 2),
  };
}
