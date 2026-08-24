package llm

import (
	"testing"

	"github.com/rushikeshg25/cool-code/internal/config"
)

func TestInferProvider(t *testing.T) {
	cases := map[string]string{
		"gemini-2.5-flash":   "google",
		"gpt-4o":             "openai",
		"o1-preview":         "openai",
		"o3-mini":            "openai",
		"claude-sonnet-5":    "anthropic",
		"claude-opus-4-8":    "anthropic",
		"some-unknown-model": "google",
	}
	for model, want := range cases {
		if got := InferProvider(model); got != want {
			t.Errorf("InferProvider(%q) = %q, want %q", model, got, want)
		}
	}
}

func TestConvertSchemaForGemini(t *testing.T) {
	in := map[string]any{
		"type": "object",
		"properties": map[string]any{
			"name": map[string]any{"type": "string"},
			"tags": map[string]any{"type": "array", "items": map[string]any{"type": "string"}},
		},
		"required": []string{"name"},
	}
	out := convertSchemaForGemini(in)
	if out["type"] != "OBJECT" {
		t.Errorf("root type = %v", out["type"])
	}
	props := out["properties"].(map[string]any)
	if props["name"].(map[string]any)["type"] != "STRING" {
		t.Errorf("name type = %v", props["name"])
	}
	tags := props["tags"].(map[string]any)
	if tags["type"] != "ARRAY" {
		t.Errorf("tags type = %v", tags["type"])
	}
	if tags["items"].(map[string]any)["type"] != "STRING" {
		t.Errorf("items type = %v", tags["items"])
	}
}

func TestMissingKeyError(t *testing.T) {
	err := &MissingKeyError{Provider: "google", EnvKey: "GOOGLE_GENERATIVE_AI_API_KEY"}
	if err.Error() == "" {
		t.Error("expected non-empty error message")
	}
}

func TestProviderEndpoint(t *testing.T) {
	cases := []struct {
		provider string
		base     string
		want     string
	}{
		{"openai", "http://localhost:8317/v1", "http://localhost:8317/v1/chat/completions"},
		{"openai", "http://localhost:8317/v1/chat/completions", "http://localhost:8317/v1/chat/completions"},
		{"anthropic", "https://proxy.test/v1", "https://proxy.test/v1/messages"},
		{"google", "https://proxy.test/v1beta", "https://proxy.test/v1beta/models/"},
	}
	for _, tc := range cases {
		if got := providerEndpoint(tc.provider, tc.base); got != tc.want {
			t.Errorf("providerEndpoint(%q, %q) = %q, want %q", tc.provider, tc.base, got, tc.want)
		}
	}
}

func TestProxyEnvironmentOverrides(t *testing.T) {
	t.Setenv("TEST_CLIPROXY_KEY", "proxy-secret")
	key, env := resolveAPIKey(config.LLM{APIKeyEnv: "TEST_CLIPROXY_KEY"}, "openai", "OPENAI_API_KEY")
	if key != "proxy-secret" || env != "TEST_CLIPROXY_KEY" {
		t.Fatalf("resolveAPIKey = %q, %q", key, env)
	}

	t.Setenv("COOLCODE_API_BASE_URL", "http://proxy.test/v1")
	if got := resolveBaseURL(config.LLM{}, "openai"); got != "http://proxy.test/v1" {
		t.Fatalf("resolveBaseURL = %q", got)
	}
}

func TestOpenAIBodyIncludesReasoningEffort(t *testing.T) {
	provider := &openaiProvider{baseProvider{model: "gpt-test", reasoningEffort: "high", maxTokens: 32}}
	body, _ := provider.buildBody(Request{Messages: []Message{{Role: RoleUser, Text: "hi"}}})
	if got := body["reasoning_effort"]; got != "high" {
		t.Fatalf("reasoning_effort = %v", got)
	}
}
