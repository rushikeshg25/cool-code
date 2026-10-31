package cmd

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"

	"github.com/rushikeshg25/cool-code/internal/config"
	"github.com/rushikeshg25/cool-code/internal/llm"
	"github.com/rushikeshg25/cool-code/internal/security"
)

func modelsCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "models",
		Short: "List models available through the configured CLI proxy (dev)",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			loadEnv()
			root, err := os.Getwd()
			if err != nil {
				return err
			}
			ctx, cancel := context.WithTimeout(cmd.Context(), 15*time.Second)
			defer cancel()
			ids, err := llm.ProxyModels(ctx, config.Load(root).LLM)
			if err != nil {
				return handleProviderError(err)
			}
			if len(ids) == 0 {
				return fmt.Errorf("CLI proxy reports no available models")
			}
			for _, id := range ids {
				fmt.Fprintln(cmd.OutOrStdout(), security.SanitizeLine(security.Redact(id)))
			}
			return nil
		},
	}
}
