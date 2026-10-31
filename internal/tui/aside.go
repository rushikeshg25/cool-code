package tui

import (
	"context"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/ansi"
	"github.com/rushikeshg25/cool-code/internal/security"
)

type asideDialog struct {
	input            textinput.Model
	question, answer string
	busy             bool
	offset           int
	cancel           context.CancelFunc
	ask              func(context.Context, string) (string, error)
}
type asideMsg struct {
	generation int
	answer     string
	err        error
}

func (m *model) openAside(question string) (tea.Model, tea.Cmd) {
	m.closeAside()
	input := textinput.New()
	input.Prompt = "❯ "
	input.Placeholder = "Ask a side question…"
	input.CharLimit = 4000
	cmd := input.Focus()
	m.aside = &asideDialog{input: input, ask: m.proc.PrepareAside()}
	if strings.TrimSpace(question) != "" {
		m.aside.input.SetValue(question)
		return m.submitAside()
	}
	return m, cmd
}
func (m *model) closeAside() {
	if m.aside != nil && m.aside.cancel != nil {
		m.aside.cancel()
	}
	m.aside = nil
	m.asideGeneration++
}
func (m *model) submitAside() (tea.Model, tea.Cmd) {
	a := m.aside
	if a.busy {
		return m, nil
	}
	question := strings.TrimSpace(a.input.Value())
	if question == "" {
		return m, nil
	}
	a.question = question
	a.answer = ""
	a.input.Reset()
	a.busy = true
	a.offset = 0
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	a.cancel = cancel
	ask, generation := a.ask, m.asideGeneration
	return m, func() tea.Msg {
		defer cancel()
		answer, err := ask(ctx, question)
		return asideMsg{generation, answer, err}
	}
}
func (m *model) receiveAside(msg asideMsg) (tea.Model, tea.Cmd) {
	if m.aside == nil || msg.generation != m.asideGeneration {
		return m, nil
	}
	a := m.aside
	a.busy = false
	a.cancel = nil
	if msg.err != nil {
		a.answer = "Could not answer: " + security.Redact(msg.err.Error())
	} else {
		a.answer = msg.answer
	}
	return m, nil
}
func (m *model) handleAsideKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	a := m.aside
	switch msg.String() {
	case "esc", "ctrl+c":
		m.closeAside()
		return m, nil
	case "enter":
		return m.submitAside()
	case "pgup", "up":
		a.offset = maxInt(0, a.offset-3)
		return m, nil
	case "pgdown", "down":
		a.offset += 3
		return m, nil
	}
	if a.busy {
		return m, nil
	}
	var cmd tea.Cmd
	a.input, cmd = a.input.Update(msg)
	return m, cmd
}
func (m *model) renderAside() string {
	a := m.aside
	width := maxInt(1, m.width-2)
	lines := []string{menuTitle.Render("BTW · side question")}
	body := "Ask a quick question. This does not change the main chat."
	if a.question != "" {
		body = "You: " + security.SanitizeTerminal(a.question) + "\n\n"
		if a.busy {
			body += "Thinking…"
		} else {
			body += a.answer
		}
	}
	rendered := strings.Split(renderMarkdown(security.SanitizeTerminal(body), width), "\n")
	limit := maxInt(2, minInt(16, m.height/2))
	a.offset = minInt(a.offset, maxInt(0, len(rendered)-limit))
	lines = append(lines, rendered[a.offset:minInt(len(rendered), a.offset+limit)]...)
	if !a.busy {
		a.input.Width = maxInt(1, width-3)
		lines = append(lines, a.input.View())
	}
	lines = append(lines, "Enter ask · PgUp/PgDn scroll · Esc close")
	for i := range lines {
		lines[i] = ansi.Truncate(lines[i], width, "…")
	}
	return strings.Join(lines, "\n")
}
