package tui

import (
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/ansi"
	"github.com/rushikeshg25/cool-code/internal/config"
	"github.com/rushikeshg25/cool-code/internal/creds"
	"github.com/rushikeshg25/cool-code/internal/llm"
	"strings"
	"testing"
)

func TestModelPickerFiltersSelectsAndPersistsProvider(t *testing.T) {
	m := newTestModel(t)
	if err := creds.SetAPIKey("openrouter", "test"); err != nil {
		t.Fatal(err)
	}
	_, cmd := m.dispatchCommand("/model")
	if cmd == nil || !m.picker.loading {
		t.Fatal("no asynchronous picker")
	}
	m.receiveCatalog(catalogMsg{m.pickerGeneration, llm.ModelCatalog{Models: []llm.ModelInfo{{Provider: "google", ID: "gemini-test"}, {Provider: "openrouter", ID: "openai/gpt-test"}}}})
	m.handlePickerKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("openrouter")})
	if len(m.picker.filtered) != 1 {
		t.Fatal("filter did not match provider")
	}
	m.handlePickerKey(tea.KeyMsg{Type: tea.KeyEnter})
	got := m.proc.LLMConfig()
	saved := config.Load(m.rootDir).LLM
	if m.picker != nil || got.Provider != "openrouter" || saved.Provider != got.Provider || saved.Model != got.Model {
		t.Fatalf("active=%+v saved=%+v", got, saved)
	}
}
func TestModelPickerCancelIgnoresLateResults(t *testing.T) {
	m := newTestModel(t)
	m.openModelPicker()
	generation := m.pickerGeneration
	m.handlePickerKey(tea.KeyMsg{Type: tea.KeyEsc})
	m.openEffortPicker()
	m.receiveCatalog(catalogMsg{generation, llm.ModelCatalog{Models: []llm.ModelInfo{{Provider: "openai", ID: "gpt-test"}}}})
	if m.picker.kind != "effort" {
		t.Fatal("late result replaced picker")
	}
}
func TestEffortPickerResetOverridesGlobalEffort(t *testing.T) {
	m := newTestModel(t)
	if err := config.SetGlobalLLM(config.LLM{ReasoningEffort: "high"}); err != nil {
		t.Fatal(err)
	}
	m.openEffortPicker()
	m.handlePickerKey(tea.KeyMsg{Type: tea.KeyEnter})
	if got := config.Load(m.rootDir).LLM.ReasoningEffort; got != "" {
		t.Fatalf("default effort did not persist: %s", got)
	}
}
func TestPickersFitAndScroll(t *testing.T) {
	for _, width := range []int{32, 80, 120} {
		m := resizeModel(t, newTestModel(t), width, 24)
		m.openModelPicker()
		catalog := llm.ModelCatalog{}
		for i := 0; i < 100; i++ {
			catalog.Models = append(catalog.Models, llm.ModelInfo{Provider: "openai", ID: fmt.Sprintf("gpt-%03d", i), Name: strings.Repeat("Name ", 30)})
		}
		m.receiveCatalog(catalogMsg{m.pickerGeneration, catalog})
		m.handlePickerKey(tea.KeyMsg{Type: tea.KeyEnd})
		rendered := m.renderPicker()
		if !strings.Contains(rendered, "gpt-099") {
			t.Fatal("cannot reach last model")
		}
		for _, line := range strings.Split(rendered, "\n") {
			if ansi.StringWidth(line) > width {
				t.Fatalf("picker wider than %d: %q", width, line)
			}
		}
	}
}
func TestDirectModeCommands(t *testing.T) {
	m := newTestModel(t)
	for command, want := range map[string]string{"/plan": "plan", "/ask": "ask", "/impl": "agent"} {
		m.dispatchCommand(command)
		if string(m.proc.Mode()) != want {
			t.Fatalf("%s did not select %s", command, want)
		}
	}
	for _, c := range commands {
		if c.name == "/mode" {
			t.Fatal("obsolete mode command remains")
		}
	}
}
func TestEffortMeterAnimation(t *testing.T) {
	if effortMeter("xhigh", 0) == effortMeter("xhigh", 1) {
		t.Fatal("xhigh should animate")
	}
	if effortMeter("low", 0) != effortMeter("low", 1) {
		t.Fatal("low should be steady")
	}
}
