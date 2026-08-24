package tui

import (
	"path/filepath"
	"strings"

	"github.com/rushikeshg25/cool-code/internal/tools"
	"github.com/rushikeshg25/cool-code/internal/types"
)

// slashCommand is a single in-session command. Reused for @-file suggestions,
// where name holds the project-relative path and desc is empty.
type slashCommand struct {
	name string
	desc string
}

// suggestKind is what the dropdown is currently completing.
type suggestKind int

const (
	suggestNone suggestKind = iota
	suggestCommand
	suggestFile
)

var commands = []slashCommand{
	{"/help", "Show available commands"},
	{"/connect", "Connect a model provider (API key)"},
	{"/mode", "Show or switch mode (plan | agent | ask)"},
	{"/effort", "Show or set reasoning effort (minimal | low | medium | high | xhigh)"},
	{"/add-dir", "Grant access to an additional directory (e.g. /add-dir ../other)"},
	{"/pin", "Pin a file into context (e.g. /pin src/main.go)"},
	{"/unpin", "Unpin a file (or list pinned files)"},
	{"/context", "Preview context, pinned files, and token usage"},
	{"/sessions", "List saved sessions for this directory"},
	{"/install-skill", "Install a skill from a local path or git URL"},
	{"/clear", "Clear the screen"},
	{"/exit", "Exit the session"},
	{"/quit", "Exit the session"},
}

var modeCycle = []types.AgentMode{types.ModePlan, types.ModeAgent, types.ModeAsk}

// nextMode returns the next mode in the plan → agent → ask cycle.
func nextMode(mode types.AgentMode) types.AgentMode {
	for i, m := range modeCycle {
		if m == mode {
			return modeCycle[(i+1)%len(modeCycle)]
		}
	}
	return types.ModeAgent
}

// matchCommands returns commands whose name prefixes the first token, but only
// while the user is typing a command (input starts with "/").
func matchCommands(input string) []slashCommand {
	if !strings.HasPrefix(input, "/") {
		return nil
	}
	token := strings.Fields(input)
	if len(token) == 0 {
		return commands
	}
	first := token[0]
	if len(first) <= 1 {
		return commands
	}
	var out []slashCommand
	for _, c := range commands {
		if strings.HasPrefix(c.name, first) {
			out = append(out, c)
		}
	}
	return out
}

const maxFileSuggestions = 8

// atToken extracts a trailing "@<token>" mention from the input: the "@" must
// start the input or follow whitespace, and no whitespace may follow it. It
// returns the token after "@" and the byte index of the "@".
func atToken(s string) (token string, at int, ok bool) {
	i := strings.LastIndex(s, "@")
	if i < 0 {
		return "", 0, false
	}
	if i > 0 && s[i-1] != ' ' && s[i-1] != '\t' {
		return "", 0, false
	}
	rest := s[i+1:]
	if strings.ContainsAny(rest, " \t") {
		return "", 0, false
	}
	return rest, i, true
}

// matchFiles returns up to maxFileSuggestions project files matching token,
// ranking basename matches ahead of full-path matches.
func matchFiles(files []string, token string) []slashCommand {
	token = strings.ToLower(token)
	var byBase, byPath []slashCommand
	for _, f := range files {
		lower := strings.ToLower(f)
		base := strings.ToLower(filepath.Base(f))
		switch {
		case token == "" || strings.Contains(base, token):
			byBase = append(byBase, slashCommand{name: f})
		case strings.Contains(lower, token):
			byPath = append(byPath, slashCommand{name: f})
		}
	}
	out := append(byBase, byPath...)
	if len(out) > maxFileSuggestions {
		out = out[:maxFileSuggestions]
	}
	return out
}

// projectFiles lazily loads and caches the file list for @-mention completion:
// project-relative paths from the primary root plus absolute paths from any
// /add-dir directories (absolute so the reference is unambiguous).
func (m *model) projectFiles() []string {
	if m.fileCache == nil {
		m.fileCache = tools.ProjectFiles(m.rootDir)
		for _, dir := range m.proc.ExtraDirs() {
			m.fileCache = append(m.fileCache, tools.ProjectFilesAbs(dir)...)
		}
	}
	return m.fileCache
}
