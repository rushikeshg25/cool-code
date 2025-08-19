#!/usr/bin/env node

import { program } from "commander";
import dotenv from "dotenv";
import { runCli } from "./ui";
import { readFileSync } from "fs";
import { join } from "path";
import chalk from "chalk";

async function main() {
  dotenv.config({
    quiet: true,
  });

  const packageJson = JSON.parse(
    readFileSync(join(__dirname, "../package.json"), "utf-8")
  );

  program
    .name("cool-code")
    .description("Cli coding Ai agent like Claude Code and Gemini Cli")
    .version(packageJson.version);

  program.action(async () => {
    // Check if API key is set
    if (!process.env.GOOGLE_GENERATIVE_AI_API_KEY) {
      console.error(chalk.red("❌ Missing API Key!"));
      console.error("");
      console.error(chalk.yellow("Please set your Google AI API key:"));
      console.error("");
      console.error(
        chalk.cyan("  export GOOGLE_GENERATIVE_AI_API_KEY=your_api_key_here")
      );
      console.error("");
      console.error(
        chalk.blue(
          "Get your API key at: https://aistudio.google.com/app/apikey"
        )
      );
      process.exit(1);
    }

    await runCli();
  });

  program.parse();
}

main().catch(console.error);
