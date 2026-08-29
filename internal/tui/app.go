// Package tui implements the interactive Bubble Tea terminal UI.
package tui

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/ansi"

	"github.com/rushikeshg25/cool-code/internal/agent"
	"github.com/rushikeshg25/cool-code/internal/llm"
	"github.com/rushikeshg25/cool-code/internal/security"
	"github.com/rushikeshg25/cool-code/internal/session"
	"github.com/rushikeshg25/cool-code/internal/types"
)

// --- messages pushed from the processor goroutine ---

type statusMsg string
type assistantMsg string
type deltaMsg string
type discardStreamMsg struct{}
type compactedMsg struct{ summary string }
type toolMsg struct {
	name, display string
	failed        bool
}
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

func (b *bridge) Status(t string)         { b.prog.Send(statusMsg(t)) }
func (b *bridge) AssistantDelta(t string) { b.prog.Send(deltaMsg(t)) }
func (b *bridge) Assistant(md string)     { b.prog.Send(assistantMsg(md)) }
func (b *bridge) AssistantDiscard()       { b.prog.Send(discardStreamMsg{}) }
func (b *bridge) Compacted(note string)   { b.prog.Send(compactedMsg{note}) }
func (b *bridge) Tool(name, display string, failed bool) {
	b.prog.Send(toolMsg{name, display, failed})
}
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
	layout        layout
	vp            viewport.Model
	ti            textarea.Model
	sp            spinner.Model
	ready         bool

	history     []entry
	prefix      string // cached rendering of every entry but the last
	prefixCount int
	streamIdx   int // index of the in-progress streaming entry, -1 when none
	inputHist   []string
	histIdx     int
	processing  bool
	status      string
	mode        types.AgentMode
	tasks       *types.TaskList
	subagents   []string
	cancelTurn  context.CancelFunc

	suggestions []slashCommand
	suggestIdx  int
	suggestMode suggestKind
	fileCache   []string // project-relative paths for @-mention completion

	confirmMsg  string
	confirmResp chan bool
	confirmOff  int

	planMenu bool
	planIdx  int

	sessionMenu bool
	sessionIdx  int
	sessionList []session.Data

	connectMenu bool
	connectIdx  int
	connectFor  int // index into connectOptions awaiting key entry, -1 = none
	keyInput    textinput.Model
}

// connectOption is one /connect choice.
type connectOption struct {
	label    string
	provider string // "" = not yet supported
	model    string // global default model applied on connect
}

