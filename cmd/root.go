// Package cmd wires the cool-code command-line interface.
package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

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
var Version = "2.3.0"

type rootFlags struct {
	yes            bool
	noInit         bool
	allowDangerous bool
	copy           bool
	continueSess   bool
	resumeID       string
	effort         string
	print          bool
	printJSON      bool
	verbose        bool
	mode           string
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
		Args:          cobra.ArbitraryArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if flags.print || flags.printJSON {
				return runPrint(flags, args)
			}
			return runInteractive(flags)
		},
	}

	root.Flags().BoolVarP(&flags.yes, "yes", "y", false, "Start without the initialization prompt")
	root.Flags().BoolVar(&flags.noInit, "no-init", false, "Exit without initializing in the current directory")
	root.Flags().BoolVar(&flags.allowDangerous, "allow-dangerous", false, "Allow dangerous actions without prompting")
	root.Flags().BoolVar(&flags.copy, "copy", false, "Copy final responses to the clipboard")
	root.Flags().BoolVar(&flags.continueSess, "continue", false, "Resume the most recent session for this directory")
	root.Flags().StringVar(&flags.resumeID, "resume", "", "Resume a specific session by id")
	root.Flags().StringVar(&flags.effort, "effort", "", "Reasoning effort: minimal, low, medium, high, or xhigh")
	root.Flags().BoolVarP(&flags.print, "print", "p", false, "Run one turn without the TUI and print the result")
	root.Flags().BoolVar(&flags.printJSON, "json", false, "Like --print, but emit a JSON object")
	root.Flags().BoolVarP(&flags.verbose, "verbose", "v", false, "With --print, report tool activity on stderr")
	root.Flags().StringVar(&flags.mode, "mode", "", "Starting mode: plan, agent, or ask")

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
	if flags.effort != "" {
		flags.effort = strings.ToLower(flags.effort)
		if !config.ValidReasoningEffort(flags.effort) {
			return fmt.Errorf("invalid effort %q (use minimal, low, medium, high, or xhigh)", flags.effort)
		}
		cfg.LLM.ReasoningEffort = flags.effort
	}
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
		banner += "\n  No API key configured - run /connect to link a provider."
	}
	// A repository cannot be allowed to set these, but discarding them in
	// silence misleads whoever wrote the file into thinking they took effect.
	if ignored := config.IgnoredProjectKeys(rootDir); len(ignored) > 0 {
		banner += "\n  Ignored global-only keys in .coolcode.json: " + strings.Join(ignored, ", ") +
			"\n  Set them with `cool-code config set <key> <value>` to apply them."
	}

	sessionID := session.NewID()
	var restoreFrom *session.Data
	switch {
	case flags.resumeID != "":
		restoreFrom = session.Load(flags.resumeID)
		// A session carries the extra roots granted by /add-dir, and those are
		// replayed below. Resuming one recorded in a different directory would
		// re-grant them here, where the user never approved them.
		if restoreFrom != nil && !sameDir(restoreFrom.Cwd, rootDir) {
			banner += "\n  Session " + flags.resumeID + " belongs to " + restoreFrom.Cwd + "; not resuming here."
			restoreFrom = nil
		}
	case flags.continueSess:
		restoreFrom = session.Latest(rootDir)
	}
	if restoreFrom != nil {
		var messages []llm.Message
		_ = json.Unmarshal(restoreFrom.Messages, &messages)
		proc.Restore(messages, restoreFrom.Summary, restoreFrom.PinnedFiles, types.AgentMode(restoreFrom.Mode))
		for _, d := range restoreFrom.ExtraDirs {
			_, _ = proc.AddDir(d) // dir may have been removed since; ignore
		}
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

// loadEnv imports only API-key-shaped values from a local .env. Endpoint,
// proxy, and process-control variables are intentionally ignored so opening a
// repository cannot redirect provider traffic or alter command execution.
func loadEnv() {
	info, err := os.Lstat(".env")
	if err != nil || !info.Mode().IsRegular() {
		return
	}
	values, err := godotenv.Read()
	if err != nil {
		return
	}
	for name, value := range values {
		upper := strings.ToUpper(name)
		if !strings.HasSuffix(upper, "_API_KEY") {
			continue
		}
		// COOLCODE_API_KEY is consulted ahead of the stored /connect
		// credential and needs no global setting to take effect, so honouring
		// it from a repository file would let a cloned project substitute the
		// key every request is billed and logged against. A named *_API_KEY is
		// inert unless the user's own global llm.apiKeyEnv points at it.
		if upper == "COOLCODE_API_KEY" {
			continue
		}
		if _, exists := os.LookupEnv(name); !exists {
			_ = os.Setenv(name, value)
		}
	}
}

// handleProviderError prints a friendly setup hint for a missing API key.
func handleProviderError(err error) error {
	var missing *llm.MissingKeyError
	if ok := asMissingKey(err, &missing); ok {
		fmt.Fprintf(os.Stderr, "\n  Missing API key for %s.\n\n", missing.Provider)
		fmt.Fprintf(os.Stderr, "  Run cool-code and use /connect to link a provider, or set:\n    export %s=your_api_key_here\n\n", missing.EnvKey)
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

// sameDir reports whether two directory paths refer to the same place, after
// resolving symlinks so /tmp and /private/tmp compare equal.
func sameDir(a, b string) bool {
	if a == "" || b == "" {
		return false
	}
	if a == b {
		return true
	}
	ra, errA := filepath.EvalSymlinks(a)
	rb, errB := filepath.EvalSymlinks(b)
	if errA != nil || errB != nil {
		return false
	}
	return ra == rb
}
