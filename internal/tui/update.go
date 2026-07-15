package tui

import (
	"context"
	"strings"

	"github.com/atotto/clipboard"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
)

func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.ti.Width = msg.Width - 4
		if !m.ready {
			m.vp = newViewport(msg.Width, m.viewportHeight())
			m.ready = true
		} else {
			m.vp.Width = msg.Width
			m.vp.Height = m.viewportHeight()
		}
		m.syncViewport()
		return m, nil

	case statusMsg:
		m.status = string(msg)
		return m, nil

	case assistantMsg:
		m.status = ""
		m.appendAssistant(string(msg))
		return m, nil

	case toolMsg:
		m.appendTool(msg.display)
		return m, nil

	case tasksMsg:
		m.tasks = msg.list
		return m, nil

	case confirmReqMsg:
		m.status = ""
		m.confirmMsg = msg.message
		m.confirmResp = msg.resp
		return m, nil

	case doneMsg:
		return m.handleDone(msg)

	case spinner.TickMsg:
		if m.processing {
			var cmd tea.Cmd
			m.sp, cmd = m.sp.Update(msg)
			return m, cmd
		}
		return m, nil

	case tea.KeyMsg:
		return m.handleKey(msg)
	}

	// Forward everything else (mouse, etc.) to the viewport for scrolling.
	if m.ready {
		var cmd tea.Cmd
		m.vp, cmd = m.vp.Update(msg)
		return m, cmd
	}
	return m, nil
}

func (m *model) handleDone(msg doneMsg) (tea.Model, tea.Cmd) {
	m.processing = false
	m.status = ""
	if msg.err != nil {
		m.appendSystem("Error: " + msg.err.Error())
		return m, nil
	}
	m.tasks = m.proc.TaskList()
	m.persist()
	if m.copy && msg.final != "" {
		if err := clipboard.WriteAll(msg.final); err != nil {
			m.appendSystem("Copy failed: " + err.Error())
		} else {
			m.appendSystem("Copied to clipboard.")
		}
	}
	if m.proc.Mode() == "plan" && msg.final != "" {
		m.planMenu = true
		m.planIdx = 0
	}
	return m, nil
}

func (m *model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Confirmation overlay takes priority.
	if m.confirmMsg != "" {
		switch strings.ToLower(msg.String()) {
		case "y":
			m.respondConfirm(true)
		case "n", "esc", "enter":
			m.respondConfirm(false)
		}
		return m, nil
	}

	// Post-plan action menu.
	if m.planMenu {
		return m.handlePlanMenu(msg)
	}

	switch msg.Type {
	case tea.KeyCtrlC:
		return m, tea.Quit
	case tea.KeyShiftTab:
		next := nextMode(m.proc.Mode())
		m.proc.SetMode(next)
		m.mode = next
		return m, nil
	case tea.KeyTab:
		if len(m.suggestions) > 0 {
			m.ti.SetValue(m.suggestions[m.suggestIdx%len(m.suggestions)].name + " ")
			m.ti.CursorEnd()
			m.suggestIdx = 0
			m.refreshSuggestions()
		}
		return m, nil
	case tea.KeyUp:
		if len(m.suggestions) > 0 {
			m.suggestIdx = (m.suggestIdx - 1 + len(m.suggestions)) % len(m.suggestions)
			return m, nil
		}
		var cmd tea.Cmd
		m.vp, cmd = m.vp.Update(msg)
		return m, cmd
	case tea.KeyDown:
		if len(m.suggestions) > 0 {
			m.suggestIdx = (m.suggestIdx + 1) % len(m.suggestions)
			return m, nil
		}
		var cmd tea.Cmd
		m.vp, cmd = m.vp.Update(msg)
		return m, cmd
	case tea.KeyEnter:
		return m.submit()
	}

	var cmd tea.Cmd
	m.ti, cmd = m.ti.Update(msg)
	m.refreshSuggestions()
	return m, cmd
}

func (m *model) handlePlanMenu(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "up":
		m.planIdx = (m.planIdx + len(planOptions) - 1) % len(planOptions)
	case "down":
		m.planIdx = (m.planIdx + 1) % len(planOptions)
	case "1":
		return m.startImplementation()
	case "2", "esc":
		m.planMenu = false
	case "enter":
		if m.planIdx == 0 {
			return m.startImplementation()
		}
		m.planMenu = false
	}
	return m, nil
}

func (m *model) startImplementation() (tea.Model, tea.Cmd) {
	m.planMenu = false
	m.proc.SetMode("agent")
	m.mode = "agent"
	return m.runQuery("The plan above is approved. Proceed with implementing it now.")
}

func (m *model) respondConfirm(v bool) {
	if m.confirmResp != nil {
		m.confirmResp <- v
		m.confirmResp = nil
	}
	m.confirmMsg = ""
}

func (m *model) refreshSuggestions() {
	if m.processing {
		m.suggestions = nil
		return
	}
	m.suggestions = matchCommands(m.ti.Value())
	if m.suggestIdx >= len(m.suggestions) {
		m.suggestIdx = 0
	}
}

func (m *model) submit() (tea.Model, tea.Cmd) {
	value := strings.TrimSpace(m.ti.Value())
	m.ti.SetValue("")
	m.suggestions = nil
	m.suggestIdx = 0
	if value == "" {
		return m, nil
	}
	if m.processing {
		m.proc.EnqueueMessage(value)
		m.appendSystem("(queued) " + value)
		return m, nil
	}
	if strings.HasPrefix(value, "/") {
		return m.dispatchCommand(value)
	}
	return m.runQuery(value)
}

func (m *model) runQuery(text string) (tea.Model, tea.Cmd) {
	m.processing = true
	m.appendUser(text)
	m.status = "Thinking…"
	proc := m.proc
	br := m.bridge
	run := func() tea.Msg {
		final, err := proc.ProcessQuery(context.Background(), text, br)
		return doneMsg{final: final, err: err}
	}
	return m, tea.Batch(run, m.sp.Tick)
}
