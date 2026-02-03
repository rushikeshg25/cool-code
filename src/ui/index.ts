import { confirm, select } from '@clack/prompts';
import chalk from 'chalk';
import { showLanding } from './landing';
import { acceptQuery } from './query';
import type { CoolCodeConfig } from '../core/config';

interface RunCliOptions {
  autoInit?: boolean;
  noInit?: boolean;
  quiet?: boolean;
  copy?: boolean;
}

export async function runCli(config: CoolCodeConfig, options: RunCliOptions = {}) {
  const rootDir = process.cwd();

  if (!options.quiet) {
    await showLanding();
  }
  if (options.noInit) {
    console.log(chalk.yellow('\nExiting without initialization.'));
    process.exit(0);
  }

  let shouldInit: any = true;
  if (!options.autoInit) {
    shouldInit = await confirm({
      message: `Initialize Cool-Code in ${rootDir}?`,
      initialValue: true,
    });
  }

  if (typeof shouldInit === 'symbol' || shouldInit === false) {
    console.log(chalk.yellow('\nExiting without initialization.'));
    process.exit(0);
  }

  let initialMode: any = 'agent';
  if (!options.quiet) {
    initialMode = await select({
      message: 'Select operational mode:',
      options: [
        { value: 'agent', label: '🤖 Agent', hint: 'Auto-execute tasks' },
        { value: 'planning', label: '📋 Planning', hint: 'Generate task list only' },
        { value: 'ask', label: '💬 Ask', hint: 'Read-only/Q&A' },
      ],
      initialValue: 'agent',
    });
  }

  if (typeof initialMode === 'symbol') {
    process.exit(0);
  }

  await acceptQuery(rootDir, config, Boolean(options.quiet), Boolean(options.copy), initialMode);
}
