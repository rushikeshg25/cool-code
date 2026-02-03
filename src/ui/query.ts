import chalk from 'chalk';
import path from 'path';
import * as fs from 'fs';
import { Processor } from '../core/processor';
import type { CoolCodeConfig } from '../core/config';
import type { TaskList, AgentMode } from '../types';
import { createPromptSession } from './prompt';
import { toolRegistery } from '../core/tools/tool-registery';
import { copyToClipboard } from './clipboard';

export async function acceptQuery(
  rootDir: string,
  config: CoolCodeConfig,
  quiet: boolean,
  copy: boolean,
  initialMode: AgentMode = 'agent'
) {
  const commands = [':help', ':exit', ':quit', ':context', ':clear', ':mode', ':pin', ':unpin'];
  const tools = toolRegistery.map((t) => t.name);
  const session = createPromptSession({
    message: chalk.cyan('Enter your query:') + '\n' + chalk.gray('❯ '),
    completions: [...commands, ...tools],
  });

  const processor = new Processor(rootDir, config, {
    quiet,
    allowDangerous: config.features?.allowDangerous,
    mode: initialMode,
    confirm: async (message) => session.confirm(message, false),
    confirmEdit: async (message, preview) => {
      console.log(chalk.yellow(`\n${message}`));
      if (preview) {
        console.log(chalk.gray('--- preview ---'));
        console.log(preview);
        console.log(chalk.gray('--- end preview ---'));
      }
      return session.confirm('Proceed?', false);
    },
  });

  // Set up non-blocking input for background queuing
  session.onInput((input) => {
    processor.enqueueMessage(input);
  });

  while (true) {
    try {
      const taskList = processor.getTaskList();
      if (taskList && !quiet) {
        renderTasks(taskList);
      }
      if (!quiet) {
        renderStatus(processor, rootDir);
      }

      const query = await session.ask();
      if (query === null) {
        console.log(chalk.yellow('\nExiting...'));
        process.exit(0);
      }
      if (!query) {
        continue;
      }

      if (query.startsWith(':')) {
        await handleCommand(query, processor, session);
        continue;
      }

      const result = await processQuery(query, processor);
      if (copy && result) {
        const copied = copyToClipboard(result);
        if (!quiet) {
          console.log(
            copied.success
              ? chalk.green('Copied to clipboard.')
              : chalk.red(`Copy failed: ${copied.error}`)
          );
        }
      }

      console.log();
    } catch (error) {
      console.log(chalk.red('Something went wrong.'));
      const show = await session.confirm('Show details?', false);
      if (show) {
        console.log(error instanceof Error ? error.stack || error.message : String(error));
      }
    }
  }
}

async function processQuery(query: string, processor: Processor) {
  try {
    return await processor.processQuery(query);
  } catch (error) {
    console.error(chalk.red('Error:'), error);
    return null;
  }
}

function renderTasks(taskList: TaskList) {
  console.log(chalk.bold.white(`\n  🎯 Goal: ${taskList.goal}`));
  for (const item of taskList.items) {
    let icon = chalk.gray('○');
    if (item.status === 'in-progress') icon = chalk.yellow('▶');
    if (item.status === 'done') icon = chalk.green('✓');
    if (item.status === 'failed') icon = chalk.red('✗');
    
    const text = item.status === 'done' ? chalk.gray.strikethrough(item.title) : item.title;
    console.log(`    ${icon} ${text}`);
  }
}

function renderStatus(processor: Processor, rootDir: string) {
  const status = processor.getStatus();
  const mode = processor.getMode();
  const modeColors: Record<AgentMode, (s: string) => string> = {
    planning: chalk.yellow,
    agent: chalk.green,
    ask: chalk.blue,
  };
  const modeIcon: Record<AgentMode, string> = {
    planning: '📋',
    agent: '🤖',
    ask: '💬',
  };
  const divider = chalk.gray('│');
  
  const statusLine = [
    modeColors[mode](`${modeIcon[mode]} ${mode.toUpperCase()}`),
    divider,
    chalk.cyan(`${status.model}`),
    divider,
    chalk.blue(`📁 ${path.basename(rootDir)}`),
    divider,
    chalk.magenta(`💬 ${status.messageCount}`),
    divider,
    chalk.green(`🪙  ${Math.round(status.totalTokens / 100) / 10}k tokens`),
  ].join(' ');

  console.log(`${chalk.gray('╾')} ${statusLine}`);
}

