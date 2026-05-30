import type { ToolResult } from '../../types';
import { getErrorMessage } from '../utils';

export interface WebFetchOptions {
  url: string;
}

const MAX_CHARS = 20000;
const TIMEOUT_MS = 15000;

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
    const res = await fetch(url, {
      signal: controller.signal,
      headers: { 'User-Agent': 'cool-code/1.0 (+https://github.com/rushikeshg25/cool-code)' },
    });
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
        LLMresult: `HTTP ${res.status} for ${url.href}\n${truncated}`,
      };
    }
    return {
      DisplayResult: `Fetched ${url.hostname}`,
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
