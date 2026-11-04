package agent

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/rushikeshg25/cool-code/internal/config"
	"github.com/rushikeshg25/cool-code/internal/llm"
	"github.com/rushikeshg25/cool-code/internal/types"
)

// Opt in explicitly: this test makes billable requests using existing proxy
// settings, but changes neither Coolcode settings nor the proxy configuration.
func TestReliabilityLiveProxy(t *testing.T) {
	if os.Getenv("COOLCODE_LIVE_PROXY_TEST") != "1" {
		t.Skip("set COOLCODE_LIVE_PROXY_TEST=1 to use the existing CLI proxy")
	}
	cfg := config.Load(t.TempDir())
	if llm.ResolveProvider(cfg.LLM) != "cliproxy" {
		t.Fatal("live test requires configured CLI proxy")
	}
	cfg.LLM.Model = "gpt-6-luna"
	cfg.LLM.ReasoningEffort = "low"
	provider, err := llm.New(cfg.LLM)
	if err != nil {
		t.Fatal("could not initialize configured proxy provider")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 180*time.Second)
	defer cancel()
	t.Run("silent_failure", func(t *testing.T) {
		p := newTestProcessor(t, t.TempDir(), provider, types.ModeAgent)
		result, err := p.ProcessQuery(ctx, "Harness test: use shell_command exactly once with command 'exit 7'. Do not retry. Report the exit code and whether the command passed or failed.", nil)
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(result, "7") || !strings.Contains(strings.ToLower(result), "fail") {
			t.Fatalf("failure not recognized: %s", result)
		}
		found := false
		for _, m := range p.ctxMgr.snapshotMessages() {
			if m.ToolName == "shell_command" && strings.Contains(m.Text, `"exit_code":7`) {
				found = true
			}
		}
		if !found {
			t.Fatal("command was not executed")
		}
	})
	t.Run("ambiguous_edit", func(t *testing.T) {
		root := t.TempDir()
		path := filepath.Join(root, "sample.txt")
		if err := os.WriteFile(path, []byte("same\nsame\n"), 0600); err != nil {
			t.Fatal(err)
		}
		p := newTestProcessor(t, root, provider, types.ModeAgent)
		_, err := p.ProcessQuery(ctx, "Harness test: read "+path+" then call edit_file exactly once on it, oldString=same, newString=changed, expected_replacements=1. This intentionally tests rejection. Do not repair or retry. Report the result.", nil)
		if err != nil {
			t.Fatal(err)
		}
		raw, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		if string(raw) != "same\nsame\n" {
			t.Fatal("ambiguous edit mutated fixture")
		}
		found := false
		for _, m := range p.ctxMgr.snapshotMessages() {
			if m.ToolName == "edit_file" && strings.Contains(m.Text, "found 2") {
				found = true
			}
		}
		if !found {
			t.Fatal("ambiguous edit rejection was not exercised")
		}
	})
	t.Run("live_compaction", func(t *testing.T) {
		p := newTestProcessor(t, t.TempDir(), provider, types.ModeAgent)
		p.ctxMgr.addUser("Use npm. Create five separate commits. Both requirements remain pending.")
		p.ctxMgr.addAssistant(llm.Message{Role: llm.RoleAssistant, ToolCalls: []llm.ToolCall{toolCall("test", "shell_command", map[string]any{"command": "exit 7"})}})
		p.ctxMgr.addToolResult(llm.ToolCall{ID: "test", Name: "shell_command"}, `{"success":false,"exit_code":7}`)
		for i := 0; i < 24; i++ {
			p.ctxMgr.addUser("Continue within the original requirements.")
		}
		note := p.Compact(ctx)
		if !strings.Contains(note, "Compacted") {
			t.Fatal(note)
		}
		summary := p.ctxMgr.currentSummary()
		if !strings.Contains(strings.ToLower(summary), "npm") || !strings.Contains(strings.ToLower(summary), "commit") || !strings.Contains(summary, "7") {
			t.Fatalf("summary lost evidence: %s", summary)
		}
	})
}
