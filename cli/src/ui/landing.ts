import cfonts from 'cfonts';
import chalk from 'chalk';
import { text, confirm } from '@clack/prompts';
import ora from 'ora';
import { AIDBProcessor } from '../core';

export async function showLanding() {
  process.on('SIGINT', () => {
    console.log(chalk.yellow('\n\nExiting...'));
    process.exit(0);
  });

  console.clear();

  cfonts.say('AI-DB-CLI', {
    font: 'block',
    align: 'center',
    colors: ['cyan', 'magenta'],
    background: 'transparent',
    letterSpacing: 1,
    lineHeight: 1,
    space: true,
    maxLength: '0',
  });

  console.log(chalk.gray('Welcome to AI Database CLI - Your database Agent'));
  console.log(chalk.gray('Press Ctrl+C to exit at any time\n'));

}