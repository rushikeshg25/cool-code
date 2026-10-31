package tui

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/rushikeshg25/cool-code/internal/config"
	"github.com/rushikeshg25/cool-code/internal/llm"
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
		m.invalidatePrefix()
		m.syncViewport()
	case "/help":
		m.appendSystem(m.helpText())
	case "/yolo":
		enabled := !m.proc.YOLO()
		if len(parts) > 2 || (arg != "" && arg != "on" && arg != "off") {
			m.appendError("Usage: /yolo [on|off]")
			break
		}
		if arg != "" {
			enabled = arg == "on"
		}
		m.proc.SetYOLO(enabled)
		if enabled {
			m.appendSystem("YOLO enabled: action and edit approvals are skipped for this run.")
		} else {
			m.appendSystem("YOLO disabled: original approval settings restored.")
		}
	case "/connect":
		m.connectMenu = true
		m.connectIdx = 0
	case "/plan", "/ask", "/impl":
		mode := types.AgentMode(strings.TrimPrefix(name, "/"))
		if name == "/impl" {
			mode = types.ModeAgent
		}
		m.proc.SetMode(mode)
		m.mode = mode
		m.appendSystem("Mode switched to " + strings.ToUpper(string(mode)))
	case "/effort":
		if arg == "" {
			return m.openEffortPicker()
		}
		if err := m.setEffort(strings.ToLower(arg)); err != nil {
			m.appendError(err.Error())
		}
	case "/add-dir":
		if arg == "" {
			dirs := m.proc.ExtraDirs()
			if len(dirs) == 0 {
				m.appendSystem("No additional directories. Usage: /add-dir <path>")
			} else {
				m.appendSystem("Additional directories:\n  " + strings.Join(dirs, "\n  "))
			}
			break
		}
		resolved, err := m.proc.AddDir(arg)
		if err != nil {
			m.appendSystem("add-dir failed: " + err.Error())
			break
		}
		m.fileCache = nil // extra-dir files become @-completable
		m.appendSystem("Added directory: " + resolved + " (read/write enabled; search tools still cover only the primary root)")
	case "/pin":
		if arg == "" {
			m.appendSystem("Usage: /pin <file>")
			break
		}
		abs := m.resolve(arg)
		if _, err := os.Stat(abs); err == nil {
			if err := m.proc.PinFile(abs); err != nil {
				m.appendSystem("Pin blocked: " + err.Error())
			} else {
				m.appendSystem("Pinned: " + arg)
			}
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
		usage := fmt.Sprintf("%s%.1fk tokens", approx, float64(status.TotalTokens)/1000)
		if status.MaxTokens > 0 {
			usage = fmt.Sprintf("%s%.1fk of %.0fk tokens (%s)", approx,
				float64(status.TotalTokens)/1000, float64(status.MaxTokens)/1000,
				formatContext(status.TotalTokens, status.MaxTokens))
		}
		line := fmt.Sprintf("Context: %d messages, %s, %d pinned",
			status.MessageCount, usage, len(m.proc.PinnedFiles()))
		if status.MaxTokens == 0 {
			// Without a window there is nothing to take a percentage of, and
			// guessing one would be worse than saying so.
			line += "\nNo context window known for " + status.Model +
				". Declare it with `cool-code config set llm.contextWindow <tokens>`."
		}
		m.appendSystem(line)
	case "/model":
		if arg == "" {
			return m.openModelPicker()
		}
		if err := m.switchModel(arg); err != nil {
			m.appendError("Could not switch model: " + err.Error())
			break
		}
		m.appendSystem("Model set to " + arg + ".")
	case "/cost":
		status := m.proc.GetStatus()
		if !status.CostKnown {
			m.appendSystem("No published rate for " + status.Model +
				", so spend is not tracked. Token usage is still shown in the status bar.")
			break
		}
		m.appendSystem(fmt.Sprintf("Session spend: %s across %d messages (model %s).",
			formatCost(status.SessionCost), status.MessageCount, status.Model))
	case "/compact":
		// Compaction talks to the provider, so it runs off the UI goroutine.
		m.appendSystem("Compacting the conversation…")
		return m, func() tea.Msg {
			return compactedMsg{summary: m.proc.Compact(context.Background())}
		}
	case "/btw":
		return m.openAside(strings.TrimSpace(strings.TrimPrefix(raw, parts[0])))
	case "/fork":
		return m.forkSession()
	case "/sessions":
		sessions := session.List(m.rootDir)
		if len(sessions) == 0 {
			m.appendSystem("No saved sessions for this directory.")
			break
		}
		if len(sessions) > 10 {
			sessions = sessions[:10]
		}
		m.sessionList = sessions
		m.sessionIdx = 0
		m.sessionMenu = true
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
		b.WriteString("\n  " + c.name + "  -  " + c.desc)
	}
	b.WriteString("\n  Shift+Tab cycles mode (plan → agent → ask)")
	b.WriteString("\nKeys:")
	b.WriteString("\n  Enter sends · Ctrl/Shift+Enter inserts a newline in enhanced terminals (Ghostty/Kitty)\n  Alt+Enter or Ctrl+J inserts a newline in legacy terminals")
	b.WriteString("\n  Ctrl+Space / Ctrl+Backspace / Ctrl+W deletes the previous word · Ctrl+Delete / Alt+D deletes the next word")
	b.WriteString("\n  Ctrl+Left/Right or Alt+B/F moves by word · Home/End or Ctrl+A/E moves to line ends")
	b.WriteString("\n  Ctrl+U/K deletes to line start/end · Ctrl+Home/End moves to prompt start/end")
	b.WriteString("\n  Esc or Ctrl+C cancels a running turn · Ctrl+C when idle quits")
	b.WriteString("\n  Up/Down recalls input history · PgUp/PgDn or mouse wheel scrolls the transcript\n  Alt+End jumps to the latest output and resumes auto-scroll")
	return b.String()
}

// switchModel rebuilds the provider for a new model id and persists it as the
// global default, the same place /connect writes to. Model choice is a trusted
// setting, so it never lands in the repository's .coolcode.json.
func (m *model) switchModel(id string) error {
	llmCfg := m.proc.LLMConfig()
	llmCfg.Model = id
	// A proxy model ID must stay on the explicitly selected proxy.
	if llm.ResolveProvider(llmCfg) != "cliproxy" {
		llmCfg.Provider = ""
	}
	if err := m.proc.ConfigureLLM(llmCfg); err != nil {
		return err
	}
	return config.SetGlobalLLM(config.LLM{Model: id, Provider: llm.ResolveProvider(llmCfg)})
}
