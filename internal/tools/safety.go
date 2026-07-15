package tools

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"unicode"

	"github.com/rushikeshg25/cool-code/internal/config"
)

// BlockedPath returns a non-empty reason when reading filePath is disallowed by
// the configured guardrails, otherwise "".
func BlockedPath(filePath string, cfg config.Config) string {
	patterns := cfg.Guardrails.BlockReadPatterns
	if len(patterns) == 0 {
		return ""
	}
	base := filepath.Base(filePath)
	for _, pattern := range patterns {
		switch {
		case pattern == base:
			return "Reading blocked for \"" + base + "\" by guardrails."
		case strings.HasPrefix(pattern, ".") && strings.HasSuffix(pattern, ".*"):
			if strings.HasPrefix(base, strings.TrimSuffix(pattern, ".*")) {
				return "Reading blocked for \"" + base + "\" by guardrails."
			}
		case strings.HasPrefix(pattern, "*."):
			if strings.HasSuffix(base, pattern[1:]) {
				return "Reading blocked for \"" + base + "\" by guardrails."
			}
		}
	}
	return ""
}

// EnsureAbsoluteWithinRoot verifies absPath is absolute and contained within
// rootPath, returning a non-empty error message otherwise.
func EnsureAbsoluteWithinRoot(absPath, rootPath string) string {
	if !filepath.IsAbs(absPath) {
		return "File path must be absolute"
	}
	resolvedRoot, _ := filepath.Abs(rootPath)
	resolvedPath, _ := filepath.Abs(absPath)
	if resolvedPath != resolvedRoot && !strings.HasPrefix(resolvedPath, resolvedRoot+string(filepath.Separator)) {
		return "Path must be within project root: " + resolvedRoot
	}
	return ""
}

func readPackageJSON(rootPath string) (map[string]any, bool) {
	raw, err := os.ReadFile(filepath.Join(rootPath, "package.json"))
	if err != nil {
		return nil, false
	}
	var pkg map[string]any
	if err := json.Unmarshal(raw, &pkg); err != nil {
		return nil, false
	}
	return pkg, true
}

func writePackageJSON(rootPath string, pkg map[string]any) error {
	data, err := json.MarshalIndent(pkg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(rootPath, "package.json"), append(data, '\n'), 0o644)
}

func shellEscapeSingleQuotes(value string) string {
	return strings.ReplaceAll(value, "'", `'\"'\"'`)
}

func toRelative(filePath, rootPath string) string {
	rel, err := filepath.Rel(rootPath, filePath)
	if err != nil {
		return filePath
	}
	return rel
}

func toPascalCase(input string) string {
	var parts []string
	var cur strings.Builder
	for _, r := range input {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			cur.WriteRune(r)
		} else if cur.Len() > 0 {
			parts = append(parts, cur.String())
			cur.Reset()
		}
	}
	if cur.Len() > 0 {
		parts = append(parts, cur.String())
	}
	var out strings.Builder
	for _, p := range parts {
		out.WriteString(strings.ToUpper(p[:1]) + p[1:])
	}
	return out.String()
}
