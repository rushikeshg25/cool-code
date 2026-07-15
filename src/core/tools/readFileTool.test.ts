import { describe, it, expect, beforeEach, afterEach } from 'vitest';
import * as fs from 'fs';
import * as os from 'os';
import * as path from 'path';
import { readFile } from './readFileTool';

let root: string;
let outsideFile: string;

beforeEach(() => {
  root = fs.mkdtempSync(path.join(os.tmpdir(), 'coolcode-root-'));
  outsideFile = fs.mkdtempSync(path.join(os.tmpdir(), 'coolcode-secret-'));
  fs.writeFileSync(path.join(outsideFile, 'secret.txt'), 'top secret');
  fs.writeFileSync(path.join(root, 'inside.txt'), 'ok content');
});

afterEach(() => {
  fs.rmSync(root, { recursive: true, force: true });
  fs.rmSync(outsideFile, { recursive: true, force: true });
});

describe('readFile confinement', () => {
  it('refuses to read a file outside the project root', async () => {
    const result = await readFile(
      { absolutePath: path.join(outsideFile, 'secret.txt') },
      root
    );
    expect(result.LLMresult).toMatch(/within project root/);
    expect(result.LLMresult).not.toContain('top secret');
  });

  it('reads a file inside the project root', async () => {
    const result = await readFile(
      { absolutePath: path.join(root, 'inside.txt') },
      root
    );
    expect(result.LLMresult).toBe('ok content');
  });
});
