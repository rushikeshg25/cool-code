import * as fs from "fs";
import * as path from "path";
import { ToolResult } from "../../types";
import { isBlockedPath } from "./toolUtils";

export interface ReadFileOptions {
  absolutePath: string;
  startLine?: number;
  endLine?: number;
}

export interface ReadFileResult {
  success: boolean;
  content: string;
  lines?: string[];
  size: number;
  path: string;
  error?: string;
  lineRange?: {
    requested: { start?: number; end?: number };
    actual: { start: number; end: number };
    totalLines: number;
  };
}

export async function readFile(
  options: ReadFileOptions,
  rootPath: string
): Promise<ToolResult> {
  const { absolutePath, startLine, endLine } = options;
  const blocked = isBlockedPath(absolutePath, rootPath);
  if (blocked) {
    return {
      DisplayResult: "Fixing Issues",
      LLMresult: blocked,
    };
  }
  const validation = validateFileForReading(absolutePath);
  if (validation) {
    return {
      DisplayResult: "Fixing Issues",
      LLMresult: validation,
    };
  }
  try {
    if (!startLine && !endLine) {
      const content = await fs.promises.readFile(absolutePath, "utf-8");
      return {
        DisplayResult: "Reading " + path.relative(rootPath, absolutePath),
        LLMresult: content,
      };
    } else {
      if (startLine === undefined || endLine === undefined) {
        return {
          DisplayResult: "Fixing Issues",
          LLMresult: "Both startLine and endLine must be provided.",
        };
      }
      if (startLine < 1 || endLine < 1) {
        return {
          DisplayResult: "Fixing Issues",
          LLMresult: "startLine and endLine must be greater than 0.",
        };
      }
      if (endLine < startLine) {
        return {
          DisplayResult: "Fixing Issues",
          LLMresult: "endLine must be greater than or equal to startLine.",
        };
      }

      // Memory efficient reading for specific lines
      const readline = require('readline');
      const instream = fs.createReadStream(absolutePath);
      const rl = readline.createInterface({
        input: instream,
        terminal: false
      });

      let currentLine = 0;
      const selectedLines: string[] = [];

      for await (const line of rl) {
        currentLine++;
        if (currentLine >= startLine && currentLine <= endLine) {
          selectedLines.push(line);
        }
        if (currentLine > endLine) {
          rl.close();
          instream.destroy();
          break;
        }
      }

      if (currentLine < startLine) {
        return {
          DisplayResult: "Fixing Issues",
          LLMresult: `File only has ${currentLine} lines, but startLine is ${startLine}.`,
        };
      }

      const finalContent = selectedLines.join("\n");
      return {
        DisplayResult:
          `Reading lines ${startLine}-${endLine} from ` +
          path.relative(rootPath, absolutePath),
        LLMresult: finalContent,
      };
    }
  } catch (error) {
    console.log("error while reading file:", error);
    return {
      DisplayResult: "Fixing Issues",
      LLMresult: error instanceof Error ? error.message : String(error),
    };
  }
}

export function validateFileForReading(filePath: string): string | null {
  if (!path.isAbsolute(filePath)) {
    return "File path must be absolute";
  }

  if (!fs.existsSync(filePath)) {
    return `File does not exist: ${filePath}`;
  }

  const stats = fs.statSync(filePath);

  if (!stats.isFile()) {
    return `Path is not a file: ${filePath}`;
  }

  try {
    fs.accessSync(filePath, fs.constants.R_OK);
  } catch {
    return `File is not readable: ${filePath}`;
  }
  return null;
}
