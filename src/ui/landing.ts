import cfonts from 'cfonts';
import chalk from 'chalk';

export async function showLanding() {
  // Handle Ctrl+C gracefully
  process.on('SIGINT', () => {
    console.log(chalk.yellow('\n\n👋 Goodbye! Thanks for using Cool-Code!'));
    process.exit(0);
  });

  console.clear();

  cfonts.say('Cool-Code', {
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
