package tui

import (
	tea "github.com/charmbracelet/bubbletea"

	"github.com/rushikeshg25/cool-code/internal/agent"
)

// RunOptions configure an interactive session.
type RunOptions struct {
	Processor *agent.Processor
	RootDir   string
	Version   string
	Copy      bool
	SessionID string
	// Banner is printed once above the UI before the program starts.
	Banner string
}

// Run starts the interactive TUI and blocks until the user exits.
func Run(opts RunOptions) error {
	m := newModel(opts.Processor, opts.RootDir, opts.Version, opts.Copy, opts.SessionID)
	if opts.Banner != "" {
		m.appendRaw(opts.Banner)
		m.appendSystem("Type your request, or /help for commands. Shift+Tab cycles mode.")
	}

	br := &bridge{}
	m.bridge = br

	prog := tea.NewProgram(m, tea.WithAltScreen())
	br.prog = prog
	opts.Processor.SetConfirmHandlers(br.confirm, br.confirmEdit)

	_, err := prog.Run()
	return err
}
