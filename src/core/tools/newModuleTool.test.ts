import { describe, it, expect, beforeEach, afterEach } from 'vitest';
import * as fs from 'fs';
import * as os from 'os';
import * as path from 'path';
import { newModule } from './newModuleTool';

let dir: string;

beforeEach(() => {
  dir = fs.mkdtempSync(path.join(os.tmpdir(), 'newmodule-test-'));
});

afterEach(() => {
  fs.rmSync(dir, { recursive: true, force: true });
});

describe('newModule', () => {
  it('rejects a module name that escapes the project root', () => {
    const result = newModule({ moduleName: '../../evil' }, dir);
    expect(result.LLMresult).toMatch(/within project root/);
    // Nothing should have been created outside the root.
    expect(fs.existsSync(path.join(dir, '..', 'evil'))).toBe(false);
  });

  it('creates a module inside the root', () => {
    const result = newModule({ moduleName: 'widget' }, dir);
    expect(result.DisplayResult).toBe('Module created');
    expect(fs.existsSync(path.join(dir, 'src', 'widget', 'widget.ts'))).toBe(true);
  });
});
