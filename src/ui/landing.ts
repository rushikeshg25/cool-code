import cfonts from 'cfonts';
import chalk from 'chalk';
import { glyph } from './theme';

export async function showLanding() {
  // Handle Ctrl+C gracefully
  process.on('SIGINT', () => {
    console.log(chalk.dim('\n\nSession ended.'));
    process.exit(0);
  });

  console.clear();

  cfonts.say('COOLCODE', {
    font: 'tiny',
    align: 'left',
    colors: ['cyan'],
  });

  const version = require('../../package.json').version;
  const sep = chalk.dim(`  ${glyph.bullet}  `);
  console.log(chalk.dim(`  v${version}`));
  console.log(
    chalk.dim('  ') +
      chalk.cyan(':help') +
      chalk.dim(' commands') +
      sep +
      chalk.cyan('shift+tab') +
      chalk.dim(' mode') +
      sep +
      chalk.cyan('ctrl+e') +
      chalk.dim(' effort') +
      sep +
      chalk.cyan('ctrl+c') +
      chalk.dim(' exit') +
      '\n'
  );
}
