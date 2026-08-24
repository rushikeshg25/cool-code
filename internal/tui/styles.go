package tui

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/rushikeshg25/cool-code/internal/types"
)

// Palette — a restrained, opencode-inspired theme with a single violet accent,
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
	confirmBox   = lipgloss.NewStyle().Foreground(fg)
	menuTitle    = lipgloss.NewStyle().Foreground(warn).Bold(true)
	menuSel      = lipgloss.NewStyle().Foreground(accent).Bold(true)
	menuNorm     = lipgloss.NewStyle().Foreground(muted)

	composerStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(accentDim).
			Padding(0, 1)
	activityStyle = lipgloss.NewStyle().Foreground(muted)
	taskStyle     = lipgloss.NewStyle().Foreground(muted)
	taskCount     = lipgloss.NewStyle().Foreground(accent).Bold(true)
	planTitle     = lipgloss.NewStyle().Foreground(warn).Bold(true)
	planCard      = lipgloss.NewStyle().BorderLeft(true).BorderStyle(lipgloss.NormalBorder()).BorderForeground(accentDim).PaddingLeft(1)
)

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
