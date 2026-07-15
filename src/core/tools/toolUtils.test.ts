import { describe, it, expect, afterEach } from 'vitest';
import * as fs from 'fs';
import * as os from 'os';
import * as path from 'path';
import {
  ensureAbsoluteWithinRoot,
  shellEscapeSingleQuotes,
  toPascalCase,
} from './toolUtils';

const root = path.resolve('/tmp/project');

describe('ensureAbsoluteWithinRoot', () => {
  it('rejects relative paths', () => {
    expect(ensureAbsoluteWithinRoot('src/index.ts', root)).toMatch(
      /must be absolute/
    );
  });

  it('accepts a path inside the root', () => {
    expect(
      ensureAbsoluteWithinRoot(path.join(root, 'src/index.ts'), root)
    ).toBeNull();
  });

  it('rejects path traversal escaping the root', () => {
    expect(
      ensureAbsoluteWithinRoot(path.join(root, '../secret.txt'), root)
    ).toMatch(/within project root/);
  });

  it('rejects a sibling directory sharing a prefix', () => {
    expect(
      ensureAbsoluteWithinRoot('/tmp/project-evil/file.txt', root)
    ).toMatch(/within project root/);
  });

  describe('symlink escape', () => {
    let realRoot: string;
    let realOutside: string;

    afterEach(() => {
      if (realRoot) fs.rmSync(realRoot, { recursive: true, force: true });
      if (realOutside) fs.rmSync(realOutside, { recursive: true, force: true });
    });

    it('rejects a path that escapes the root through a symlinked directory', () => {
      realRoot = fs.mkdtempSync(path.join(os.tmpdir(), 'coolcode-root-'));
      realOutside = fs.mkdtempSync(path.join(os.tmpdir(), 'coolcode-out-'));
      // A symlink inside the root pointing outside it.
      const link = path.join(realRoot, 'escape');
      fs.symlinkSync(realOutside, link);
      expect(
        ensureAbsoluteWithinRoot(path.join(link, 'secret.txt'), realRoot)
      ).toMatch(/within project root/);
    });

    it('still accepts a real path inside the root', () => {
      realRoot = fs.mkdtempSync(path.join(os.tmpdir(), 'coolcode-root-'));
      expect(
        ensureAbsoluteWithinRoot(path.join(realRoot, 'src/index.ts'), realRoot)
      ).toBeNull();
    });
  });
});

describe('shellEscapeSingleQuotes', () => {
  it('escapes embedded single quotes for safe single-quoting', () => {
    // The shell-safe encoding of a single quote is '\'' -> '"'"'
    expect(shellEscapeSingleQuotes("it's")).toBe(`it'"'"'s`);
  });

  it('leaves strings without single quotes untouched', () => {
    expect(shellEscapeSingleQuotes('hello world')).toBe('hello world');
  });
});

describe('toPascalCase', () => {
  it('converts kebab and snake case to PascalCase', () => {
    expect(toPascalCase('my-cool_module')).toBe('MyCoolModule');
  });
});
