package agent

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/rushikeshg25/cool-code/internal/llm"
	"github.com/rushikeshg25/cool-code/internal/types"
)

func TestCompactionIncludesDiscardedToolBoundary(t *testing.T) {
	c := &contextManager{}
	c.addUser("create five commits")
	c.addAssistant(llm.Message{Role: llm.RoleAssistant, ToolCalls: []llm.ToolCall{{ID: "call-1", Name: "shell_command", Arguments: json.RawMessage(`{"command":"exit 7"}`)}}})
	c.addToolResult(llm.ToolCall{ID: "call-1", Name: "shell_command"}, `{"exit_code":7}`)
	for i := 0; i < 19; i++ {
		c.addUser("later request")
	}
	transcript := c.transcriptForSummary(24000)
	for _, want := range []string{"create five commits", "shell_command", "exit 7", "exit_code", "call-1"} {
		if !strings.Contains(transcript, want) {
			t.Fatalf("missing %s: %s", want, transcript)
		}
	}
	c.applySummary("five commits still pending; command failed")
	if len(c.messages) != 19 || c.messages[0].Role != llm.RoleUser {
		t.Fatal("incorrect discard boundary")
	}
}

func TestCompactionOversizedResultDoesNotHideRequest(t *testing.T) {
	c := &contextManager{}
	c.addUser("use npm and commit separately")
	c.addToolResult(llm.ToolCall{ID: "large", Name: "shell_command"}, strings.Repeat("x", 200000))
	for i := 0; i < 20; i++ {
		c.addUser("recent")
	}
	got := c.transcriptForSummary(24000)
	if !strings.Contains(got, "use npm and commit separately") || !strings.Contains(got, "TRUNCATED") || estimateTokens(got) > 24000 {
		t.Fatal("lost requirements or exceeded budget")
	}
}

func TestCompactionHandoffSurvivesRestoreAndSecondCompaction(t *testing.T) {
	provider := &fakeProvider{responses: []llm.Message{{Text: "User requirements: five commits pending; npm required"}, {Text: "User requirements: five commits pending; npm required; tests failed"}}}
	p := newTestProcessor(t, t.TempDir(), provider, types.ModeAgent)
	for i := 0; i < 25; i++ {
		p.ctxMgr.addUser("five commits; npm")
	}
	p.Compact(context.Background())
	messages, summary, pinned := p.ctxMgr.sessionState()
	restored := newTestProcessor(t, t.TempDir(), provider, types.ModeAgent)
	restored.ctxMgr.restore(messages, summary, pinned)
	for i := 0; i < 10; i++ {
		restored.ctxMgr.addUser("tests failed")
	}
	restored.Compact(context.Background())
	if !strings.Contains(provider.lastReq.Messages[0].Text, "five commits pending") || provider.lastReq.System != compactionPrompt {
		t.Fatal("prior obligations lost")
	}
	if !strings.Contains(restored.ctxMgr.currentSummary(), "tests failed") {
		t.Fatal("new evidence lost")
	}
}
