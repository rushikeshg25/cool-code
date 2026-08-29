package tui

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"

	"github.com/rushikeshg25/cool-code/internal/security"
	"github.com/rushikeshg25/cool-code/internal/types"
)

func newViewport(width, height int) viewport.Model {
	return viewport.New(width, height)
}

func (m *model) View() string {
	if !m.ready {
		return "\n  Starting cool-code…\n"
	}
	// The footer is built once and measured, then the layout is derived from
	// it. Rendering must not mutate the model, so the confirmation overlay's
	// scroll offset is clamped before anything is drawn.
	m.clampConfirmOffset()
	footer := m.footer()
	l := computeLayout(m.width, m.height, lipgloss.Height(footer), m.hasOverlay())
	m.layout = l

	m.vp.Height = l.transcriptHeight
	m.vp.Width = l.transcriptWidth

	body := m.vp.View()
	if l.showSidebar {
		body = m.joinSidebar(body, l)
	}

	sections := make([]string, 0, 3)
	if l.showHeader {
		sections = append(sections, m.renderHeader(l))
	}
	sections = append(sections, body)
	return strings.Join(sections, "\n") + "\n" + footer
}

// joinSidebar places the task and agent panel beside the transcript.
func (m *model) joinSidebar(body string, l layout) string {
	panel := m.renderSidebar(l)
	bodyLines := padLines(strings.Split(body, "\n"), l.transcriptHeight, l.transcriptWidth)
	panelLines := padLines(strings.Split(panel, "\n"), l.transcriptHeight, l.sidebarWidth)

	rule := sidebarRule.Render("│")
	out := make([]string, l.transcriptHeight)
	for i := range out {
		out[i] = bodyLines[i] + " " + rule + " " + panelLines[i]
	}
	return strings.Join(out, "\n")
}

// padLines trims or extends lines to exactly n entries, each padded to width so
// the two columns stay aligned.
func padLines(lines []string, n, width int) []string {
	out := make([]string, n)
	for i := range out {
		line := ""
		if i < len(lines) {
			line = ansi.Truncate(lines[i], maxInt(1, width), "…")
		}
		if pad := width - ansi.StringWidth(line); pad > 0 {
			line += strings.Repeat(" ", pad)
		}
		out[i] = line
	}
	return out
}

// renderHeader is the one row that never scrolls away.
func (m *model) renderHeader(l layout) string {
	status := m.proc.GetStatus()
	left := headerName.Render("◆ cool-code") + headerStyle.Render(" v"+m.version)

	right := []string{modeStyle(m.mode).Render(string(m.mode))}
	if l.width >= taskListMinWidth && status.Model != "" {
		right = append(right, headerStyle.Render(status.Model))
	}
	if l.width >= sidebarMinWidth && status.Effort != "" {
		right = append(right, headerStyle.Render(status.Effort+" effort"))
	}
	rightText := strings.Join(right, headerStyle.Render("  ·  "))

	gap := l.width - ansi.StringWidth(left) - ansi.StringWidth(rightText) - 2
	if gap < 1 {
		return ansi.Truncate(left+" "+rightText, maxInt(1, l.width), "…")
	}
	return left + " " + headerRule.Render(strings.Repeat("─", gap)) + " " + rightText
}

// renderSidebar lists the task items and any running subagents. The old task
// panel was a single aggregated line; the individual items existed in the model
// but were never shown.
func (m *model) renderSidebar(l layout) string {
	var lines []string
	w := l.sidebarWidth

	if m.tasks != nil && len(m.tasks.Items) > 0 {
		done := 0
		for _, item := range m.tasks.Items {
			if item.Status == types.TaskDone {
				done++
			}
		}
		lines = append(lines, sidebarTitle.Render("Tasks")+
			sidebarTodo.Render(fmt.Sprintf("  %d/%d", done, len(m.tasks.Items))))
		for _, item := range m.tasks.Items {
			glyph, style := taskGlyph(item.Status)
			lines = append(lines, ansi.Truncate(style.Render(glyph+" ")+sidebarTodo.Render(item.Title), maxInt(1, w), "…"))
		}
	}

	if len(m.subagents) > 0 {
		if len(lines) > 0 {
			lines = append(lines, "")
		}
		lines = append(lines, sidebarTitle.Render("Agents"))
		for _, sub := range m.subagents {
			lines = append(lines, ansi.Truncate(sidebarNow.Render("◆ ")+sidebarTodo.Render(sub), maxInt(1, w), "…"))
		}
	}

	if len(lines) == 0 {
		lines = append(lines, sidebarTodo.Render("No active tasks"))
	}
	return strings.Join(lines, "\n")
}

