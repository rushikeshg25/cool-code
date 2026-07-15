package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/rushikeshg25/cool-code/internal/skills"
)

func skillCmd() *cobra.Command {
	c := &cobra.Command{
		Use:   "skill",
		Short: "Manage skills",
	}

	var global bool
	install := &cobra.Command{
		Use:   "install <source>",
		Short: "Install a skill from a local path or git URL",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			rootDir, _ := os.Getwd()
			result := skills.Install(args[0], global, rootDir)
			if result.Error != "" {
				fmt.Fprintln(os.Stderr, "Install failed:", result.Error)
				os.Exit(1)
			}
			if len(result.Installed) == 0 {
				fmt.Println("No skills found in the source.")
				return nil
			}
			fmt.Printf("Installed %d skill(s): %s\n", len(result.Installed), strings.Join(result.Installed, ", "))
			fmt.Printf("Into: %s\n", result.Dest)
			return nil
		},
	}
	install.Flags().BoolVar(&global, "global", false, "Install to ~/.coolcode/skills instead of the project")

	c.AddCommand(install)
	return c
}
