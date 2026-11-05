package agent

import (
	"encoding/json"
	"github.com/rushikeshg25/cool-code/internal/llm"
	"strings"
	"testing"
)

func TestCompactionRetainsIncompleteToolCalls(t *testing.T) {
	c := &contextManager{}
	c.addUser("unfinished request")
	c.addAssistant(llm.Message{Role: llm.RoleAssistant, Text: "interrupted edit", ToolCalls: []llm.ToolCall{{ID: "partial-1", Name: "edit_file", Arguments: json.RawMessage("{")}}})
	for i := 0; i < 20; i++ {
		c.addUser("recent")
	}
	got := c.transcriptForSummary(24000)
	for _, want := range []string{"interrupted edit", "partial-1", "edit_file", "{"} {
		if !strings.Contains(got, want) {
			t.Fatalf("lost %s in summary input", want)
		}
	}
	if string(c.messages[1].ToolCalls[0].Arguments) != "{" {
		t.Fatal("summary mutated conversation")
	}
}
