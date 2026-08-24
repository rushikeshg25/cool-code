package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/rushikeshg25/cool-code/internal/config"
)

func configCmd() *cobra.Command {
	c := &cobra.Command{
		Use:   "config",
		Short: "Get or set configuration in .coolcode.json",
	}

	get := &cobra.Command{
		Use:   "get <key>",
		Short: "Read a config value (e.g. llm.model)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			rootDir, _ := os.Getwd()
			cfg := config.Load(rootDir)
			value, ok := config.GetByPath(cfg, args[0])
			if !ok {
				fmt.Printf("\nNo value set for %q.\nConfig file: %s\n", args[0], config.Path(rootDir))
				return nil
			}
			out, _ := json.MarshalIndent(value, "", "  ")
			fmt.Println(string(out))
			return nil
		},
	}

	set := &cobra.Command{
		Use:   "set <key> <value>",
		Short: "Set a config value (JSON supported, e.g. true or 1024)",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			rootDir, _ := os.Getwd()
			path, err := config.Set(rootDir, args[0], config.ParseValue(args[1]))
			if err != nil {
				return err
			}
			fmt.Printf("Updated %s.\nConfig file: %s\n", args[0], path)
			return nil
		},
	}

	c.AddCommand(get, set)
	return c
}
