package agent

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/rushikeshg25/cool-code/internal/config"
	"github.com/rushikeshg25/cool-code/internal/llm"
	"github.com/rushikeshg25/cool-code/internal/types"
)

func TestQueuedSecretsDoNotReachProviderOrSession(t *testing.T) {
	t.Setenv("COOLCODE_TEST_API_KEY", "queued-environment-value")
	provider := &fakeProvider{responses: []llm.Message{{Role: llm.RoleAssistant, ToolCalls: []llm.ToolCall{toolCall("1", "project_summary", map[string]any{})}}}}
	p := newTestProcessor(t, t.TempDir(), provider, types.ModeAgent)
	p.EnqueueMessage("follow-up queued-environment-value password=synthetic-password")
	if _, err := p.ProcessQuery(context.Background(), "inspect", nil); err != nil {
		t.Fatal(err)
	}
	messages, summary, _, _, _ := p.Snapshot()
	persisted, _ := json.Marshal(map[string]any{"messages": messages, "summary": summary})
	request, _ := json.Marshal(provider.lastReq)
	for _, data := range []string{string(persisted), string(request)} {
		for _, secret := range []string{"queued-environment-value", "synthetic-password"} {
			if strings.Contains(data, secret) {
				t.Fatal("secret leaked")
			}
		}
		if !strings.Contains(data, "follow-up") || !strings.Contains(data, "[REDACTED]") {
			t.Fatalf("missing sanitized follow-up: %s", data)
		}
	}
}

func TestRestoredConversationRedactsLegacyQueuedSecrets(t *testing.T) {
	p := newTestProcessor(t, t.TempDir(), &fakeProvider{}, types.ModeAgent)
	p.Restore([]llm.Message{{Role: llm.RoleUser, Text: "password=legacy-secret"}}, "password=legacy-secret", nil, types.ModeAsk)
	messages, summary, _, _, _ := p.Snapshot()
	data, _ := json.Marshal(messages)
	if strings.Contains(string(data)+summary, "legacy-secret") {
		t.Fatal("legacy secret retained")
	}
}

func TestTaskGoalIsRedactedAtProviderEgress(t *testing.T) {
	t.Setenv("TASK_TEST_API_KEY", "task-environment-secret")
	received := make(chan string, 1)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		received <- string(body)
		w.Header().Set("Content-Type", "application/json")
		io.WriteString(w, `{"choices":[{"message":{"role":"assistant","content":"{\"goal\":\"safe plan\",\"steps\":[]}"}}]}`)
	}))
	defer srv.Close()
	cfg := config.Default()
	cfg.LLM.Provider = "openai"
	cfg.LLM.Model = "gpt-4o"
	cfg.LLM.BaseURL = srv.URL
	cfg.LLM.APIKeyEnv = "TASK_TEST_API_KEY"
	plan, err := CreateTaskPlan(context.Background(), cfg, "ordinary goal task-environment-secret password=task-inline-secret")
	if err != nil {
		t.Fatal(err)
	}
	if plan == nil || plan.Goal != "safe plan" {
		t.Fatalf("plan: %+v", plan)
	}
	body := <-received
	if strings.Contains(body, "task-environment-secret") || strings.Contains(body, "task-inline-secret") {
		t.Fatal("task secret leaked")
	}
	if !strings.Contains(body, "ordinary goal") || !strings.Contains(body, "[REDACTED]") {
		t.Fatal("goal lost")
	}
}
