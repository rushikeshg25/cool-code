package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/rushikeshg25/cool-code/internal/config"
	"github.com/rushikeshg25/cool-code/internal/project"
)

func scanCmd() *cobra.Command {
	var refresh, asJSON bool
	c := &cobra.Command{
		Use:   "scan",
		Short: "Summarize the current project structure",
		RunE: func(cmd *cobra.Command, args []string) error {
			rootDir, _ := os.Getwd()
			cfg := config.Load(rootDir)
			cachePath := filepath.Join(rootDir, ".coolcode.scan.json")

			if !refresh && cfg.ScanCache() {
				if raw, err := os.ReadFile(cachePath); err == nil {
					var cached project.Scan
					if json.Unmarshal(raw, &cached) == nil {
						printScan(cached, asJSON)
						return nil
					}
				}
			}
			scan := project.ScanProject(rootDir)
			if cfg.ScanCache() {
				if data, err := json.MarshalIndent(scan, "", "  "); err == nil {
					_ = os.WriteFile(cachePath, data, 0o644)
				}
			}
			printScan(scan, asJSON)
			return nil
		},
	}
	c.Flags().BoolVar(&refresh, "refresh", false, "Refresh cached scan results")
	c.Flags().BoolVar(&asJSON, "json", false, "Output raw JSON")
	return c
}

func printScan(scan project.Scan, asJSON bool) {
	if asJSON {
		out, _ := json.MarshalIndent(scan, "", "  ")
		fmt.Println(string(out))
		return
	}
	fmt.Println("\nProject Scan")
	fmt.Printf("  Root:         %s\n", scan.RootDir)
	fmt.Printf("  Entry points: %s\n", list(scan.Entrypoints))
	fmt.Printf("  Frameworks:   %s\n", list(scan.Frameworks))
	fmt.Printf("  Scripts:      %s\n", list(scan.Scripts))
	fmt.Printf("  Languages:    %s\n", list(scan.Languages))
	fmt.Printf("  tsconfig:     %s\n", yesno(scan.HasTsConfig))
	fmt.Printf("  README:       %s\n", yesno(scan.HasReadme))
}

func list(items []string) string {
	if len(items) == 0 {
		return "none"
	}
	return strings.Join(items, ", ")
}

func yesno(b bool) string {
	if b {
		return "yes"
	}
	return "no"
}
