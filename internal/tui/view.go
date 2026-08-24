package tui

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"

	"github.com/rushikeshg25/cool-code/internal/types"
)

func newViewport(width, height int) viewport.Model {
	return viewport.New(width, height)
}

// viewportHeight reserves the exact number of rows used by the footer plus
// the separator newline emitted by View. Footer content is deliberately
// bounded so even small terminals retain transcript space.
func (m *model) viewportHeight() int {
	h := m.height - lipgloss.Height(m.footer()) - 1
	if h < 1 {
		h = 1
	}
	return h
}

func (m *model) View() string {
	if !m.ready {
		return "\n  Starting cool-code…\n"
	}
	m.vp.Height = m.viewportHeight()
	return m.vp.View() + "\n" + m.footer()
}

// footer is a compact workbench: optional task progress, session metadata,
// then either a bounded overlay or the activity/palette/composer stack.
func (m *model) footer() string {
	sections := make([]string, 0, 4)
	if !m.hasOverlay() && (m.processing || m.status != "") {
		sections = append(sections, m.renderActivity())
	}
	if m.tasks != nil && len(m.tasks.Items) > 0 {
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
			menuTitle.Render("Connect "+connectOptions[m.connectFor].provider),
			m.keyInput.View(),
			faintStyle.Render("Enter save  ·  Esc cancel"),
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
	line := m.sp.View() + " " + activityStyle.Render(label) + faintStyle.Render("  ·  Esc cancel")
	lines := []string{ansi.Truncate(line, maxInt(1, m.width), "…")}
	for _, sub := range m.subagents {
		if sub == "" || len(lines) > 3 {
			continue
		}
		lines = append(lines, ansi.Truncate(faintStyle.Render("  └─ "+sub), maxInt(1, m.width), "…"))
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
	lines = append(lines, faintStyle.Render("  ↑/↓ navigate  ·  "+action))
	return strings.Join(lines, "\n")
}

func (m *model) renderConfirmation() string {
	wrapped := ansi.Wordwrap(m.confirmMsg, maxInt(20, m.width-2), " /")
	lines := strings.Split(wrapped, "\n")
	limit := maxInt(3, minInt(10, m.height/2))
	maxOffset := maxInt(0, len(lines)-limit)
	if m.confirmOff > maxOffset {
		m.confirmOff = maxOffset
	}
	end := minInt(len(lines), m.confirmOff+limit)
	preview := colorizeDiff(strings.Join(lines[m.confirmOff:end], "\n"))
	hint := "y approve  ·  n/Enter reject"
	if len(lines) > limit {
		hint = fmt.Sprintf("↑/↓ scroll %d-%d/%d  ·  %s", m.confirmOff+1, end, len(lines), hint)
	}
	return lipgloss.JoinVertical(lipgloss.Left,
		confirmTitle.Render("⚠ Confirmation required"),
		confirmBox.Render(preview),
		faintStyle.Render(hint),
	)
}

func (m *model) renderConnectMenu() string {
	lines := []string{menuTitle.Render("Connect a provider")}
	for i, option := range connectOptions {
		line := fmt.Sprintf("  %d. %s", i+1, option.label)
		if i == m.connectIdx {
			line = menuSel.Render(fmt.Sprintf("› %d. %s", i+1, option.label))
		} else {
			line = menuNorm.Render(line)
		}
		lines = append(lines, ansi.Truncate(line, maxInt(1, m.width), "…"))
	}
	lines = append(lines, faintStyle.Render("↑/↓ navigate  ·  Enter select  ·  Esc cancel"))
	return strings.Join(lines, "\n")
}

func (m *model) renderSessionMenu() string {
	n := len(m.sessionList)
	start, end := visibleWindow(n, m.sessionIdx, m.menuLimit())
	lines := []string{menuTitle.Render("Resume a session")}
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
	lines = append(lines, faintStyle.Render("↑/↓ navigate  ·  Enter resume  ·  Esc cancel"))
	return strings.Join(lines, "\n")
}

func (m *model) renderPlanMenu() string {
	lines := []string{menuTitle.Render("Plan ready. What next?")}
	for i, label := range planOptions {
		line := fmt.Sprintf("  %d. %s", i+1, label)
		if i == m.planIdx {
			line = menuSel.Render(fmt.Sprintf("› %d. %s", i+1, label))
		} else {
			line = menuNorm.Render(line)
		}
		lines = append(lines, ansi.Truncate(line, maxInt(1, m.width), "…"))
	}
	lines = append(lines, faintStyle.Render("↑/↓ navigate  ·  Enter select  ·  Esc keep planning"))
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
