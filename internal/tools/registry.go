package tools

import (
	"encoding/json"
	"regexp"

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

var riskyShellPatterns = []*regexp.Regexp{
	regexp.MustCompile(`(?i)\brm\b.*(-rf|-fr)\b`),
	regexp.MustCompile(`(?i)\bsudo\b`),
	regexp.MustCompile(`(?i)\bmkfs\b`),
	regexp.MustCompile(`(?i)\bdd\b`),
	regexp.MustCompile(`(?i)\bshutdown\b|\breboot\b|\bpoweroff\b`),
	regexp.MustCompile(`(?i)\bkill\s+-9\b`),
	regexp.MustCompile(`(?i)\bgit\s+reset\s+--hard\b`),
	regexp.MustCompile(`(?i)\bgit\s+clean\b`),
	regexp.MustCompile(`(?i)\bgit\s+push\b.*--force\b`),
	regexp.MustCompile(`(?i)\bcurl\b.*\s*\|\s*(bash|sh)\b`),
	regexp.MustCompile(`(?i)\bwget\b.*\s*\|\s*(bash|sh)\b`),
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
		for _, re := range riskyShellPatterns {
			if re.MatchString(a.Command) {
				return "shell command"
			}
		}
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

// EditPreview returns a human-readable preview for edit-like tools, or "".
func EditPreview(name string, args json.RawMessage) string {
	truncate := func(s string) string {
		if len(s) > 400 {
			return s[:400] + "\n... (truncated)"
		}
		return s
	}
	switch name {
	case "edit_file":
		var a struct {
			FilePath  string `json:"filePath"`
			OldString string `json:"oldString"`
			NewString string `json:"newString"`
		}
		_ = json.Unmarshal(args, &a)
		return "File: " + a.FilePath + "\n--- old ---\n" + truncate(a.OldString) + "\n--- new ---\n" + truncate(a.NewString)
	case "new_file":
		var a struct {
			FilePath string `json:"filePath"`
			Content  string `json:"content"`
		}
		_ = json.Unmarshal(args, &a)
		return "File: " + a.FilePath + "\n--- content ---\n" + truncate(a.Content)
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
