package tools

import (
	"encoding/json"
	"strings"

	"github.com/rushikeshg25/cool-code/internal/types"
)

var gitStatusTool = Tool{
	Name:        "git_status",
	Description: "Shows a git status summary.",
	ReadOnly:    true,
	Schema:      obj(map[string]any{}),
	Execute: func(ctx Context, _ json.RawMessage) types.ToolResult {
		res := execArgv(ctx.Context(), ctx.RootDir, 0, "git", "status", "--short", "-b")
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
	ReadOnly:    true,
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
		gitArgs := []string{"diff"}
		if a.Staged {
			gitArgs = append(gitArgs, "--staged")
		}
		if a.FilePath != "" {
			resolved, v := ResolveReadPath(a.FilePath, ctx)
			if v != "" {
				return fail("Invalid path", v)
			}
			gitArgs = append(gitArgs, "--", pathArg(toRelative(resolved, ctx.RootDir)))
		} else if specs := GitExcludePathspecs(ctx.Config); len(specs) > 0 {
			// A bare diff would print every tracked file, guardrailed ones
			// included, so exclude them by pathspec.
			gitArgs = append(gitArgs, "--", ".")
			gitArgs = append(gitArgs, specs...)
		}
		res := execArgv(ctx.Context(), ctx.RootDir, 0, "git", gitArgs...)
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
			if res := execArgv(ctx.Context(), ctx.RootDir, 0, "git", "add", "-A"); !res.success {
				return types.ToolResult{Display: "Git add failed", LLMResult: res.combined()}
			}
		} else if len(a.Files) > 0 {
			addArgs := []string{"add", "--"}
			for _, f := range a.Files {
				if v := EnsureAbsoluteWithinRoots(f, ctx.Roots()); v != "" {
					return fail("Invalid path", v)
				}
				addArgs = append(addArgs, pathArg(toRelative(f, ctx.RootDir)))
			}
			if res := execArgv(ctx.Context(), ctx.RootDir, 0, "git", addArgs...); !res.success {
				return types.ToolResult{Display: "Git add failed", LLMResult: res.combined()}
			}
		}
		res := execArgv(ctx.Context(), ctx.RootDir, 0, "git", "commit", "-m", a.Message)
		display := "Commit created"
		if !res.success {
			display = "Commit failed"
		}
		return types.ToolResult{Display: display, LLMResult: res.combined()}
	},
}
