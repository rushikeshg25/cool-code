// Package tui implements the interactive Bubble Tea terminal UI.
package tui

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/rushikeshg25/cool-code/internal/agent"
	"github.com/rushikeshg25/cool-code/internal/session"
	"github.com/rushikeshg25/cool-code/internal/types"
)

// --- messages pushed from the processor goroutine ---

type statusMsg string
type assistantMsg string
type deltaMsg string
type toolMsg struct{ name, display string }
type tasksMsg struct{ list *types.TaskList }
type subagentsMsg struct{ lines []string }
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
func (b *bridge) AssistantDelta(t string)    { b.prog.Send(deltaMsg(t)) }
func (b *bridge) Assistant(md string)        { b.prog.Send(assistantMsg(md)) }
func (b *bridge) Tool(name, display string)  { b.prog.Send(toolMsg{name, display}) }
func (b *bridge) Tasks(list *types.TaskList) { b.prog.Send(tasksMsg{list}) }
func (b *bridge) Subagents(lines []string)   { b.prog.Send(subagentsMsg{lines}) }

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
	ti            textarea.Model
	sp            spinner.Model
	ready         bool

	history    []entry
	streamIdx  int // index of the in-progress streaming entry, -1 when none
	inputHist  []string
	histIdx    int
	processing bool
	status     string
	mode       types.AgentMode
	tasks      *types.TaskList
	subagents  []string
	cancelTurn context.CancelFunc

	suggestions []slashCommand
	suggestIdx  int

	confirmMsg  string
	confirmResp chan bool

	planMenu bool
	planIdx  int
}

func newModel(proc *agent.Processor, rootDir, version string, copyOut bool, sessionID string) *model {
	ti := textarea.New()
	ti.Placeholder = "Ask, plan, or build something…"
	ti.Prompt = "❯ "
	ti.FocusedStyle.Prompt = promptGlyph
	ti.ShowLineNumbers = false
	ti.SetHeight(1)
	ti.CharLimit = 0
	// Enter submits; Alt+Enter or Ctrl+J inserts a newline (Shift+Enter is
	// not distinguishable from Enter in classic terminal input).
	ti.KeyMap.InsertNewline.SetKeys("alt+enter", "ctrl+j")
	ti.Focus()

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
		streamIdx: -1,
	}
}

func (m *model) Init() tea.Cmd {
	return textarea.Blink
}

// --- history helpers ---

// entry is one transcript item: raw source plus its rendering at the current
// width, so history can be re-rendered on resize.
type entryKind int

const (
	entryRaw entryKind = iota // pre-styled, never re-rendered (banner)
	entryUser
	entryAssistant
	entryTool
	entrySystem
	entryStream // assistant text still streaming: plain wrap, no markdown
)

type entry struct {
	kind     entryKind
	raw      string
	rendered string
}

func (m *model) contentWidth() int {
	if m.width < 10 {
		return 80
	}
	return m.width
}

func (m *model) renderEntry(e entry) string {
	switch e.kind {
	case entryUser:
		return userPrefix.Render("❯ ") + userText.Render(e.raw)
	case entryAssistant:
		return renderMarkdown(e.raw, m.contentWidth())
	case entryTool:
		return toolGlyph.Render("  ● ") + toolStyle.Render(e.raw)
	case entrySystem:
		return systemStyle.Render(e.raw)
	case entryStream:
		return lipgloss.NewStyle().Width(m.contentWidth()).Render(e.raw)
	default:
		return e.raw
	}
}

// appendDelta grows the in-progress streaming entry (creating it on the first
// fragment). Cheap per fragment: plain wrapping, markdown renders once at the
// end via finishStream.
func (m *model) appendDelta(delta string) {
	if m.streamIdx < 0 {
		m.appendEntry(entryStream, delta)
		m.streamIdx = len(m.history) - 1
		return
	}
	e := &m.history[m.streamIdx]
	e.raw += delta
	e.rendered = m.renderEntry(*e)
	m.syncViewport()
}

// finishStream replaces the streaming entry with the final markdown rendering.
func (m *model) finishStream(md string) {
	if m.streamIdx < 0 {
		m.appendAssistant(md)
		return
	}
	m.history[m.streamIdx] = entry{kind: entryAssistant, raw: md}
	m.history[m.streamIdx].rendered = m.renderEntry(m.history[m.streamIdx])
	m.streamIdx = -1
	m.syncViewport()
}

func (m *model) appendEntry(kind entryKind, raw string) {
	e := entry{kind: kind, raw: raw}
	e.rendered = m.renderEntry(e)
	m.history = append(m.history, e)
	m.syncViewport()
}

func (m *model) appendRaw(s string)         { m.appendEntry(entryRaw, s) }
func (m *model) appendUser(text string)     { m.appendEntry(entryUser, text) }
func (m *model) appendAssistant(md string)  { m.appendEntry(entryAssistant, md) }
func (m *model) appendTool(display string)  { m.appendEntry(entryTool, display) }
func (m *model) appendSystem(text string)   { m.appendEntry(entrySystem, text) }

// rerenderHistory refreshes every entry's rendering, e.g. after a resize.
func (m *model) rerenderHistory() {
	for i := range m.history {
		if m.history[i].kind != entryRaw {
			m.history[i].rendered = m.renderEntry(m.history[i])
		}
	}
}

func (m *model) syncViewport() {
	if !m.ready {
		return
	}
	atBottom := m.vp.AtBottom()
	parts := make([]string, len(m.history))
	for i, e := range m.history {
		parts[i] = e.rendered
	}
	m.vp.SetContent(strings.Join(parts, "\n\n"))
	if atBottom {
		m.vp.GotoBottom()
	}
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
