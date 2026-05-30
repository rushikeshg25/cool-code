import { confirm, select } from '@clack/prompts';
import chalk from 'chalk';
import { showLanding } from './landing';
import { Processor } from '../core/processor';
import type { CoolCodeConfig } from '../core/config';
import type { AgentMode } from '../types';
import { loadSession, latestSession, newSessionId } from '../core/session';

interface RunCliOptions {
  autoInit?: boolean;
  noInit?: boolean;
  quiet?: boolean;
  copy?: boolean;
  continueSession?: boolean;
  resumeId?: string;
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

  let initialMode: AgentMode = 'agent';
  if (!options.quiet) {
    const selected = await select({
      message: 'Select operational mode:',
      options: [
        { value: 'agent', label: 'Agent', hint: 'Auto-execute tasks' },
        { value: 'plan', label: 'Plan', hint: 'Investigate and produce a detailed plan only' },
        { value: 'ask', label: 'Ask', hint: 'Read-only / Q&A' },
      ],
      initialValue: 'agent',
    });
    if (typeof selected === 'symbol') process.exit(0);
    initialMode = selected as AgentMode;
  }

  const processor = new Processor(rootDir, config, {
    quiet: options.quiet,
    allowDangerous: config.features?.allowDangerous,
    mode: initialMode,
  });

  // Resolve and restore a prior session if requested.
  let sessionId = newSessionId();
  const restoreFrom = options.resumeId
    ? loadSession(options.resumeId)
    : options.continueSession
    ? latestSession(rootDir)
    : null;
  if (restoreFrom) {
    sessionId = restoreFrom.id;
    processor.restoreSnapshot(restoreFrom);
    console.log(
      chalk.green(`Resumed session ${sessionId.slice(0, 8)} (${restoreFrom.conversations.length} messages).`)
    );
  }

  // Lazy-load the Ink app so non-interactive CLI commands don't pull in Ink/React.
  const { runApp } = await import('./app');
  await runApp({
    processor,
    rootDir,
    copy: Boolean(options.copy),
    sessionId,
  });
  process.exit(0);
}
