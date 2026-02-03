#!/usr/bin/env node

import { program } from "commander";
import dotenv from "dotenv";
import { runCli } from "./ui";
import { readFileSync, existsSync, writeFileSync } from "fs";
import { join } from "path";
import chalk from "chalk";
import {
  loadConfig,
  saveConfig,
  setByPath,
  getByPath,
  parseConfigValue,
  getConfigPath,
} from "./core/config";
import { scanProject } from "./core/utils";
import { createTaskPlan } from "./core/taskPlanner";

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

  program
    .option("-y, --yes", "Initialize in current directory without prompting")
    .option("--no-init", "Exit without initializing in the current directory")
    .option("--quiet", "Reduce UI output for automation")
    .option("--allow-dangerous", "Allow dangerous actions without prompting")
    .option("--copy", "Copy final responses to clipboard");

  program.addHelpText(
    "after",
    `
Examples:
  cool-code
  cool-code --yes
  cool-code --copy
  cool-code scan
  cool-code scan --json
  cool-code task "Add user authentication"
  cool-code config set llm.model "gemini-2.5-flash"
`
  );

  const configCmd = program
    .command("config")
    .description("Get or set configuration in .coolcode.json");

  configCmd
    .command("get")
    .argument("<key>", "Config key path, e.g. llm.model")
    .action((key: string) => {
      const rootDir = process.cwd();
      const config = loadConfig(rootDir);
      const value = getByPath(config, key);
      if (value === undefined) {
        console.log("");
        console.log(`No value set for "${key}".`);
        console.log(`Config file: ${getConfigPath(rootDir)}`);
        return;
      }
      console.log(JSON.stringify(value, null, 2));
    });

  configCmd
    .command("set")
    .argument("<key>", "Config key path, e.g. llm.model")
    .argument("<value>", "Value. JSON supported, e.g. true or 1024")
    .action((key: string, value: string) => {
      const rootDir = process.cwd();
      const config = loadConfig(rootDir);
      setByPath(config, key, parseConfigValue(value));
      saveConfig(rootDir, config);
      console.log(`Updated ${key}.`);
      console.log(`Config file: ${getConfigPath(rootDir)}`);
    });

  program
    .command("scan")
    .description("Summarize the current project structure")
    .option("--refresh", "Refresh cached scan results")
    .option("--json", "Output raw JSON")
    .action((options: { refresh?: boolean; json?: boolean }) => {
      const rootDir = process.cwd();
      const config = loadConfig(rootDir);
      const cachePath = join(rootDir, ".coolcode.scan.json");
      if (!options.refresh && config.features?.scanCache && existsSync(cachePath)) {
        const cached = JSON.parse(readFileSync(cachePath, "utf-8"));
        if (options.json) {
          console.log(JSON.stringify(cached, null, 2));
        } else {
          printScan(cached);
        }
        return;
      }
      const scan = scanProject(rootDir);
      if (config.features?.scanCache) {
        writeFileSync(cachePath, JSON.stringify(scan, null, 2));
      }
      if (options.json) {
        console.log(JSON.stringify(scan, null, 2));
      } else {
        printScan(scan);
      }
    });

  program
    .command("task")
    .description("Generate a structured plan for a task")
    .argument("<goal>", "Task goal in plain language")
    .option("--json", "Output raw JSON")
    .action(async (goal: string, options: { json?: boolean }) => {
      const rootDir = process.cwd();
      const config = loadConfig(rootDir);
      const plan = await createTaskPlan(goal, config);
      if (!plan) {
        console.log("Could not generate a plan. Try rephrasing the goal.");
        return;
      }
      if (options.json) {
        console.log(JSON.stringify(plan, null, 2));
      } else {
        printTaskPlan(plan);
      }
    });

  program.action(async () => {
    const rootDir = process.cwd();
    const config = loadConfig(rootDir);
    const options = program.opts<{ yes?: boolean; noInit?: boolean; quiet?: boolean; allowDangerous?: boolean; copy?: boolean }>();
    if (options.allowDangerous) {
      config.features = config.features || {};
      config.features.allowDangerous = true;
    }
    await runCli(config, {
      autoInit: options.yes,
      noInit: options.noInit,
      quiet: options.quiet,
      copy: options.copy,
    });
  });

  program.parse();
}

main().catch(console.error);

function printScan(scan: any) {
  console.log("");
  console.log(chalk.cyan.bold("Project Scan"));
  console.log(`${chalk.gray("Root:")} ${scan.rootDir}`);
  console.log(`${chalk.gray("Timestamp:")} ${scan.timestamp}`);
  console.log(`${chalk.gray("Entry points:")} ${formatList(scan.entrypoints)}`);
  console.log(`${chalk.gray("Frameworks:")} ${formatList(scan.frameworks)}`);
  console.log(`${chalk.gray("Scripts:")} ${formatList(scan.scripts)}`);
  console.log(`${chalk.gray("Languages:")} ${formatList(scan.languages)}`);
  console.log(`${chalk.gray("tsconfig.json:")} ${scan.hasTsConfig ? "yes" : "no"}`);
  console.log(`${chalk.gray("README.md:")} ${scan.hasReadme ? "yes" : "no"}`);
}

function printTaskPlan(plan: any) {
  console.log("");
  console.log(`${chalk.magenta.bold("Goal:")} ${plan.goal}`);
  console.log("");
  console.log(chalk.cyan.bold("Steps:"));
  for (let i = 0; i < plan.steps.length; i++) {
    const step = plan.steps[i];
    console.log(`${i + 1}. ${step.title}`);
    console.log(`   ${step.detail}`);
  }
  console.log("");
  console.log(`${chalk.yellow.bold("Assumptions:")} ${formatList(plan.assumptions)}`);
  console.log(`${chalk.red.bold("Risks:")} ${formatList(plan.risks)}`);
}

function formatList(items: string[]) {
  if (!items || items.length === 0) return "none";
  return items.join(", ");
}
