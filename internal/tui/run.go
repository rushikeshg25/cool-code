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
	// Notices are shown once at the top of the transcript at startup.
	Notices []string
}

// Run starts the interactive TUI and blocks until the user exits.
func Run(opts RunOptions) error {
	m := newModel(opts.Processor, opts.RootDir, opts.Version, opts.Copy, opts.SessionID)
	for _, notice := range opts.Notices {
		m.appendSystem(notice)
	}

	// Repopulate the transcript from a resumed session so the prior
	// conversation is visible, not just restored into the model's context.
	m.repopulateTranscript()

	br := &bridge{}
	m.bridge = br

	prog := tea.NewProgram(m, tea.WithAltScreen(), tea.WithMouseCellMotion())
	br.prog = prog
	opts.Processor.SetConfirmHandlers(br.confirm, br.confirmEdit)

	_, err := prog.Run()
	return err
}
