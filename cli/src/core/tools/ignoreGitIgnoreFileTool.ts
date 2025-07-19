import * as fs from 'fs';
import * as path from 'path';

interface GitIgnoreChecker {
  (filePath: string): boolean;
}

export function createGitIgnoreChecker(rootDir: string): GitIgnoreChecker {
  const gitignorePath = path.join(rootDir, '.gitignore');

  if (!fs.existsSync(gitignorePath)) {
    return () => false;
  }

  try {
    const gitignoreContent = fs.readFileSync(gitignorePath, 'utf-8');
    const patterns = parseGitIgnorePatterns(gitignoreContent);
    const regexPatterns = patterns.map((pattern) =>
      createRegexFromGitIgnorePattern(pattern)
    );

    return (filePath: string): boolean => {
      const normalizedPath = path.relative(
        rootDir,
        path.resolve(rootDir, filePath)
      );

      return regexPatterns.some((regex) => regex.test(normalizedPath));
    };
  } catch (error) {
    console.warn(`Warning: Could not read .gitignore file: ${error}`);
    return () => false;
  }
}

function parseGitIgnorePatterns(content: string): string[] {
  return content
    .split('\n')
    .map((line) => line.trim())
    .filter((line) => line.length > 0 && !line.startsWith('#'))
    .filter((line) => !line.startsWith('!'));
}

function createRegexFromGitIgnorePattern(pattern: string): RegExp {
  const isAbsolute = pattern.startsWith('/');
  if (isAbsolute) {
    pattern = pattern.slice(1);
  }

  let regexPattern = pattern
    .replace(/[.+^${}()|[\]\\]/g, '\\$&')
    .replace(/\*\*/g, '§DOUBLESTAR§')
    .replace(/\*/g, '[^/]*')
    .replace(/§DOUBLESTAR§/g, '.*')
    .replace(/\?/g, '[^/]');

  if (pattern.endsWith('/')) {
    regexPattern = regexPattern.slice(0, -1) + '(/.*)?$';
  } else {
    regexPattern += '(/.*)?$';
  }

  if (isAbsolute || pattern.includes('/')) {
    regexPattern = '^' + regexPattern;
  } else {
    regexPattern = '(^|/)' + regexPattern;
  }

  return new RegExp(regexPattern);
}
