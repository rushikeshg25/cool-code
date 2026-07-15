package agent

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/rushikeshg25/cool-code/internal/config"
	"github.com/rushikeshg25/cool-code/internal/llm"
)

// TaskPlanStep is a single planned step.
type TaskPlanStep struct {
	Title  string `json:"title"`
	Detail string `json:"detail"`
}

// TaskPlan is a one-shot structured plan produced by the `task` subcommand.
type TaskPlan struct {
	Goal        string         `json:"goal"`
	Steps       []TaskPlanStep `json:"steps"`
	Assumptions []string       `json:"assumptions"`
	Risks       []string       `json:"risks"`
}

const taskPromptTemplate = `You are an expert engineering lead. Produce a concise, actionable plan.
Return ONLY valid JSON with this shape:
{
  "goal": string,
  "steps": [{"title": string, "detail": string}],
  "assumptions": string[],
  "risks": string[]
}

Constraints:
- 4 to 8 steps.
- Each detail should be 1-2 sentences.
- Keep assumptions and risks short.

Goal: `

// CreateTaskPlan asks the model for a structured plan, or nil on failure.
func CreateTaskPlan(ctx context.Context, cfg config.Config, goal string) (*TaskPlan, error) {
	provider, err := llm.New(cfg.LLM)
	if err != nil {
		return nil, err
	}
	resp, err := provider.Complete(ctx, llm.Request{
		Messages: []llm.Message{{Role: llm.RoleUser, Text: taskPromptTemplate + goal}},
	})
	if err != nil {
		return nil, err
	}
	plan := parseTaskPlan(resp.Text)
	if plan == nil || plan.Goal == "" {
		return nil, nil
	}
	return plan, nil
}

func parseTaskPlan(raw string) *TaskPlan {
	cleaned := strings.TrimSpace(raw)
	if strings.HasPrefix(cleaned, "```") {
		if i := strings.IndexByte(cleaned, '\n'); i != -1 {
			cleaned = cleaned[i+1:]
		}
		cleaned = strings.TrimSuffix(strings.TrimSpace(cleaned), "```")
	}
	// Trim to the outermost JSON object.
	if s := strings.IndexByte(cleaned, '{'); s != -1 {
		if e := strings.LastIndexByte(cleaned, '}'); e > s {
			cleaned = cleaned[s : e+1]
		}
	}
	var plan TaskPlan
	if err := json.Unmarshal([]byte(cleaned), &plan); err != nil {
		return nil
	}
	return &plan
}
