package tools

import (
	"bufio"
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/rushikeshg25/cool-code/internal/types"
)

var readFileTool = Tool{
	Name: "read_file",
	Description: "Reads and returns the content of a specified file from the local filesystem. " +
		"For text files, it can read specific line ranges.",
	ReadOnly: true,
	Schema: obj(map[string]any{
		"absolutePath": strProp("The absolute path to the file to read. Relative paths are not supported."),
		"startLine":    numProp("Optional: 1-based line to start reading from. Requires endLine."),
		"endLine":      numProp("Optional: 1-based line to read until. Requires startLine."),
	}, "absolutePath"),
	Execute: func(ctx Context, args json.RawMessage) types.ToolResult {
		var a struct {
			AbsolutePath string `json:"absolutePath"`
			StartLine    *int   `json:"startLine"`
			EndLine      *int   `json:"endLine"`
		}
		if err := json.Unmarshal(args, &a); err != nil {
			return fail("Invalid arguments", err.Error())
		}
		resolved, reason := ResolveReadPath(a.AbsolutePath, ctx)
		if reason != "" {
			return fail("Blocked by guardrails", reason)
		}
		if v := validateFileForReading(resolved); v != "" {
			return fail("Read failed", v)
		}
		rel := toRelative(resolved, ctx.RootDir)
		if a.StartLine == nil && a.EndLine == nil {
			content, err := os.ReadFile(resolved)
			if err != nil {
				return fail("Read failed", err.Error())
			}
			return types.ToolResult{Display: "Reading " + rel, LLMResult: string(content)}
		}
		if a.StartLine == nil || a.EndLine == nil {
			return fail("Invalid arguments", "Both startLine and endLine must be provided.")
		}
		if *a.StartLine < 1 || *a.EndLine < 1 {
			return fail("Invalid arguments", "startLine and endLine must be greater than 0.")
		}
		if *a.EndLine < *a.StartLine {
			return fail("Invalid arguments", "endLine must be greater than or equal to startLine.")
		}
		f, err := os.Open(resolved)
		if err != nil {
			return fail("Read failed", err.Error())
		}
		defer f.Close()
		scanner := bufio.NewScanner(f)
		scanner.Buffer(make([]byte, 0, 64*1024), 10*1024*1024)
		var selected []string
		cur := 0
		for scanner.Scan() {
			cur++
			if cur >= *a.StartLine && cur <= *a.EndLine {
				selected = append(selected, scanner.Text())
			}
			if cur > *a.EndLine {
				break
			}
		}
		if cur < *a.StartLine {
			return fail("Read failed", "File only has fewer lines than startLine.")
		}
		return types.ToolResult{
			Display:   "Reading lines from " + rel,
			LLMResult: strings.Join(selected, "\n"),
		}
	},
}

var openFileAtTool = Tool{
	Name:        "open_file_at",
	Description: "Reads a file or a specific line range. Line numbers are 1-based.",
	ReadOnly:    true,
	Schema: obj(map[string]any{
		"absolutePath": strProp("Absolute path to the file to read."),
		"startLine":    numProp("Start line (1-based). Use with endLine."),
		"endLine":      numProp("End line (1-based). Use with startLine."),
	}, "absolutePath"),
	Execute: readFileTool.Execute,
}

