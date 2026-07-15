package tui

import "github.com/charmbracelet/lipgloss"

// asciiLogo is the COOLCODE wordmark shown at startup.
const asciiLogo = ` ██████╗ ██████╗  ██████╗ ██╗       ██████╗ ██████╗ ██████╗ ███████╗
██╔════╝██╔═══██╗██╔═══██╗██║      ██╔════╝██╔═══██╗██╔══██╗██╔════╝
██║     ██║   ██║██║   ██║██║      ██║     ██║   ██║██║  ██║█████╗
██║     ██║   ██║██║   ██║██║      ██║     ██║   ██║██║  ██║██╔══╝
╚██████╗╚██████╔╝╚██████╔╝███████╗ ╚██████╗╚██████╔╝██████╔╝███████╗
 ╚═════╝ ╚═════╝  ╚═════╝ ╚══════╝  ╚═════╝ ╚═════╝ ╚═════╝ ╚══════╝`

// Banner renders the startup banner with the version and tagline.
func Banner(version string) string {
	logo := logoStyle.Render(asciiLogo)
	tagline := subtleStyle.Render("  A fast, native CLI coding agent  ·  v" + version)
	return lipgloss.JoinVertical(lipgloss.Left, logo, tagline)
}
