package tui

import (
	"context"
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/ansi"
	"github.com/rushikeshg25/cool-code/internal/config"
	"github.com/rushikeshg25/cool-code/internal/llm"
	"github.com/rushikeshg25/cool-code/internal/security"
)

type pickerItem struct{ label, description, value, provider string }
type selectionPicker struct {
	kind     string
	items    []pickerItem
	filtered []int
	selected int
	query    string
	loading  bool
	cancel   context.CancelFunc
	frame    int
}
type catalogMsg struct {
	generation int
	catalog    llm.ModelCatalog
}
type effortTickMsg struct{ generation int }

func effortTick(generation int) tea.Cmd {
	return tea.Tick(250*time.Millisecond, func(time.Time) tea.Msg { return effortTickMsg{generation} })
}

func (m *model) closePicker() {
	if m.picker != nil && m.picker.cancel != nil {
		m.picker.cancel()
	}
	m.picker = nil
	m.pickerGeneration++
}
func (m *model) openModelPicker() (tea.Model, tea.Cmd) {
	m.closePicker()
	ctx, cancel := context.WithCancel(context.Background())
	m.picker = &selectionPicker{kind: "model", loading: true, cancel: cancel}
	generation := m.pickerGeneration
	cfg := m.proc.LLMConfig()
	return m, func() tea.Msg { return catalogMsg{generation, llm.DiscoverModels(ctx, cfg)} }
}
func (m *model) receiveCatalog(msg catalogMsg) (tea.Model, tea.Cmd) {
	if m.picker == nil || m.picker.kind != "model" || msg.generation != m.pickerGeneration {
		return m, nil
	}
	p := m.picker
	p.loading = false
	cfg := m.proc.LLMConfig()
	for _, model := range msg.catalog.Models {
		label := model.Provider + " / " + model.ID
		if model.Provider == llm.ResolveProvider(cfg) && model.ID == cfg.Model {
			label += " (current)"
		}
		p.items = append(p.items, pickerItem{label: label, description: model.Name, value: model.ID, provider: model.Provider})
	}
	for _, warning := range msg.catalog.Warnings {
		m.appendError("Model discovery: " + warning)
	}
	p.filter()
	for i, index := range p.filtered {
		item := p.items[index]
		if item.provider == llm.ResolveProvider(cfg) && item.value == cfg.Model {
			p.selected = i
			break
		}
	}
	return m, nil
}

var effortLevels = []string{"", "minimal", "low", "medium", "high", "xhigh"}
var effortDescriptions = []string{"Use the provider default", "Least reasoning", "Quick responses", "Balance speed and reasoning", "More reasoning for difficult tasks", "Most reasoning; can take longer"}

