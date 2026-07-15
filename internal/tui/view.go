package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	"github.com/charmbracelet/lipgloss"

	"github.com/rushikeshg25/cool-code/internal/types"
)

func newViewport(width, height int) viewport.Model {
	vp := viewport.New(width, height)
	return vp
}

// footerHeight estimates the lines used by the footer so the viewport can be
// sized. It is recomputed from the current footer string.
func (m *model) viewportHeight() int {
	fh := lipgloss.Height(m.footer())
	h := m.height - fh
	if h < 3 {
		h = 3
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

// footer renders the task panel, status bar, and the input / overlay region.
func (m *model) footer() string {
	var sections []string

	if m.tasks != nil && len(m.tasks.Items) > 0 {
		sections = append(sections, m.renderTasks())
	}
	sections = append(sections, m.renderStatusBar())
	sections = append(sections, m.renderInputRegion())

	return lipgloss.JoinVertical(lipgloss.Left, sections...)
}

func (m *model) renderTasks() string {
	var b strings.Builder
	b.WriteString(lipgloss.NewStyle().Foreground(fg).Bold(true).Render("Goal: "+m.tasks.Goal) + "\n")
	for _, it := range m.tasks.Items {
		glyph, style := taskGlyph(it.Status)
		b.WriteString(style.Render(glyph+" "+it.Title) + "\n")
		if it.Detail != "" {
			b.WriteString(faintStyle.Render("      "+it.Detail) + "\n")
		}
	}
	return strings.TrimRight(b.String(), "\n")
}

func taskGlyph(status types.TaskStatus) (string, lipgloss.Style) {
	switch status {
	case types.TaskDone:
		return "[x]", lipgloss.NewStyle().Foreground(success)
	case types.TaskInProgress:
		return "[~]", lipgloss.NewStyle().Foreground(accent)
	case types.TaskFailed:
		return "[!]", lipgloss.NewStyle().Foreground(danger)
	default:
		return "[ ]", lipgloss.NewStyle().Foreground(muted)
	}
}

func (m *model) renderStatusBar() string {
	status := m.proc.GetStatus()
	sep := sepStyle.Render("  ·  ")
	dir := statusValue.Render(shortPath(m.rootDir))
	mode := modeStyle(m.mode).Render(string(m.mode))
	model := statusValue.Render(status.Model)
	msgs := statusBar.Render(fmt.Sprintf("%d msgs", status.MessageCount))
	toks := statusBar.Render(fmt.Sprintf("%.1fk tok", float64(status.TotalTokens)/1000))
	return statusBar.Render(dir + sep + mode + sep + model + sep + msgs + sep + toks)
}

func (m *model) renderInputRegion() string {
	switch {
	case m.status != "":
		return m.sp.View() + " " + subtleStyle.Render(m.status)
	case m.confirmMsg != "":
		return lipgloss.JoinVertical(lipgloss.Left,
			confirmTitle.Render("⚠ Confirmation required"),
			confirmBox.Render(m.confirmMsg),
			faintStyle.Render("Proceed? [y/N]"),
		)
	case m.planMenu:
		var b strings.Builder
		b.WriteString(menuTitle.Render("Plan ready. What next?") + "\n")
		for i, label := range planOptions {
			if i == m.planIdx {
				b.WriteString(menuSel.Render(fmt.Sprintf("› %d. %s", i+1, label)) + "\n")
			} else {
				b.WriteString(menuNorm.Render(fmt.Sprintf("  %d. %s", i+1, label)) + "\n")
			}
		}
		b.WriteString(faintStyle.Render("↑/↓ + Enter, or press 1/2 — Esc to keep planning"))
		return b.String()
	default:
		out := m.ti.View()
		if len(m.suggestions) > 0 {
			out += "\n" + m.renderSuggestions()
		}
		return out
	}
}

func (m *model) renderSuggestions() string {
	var b strings.Builder
	for i, s := range m.suggestions {
		nameStyle := suggestNorm
		if i == m.suggestIdx%len(m.suggestions) {
			nameStyle = suggestSel
		}
		b.WriteString("  " + nameStyle.Render(s.name) + "  " + suggestDesc.Render(s.desc))
		if i < len(m.suggestions)-1 {
			b.WriteString("\n")
		}
	}
	return b.String()
}

func shortPath(p string) string {
	parts := strings.Split(p, "/")
	if len(parts) <= 3 {
		return p
	}
	return ".../" + strings.Join(parts[len(parts)-2:], "/")
}
