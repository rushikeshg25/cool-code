package tui

import (
	"context"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/ansi"
)

func TestBTWDuringTurnDoesNotQueueOrChangeHistory(t *testing.T) {
	m := newTestModel(t)
	m.processing = true
	m.ti.SetValue("/btw")
	m.submit()
	if m.aside == nil || !m.processing || len(m.history) != 0 {
		t.Fatal("did not open independent window")
	}
	_, _, _, _, count := m.proc.Snapshot()
	if count != 0 {
		t.Fatal("polluted main history")
	}
	m.aside.ask = func(context.Context, string) (string, error) { return "**Side answer**", nil }
	m.aside.input.SetValue("Normal question")
	_, cmd := m.submitAside()
	m.Update(cmd())
	if m.aside.answer != "**Side answer**" || len(m.history) != 0 {
		t.Fatal("answer escaped side window")
	}
	m.handleAsideKey(tea.KeyMsg{Type: tea.KeyEsc})
	if m.aside != nil || !m.processing {
		t.Fatal("closing side window interrupted main turn")
	}
}
func TestAsideCancellationStaleResultAndResize(t *testing.T) {
	m := newTestModel(t)
	m.openAside("")
	ctx, cancel := context.WithCancel(context.Background())
	m.aside.cancel = cancel
	old := m.asideGeneration
	m.closeAside()
	if ctx.Err() == nil {
		t.Fatal("not cancelled")
	}
	m.openAside("")
	m.receiveAside(asideMsg{generation: old, answer: "stale"})
	if m.aside.answer != "" {
		t.Fatal("stale answer reopened window")
	}
	m.aside.question = "Q"
	m.aside.answer = strings.Repeat("A line of **Markdown**\n\n", 100)
	for _, width := range []int{30, 80, 120} {
		m.Update(tea.WindowSizeMsg{Width: width, Height: 24})
		for _, line := range strings.Split(m.renderAside(), "\n") {
			if ansi.StringWidth(line) > width {
				t.Fatal("overflow")
			}
		}
	}
}
