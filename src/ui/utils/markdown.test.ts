import { describe, it, expect } from 'vitest';
import { renderMarkdown } from './markdown';

// chalk styling may be disabled in the test env, so assert on the surviving
// text content rather than on ANSI codes.
describe('renderMarkdown', () => {
  it('preserves fenced code block contents (no placeholder mangling)', () => {
    const input = 'Here:\n```ts\nconst x = 1;\nfoo_bar();\n```\ndone';
    const out = renderMarkdown(input);
    expect(out).toContain('const x = 1;');
    expect(out).toContain('foo_bar();');
    // The internal placeholder must not leak into the output.
    expect(out).not.toContain('CODE_BLOCK');
  });

  it('handles multiple code blocks', () => {
    const input = '```\nblockA\n```\nmiddle\n```\nblockB\n```';
    const out = renderMarkdown(input);
    expect(out).toContain('blockA');
    expect(out).toContain('blockB');
    expect(out).toContain('middle');
  });
});
