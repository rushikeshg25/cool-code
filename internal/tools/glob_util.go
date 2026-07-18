package tools

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/bmatcuk/doublestar/v4"
)

var defaultIgnores = []string{
	"**/node_modules/**", "**/.git/**", "**/dist/**", "**/build/**",
}

// globFiles returns absolute file paths (no directories) under rootDir matching
// the include pattern, skipping dotfiles and any of the ignore globs.
func globFiles(rootDir, include string, extraIgnore ...string) []string {
	if include == "" {
		include = "**/*"
	}
	ignores := append([]string{}, defaultIgnores...)
	ignores = append(ignores, extraIgnore...)

	fsys := os.DirFS(rootDir)
	matches, err := doublestar.Glob(fsys, include)
	if err != nil {
		return nil
	}
	var out []string
	for _, rel := range matches {
		if hasDotComponent(rel) {
			continue
		}
		if matchesAny(rel, ignores) {
			continue
		}
		abs := filepath.Join(rootDir, rel)
		info, err := os.Stat(abs)
		if err != nil || info.IsDir() {
			continue
		}
		out = append(out, abs)
	}
	return out
}

// ProjectFiles returns project-relative (slash-separated) file paths under
// rootDir, skipping dotfiles and the standard ignore directories. Used for
// @-mention file completion in the TUI.
func ProjectFiles(rootDir string) []string {
	abs := globFiles(rootDir, "**/*")
	out := make([]string, 0, len(abs))
	for _, p := range abs {
		if rel, err := filepath.Rel(rootDir, p); err == nil {
			out = append(out, filepath.ToSlash(rel))
		}
	}
	return out
}

func hasDotComponent(rel string) bool {
	for _, part := range strings.Split(filepath.ToSlash(rel), "/") {
		if strings.HasPrefix(part, ".") && part != "." && part != ".." {
			return true
		}
	}
	return false
}

func matchesAny(rel string, patterns []string) bool {
	rel = filepath.ToSlash(rel)
	for _, p := range patterns {
		if ok, _ := doublestar.Match(p, rel); ok {
			return true
		}
	}
	return false
}

// walkFiles walks rootDir invoking fn for each regular file, skipping the
// standard ignore directories.
func walkFiles(rootDir string, fn func(path string, d fs.DirEntry)) {
	_ = filepath.WalkDir(rootDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() {
			switch d.Name() {
			case "node_modules", ".git", "dist", "build":
				return filepath.SkipDir
			}
			return nil
		}
		fn(path, d)
		return nil
	})
}
