#!/usr/bin/env node

import { program } from "commander";
import dotenv from "dotenv";
import { runCli } from "./ui";
import { readFileSync } from "fs";
import { join } from "path";

async function main() {
  // Suppress dotenv output by temporarily redirecting console
  const originalLog = console.log;
  console.log = () => {};
  
  dotenv.config();

  console.log = originalLog;

  const packageJson = JSON.parse(
    readFileSync(join(__dirname, "../package.json"), "utf-8")
  );

  program
    .name("cool-code")
    .description("Cli coding Ai agent like Claude Code and Gemini Cli")
    .version(packageJson.version);

  program.action(async () => {
    await runCli();
  });

  program.parse();
}

main().catch(console.error);
