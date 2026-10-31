package tui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/rushikeshg25/cool-code/internal/session"
)

func (m *model) forkSession() (tea.Model, tea.Cmd) {
	if m.processing {
		m.appendError("Wait for the running turn to finish before forking.")
		return m, nil
	}
	parent := m.sessionID
	if err := m.saveSession(parent); err != nil {
		m.appendError("Could not save the original chat: " + err.Error())
		return m, nil
	}
	child := session.NewID()
	if err := m.saveSession(child); err != nil {
		m.appendError("Could not create the fork: " + err.Error())
		return m, nil
	}
	m.sessionID = child
	m.appendSystem("Forked chat " + parent[:8] + " → " + child[:8] + ". You are now in the copy. Use /sessions to return to the original.")
	return m, nil
}
