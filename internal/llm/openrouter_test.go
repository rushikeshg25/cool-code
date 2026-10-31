package llm

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/rushikeshg25/cool-code/internal/config"
	"github.com/rushikeshg25/cool-code/internal/creds"
)

func isolateProviderEnv(t *testing.T) {
	t.Helper()
	t.Setenv("HOME", t.TempDir())
	for _, key := range []string{"COOLCODE_API_KEY", "COOLCODE_API_BASE_URL", "OPENROUTER_API_KEY", "OPENROUTER_BASE_URL", "OPENAI_API_KEY", "OPENAI_BASE_URL"} {
		t.Setenv(key, "")
	}
}

func TestOpenRouterCredentialsAndInference(t *testing.T) {
	isolateProviderEnv(t)
	for _, model := range []string{"openai/gpt-4.1", "anthropic/claude-sonnet-4.5", "google/gemini-2.5-flash", "vendor/new-model:free"} {
		if got := InferProvider(model); got != "openrouter" {
			t.Errorf("InferProvider(%q) = %q", model, got)
		}
	}
	cfg := config.LLM{Model: "google/gemini-2.5-flash"}
	t.Setenv("OPENAI_API_KEY", "unrelated-key")
	_, err := New(cfg)
	var missing *MissingKeyError
	if !errors.As(err, &missing) || missing.EnvKey != "OPENROUTER_API_KEY" {
		t.Fatalf("missing key error = %v", err)
	}
	t.Setenv("OPENROUTER_API_KEY", "router-key")
	p, err := New(cfg)
	if err != nil {
		t.Fatal(err)
	}
	if p.Name() != "openrouter" || p.Model() != cfg.Model {
		t.Fatalf("provider = %s, %s", p.Name(), p.Model())
	}
	if err := creds.SetAPIKey("openrouter", "stored-key"); err != nil {
		t.Fatal(err)
	}
	key, _ := resolveAPIKey(cfg, "openrouter", "OPENROUTER_API_KEY", false)
	if key != "stored-key" {
		t.Fatal("stored key did not take precedence")
	}
	cfg.BaseURL = "https://proxy.example/v1"
	if _, err := New(cfg); !errors.As(err, &missing) || missing.EnvKey != "COOLCODE_API_KEY" {
		t.Fatalf("custom endpoint must require explicit key: %v", err)
	}
	cfg.BaseURL = ""
	t.Setenv("OPENROUTER_BASE_URL", "https://proxy.example/v1")
	if _, err := New(cfg); !errors.As(err, &missing) || missing.EnvKey != "COOLCODE_API_KEY" {
		t.Fatalf("environment override must require explicit key: %v", err)
	}
}

