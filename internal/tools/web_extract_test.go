package tools

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"
)

func TestExtractHTMLPreservesDocumentation(t *testing.T) {
	base, _ := url.Parse("https://example.com/docs/start")
	source := `<nav>Ignore menu</nav><main><h2>Install</h2><p>Use <strong>Go</strong> &amp; tools &#x2713;.</p><a href="../api?q=a&amp;x=b">API</a><ul><li>First</li><li>Second</li></ul><pre><code>if a &lt; b {
  print(a)
}</code></pre><script>secretScript()</script><div hidden>hidden text</div></main>`
	got, err := extractHTML(source, base, true)
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"## Install", "**Go**", "✓", "https://example.com/api?q=a&x=b", "- First", "- Second", "if a < b {\n  print(a)\n}"} {
		if !strings.Contains(got, want) {
			t.Errorf("missing %q in %s", want, got)
		}
	}
	for _, bad := range []string{"Ignore menu", "secretScript", "hidden text"} {
		if strings.Contains(got, bad) {
			t.Errorf("exposed %s", bad)
		}
	}
}
func TestExtractHTMLDoesNotPreserveExecutableLinks(t *testing.T) {
	got, _ := extractHTML(`<p><a href="javascript:alert(1)">Read me</a><code>x</code></p>`, nil, true)
	if strings.Contains(got, "javascript:") || !strings.Contains(got, "Read me") {
		t.Fatal(got)
	}
	plain, _ := extractHTML(`<h1>Title</h1><p><b>Bold</b></p>`, nil, false)
	if strings.Contains(plain, "#") || strings.Contains(plain, "**") {
		t.Fatal(plain)
	}
}

type webTransport func(*http.Request) (*http.Response, error)

func (f webTransport) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func TestWebFetchFormatsAndPaginates(t *testing.T) {
	original := safeWebClient
	defer func() { safeWebClient = original }()
	body, media, status := "<h1>Docs</h1><a href='/api'>API</a>", "text/html", 200
	safeWebClient = &http.Client{Transport: webTransport(func(r *http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: status, Header: http.Header{"Content-Type": {media}}, Body: io.NopCloser(strings.NewReader(body)), Request: r}, nil
	})}
	run := func(offset int) string {
		t.Helper()
		r := webFetchTool.Execute(Context{}, json.RawMessage(fmt.Sprintf(`{"url":"https://8.8.8.8/docs","offset":%d}`, offset)))
		if r.Failed {
			t.Fatal(r.LLMResult)
		}
		return r.LLMResult
	}
	got := run(0)
	if !strings.Contains(got, "# Docs") || !strings.Contains(got, "https://8.8.8.8/api") || !strings.Contains(got, beginUntrusted) {
		t.Fatal(got)
	}
	body, media = strings.Repeat("✓", webMaxChars)+"tail", "text/plain"
	if !strings.Contains(run(0), "offset 20000") || !strings.Contains(run(webMaxChars), "tail") {
		t.Fatal("pagination failed")
	}
	status, media, body = 403, "text/plain", "untrusted error"
	result := webFetchTool.Execute(Context{}, json.RawMessage(`{"url":"https://8.8.8.8"}`))
	if !result.Failed || !strings.Contains(result.LLMResult, beginUntrusted) {
		t.Fatal("error content escaped untrusted boundary")
	}
	media = "image/png"
	if !webFetchTool.Execute(Context{}, json.RawMessage(`{"url":"https://8.8.8.8"}`)).Failed {
		t.Fatal("accepted binary")
	}
}
