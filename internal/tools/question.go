package tools

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/rushikeshg25/cool-code/internal/types"
)

// Question requests one answer. Options are suggestions; free text is always allowed.
type Question struct {
	Prompt  string   `json:"prompt"`
	Options []string `json:"options,omitempty"`
}
type QuestionHandler func(context.Context, Question) (string, error)

var questionTool = Tool{
	Name:        "question",
	Description: "Ask the user for missing information or a choice, then wait for their answer. Provide one concise prompt and optionally 2-5 short suggested answers. Free text is always available. Dismissal is not approval. Unavailable in noninteractive runs and subagents.",
	Schema:      obj(map[string]any{"prompt": strProp("One concise question with necessary context, up to 2000 characters."), "options": arrProp("Optional 2-5 suggested answers, each up to 200 characters; user can also type an answer.")}, "prompt"),
	Execute: func(ctx Context, raw json.RawMessage) types.ToolResult {
		var q Question
		if err := json.Unmarshal(raw, &q); err != nil {
			return fail("Invalid question", err.Error())
		}
		if strings.TrimSpace(q.Prompt) == "" || len([]rune(q.Prompt)) > 2000 || len(q.Options) == 1 || len(q.Options) > 5 {
			return fail("Invalid question", "Provide a prompt up to 2000 characters and either no options or 2-5 options.")
		}
		for _, option := range q.Options {
			if strings.TrimSpace(option) == "" || len([]rune(option)) > 200 {
				return fail("Invalid question", "Options must contain 1-200 characters.")
			}
		}
		if ctx.AskQuestion == nil {
			return fail("Question unavailable", "Interactive user input is unavailable. Explain what information is missing; do not assume an answer.")
		}
		if err := ctx.Context().Err(); err != nil {
			return fail("Question cancelled", err.Error())
		}
		answer, err := ctx.AskQuestion(ctx.Context(), q)
		if err != nil {
			return fail("Question unanswered", err.Error()+" Do not assume an answer or approval.")
		}
		if strings.TrimSpace(answer) == "" {
			return fail("Question unanswered", "empty answer; do not assume approval")
		}
		return types.ToolResult{Display: "User answered a question", LLMResult: "User answer: " + answer}
	},
}
