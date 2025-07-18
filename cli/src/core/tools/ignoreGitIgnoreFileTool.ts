import * as fs from 'fs';
import * as path from 'path';

interface GitIgnoreChecker {
  (filePath: string): boolean;
}

/**
 * Creates a function that checks if a file should be ignored based on .gitignore patterns
 * @param rootDir - The root directory to look for .gitignore file
 * @returns A function that returns true if the file should be ignored, false otherwise
 */
export function createGitIgnoreChecker(rootDir: string): GitIgnoreChecker {
  const gitignorePath = path.join(rootDir, '.gitignore');
  
  if (!fs.existsSync(gitignorePath)) {
    return () => false;
  }

  try {
    const gitignoreContent = fs.readFileSync(gitignorePath, 'utf-8');
    const patterns = parseGitIgnorePatterns(gitignoreContent);
    const regexPatterns = patterns.map(pattern => createRegexFromGitIgnorePattern(pattern));

    return (filePath: string): boolean => {
      const normalizedPath = path.relative(rootDir, path.resolve(rootDir, filePath));
      
      return regexPatterns.some(regex => regex.test(normalizedPath));
    };
  } catch (error) {
    console.warn(`Warning: Could not read .gitignore file: ${error}`);
    return () => false;
  }
}


function parseGitIgnorePatterns(content: string): string[] {
  return content
    .split('\n')
    .map(line => line.trim())
    .filter(line => line.length > 0 && !line.startsWith('#')) 
    .filter(line => !line.startsWith('!')); 
}

/**
 * Converts a gitignore pattern to a regular expression
 */
function createRegexFromGitIgnorePattern(pattern: string): RegExp {
  // Handle leading slash (absolute path from repo root)
  const isAbsolute = pattern.startsWith('/');
  if (isAbsolute) {
    pattern = pattern.slice(1);
  }

  let regexPattern = pattern
    .replace(/[.+^${}()|[\]\\]/g, '\\$&')
    .replace(/\*\*/g, '§DOUBLESTAR§') // Temporary placeholder
    .replace(/\*/g, '[^/]*') // Single * matches anything except /
    .replace(/§DOUBLESTAR§/g, '.*') // ** matches anything including /
    .replace(/\?/g, '[^/]'); // ? matches any single character except /

  if (pattern.endsWith('/')) {
    regexPattern = regexPattern.slice(0, -1) + '(/.*)?$';
  } else {
    regexPattern += '(/.*)?$';
  }

  if (isAbsolute || pattern.includes('/')) {
    // Pattern is absolute or contains /, match from root
    regexPattern = '^' + regexPattern;
  } else {
    // Pattern doesn't contain /, match at any level
    regexPattern = '(^|/)' + regexPattern;
  }

  return new RegExp(regexPattern);
}

/**
 * Example usage:
 * 
 * const checker = createGitIgnoreChecker('/path/to/project');
 * 
 * console.log(checker('node_modules/package.json')); // true if node_modules is in .gitignore
 * console.log(checker('src/index.ts')); // false if src files are not ignored
 * console.log(checker('.env')); // true if .env is in .gitignore
 */