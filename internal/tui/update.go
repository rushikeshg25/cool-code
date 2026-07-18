package tui

import (
	"context"
	"errors"
	"strings"

	"github.com/atotto/clipboard"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/rushikeshg25/cool-code/internal/config"
	"github.com/rushikeshg25/cool-code/internal/creds"
)

func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		widthChanged := m.width != msg.Width
		m.width = msg.Width
		m.height = msg.Height
		m.ti.SetWidth(msg.Width - 4)
		if !m.ready {
			m.vp = newViewport(msg.Width, m.viewportHeight())
			m.ready = true
		} else {
			m.vp.Width = msg.Width
			m.vp.Height = m.viewportHeight()
		}
		if widthChanged {
			m.rerenderHistory()
		}
		m.syncViewport()
		return m, nil

	case statusMsg:
		m.status = string(msg)
		return m, nil

	case deltaMsg:
		m.appendDelta(string(msg))
		return m, nil

	case assistantMsg:
		m.status = ""
		m.finishStream(string(msg))
		return m, nil

	case toolMsg:
		m.appendTool(msg.display)
		return m, nil

	case tasksMsg:
		m.tasks = msg.list
		return m, nil

	case subagentsMsg:
		m.subagents = msg.lines
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
	m.cancelTurn = nil
	m.subagents = nil
	m.streamIdx = -1
	if errors.Is(msg.err, context.Canceled) {
		m.appendSystem("Cancelled.")
		m.persist()
		return m, nil
	}
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

	// /connect key entry, then provider menu.
	if m.connectFor >= 0 {
		return m.handleConnectKey(msg)
	}
	if m.connectMenu {
		return m.handleConnectMenu(msg)
	}

	switch msg.Type {
	case tea.KeyCtrlC:
		if m.processing {
			m.interruptTurn()
			return m, nil
		}
		return m.quit()
	case tea.KeyEsc:
		if m.processing {
			m.interruptTurn()
		}
		return m, nil
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
		// Recall older input when the cursor is on the first line.
		if m.ti.Line() == 0 && m.histIdx > 0 {
			m.histIdx--
			m.setInput(m.inputHist[m.histIdx])
			return m, nil
		}
	case tea.KeyDown:
		if len(m.suggestions) > 0 {
			m.suggestIdx = (m.suggestIdx + 1) % len(m.suggestions)
			return m, nil
		}
		// Recall newer input when the cursor is on the last line.
		if m.ti.Line() == m.ti.LineCount()-1 && m.histIdx < len(m.inputHist) {
			m.histIdx++
			if m.histIdx == len(m.inputHist) {
				m.setInput("")
			} else {
				m.setInput(m.inputHist[m.histIdx])
			}
			return m, nil
		}
	case tea.KeyPgUp, tea.KeyPgDown:
		var cmd tea.Cmd
		m.vp, cmd = m.vp.Update(msg)
		return m, cmd
	case tea.KeyEnter:
		if !msg.Alt {
			// When the slash-command dropdown is open and no arguments have
			// been typed, Enter accepts the highlighted command and runs it.
			// Typed arguments (e.g. "/pin foo.go") are preserved as-is.
			if len(m.suggestions) > 0 {
				if fields := strings.Fields(m.ti.Value()); len(fields) <= 1 {
					m.ti.SetValue(m.suggestions[m.suggestIdx%len(m.suggestions)].name)
				}
			}
			return m.submit()
		}
		// Alt+Enter falls through to the textarea's newline binding.
	}

	var cmd tea.Cmd
	m.ti, cmd = m.ti.Update(msg)
	m.fitInputHeight()
	m.refreshSuggestions()
	return m, cmd
}

// setInput replaces the input contents (history recall).
func (m *model) setInput(value string) {
	m.ti.SetValue(value)
	m.fitInputHeight()
	m.refreshSuggestions()
}

