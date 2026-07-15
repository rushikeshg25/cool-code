import { describe, it, expect, vi, afterEach } from 'vitest';
import * as dns from 'dns';
import { webFetch, htmlToText, isPrivateAddress } from './webFetchTool';

afterEach(() => {
  vi.unstubAllGlobals();
  vi.restoreAllMocks();
});

// Resolve any hostname to a fixed public IP so tests never hit the network.
function stubDnsPublic() {
  vi.spyOn(dns.promises, 'lookup').mockResolvedValue([
    { address: '93.184.216.34', family: 4 },
  ] as never);
}

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
    stubDnsPublic();
    vi.stubGlobal(
      'fetch',
      vi.fn().mockResolvedValue({
        ok: true,
        status: 200,
        headers: {
          get: (k: string) =>
            k.toLowerCase() === 'content-type' ? 'text/html; charset=utf-8' : null,
        },
        text: async () => '<html><body><h1>Docs</h1><p>content here</p></body></html>',
      })
    );
    const result = await webFetch({ url: 'https://example.com' });
    expect(result.LLMresult).toContain('Docs');
    expect(result.LLMresult).toContain('content here');
    expect(result.LLMresult).not.toContain('<h1>');
  });

  it('blocks the cloud metadata endpoint without fetching', async () => {
    const fetchMock = vi.fn();
    vi.stubGlobal('fetch', fetchMock);
    const result = await webFetch({ url: 'http://169.254.169.254/latest/meta-data/' });
    expect(result.DisplayResult).toBe('Blocked URL');
    expect(fetchMock).not.toHaveBeenCalled();
  });

  it('blocks loopback addresses without fetching', async () => {
    const fetchMock = vi.fn();
    vi.stubGlobal('fetch', fetchMock);
    const result = await webFetch({ url: 'http://127.0.0.1:8080/admin' });
    expect(result.DisplayResult).toBe('Blocked URL');
    expect(fetchMock).not.toHaveBeenCalled();
  });

  it('blocks a hostname that resolves to a private address', async () => {
    vi.spyOn(dns.promises, 'lookup').mockResolvedValue([
      { address: '10.0.0.5', family: 4 },
    ] as never);
    const fetchMock = vi.fn();
    vi.stubGlobal('fetch', fetchMock);
    const result = await webFetch({ url: 'https://internal.example' });
    expect(result.DisplayResult).toBe('Blocked URL');
    expect(fetchMock).not.toHaveBeenCalled();
  });
});

describe('isPrivateAddress', () => {
  it('flags loopback, private, link-local and metadata addresses', () => {
    for (const ip of ['127.0.0.1', '10.1.2.3', '172.16.0.1', '192.168.1.1', '169.254.169.254', '::1', 'fe80::1', 'fd00::1']) {
      expect(isPrivateAddress(ip)).toBe(true);
    }
  });
  it('allows public addresses', () => {
    for (const ip of ['93.184.216.34', '8.8.8.8', '2606:2800:220:1:248:1893:25c8:1946']) {
      expect(isPrivateAddress(ip)).toBe(false);
    }
  });
});
