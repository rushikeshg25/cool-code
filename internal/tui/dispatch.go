package tui

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/rushikeshg25/cool-code/internal/session"
	"github.com/rushikeshg25/cool-code/internal/skills"
	"github.com/rushikeshg25/cool-code/internal/types"
)

func (m *model) dispatchCommand(raw string) (tea.Model, tea.Cmd) {
	parts := strings.Fields(raw)
	name := strings.ToLower(parts[0])
	var arg string
	for _, p := range parts[1:] {
		if !strings.HasPrefix(p, "--") {
			arg = p
			break
		}
	}

	switch name {
	case "/exit", "/quit":
		return m.quit()
	case "/clear":
		m.history = nil
		m.syncViewport()
	case "/help":
		m.appendSystem(m.helpText())
	case "/connect":
		m.connectMenu = true
		m.connectIdx = 0
	case "/mode":
		if arg != "" && types.AgentMode(arg).Valid() {
			m.proc.SetMode(types.AgentMode(arg))
			m.mode = types.AgentMode(arg)
			m.appendSystem("Mode switched to " + strings.ToUpper(arg))
		} else {
			m.appendSystem("Current mode: " + strings.ToUpper(string(m.mode)) + " (use /mode plan|agent|ask)")
		}
	case "/pin":
		if arg == "" {
			m.appendSystem("Usage: /pin <file>")
			break
		}
		abs := m.resolve(arg)
		if _, err := os.Stat(abs); err == nil {
			m.proc.PinFile(abs)
			m.appendSystem("Pinned: " + arg)
		} else {
			m.appendSystem("File not found: " + arg)
		}
	case "/unpin":
		if arg != "" {
			m.proc.UnpinFile(m.resolve(arg))
			m.appendSystem("Unpinned: " + arg)
		} else {
			pinned := m.proc.PinnedFiles()
			if len(pinned) == 0 {
				m.appendSystem("No files pinned.")
			} else {
				var rels []string
				for _, p := range pinned {
					rel, _ := filepath.Rel(m.rootDir, p)
					rels = append(rels, rel)
				}
				m.appendSystem("Pinned: " + strings.Join(rels, ", "))
			}
		}
	case "/context":
		status := m.proc.GetStatus()
		approx := ""
		if status.Estimated {
			approx = "~"
		}
		m.appendSystem(fmt.Sprintf("Context: %d messages, %s%.1fk tokens, %d pinned",
			status.MessageCount, approx, float64(status.TotalTokens)/1000, len(m.proc.PinnedFiles())))
	case "/sessions":
		sessions := session.List(m.rootDir)
		if len(sessions) == 0 {
			m.appendSystem("No saved sessions for this directory.")
			break
		}
		var b strings.Builder
		b.WriteString("Saved sessions (newest first):")
		for i, s := range sessions {
			if i >= 10 {
				break
			}
			id := s.ID
			if len(id) > 8 {
				id = id[:8]
			}
			b.WriteString(fmt.Sprintf("\n  %s  %s  %d msgs", id, s.UpdatedAt, s.MessageCount))
		}
		m.appendSystem(b.String())
	case "/install-skill":
		if arg == "" {
			m.appendSystem("Usage: /install-skill <local-path|git-url> [--global]")
			break
		}
		global := false
		for _, p := range parts[1:] {
			if p == "--global" {
				global = true
			}
		}
		result := skills.Install(arg, global, m.rootDir)
		if result.Error != "" {
			m.appendSystem("Install failed: " + result.Error)
		} else if len(result.Installed) == 0 {
			m.appendSystem("No skills found in the source.")
		} else {
			m.proc.ReloadSkills()
			m.appendSystem("Installed: " + strings.Join(result.Installed, ", "))
		}
	default:
		m.appendSystem("Unknown command: " + name + ". Type /help.")
	}
	return m, nil
}

func (m *model) resolve(arg string) string {
	if filepath.IsAbs(arg) {
		return arg
	}
	return filepath.Join(m.rootDir, arg)
}

func (m *model) helpText() string {
	var b strings.Builder
	b.WriteString("Commands:")
	for _, c := range commands {
		b.WriteString("\n  " + c.name + "  —  " + c.desc)
	}
	b.WriteString("\n  Shift+Tab cycles mode (plan → agent → ask)")
	b.WriteString("\nKeys:")
	b.WriteString("\n  Enter submits · Alt+Enter or Ctrl+J inserts a newline (Shift+Enter is not detectable in terminals)")
	b.WriteString("\n  Esc or Ctrl+C cancels a running turn · Ctrl+C when idle quits")
	b.WriteString("\n  Up/Down recalls input history · PgUp/PgDn or mouse wheel scrolls the transcript")
	return b.String()
}
