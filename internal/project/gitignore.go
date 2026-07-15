// Package project provides read-only inspection of the working directory:
// gitignore matching, an ASCII folder tree, and a lightweight project scan.
package project

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// GitIgnoreChecker reports whether a path (relative to the project root) is
// ignored by the project's .gitignore. Negation patterns are ignored, matching
// the original implementation.
type GitIgnoreChecker func(relPath string) bool

// NewGitIgnoreChecker builds a checker from rootDir/.gitignore. When the file is
// absent or unreadable, the returned checker matches nothing.
func NewGitIgnoreChecker(rootDir string) GitIgnoreChecker {
	raw, err := os.ReadFile(filepath.Join(rootDir, ".gitignore"))
	if err != nil {
		return func(string) bool { return false }
	}
	var regexes []*regexp.Regexp
	for _, line := range strings.Split(string(raw), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, "!") {
			continue
		}
		if re := compileGitIgnorePattern(line); re != nil {
			regexes = append(regexes, re)
		}
	}
	return func(relPath string) bool {
		normalized := filepath.ToSlash(relPath)
		for _, re := range regexes {
			if re.MatchString(normalized) {
				return true
			}
		}
		return false
	}
}

var reEscape = regexp.MustCompile(`[.+^${}()|\[\]\\]`)

func compileGitIgnorePattern(pattern string) *regexp.Regexp {
	isAbsolute := strings.HasPrefix(pattern, "/")
	if isAbsolute {
		pattern = pattern[1:]
	}

	escaped := reEscape.ReplaceAllString(pattern, `\$0`)
	escaped = strings.ReplaceAll(escaped, "**", "\x00")
	escaped = strings.ReplaceAll(escaped, "*", "[^/]*")
	escaped = strings.ReplaceAll(escaped, "\x00", ".*")
	escaped = strings.ReplaceAll(escaped, "?", "[^/]")

	if strings.HasSuffix(pattern, "/") {
		escaped = strings.TrimSuffix(escaped, "/") + "(/.*)?$"
	} else {
		escaped += "(/.*)?$"
	}

	if isAbsolute || strings.Contains(pattern, "/") {
		escaped = "^" + escaped
	} else {
		escaped = "(^|/)" + escaped
	}

	re, err := regexp.Compile(escaped)
	if err != nil {
		return nil
	}
	return re
}