var editFileTool = Tool{
	Name: "edit_file",
	Description: "Replaces literal text within a file. Provide significant surrounding context in oldString " +
		"to target a unique location. Read the file first. Set expected_replacements to replace multiple occurrences.",
	Mutating: true,
	Schema: obj(map[string]any{
		"filePath":              strProp("The absolute path to the file to modify."),
		"oldString":             strProp("The exact literal text to replace, including whitespace and indentation."),
		"newString":             strProp("The exact literal replacement text."),
		"expected_replacements": numProp("Number of replacements expected. Defaults to 1."),
	}, "filePath", "oldString", "newString"),
	Execute: func(ctx Context, args json.RawMessage) types.ToolResult {
		var a struct {
			FilePath             string `json:"filePath"`
			OldString            string `json:"oldString"`
			NewString            string `json:"newString"`
			ExpectedReplacements *int   `json:"expected_replacements"`
		}
		if err := json.Unmarshal(args, &a); err != nil {
			return fail("Invalid arguments", err.Error())
		}
		expected := 1
		if a.ExpectedReplacements != nil && *a.ExpectedReplacements > 0 {
			expected = *a.ExpectedReplacements
		}
		resolved, v := ResolveWritePath(a.FilePath, ctx)
		if v != "" {
			return fail("Edit blocked", v)
		}
		info, err := os.Stat(resolved)
		if err != nil {
			return fail("Edit failed", "File does not exist: "+a.FilePath)
		}
		if info.IsDir() {
			return fail("Edit failed", "Path is not a file: "+a.FilePath)
		}
		if a.OldString == "" {
			return fail("Invalid arguments", "oldString cannot be empty.")
		}
		raw, err := os.ReadFile(resolved)
		if err != nil {
			return fail("Edit failed", err.Error())
		}
		content := string(raw)
		count := 0
		var b strings.Builder
		idx := 0
		for count < expected {
			pos := strings.Index(content[idx:], a.OldString)
			if pos == -1 {
				break
			}
			b.WriteString(content[idx : idx+pos])
			b.WriteString(a.NewString)
			idx += pos + len(a.OldString)
			count++
		}
		if count == 0 {
			return fail("No replacements made", "No occurrences of the old string were found.")
		}
		b.WriteString(content[idx:])
		newContent := b.String()
		if err := os.WriteFile(resolved, []byte(newContent), info.Mode().Perm()); err != nil {
			return fail("Edit failed", err.Error())
		}
		return types.ToolResult{
			Display:   "Edited " + filepath.Base(resolved),
			LLMResult: newContent,
		}
	},
}

var newFileTool = Tool{
	Name:        "new_file",
	Description: "Creates a new file at the given absolute path with the provided content.",
	Mutating:    true,
	Schema: obj(map[string]any{
		"filePath": strProp("The absolute path (including filename) of the file to create."),
		"content":  strProp("The content to write into the new file."),
	}, "filePath", "content"),
	Execute: func(ctx Context, args json.RawMessage) types.ToolResult {
		var a struct {
			FilePath string `json:"filePath"`
			Content  string `json:"content"`
		}
		if err := json.Unmarshal(args, &a); err != nil {
			return fail("Invalid arguments", err.Error())
		}
		resolved, v := ResolveWritePath(a.FilePath, ctx)
		if v != "" {
			return fail("Invalid path", v)
		}
		if err := os.MkdirAll(filepath.Dir(resolved), 0o755); err != nil {
			return fail("Error creating file "+a.FilePath, err.Error())
		}
		if err := os.WriteFile(resolved, []byte(a.Content), 0o644); err != nil {
			return fail("Error creating file "+a.FilePath, err.Error())
		}
		msg := "File " + toRelative(resolved, ctx.RootDir) + " created successfully"
		return types.ToolResult{Display: msg, LLMResult: msg}
	},
}

var renameFileTool = Tool{
	Name:        "rename_file",
	Description: "Renames or moves a file within the project.",
	Mutating:    true,
	Schema: obj(map[string]any{
		"fromPath":  strProp("Absolute path to the source file."),
		"toPath":    strProp("Absolute path to the destination."),
		"overwrite": boolProp("If true, overwrite destination if it exists."),
	}, "fromPath", "toPath"),
	Execute: func(ctx Context, args json.RawMessage) types.ToolResult {
		var a struct {
			FromPath  string `json:"fromPath"`
			ToPath    string `json:"toPath"`
			Overwrite bool   `json:"overwrite"`
		}
		if err := json.Unmarshal(args, &a); err != nil {
			return fail("Invalid arguments", err.Error())
		}
		fromResolved, v := ResolveWritePath(a.FromPath, ctx)
		if v != "" {
			return fail("Invalid path", v)
		}
		toResolved, v := ResolveWritePath(a.ToPath, ctx)
		if v != "" {
			return fail("Invalid path", v)
		}
		if _, err := os.Stat(fromResolved); err != nil {
			return fail("Rename failed", "Source file does not exist: "+a.FromPath)
		}
		if _, err := os.Stat(toResolved); err == nil && !a.Overwrite {
			return fail("Rename failed", "Target already exists: "+a.ToPath)
		}
		if err := os.MkdirAll(filepath.Dir(toResolved), 0o755); err != nil {
			return fail("Rename failed", err.Error())
		}
		if err := os.Rename(fromResolved, toResolved); err != nil {
			return fail("Rename failed", err.Error())
		}
		return types.ToolResult{Display: "File renamed", LLMResult: "Renamed " + a.FromPath + " to " + a.ToPath}
	},
}

