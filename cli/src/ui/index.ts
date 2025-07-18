import { showLanding } from "./landing";
import { acceptQuery } from "./query";


export async function runCli() {
  const rootDir = process.cwd();

  await showLanding();
  await acceptQuery(rootDir);
}
