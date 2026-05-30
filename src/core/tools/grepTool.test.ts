import { describe, it, expect, beforeEach, afterEach } from 'vitest';
import * as fs from 'fs';
import * as os from 'os';
import * as path from 'path';
import { grepTool } from './grepTool';

let dir: string;

beforeEach(() => {
  dir = fs.mkdtempSync(path.join(os.tmpdir(), 'grep-test-'));
  fs.writeFileSync(path.join(dir, 'a.ts'), 'const needle = 1;\nother line\n');
  fs.mkdirSync(path.join(dir, 'node_modules', 'pkg'), { recursive: true });
  fs.writeFileSync(
    path.join(dir, 'node_modules', 'pkg', 'index.js'),
    'const needle = 2;\n'
  );
});

afterEach(() => {
  fs.rmSync(dir, { recursive: true, force: true });
});

describe('grepTool', () => {
  it('returns a clean error for an invalid regex instead of throwing', async () => {
    const result = await grepTool({ pattern: '(', path: dir });
    expect(result.DisplayResult).toBe('Invalid pattern');
    expect(result.LLMresult).toMatch(/Invalid search pattern/);
  });

  it('finds matches in source files', async () => {
    const result = await grepTool({ pattern: 'needle', path: dir });
    expect(result.LLMresult).toContain('a.ts');
  });

  it('ignores node_modules', async () => {
    const result = await grepTool({ pattern: 'needle', path: dir });
    expect(result.LLMresult).not.toContain('node_modules');
  });
});
