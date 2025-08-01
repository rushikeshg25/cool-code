#!/usr/bin/env node

import { program } from 'commander';
import dotenv from 'dotenv';
import { runCli } from './ui';

async function main() {
  dotenv.config({
    quiet: true,
  });
  program
    .name('cool-code')
    .description('Cli coding Ai agent like Claude Code and Gemini Cli')
    .version('1.0.0');

  program.action(async () => {
    await runCli();
  });

  program.parse();
}

main().catch(console.error);
