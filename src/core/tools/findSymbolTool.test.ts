import { describe, it, expect, vi, beforeEach } from 'vitest';

const execCommand = vi.fn();
vi.mock('./shellTool', () => ({
  execCommand: (...args: unknown[]) => execCommand(...args),
}));

import { findSymbol } from './findSymbolTool';

beforeEach(() => {
  execCommand.mockReset();
  execCommand.mockResolvedValue({ success: true, stdout: '', stderr: '' });
});

describe('findSymbol', () => {
  it('single-quote escapes the pattern so it cannot inject shell commands', async () => {
    await findSymbol({ pattern: 'foo"; echo $(whoami)' }, '/tmp/repo');
    const cmd = (execCommand.mock.calls[0][0] as { command: string }).command;
    // The whole payload must sit inside a single-quoted argument (no single
    // quotes in it, so nothing escapes the quoting) — the shell sees one arg.
    expect(cmd).toContain(`'foo"; echo $(whoami)'`);
  });
});
