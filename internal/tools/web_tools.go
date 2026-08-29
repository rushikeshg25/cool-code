package tools

import (
	"context"
	"encoding/json"
	"io"
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

var (
	scriptRe  = regexp.MustCompile(`(?is)<script.*?</script>`)
	styleRe   = regexp.MustCompile(`(?is)<style.*?</style>`)
	commentRe = regexp.MustCompile(`(?s)<!--.*?-->`)
	tagRe     = regexp.MustCompile(`<[^>]+>`)
	wsRe      = regexp.MustCompile(`[ \t]+`)
	blankRe   = regexp.MustCompile(`\n\s*\n\s*\n+`)
)

// htmlToText strips an HTML document down to readable text.
func htmlToText(html string) string {
	s := scriptRe.ReplaceAllString(html, " ")
	s = styleRe.ReplaceAllString(s, " ")
	s = commentRe.ReplaceAllString(s, " ")
	s = tagRe.ReplaceAllString(s, " ")
	replacer := strings.NewReplacer(
		"&nbsp;", " ", "&amp;", "&", "&lt;", "<", "&gt;", ">", "&quot;", "\"", "&#39;", "'",
	)
	s = replacer.Replace(s)
	s = wsRe.ReplaceAllString(s, " ")
	s = blankRe.ReplaceAllString(s, "\n\n")
	return strings.TrimSpace(s)
}

var webFetchTool = Tool{
	Name:        "web_fetch",
	Description: "Fetches the contents of an http(s) URL and returns it as readable text (HTML is stripped).",
	ReadOnly:    true,
	Schema: obj(map[string]any{
		"url": strProp("The absolute http(s) URL to fetch."),
	}, "url"),
	Execute: func(ctx Context, args json.RawMessage) types.ToolResult {
		var a struct {
			URL string `json:"url"`
		}
		if err := json.Unmarshal(args, &a); err != nil {
			return fail("Invalid arguments", err.Error())
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
		raw, _ := io.ReadAll(io.LimitReader(res.Body, 4*1024*1024))
		text := strings.TrimSpace(string(raw))
		if strings.Contains(res.Header.Get("Content-Type"), "html") {
			text = htmlToText(string(raw))
		}
		if len(text) > webMaxChars {
			text = text[:webMaxChars] + "\n... (truncated)"
		}
		if res.StatusCode >= 400 {
			return fail("Fetch failed", "HTTP "+itoa(res.StatusCode)+" for "+u.String()+"\n"+text)
		}
		if text == "" {
			text = "(empty response)"
		}
		return types.ToolResult{
			Display:   "Fetched " + u.Hostname(),
			LLMResult: Untrusted("web page "+u.Hostname(), text),
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
