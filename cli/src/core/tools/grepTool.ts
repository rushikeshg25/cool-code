import { ToolResult } from "../../types";
import { getErrorMessage } from "../utils";
import { GrepSchema } from "./toolValidator";
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

export async function grepTool(options: GrepToolOptions): Promise<ToolResult> {
  const validationResult = validateToolParams(options);
  if (validationResult) {
    return {
      LLMresult: validationResult,
      DisplayResult: "Fixing Errors while finding files",
    };
  }

  const searchPath = options.path ? path.resolve(options.path) : process.cwd();
  const includePattern = options.include ? new RegExp(options.include) : null;
  const regex = new RegExp(options.pattern);
  const matches: GrepMatch[] = [];

  function searchFile(filePath: string) {
    const lines = fs.readFileSync(filePath, "utf-8").split("\n");
    lines.forEach((line, idx) => {
      if (regex.test(line)) {
        matches.push({ filePath: filePath, lineNumber: idx + 1, line });
      }
    });
  }

  function walkDir(dir: string) {
    const entries = fs.readdirSync(dir, { withFileTypes: true });
    for (const entry of entries) {
      const fullPath = path.join(dir, entry.name);
      if (entry.isDirectory()) {
        walkDir(fullPath);
      } else if (entry.isFile()) {
        if (!includePattern || includePattern.test(entry.name)) {
          searchFile(fullPath);
        }
      }
    }
  }

  try {
    const stat = fs.statSync(searchPath);
    if (stat.isFile()) {
      if (!includePattern || includePattern.test(path.basename(searchPath))) {
        searchFile(searchPath);
      }
    } else if (stat.isDirectory()) {
      walkDir(searchPath);
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

function validateToolParams(options: GrepToolOptions): string | null {
  const result = GrepSchema.safeParse(options);
  if (!result.success) {
    return `Invalid grep parameters: ${result.error.message}`;
  }

  try {
    new RegExp(options.pattern);
  } catch (error) {
    return `Invalid regular expression pattern provided: ${options.pattern}. Error: ${getErrorMessage(error)}`;
  }

  return null;
}
