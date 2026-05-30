import { describe, it, expect } from 'vitest';
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
