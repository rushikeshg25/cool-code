import type { ToolResult } from '../../types';
import { getErrorMessage } from '../utils';
import { htmlToText } from './webFetchTool';

export interface WebSearchOptions {
  query: string;
}

const TIMEOUT_MS = 15000;
const MAX_RESULTS = 8;

export interface SearchResult {
  title: string;
  url: string;
  snippet: string;
}

// Parses the DuckDuckGo HTML results page. Each result is an <a class="result__a">
// (title + link) with a sibling <a class="result__snippet"> (description).
export function parseDuckDuckGoHtml(html: string): SearchResult[] {
  const results: SearchResult[] = [];
  const linkRe =
    /<a[^>]*class="result__a"[^>]*href="([^"]+)"[^>]*>([\s\S]*?)<\/a>/g;
  const snippetRe =
    /<a[^>]*class="result__snippet"[^>]*>([\s\S]*?)<\/a>/g;

  const snippets: string[] = [];
  let sm: RegExpExecArray | null;
  while ((sm = snippetRe.exec(html)) !== null) {
    snippets.push(htmlToText(sm[1]));
  }

  let lm: RegExpExecArray | null;
  let i = 0;
  while ((lm = linkRe.exec(html)) !== null && results.length < MAX_RESULTS) {
    let url = lm[1];
    // DuckDuckGo wraps links as /l/?uddg=<encoded-url>; unwrap when present.
    const uddg = url.match(/[?&]uddg=([^&]+)/);
    if (uddg) {
      try {
        url = decodeURIComponent(uddg[1]);
      } catch {
        /* keep raw */
      }
    }
    results.push({
      title: htmlToText(lm[2]),
      url,
      snippet: snippets[i] ?? '',
    });
    i++;
  }
  return results;
}

export async function webSearch(options: WebSearchOptions): Promise<ToolResult> {
  if (!options.query || options.query.trim() === '') {
    return { DisplayResult: 'Fixing Issues', LLMresult: 'query is required.' };
  }

  const endpoint = `https://html.duckduckgo.com/html/?q=${encodeURIComponent(
    options.query
  )}`;
  const controller = new AbortController();
  const timer = setTimeout(() => controller.abort(), TIMEOUT_MS);
  try {
    const res = await fetch(endpoint, {
      signal: controller.signal,
      headers: { 'User-Agent': 'Mozilla/5.0 (compatible; cool-code/1.0)' },
    });
    const html = await res.text();
    const results = parseDuckDuckGoHtml(html);
    if (results.length === 0) {
      return {
        DisplayResult: 'No results',
        LLMresult: `No search results for "${options.query}".`,
      };
    }
    const formatted = results
      .map((r, idx) => `${idx + 1}. ${r.title}\n   ${r.url}\n   ${r.snippet}`)
      .join('\n');
    return {
      DisplayResult: `Found ${results.length} result(s)`,
      LLMresult: formatted,
    };
  } catch (error) {
    return {
      DisplayResult: 'Search failed',
      LLMresult: `Error searching for "${options.query}": ${getErrorMessage(error)}`,
    };
  } finally {
    clearTimeout(timer);
  }
}