// footer is the bottom stack: activity, an optional task summary when there is
// no sidebar to hold it, session metadata, then the overlay or composer.
func (m *model) footer() string {
	l := computeLayout(m.width, m.height, 0, m.hasOverlay())
	sections := make([]string, 0, 4)
	if !m.hasOverlay() && (m.processing || m.status != "") {
		sections = append(sections, m.renderActivity())
	}
	if !l.showSidebar && m.tasks != nil && len(m.tasks.Items) > 0 {
		sections = append(sections, m.renderTasks())
	}
	sections = append(sections, m.renderStatusBar(), m.renderInputRegion())
	return lipgloss.JoinVertical(lipgloss.Left, sections...)
}

func (m *model) hasOverlay() bool {
	return m.confirmMsg != "" || m.connectFor >= 0 || m.connectMenu || m.sessionMenu || m.planMenu
}

func (m *model) renderTasks() string {
	done := 0
	current := ""
	failed := false
	for _, item := range m.tasks.Items {
		switch item.Status {
		case types.TaskDone:
			done++
		case types.TaskInProgress:
			current = item.Title
		case types.TaskFailed:
			failed = true
			if current == "" {
				current = item.Title
			}
		}
	}
	glyph := "◇"
	if failed {
		glyph = "!"
	} else if done == len(m.tasks.Items) {
		glyph = "✓"
	}
	count := taskCount.Render(fmt.Sprintf("%d/%d", done, len(m.tasks.Items)))
	label := m.tasks.Goal
	if current != "" {
		label = current
	}
	line := taskStyle.Render(glyph+" Plan ") + count + taskStyle.Render("  ·  "+label)
	return ansi.Truncate(line, maxInt(1, m.width), "…")
}

func (m *model) renderStatusBar() string {
	status := m.proc.GetStatus()
	project := filepath.Base(m.rootDir)
	if project == "." || project == string(filepath.Separator) || project == "" {
		project = m.rootDir
	}
	if n := len(m.proc.ExtraDirs()); n > 0 {
		project += fmt.Sprintf(" +%d", n)
	}

	sep := sepStyle.Render("  ·  ")
	parts := []string{statusValue.Render(project), modeStyle(m.mode).Render(string(m.mode))}
	if m.width >= 72 {
		parts = append(parts, statusValue.Render(status.Model))
		if status.Effort != "" {
			parts = append(parts, statusBar.Render(status.Effort+" effort"))
		}
	}
	if m.width >= 100 {
		parts = append(parts, statusBar.Render(fmt.Sprintf("%d msgs", status.MessageCount)))
	}
	prefix := ""
	if status.Estimated {
		prefix = "~"
	}
	parts = append(parts, statusBar.Render(fmt.Sprintf("%s%.1fk ctx", prefix, float64(status.TotalTokens)/1000)))
	return ansi.Truncate(strings.Join(parts, sep), maxInt(1, m.width), "…")
}

func (m *model) renderInputRegion() string {
	switch {
	case m.confirmMsg != "":
		return m.renderConfirmation()
	case m.connectFor >= 0:
		return lipgloss.JoinVertical(lipgloss.Left,
			menuTitle.Render(truncate("Connect "+connectOptions[m.connectFor].provider, m.width)),
			keyFieldStyle.Width(maxInt(12, m.width-2)).Render(m.keyInput.View()),
			faintStyle.Render(truncate("Enter save  ·  Esc cancel", m.width)),
		)
	case m.connectMenu:
		return m.renderConnectMenu()
	case m.sessionMenu:
		return m.renderSessionMenu()
	case m.planMenu:
		return m.renderPlanMenu()
	default:
		var sections []string
		if len(m.suggestions) > 0 {
			sections = append(sections, m.renderSuggestions())
		}
		sections = append(sections, m.renderComposer())
		return lipgloss.JoinVertical(lipgloss.Left, sections...)
	}
}

