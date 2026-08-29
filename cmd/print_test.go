package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/rushikeshg25/cool-code/internal/types"
)

func TestPrintPromptPrefersArguments(t *testing.T) {
	got, err := printPrompt([]string{"list", "the", "go", "files"})
	if err != nil {
		t.Fatal(err)
	}
	if got != "list the go files" {
		t.Errorf("printPrompt = %q", got)
	}
}

// TestPrintPromptReadsPipedStdin covers the plumbing that did not exist at all
// before: the binary had no non-interactive flag and never read stdin, so
// nothing could script or CI the agent.
func TestPrintPromptReadsPipedStdin(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "prompt")
	if err := os.WriteFile(path, []byte("summarize the repo\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	f, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	prev := os.Stdin
	os.Stdin = f
	t.Cleanup(func() { os.Stdin = prev })

	got, err := printPrompt(nil)
	if err != nil {
		t.Fatal(err)
	}
	if got != "summarize the repo\n" {
		t.Errorf("printPrompt from stdin = %q", got)
	}
}

func TestParseMode(t *testing.T) {
	cases := map[string]types.AgentMode{
		"plan":  types.ModePlan,
		"AGENT": types.ModeAgent,
		" ask ": types.ModeAsk,
	}
	for in, want := range cases {
		got, ok := parseMode(in)
		if !ok || got != want {
			t.Errorf("parseMode(%q) = %v, %v", in, got, ok)
		}
	}
	if _, ok := parseMode("supervisor"); ok {
		t.Error("parseMode accepted an unknown mode")
	}
}

// TestPrintReporterKeepsStdoutClean checks that progress goes to stderr, so
// the result on stdout stays pipeable.
func TestPrintReporterKeepsStdoutClean(t *testing.T) {
	var buf writerRecorder
	r := &printReporter{verbose: true, out: &buf}
	r.Tool("read_file", "Reading main.go", false)
	r.Tool("shell_command", "Command failed", true)

	if len(r.tools) != 2 {
		t.Errorf("tools recorded = %v", r.tools)
	}
	if !strings.Contains(buf.String(), "read_file") || !strings.Contains(buf.String(), "✗") {
		t.Errorf("reporter output = %q", buf.String())
	}

	quiet := &printReporter{verbose: false, out: &writerRecorder{}}
	quiet.Tool("read_file", "Reading main.go", false)
	if len(quiet.tools) != 1 {
		t.Error("quiet reporter should still record tool names for --json")
	}
}

type writerRecorder struct{ b []byte }

func (w *writerRecorder) Write(p []byte) (int, error) {
	w.b = append(w.b, p...)
	return len(p), nil
}
func (w *writerRecorder) String() string { return string(w.b) }
