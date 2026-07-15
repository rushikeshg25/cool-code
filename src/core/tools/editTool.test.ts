import { describe, it, expect, beforeEach, afterEach } from 'vitest';
import * as fs from 'fs';
import * as os from 'os';
import * as path from 'path';
import { editFile } from './editTool';

let root: string;
let outside: string;

beforeEach(() => {
  root = fs.mkdtempSync(path.join(os.tmpdir(), 'coolcode-root-'));
  outside = fs.mkdtempSync(path.join(os.tmpdir(), 'coolcode-outside-'));
});

afterEach(() => {
  fs.rmSync(root, { recursive: true, force: true });
  fs.rmSync(outside, { recursive: true, force: true });
});

describe('editFile confinement', () => {
  it('refuses to edit a file outside the project root and leaves it untouched', () => {
    const target = path.join(outside, 'rc');
    fs.writeFileSync(target, 'original');
    const result = editFile(
      {
        filePath: target,
        oldString: 'original',
        newString: 'pwned',
        expected_replacements: 1,
      },
      root
    );
    expect(result.LLMresult).toMatch(/within project root/);
    expect(fs.readFileSync(target, 'utf-8')).toBe('original');
  });

  it('edits a file inside the project root', () => {
    const target = path.join(root, 'file.txt');
    fs.writeFileSync(target, 'hello world');
    const result = editFile(
      {
        filePath: target,
        oldString: 'world',
        newString: 'there',
        expected_replacements: 1,
      },
      root
    );
    expect(fs.readFileSync(target, 'utf-8')).toBe('hello there');
  });
});
