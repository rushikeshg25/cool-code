import * as fs from "fs";
import * as path from "path";
import { ToolResult } from "../../types";

export interface EditOptions {
  filePath: string;
  oldString: string;
  newString: string;
  expected_replacements: number;
}

export function editFile(options: EditOptions): ToolResult {
  const { filePath, newString, oldString, expected_replacements } = options;

  if (!path.isAbsolute(filePath)) {
    return {
      DisplayResult: "Fixing Issues",
      LLMresult: "File path must be absolute.",
    };
  }
  if (!fs.existsSync(filePath)) {
    return {
      DisplayResult: "Fixing Issues",
      LLMresult: `File does not exist: ${filePath}`,
    };
  }
  const stats = fs.statSync(filePath);
  if (!stats.isFile()) {
    return {
      DisplayResult: "Fixing Issues",
      LLMresult: `Path is not a file: ${filePath}`,
    };
  }
  try {
    fs.accessSync(filePath, fs.constants.R_OK | fs.constants.W_OK);
  } catch {
    return {
      DisplayResult: "Fixing Issues",
      LLMresult: `File is not readable or writable: ${filePath}`,
    };
  }

  // Read file
  let content = fs.readFileSync(filePath, "utf-8");
  let count = 0;
  let newContent = content;
  if (oldString === "") {
    return {
      DisplayResult: "Fixing Issues",
      LLMresult: "oldString cannot be empty.",
    };
  }

  let idx = 0;
  while (count < expected_replacements) {
    idx = newContent.indexOf(oldString, idx);
    if (idx === -1) break;
    newContent =
      newContent.slice(0, idx) +
      newString +
      newContent.slice(idx + oldString.length);
    count++;
    idx += newString.length;
  }

  if (count === 0) {
    return {
      DisplayResult: "No replacements made",
      LLMresult: "No occurrences of the old string were found.",
    };
  }

  fs.writeFileSync(filePath, newContent, "utf-8");

  return {
    DisplayResult: `Replaced ${count} occurrence(s) of "${oldString}" with "${newString}" in ${path.basename(filePath)}`,
    LLMresult: newContent,
  };
}