// fitInputHeight grows the input with its content, up to 5 lines.
func (m *model) fitInputHeight() {
	h := m.ti.LineCount()
	if h > 5 {
		h = 5
	}
	if h < 1 {
		h = 1
	}
	m.ti.SetHeight(h)
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

func (m *model) handleConnectMenu(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "up":
		m.connectIdx = (m.connectIdx + len(connectOptions) - 1) % len(connectOptions)
	case "down":
		m.connectIdx = (m.connectIdx + 1) % len(connectOptions)
	case "esc":
		m.connectMenu = false
	case "1", "2", "3", "4", "5":
		m.connectIdx = int(msg.String()[0] - '1')
		return m.selectConnectOption()
	case "enter":
		return m.selectConnectOption()
	}
	return m, nil
}

func (m *model) selectConnectOption() (tea.Model, tea.Cmd) {
	opt := connectOptions[m.connectIdx]
	m.connectMenu = false
	if opt.provider == "" {
		m.appendSystem("Subscription sign-in isn't supported yet — pick an API key option.")
		return m, nil
	}
	m.connectFor = m.connectIdx
	m.keyInput.Reset()
	return m, m.keyInput.Focus()
}

func (m *model) handleConnectKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEsc:
		m.connectFor = -1
		m.keyInput.Blur()
		m.appendSystem("Connect cancelled.")
		return m, nil
	case tea.KeyEnter:
		opt := connectOptions[m.connectFor]
		key := strings.TrimSpace(m.keyInput.Value())
		m.connectFor = -1
		m.keyInput.Blur()
		m.keyInput.Reset()
		if key == "" {
			m.appendSystem("Connect cancelled (empty key).")
			return m, nil
		}
		if err := creds.SetAPIKey(opt.provider, key); err != nil {
			m.appendSystem("Failed to store key: " + err.Error())
			return m, nil
		}
		if err := config.SetGlobalLLM(config.LLM{Model: opt.model, Provider: opt.provider}); err != nil {
			m.appendSystem("Key stored, but saving global settings failed: " + err.Error())
		}
		cfg := config.Load(m.rootDir)
		if err := m.proc.ConfigureLLM(cfg.LLM); err != nil {
			m.appendSystem("Key stored, but provider setup failed: " + err.Error())
			return m, nil
		}
		m.appendSystem("Connected " + opt.provider + " — model " + cfg.LLM.Model + ". Key stored in " + creds.Path() + ".")
		return m, nil
	}
	var cmd tea.Cmd
	m.keyInput, cmd = m.keyInput.Update(msg)
	return m, cmd
}

func (m *model) respondConfirm(v bool) {
	if m.confirmResp != nil {
		m.confirmResp <- v
		m.confirmResp = nil
	}
	m.confirmMsg = ""
}

func (m *model) refreshSuggestions() {
	if m.processing || m.ti.LineCount() > 1 {
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
	m.ti.Reset()
	m.ti.SetHeight(1)
	m.suggestions = nil
	m.suggestIdx = 0
	if value == "" {
		return m, nil
	}
	m.inputHist = append(m.inputHist, value)
	m.histIdx = len(m.inputHist)
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
	ctx, cancel := context.WithCancel(context.Background())
	m.cancelTurn = cancel
	proc := m.proc
	br := m.bridge
	run := func() tea.Msg {
		final, err := proc.ProcessQuery(ctx, text, br)
		return doneMsg{final: final, err: err}
	}
	return m, tea.Batch(run, m.sp.Tick)
}

// interruptTurn cancels the in-flight turn; handleDone reports the outcome.
func (m *model) interruptTurn() {
	if m.cancelTurn != nil {
		m.status = "Cancelling…"
		m.cancelTurn()
	}
}

// quit persists the session before exiting so nothing is lost on Ctrl+C.
func (m *model) quit() (tea.Model, tea.Cmd) {
	m.persist()
	return m, tea.Quit
}
