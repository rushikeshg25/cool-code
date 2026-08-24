package tui

import "github.com/charmbracelet/lipgloss"

// Banner renders the startup banner with the version and tagline.
func Banner(version string) string {
	logo := logoStyle.Render("◆ cool-code")
	meta := subtleStyle.Render("v" + version + "  ·  native coding agent")
	return lipgloss.JoinHorizontal(lipgloss.Center, logo, "  ", meta)
}
