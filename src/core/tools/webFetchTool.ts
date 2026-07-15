import * as dns from 'dns';
import * as net from 'net';
import type { ToolResult } from '../../types';
import { getErrorMessage } from '../utils';

export interface WebFetchOptions {
  url: string;
}

const MAX_CHARS = 20000;
const TIMEOUT_MS = 15000;
const MAX_REDIRECTS = 5;

function ipv4IsPrivate(ip: string): boolean {
  const parts = ip.split('.').map(Number);
  if (parts.length !== 4 || parts.some((p) => Number.isNaN(p))) return false;
  const [a, b] = parts;
  if (a === 0 || a === 127) return true; // this-network / loopback
  if (a === 10) return true; // private
  if (a === 169 && b === 254) return true; // link-local (incl. 169.254.169.254 metadata)
  if (a === 172 && b >= 16 && b <= 31) return true; // private
  if (a === 192 && b === 168) return true; // private
  return false;
}

// True for loopback, private, link-local and unique-local addresses — the ones
// an SSRF payload would target (internal services, cloud metadata endpoints).
export function isPrivateAddress(ip: string): boolean {
  const mapped = ip.match(/^::ffff:(\d+\.\d+\.\d+\.\d+)$/i);
  if (mapped) return ipv4IsPrivate(mapped[1]);
  if (net.isIPv4(ip)) return ipv4IsPrivate(ip);
  const lower = ip.toLowerCase();
  if (lower === '::1' || lower === '::') return true; // loopback / unspecified
  if (lower.startsWith('fe80')) return true; // link-local
  if (lower.startsWith('fc') || lower.startsWith('fd')) return true; // unique-local fc00::/7
  return false;
}

// Rejects URLs that resolve to a private/loopback/link-local address. Returns
// an error string when the URL should be blocked, otherwise null.
export async function assertUrlAllowed(url: URL): Promise<string | null> {
  const hostname = url.hostname.replace(/^\[|\]$/g, ''); // strip IPv6 brackets
  let addresses: string[];
  if (net.isIP(hostname)) {
    addresses = [hostname];
  } else {
    try {
      const looked = await dns.promises.lookup(hostname, { all: true });
      addresses = looked.map((a) => a.address);
    } catch {
      return `Could not resolve host: ${hostname}`;
    }
  }
  if (addresses.some(isPrivateAddress)) {
    return `Blocked request to a private or loopback address (${hostname}).`;
  }
  return null;
}

// Strips a raw HTML document down to readable text: removes script/style
// blocks, drops tags, decodes a few common entities, collapses whitespace.
export function htmlToText(html: string): string {
  return html
    .replace(/<script[\s\S]*?<\/script>/gi, ' ')
    .replace(/<style[\s\S]*?<\/style>/gi, ' ')
    .replace(/<!--[\s\S]*?-->/g, ' ')
    .replace(/<[^>]+>/g, ' ')
    .replace(/&nbsp;/g, ' ')
    .replace(/&amp;/g, '&')
    .replace(/&lt;/g, '<')
    .replace(/&gt;/g, '>')
    .replace(/&quot;/g, '"')
    .replace(/&#39;/g, "'")
    .replace(/[ \t]+/g, ' ')
    .replace(/\n\s*\n\s*\n+/g, '\n\n')
    .trim();
}

export async function webFetch(options: WebFetchOptions): Promise<ToolResult> {
  let url: URL;
  try {
    url = new URL(options.url);
  } catch {
    return { DisplayResult: 'Invalid URL', LLMresult: `Invalid URL: ${options.url}` };
  }
  if (url.protocol !== 'http:' && url.protocol !== 'https:') {
    return {
      DisplayResult: 'Invalid URL',
      LLMresult: 'Only http and https URLs are allowed.',
    };
  }

  const controller = new AbortController();
  const timer = setTimeout(() => controller.abort(), TIMEOUT_MS);
  try {
    // Follow redirects manually so every hop is re-validated — a public URL
    // must not be able to 302 into an internal/metadata endpoint.
    let current = url;
    let res: Response;
    for (let hop = 0; ; hop++) {
      const blocked = await assertUrlAllowed(current);
      if (blocked) {
        return { DisplayResult: 'Blocked URL', LLMresult: blocked };
      }
      res = await fetch(current, {
        signal: controller.signal,
        redirect: 'manual',
        headers: { 'User-Agent': 'cool-code/1.0 (+https://github.com/rushikeshg25/cool-code)' },
      });
      const location = res.headers.get('location');
      if (res.status >= 300 && res.status < 400 && location) {
        if (hop >= MAX_REDIRECTS) {
          return {
            DisplayResult: 'Too many redirects',
            LLMresult: `Too many redirects fetching ${url.href}`,
          };
        }
        current = new URL(location, current);
        continue;
      }
      break;
    }
    const contentType = res.headers.get('content-type') || '';
    const raw = await res.text();
    const text = contentType.includes('html') ? htmlToText(raw) : raw.trim();
    const truncated =
      text.length > MAX_CHARS
        ? text.slice(0, MAX_CHARS) + '\n... (truncated)'
        : text;

    if (!res.ok) {
      return {
        DisplayResult: `Fetch failed (${res.status})`,
        LLMresult: `HTTP ${res.status} for ${current.href}\n${truncated}`,
      };
    }
    return {
      DisplayResult: `Fetched ${current.hostname}`,
      LLMresult: truncated || '(empty response)',
    };
  } catch (error) {
    return {
      DisplayResult: 'Fetch failed',
      LLMresult: `Error fetching ${url.href}: ${getErrorMessage(error)}`,
    };
  } finally {
    clearTimeout(timer);
  }
}
