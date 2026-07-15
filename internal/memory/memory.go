// Package memory loads persistent project instructions (COOLCODE.md).
package memory

import (
	"os"
	"path/filepath"
	"strings"
)

// File is the project memory filename.
const File = "COOLCODE.md"

// LoadProjectInstructions returns the project COOLCODE.md, falling back to the
// global ~/.coolcode/COOLCODE.md, or "" when neither exists.
func LoadProjectInstructions(rootDir string) string {
	home, _ := os.UserHomeDir()
	candidates := []string{
		filepath.Join(rootDir, File),
		filepath.Join(home, ".coolcode", File),
	}
	for _, c := range candidates {
		if raw, err := os.ReadFile(c); err == nil {
			if content := strings.TrimSpace(string(raw)); content != "" {
				return content
			}
		}
	}
	return ""
}
