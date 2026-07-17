package tools

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/bmatcuk/doublestar/v4"
	"github.com/rushikeshg25/cool-code/internal/types"
)

const maxGrepFileSize = 5 * 1024 * 1024

var globTool = Tool{
	Name: "glob",
	Description: "Finds files matching a glob pattern (e.g. src/**/*.go, **/*.md), returning absolute paths. " +
		"Respects .gitignore and skips node_modules.",
	ReadOnly: true,
	Schema: obj(map[string]any{
		"pattern": strProp("The glob pattern to match against (e.g. **/*.py, docs/*.md)."),
	}, "pattern"),
	Execute: func(ctx Context, args json.RawMessage) types.ToolResult {
		var a struct {
			Pattern string `json:"pattern"`
		}
		if err := json.Unmarshal(args, &a); err != nil {
			return fail("Invalid arguments", err.Error())
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
			if ctx.GitIgnore != nil && ctx.GitIgnore(rel) {
				continue
			}
			abs := filepath.Join(ctx.RootDir, rel)
			if info, err := os.Stat(abs); err == nil && !info.IsDir() {
				out = append(out, abs)
			}
		}
		return types.ToolResult{
			Display:   "Found " + itoa(len(out)) + " file(s)",
			LLMResult: strings.Join(out, "\n"),
		}
	},
}

var grepTool = Tool{
	Name: "grep",
	Description: "Searches for a regular expression within file contents under a directory. " +
		"Optionally filter files by an include regex. Returns matching lines with file paths and line numbers.",
	ReadOnly: true,
	Schema: obj(map[string]any{
		"pattern": strProp("The regular expression to search for within file contents."),
		"path":    strProp("Optional absolute path to a directory or file to search. Defaults to project root."),
		"include": strProp("Optional regex to filter which file basenames are searched."),
	}, "pattern"),
	Execute: func(ctx Context, args json.RawMessage) types.ToolResult {
		var a struct {
			Pattern string `json:"pattern"`
			Path    string `json:"path"`
			Include string `json:"include"`
		}
		if err := json.Unmarshal(args, &a); err != nil {
			return fail("Invalid arguments", err.Error())
		}
		re, err := regexp.Compile(a.Pattern)
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
		var matches []string
		search := func(file string) {
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
					matches = append(matches, file+":"+itoa(i+1)+": "+line)
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
		if len(matches) == 0 {
			return types.ToolResult{Display: "No matches found", LLMResult: "No matches found."}
		}
		return types.ToolResult{
			Display:   "Found " + itoa(len(matches)) + " match(es)",
			LLMResult: strings.Join(matches, "\n"),
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
		includeFlag := ""
		if a.Include != "" {
			includeFlag = " -g '" + shellEscapeSingleQuotes(a.Include) + "'"
		}
		command := "rg -n --hidden --glob '!.git/*' --glob '!node_modules/*'" + includeFlag +
			" '" + shellEscapeSingleQuotes(a.Pattern) + "' '" + shellEscapeSingleQuotes(searchPath) + "'"
		res := execCommand(ctx.Context(), command, ctx.RootDir, 0)
		display := "Symbol search results"
		if !res.success {
			display = "Symbol search failed"
		}
		return types.ToolResult{Display: display, LLMResult: res.combined()}
	},
}

func isBinary(buf []byte) bool {
	n := len(buf)
	if n > 8000 {
		n = 8000
	}
	return bytes.IndexByte(buf[:n], 0) != -1
}
