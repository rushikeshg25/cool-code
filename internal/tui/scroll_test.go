package tui

import (
	tea "github.com/charmbracelet/bubbletea"
	"strings"
	"testing"
)

func TestFollowOutputSurvivesComposerResize(t *testing.T) {
	m := newTestModel(t)
	m.appendAssistant(strings.Repeat("line\n\n", 80))
	m.View()
	if !m.vp.AtBottom() {
		t.Fatal("initial footer layout left latest output off screen")
	}
	m.ti.SetValue("first\nsecond\nthird")
	m.fitInputHeight()
	m.View()
	m.appendDelta("latest output")
	m.View()
	if !m.vp.AtBottom() {
		t.Fatal("composer growth disabled auto-scroll")
	}
	m.Update(tea.WindowSizeMsg{Width: 80, Height: 18})
	m.View()
	if !m.vp.AtBottom() {
		t.Fatal("terminal resize disabled auto-scroll")
	}
}
func TestManualScrollPausesFollowOutput(t *testing.T) {
	m := newTestModel(t)
	m.appendAssistant(strings.Repeat("line\n\n", 80))
	m.View()
	m.vp.GotoBottom()
	m.Update(tea.KeyMsg{Type: tea.KeyPgUp})
	offset := m.vp.YOffset
	m.appendDelta("new output\n")
	m.View()
	if m.vp.YOffset != offset {
		t.Fatal("new output moved user away from reading position")
	}
	m.Update(tea.KeyMsg{Type: tea.KeyEnd, Alt: true})
	m.View()
	if !m.vp.AtBottom() {
		t.Fatal("Alt+End did not return to latest output")
	}
	m.appendDelta("more output\n")
	m.View()
	if !m.vp.AtBottom() {
		t.Fatal("Alt+End did not resume follow")
	}
}
