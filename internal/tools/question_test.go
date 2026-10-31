package tools

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"
)

func TestQuestionRequiresActualAnswer(t *testing.T) {
	args := json.RawMessage(`{"prompt":"Which database?","options":["Postgres","SQLite"]}`)
	if r := Run(Context{}, "question", args); !r.Failed {
		t.Fatal("noninteractive question guessed answer")
	}
	ctx := Context{AskQuestion: func(context.Context, Question) (string, error) { return "My own answer", nil }}
	if r := Run(ctx, "question", args); r.Failed || !strings.Contains(r.LLMResult, "My own answer") {
		t.Fatal(r)
	}
	ctx.AskQuestion = func(context.Context, Question) (string, error) { return "", errors.New("dismissed") }
	if r := Run(ctx, "question", args); !r.Failed {
		t.Fatal("dismissal accepted")
	}
	if IsReadOnly("question") || IsMutating("question") {
		t.Fatal("question should be sequential, available in main read-only modes, absent from subagents")
	}
	if !RunReadOnly(ctx, "question", args).Failed {
		t.Fatal("subagent question bypass")
	}
}
func TestQuestionValidation(t *testing.T) {
	for _, args := range []string{`{}`, `{"prompt":"Q","options":["one"]}`, `{"prompt":"Q","options":["one",""]}`} {
		if !Run(Context{}, "question", json.RawMessage(args)).Failed {
			t.Fatal(args)
		}
	}
}
