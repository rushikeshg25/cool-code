import { glob } from "glob";
import { configType } from "../processor";
import { ToolResult } from "../../types";

interface globOptions {
  pattern: string;
}

export async function globFiles(
  { pattern }: globOptions,
  config: configType
): Promise<ToolResult> {
  const results = await glob(pattern, { ignore: "node_modules/**" });
  const filteredResults = results.filter(
    (filePath) => !config.doesExistInGitIgnore(filePath)
  );
  const LLMresult = filteredResults.join("\n");
  return {
    LLMresult,
    DisplayResult: `Found ${filteredResults.length} file(s)`,
  };
}