func (m *model) openEffortPicker() (tea.Model, tea.Cmd) {
	m.closePicker()
	p := &selectionPicker{kind: "effort"}
	m.picker = p
	current := m.proc.LLMConfig().ReasoningEffort
	provider := llm.ResolveProvider(m.proc.LLMConfig())
	levels := effortLevels
	// These adapters currently do not send reasoning controls.
	if provider == "google" || provider == "anthropic" {
		levels = effortLevels[:1]
	}
	for i, level := range levels {
		label := level
		if label == "" {
			label = "provider default"
		}
		if level == current {
			label += " (current)"
			p.selected = i
		}
		p.items = append(p.items, pickerItem{label: label, description: effortDescriptions[i], value: level})
	}
	p.filter()
	for i, item := range p.items {
		if item.value == current {
			p.selected = i
		}
	}
	return m, effortTick(m.pickerGeneration)
}
func (m *model) setEffort(effort string) error {
	if effort == "default" {
		effort = ""
	}
	if !config.ValidReasoningEffort(effort) {
		return fmt.Errorf("invalid effort: use default, minimal, low, medium, high, or xhigh")
	}
	cfg := m.proc.LLMConfig()
	cfg.ReasoningEffort = effort
	if err := m.proc.ConfigureLLM(cfg); err != nil {
		return fmt.Errorf("could not set effort: %w", err)
	}
	saved := effort
	if saved == "" {
		saved = "default"
	}
	if _, err := config.Set(m.rootDir, "llm.reasoningEffort", saved); err != nil {
		return fmt.Errorf("effort changed for this session but could not be saved: %w", err)
	}
	label := effort
	if label == "" {
		label = "provider default"
	}
	m.appendSystem("Reasoning effort: " + effortMeter(effort, 0) + " " + strings.ToUpper(label))
	return nil
}
func (p *selectionPicker) filter() {
	p.filtered = nil
	p.selected = 0
	query := strings.ToLower(p.query)
	for i, item := range p.items {
		if strings.Contains(strings.ToLower(item.label+" "+item.description), query) {
			p.filtered = append(p.filtered, i)
		}
	}
}
func (m *model) handlePickerKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	p := m.picker
	n := len(p.filtered)
	switch msg.String() {
	case "esc", "ctrl+c":
		m.closePicker()
		return m, nil
	case "up":
		if n > 0 {
			p.selected = (p.selected + n - 1) % n
		}
	case "down":
		if n > 0 {
			p.selected = (p.selected + 1) % n
		}
	case "pgup":
		p.selected = maxInt(0, p.selected-m.menuLimit())
	case "pgdown":
		if n > 0 {
			p.selected = minInt(n-1, p.selected+m.menuLimit())
		}
	case "home":
		p.selected = 0
	case "end":
		if n > 0 {
			p.selected = n - 1
		}
	case "enter":
		if p.loading || n == 0 {
			return m, nil
		}
		item := p.items[p.filtered[p.selected]]
		var err error
		if p.kind == "effort" {
			err = m.setEffort(item.value)
		} else {
			var cfg config.LLM
			cfg, err = llm.ConfigForModel(m.proc.LLMConfig(), item.provider, item.value)
			if err == nil {
				err = m.proc.ConfigureLLM(cfg)
			}
			if err == nil {
				if saveErr := config.SetGlobalModel(cfg); saveErr != nil {
					m.appendError("Model changed for this session, but could not save it: " + saveErr.Error())
				}
				m.appendSystem("Model set to " + item.provider + " / " + item.value + ".")
			}
		}
		if err != nil {
			m.appendError(err.Error())
			return m, nil
		}
		m.closePicker()
	case "backspace", "ctrl+h":
		r := []rune(p.query)
		if len(r) > 0 {
			p.query = string(r[:len(r)-1])
			p.filter()
		}
	default:
		if msg.Type == tea.KeyRunes && p.kind == "model" {
			p.query += security.SanitizeLine(string(msg.Runes))
			p.filter()
		}
	}
	return m, nil
}
func (m *model) renderPicker() string {
	p := m.picker
	title := "Select Model"
	if p.kind == "effort" {
		title = "Select Reasoning Effort"
	}
	lines := []string{menuTitle.Render(title)}
	if p.loading {
		lines = append(lines, "Loading connected provider models…")
	} else {
		if p.kind == "model" {
			lines = append(lines, fmt.Sprintf("Filter: %s  (%d models)", p.query, len(p.filtered)))
		}
		start, end := visibleWindow(len(p.filtered), p.selected, m.menuLimit())
		for i := start; i < end; i++ {
			item := p.items[p.filtered[i]]
			label := item.label
			if p.kind == "effort" {
				frame := 0
				if i == p.selected {
					frame = p.frame
				}
				label = effortMeter(item.value, frame) + " " + label
			}
			line := fmt.Sprintf("  %d. %s", i+1, label)
			if m.width >= 100 && item.description != "" {
				line += "  " + item.description
			}
			if i == p.selected {
				line = menuSel.Render("›" + strings.TrimPrefix(line, " "))
			} else {
				line = menuNorm.Render(line)
			}
			lines = append(lines, line)
		}
		if len(p.filtered) == 0 {
			lines = append(lines, "No models found. Use /connect or change the filter.")
		}
	}
	lines = append(lines, faintStyle.Render("↑/↓ select · Enter apply · Esc back"))
	for i, line := range lines {
		lines[i] = ansi.Truncate(line, maxInt(1, m.width), "…")
	}
	return strings.Join(lines, "\n")
}
func effortMeter(effort string, frame int) string {
	n := 0
	for i, value := range effortLevels {
		if value == effort {
			n = i
			break
		}
	}
	bars := strings.Repeat("▰", n) + strings.Repeat("▱", 5-n)
	if effort == "xhigh" && frame%4 != 0 {
		r := []rune(bars)
		r[(frame-1)%5] = '▱'
		bars = string(r)
	}
	return bars
}
