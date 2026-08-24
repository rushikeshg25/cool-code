// Package llm provides a provider-agnostic chat interface with native
// tool-calling over Anthropic, OpenAI, and Google (Gemini) REST APIs.
package llm

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/rushikeshg25/cool-code/internal/config"
	"github.com/rushikeshg25/cool-code/internal/creds"
)

// Role identifies the author of a message.
type Role string

const (
	RoleUser      Role = "user"
	RoleAssistant Role = "assistant"
	RoleTool      Role = "tool"
)

// ToolCall is a model request to invoke a tool.
type ToolCall struct {
	ID        string          `json:"id"`
	Name      string          `json:"name"`
	Arguments json.RawMessage `json:"arguments"`
}

// Usage reports token counts from the provider for one completion.
type Usage struct {
	Input  int `json:"input,omitempty"`
	Output int `json:"output,omitempty"`
}

// Message is one turn in the conversation. Assistant turns may carry ToolCalls;
// tool turns carry a result addressed to a prior ToolCall via ToolCallID.
type Message struct {
	Role       Role       `json:"role"`
	Text       string     `json:"text,omitempty"`
	ToolCalls  []ToolCall `json:"toolCalls,omitempty"`
	ToolCallID string     `json:"toolCallId,omitempty"`
	ToolName   string     `json:"toolName,omitempty"`
	Usage      Usage      `json:"usage,omitzero"`
}

// ToolDef describes a tool for the provider (JSON-schema parameters).
type ToolDef struct {
	Name        string
	Description string
	Parameters  map[string]any
}

// Request is a single completion request.
type Request struct {
	System      string
	Messages    []Message
	Tools       []ToolDef
	Temperature *float64
	MaxTokens   int
}

// Provider performs completions for one backend.
type Provider interface {
	Name() string
	Model() string
	Complete(ctx context.Context, req Request) (Message, error)
}

// Streamer is an optional Provider capability: stream the completion, calling
// onDelta with each assistant-text fragment, and return the fully assembled
// message (identical to what Complete would have returned).
type Streamer interface {
	Stream(ctx context.Context, req Request, onDelta func(text string)) (Message, error)
}

// providerInfo maps a provider to its API-key env var and signup URL.
type providerInfo struct {
	kind   string
	envKey string
	keyURL string
}

var providers = map[string]providerInfo{
	"google":    {"google", "GOOGLE_GENERATIVE_AI_API_KEY", "https://aistudio.google.com/app/apikey"},
	"openai":    {"openai", "OPENAI_API_KEY", "https://platform.openai.com/api-keys"},
	"anthropic": {"anthropic", "ANTHROPIC_API_KEY", "https://console.anthropic.com/settings/keys"},
}

// InferProvider guesses the provider from a model id.
func InferProvider(model string) string {
	m := strings.ToLower(model)
	if strings.Contains(m, "gpt") || strings.HasPrefix(m, "o1") || strings.HasPrefix(m, "o3") {
		return "openai"
	}
	if strings.Contains(m, "claude") {
		return "anthropic"
	}
	return "google"
}

// ResolveProvider returns the configured provider or the inferred one.
func ResolveProvider(cfg config.LLM) string {
	if cfg.Provider != "" {
		return cfg.Provider
	}
	return InferProvider(cfg.Model)
}

// MissingKeyError signals that the required API key env var is unset.
type MissingKeyError struct {
	Provider string
	EnvKey   string
	KeyURL   string
}

func (e *MissingKeyError) Error() string {
	return fmt.Sprintf("missing API key for %s (set %s)", e.Provider, e.EnvKey)
}

// New builds a Provider for the given config, returning *MissingKeyError when
// the required API key is absent. Keys come from the /connect credentials
// store first, then the provider's env var.
func New(cfg config.LLM) (Provider, error) {
	name := ResolveProvider(cfg)
	info, ok := providers[name]
	if !ok {
		return nil, fmt.Errorf("unknown provider: %s", name)
	}
	key, keyEnv := resolveAPIKey(cfg, name, info.envKey)
	if key == "" {
		return nil, &MissingKeyError{Provider: name, EnvKey: keyEnv, KeyURL: info.keyURL}
	}
	base := baseProvider{
		model:           cfg.Model,
		apiKey:          key,
		baseURL:         resolveBaseURL(cfg, name),
		reasoningEffort: cfg.ReasoningEffort,
		temperature:     cfg.Temperature,
		maxTokens:       4096,
	}
	if cfg.MaxTokens != nil && *cfg.MaxTokens > 0 {
		base.maxTokens = *cfg.MaxTokens
	}
	switch name {
	case "anthropic":
		return &anthropicProvider{base}, nil
	case "openai":
		return &openaiProvider{base}, nil
	default:
		return &googleProvider{base}, nil
	}
}

// resolveAPIKey supports proxy credentials without persisting the secret in
// .coolcode.json. apiKeyEnv stores only the name of the environment variable.
func resolveAPIKey(cfg config.LLM, provider, providerEnv string) (key, envName string) {
	if cfg.APIKeyEnv != "" {
		return os.Getenv(cfg.APIKeyEnv), cfg.APIKeyEnv
	}
	if key = os.Getenv("COOLCODE_API_KEY"); key != "" {
		return key, "COOLCODE_API_KEY"
	}
	if key = creds.APIKey(provider); key != "" {
		return key, providerEnv
	}
	return os.Getenv(providerEnv), providerEnv
}

// resolveBaseURL returns an optional provider endpoint override. The generic
// COOLCODE variable is useful for OpenAI-compatible gateways and CLI proxies;
// provider-specific variables preserve common ecosystem conventions.
func resolveBaseURL(cfg config.LLM, provider string) string {
	if cfg.BaseURL != "" {
		return cfg.BaseURL
	}
	if base := os.Getenv("COOLCODE_API_BASE_URL"); base != "" {
		return base
	}
	var env string
	switch provider {
	case "openai":
		env = "OPENAI_BASE_URL"
	case "anthropic":
		env = "ANTHROPIC_BASE_URL"
	case "google":
		env = "GOOGLE_GENERATIVE_AI_BASE_URL"
	}
	return os.Getenv(env)
}

type baseProvider struct {
	model           string
	apiKey          string
	baseURL         string
	reasoningEffort string
	temperature     *float64
	maxTokens       int
}

func (b baseProvider) Model() string { return b.model }
