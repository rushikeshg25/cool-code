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
	case "/effort":
		if arg == "" {
			effort := m.proc.GetStatus().Effort
			if effort == "" {
				effort = "provider default"
			}
			m.appendSystem("Reasoning effort: " + effort + " (use /effort minimal|low|medium|high|xhigh)")
			break
		}
		arg = strings.ToLower(arg)
		if !config.ValidReasoningEffort(arg) || arg == "" {
			m.appendSystem("Invalid effort. Use minimal, low, medium, high, or xhigh.")
			break
		}
		cfg := config.Load(m.rootDir)
		cfg.LLM.ReasoningEffort = arg
		if err := m.proc.ConfigureLLM(cfg.LLM); err != nil {
			m.appendSystem("Could not set effort: " + err.Error())
			break
		}
		if _, err := config.Set(m.rootDir, "llm.reasoningEffort", arg); err != nil {
			m.appendSystem("Effort changed for this session, but could not save it: " + err.Error())
			break
		}
		m.appendSystem("Reasoning effort set to " + strings.ToUpper(arg) + ".")
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
		m.appendSystem(fmt.Sprintf("Context: %d messages, %s%.1fk of %.0fk tokens (%s), %d pinned",
			status.MessageCount, approx, float64(status.TotalTokens)/1000,
			float64(status.MaxTokens)/1000, formatContext(status.TotalTokens, status.MaxTokens),
			len(m.proc.PinnedFiles())))
	case "/model":
		if arg == "" {
			status := m.proc.GetStatus()
			rate := "no published rate; cost is not tracked"
			if price, ok := llm.PriceFor(status.Model); ok {
				rate = fmt.Sprintf("$%g in / $%g out per million tokens", price.Input, price.Output)
			}
			m.appendSystem("Model: " + status.Model + "\n" + rate +
				"\nSwitch with /model <id>, for example /model claude-sonnet-4-5")
			break
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
	b.WriteString("\n  Enter submits · Alt+Enter or Ctrl+J inserts a newline (Shift+Enter is not detectable in terminals)")
	b.WriteString("\n  Esc or Ctrl+C cancels a running turn · Ctrl+C when idle quits")
	b.WriteString("\n  Up/Down recalls input history · PgUp/PgDn or mouse wheel scrolls the transcript")
	return b.String()
}

// switchModel rebuilds the provider for a new model id and persists it as the
// global default, the same place /connect writes to. Model choice is a trusted
// setting, so it never lands in the repository's .coolcode.json.
func (m *model) switchModel(id string) error {
	llmCfg := m.proc.LLMConfig()
	llmCfg.Model = id
	// The provider is inferred from the id, so clear any stale pin.
	llmCfg.Provider = ""
	if err := m.proc.ConfigureLLM(llmCfg); err != nil {
		return err
	}
	return config.SetGlobalLLM(config.LLM{Model: id})
}
