import * as fs from 'fs';
import * as os from 'os';
import * as path from 'path';
import { randomUUID } from 'crypto';
import type { Message } from './contextManager';
import type { AgentMode } from '../types';

export interface SessionData {
  id: string;
  cwd: string;
  updatedAt: string;
  mode: AgentMode;
  conversations: Message[];
  summary: string | null;
  pinnedFiles: string[];
}

function sessionsDir(): string {
  return path.join(os.homedir(), '.coolcode', 'sessions');
}

function sessionPath(id: string): string {
  return path.join(sessionsDir(), `${id}.json`);
}

export function newSessionId(): string {
  return randomUUID();
}

export function saveSession(data: SessionData): void {
  const dir = sessionsDir();
  fs.mkdirSync(dir, { recursive: true });
  fs.writeFileSync(sessionPath(data.id), JSON.stringify(data, null, 2), 'utf-8');
}

export function loadSession(id: string): SessionData | null {
  try {
    const raw = fs.readFileSync(sessionPath(id), 'utf-8');
    return JSON.parse(raw) as SessionData;
  } catch {
    return null;
  }
}

// All sessions recorded for a given working directory, newest first.
export function listSessions(cwd: string): SessionData[] {
  let files: string[];
  try {
    files = fs.readdirSync(sessionsDir());
  } catch {
    return [];
  }
  const sessions: SessionData[] = [];
  for (const file of files) {
    if (!file.endsWith('.json')) continue;
    try {
      const data = JSON.parse(
        fs.readFileSync(path.join(sessionsDir(), file), 'utf-8')
      ) as SessionData;
      if (data.cwd === cwd) sessions.push(data);
    } catch {
      // skip corrupt session files
    }
  }
  return sessions.sort((a, b) => b.updatedAt.localeCompare(a.updatedAt));
}

export function latestSession(cwd: string): SessionData | null {
  return listSessions(cwd)[0] ?? null;
}
