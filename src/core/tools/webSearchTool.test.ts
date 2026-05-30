import { describe, it, expect, vi, afterEach } from 'vitest';
import { webSearch, parseDuckDuckGoHtml } from './webSearchTool';

afterEach(() => {
  vi.unstubAllGlobals();
});

const SAMPLE_HTML = `
<div class="result">
  <a class="result__a" href="/l/?uddg=https%3A%2F%2Fexample.com%2Fdocs">Example Docs</a>
  <a class="result__snippet">The official docs for Example.</a>
</div>
<div class="result">
  <a class="result__a" href="https://foo.dev/page">Foo Page</a>
  <a class="result__snippet">All about Foo.</a>
</div>
`;

describe('parseDuckDuckGoHtml', () => {
  it('extracts titles, unwrapped urls, and snippets', () => {
    const results = parseDuckDuckGoHtml(SAMPLE_HTML);
    expect(results).toHaveLength(2);
    expect(results[0].title).toBe('Example Docs');
    expect(results[0].url).toBe('https://example.com/docs');
    expect(results[0].snippet).toBe('The official docs for Example.');
    expect(results[1].url).toBe('https://foo.dev/page');
  });
});

describe('webSearch', () => {
  it('formats parsed results', async () => {
    vi.stubGlobal(
      'fetch',
      vi.fn().mockResolvedValue({ ok: true, status: 200, text: async () => SAMPLE_HTML })
    );
    const result = await webSearch({ query: 'example' });
    expect(result.DisplayResult).toMatch(/Found 2/);
    expect(result.LLMresult).toContain('Example Docs');
    expect(result.LLMresult).toContain('https://example.com/docs');
  });

  it('requires a query', async () => {
    const result = await webSearch({ query: '   ' });
    expect(result.LLMresult).toMatch(/required/);
  });
});
