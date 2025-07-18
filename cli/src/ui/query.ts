import { text } from "@clack/prompts";
import chalk from "chalk";
import { Processor } from "../core";
import { DynamicSpinner } from "./spinner";

export async function acceptQuery(rootDir: string) {
  const processor = new Processor(rootDir);

  while (true) {
    try {
      const query = await text({
        message: "Enter your query:",
        placeholder:
          'e.g., "Create a users table with email and password fields"',
        validate: (value) => {
          if (!value || value.trim().length === 0) {
            return "Please enter a query";
          }
        },
      });

      if (typeof query === "symbol") {
        console.log(chalk.yellow("\nExiting..."));
        process.exit(0);
      }

      if (typeof query === "string" && query.trim()) {
        await processQueryandShowLoader(query.trim(), processor);
      }

      console.log();
    } catch (error) {
      console.log(chalk.red("Error:"), error);
      process.exit(0);
    }
  }
}

async function processQueryandShowLoader(query: string, processor: Processor) {
  const spinner = new DynamicSpinner("Starting query processing...");

  try {
    await processor.processQuery(query, spinner);

    spinner.succeed("Query processed successfully!");
  } catch (error) {
    spinner.fail("Failed to process query");
    console.error(chalk.red("Error:"), error);
  }
}
