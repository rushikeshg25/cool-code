package agent

import (
	"context"
	"encoding/json"
	"strconv"
	"strings"

	"github.com/rushikeshg25/cool-code/internal/llm"
	"github.com/rushikeshg25/cool-code/internal/security"
	"github.com/rushikeshg25/cool-code/internal/tools"
)

const spawnAgentTool = "spawn_agent"

const subagentMaxIterations = 15

var spawnAgentDef = llm.ToolDef{
	Name: spawnAgentTool,
	Description: "Spawns a read-only explore subagent that investigates the codebase and reports back. " +
		"Use for broad searches or multi-file understanding; issue several spawn_agent calls in one turn " +
		"to explore independent areas in parallel. The subagent cannot edit files or run shell commands.",
	Parameters: map[string]any{
		"type": "object",
		"properties": map[string]any{
			"task": map[string]any{
				"type":        "string",
				"description": "The exploration goal: what to investigate and what to report back.",
			},
		},
		"required": []string{"task"},
	},
}

// subagentToolDefs is the read-only tool surface exposed to subagents. Built
// from tools.All at init, so spawn_agent itself is structurally excluded and
// subagents can never nest.
var subagentToolDefs = func() []llm.ToolDef {
	var defs []llm.ToolDef
	for _, t := range tools.ReadOnlyTools() {
		defs = append(defs, llm.ToolDef{Name: t.Name, Description: t.Description, Parameters: t.Schema})
	}
	return defs
}()

// runSubagent runs one explore-only mini agent loop to completion and returns
// its final report text. Errors are folded into the returned text so the main
// model can react to them. report receives short state updates for the TUI.
func (p *Processor) runSubagent(ctx context.Context, task string, report func(state string)) string {
	system := subagentPrompt + "\n\n--- Project State ---\nCWD: " + p.rootDir + "\nFile Tree:\n" + p.ctxMgr.tree()
	if extras := p.ctxMgr.extraDirList(); len(extras) > 0 {
		system += "\nAdditional directories (use absolute paths): " + strings.Join(extras, ", ")
	}
	messages := []llm.Message{{Role: llm.RoleUser, Text: task}}
	toolCtx := p.toolCtx
	toolCtx.Ctx = ctx

	toolsRun := 0
	lastText := ""
	for iter := 0; iter < subagentMaxIterations; iter++ {
		if err := ctx.Err(); err != nil {
			return "[subagent cancelled] " + lastText
		}
		resp, err := p.provider.Complete(ctx, llm.Request{
			System:   system,
			Messages: messages,
			Tools:    subagentToolDefs,
		})
		if err != nil {
			return "[subagent error] " + err.Error()
		}
		messages = append(messages, resp)
		if resp.Text != "" {
			lastText = resp.Text
		}
		if len(resp.ToolCalls) == 0 {
			return lastText
		}
		for _, call := range resp.ToolCalls {
			result := tools.Run(toolCtx, call.Name, call.Arguments)
			messages = append(messages, llm.Message{
				Role: llm.RoleTool,
				// Redact here as the parent loop does. These messages are
				// resent to the provider on every later subagent iteration,
				// and only the subagent's final text passes through the
				// parent's own redaction.
				Text:       security.Redact(result.LLMResult),
				ToolCallID: call.ID,
				ToolName:   call.Name,
			})
			toolsRun++
		}
		report("exploring (" + strconv.Itoa(toolsRun) + " tools)")
	}
	return "[subagent hit its iteration limit; findings may be incomplete]\n" + lastText
}

// parseSpawnTask extracts the task argument of a spawn_agent call.
func parseSpawnTask(args json.RawMessage) (string, bool) {
	var a struct {
		Task string `json:"task"`
	}
	if err := json.Unmarshal(args, &a); err != nil || a.Task == "" {
		return "", false
	}
	return a.Task, true
}

// truncateTask shortens a task description for status lines.
func truncateTask(task string) string {
	const max = 40
	if len(task) <= max {
		return task
	}
	return task[:max] + "…"
}
