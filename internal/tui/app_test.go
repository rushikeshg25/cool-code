package tui

import (
	"os"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/rushikeshg25/cool-code/internal/agent"
	"github.com/rushikeshg25/cool-code/internal/config"
	"github.com/rushikeshg25/cool-code/internal/types"
)

// newTestModel builds a model backed by a real processor. A fake API key lets
// agent.New succeed without any network calls (slash commands never hit the LLM).
func newTestModel(t *testing.T) *model {
	t.Helper()
	t.Setenv("GOOGLE_GENERATIVE_AI_API_KEY", "fake-key")
	root := t.TempDir()
	proc, err := agent.New(root, config.Default(), agent.Options{Mode: types.ModeAgent})
	if err != nil {
		t.Fatalf("agent.New: %v", err)
	}
	m := newModel(proc, root, "test", false, "sess-1")
	// Simulate an initial resize so the viewport is ready.
	updated, _ := m.Update(tea.WindowSizeMsg{Width: 100, Height: 30})
	return updated.(*model)
}

func typeLine(t *testing.T, m *model, line string) *model {
	t.Helper()
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(line)})
	m = updated.(*model)
	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updated.(*model)
	_ = cmd
	return m
}

func TestTypingUpdatesInput(t *testing.T) {
	m := newTestModel(t)
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("hello")})
	m = updated.(*model)
	if m.ti.Value() != "hello" {
		t.Fatalf("input value = %q, want hello", m.ti.Value())
	}
}

func TestSlashModeSwitch(t *testing.T) {
	m := newTestModel(t)
	m = typeLine(t, m, "/mode plan")
	if m.mode != types.ModePlan {
		t.Fatalf("mode = %q, want plan", m.mode)
	}
	if m.proc.Mode() != types.ModePlan {
		t.Fatalf("processor mode = %q, want plan", m.proc.Mode())
	}
}

func TestSlashHelpAppendsHistory(t *testing.T) {
	m := newTestModel(t)
	m = typeLine(t, m, "/help")
	var parts []string
	for _, e := range m.history {
		parts = append(parts, e.rendered)
	}
	joined := strings.Join(parts, "\n")
	if !strings.Contains(joined, "/context") || !strings.Contains(joined, "/install-skill") {
		t.Fatalf("help output missing commands:\n%s", joined)
	}
}

func TestSlashPin(t *testing.T) {
	m := newTestModel(t)
	file := m.rootDir + "/note.txt"
	if err := os.WriteFile(file, []byte("hi"), 0o644); err != nil {
		t.Fatal(err)
	}
	m = typeLine(t, m, "/pin note.txt")
	if len(m.proc.PinnedFiles()) != 1 {
		t.Fatalf("expected 1 pinned file, got %v", m.proc.PinnedFiles())
	}
}

func TestShiftTabCyclesMode(t *testing.T) {
	m := newTestModel(t)
	if m.mode != types.ModeAgent {
		t.Fatalf("start mode = %q", m.mode)
	}
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyShiftTab})
	m = updated.(*model)
	if m.mode != nextMode(types.ModeAgent) {
		t.Fatalf("after shift-tab mode = %q, want %q", m.mode, nextMode(types.ModeAgent))
	}
}

func TestSlashEffortUpdatesProviderAndProjectConfig(t *testing.T) {
	m := newTestModel(t)
	m = typeLine(t, m, "/effort high")
	if got := m.proc.GetStatus().Effort; got != "high" {
		t.Fatalf("processor effort = %q", got)
	}
	cfg := config.Load(m.rootDir)
	if cfg.LLM.ReasoningEffort != "high" {
		t.Fatalf("saved effort = %q", cfg.LLM.ReasoningEffort)
	}
	if !strings.Contains(lastSystemEntry(m), "HIGH") {
		t.Fatalf("effort confirmation = %q", lastSystemEntry(m))
	}
}

func TestExitQuits(t *testing.T) {
	m := newTestModel(t)
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("/exit")})
	m = updated.(*model)
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("expected a quit command")
	}
	if _, ok := cmd().(tea.QuitMsg); !ok {
		t.Fatal("expected tea.QuitMsg from /exit")
	}
}

func TestCommandSuggestions(t *testing.T) {
	m := newTestModel(t)
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("/mo")})
	m = updated.(*model)
	if len(m.suggestions) == 0 {
		t.Fatal("expected suggestions for /mo")
	}
	found := false
	for _, s := range m.suggestions {
		if s.name == "/mode" {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected /mode in suggestions, got %v", m.suggestions)
	}
}
