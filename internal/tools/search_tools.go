package tools

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/bmatcuk/doublestar/v4"
	"github.com/rushikeshg25/cool-code/internal/security"
	"github.com/rushikeshg25/cool-code/internal/types"
)

const maxGrepFileSize = 5 * 1024 * 1024

var globTool = Tool{
	Name: "glob",
	Description: "Finds files matching a glob pattern (e.g. src/**/*.go, **/*.md), returning absolute paths. " +
		"Respects .gitignore and skips node_modules. Returns sorted pages; use offset and limit to continue.",
	ReadOnly: true,
	Schema: obj(map[string]any{
		"offset":  numProp("Zero-based result offset (default 0)."),
		"limit":   numProp("Page size, 1-500 (default 200)."),
		"pattern": strProp("The glob pattern to match against (e.g. **/*.py, docs/*.md)."),
	}, "pattern"),
	Execute: func(ctx Context, args json.RawMessage) types.ToolResult {
		var a struct {
			Pattern string `json:"pattern"`
			Offset  int    `json:"offset"`
			Limit   *int   `json:"limit"`
		}
		if err := json.Unmarshal(args, &a); err != nil {
			return fail("Invalid arguments", err.Error())
		}
		if a.Offset < 0 {
			return fail("Invalid offset", "offset must be nonnegative.")
		}
		limit := 200
		if a.Limit != nil {
			if *a.Limit < 1 || *a.Limit > 500 {
				return fail("Invalid limit", "limit must be between 1 and 500.")
			}
			limit = *a.Limit
		}
		fsys := os.DirFS(ctx.RootDir)
		matches, err := doublestar.Glob(fsys, a.Pattern)
		if err != nil {
			return fail("Invalid pattern", err.Error())
		}
		var out []string
		for _, rel := range matches {
			if matchesAny(rel, []string{"**/node_modules/**", "node_modules/**"}) {
				continue
			}
			// glob went straight to doublestar rather than through globFiles,
			// so "**/.git/**" listed the repository's git directory, whose
			// config commonly holds a remote URL with an embedded token.
			if hasDotComponent(rel) {
				continue
			}
			if ctx.GitIgnore != nil && ctx.GitIgnore(rel) {
				continue
			}
			abs := filepath.Join(ctx.RootDir, rel)
			if BlockedPath(abs, ctx.Config) != "" {
				continue
			}
			if info, err := os.Stat(abs); err == nil && !info.IsDir() {
				out = append(out, abs)
			}
		}
		sort.Strings(out)
		total := len(out)
		start := a.Offset
		if start > total {
			start = total
		}
		count := total - start
		if count > limit {
			count = limit
		}
		end := start + count
		text := strings.Join(out[start:end], "\n")
		if end < total {
			text += "\n[More files available; continue with offset=" + itoa(end) + ".]"
		}
		return types.ToolResult{
			Display:   "Showing " + itoa(count) + " of " + itoa(total) + " file(s)",
			LLMResult: text,
		}
	},
}

var grepTool = Tool{
	Name: "grep",
	Description: "Searches for a regular expression within file contents under a directory. " +
		"Optionally filter files by an include regex. Supports literal and case-insensitive matching. Returns bounded line previews with paths and line numbers.",
	ReadOnly: true,
	Schema: obj(map[string]any{
		"pattern":    strProp("The regular expression to search for within file contents."),
		"path":       strProp("Optional absolute path to a directory or file to search. Defaults to project root."),
		"literal":    boolProp("Treat pattern as literal text instead of a regular expression."),
		"ignoreCase": boolProp("Match without case sensitivity (default false)."),
		"limit":      numProp("Maximum matching lines to return, 1-200 (default 100). Long lines are shortened."),
		"include":    strProp("Optional regex to filter which file basenames are searched."),
	}, "pattern"),
	Execute: func(ctx Context, args json.RawMessage) types.ToolResult {
		var a struct {
			Pattern    string `json:"pattern"`
			Path       string `json:"path"`
			Include    string `json:"include"`
			Literal    bool   `json:"literal"`
			IgnoreCase bool   `json:"ignoreCase"`
			Limit      *int   `json:"limit"`
		}
		if err := json.Unmarshal(args, &a); err != nil {
			return fail("Invalid arguments", err.Error())
		}
		if a.Pattern == "" {
			return fail("Invalid pattern", "pattern must not be empty.")
		}
		limit := 100
		if a.Limit != nil {
			if *a.Limit < 1 || *a.Limit > 200 {
				return fail("Invalid limit", "limit must be between 1 and 200.")
			}
			limit = *a.Limit
		}
		pattern := a.Pattern
		if a.Literal {
			pattern = regexp.QuoteMeta(pattern)
		}
		if a.IgnoreCase {
			pattern = "(?i)" + pattern
		}
		re, err := regexp.Compile(pattern)
		if err != nil {
			return fail("Invalid pattern", "Invalid search pattern: "+err.Error())
		}
		var includeRe *regexp.Regexp
		if a.Include != "" {
			includeRe, err = regexp.Compile(a.Include)
			if err != nil {
				return fail("Invalid pattern", "Invalid include pattern: "+err.Error())
			}
		}
		searchPath := a.Path
		if searchPath == "" {
			searchPath = ctx.RootDir
		}
		if reason := EnsureAbsoluteWithinRoots(searchPath, ctx.Roots()); reason != "" {
			return fail("Search blocked", reason)
		}
		if info, err := os.Stat(searchPath); err == nil && !info.IsDir() {
			if reason := ValidateReadPath(searchPath, ctx); reason != "" {
				return fail("Search blocked", reason)
			}
		}
		var matches []string
		truncated := false
		search := func(file string) {
			if truncated || ctx.Context().Err() != nil {
				return
			}
			if ValidateReadPath(file, ctx) != "" {
				return
			}
			if rel, err := filepath.Rel(ctx.RootDir, file); err == nil && ctx.GitIgnore != nil && ctx.GitIgnore(rel) {
				return
			}
			info, err := os.Stat(file)
			if err != nil || info.Size() > maxGrepFileSize {
				return
			}
			raw, err := os.ReadFile(file)
			if err != nil || isBinary(raw) {
				return
			}
			for i, line := range strings.Split(string(raw), "\n") {
				if re.MatchString(line) {
					if len(matches) == limit {
						truncated = true
						return
					}
					line = security.Redact(line)
					preview := truncateRunes(line, 1000)
					if len(preview) < len(line) {
						preview += " [line truncated]"
					}
					matches = append(matches, file+":"+itoa(i+1)+": "+preview)
				}
			}
		}
		info, err := os.Stat(searchPath)
		if err != nil {
			return fail("Error while searching", "Error reading path: "+err.Error())
		}
		if info.IsDir() {
			walkFiles(searchPath, func(path string, _ os.DirEntry) {
				if includeRe == nil || includeRe.MatchString(filepath.Base(path)) {
					search(path)
				}
			})
		} else {
			if includeRe == nil || includeRe.MatchString(filepath.Base(searchPath)) {
				search(searchPath)
			}
		}
		if err := ctx.Context().Err(); err != nil {
			return fail("Search cancelled", err.Error())
		}
		if len(matches) == 0 {
			return types.ToolResult{Display: "No matches found", LLMResult: "No matches found."}
		}
		text := strings.Join(matches, "\n")
		if truncated {
			text += "\n[Results truncated; narrow path, include, or pattern.]"
		}
		return types.ToolResult{
			Display:   "Found " + itoa(len(matches)) + " match(es)",
			LLMResult: text,
		}
	},
}

