package tui

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/rushikeshg25/cool-code/internal/types"
)

// Palette - a restrained, opencode-inspired theme with a single violet accent,
// muted secondary text, and adaptive light/dark values.
var (
	accent    = lipgloss.AdaptiveColor{Light: "#6D5AE6", Dark: "#A896FF"}
	accentDim = lipgloss.AdaptiveColor{Light: "#8B7BE8", Dark: "#7B6FD0"}
	fg        = lipgloss.AdaptiveColor{Light: "#1F2430", Dark: "#E4E4EF"}
	muted     = lipgloss.AdaptiveColor{Light: "#6B7280", Dark: "#8A8FA3"}
	faint     = lipgloss.AdaptiveColor{Light: "#9AA0AE", Dark: "#5B6070"}
	success   = lipgloss.AdaptiveColor{Light: "#0E9F6E", Dark: "#4ADE9A"}
	warn      = lipgloss.AdaptiveColor{Light: "#B7791F", Dark: "#F0C674"}
	danger    = lipgloss.AdaptiveColor{Light: "#D64545", Dark: "#F08A8A"}
	blue      = lipgloss.AdaptiveColor{Light: "#2B6CB0", Dark: "#7FB4F0"}
)

var (
	logoStyle    = lipgloss.NewStyle().Foreground(accent).Bold(true)
	subtleStyle  = lipgloss.NewStyle().Foreground(muted)
	faintStyle   = lipgloss.NewStyle().Foreground(faint)
	userPrefix   = lipgloss.NewStyle().Foreground(accent).Bold(true)
	userText     = lipgloss.NewStyle().Foreground(fg)
	toolStyle    = lipgloss.NewStyle().Foreground(muted)
	toolGlyph    = lipgloss.NewStyle().Foreground(accentDim)
	systemStyle  = lipgloss.NewStyle().Foreground(faint).Italic(true)
	promptGlyph  = lipgloss.NewStyle().Foreground(accent).Bold(true)
	spinnerStyle = lipgloss.NewStyle().Foreground(accent)

	statusBar   = lipgloss.NewStyle().Foreground(muted)
	statusValue = lipgloss.NewStyle().Foreground(fg)
	sepStyle    = lipgloss.NewStyle().Foreground(faint)

	suggestSel  = lipgloss.NewStyle().Foreground(accent).Bold(true)
	suggestNorm = lipgloss.NewStyle().Foreground(muted)
	suggestDesc = lipgloss.NewStyle().Foreground(faint)

	confirmTitle = lipgloss.NewStyle().Foreground(warn).Bold(true)
	// The most consequential moment in the UI deserves a frame around it.
	confirmBox = lipgloss.NewStyle().
			Foreground(fg).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(warn).
			Padding(0, 1)
	// menuTitle was warn, the same amber as the confirmation warning, so
	// "Resume a session" read as an alert. Titles are neutral now.
	menuTitle = lipgloss.NewStyle().Foreground(fg).Bold(true)
	menuSel   = lipgloss.NewStyle().Foreground(accent).Bold(true)
	menuNorm  = lipgloss.NewStyle().Foreground(muted)

	composerStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(accentDim).
			Padding(0, 1)
	activityStyle = lipgloss.NewStyle().Foreground(muted)
	taskStyle     = lipgloss.NewStyle().Foreground(muted)
	taskCount     = lipgloss.NewStyle().Foreground(accent).Bold(true)
	planTitle     = lipgloss.NewStyle().Foreground(warn).Bold(true)
	planCard      = lipgloss.NewStyle().BorderLeft(true).BorderStyle(lipgloss.NormalBorder()).BorderForeground(accentDim).PaddingLeft(1)

	// Errors used to render as entrySystem, faint and italic, which made a
	// failure the least prominent thing on screen. danger was reserved for
	// diff lines and used nowhere else.
	errorStyle  = lipgloss.NewStyle().Foreground(danger)
	errorGlyph  = lipgloss.NewStyle().Foreground(danger).Bold(true)
	toolFailure = lipgloss.NewStyle().Foreground(danger)

	// Hoisted out of colorizeDiff, which allocated a style per line per frame.
	diffAdd = lipgloss.NewStyle().Foreground(success)
	diffDel = lipgloss.NewStyle().Foreground(danger)

	// Persistent header. The banner used to be transcript entry zero, so the
	// version, project and model scrolled away and never came back.
	headerStyle  = lipgloss.NewStyle().Foreground(muted)
	headerName   = lipgloss.NewStyle().Foreground(accent).Bold(true)
	headerRule   = lipgloss.NewStyle().Foreground(faint)
	sidebarRule  = lipgloss.NewStyle().Foreground(faint)
	sidebarTitle = lipgloss.NewStyle().Foreground(muted).Bold(true)
	sidebarDone  = lipgloss.NewStyle().Foreground(success)
	sidebarNow   = lipgloss.NewStyle().Foreground(accent)
	sidebarTodo  = lipgloss.NewStyle().Foreground(faint)
	sidebarFail  = lipgloss.NewStyle().Foreground(danger)

	// The /connect key field rendered bare, with a default-coloured prompt,
	// in the slot the bordered composer normally occupies.
	keyFieldStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(warn).
			Padding(0, 1)
)

// taskGlyph returns the marker and style for one task item.
func taskGlyph(status types.TaskStatus) (string, lipgloss.Style) {
	switch status {
	case types.TaskDone:
		return "✓", sidebarDone
	case types.TaskInProgress:
		return "◆", sidebarNow
	case types.TaskFailed:
		return "✗", sidebarFail
	default:
		return "○", sidebarTodo
	}
}

func modeStyle(mode types.AgentMode) lipgloss.Style {
	switch mode {
	case types.ModePlan:
		return lipgloss.NewStyle().Foreground(warn).Bold(true)
	case types.ModeAsk:
		return lipgloss.NewStyle().Foreground(blue).Bold(true)
	default:
		return lipgloss.NewStyle().Foreground(success).Bold(true)
	}
}
