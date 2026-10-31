package tools

import (
	"context"
	"encoding/json"
	"io"
	"mime"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/rushikeshg25/cool-code/internal/types"
)

const (
	webMaxChars   = 20000
	webTimeout    = 15 * time.Second
	maxSearchHits = 8
)

// htmlToText is used for short search-result fragments.
func htmlToText(source string) string {
	text, err := extractHTML(source, nil, false)
	if err != nil {
		return ""
	}
	return strings.Join(strings.Fields(text), " ")
}

var webFetchTool = Tool{
	Name:        "web_fetch",
	Description: "Reads an HTTPS page as Markdown or text without JavaScript. Preserves documentation links and code. Use offset to read further chunks; use browser tools for JavaScript-rendered pages.",
	ReadOnly:    true,
	Schema: obj(map[string]any{
		"url":    strProp("The absolute HTTPS URL to fetch."),
		"format": map[string]any{"type": "string", "enum": []string{"markdown", "text"}, "description": "Output format; defaults to markdown."},
		"offset": numProp("Optional character offset for the next chunk, starting at zero."),
	}, "url"),
	Execute: func(ctx Context, args json.RawMessage) types.ToolResult {
		var a struct {
			URL    string `json:"url"`
			Format string `json:"format"`
			Offset int    `json:"offset"`
		}
		if err := json.Unmarshal(args, &a); err != nil {
			return fail("Invalid arguments", err.Error())
		}
		if a.Offset < 0 || (a.Format != "" && a.Format != "markdown" && a.Format != "text") {
			return fail("Invalid arguments", "Use a nonnegative offset and format markdown or text.")
		}
		u, err := url.Parse(a.URL)
		if err != nil {
			return fail("Invalid URL", "Only absolute HTTPS URLs are allowed.")
		}
		ctxT, cancel := context.WithTimeout(ctx.Context(), webTimeout)
		defer cancel()
		if err := validateWebURL(ctxT, u); err != nil {
			return fail("Fetch blocked", err.Error())
		}
		req, _ := http.NewRequestWithContext(ctxT, http.MethodGet, u.String(), nil)
		req.Header.Set("User-Agent", "cool-code/2.0 (+https://github.com/rushikeshg25/cool-code)")
		res, err := safeWebClient.Do(req)
		if err != nil {
			return fail("Fetch failed", "Secure fetch failed: "+err.Error())
		}
		defer res.Body.Close()
		raw, err := io.ReadAll(io.LimitReader(res.Body, 4*1024*1024+1))
		if err != nil {
			return fail("Fetch failed", "Could not read the complete response.")
		}
		if len(raw) > 4*1024*1024 {
			return fail("Fetch failed", "Page exceeds the 4 MiB download limit.")
		}
		media, _, _ := mime.ParseMediaType(res.Header.Get("Content-Type"))
		if media == "" {
			media, _, _ = mime.ParseMediaType(http.DetectContentType(raw))
		}
		text := strings.TrimSpace(string(raw))
		if media == "text/html" || media == "application/xhtml+xml" {
			text, err = extractHTML(string(raw), res.Request.URL, a.Format != "text")
			if err != nil {
				return fail("Fetch failed", "Could not parse page HTML.")
			}
		} else if !strings.HasPrefix(media, "text/") && media != "application/json" && media != "application/xml" && !strings.HasSuffix(media, "+json") && !strings.HasSuffix(media, "+xml") {
			return fail("Unsupported content", "Use web_fetch for text, HTML, JSON, or XML; binary content is not supported.")
		}
		chars := []rune(text)
		if a.Offset > len(chars) {
			return fail("Invalid offset", "Page contains "+itoa(len(chars))+" characters.")
		}
		end := a.Offset + webMaxChars
		if end > len(chars) {
			end = len(chars)
		}
		text = string(chars[a.Offset:end])
		if end < len(chars) {
			text += "\n[More content: call web_fetch with offset " + itoa(end) + ".]"
		}

		if res.StatusCode >= 400 {
			return fail("Fetch failed", Untrusted("web error", "HTTP "+itoa(res.StatusCode)+" for "+u.String()+"\n"+text))
		}
		if text == "" {
			text = "(empty response)"
		}
		return types.ToolResult{
			Display:   "Fetched " + u.Hostname(),
			LLMResult: Untrusted("web page "+u.Hostname(), "Source: "+res.Request.URL.String()+"\n\n"+text),
		}
	},
}

var (
	linkRe    = regexp.MustCompile(`(?is)<a[^>]*class="result__a"[^>]*href="([^"]+)"[^>]*>(.*?)</a>`)
	snippetRe = regexp.MustCompile(`(?is)<a[^>]*class="result__snippet"[^>]*>(.*?)</a>`)
	uddgRe    = regexp.MustCompile(`[?&]uddg=([^&]+)`)
)

type searchResult struct {
	title   string
	url     string
	snippet string
}

func parseDuckDuckGo(html string) []searchResult {
	var snippets []string
	for _, m := range snippetRe.FindAllStringSubmatch(html, -1) {
		snippets = append(snippets, htmlToText(m[1]))
	}
	var out []searchResult
	for i, m := range linkRe.FindAllStringSubmatch(html, -1) {
		if len(out) >= maxSearchHits {
			break
		}
		link := m[1]
		if uddg := uddgRe.FindStringSubmatch(link); uddg != nil {
			if decoded, err := url.QueryUnescape(uddg[1]); err == nil {
				link = decoded
			}
		}
		snippet := ""
		if i < len(snippets) {
			snippet = snippets[i]
		}
		out = append(out, searchResult{title: htmlToText(m[2]), url: link, snippet: snippet})
	}
	return out
}

var webSearchTool = Tool{
	Name:        "web_search",
	Description: "Searches the web and returns result titles, URLs, and snippets. Follow up with web_fetch for details.",
	ReadOnly:    true,
	Schema: obj(map[string]any{
		"query": strProp("The search query."),
	}, "query"),
	Execute: func(ctx Context, args json.RawMessage) types.ToolResult {
		var a struct {
			Query string `json:"query"`
		}
		if err := json.Unmarshal(args, &a); err != nil {
			return fail("Invalid arguments", err.Error())
		}
		if strings.TrimSpace(a.Query) == "" {
			return fail("Invalid arguments", "query is required.")
		}
		endpoint := "https://html.duckduckgo.com/html/?q=" + url.QueryEscape(a.Query)
		ctxT, cancel := context.WithTimeout(ctx.Context(), webTimeout)
		defer cancel()
		req, _ := http.NewRequestWithContext(ctxT, http.MethodGet, endpoint, nil)
		req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; cool-code/2.0)")
		res, err := safeWebClient.Do(req)
		if err != nil {
			return fail("Search failed", "Error searching for \""+a.Query+"\": "+err.Error())
		}
		defer res.Body.Close()
		raw, _ := io.ReadAll(io.LimitReader(res.Body, 4*1024*1024))
		results := parseDuckDuckGo(string(raw))
		if len(results) == 0 {
			return fail("No results", "No search results for \""+a.Query+"\".")
		}
		var b strings.Builder
		for i, r := range results {
			b.WriteString(itoa(i+1) + ". " + r.title + "\n   " + r.url + "\n   " + r.snippet + "\n")
		}
		return types.ToolResult{
			Display:   "Found " + itoa(len(results)) + " result(s)",
			LLMResult: Untrusted("web search results", strings.TrimRight(b.String(), "\n")),
		}
	},
}
