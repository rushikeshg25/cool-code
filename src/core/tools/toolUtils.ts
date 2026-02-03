import * as fs from 'fs';
import * as path from 'path';

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
