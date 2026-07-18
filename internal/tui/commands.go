package tui

import (
	"strings"

	"github.com/rushikeshg25/cool-code/internal/types"
)

// slashCommand is a single in-session command.
type slashCommand struct {
	name string
	desc string
}

var commands = []slashCommand{
	{"/help", "Show available commands"},
	{"/connect", "Connect a model provider (API key)"},
	{"/mode", "Show or switch mode (plan | agent | ask)"},
	{"/pin", "Pin a file into context (e.g. /pin src/main.go)"},
	{"/unpin", "Unpin a file (or list pinned files)"},
	{"/context", "Preview context, pinned files, and token usage"},
	{"/sessions", "List saved sessions for this directory"},
	{"/install-skill", "Install a skill from a local path or git URL"},
	{"/clear", "Clear the screen"},
	{"/exit", "Exit the session"},
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
