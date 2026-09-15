package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/google/uuid"

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

			if !refresh && cfg.ScanCache() {
				if raw, err := readScanCache(rootDir); err == nil {
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
					_ = writeScanCache(rootDir, data)
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

const scanCacheName = ".coolcode.scan.json"

func readScanCache(rootDir string) ([]byte, error) {
	root, err := os.OpenRoot(rootDir)
	if err != nil {
		return nil, err
	}
	defer root.Close()
	info, err := root.Lstat(scanCacheName)
	if err != nil {
		return nil, err
	}
	if !info.Mode().IsRegular() {
		return nil, fmt.Errorf("scan cache is not a regular file")
	}
	return root.ReadFile(scanCacheName)
}

// Replace the directory entry rather than truncating a potentially linked
// target. Root keeps even concurrent symlink replacements inside the workspace.
func writeScanCache(rootDir string, data []byte) error {
	root, err := os.OpenRoot(rootDir)
	if err != nil {
		return err
	}
	defer root.Close()
	if info, err := root.Lstat(scanCacheName); err == nil {
		if !info.Mode().IsRegular() {
			return fmt.Errorf("scan cache is not a regular file")
		}
	} else if !os.IsNotExist(err) {
		return err
	}
	temp := ".coolcode-scan-" + uuid.NewString()
	file, err := root.OpenFile(temp, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o600)
	if err != nil {
		return err
	}
	defer root.Remove(temp)
	_, writeErr := file.Write(data)
	closeErr := file.Close()
	if writeErr != nil {
		return writeErr
	}
	if closeErr != nil {
		return closeErr
	}
	return root.Rename(temp, scanCacheName)
}
