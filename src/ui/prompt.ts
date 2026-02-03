import * as fs from 'fs';
import * as os from 'os';
import * as path from 'path';
import readline from 'readline';

export interface PromptOptions {
  message: string;
  historyFile?: string;
  completions?: string[];
}

export interface PromptSession {
  ask(): Promise<string | null>;
  confirm(message: string, defaultValue?: boolean): Promise<boolean>;
  onInput(callback: (line: string) => void): void;
  pause(): void;
  resume(): void;
  close(): void;
}

const DEFAULT_HISTORY_FILE = path.join(os.homedir(), '.coolcode_history');
const MAX_HISTORY = 200;

export function createPromptSession(options: PromptOptions): PromptSession {
  const historyFile = options.historyFile || DEFAULT_HISTORY_FILE;
  const completions = options.completions || [];

  const rl = readline.createInterface({
    input: process.stdin,
    output: process.stdout,
    historySize: MAX_HISTORY,
    completer: (line: string) => {
      const pathHits = completePath(line);
      if (pathHits) {
        return pathHits;
      }
      const hits = completions.filter((c) => c.startsWith(line));
      return [hits.length ? hits : completions, line];
    },
  });

  loadHistory(rl, historyFile);

  rl.on('SIGINT', () => {
    rl.close();
    process.exit(0);
  });

  return {
    async ask() {
      const value = await question(rl, options.message);
      if (value === null) return null;
      const trimmed = value.trim();
      if (trimmed.length > 0) {
        appendHistory(historyFile, trimmed);
      }
      return trimmed;
    },
    async confirm(message: string, defaultValue: boolean = false) {
      const suffix = defaultValue ? ' [Y/n]' : ' [y/N]';
      const answer = await question(rl, `${message}${suffix} `);
      if (answer === null) return false;
      const normalized = answer.trim().toLowerCase();
      if (!normalized) return defaultValue;
      return normalized === 'y' || normalized === 'yes';
    },
    onInput(callback: (line: string) => void) {
      rl.on('line', (line) => {
        const trimmed = line.trim();
        if (trimmed) {
          appendHistory(historyFile, trimmed);
          callback(trimmed);
        }
      });
    },
    pause() {
      rl.pause();
    },
    resume() {
      rl.resume();
    },
    close() {
      rl.removeAllListeners('line');
      rl.close();
    },
  };
}

function question(rl: readline.Interface, message: string): Promise<string | null> {
  return new Promise((resolve) => {
    rl.question(message, (answer) => resolve(answer));
  });
}

function loadHistory(rl: readline.Interface, filePath: string) {
  if (!fs.existsSync(filePath)) return;
  try {
    const lines = fs
      .readFileSync(filePath, 'utf-8')
      .split('\n')
      .map((l) => l.trim())
      .filter(Boolean);
    (rl as any).history = lines.slice(-MAX_HISTORY).reverse();
  } catch {
    // ignore history errors
  }
}

function appendHistory(filePath: string, line: string) {
  try {
    fs.appendFileSync(filePath, `${line}\n`);
  } catch {
    // ignore history errors
  }
}

function completePath(line: string): [string[], string] | null {
  const parts = line.split(/\s+/);
  const token = parts[parts.length - 1];
  if (!token || (!token.includes('/') && !token.startsWith('.'))) {
    return null;
  }

  const baseDir =
    token === '.' || token.endsWith('/')
      ? token
      : path.dirname(token);
  const prefix =
    token.endsWith('/') ? '' : path.basename(token);

  const resolvedDir = path.resolve(process.cwd(), baseDir);
  if (!fs.existsSync(resolvedDir)) return null;

  try {
    const entries = fs.readdirSync(resolvedDir);
    const matches = entries
      .filter((entry) => entry.startsWith(prefix))
      .map((entry) => {
        const separator = baseDir.endsWith('/') || baseDir === '.' ? '' : '/';
        return baseDir === '.'
          ? entry
          : `${baseDir}${separator}${entry}`;
      });
    return [matches.length ? matches : entries, token];
  } catch {
    return null;
  }
}
