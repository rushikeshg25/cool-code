#!/usr/bin/env node

import { program } from 'commander';
import dotenv from 'dotenv';
import { PrismaClient } from '@prisma/client';
import { runCli } from './ui';

async function main() {
  dotenv.config();
  program
    .name('ai-db-cli')
    .description('AI Database CLI Tool')
    .version('1.0.0');

  const prisma = new PrismaClient();

  program.action(async () => {
    await runCli(prisma);
  });

  program.parse();
}

main().catch(console.error);