var findSymbolTool = Tool{
	Name:        "find_symbol",
	Description: "Searches for a symbol or pattern using ripgrep (rg).",
	ReadOnly:    true,
	Schema: obj(map[string]any{
		"pattern": strProp("Regex or string pattern to search for."),
		"include": strProp("Optional glob filter for files."),
		"path":    strProp("Optional absolute path to search."),
	}, "pattern"),
	Execute: func(ctx Context, args json.RawMessage) types.ToolResult {
		var a struct {
			Pattern string `json:"pattern"`
			Include string `json:"include"`
			Path    string `json:"path"`
		}
		if err := json.Unmarshal(args, &a); err != nil {
			return fail("Invalid arguments", err.Error())
		}
		if strings.TrimSpace(a.Pattern) == "" {
			return fail("Invalid arguments", "pattern is required.")
		}
		searchPath := a.Path
		if searchPath == "" {
			searchPath = ctx.RootDir
		}
		resolved, reason := ResolveReadPath(searchPath, ctx)
		if reason != "" {
			return fail("Search blocked", reason)
		}
		info, err := os.Stat(resolved)
		if err != nil {
			return fail("Search failed", err.Error())
		}
		// Enumerate filenames without reading content. External glob syntax
		// and directory aliases cannot replace the application's read policy.
		listArgs := []string{"--files", "--null", "--hidden", "--no-follow"}
		if a.Include != "" {
			listArgs = append(listArgs, "-g", a.Include)
		}
		listArgs = append(listArgs, "-g", "!**/.git/**", "-g", "!**/node_modules/**", "--", resolved)
		listed := runCommandRaw(ctx.Context(), ctx.RootDir, 0, "rg", listArgs...)
		if !listed.success && listed.exitCode != 1 {
			return fail("Search failed", security.Redact(listed.combined()))
		}
		var permitted []string
		for _, file := range strings.Split(listed.stdout, "\x00") {
			if file == "" {
				continue
			}
			original := searchPath
			if info.IsDir() {
				rel, err := filepath.Rel(resolved, file)
				if err != nil {
					return fail("Search failed", err.Error())
				}
				original += string(filepath.Separator) + rel
			}
			if _, reason := ResolveReadPath(original, ctx); reason != "" {
				continue
			}
			canonical, reason := ResolveReadPath(file, ctx)
			if reason != "" {
				continue
			}
			permitted = append(permitted, canonical)
		}
		var output strings.Builder
		for start := 0; start < len(permitted); start += 128 {
			end := min(start+128, len(permitted))
			argv := []string{"-n", "--with-filename", "--no-heading", "--color", "never", "--", a.Pattern}
			argv = append(argv, permitted[start:end]...)
			res := execArgv(ctx.Context(), ctx.RootDir, 0, "rg", argv...)
			if !res.success && res.exitCode != 1 {
				return fail("Symbol search failed", res.combined())
			}
			output.WriteString(res.combined())
		}
		if output.Len() == 0 {
			return types.ToolResult{Display: "No matches found", LLMResult: "No matches found."}
		}
		return types.ToolResult{Display: "Symbol search results", LLMResult: output.String()}
	},
}

func isBinary(buf []byte) bool {
	n := len(buf)
	if n > 8000 {
		n = 8000
	}
	return bytes.IndexByte(buf[:n], 0) != -1
}
