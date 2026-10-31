package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/rushikeshg25/cool-code/internal/config"
	"github.com/rushikeshg25/cool-code/internal/creds"
	"github.com/rushikeshg25/cool-code/internal/llm"
)

func TestConnectCLIProxyAndKeepModelRouting(t *testing.T) {
	m := newTestModel(t)
	for _, key := range []string{"COOLCODE_API_KEY", "COOLCODE_API_BASE_URL", "CLIPROXY_API_KEY", "CLIPROXY_BASE_URL"} {
		t.Setenv(key, "")
	}
	found := false
	for i, opt := range connectOptions {
		if opt.provider == "cliproxy" {
			m.connectIdx = i
			found = true
			break
		}
	}
	if !found {
		t.Fatal("CLI proxy connect choice missing")
	}
	m.selectConnectOption()
	m.keyInput.SetValue("proxy-key")
	m.handleConnectKey(tea.KeyMsg{Type: tea.KeyEnter})
	if !m.proc.Connected() || creds.APIKey("cliproxy") != "proxy-key" {
		t.Fatal("proxy not connected")
	}
	for _, id := range []string{"gpt-6.1-sol", "claude-proxy-alias", "vendor/custom"} {
		if err := m.switchModel(id); err != nil {
			t.Fatal(err)
		}
		cfg := config.Load(m.rootDir).LLM
		p, err := llm.New(cfg)
		if err != nil || p.Name() != "cliproxy" || p.Model() != id {
			t.Fatalf("model switch left proxy: %v", err)
		}
	}
}
