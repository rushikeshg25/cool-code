package tools

import (
	"encoding/json"
	"path/filepath"
	"strings"

	"github.com/rushikeshg25/cool-code/internal/security"
	"github.com/rushikeshg25/cool-code/internal/types"
)

// All is the ordered list of every tool the agent can call.
var All = []Tool{
	readFileTool,
	editFileTool,
	newFileTool,
	openFileAtTool,
	renameFileTool,
	listRecentFilesTool,
	replaceInFilesTool,
	newModuleTool,
	globTool,
	grepTool,
	findSymbolTool,
	shellCommandTool,
	runTestsTool,
	lintFixTool,
	formatFileTool,
	addScriptTool,
	gitStatusTool,
	gitDiffTool,
	gitCommitTool,
	projectSummaryTool,
	generateReadmeSectionTool,
	useSkillTool,
	webFetchTool,
	webSearchTool,
}

var byName = func() map[string]Tool {
	m := make(map[string]Tool, len(All))
	for _, t := range All {
		m[t.Name] = t
	}
	return m
}()

// Lookup returns the tool with the given name.
func Lookup(name string) (Tool, bool) {
	t, ok := byName[name]
	return t, ok
}

// Run executes a named tool, returning an error result string when unknown.
func Run(ctx Context, name string, args json.RawMessage) types.ToolResult {
	t, ok := byName[name]
	if !ok {
		return fail("Unknown tool", "Unknown tool: "+name)
	}
	return t.Execute(ctx, args)
}

// DangerReason returns a short reason string when a tool call is potentially
// dangerous and should be confirmed, or "" otherwise.
func DangerReason(name string, args json.RawMessage) string {
	switch name {
	case "shell_command":
		var a struct {
			Command string `json:"command"`
		}
		_ = json.Unmarshal(args, &a)
		command := strings.TrimSpace(security.Redact(a.Command))
		if len(command) > 160 {
			command = command[:160] + "..."
		}
		return "shell command: " + command
	case "run_tests", "lint_fix":
		return "project code execution"
	case "format_file":
		// The fallback formatter is `npx prettier`, which resolves
		// ./node_modules/.bin/prettier before anything else, so a repository
		// chooses what runs here.
		var a struct {
			AbsolutePath string `json:"absolutePath"`
		}
		_ = json.Unmarshal(args, &a)
		return "run a formatter on " + filepath.Base(a.AbsolutePath)
	case "add_script":
		// Whatever lands in package.json scripts is what run_tests and
		// lint_fix will later execute through `npm run`.
		var a struct {
			Name    string `json:"name"`
			Command string `json:"command"`
		}
		_ = json.Unmarshal(args, &a)
		return "add package.json script \"" + a.Name + "\": " + security.Redact(a.Command)
	case "git_commit":
		return "create a git commit"
	case "replace_in_files":
		var a struct {
			DryRun *bool `json:"dryRun"`
		}
		_ = json.Unmarshal(args, &a)
		if a.DryRun != nil && !*a.DryRun {
			return "bulk replace (write)"
		}
	case "rename_file":
		var a struct {
			Overwrite bool `json:"overwrite"`
		}
		_ = json.Unmarshal(args, &a)
		if a.Overwrite {
			return "file overwrite rename"
		}
	}
	return ""
}

// EditPreview returns a diff-style preview for edit-like tools, or "". Removed
// lines are prefixed "- " and added lines "+ " so the UI can colorize them.
func EditPreview(name string, args json.RawMessage) string {
	truncate := func(s string) string {
		s = security.Redact(s)
		if len(s) > 400 {
			return s[:400] + "\n... (truncated)"
		}
		return s
	}
	prefixLines := func(s, prefix string) string {
		lines := strings.Split(truncate(s), "\n")
		for i, l := range lines {
			lines[i] = prefix + l
		}
		return strings.Join(lines, "\n")
	}
	switch name {
	case "edit_file":
		var a struct {
			FilePath  string `json:"filePath"`
			OldString string `json:"oldString"`
			NewString string `json:"newString"`
		}
		_ = json.Unmarshal(args, &a)
		return "File: " + a.FilePath + "\n" + prefixLines(a.OldString, "- ") + "\n" + prefixLines(a.NewString, "+ ")
	case "new_file":
		var a struct {
			FilePath string `json:"filePath"`
			Content  string `json:"content"`
		}
		_ = json.Unmarshal(args, &a)
		return "File: " + a.FilePath + "\n" + prefixLines(a.Content, "+ ")
	}
	return ""
}

// IsMutating reports whether a tool mutates state (blocked in ask mode).
func IsMutating(name string) bool {
	t, ok := byName[name]
	return ok && t.Mutating
}

// IsReadOnly reports whether a tool is side-effect-free.
func IsReadOnly(name string) bool {
	t, ok := byName[name]
	return ok && t.ReadOnly
}

// ReadOnlyTools returns the read-only subset of All, in order.
func ReadOnlyTools() []Tool {
	var out []Tool
	for _, t := range All {
		if t.ReadOnly {
			out = append(out, t)
		}
	}
	return out
}