func (m *model) renderActivity() string {
	label := m.status
	if label == "" {
		label = "Working…"
	}
	// spinner.Dot's frames already end in a space, so do not add another.
	line := m.sp.View() + activityStyle.Render(label) + faintStyle.Render("  ·  Esc cancel")
	lines := []string{m.truncateToWidth(line)}

	// The sidebar lists the running agents when it is showing, so repeating
	// them here would draw each one twice.
	if computeLayout(m.width, m.height, 0, m.hasOverlay()).showSidebar {
		return lines[0]
	}
	for i, sub := range m.subagents {
		if sub == "" || len(lines) > 3 {
			continue
		}
		// Only the final entry is a last child.
		branch := "├─ "
		if i == len(m.subagents)-1 {
			branch = "└─ "
		}
		lines = append(lines, m.truncateToWidth(faintStyle.Render("  "+branch+sub)))
	}
	return strings.Join(lines, "\n")
}

func (m *model) renderComposer() string {
	box := composerStyle.Width(maxInt(12, m.width-2)).Render(m.ti.View())
	hint := "Enter send  ·  Alt+Enter newline  ·  / commands  ·  @ files"
	if m.processing {
		hint = "Enter queue follow-up  ·  Alt+Enter newline  ·  Esc cancel"
	}
	return box + "\n" + ansi.Truncate(faintStyle.Render(hint), maxInt(1, m.width), "…")
}

func (m *model) renderSuggestions() string {
	n := len(m.suggestions)
	selected := m.suggestIdx % n
	start, end := visibleWindow(n, selected, m.menuLimit())
	lines := make([]string, 0, end-start+1)
	for i := start; i < end; i++ {
		suggestion := m.suggestions[i]
		marker := "  "
		nameStyle := suggestNorm
		if i == selected {
			marker = "› "
			nameStyle = suggestSel
		}
		line := marker + nameStyle.Render(suggestion.name)
		if suggestion.desc != "" {
			line += "  " + suggestDesc.Render(suggestion.desc)
		}
		lines = append(lines, ansi.Truncate(line, maxInt(1, m.width), "…"))
	}
	action := "Enter run  ·  Tab complete"
	if m.suggestMode == suggestFile {
		action = "Tab/Enter insert path"
	}
	if n > end-start {
		action = fmt.Sprintf("%d/%d  ·  %s", selected+1, n, action)
	}
	lines = append(lines, m.truncateToWidth(faintStyle.Render("  ↑/↓ navigate  ·  "+action)))
	return strings.Join(lines, "\n")
}

// confirmLines is the wrapped, sanitized body of the confirmation overlay.
// The text quotes a model-supplied command, so escapes are stripped before it
// can redraw over what the user is being asked to approve.
func (m *model) confirmLines() []string {
	wrapped := ansi.Wordwrap(security.SanitizeTerminal(m.confirmMsg), maxInt(20, m.width-4), " /")
	return strings.Split(wrapped, "\n")
}

// confirmLimit is how many body lines the overlay may show. It has to account
// for its own chrome (title, two border rows, hint) plus the header, status bar
// and the separator row, or a short terminal overflows.
func (m *model) confirmLimit() int {
	const chrome = 8
	return maxInt(1, minInt(10, minInt(m.height/2, m.height-chrome)))
}

// clampConfirmOffset keeps the scroll offset in range. This used to happen
// inside renderConfirmation, which meant drawing a frame mutated the model.
func (m *model) clampConfirmOffset() {
	if m.confirmMsg == "" {
		return
	}
	maxOffset := maxInt(0, len(m.confirmLines())-m.confirmLimit())
	if m.confirmOff > maxOffset {
		m.confirmOff = maxOffset
	}
	if m.confirmOff < 0 {
		m.confirmOff = 0
	}
}

func (m *model) renderConfirmation() string {
	lines := m.confirmLines()
	limit := m.confirmLimit()
	offset := minInt(m.confirmOff, maxInt(0, len(lines)-limit))
	end := minInt(len(lines), offset+limit)
	preview := colorizeDiff(strings.Join(lines[offset:end], "\n"))
	hint := "y approve  ·  n/Enter reject"
	if len(lines) > limit {
		hint = fmt.Sprintf("↑/↓ scroll %d-%d/%d  ·  %s", offset+1, end, len(lines), hint)
	}
	return lipgloss.JoinVertical(lipgloss.Left,
		confirmTitle.Render("⚠ Confirmation required"),
		confirmBox.Width(maxInt(12, m.width-2)).Render(preview),
		faintStyle.Render(truncate(hint, m.width)),
	)
}

