package tui

import (
	"context"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/rushikeshg25/cool-code/internal/security"
)

func (m *model) markQueuedSent(messages []string) {
	for _, message := range messages {
		for i := range m.history {
			if m.history[i].kind == entrySystem && m.history[i].raw == "(queued) "+message {
				m.history[i].raw = "(sent follow-up) " + message
				m.history[i].rendered = m.renderEntry(m.history[i])
				break
			}
		}
	}
	m.invalidatePrefix()
	m.syncViewport()
}

func (m *model) sendQueuedNow() (tea.Model, tea.Cmd) {
	// Do not submit a hidden main draft from another text-entry overlay.
	if !m.hasOverlay() {
		if draft := strings.TrimSpace(m.ti.Value()); draft != "" {
			m.proc.EnqueueMessage(draft)
			m.appendSystem("(queued) " + security.Redact(draft))
			m.ti.Reset()
			m.fitInputHeight()
			m.refreshSuggestions()
		}
	}
	messages := m.proc.TakeQueuedMessages()
	if len(messages) == 0 {
		return m, nil
	}
	m.followOutput = true
	if m.processing {
		m.sendNowMessages = append(m.sendNowMessages, messages...)
		// Cancel first, then release any synchronous approval callback. The worker
		// sees cancellation before it can execute a just-declined queued action.
		m.interruptTurn()
		if m.confirmResp != nil {
			m.respondConfirm(false)
		}
		m.answerQuestion("", context.Canceled)
		return m, nil
	}
	m.markQueuedSent(messages)
	m.planMenu = false
	return m.runQuery(strings.Join(messages, "\n"))
}
