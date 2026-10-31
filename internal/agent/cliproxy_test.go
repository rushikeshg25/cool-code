package agent

import (
	"testing"

	"github.com/rushikeshg25/cool-code/internal/config"
	"github.com/rushikeshg25/cool-code/internal/llm"
)

func TestCLIProxyDoesNotBorrowAPIBillingOrContext(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Setenv("CLIPROXY_API_KEY", "test-proxy-key")
	t.Setenv("COOLCODE_API_BASE_URL", "")
	t.Setenv("CLIPROXY_BASE_URL", "")
	cfg := config.Default()
	cfg.LLM.Provider = "cliproxy"
	cfg.LLM.Model = "gpt-4o"
	p, err := New(t.TempDir(), cfg, Options{})
	if err != nil {
		t.Fatal(err)
	}
	p.totalUsage = llm.Usage{Input: 100, Output: 100}
	status := p.GetStatus()
	if status.CostKnown || status.MaxTokens != 0 {
		t.Fatal("proxy borrowed first-party metadata")
	}
	declared := 32000
	cfg.LLM.ContextWindow = &declared
	if err := p.ConfigureLLM(cfg.LLM); err != nil {
		t.Fatal(err)
	}
	if p.GetStatus().MaxTokens != declared {
		t.Fatal("explicit context window ignored")
	}
}
