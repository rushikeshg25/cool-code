import { describe, it, expect, beforeEach, afterEach } from 'vitest';
import * as fs from 'fs';
import * as os from 'os';
import * as path from 'path';
import { replaceInFiles } from './replaceInFilesTool';

let dir: string;

beforeEach(() => {
  dir = fs.mkdtempSync(path.join(os.tmpdir(), 'replace-test-'));
  fs.writeFileSync(path.join(dir, 'a.txt'), 'foo foo bar');
});

afterEach(() => {
  fs.rmSync(dir, { recursive: true, force: true });
});

describe('replaceInFiles', () => {
  it('defaults to a dry run that does not modify files', () => {
    const result = replaceInFiles(
      { pattern: 'foo', replacement: 'baz' },
      dir
    );
    expect(result.DisplayResult).toMatch(/Dry run/);
    expect(fs.readFileSync(path.join(dir, 'a.txt'), 'utf-8')).toBe('foo foo bar');
  });

  it('writes literal replacements when dryRun is false', () => {
    replaceInFiles(
      { pattern: 'foo', replacement: 'baz', dryRun: false },
      dir
    );
    expect(fs.readFileSync(path.join(dir, 'a.txt'), 'utf-8')).toBe('baz baz bar');
  });

  it('returns a clean error for an invalid regex instead of throwing', () => {
    const result = replaceInFiles(
      { pattern: '(', replacement: 'X', useRegex: true, dryRun: false },
      dir
    );
    expect(result.DisplayResult).toBe('Invalid pattern');
    expect(result.LLMresult).toMatch(/Invalid regex/);
    // File must be untouched.
    expect(fs.readFileSync(path.join(dir, 'a.txt'), 'utf-8')).toBe('foo foo bar');
  });

  it('supports regex replacements', () => {
    const result = replaceInFiles(
      { pattern: 'f.o', replacement: 'X', useRegex: true, dryRun: false },
      dir
    );
    expect(result.LLMresult).toContain('"totalReplacements": 2');
    expect(fs.readFileSync(path.join(dir, 'a.txt'), 'utf-8')).toBe('X X bar');
  });
});
