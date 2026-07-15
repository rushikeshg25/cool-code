import { describe, it, expect, beforeEach, afterEach } from 'vitest';
import * as fs from 'fs';
import * as os from 'os';
import * as path from 'path';
import {
  saveSession,
  loadSession,
  listSessions,
  latestSession,
  newSessionId,
  type SessionData,
} from './session';

let home: string;
let originalHome: string | undefined;

function makeSession(over: Partial<SessionData> = {}): SessionData {
  return {
    id: newSessionId(),
    cwd: '/proj/a',
    updatedAt: new Date().toISOString(),
    mode: 'agent',
    conversations: [{ role: 'user', content: 'hi' }],
    summary: null,
    pinnedFiles: [],
    ...over,
  };
}

beforeEach(() => {
  home = fs.mkdtempSync(path.join(os.tmpdir(), 'session-home-'));
  originalHome = process.env.HOME;
  process.env.HOME = home; // os.homedir() honors $HOME on posix
});

afterEach(() => {
  if (originalHome === undefined) delete process.env.HOME;
  else process.env.HOME = originalHome;
  fs.rmSync(home, { recursive: true, force: true });
});

describe('session persistence', () => {
  it('saves and loads a session round-trip', () => {
    const s = makeSession({
      conversations: [
        { role: 'user', content: 'q1' },
        { role: 'model', content: 'a1' },
      ],
      summary: 'a summary',
      pinnedFiles: ['/proj/a/x.ts'],
    });
    saveSession(s);
    const loaded = loadSession(s.id);
    expect(loaded).not.toBeNull();
    expect(loaded!.conversations).toHaveLength(2);
    expect(loaded!.summary).toBe('a summary');
    expect(loaded!.pinnedFiles).toEqual(['/proj/a/x.ts']);
  });

  it('lists sessions for a cwd newest-first and finds the latest', () => {
    saveSession(makeSession({ cwd: '/proj/a', updatedAt: '2026-01-01T00:00:00.000Z' }));
    const newer = makeSession({ cwd: '/proj/a', updatedAt: '2026-05-01T00:00:00.000Z' });
    saveSession(newer);
    saveSession(makeSession({ cwd: '/proj/other' }));

    const list = listSessions('/proj/a');
    expect(list).toHaveLength(2);
    expect(list[0].id).toBe(newer.id);
    expect(latestSession('/proj/a')!.id).toBe(newer.id);
  });

  it('returns null for an unknown id and empty list for an unknown cwd', () => {
    expect(loadSession('does-not-exist')).toBeNull();
    expect(listSessions('/nope')).toEqual([]);
  });

  it.skipIf(process.platform === 'win32')(
    'writes the session file 0600 and its directory 0700',
    () => {
      const s = makeSession();
      saveSession(s);
      const dir = path.join(home, '.coolcode', 'sessions');
      const file = path.join(dir, `${s.id}.json`);
      expect(fs.statSync(file).mode & 0o777).toBe(0o600);
      expect(fs.statSync(dir).mode & 0o777).toBe(0o700);
    }
  );
});
