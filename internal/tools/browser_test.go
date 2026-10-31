package tools

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

func TestBrowserCommandValidation(t *testing.T) {
	for _, a := range []browserArgs{{Action: "navigate", URL: "file:///etc/passwd"}, {Action: "navigate", URL: "javascript:alert(1)"}, {Action: "navigate", URL: "https://user:pass@example.com"}, {Action: "click", Ref: "getByRole('button')"}, {Action: "run-code"}, {Action: "press", Key: "Control+L"}} {
		if _, err := browserCommand(a); err == nil {
			t.Fatalf("accepted %#v", a)
		}
	}
	args, err := browserCommand(browserArgs{Action: "fill", Ref: "e12", Text: "--config=/tmp/evil"})
	if err != nil || strings.Join(args, "|") != "fill|--|e12|--config=/tmp/evil" {
		t.Fatalf("unsafe argv: %v %v", args, err)
	}
	r := RunReadOnly(Context{}, "browser", json.RawMessage(`{"action":"navigate","url":"http://localhost:3000"}`))
	if !r.Failed || DangerReason("browser", json.RawMessage(`{"action":"click","ref":"e1"}`)) == "" {
		t.Fatal("browser bypasses mode/approval checks")
	}
}
func TestBrowserIsLazyAndMissingDependencyIsActionable(t *testing.T) {
	b := &Browser{}
	if err := b.Close(); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", t.TempDir())
	_, err := b.run(context.Background(), browserArgs{Action: "navigate", URL: "https://example.com"})
	if err == nil || !strings.Contains(err.Error(), "npm install") || b.started || b.directory != "" {
		t.Fatalf("%v %#v", err, b)
	}
}
func TestBrowserOutputBound(t *testing.T) {
	var out browserOutput
	s := strings.Repeat("a", 2*1024*1024)
	n, err := out.Write([]byte(s))
	if err != nil || n != len(s) || out.Len() != 1024*1024 {
		t.Fatal("unbounded output")
	}
}

// Opt-in: requires a user-installed playwright-cli and Chrome, never installs them.
func TestBrowserLive(t *testing.T) {
	if os.Getenv("COOLCODE_BROWSER_TEST") != "1" {
		t.Skip("set COOLCODE_BROWSER_TEST=1 with Playwright CLI and Chrome installed")
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		_, _ = w.Write([]byte(`<main><h1>Browser fixture</h1><label>Name<input id="name"></label><button onclick="document.querySelector('#result').textContent=document.querySelector('#name').value; console.log('fixture clicked')">Apply</button><p id="result"></p><script>document.querySelector('h1').textContent='Rendered by JavaScript'</script></main>`))
	}))
	defer server.Close()
	b := &Browser{}
	defer func() {
		if err := b.Close(); err != nil {
			t.Error(err)
		}
	}()
	run := func(a browserArgs) string {
		t.Helper()
		out, err := b.run(context.Background(), a)
		if err != nil {
			t.Fatal(err)
		}
		return out
	}
	out := run(browserArgs{Action: "navigate", URL: server.URL})
	if !strings.Contains(out, "Rendered by JavaScript") {
		t.Fatal(out)
	}
	ref := func(kind string) string {
		t.Helper()
		match := regexp.MustCompile(kind + `[^\n]*\[ref=(e\d+)\]`).FindStringSubmatch(out)
		if len(match) != 2 {
			t.Fatalf("no %s ref: %s", kind, out)
		}
		return match[1]
	}
	input, button := ref("textbox"), ref("button")
	run(browserArgs{Action: "fill", Ref: input, Text: "--config=literal text"})
	out = run(browserArgs{Action: "click", Ref: button})
	if !strings.Contains(out, "--config=literal text") {
		t.Fatal(out)
	}
	if !strings.Contains(run(browserArgs{Action: "find", Text: "literal text"}), "literal text") {
		t.Fatal("find lost text")
	}
	if !strings.Contains(run(browserArgs{Action: "console"}), "fixture clicked") {
		t.Fatal("missing console output")
	}
	run(browserArgs{Action: "network"})
	directory := b.directory
	raw, err := os.ReadFile(filepath.Join(directory, "config.json"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(raw), `"isolated":true`) || !strings.Contains(string(raw), `"headless":true`) {
		t.Fatal(string(raw))
	}
	run(browserArgs{Action: "close"})
	if _, err := os.Stat(directory); !os.IsNotExist(err) {
		t.Fatalf("temporary directory remains: %v", err)
	}
}
