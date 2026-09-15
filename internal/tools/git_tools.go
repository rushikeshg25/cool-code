package tools

import (
	"encoding/json"
	"path/filepath"
	"strings"

	"github.com/rushikeshg25/cool-code/internal/security"
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
		return types.ToolResult{Display: display, LLMResult: res.combined(), Failed: !res.success}
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
		gitArgs := []string{"diff", "--no-renames", "--no-ext-diff", "--no-textconv"}
		if a.Staged {
			gitArgs = append(gitArgs, "--staged")
		}
		selection := "."
		if a.FilePath != "" {
			resolved, reason := ResolveReadPath(a.FilePath, ctx)
			if reason != "" {
				return fail("Invalid path", reason)
			}
			selection = ":(literal)" + filepath.ToSlash(toRelative(resolved, ctx.RootDir))
		}
		// Git wildmatch and our doublestar policy have different grammars
		// (notably braces). Enumerate names only, then apply the actual policy
		// before requesting any content. Disable renames so excluded source
		// content cannot be included via a permitted destination's rename diff.
		listArgs := append(append([]string{}, gitArgs...), "--name-only", "-z", "--", selection)
		listed := runCommandRaw(ctx.Context(), ctx.RootDir, 0, "git", listArgs...)
		if !listed.success {
			return fail("Git diff failed", security.Redact(listed.combined()))
		}
		var permitted []string
		for _, name := range strings.Split(listed.stdout, "\x00") {
			if name == "" {
				continue
			}
			if _, reason := ResolveReadPath(filepath.Join(ctx.RootDir, name), ctx); reason == "" {
				permitted = append(permitted, ":(literal)"+name)
			}
		}
		if len(permitted) == 0 {
			return types.ToolResult{Display: "Git diff", LLMResult: "No permitted changes."}
		}
		// Bound argv size while preserving all permitted file changes.
		var output strings.Builder
		for start := 0; start < len(permitted); start += 128 {
			end := min(start+128, len(permitted))
			argv := append(append([]string{}, gitArgs...), "--")
			argv = append(argv, permitted[start:end]...)
			result := execArgv(ctx.Context(), ctx.RootDir, 0, "git", argv...)
			if !result.success {
				return fail("Git diff failed", result.combined())
			}
			output.WriteString(result.combined())
		}
		return types.ToolResult{Display: "Git diff", LLMResult: output.String()}
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
				return types.ToolResult{Display: "Git add failed", LLMResult: res.combined(), Failed: true}
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
				return types.ToolResult{Display: "Git add failed", LLMResult: res.combined(), Failed: true}
			}
		}
		res := execArgv(ctx.Context(), ctx.RootDir, 0, "git", "commit", "-m", a.Message)
		display := "Commit created"
		if !res.success {
			display = "Commit failed"
		}
		return types.ToolResult{Display: display, LLMResult: res.combined(), Failed: !res.success}
	},
}