var listRecentFilesTool = Tool{
	Name:        "list_recent_files",
	Description: "Lists recently modified files in the project, most recent first.",
	ReadOnly:    true,
	Schema: obj(map[string]any{
		"limit":   numProp("Maximum number of files to return. Defaults to 20."),
		"include": strProp("Optional glob to include files (e.g. src/**/*.ts)."),
		"exclude": strProp("Optional glob to exclude files."),
	}),
	Execute: func(ctx Context, args json.RawMessage) types.ToolResult {
		var a struct {
			Limit   *int   `json:"limit"`
			Include string `json:"include"`
			Exclude string `json:"exclude"`
		}
		_ = json.Unmarshal(args, &a)
		limit := 20
		if a.Limit != nil && *a.Limit > 0 {
			limit = *a.Limit
		}
		var extra []string
		if a.Exclude != "" {
			extra = append(extra, a.Exclude)
		}
		files := globFiles(ctx.RootDir, a.Include, extra...)
		type fm struct {
			path  string
			mtime time.Time
		}
		var stats []fm
		for _, f := range files {
			if ValidateReadPath(f, ctx) != "" {
				continue
			}
			if info, err := os.Stat(f); err == nil {
				stats = append(stats, fm{f, info.ModTime()})
			}
		}
		sort.Slice(stats, func(i, j int) bool { return stats[i].mtime.After(stats[j].mtime) })
		if len(stats) > limit {
			stats = stats[:limit]
		}
		type entry struct {
			Path       string `json:"path"`
			ModifiedAt string `json:"modifiedAt"`
		}
		out := make([]entry, 0, len(stats))
		for _, s := range stats {
			out = append(out, entry{toRelative(s.path, ctx.RootDir), s.mtime.UTC().Format(time.RFC3339)})
		}
		data, _ := json.MarshalIndent(out, "", "  ")
		return types.ToolResult{
			Display:   "Found recent files",
			LLMResult: string(data),
		}
	},
}

var replaceInFilesTool = Tool{
	Name:        "replace_in_files",
	Description: "Replaces text across files. Supports dry run (default) and regex mode.",
	Mutating:    true,
	Schema: obj(map[string]any{
		"pattern":     strProp("Text or regex pattern to replace."),
		"replacement": strProp("Replacement text."),
		"include":     strProp("Optional glob to include files."),
		"exclude":     strProp("Optional glob to exclude files."),
		"useRegex":    boolProp("If true, treat pattern as a regular expression."),
		"dryRun":      boolProp("If true (default), do not write changes."),
	}, "pattern", "replacement"),
	Execute: func(ctx Context, args json.RawMessage) types.ToolResult {
		var a struct {
			Pattern     string `json:"pattern"`
			Replacement string `json:"replacement"`
			Include     string `json:"include"`
			Exclude     string `json:"exclude"`
			UseRegex    bool   `json:"useRegex"`
			DryRun      *bool  `json:"dryRun"`
		}
		if err := json.Unmarshal(args, &a); err != nil {
			return fail("Invalid arguments", err.Error())
		}
		if a.Pattern == "" {
			return fail("Invalid arguments", "pattern is required.")
		}
		var re *regexp.Regexp
		if a.UseRegex {
			var err error
			re, err = regexp.Compile(a.Pattern)
			if err != nil {
				return fail("Invalid pattern", "Invalid regex pattern: "+err.Error())
			}
		}
		dryRun := a.DryRun == nil || *a.DryRun
		var extra []string
		if a.Exclude != "" {
			extra = append(extra, a.Exclude)
		}
		files := globFiles(ctx.RootDir, a.Include, extra...)

		type fileResult struct {
			File         string `json:"file"`
			Replacements int    `json:"replacements"`
		}
		var results []fileResult
		total := 0
		for _, f := range files {
			resolved, reason := ResolveWritePath(f, ctx)
			if reason != "" || BlockedPath(f, ctx.Config) != "" {
				continue
			}
			f = resolved
			info, err := os.Stat(f)
			if err != nil {
				continue
			}
			raw, err := os.ReadFile(f)
			if err != nil {
				continue
			}
			content := string(raw)
			var replaced string
			count := 0
			if a.UseRegex {
				count = len(re.FindAllStringIndex(content, -1))
				if count > 0 {
					replaced = re.ReplaceAllString(content, a.Replacement)
				}
			} else {
				count = strings.Count(content, a.Pattern)
				if count > 0 {
					replaced = strings.ReplaceAll(content, a.Pattern, a.Replacement)
				}
			}
			if count > 0 {
				total += count
				results = append(results, fileResult{toRelative(f, ctx.RootDir), count})
				if !dryRun {
					// Keep the original mode: rewriting a 0600 file at 0644
					// would make it world readable.
					_ = os.WriteFile(f, []byte(replaced), info.Mode().Perm())
				}
			}
		}
		shown := results
		if len(shown) > 50 {
			shown = shown[:50]
		}
		summary := map[string]any{
			"dryRun":            dryRun,
			"filesChanged":      len(results),
			"totalReplacements": total,
			"files":             shown,
		}
		data, _ := json.MarshalIndent(summary, "", "  ")
		display := "Replaced " + itoa(total) + " occurrences in " + itoa(len(results)) + " files"
		if dryRun {
			display = "Dry run: " + itoa(total) + " replacements in " + itoa(len(results)) + " files"
		}
		return types.ToolResult{Display: display, LLMResult: string(data)}
	},
}

