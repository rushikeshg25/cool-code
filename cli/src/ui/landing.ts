import cfonts from "cfonts";
import chalk from "chalk";
import { text, confirm } from "@clack/prompts";
import ora from "ora";
import { Processor } from "../core";

export async function showLanding() {
  // Handle Ctrl+C gracefully
  process.on("SIGINT", () => {
    console.log(chalk.yellow("\n\n👋 Goodbye! Thanks for using AI-DB-CLI!"));
    process.exit(0);
  });

  console.clear();

  cfonts.say("AI-DB-CLI", {
    font: "block",
    align: "center",
    colors: ["cyan", "magenta"],
    background: "transparent",
    letterSpacing: 1,
    lineHeight: 1,
    space: true,
    maxLength: "0",
  });

  console.log(chalk.gray("Welcome to AI Database CLI - Your database Agent"));
  console.log(chalk.gray("Press Ctrl+C to exit at any time\n"));
}
