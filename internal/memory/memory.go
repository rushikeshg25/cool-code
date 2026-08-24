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
	candidates := []struct{ base, file string }{
		{rootDir, filepath.Join(rootDir, File)},
		{filepath.Join(home, ".coolcode"), filepath.Join(home, ".coolcode", File)},
	}
	for _, c := range candidates {
		baseInfo, err := os.Lstat(c.base)
		if err != nil || !baseInfo.IsDir() {
			continue
		}
		info, err := os.Lstat(c.file)
		if err != nil || !info.Mode().IsRegular() {
			continue
		}
		if raw, err := os.ReadFile(c.file); err == nil {
			if content := strings.TrimSpace(string(raw)); content != "" {
				return content
			}
		}
	}
	return ""
}
