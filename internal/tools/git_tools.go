package tools

import (
	"encoding/json"
	"strings"

	"github.com/rushikeshg25/cool-code/internal/types"
)

var gitStatusTool = Tool{
	Name:        "git_status",
	Description: "Shows a git status summary.",
	Schema:      obj(map[string]any{}),
	Execute: func(ctx Context, _ json.RawMessage) types.ToolResult {
		res := execCommand(ctx.Context(), "git status --short -b", ctx.RootDir, 0)
		display := "Git status"
		if !res.success {
			display = "Git status failed"
		}
		return types.ToolResult{Display: display, LLMResult: res.combined()}
	},
}

var gitDiffTool = Tool{
	Name:        "git_diff",
	Description: "Shows a git diff. Optionally specify a file path and/or the staged diff.",
	Schema: obj(map[string]any{
		"filePath": strProp("Absolute path to a file to diff (optional)."),
		"staged":   boolProp("If true, show the staged diff."),
	}),
	Execute: func(ctx Context, args json.RawMessage) types.ToolResult {
		var a struct {
			FilePath string `json:"filePath"`
			Staged   bool   `json:"staged"`
		}
		_ = json.Unmarshal(args, &a)
		fileArg := ""
		if a.FilePath != "" {
			if v := EnsureAbsoluteWithinRoot(a.FilePath, ctx.RootDir); v != "" {
				return fail("Invalid path", v)
			}
			fileArg = " -- '" + shellEscapeSingleQuotes(toRelative(a.FilePath, ctx.RootDir)) + "'"
		}
		command := "git diff"
		if a.Staged {
			command += " --staged"
		}
		command += fileArg
		res := execCommand(ctx.Context(), command, ctx.RootDir, 0)
		display := "Git diff"
		if !res.success {
			display = "Git diff failed"
		}
		return types.ToolResult{Display: display, LLMResult: res.combined()}
	},
}

var gitCommitTool = Tool{
	Name:        "git_commit",
	Description: "Stages files and creates a git commit.",
	Mutating:    true,
	Schema: obj(map[string]any{
		"message": strProp("Commit message."),
		"all":     boolProp("If true, stage all changes."),
		"files":   arrProp("Array of absolute file paths to stage."),
	}, "message"),
	Execute: func(ctx Context, args json.RawMessage) types.ToolResult {
		var a struct {
			Message string   `json:"message"`
			All     bool     `json:"all"`
			Files   []string `json:"files"`
		}
		if err := json.Unmarshal(args, &a); err != nil {
			return fail("Invalid arguments", err.Error())
		}
		if strings.TrimSpace(a.Message) == "" {
			return fail("Invalid arguments", "Commit message is required.")
		}
		if a.All {
			if res := execCommand(ctx.Context(), "git add -A", ctx.RootDir, 0); !res.success {
				return types.ToolResult{Display: "Git add failed", LLMResult: res.combined()}
			}
		} else if len(a.Files) > 0 {
			var rels []string
			for _, f := range a.Files {
				if v := EnsureAbsoluteWithinRoot(f, ctx.RootDir); v != "" {
					return fail("Invalid path", v)
				}
				rels = append(rels, "'"+shellEscapeSingleQuotes(toRelative(f, ctx.RootDir))+"'")
			}
			if res := execCommand(ctx.Context(), "git add -- "+strings.Join(rels, " "), ctx.RootDir, 0); !res.success {
				return types.ToolResult{Display: "Git add failed", LLMResult: res.combined()}
			}
		}
		res := execCommand(ctx.Context(), "git commit -m '"+shellEscapeSingleQuotes(a.Message)+"'", ctx.RootDir, 0)
		display := "Commit created"
		if !res.success {
			display = "Commit failed"
		}
		return types.ToolResult{Display: display, LLMResult: res.combined()}
	},
}
