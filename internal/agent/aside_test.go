package agent

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/rushikeshg25/cool-code/internal/llm"
	"github.com/rushikeshg25/cool-code/internal/types"
)

func TestAsideKeepsConversationAndToolsSeparate(t *testing.T) {
	provider := &fakeProvider{responses: []llm.Message{{Text: "Side answer", Usage: llm.Usage{Input: 20, Output: 4}}}}
	p := newTestProcessor(t, t.TempDir(), provider, types.ModePlan)
	p.ctxMgr.addUser("Main question")
	p.ctxMgr.addAssistant(llm.Message{Role: llm.RoleAssistant, Text: "Main answer"})
	before, _, _, _, _ := p.Snapshot()
	rawBefore, _ := json.Marshal(before)
	ask := p.PrepareAside()
	p.ctxMgr.addUser("Later main turn") // Captured context must not follow subsequent edits.
	got, err := ask(context.Background(), "A side question")
	if err != nil || got != "Side answer" {
		t.Fatalf("%q %v", got, err)
	}
	req := provider.lastReq
	if len(req.Tools) != 0 || len(req.Messages) != 1 || req.Messages[0].Text != "A side question" || strings.Contains(req.System, "Later main turn") || !strings.Contains(req.System, "Main question") {
		t.Fatalf("bad aside request: %+v", req)
	}
	after, _, _, mode, _ := p.Snapshot()
	rawAfter, _ := json.Marshal(after[:len(before)])
	if string(rawAfter) != string(rawBefore) || len(after) != len(before)+1 || mode != types.ModePlan {
		t.Fatal("aside altered main conversation")
	}
	if p.totalUsage.Input != 20 || p.lastUsage.Input != 0 {
		t.Fatal("aside usage overwrote main context usage")
	}
}
func TestAsideRejectsToolsAndCancellation(t *testing.T) {
	provider := &fakeProvider{responses: []llm.Message{{ToolCalls: []llm.ToolCall{{Name: "shell_command"}}}}}
	p := newTestProcessor(t, t.TempDir(), provider, types.ModeAgent)
	ask := p.PrepareAside()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := ask(ctx, "Q"); err == nil || provider.calls != 0 {
		t.Fatal("called provider after cancellation")
	}
	if _, err := ask(context.Background(), "Q"); err == nil {
		t.Fatal("accepted tool call")
	}
}
