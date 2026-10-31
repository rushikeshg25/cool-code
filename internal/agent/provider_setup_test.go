package agent

import (
	"testing"

	"github.com/rushikeshg25/cool-code/internal/config"
)

func TestUnsupportedProviderCanRecoverInteractively(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	cfg := config.Default()
	cfg.LLM.Provider = "retired-subscription"
	if _, err := New(t.TempDir(), cfg, Options{}); err == nil {
		t.Fatal("one-shot startup must reject unsupported provider")
	}
	p, err := New(t.TempDir(), cfg, Options{AllowMissingKey: true})
	if err != nil {
		t.Fatalf("interactive startup must allow /connect recovery: %v", err)
	}
	if p.SetupError() == nil {
		t.Fatal("missing startup diagnostic")
	}
	if p.Connected() {
		t.Fatal("unsupported provider must not be connected")
	}
	t.Setenv("OPENAI_API_KEY", "fake-key")
	t.Setenv("OPENAI_BASE_URL", "")
	t.Setenv("COOLCODE_API_BASE_URL", "")
	cfg.LLM.Provider = "openai"
	cfg.LLM.Model = "gpt-4.1"
	if err := p.ConfigureLLM(cfg.LLM); err != nil {
		t.Fatal(err)
	}
	if p.SetupError() != nil {
		t.Fatal("stale setup error after recovery")
	}
	if !p.Connected() {
		t.Fatal("provider did not recover")
	}
}

func TestInvalidEndpointStillBlocksInteractiveStartup(t *testing.T) {
	cfg := config.Default()
	cfg.LLM.Provider = "openai"
	cfg.LLM.BaseURL = "http://untrusted.example/v1"
	if _, err := New(t.TempDir(), cfg, Options{AllowMissingKey: true}); err == nil {
		t.Fatal("invalid endpoint was silently ignored")
	}
}
