package tui

import (
	"context"
	"errors"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/ansi"
	"github.com/rushikeshg25/cool-code/internal/security"
	"github.com/rushikeshg25/cool-code/internal/tools"
)

type questionReply struct {
	answer string
	err    error
}
type questionReqMsg struct {
	ctx      context.Context
	question tools.Question
	reply    chan questionReply
}
type questionDialog struct {
	request          questionReqMsg
	selected, offset int
	input            textinput.Model
}

func (b *bridge) askQuestion(ctx context.Context, q tools.Question) (string, error) {
	ch := make(chan questionReply, 1)
	b.prog.Send(questionReqMsg{ctx, q, ch})
	select {
	case r := <-ch:
		return r.answer, r.err
	case <-ctx.Done():
		return "", ctx.Err()
	}
}
func (m *model) receiveQuestion(req questionReqMsg) (tea.Model, tea.Cmd) {
	if req.ctx.Err() != nil {
		return m, nil
	}
	input := textinput.New()
	input.Prompt = "Answer: "
	input.Placeholder = "Type your answer"
	input.CharLimit = 2000
	cmd := input.Focus()
	m.question = &questionDialog{request: req, input: input}
	return m, cmd
}
func (m *model) answerQuestion(answer string, err error) {
	if m.question == nil {
		return
	}
	select {
	case m.question.request.reply <- questionReply{answer, err}:
	default:
	}
	m.question = nil
}
func (m *model) handleQuestionKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	q := m.question
	switch msg.String() {
	case "esc":
		m.answerQuestion("", errors.New("user dismissed the question"))
		return m, nil
	case "ctrl+c":
		m.answerQuestion("", context.Canceled)
		m.interruptTurn()
		return m, nil
	case "up":
		q.selected = maxInt(0, q.selected-1)
	case "down", "tab":
		q.selected = (q.selected + 1) % (len(q.request.question.Options) + 1)
	case "pgup":
		q.offset = maxInt(0, q.offset-3)
	case "pgdown":
		q.offset += 3
	case "enter":
		if q.selected < len(q.request.question.Options) {
			m.answerQuestion(q.request.question.Options[q.selected], nil)
		} else if answer := strings.TrimSpace(q.input.Value()); answer != "" {
			m.answerQuestion(answer, nil)
		}
	default:
		if q.selected == len(q.request.question.Options) {
			var cmd tea.Cmd
			q.input, cmd = q.input.Update(msg)
			return m, cmd
		}
	}
	return m, nil
}
func (m *model) renderQuestion() string {
	q := m.question
	width := maxInt(1, m.width-2)
	prompt := strings.Split(ansi.Wordwrap(security.SanitizeTerminal(q.request.question.Prompt), width, " "), "\n")
	limit := maxInt(1, minInt(4, m.height/4))
	q.offset = minInt(q.offset, maxInt(0, len(prompt)-limit))
	lines := []string{menuTitle.Render("Question")}
	lines = append(lines, prompt[q.offset:minInt(len(prompt), q.offset+limit)]...)
	options := append(append([]string(nil), q.request.question.Options...), "Write my own answer")
	start, end := visibleWindow(len(options), q.selected, maxInt(1, minInt(6, m.height/3)))
	for i := start; i < end; i++ {
		line := "  " + security.SanitizeLine(options[i])
		if i == q.selected {
			line = menuSel.Render("› " + security.SanitizeLine(options[i]))
		}
		lines = append(lines, line)
	}
	if q.selected == len(q.request.question.Options) {
		q.input.Width = maxInt(1, width-8)
		lines = append(lines, q.input.View())
	}
	lines = append(lines, "↑/↓ choose · Enter answer · Esc dismiss", "PgUp/PgDn scroll question")
	for i := range lines {
		lines[i] = ansi.Truncate(lines[i], width, "…")
	}
	return strings.Join(lines, "\n")
}
