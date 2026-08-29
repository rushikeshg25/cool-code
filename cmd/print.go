package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"

	"github.com/rushikeshg25/cool-code/internal/agent"
	"github.com/rushikeshg25/cool-code/internal/config"
	"github.com/rushikeshg25/cool-code/internal/types"
)

// runPrint executes one turn without the TUI and writes the result to stdout,
// so the agent can be scripted or run in CI. The prompt comes from the
// argument, or from stdin when it is piped.
//
// The binary previously could not be driven at all: there was no non
// interactive flag and nothing read stdin. Processor and Reporter were already
// free of any UI dependency, so this is mostly plumbing.
func runPrint(flags rootFlags, args []string) error {
	loadEnv()

	prompt, err := printPrompt(args)
	if err != nil {
		return err
	}
	if strings.TrimSpace(prompt) == "" {
		return fmt.Errorf("no prompt given: pass one as an argument or pipe it on stdin")
	}

	rootDir, err := os.Getwd()
	if err != nil {
		return err
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

	mode := types.ModeAgent
	if flags.mode != "" {
		parsed, ok := parseMode(flags.mode)
		if !ok {
			return fmt.Errorf("invalid mode %q (use plan, agent, or ask)", flags.mode)
		}
		mode = parsed
	}

	proc, err := agent.New(rootDir, cfg, agent.Options{
		Mode:           mode,
		AllowDangerous: cfg.AllowDangerous(),
	})
	if err != nil {
		return handleProviderError(err)
	}

	// There is no one to answer a prompt in this mode. Without
	// --allow-dangerous a confirmation is a refusal, which is the safe
	// default: a scripted run must not silently approve a shell command.
	approved := flags.allowDangerous
	proc.SetConfirmHandlers(
		func(string) bool { return approved },
		func(string, string) bool { return approved },
	)

	rep := &printReporter{verbose: flags.verbose, out: os.Stderr}
	final, err := proc.ProcessQuery(context.Background(), prompt, rep)
	if err != nil {
		return err
	}

	if flags.printJSON {
		status := proc.GetStatus()
		out := map[string]any{
			"result":   final,
			"mode":     string(mode),
			"model":    status.Model,
			"messages": status.MessageCount,
			"tools":    rep.tools,
		}
		if status.CostKnown {
			out["cost_usd"] = status.SessionCost
		}
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(out)
	}

	fmt.Println(final)
	return nil
}

// printPrompt takes the prompt from the arguments, or from stdin when it is
// piped rather than attached to a terminal.
func printPrompt(args []string) (string, error) {
	if len(args) > 0 {
		return strings.Join(args, " "), nil
	}
	info, err := os.Stdin.Stat()
	if err != nil {
		return "", err
	}
	if info.Mode()&os.ModeCharDevice != 0 {
		// A terminal: reading would block forever waiting for input.
		return "", nil
	}
	raw, err := io.ReadAll(os.Stdin)
	if err != nil {
		return "", err
	}
	return string(raw), nil
}

func parseMode(name string) (types.AgentMode, bool) {
	switch strings.ToLower(strings.TrimSpace(name)) {
	case "plan":
		return types.ModePlan, true
	case "agent":
		return types.ModeAgent, true
	case "ask":
		return types.ModeAsk, true
	}
	return "", false
}

// printReporter sends progress to stderr, leaving stdout for the result alone
// so the output stays pipeable.
type printReporter struct {
	mu      sync.Mutex
	verbose bool
	out     io.Writer
	tools   []string
}

func (r *printReporter) Status(string)         {}
func (r *printReporter) AssistantDelta(string) {}
func (r *printReporter) Assistant(string)      {}
func (r *printReporter) AssistantDiscard()     {}

func (r *printReporter) Tool(name, display string, failed bool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.tools = append(r.tools, name)
	if !r.verbose {
		return
	}
	marker := "•"
	if failed {
		marker = "✗"
	}
	fmt.Fprintf(r.out, "%s %s: %s\n", marker, name, display)
}

func (r *printReporter) Compacted(note string) {
	if r.verbose {
		fmt.Fprintln(r.out, "• "+note)
	}
}

func (r *printReporter) Tasks(*types.TaskList) {}

func (r *printReporter) Subagents(lines []string) {
	if !r.verbose {
		return
	}
	for _, line := range lines {
		fmt.Fprintln(r.out, "• "+line)
	}
}
