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
	if !withinRoot(resolvedPath, resolvedRoot) {
		return "Path must be within project root: " + resolvedRoot
	}
	return ""
}

// ValidateReadPath applies the same root jail and guardrails to every tool
// capable of returning local file contents to the model.
func ValidateReadPath(filePath string, ctx Context) string {
	_, reason := ResolveReadPath(filePath, ctx)
	return reason
}

// ResolveReadPath validates filePath and returns the canonical path to read.
// Tools that go on to open the file must use the returned path: validating one
// string and opening another is what lets "dir/link/../secret" past the jail.
func ResolveReadPath(filePath string, ctx Context) (string, string) {
	resolved, reason := ResolveWithinRoots(filePath, ctx.Roots())
	if reason != "" {
		return "", reason
	}
	// Check guardrails against both spellings so a link cannot launder a
	// blocked name into an allowed one, or the reverse.
	if blocked := BlockedPath(filePath, ctx.Config); blocked != "" {
		return "", blocked
	}
	if blocked := BlockedPath(resolved, ctx.Config); blocked != "" {
		return "", blocked
	}
	return resolved, ""
}

// protectedWriteComponents names directories a tool must never write into,
// even though they sit inside the workspace. A write to .git/config or
// .git/hooks turns the next git command into code execution and outlives the
// session, and .coolcode holds skills, settings and credentials.
var protectedWriteComponents = map[string]bool{
	".git":      true,
	".coolcode": true,
}

// protectedWriteNames names individual files a tool must never write.
var protectedWriteNames = map[string]bool{
	".coolcode.json": true,
}

// ResolveWritePath validates absPath for writing and returns the canonical
// path callers must write to.
func ResolveWritePath(absPath string, ctx Context) (string, string) {
	resolved, reason := ResolveWithinRoots(absPath, ctx.Roots())
	if reason != "" {
		return "", reason
	}
	if reason := protectedWrite(resolved, ctx.Roots()); reason != "" {
		return "", reason
	}
	// Guardrailed files are off limits to writes as well as reads. edit_file
	// in particular reports what it wrote, so an edit is also a disclosure.
	if blocked := BlockedPath(absPath, ctx.Config); blocked != "" {
		return "", blocked
	}
	if blocked := BlockedPath(resolved, ctx.Config); blocked != "" {
		return "", blocked
	}
	return resolved, ""
}

// GitExcludePathspecs renders the read guardrails as git pathspecs so the git
// tools cannot print the contents of a blocked file.
func GitExcludePathspecs(cfg config.Config) []string {
	var specs []string
	for _, pattern := range cfg.Guardrails.BlockReadPatterns {
		pattern = filepath.ToSlash(strings.TrimSpace(pattern))
		if pattern == "" {
			continue
		}
		specs = append(specs, ":(exclude,glob)"+pattern)
		if !strings.HasPrefix(pattern, "**/") {
			specs = append(specs, ":(exclude,glob)**/"+strings.TrimPrefix(pattern, "/"))
		}
	}
	return specs
}

// protectedWrite reports why resolved must not be written, or "".
func protectedWrite(resolved string, roots []string) string {
	rel := resolved
	for _, root := range roots {
		resolvedRoot, err := canonicalPath(root)
		if err != nil || !withinRoot(resolved, resolvedRoot) {
			continue
		}
		if candidate, err := filepath.Rel(resolvedRoot, resolved); err == nil {
			rel = candidate
			break
		}
	}
	for _, part := range strings.Split(filepath.ToSlash(rel), "/") {
		if protectedWriteComponents[part] {
			return "Writing inside \"" + part + "\" is not allowed."
		}
	}
	if base := filepath.Base(rel); protectedWriteNames[base] {
		return "Writing \"" + base + "\" is not allowed."
	}
	return ""
}

