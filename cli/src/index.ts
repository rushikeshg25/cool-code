#!/usr/bin/env node

import { program } from 'commander';
import dotenv from 'dotenv';
import { runCli } from './ui';

async function main() {
  dotenv.config({
    quiet: true,
  });
  program
    .name('ai-db-cli')
    .description('AI Database CLI Tool')
    .version('1.0.0');

  program.action(async () => {
    await runCli();
  });

  program.parse();
}

main().catch(console.error);
