package tui

import (
	"context"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/ansi"
	"github.com/rushikeshg25/cool-code/internal/tools"
	"strings"
	"testing"
)

func TestQuestionChoiceAndCustomAnswer(t *testing.T) {
	m := newTestModel(t)
	req := questionReqMsg{context.Background(), tools.Question{Prompt: "Which?", Options: []string{"One", "Two"}}, make(chan questionReply, 1)}
	m.receiveQuestion(req)
	select {
	case <-req.reply:
		t.Fatal("selected without input")
	default:
	}
	m.handleQuestionKey(tea.KeyMsg{Type: tea.KeyDown})
	m.handleQuestionKey(tea.KeyMsg{Type: tea.KeyEnter})
	if r := <-req.reply; r.answer != "Two" || r.err != nil {
		t.Fatal(r)
	}
	m.receiveQuestion(req)
	m.handleQuestionKey(tea.KeyMsg{Type: tea.KeyDown})
	m.handleQuestionKey(tea.KeyMsg{Type: tea.KeyDown})
	m.handleQuestionKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("Custom")})
	m.handleQuestionKey(tea.KeyMsg{Type: tea.KeyEnter})
	if r := <-req.reply; r.answer != "Custom" {
		t.Fatal(r)
	}
	m.receiveQuestion(req)
	m.handleQuestionKey(tea.KeyMsg{Type: tea.KeyEsc})
	if r := <-req.reply; r.err == nil || r.answer != "" {
		t.Fatal("dismissal guessed answer")
	}
}
func TestQuestionCancelledAndNarrowLayout(t *testing.T) {
	m := newTestModel(t)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	req := questionReqMsg{ctx, tools.Question{Prompt: "Cancelled"}, make(chan questionReply, 1)}
	m.receiveQuestion(req)
	if m.question != nil {
		t.Fatal("stale request reopened")
	}
	req.ctx = context.Background()
	req.question.Prompt = strings.Repeat("Long question ", 100)
	m.receiveQuestion(req)
	for _, width := range []int{30, 80, 120} {
		m.Update(tea.WindowSizeMsg{Width: width, Height: 24})
		for _, line := range strings.Split(m.renderQuestion(), "\n") {
			if ansi.StringWidth(line) > width {
				t.Fatal("overflow")
			}
		}
	}
	m.handleDone(doneMsg{err: context.Canceled})
	if m.question != nil {
		t.Fatal("dialog left after cancellation")
	}
}
