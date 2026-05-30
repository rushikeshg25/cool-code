import { describe, it, expect, vi, afterEach } from 'vitest';
import { webFetch, htmlToText } from './webFetchTool';

afterEach(() => {
  vi.unstubAllGlobals();
});

describe('htmlToText', () => {
  it('drops scripts/styles/tags and decodes entities', () => {
    const html =
      '<html><head><style>.x{}</style></head><body><script>evil()</script><p>Hello &amp; welcome</p></body></html>';
    const text = htmlToText(html);
    expect(text).toContain('Hello & welcome');
    expect(text).not.toContain('evil()');
    expect(text).not.toContain('<p>');
  });
});

describe('webFetch', () => {
  it('rejects non-http(s) URLs without fetching', async () => {
    const fetchMock = vi.fn();
    vi.stubGlobal('fetch', fetchMock);
    const result = await webFetch({ url: 'file:///etc/passwd' });
    expect(result.DisplayResult).toBe('Invalid URL');
    expect(fetchMock).not.toHaveBeenCalled();
  });

  it('returns stripped text for an HTML response', async () => {
    vi.stubGlobal(
      'fetch',
      vi.fn().mockResolvedValue({
        ok: true,
        status: 200,
        headers: { get: () => 'text/html; charset=utf-8' },
        text: async () => '<html><body><h1>Docs</h1><p>content here</p></body></html>',
      })
    );
    const result = await webFetch({ url: 'https://example.com' });
    expect(result.LLMresult).toContain('Docs');
    expect(result.LLMresult).toContain('content here');
    expect(result.LLMresult).not.toContain('<h1>');
  });
});
