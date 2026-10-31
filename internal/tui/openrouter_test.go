package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/rushikeshg25/cool-code/internal/agent"
	"github.com/rushikeshg25/cool-code/internal/config"
	"github.com/rushikeshg25/cool-code/internal/creds"
	"github.com/rushikeshg25/cool-code/internal/llm"
)

func TestConnectOpenRouterAndSwitchModels(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	for _, key := range []string{"COOLCODE_API_KEY", "COOLCODE_API_BASE_URL", "OPENROUTER_API_KEY", "OPENROUTER_BASE_URL", "OPENAI_BASE_URL"} {
		t.Setenv(key, "")
	}
	t.Setenv("OPENAI_API_KEY", "fake-openai-key")
	m := newTestModel(t)
	found := false
	for i, opt := range connectOptions {
		if opt.provider == "openrouter" {
			m.connectIdx = i
			found = true
			break
		}
	}
	if !found {
		t.Fatal("OpenRouter is missing from /connect")
	}
	m.selectConnectOption()
	m.keyInput.SetValue("fake-router-key")
	m.handleConnectKey(tea.KeyMsg{Type: tea.KeyEnter})
	if creds.APIKey("openrouter") != "fake-router-key" {
		t.Fatal("key not persisted")
	}
	if !m.proc.Connected() {
		t.Fatal("provider not configured")
	}
	for _, step := range []struct{ id, provider string }{
		{"anthropic/claude-sonnet-4.5", "openrouter"},
		{"gpt-4.1", "openai"},
		{"google/gemini-2.5-flash", "openrouter"},
	} {
		if err := m.switchModel(step.id); err != nil {
			t.Fatal(err)
		}
		cfg := config.Load(m.rootDir).LLM
		if cfg.Provider != step.provider || cfg.Model != step.id {
			t.Fatalf("persisted model/provider = %s/%s", cfg.Provider, cfg.Model)
		}
		p, err := llm.New(cfg)
		if err != nil || p.Name() != step.provider {
			t.Fatalf("restart provider = %v, error = %v", p, err)
		}
	}
}

// An obsolete subscription selection must not prevent reaching /connect.
func TestConnectReplacesUnsupportedSavedProvider(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	for _, key := range []string{"COOLCODE_API_KEY", "COOLCODE_API_BASE_URL", "OPENROUTER_BASE_URL"} {
		t.Setenv(key, "")
	}
	if err := config.SetGlobalLLM(config.LLM{Provider: "retired-subscription", Model: "gpt-5"}); err != nil {
		t.Fatal(err)
	}
	root := t.TempDir()
	proc, err := agent.New(root, config.Load(root), agent.Options{AllowMissingKey: true})
	if err != nil {
		t.Fatal(err)
	}
	m := newModel(proc, root, "test", false, "session")
	for i, opt := range connectOptions {
		if opt.provider == "openrouter" {
			m.connectIdx = i
			break
		}
	}
	m.selectConnectOption()
	m.keyInput.SetValue("fake-router-key")
	m.handleConnectKey(tea.KeyMsg{Type: tea.KeyEnter})
	if !proc.Connected() || proc.SetupError() != nil {
		t.Fatal("/connect did not recover")
	}
	cfg := config.Load(root).LLM
	if cfg.Provider != "openrouter" {
		t.Fatalf("saved provider = %q", cfg.Provider)
	}
	if _, err := llm.New(cfg); err != nil {
		t.Fatalf("restart failed: %v", err)
	}
}
