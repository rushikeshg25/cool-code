import { showLanding } from './landing';
import { acceptQuery } from './query';

export async function runCli() {
  await showLanding();
  await acceptQuery();
}
