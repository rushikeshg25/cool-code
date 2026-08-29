package tools

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"unicode"

	"github.com/bmatcuk/doublestar/v4"
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
	normalized := strings.TrimPrefix(filepath.ToSlash(filepath.Clean(filePath)), "/")
	for _, pattern := range patterns {
		pattern = filepath.ToSlash(strings.TrimSpace(pattern))
		matched, _ := doublestar.Match(pattern, base)
		if !matched {
			matched, _ = doublestar.Match(pattern, normalized)
		}
		if !matched {
			matched, _ = doublestar.Match("**/"+strings.TrimPrefix(pattern, "/"), normalized)
		}
		if matched {
			return "Reading blocked for \"" + base + "\" by guardrails."
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
	resolvedRoot, err := canonicalPath(rootPath)
	if err != nil {
		return "Could not resolve project root"
	}
	resolvedPath, err := canonicalPath(absPath)
	if err != nil {
		return "Could not resolve file path"
	}
	rel, err := filepath.Rel(resolvedRoot, resolvedPath)
	if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) || filepath.IsAbs(rel) {
		return "Path must be within project root: " + resolvedRoot
	}
	return ""
}

// ValidateReadPath applies the same root jail and guardrails to every tool
// capable of returning local file contents to the model.
func ValidateReadPath(filePath string, ctx Context) string {
	if reason := EnsureAbsoluteWithinRoots(filePath, ctx.Roots()); reason != "" {
		return reason
	}
	return BlockedPath(filePath, ctx.Config)
}

// canonicalPath resolves symlinks in the existing portion of path and then
// appends any not-yet-created suffix. This protects both reads and writes.
func canonicalPath(path string) (string, error) {
	abs, err := filepath.Abs(filepath.Clean(path))
	if err != nil {
		return "", err
	}
	cur := abs
	var suffix []string
	for {
		resolved, evalErr := filepath.EvalSymlinks(cur)
		if evalErr == nil {
			for i := len(suffix) - 1; i >= 0; i-- {
				resolved = filepath.Join(resolved, suffix[i])
			}
			return filepath.Clean(resolved), nil
		}
		parent := filepath.Dir(cur)
		if parent == cur {
			return abs, nil
		}
		suffix = append(suffix, filepath.Base(cur))
		cur = parent
	}
}

// EnsureAbsoluteWithinRoots verifies absPath is absolute and contained within
// any of roots, returning a non-empty error message otherwise.
func EnsureAbsoluteWithinRoots(absPath string, roots []string) string {
	if !filepath.IsAbs(absPath) {
		return "File path must be absolute"
	}
	for _, root := range roots {
		if EnsureAbsoluteWithinRoot(absPath, root) == "" {
			return ""
		}
	}
	return "Path must be within project root or an added directory: " + strings.Join(roots, ", ")
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

// pathArg prepares a relative path for use as a command argument. Paths are
// passed to argv directly, so the only remaining hazard is a leading dash
// being read as an option by tools that do not support a "--" separator.
func pathArg(rel string) string {
	if rel == "" || filepath.IsAbs(rel) || strings.HasPrefix(rel, "./") {
		return rel
	}
	return "./" + rel
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
