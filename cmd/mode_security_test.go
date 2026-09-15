package cmd

import (
	"testing"

	"github.com/rushikeshg25/cool-code/internal/agent"
	"github.com/rushikeshg25/cool-code/internal/config"
	"github.com/rushikeshg25/cool-code/internal/session"
	"github.com/rushikeshg25/cool-code/internal/types"
)

func TestStartupAndResumeHonorExplicitMode(t *testing.T) {
	for _, requested := range []string{"", "agent", "ask", "plan"} {
		for _, saved := range []types.AgentMode{types.ModeAgent, types.ModeAsk, types.ModePlan} {
			t.Run(requested+"/"+string(saved), func(t *testing.T) {
				mode, err := startingMode(requested)
				if err != nil {
					t.Fatal(err)
				}
				cfg := config.Default()
				p, err := agent.New(t.TempDir(), cfg, agent.Options{Mode: mode, AllowMissingKey: true})
				if err != nil {
					t.Fatal(err)
				}
				if p.Mode() != mode {
					t.Fatalf("startup mode %s", p.Mode())
				}
				restoreConversation(p, &session.Data{Mode: string(saved)}, requested, mode)
				want := saved
				if requested != "" {
					want = mode
				}
				if p.Mode() != want {
					t.Fatalf("mode %s, want %s", p.Mode(), want)
				}
			})
		}
	}
}

func TestInteractiveRejectsInvalidModeBeforeStartup(t *testing.T) {
	if err := runInteractive(rootFlags{mode: "supervisor"}); err == nil {
		t.Fatal("invalid mode accepted")
	}
}