func (m *model) renderConnectMenu() string {
	lines := []string{m.truncateToWidth(menuTitle.Render("Connect a provider"))}
	for i, option := range connectOptions {
		line := fmt.Sprintf("  %d. %s", i+1, option.label)
		if i == m.connectIdx {
			line = menuSel.Render(fmt.Sprintf("› %d. %s", i+1, option.label))
		} else {
			line = menuNorm.Render(line)
		}
		lines = append(lines, ansi.Truncate(line, maxInt(1, m.width), "…"))
	}
	lines = append(lines, m.truncateToWidth(faintStyle.Render("↑/↓ navigate  ·  Enter select  ·  Esc cancel")))
	return strings.Join(lines, "\n")
}

func (m *model) renderSessionMenu() string {
	n := len(m.sessionList)
	start, end := visibleWindow(n, m.sessionIdx, m.menuLimit())
	lines := []string{m.truncateToWidth(menuTitle.Render("Resume a session"))}
	for i := start; i < end; i++ {
		sess := m.sessionList[i]
		short := sess.ID
		if len(short) > 8 {
			short = short[:8]
		}
		label := fmt.Sprintf("%d. %s  %s  %d msgs", i+1, short, sess.UpdatedAt, sess.MessageCount)
		if i == m.sessionIdx {
			label = menuSel.Render("› " + label)
		} else {
			label = menuNorm.Render("  " + label)
		}
		lines = append(lines, ansi.Truncate(label, maxInt(1, m.width), "…"))
	}
	lines = append(lines, m.truncateToWidth(faintStyle.Render("↑/↓ navigate  ·  Enter resume  ·  Esc cancel")))
	return strings.Join(lines, "\n")
}

func (m *model) renderPlanMenu() string {
	lines := []string{m.truncateToWidth(menuTitle.Render("Plan ready. What next?"))}
	for i, label := range planOptions {
		line := fmt.Sprintf("  %d. %s", i+1, label)
		if i == m.planIdx {
			line = menuSel.Render(fmt.Sprintf("› %d. %s", i+1, label))
		} else {
			line = menuNorm.Render(line)
		}
		lines = append(lines, ansi.Truncate(line, maxInt(1, m.width), "…"))
	}
	lines = append(lines, m.truncateToWidth(faintStyle.Render("↑/↓ navigate  ·  Enter select  ·  Esc keep planning")))
	return strings.Join(lines, "\n")
}

func (m *model) menuLimit() int {
	if m.height <= 20 {
		return 4
	}
	return 6
}

func visibleWindow(total, selected, limit int) (start, end int) {
	if total <= limit {
		return 0, total
	}
	start = selected - limit/2
	if start < 0 {
		start = 0
	}
	if start+limit > total {
		start = total - limit
	}
	return start, start + limit
}

// colorizeDiff styles edit previews without relying on color alone: the +/-
// prefixes remain visible in monochrome terminals.
func colorizeDiff(s string) string {
	lines := strings.Split(s, "\n")
	for i, line := range lines {
		switch {
		case strings.HasPrefix(line, "- "):
			lines[i] = lipgloss.NewStyle().Foreground(danger).Render(line)
		case strings.HasPrefix(line, "+ "):
			lines[i] = lipgloss.NewStyle().Foreground(success).Render(line)
		}
	}
	return strings.Join(lines, "\n")
}

func shortPath(path string) string {
	parts := strings.Split(filepath.Clean(path), string(filepath.Separator))
	if len(parts) <= 3 {
		return path
	}
	return "…" + string(filepath.Separator) + filepath.Join(parts[len(parts)-2:]...)
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// truncate fits a line to the terminal width. Several hint and title lines
// skipped this and wrapped on narrow terminals, which broke the height budget.
func (m *model) truncateToWidth(s string) string { return truncate(s, m.width) }

func truncate(s string, width int) string {
	return ansi.Truncate(s, maxInt(1, width), "…")
}
