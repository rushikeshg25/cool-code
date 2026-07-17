// Package tui implements the interactive Bubble Tea terminal UI.
package tui

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/rushikeshg25/cool-code/internal/agent"
	"github.com/rushikeshg25/cool-code/internal/session"
	"github.com/rushikeshg25/cool-code/internal/types"
)

// --- messages pushed from the processor goroutine ---

type statusMsg string
type assistantMsg string
type toolMsg struct{ name, display string }
type tasksMsg struct{ list *types.TaskList }
type doneMsg struct {
	final string
	err   error
}
type confirmReqMsg struct {
	message string
	resp    chan bool
}

// bridge implements agent.Reporter and the confirm callbacks by forwarding to
// the running tea.Program.
type bridge struct{ prog *tea.Program }

func (b *bridge) Status(t string)            { b.prog.Send(statusMsg(t)) }
func (b *bridge) Assistant(md string)        { b.prog.Send(assistantMsg(md)) }
func (b *bridge) Tool(name, display string)  { b.prog.Send(toolMsg{name, display}) }
func (b *bridge) Tasks(list *types.TaskList) { b.prog.Send(tasksMsg{list}) }

func (b *bridge) confirm(message string) bool {
	ch := make(chan bool, 1)
	b.prog.Send(confirmReqMsg{message: message, resp: ch})
	return <-ch
}

func (b *bridge) confirmEdit(message, preview string) bool {
	return b.confirm(message + "\n" + preview)
}

var planOptions = []string{
	"Start implementation (switch to Agent and proceed)",
	"Keep planning (talk / refine the plan)",
}

type model struct {
	proc      *agent.Processor
	bridge    *bridge
	rootDir   string
	version   string
	copy      bool
	sessionID string

	width, height int
	vp            viewport.Model
	ti            textinput.Model
	sp            spinner.Model
	ready         bool

	history    []string
	processing bool
	status     string
	mode       types.AgentMode
	tasks      *types.TaskList
	cancelTurn context.CancelFunc

	suggestions []slashCommand
	suggestIdx  int

	confirmMsg  string
	confirmResp chan bool

	planMenu bool
	planIdx  int
}

func newModel(proc *agent.Processor, rootDir, version string, copyOut bool, sessionID string) *model {
	ti := textinput.New()
	ti.Placeholder = "Ask, plan, or build something…"
	ti.Prompt = "❯ "
	ti.PromptStyle = promptGlyph
	ti.Focus()
	ti.CharLimit = 0

	sp := spinner.New()
	sp.Spinner = spinner.Dot
	sp.Style = spinnerStyle

	return &model{
		proc:      proc,
		rootDir:   rootDir,
		version:   version,
		copy:      copyOut,
		sessionID: sessionID,
		ti:        ti,
		sp:        sp,
		mode:      proc.Mode(),
		tasks:     proc.TaskList(),
	}
}

func (m *model) Init() tea.Cmd {
	return textinput.Blink
}

// --- history helpers ---

func (m *model) contentWidth() int {
	if m.width < 10 {
		return 80
	}
	return m.width
}

func (m *model) appendRaw(s string) {
	m.history = append(m.history, s)
	m.syncViewport()
}

func (m *model) appendUser(text string) {
	m.appendRaw(userPrefix.Render("❯ ") + userText.Render(text))
}

func (m *model) appendAssistant(md string) {
	m.appendRaw(renderMarkdown(md, m.contentWidth()))
}

func (m *model) appendTool(display string) {
	m.appendRaw(toolGlyph.Render("  ● ") + toolStyle.Render(display))
}

func (m *model) appendSystem(text string) {
	m.appendRaw(systemStyle.Render(text))
}

func (m *model) syncViewport() {
	if !m.ready {
		return
	}
	m.vp.SetContent(strings.Join(m.history, "\n\n"))
	m.vp.GotoBottom()
}

func (m *model) persist() {
	messages, summary, pinned, mode, count := m.proc.Snapshot()
	raw, err := json.Marshal(messages)
	if err != nil {
		return
	}
	_ = session.Save(session.Data{
		ID:           m.sessionID,
		Cwd:          m.rootDir,
		UpdatedAt:    time.Now().UTC().Format(time.RFC3339Nano),
		Mode:         string(mode),
		Messages:     raw,
		Summary:      summary,
		PinnedFiles:  pinned,
		MessageCount: count,
	})
}
