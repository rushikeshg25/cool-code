package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/rushikeshg25/cool-code/internal/agent"
	"github.com/rushikeshg25/cool-code/internal/config"
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
			if asJSON {
				out, _ := json.MarshalIndent(plan, "", "  ")
				fmt.Println(string(out))
				return nil
			}
			fmt.Printf("\nGoal: %s\n\nSteps:\n", plan.Goal)
			for i, step := range plan.Steps {
				fmt.Printf("%d. %s\n   %s\n", i+1, step.Title, step.Detail)
			}
			fmt.Printf("\nAssumptions: %s\n", list(plan.Assumptions))
			fmt.Printf("Risks: %s\n", list(plan.Risks))
			return nil
		},
	}
	c.Flags().BoolVar(&asJSON, "json", false, "Output raw JSON")
	return c
}