var connectOptions = []connectOption{
	{"Claude API key (Anthropic)", "anthropic", "claude-sonnet-4-5"},
	{"OpenAI API key", "openai", "gpt-5"},
	{"Gemini API key (Google)", "google", "gemini-2.5-flash"},
	{"Claude Pro/Max subscription - coming soon", "", ""},
	{"ChatGPT/Codex subscription - coming soon", "", ""},
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

	ki := textinput.New()
	ki.Placeholder = "paste API key"
	ki.Prompt = "❯ "
	ki.EchoMode = textinput.EchoPassword

	return &model{
		proc:       proc,
		rootDir:    rootDir,
		version:    version,
		copy:       copyOut,
		sessionID:  sessionID,
		ti:         ti,
		sp:         sp,
		mode:       proc.Mode(),
		tasks:      proc.TaskList(),
		streamIdx:  -1,
		connectFor: -1,
		keyInput:   ki,
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
	entryUser entryKind = iota
	entryAssistant
	entryPlan
	entryTool
	entryToolFailed
	entrySystem
	entryError
	entryStream // assistant text still streaming: plain wrap, no markdown
)

type entry struct {
	kind     entryKind
	raw      string
	rendered string
}

func (m *model) contentWidth() int {
	if m.layout.transcriptWidth > 0 {
		return m.layout.contentWidth()
	}
	if m.width < 10 {
		return 80
	}
	return m.width
}

func (m *model) renderEntry(e entry) string {
	// Every entry carries text from a model, a tool or a repository, so its
	// escape sequences are stripped before it can reach the terminal.
	raw := security.SanitizeTerminal(e.raw)
	switch e.kind {
	case entryUser:
		// Wrap, and indent the continuation under the sigil. A long single
		// line of user input used to run past the terminal width unwrapped.
		wrapped := ansi.Wordwrap(raw, maxInt(10, m.contentWidth()-2), " /")
		lines := strings.Split(wrapped, "\n")
		for i, line := range lines {
			if i == 0 {
				lines[i] = userPrefix.Render("› ") + userText.Render(line)
				continue
			}
			lines[i] = "  " + userText.Render(line)
		}
		return strings.Join(lines, "\n")
	case entryAssistant:
		return renderMarkdown(raw, m.contentWidth())
	case entryPlan:
		body := planCard.Render(renderMarkdown(raw, maxInt(20, m.contentWidth()-3)))
		return planTitle.Render("◆ PLAN READY") + "\n" + body
	case entryTool:
		return toolGlyph.Render("  ├─ ") + toolStyle.Render(raw)
	case entryToolFailed:
		return errorGlyph.Render("  ✗  ") + toolFailure.Render(raw)
	case entrySystem:
		return systemStyle.Render(raw)
	case entryError:
		return errorGlyph.Render("⚠ ") + errorStyle.Render(raw)
	case entryStream:
		return ansi.Wordwrap(raw, m.contentWidth(), " /")
	default:
		return raw
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

// discardStream removes the in-progress streaming entry. Text streamed before
// a tool call is the model reasoning aloud; the tool lines that follow are the
// record of what actually happened.
func (m *model) discardStream() {
	if m.streamIdx < 0 {
		return
	}
	m.history = append(m.history[:m.streamIdx], m.history[m.streamIdx+1:]...)
	m.streamIdx = -1
	m.invalidatePrefix()
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
	m.extendPrefix()
	m.syncViewport()
}

func (m *model) appendUser(text string)    { m.appendEntry(entryUser, text) }
func (m *model) appendAssistant(md string) { m.appendEntry(entryAssistant, md) }
func (m *model) appendTool(display string) { m.appendEntry(entryTool, display) }
func (m *model) appendSystem(text string)  { m.appendEntry(entrySystem, text) }
func (m *model) appendError(text string)   { m.appendEntry(entryError, text) }

// appendToolResult draws a failed call differently from a successful one.
// Both used to render as the same muted branch line.
func (m *model) appendToolResult(display string, failed bool) {
	if failed {
		m.appendEntry(entryToolFailed, display)
		return
	}
	m.appendEntry(entryTool, display)
}

// rerenderHistory refreshes every entry's rendering, e.g. after a resize.
func (m *model) rerenderHistory() {
	for i := range m.history {
		m.history[i].rendered = m.renderEntry(m.history[i])
	}
	m.invalidatePrefix()
}

// repopulateTranscript rebuilds the visible transcript from the processor's
// current message history, so a resumed conversation is shown and not just
// restored into context. Shared by startup resume and in-session /sessions.
func (m *model) repopulateTranscript() {
	messages, _, _, _, _ := m.proc.Snapshot()
	for _, msg := range messages {
		switch msg.Role {
		case llm.RoleUser:
			if !strings.HasPrefix(msg.Text, "[SYSTEM:") {
				m.appendUser(msg.Text)
			}
		case llm.RoleAssistant:
			if msg.Text != "" {
				m.appendAssistant(msg.Text)
			}
		}
	}
	if m.proc.Mode() == types.ModePlan {
		m.promoteLastPlan("")
	}
}

// promoteLastPlan turns the final assistant response in a Plan-mode turn into
// a visually distinct plan card. final may be empty when restoring a session.
func (m *model) promoteLastPlan(final string) {
	for i := len(m.history) - 1; i >= 0; i-- {
		entry := &m.history[i]
		if entry.kind != entryAssistant || (final != "" && entry.raw != final) {
			continue
		}
		entry.kind = entryPlan
		entry.rendered = m.renderEntry(*entry)
		m.invalidatePrefix()
		m.syncViewport()
		return
	}
}

// syncViewport rebuilds the transcript. The joined prefix of every entry
// before the last is cached, because appendDelta calls this once per streamed
// token and rebuilding the whole history each time is O(history) per token.
func (m *model) syncViewport() {
	if !m.ready {
		return
	}
	atBottom := m.vp.AtBottom()

	if m.prefixCount > len(m.history)-1 {
		// History shrank or was rewritten; drop the cache.
		m.prefixCount = 0
		m.prefix = ""
	}
	var b strings.Builder
	b.WriteString(m.prefix)
	for i := m.prefixCount; i < len(m.history); i++ {
		if i > 0 {
			b.WriteString(entrySeparator(m.history[i].kind, m.history[i-1].kind))
		}
		b.WriteString(m.history[i].rendered)
	}
	m.vp.SetContent(b.String())
	if atBottom {
		m.vp.GotoBottom()
	}
}

// extendPrefix folds every entry but the last into the cached prefix. Only
// the last entry can still change, since that is the one being streamed into.
func (m *model) extendPrefix() {
	var b strings.Builder
	b.WriteString(m.prefix)
	for i := m.prefixCount; i < len(m.history)-1; i++ {
		if i > 0 {
			b.WriteString(entrySeparator(m.history[i].kind, m.history[i-1].kind))
		}
		b.WriteString(m.history[i].rendered)
	}
	m.prefix = b.String()
	m.prefixCount = maxInt(0, len(m.history)-1)
}

// invalidatePrefix forces a full rebuild, for anything that rewrites history.
func (m *model) invalidatePrefix() {
	m.prefix = ""
	m.prefixCount = 0
}

func entrySeparator(kind, prev entryKind) string {
	if isToolEntry(kind) && isToolEntry(prev) {
		return "\n"
	}
	return "\n\n"
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
		ExtraDirs:    m.proc.ExtraDirs(),
		MessageCount: count,
	})
}

// isToolEntry reports whether a kind is one of the compact tool lines, which
// are joined by a single newline rather than a blank line.
func isToolEntry(k entryKind) bool {
	return k == entryTool || k == entryToolFailed
}
