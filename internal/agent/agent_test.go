package agent

import (
	"testing"

	"github.com/rushikeshg25/cool-code/internal/config"
	"github.com/rushikeshg25/cool-code/internal/llm"
	"github.com/rushikeshg25/cool-code/internal/project"
	"github.com/rushikeshg25/cool-code/internal/types"
)

func TestParseTaskPlanFencedJSON(t *testing.T) {
	raw := "```json\n{\"goal\":\"Add auth\",\"steps\":[{\"title\":\"a\",\"detail\":\"b\"}],\"assumptions\":[],\"risks\":[]}\n```"
	plan := parseTaskPlan(raw)
	if plan == nil {
		t.Fatal("expected a plan")
	}
	if plan.Goal != "Add auth" || len(plan.Steps) != 1 {
		t.Fatalf("plan = %+v", plan)
	}
}

func TestParseTaskPlanWithPreamble(t *testing.T) {
	raw := "Here is the plan:\n{\"goal\":\"X\",\"steps\":[],\"assumptions\":[],\"risks\":[]}\nThanks!"
	plan := parseTaskPlan(raw)
	if plan == nil || plan.Goal != "X" {
		t.Fatalf("plan = %+v", plan)
	}
}

func TestContextWindowStartsAtUser(t *testing.T) {
	cm := newContextManager(t.TempDir(), config.Default(), project.NewGitIgnoreChecker(t.TempDir()))
	cm.maxTokens = 10_000
	cm.addUser("first question")
	cm.addAssistant(llm.Message{Role: llm.RoleAssistant, ToolCalls: []llm.ToolCall{{ID: "1", Name: "read_file"}}})
	cm.addToolResult(llm.ToolCall{ID: "1", Name: "read_file"}, "file contents")
	cm.addAssistant(llm.Message{Role: llm.RoleAssistant, Text: "done"})

	win := cm.window()
	if len(win) == 0 || win[0].Role != llm.RoleUser {
		t.Fatalf("window must start at a user message, got %+v", win)
	}
}

func TestModeValidity(t *testing.T) {
	if !types.ModePlan.Valid() || !types.ModeAgent.Valid() || !types.ModeAsk.Valid() {
		t.Fatal("known modes should be valid")
	}
	if types.AgentMode("bogus").Valid() {
		t.Fatal("bogus mode should be invalid")
	}
}