var newModuleTool = Tool{
	Name:        "new_module",
	Description: "Scaffolds a new module folder with an index export.",
	Mutating:    true,
	Schema: obj(map[string]any{
		"moduleName":          strProp("Module name (folder name)."),
		"baseDir":             strProp("Base directory relative to project root (default: src)."),
		"exportFromRootIndex": boolProp("If true, export from baseDir/index.ts."),
	}, "moduleName"),
	Execute: func(ctx Context, args json.RawMessage) types.ToolResult {
		var a struct {
			ModuleName          string `json:"moduleName"`
			BaseDir             string `json:"baseDir"`
			ExportFromRootIndex bool   `json:"exportFromRootIndex"`
		}
		if err := json.Unmarshal(args, &a); err != nil {
			return fail("Invalid arguments", err.Error())
		}
		if strings.TrimSpace(a.ModuleName) == "" {
			return fail("Invalid arguments", "moduleName is required.")
		}
		baseDir := a.BaseDir
		if baseDir == "" {
			baseDir = "src"
		}
		moduleDir := filepath.Join(ctx.RootDir, baseDir, a.ModuleName)
		if _, v := ResolveWritePath(moduleDir, ctx); v != "" {
			return fail("Invalid path", v)
		}
		if err := os.MkdirAll(moduleDir, 0o755); err != nil {
			return fail("Create failed", "Failed to create module: "+err.Error())
		}
		exportName := toPascalCase(a.ModuleName)
		if exportName == "" {
			exportName = "NewModule"
		}
		moduleFile := filepath.Join(moduleDir, a.ModuleName+".ts")
		if _, err := os.Stat(moduleFile); err != nil {
			_ = os.WriteFile(moduleFile, []byte("export const "+exportName+" = () => {\n  // TODO: implement\n};\n"), 0o644)
		}
		indexFile := filepath.Join(moduleDir, "index.ts")
		if _, err := os.Stat(indexFile); err != nil {
			_ = os.WriteFile(indexFile, []byte("export * from './"+a.ModuleName+"';\n"), 0o644)
		}
		if a.ExportFromRootIndex {
			rootIndex := filepath.Join(ctx.RootDir, baseDir, "index.ts")
			line := "export * from './" + a.ModuleName + "';\n"
			if raw, err := os.ReadFile(rootIndex); err == nil {
				if !strings.Contains(string(raw), line) {
					if f, err := os.OpenFile(rootIndex, os.O_APPEND|os.O_WRONLY, 0o644); err == nil {
						_, _ = f.WriteString(line)
						_ = f.Close()
					}
				}
			} else {
				_ = os.WriteFile(rootIndex, []byte(line), 0o644)
			}
		}
		return types.ToolResult{Display: "Module created", LLMResult: "Created " + moduleDir}
	},
}

func validateFileForReading(filePath string) string {
	if !filepath.IsAbs(filePath) {
		return "File path must be absolute"
	}
	info, err := os.Stat(filePath)
	if err != nil {
		return "File does not exist: " + filePath
	}
	if info.IsDir() {
		return "Path is not a file: " + filePath
	}
	return ""
}