// ResolveWithinRoots validates absPath against roots and returns the canonical
// path callers must use for I/O.
func ResolveWithinRoots(absPath string, roots []string) (string, string) {
	if !filepath.IsAbs(absPath) {
		return "", "File path must be absolute"
	}
	resolved, err := canonicalPath(absPath)
	if err != nil {
		return "", "Could not resolve file path"
	}
	for _, root := range roots {
		resolvedRoot, err := canonicalPath(root)
		if err != nil {
			continue
		}
		if withinRoot(resolved, resolvedRoot) {
			return resolved, ""
		}
	}
	return "", "Path must be within project root or an added directory: " + strings.Join(roots, ", ")
}

func withinRoot(resolvedPath, resolvedRoot string) bool {
	rel, err := filepath.Rel(resolvedRoot, resolvedPath)
	if err != nil || filepath.IsAbs(rel) {
		return false
	}
	return rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator))
}

// maxSymlinkHops bounds link traversal so a symlink cycle cannot spin forever.
const maxSymlinkHops = 40

// canonicalPath resolves path the way the operating system does: every
// component is resolved in turn, and ".." is applied to whatever the preceding
// components resolved to. Components that do not exist yet are kept as-is, so
// writes to new files are checked against their real parent directory.
//
// Resolving component by component matters. filepath.Clean (and therefore
// filepath.Abs and filepath.EvalSymlinks) collapses ".." lexically before any
// link is followed, so "dir/link/../x" cleans to "dir/x" while the kernel
// reads it as "<link target>/../x". Callers must perform I/O on the path
// returned here, never on the string they were given.
func canonicalPath(path string) (string, error) {
	if path == "" {
		return "", os.ErrInvalid
	}
	sep := string(filepath.Separator)
	if !filepath.IsAbs(path) {
		wd, err := os.Getwd()
		if err != nil {
			return "", err
		}
		// Concatenate rather than filepath.Join, which would clean the path.
		path = wd + sep + path
	}

	vol := filepath.VolumeName(path)
	resolved := vol + sep
	hops := 0

	var walk func(string) error
	walk = func(p string) error {
		for _, part := range strings.Split(p, sep) {
			switch part {
			case "", ".":
				continue
			case "..":
				resolved = filepath.Dir(resolved)
				continue
			}
			next := filepath.Join(resolved, part)
			target, err := os.Readlink(next)
			if err != nil {
				// Not a symlink, or does not exist yet. Either way it
				// contributes itself and nothing needs following.
				resolved = next
				continue
			}
			hops++
			if hops > maxSymlinkHops {
				return os.ErrInvalid
			}
			if filepath.IsAbs(target) {
				resolved = filepath.VolumeName(target) + sep
			}
			// A relative target resolves against the link's directory, which
			// is exactly what resolved still holds.
			if err := walk(target); err != nil {
				return err
			}
		}
		return nil
	}

	if err := walk(path[len(vol):]); err != nil {
		return "", err
	}
	return filepath.Clean(resolved), nil
}

// EnsureAbsoluteWithinRoots verifies absPath is absolute and contained within
// any of roots, returning a non-empty error message otherwise. Prefer
// ResolveWithinRoots when the caller goes on to open the path.
func EnsureAbsoluteWithinRoots(absPath string, roots []string) string {
	_, reason := ResolveWithinRoots(absPath, roots)
	return reason
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

// toRelative renders filePath for display relative to rootPath. Both sides are
// canonicalized first: callers now pass resolved paths, and comparing one of
// those against an unresolved root (/var against /private/var, say) yields a
// long "../.." chain instead of a readable name.
func toRelative(filePath, rootPath string) string {
	if rel, err := filepath.Rel(rootPath, filePath); err == nil && !strings.HasPrefix(rel, "..") {
		return rel
	}
	resolvedRoot, rootErr := canonicalPath(rootPath)
	resolvedFile, fileErr := canonicalPath(filePath)
	if rootErr != nil || fileErr != nil {
		return filePath
	}
	rel, err := filepath.Rel(resolvedRoot, resolvedFile)
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
