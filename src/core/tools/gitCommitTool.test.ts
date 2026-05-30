import { describe, it, expect, vi, beforeEach } from 'vitest';
import * as path from 'path';

const execCommand = vi.fn();
vi.mock('./shellTool', () => ({
  execCommand: (...args: unknown[]) => execCommand(...args),
}));

import { gitCommit } from './gitCommitTool';

const root = path.resolve('/tmp/repo');

beforeEach(() => {
  execCommand.mockReset();
  execCommand.mockResolvedValue({ success: true, stdout: 'ok', stderr: '' });
});

describe('gitCommit', () => {
  it('single-quote escapes staged file paths to prevent injection', async () => {
    const evil = path.join(root, "a'; rm -rf ~ #.ts");
    await gitCommit({ message: 'msg', files: [evil] }, root);

    const addCall = execCommand.mock.calls.find((c) =>
      (c[0] as { command: string }).command.startsWith('git add')
    );
    expect(addCall).toBeTruthy();
    const cmd = (addCall![0] as { command: string }).command;
    // The path is wrapped in single quotes with the embedded quote escaped,
    // so the shell sees one literal argument and cannot execute `rm`.
    expect(cmd).toContain(`'"'"'`);
    expect(cmd).not.toMatch(/git add -- "/);
  });

  it('rejects an empty commit message', async () => {
    const result = await gitCommit({ message: '   ' }, root);
    expect(result.LLMresult).toMatch(/required/);
    expect(execCommand).not.toHaveBeenCalled();
  });
});
