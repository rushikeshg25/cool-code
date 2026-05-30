import * as fs from 'fs';
import * as path from 'path';
import { loadConfig } from '../config';

// Returns a block reason if reading the file is disallowed by the configured
// guardrails (blockReadPatterns), otherwise null. Shared by read_file, pinned
// files, and the :pin command so the guardrail can't be bypassed.
export function isBlockedPath(filePath: string, rootPath: string): string | null {
  const config = loadConfig(rootPath);
  const patterns = config.guardrails?.blockReadPatterns || [];
  if (patterns.length === 0) return null;

  const base = path.basename(filePath);
  for (const pattern of patterns) {
    if (pattern === base) {
      return `Reading blocked for "${base}" by guardrails.`;
    }
    if (pattern.startsWith('.') && pattern.endsWith('.*')) {
      const prefix = pattern.slice(0, -2);
      if (base.startsWith(prefix)) {
        return `Reading blocked for "${base}" by guardrails.`;
      }
    }
    if (pattern.startsWith('*.')) {
      const ext = pattern.slice(1);
      if (base.endsWith(ext)) {
        return `Reading blocked for "${base}" by guardrails.`;
      }
    }
  }
  return null;
}

export function ensureAbsoluteWithinRoot(
  absolutePath: string,
  rootPath: string
): string | null {
  if (!path.isAbsolute(absolutePath)) {
    return 'File path must be absolute';
  }
  const resolvedRoot = path.resolve(rootPath);
  const resolvedPath = path.resolve(absolutePath);
  if (!resolvedPath.startsWith(resolvedRoot + path.sep)) {
    return `Path must be within project root: ${resolvedRoot}`;
  }
  return null;
}

export function readPackageJson(rootPath: string): any | null {
  const pkgPath = path.join(rootPath, 'package.json');
  if (!fs.existsSync(pkgPath)) return null;
  try {
    return JSON.parse(fs.readFileSync(pkgPath, 'utf-8'));
  } catch {
    return null;
  }
}

export function writePackageJson(rootPath: string, data: any) {
  const pkgPath = path.join(rootPath, 'package.json');
  fs.writeFileSync(pkgPath, JSON.stringify(data, null, 2));
}

export function shellEscapeSingleQuotes(value: string): string {
  return value.replace(/'/g, `'\"'\"'`);
}

export function toRelativePath(filePath: string, rootPath: string) {
  return path.relative(rootPath, filePath);
}

export function toPascalCase(input: string): string {
  return input
    .replace(/[^a-zA-Z0-9]+/g, ' ')
    .trim()
    .split(' ')
    .filter(Boolean)
    .map((part) => part[0].toUpperCase() + part.slice(1))
    .join('');
}
