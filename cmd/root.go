// Package cmd wires the cool-code command-line interface.
package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/joho/godotenv"
	"github.com/spf13/cobra"

	"github.com/rushikeshg25/cool-code/internal/agent"
	"github.com/rushikeshg25/cool-code/internal/config"
	"github.com/rushikeshg25/cool-code/internal/llm"
	"github.com/rushikeshg25/cool-code/internal/session"
	"github.com/rushikeshg25/cool-code/internal/tui"
	"github.com/rushikeshg25/cool-code/internal/types"
)

// Version is the CLI version, overridable at build time via -ldflags.
var Version = "2.0.0"

type rootFlags struct {
	yes            bool
	noInit         bool
	allowDangerous bool
	copy           bool
	continueSess   bool
	resumeID       string
}

// Execute runs the root command.
func Execute() {
	var flags rootFlags

	root := &cobra.Command{
		Use:           "cool-code",
		Short:         "A fast, native CLI coding agent",
		Long:          "cool-code is an interactive CLI coding agent. Describe what you want; it reads, plans, and edits your codebase using live tools.",
		Version:       Version,
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runInteractive(flags)
		},
	}

	root.Flags().BoolVarP(&flags.yes, "yes", "y", false, "Start without the initialization prompt")
	root.Flags().BoolVar(&flags.noInit, "no-init", false, "Exit without initializing in the current directory")
	root.Flags().BoolVar(&flags.allowDangerous, "allow-dangerous", false, "Allow dangerous actions without prompting")
	root.Flags().BoolVar(&flags.copy, "copy", false, "Copy final responses to the clipboard")
	root.Flags().BoolVar(&flags.continueSess, "continue", false, "Resume the most recent session for this directory")
	root.Flags().StringVar(&flags.resumeID, "resume", "", "Resume a specific session by id")

	root.SetVersionTemplate("cool-code v{{.Version}}\n")
	root.AddCommand(configCmd(), scanCmd(), skillCmd(), taskCmd())

	if err := root.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}
}

func runInteractive(flags rootFlags) error {
	loadEnv()

	rootDir, err := os.Getwd()
	if err != nil {
		return err
	}
	if flags.noInit {
		fmt.Println("Exiting without initialization.")
		return nil
	}

	cfg := config.Load(rootDir)
	if flags.allowDangerous {
		t := true
		cfg.Features.AllowDangerous = &t
	}

	proc, err := agent.New(rootDir, cfg, agent.Options{
		Mode:            types.ModeAgent,
		AllowDangerous:  cfg.AllowDangerous(),
		AllowMissingKey: true,
	})
	if err != nil {
		return handleProviderError(err)
	}

	banner := tui.Banner(Version)
	if !proc.Connected() {
		banner += "\n  No API key configured — run /connect to link a provider."
	}

	sessionID := session.NewID()
	var restoreFrom *session.Data
	switch {
	case flags.resumeID != "":
		restoreFrom = session.Load(flags.resumeID)
	case flags.continueSess:
		restoreFrom = session.Latest(rootDir)
	}
	if restoreFrom != nil {
		var messages []llm.Message
		_ = json.Unmarshal(restoreFrom.Messages, &messages)
		proc.Restore(messages, restoreFrom.Summary, restoreFrom.PinnedFiles, types.AgentMode(restoreFrom.Mode))
		sessionID = restoreFrom.ID
		short := restoreFrom.ID
		if len(short) > 8 {
			short = short[:8]
		}
		banner += fmt.Sprintf("\n  Resumed session %s (%d messages).", short, restoreFrom.MessageCount)
	}

	return tui.Run(tui.RunOptions{
		Processor: proc,
		RootDir:   rootDir,
		Version:   Version,
		Copy:      flags.copy,
		SessionID: sessionID,
		Banner:    banner,
	})
}

// loadEnv loads a local .env without emitting output.
func loadEnv() {
	_ = godotenv.Load()
}

// handleProviderError prints a friendly setup hint for a missing API key.
func handleProviderError(err error) error {
	var missing *llm.MissingKeyError
	if ok := asMissingKey(err, &missing); ok {
		fmt.Fprintf(os.Stderr, "\n  Missing API key for %s.\n\n", missing.Provider)
		fmt.Fprintf(os.Stderr, "  Set it with:\n    export %s=your_api_key_here\n\n", missing.EnvKey)
		fmt.Fprintf(os.Stderr, "  Get a key at: %s\n\n", missing.KeyURL)
		os.Exit(1)
	}
	return err
}

func asMissingKey(err error, target **llm.MissingKeyError) bool {
	if mk, ok := err.(*llm.MissingKeyError); ok {
		*target = mk
		return true
	}
	return false
}