async function handleCommand(
  cmd: string,
  processor: Processor,
  session: ReturnType<typeof createPromptSession>
) {
  const command = cmd.trim().toLowerCase();
  if (command === ':exit' || command === ':quit') {
    console.log(chalk.yellow('\nExiting...'));
    process.exit(0);
  }
  if (command === ':help') {
    console.log(chalk.bold.white('\n  Interactive Commands:'));
    console.log(chalk.gray('  ╾─────────────────────'));
    console.log(`  ${chalk.cyan(':help')}      Show this help menu`);
    console.log(`  ${chalk.cyan(':mode')}      Switch between ${chalk.yellow('planning')}, ${chalk.green('agent')}, or ${chalk.blue('ask')} modes`);
    console.log(`  ${chalk.cyan(':pin')}       Pin a file to context (e.g. :pin src/index.ts)`);
    console.log(`  ${chalk.cyan(':unpin')}     Unpin a file from context`);
    console.log(`  ${chalk.cyan(':context')}   Preview context, pinned files, and token usage`);
    console.log(`  ${chalk.cyan(':clear')}     Clear the terminal screen`);
    console.log(`  ${chalk.cyan(':exit')}      Close the session`);
    console.log(chalk.gray('  ╾─────────────────────\n'));
    return;
  }
  if (command.startsWith(':mode')) {
    const parts = command.split(/\s+/);
    if (parts.length > 1) {
      const targetMode = parts[1] as AgentMode;
      if (['planning', 'agent', 'ask'].includes(targetMode)) {
        processor.setMode(targetMode);
        console.log(chalk.green(`\n✓ Mode switched to ${targetMode.toUpperCase()}`));
      } else {
        console.log(chalk.red(`\nInvalid mode: ${targetMode}. Available: planning, agent, ask.`));
      }
    } else {
      console.log(chalk.cyan(`\nCurrent mode: ${processor.getMode().toUpperCase()}`));
      console.log(chalk.gray('  Use :mode <planning|agent|ask> to switch modes.'));
    }
    return;
  }
  if (command.startsWith(':pin')) {
    const parts = cmd.trim().split(' ');
    if (parts.length > 1) {
      const relPath = parts[1];
      const absPath = path.resolve(processor.config.rootDir, relPath);
      if (fs.existsSync(absPath)) {
        processor.pinFile(absPath);
        console.log(chalk.green(`\nFile pinned: ${relPath}`));
      } else {
        console.log(chalk.red(`\nFile not found: ${relPath}`));
      }
    } else {
      console.log(chalk.yellow('\nUsage: :pin <file_path>'));
    }
    return;
  }
  if (command.startsWith(':unpin')) {
    const parts = cmd.trim().split(' ');
    if (parts.length > 1) {
      const relPath = parts[1];
      const absPath = path.resolve(processor.config.rootDir, relPath);
      processor.unpinFile(absPath);
      console.log(chalk.green(`\nFile unpinned: ${relPath}`));
    } else {
      const pinned = processor.getPinnedFiles();
      if (pinned.length > 0) {
        console.log(chalk.cyan('\nPinned files:'));
        pinned.forEach(f => console.log(`  - ${path.relative(processor.config.rootDir, f)}`));
      } else {
        console.log(chalk.yellow('\nNo files pinned. Usage: :unpin <file_path>'));
      }
    }
    return;
  }
  if (command === ':clear') {
    console.clear();
    return;
  }
  if (command === ':context') {
    const preview = processor.getContextPreview();
    console.log('');
    console.log(
      `Context: messages=${preview.stats.messageCount}, fileTreeChars=${preview.stats.fileTreeChars}, promptChars=${preview.prompt.length}`
    );
    const show = await session.confirm('Show full prompt?', false);
    if (show) {
      console.log(preview.prompt);
    }
    return;
  }
  console.log(chalk.yellow('Unknown command. Type :help for commands.'));
}
