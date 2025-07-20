import { showLanding } from './landing';
import { acceptQuery } from './query';

import { PrismaClient } from '@prisma/client';

export async function runCli(prisma: PrismaClient) {
  const rootDir = process.cwd();

  await showLanding();
  await acceptQuery(rootDir, prisma);
}
