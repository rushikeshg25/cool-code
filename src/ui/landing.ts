import cfonts from 'cfonts';
import chalk from 'chalk';

export async function showLanding() {
  // Handle Ctrl+C gracefully
  process.on('SIGINT', () => {
    console.log(chalk.yellow('\n\n👋 Goodbye! Thanks for using Cool-Code!'));
    process.exit(0);
  });

  console.clear();
 
  cfonts.say('COOLCODE', {
    font: 'tiny',
    align: 'left',
    colors: ['cyan', 'magenta'],
  });

  const version = require('../../package.json').version;
  console.log(chalk.gray(`  v${version}`));
  console.log(chalk.gray('  Press ') + chalk.white.bold(':help') + chalk.gray(' for commands, ') + chalk.white.bold(':mode') + chalk.gray(' to switch modes, or ') + chalk.white.bold('Ctrl+C') + chalk.gray(' to exit\n'));
}
