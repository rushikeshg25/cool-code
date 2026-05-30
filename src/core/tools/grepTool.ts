import { ToolResult } from "../../types";
import { getErrorMessage } from "../utils";
import { globSync } from "glob";
import * as fs from "fs";
import * as path from "path";

export interface GrepToolOptions {
  pattern: string;
  path?: string;
  include?: string;
}

interface GrepMatch {
  filePath: string;
  lineNumber: number;
  line: string;
}

const IGNORE_GLOBS = [
  "**/node_modules/**",
  "**/.git/**",
  "**/dist/**",
  "**/build/**",
];

const MAX_FILE_SIZE = 5 * 1024 * 1024; // 5MB

function isBinary(buffer: Buffer): boolean {
  const len = Math.min(buffer.length, 8000);
  for (let i = 0; i < len; i++) {
    if (buffer[i] === 0) return true;
  }
  return false;
}

export async function grepTool(options: GrepToolOptions): Promise<ToolResult> {
  let regex: RegExp;
  try {
    regex = new RegExp(options.pattern);
  } catch (error) {
    return {
      LLMresult: `Invalid search pattern: ${getErrorMessage(error)}`,
      DisplayResult: "Invalid pattern",
    };
  }

  let includePattern: RegExp | null = null;
  if (options.include) {
    try {
      includePattern = new RegExp(options.include);
    } catch (error) {
      return {
        LLMresult: `Invalid include pattern: ${getErrorMessage(error)}`,
        DisplayResult: "Invalid pattern",
      };
    }
  }

  const searchPath = options.path ? path.resolve(options.path) : process.cwd();
  const matches: GrepMatch[] = [];

  async function searchFile(filePath: string) {
    let buffer: Buffer;
    try {
      const stat = await fs.promises.stat(filePath);
      if (stat.size > MAX_FILE_SIZE) return;
      buffer = await fs.promises.readFile(filePath);
    } catch {
      return;
    }
    if (isBinary(buffer)) return;
    const lines = buffer.toString("utf-8").split("\n");
    lines.forEach((line, idx) => {
      if (regex.test(line)) {
        matches.push({ filePath, lineNumber: idx + 1, line });
      }
    });
  }

  try {
    const stat = await fs.promises.stat(searchPath);
    if (stat.isFile()) {
      if (!includePattern || includePattern.test(path.basename(searchPath))) {
        await searchFile(searchPath);
      }
    } else if (stat.isDirectory()) {
      const files = globSync("**/*", {
        cwd: searchPath,
        nodir: true,
        dot: false,
        ignore: IGNORE_GLOBS,
      }).map((rel) => path.join(searchPath, rel));
      for (const file of files) {
        if (!includePattern || includePattern.test(path.basename(file))) {
          await searchFile(file);
        }
      }
    } else {
      return {
        LLMresult: "Provided path is neither a file nor a directory.",
        DisplayResult: "Invalid path",
      };
    }
  } catch (error) {
    return {
      LLMresult: `Error reading path: ${getErrorMessage(error)}`,
      DisplayResult: "Error while searching",
    };
  }

  if (matches.length === 0) {
    return {
      LLMresult: "No matches found.",
      DisplayResult: "No matches found",
    };
  }

  const resultString = matches
    .map((m) => `${m.filePath}:${m.lineNumber}: ${m.line}`)
    .join("\n");

  return {
    LLMresult: resultString,
    DisplayResult: `Found ${matches.length} match(es)`,
  };
}
