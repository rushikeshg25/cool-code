import { text } from '@clack/prompts';
import chalk from 'chalk';
import ora from 'ora';
import { AIDBProcessor } from '../core';

export async function acceptQuery() {
  while (true) {
    try {
      const query = await text({
        message: 'Enter your query:',
        placeholder:
          'e.g., "Create a users table with email and password fields"',
        validate: (value) => {
          if (!value || value.trim().length === 0) {
            return 'Please enter a query';
          }
        },
      });

      if (typeof query === 'symbol') {
        console.log(chalk.yellow('\nExiting...'));
        process.exit(0);
      }

      if (typeof query === 'string' && query.trim()) {
        await processQuery(query.trim());
      }

      console.log();
    } catch (error) {
      console.log(chalk.red('Error:'), error);
      process.exit(0);
    }
  }
}

async function processQuery(query: string) {
  const loadingSpinner = ora('Processing your query...').start();
  const processor = new AIDBProcessor();

  try {
    // Simulate processing time
    await new Promise((resolve) => setTimeout(resolve, 1000));

    const result = await processor.processQuery(query);

    loadingSpinner.succeed('Query processed successfully!');

    console.log(chalk.blue('\n📝 Query:'), result.query);
    console.log(chalk.green('✨ Response:'), result.response);

    if (result.suggestions && result.suggestions.length > 0) {
      console.log(chalk.yellow('💡 Suggestions:'));
      result.suggestions.forEach((suggestion) => {
        console.log(chalk.yellow(`  • ${suggestion}`));
      });
    }

    console.log(
      chalk.gray(
        `\n⏰ Processed at: ${result.timestamp.toLocaleTimeString()}\n`
      )
    );
  } catch (error) {
    loadingSpinner.fail('Failed to process query');
    console.error(chalk.red('Error:'), error);
  }
}
