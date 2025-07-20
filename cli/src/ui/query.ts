import { text } from "@clack/prompts";
import chalk from "chalk";
import { Processor } from "../core/processor";

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
  try {
    await processor.processQuery(query);
    // Spinner completion is handled in processor during streaming
  } catch (error) {
    console.error(chalk.red("Error:"), error);
  }
}
