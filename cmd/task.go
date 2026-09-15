package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"

	"github.com/rushikeshg25/cool-code/internal/agent"
	"github.com/rushikeshg25/cool-code/internal/config"
	"github.com/rushikeshg25/cool-code/internal/security"
)

func taskCmd() *cobra.Command {
	var asJSON bool
	c := &cobra.Command{
		Use:   "task <goal>",
		Short: "Generate a structured plan for a task",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			loadEnv()
			rootDir, _ := os.Getwd()
			cfg := config.Load(rootDir)
			plan, err := agent.CreateTaskPlan(context.Background(), cfg, args[0])
			if err != nil {
				return handleProviderError(err)
			}
			if plan == nil {
				fmt.Println("Could not generate a plan. Try rephrasing the goal.")
				return nil
			}
			return printTaskPlan(os.Stdout, plan, asJSON)
		},
	}
	c.Flags().BoolVar(&asJSON, "json", false, "Output raw JSON")
	return c
}

func printTaskPlan(out io.Writer, plan *agent.TaskPlan, asJSON bool) error {
	if asJSON {
		enc := json.NewEncoder(out)
		enc.SetIndent("", "  ")
		return enc.Encode(plan)
	}
	fmt.Fprintf(out, "\nGoal: %s\n\nSteps:\n", security.SanitizeTerminal(plan.Goal))
	for i, step := range plan.Steps {
		fmt.Fprintf(out, "%d. %s\n   %s\n", i+1, security.SanitizeLine(step.Title), security.SanitizeTerminal(step.Detail))
	}
	fmt.Fprintf(out, "\nAssumptions: %s\n", security.SanitizeTerminal(list(plan.Assumptions)))
	fmt.Fprintf(out, "Risks: %s\n", security.SanitizeTerminal(list(plan.Risks)))
	return nil
}