func TestOpenRouterCompletions(t *testing.T) {
	for _, stream := range []bool{false, true} {
		t.Run(fmt.Sprint(stream), func(t *testing.T) {
			isolateProviderEnv(t)
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != "/api/v1/chat/completions" || r.Header.Get("Authorization") != "Bearer router-key" {
					t.Error("incorrect endpoint or credential")
				}
				var body map[string]any
				if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
					t.Error(err)
					return
				}
				if body["model"] != "google/gemini-2.5-flash" || body["max_tokens"] != float64(32) {
					t.Errorf("unexpected model/token limit: %v", body)
				}
				if body["reasoning"].(map[string]any)["effort"] != "high" {
					t.Error("missing reasoning effort")
				}
				if len(body["tools"].([]any)) != 1 || len(body["messages"].([]any)) != 4 {
					t.Error("missing tools or conversation history")
				}
				if stream {
					w.Header().Set("Content-Type", "text/event-stream")
					fmt.Fprint(w, ": OPENROUTER PROCESSING\n\ndata: {\"choices\":[{\"delta\":{\"content\":\"Hi\",\"tool_calls\":[{\"index\":0,\"id\":\"call-2\",\"function\":{\"name\":\"read_file\",\"arguments\":\"{}\"}}]}}]}\n\ndata: {\"choices\":[{\"delta\":{},\"finish_reason\":\"tool_calls\"}],\"usage\":{\"prompt_tokens\":11,\"completion_tokens\":3}}\n\ndata: [DONE]\n\n")
				} else {
					fmt.Fprint(w, `{"choices":[{"message":{"content":"Hi","tool_calls":[{"id":"call-2","function":{"name":"read_file","arguments":"{}"}}]}}],"usage":{"prompt_tokens":11,"completion_tokens":3}}`)
				}
			}))
			defer srv.Close()
			t.Setenv("COOLCODE_API_KEY", "router-key")
			maxTokens := 32
			p, err := New(config.LLM{Provider: "openrouter", Model: "google/gemini-2.5-flash", BaseURL: srv.URL + "/api/v1", ReasoningEffort: "high", MaxTokens: &maxTokens})
			if err != nil {
				t.Fatal(err)
			}
			req := Request{System: "system", Messages: []Message{{Role: RoleUser, Text: "hi"}, {Role: RoleAssistant, ToolCalls: []ToolCall{{ID: "call-1", Name: "read_file", Arguments: json.RawMessage(`{}`)}}}, {Role: RoleTool, ToolCallID: "call-1", Text: "result"}}, Tools: []ToolDef{{Name: "read_file", Parameters: map[string]any{"type": "object"}}}}
			var msg Message
			if stream {
				var deltas strings.Builder
				msg, err = p.(Streamer).Stream(context.Background(), req, func(s string) { deltas.WriteString(s) })
				if deltas.String() != "Hi" {
					t.Errorf("deltas = %q", deltas.String())
				}
			} else {
				msg, err = p.Complete(context.Background(), req)
			}
			if err != nil {
				t.Fatal(err)
			}
			if msg.Text != "Hi" || msg.Usage.Input != 11 || msg.Usage.Output != 3 || len(msg.ToolCalls) != 1 || msg.ToolCalls[0].ID != "call-2" {
				t.Fatalf("response = %+v", msg)
			}
		})
	}
}

func TestOpenRouterEndpointAndMetadata(t *testing.T) {
	if got := providerEndpoint("openrouter", ""); got != "https://openrouter.ai/api/v1/chat/completions" {
		t.Fatalf("endpoint = %s", got)
	}
	for _, id := range []string{"google/gemini-2.5-flash", "openai/gpt-4.1", "anthropic/claude-sonnet-4.5"} {
		if _, ok := PriceFor(id); !ok {
			t.Errorf("missing price for %s", id)
		}
		if _, ok := ContextWindow(id); !ok {
			t.Errorf("missing window for %s", id)
		}
	}
	for _, id := range []string{"openai/gpt-4.1:free", "vendor/unknown"} {
		if _, ok := PriceFor(id); ok {
			t.Errorf("invented price for %s", id)
		}
		if _, ok := ContextWindow(id); ok {
			t.Errorf("invented window for %s", id)
		}
	}
}

func TestOpenRouterUpstreamErrors(t *testing.T) {
	for _, stream := range []bool{false, true} {
		t.Run(fmt.Sprint(stream), func(t *testing.T) {
			isolateProviderEnv(t)
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if stream {
					fmt.Fprint(w, "data: {\"choices\":[{\"delta\":{\"content\":\"partial\"}}]}\n\ndata: {\"error\":{\"code\":502,\"message\":\"upstream unavailable\"}}\n\ndata: [DONE]\n\n")
				} else {
					fmt.Fprint(w, `{"error":{"code":502,"message":"upstream unavailable"}}`)
				}
			}))
			defer srv.Close()
			t.Setenv("COOLCODE_API_KEY", "fake-key")
			p, err := New(config.LLM{Provider: "openrouter", Model: "google/gemini-2.5-flash", BaseURL: srv.URL})
			if err != nil {
				t.Fatal(err)
			}
			if stream {
				_, err = p.(Streamer).Stream(context.Background(), Request{}, func(string) {})
			} else {
				_, err = p.Complete(context.Background(), Request{})
			}
			if err == nil || !strings.Contains(err.Error(), "upstream unavailable") {
				t.Fatalf("error = %v", err)
			}
		})
	}
}
