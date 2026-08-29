package tools

import (
	"encoding/json"
	"path/filepath"
	"strings"

	"github.com/rushikeshg25/cool-code/internal/types"
)

var shellCommandTool = Tool{
	Name:        "shell_command",
	Description: "Executes a given shell command via `bash -c`. Returns stdout, stderr, and exit code.",
	Mutating:    true,
	Schema: obj(map[string]any{
		"command":     strProp("Exact bash command to execute."),
		"description": strProp("Brief description of the command for the user."),
		"directory":   strProp("Optional directory to run in, relative to the project root."),
		"timeout":     numProp("Optional timeout in seconds (default 30, maximum 1800). Raise it for builds, installs, and long test runs."),
	}, "command"),
	Execute: func(ctx Context, args json.RawMessage) types.ToolResult {
		var a struct {
			Command     string `json:"command"`
			Description string `json:"description"`
			Directory   string `json:"directory"`
			Timeout     int    `json:"timeout"`
		}
		if err := json.Unmarshal(args, &a); err != nil {
			return fail("Invalid arguments", err.Error())
		}
		if strings.TrimSpace(a.Command) == "" {
			return fail("Invalid command", "command must not be empty.")
		}
		dir := ctx.RootDir
		if a.Directory != "" {
			dir = a.Directory
			if !filepath.IsAbs(dir) {
				dir = filepath.Join(ctx.RootDir, dir)
			}
			resolved, v := ResolveWithinRoots(dir, ctx.Roots())
			if v != "" {
				return fail("Invalid directory", v)
			}
			dir = resolved
		}
		res := execCommand(ctx.Context(), a.Command, dir, commandTimeout(a.Timeout))
		llm := res.stdout
		if res.stderr != "" {
			llm += "\nSTDERR:\n" + res.stderr
		}
		display := "Command executed successfully (exit code: " + itoa(res.exitCode) + ")"
		if !res.success {
			display = "Command failed (exit code: " + itoa(res.exitCode) + ")"
			if res.errMsg != "" {
				display += ": " + res.errMsg
			}
		}
		return types.ToolResult{Display: display, LLMResult: llm, Failed: !res.success}
	},
}

var runTestsTool = Tool{
	Name:        "run_tests",
	Description: "Runs project tests. Auto-detects Go, Rust, Python, or Node when no command is provided.",
	Mutating:    true,
	Schema: obj(map[string]any{
		"command": strProp("Optional command to run tests."),
		"timeout": numProp("Optional timeout in seconds (default 30, maximum 1800)."),
	}),
	Execute: func(ctx Context, args json.RawMessage) types.ToolResult {
		var a struct {
			Command string `json:"command"`
			Timeout int    `json:"timeout"`
		}
		_ = json.Unmarshal(args, &a)
		command := a.Command
		if command == "" {
			command = defaultTestCommand(ctx.RootDir)
		}
		if command == "" {
			return fail("No test command found", "No test command found. Provide a command explicitly.")
		}
		res := execCommand(ctx.Context(), command, ctx.RootDir, commandTimeout(a.Timeout))
		display := "Tests completed successfully"
		if !res.success {
			display = "Tests failed"
		}
		return types.ToolResult{Display: display, LLMResult: res.combined(), Failed: !res.success}
	},
}

var lintFixTool = Tool{
	Name:        "lint_fix",
	Description: "Runs lint/format with auto-fix. Auto-detects Go, Rust, Python, or Node when no command is provided.",
	Mutating:    true,
	Schema: obj(map[string]any{
		"command": strProp("Optional command to run lint/format."),
		"timeout": numProp("Optional timeout in seconds (default 30, maximum 1800)."),
	}),
	Execute: func(ctx Context, args json.RawMessage) types.ToolResult {
		var a struct {
			Command string `json:"command"`
			Timeout int    `json:"timeout"`
		}
		_ = json.Unmarshal(args, &a)
		command := a.Command
		if command == "" {
			command = defaultLintFixCommand(ctx.RootDir)
		}
		if command == "" {
			return fail("No lint/format command found", "No lint/format command found. Provide a command explicitly.")
		}
		res := execCommand(ctx.Context(), command, ctx.RootDir, commandTimeout(a.Timeout))
		display := "Lint/format completed successfully"
		if !res.success {
			display = "Lint/format failed"
		}
		return types.ToolResult{Display: display, LLMResult: res.combined(), Failed: !res.success}
	},
}

var formatFileTool = Tool{
	Name:        "format_file",
	Description: "Formats a file with the right formatter for its type (gofmt, rustfmt, ruff, or prettier).",
	Mutating:    true,
	Schema: obj(map[string]any{
		"absolutePath": strProp("Absolute path to the file to format."),
	}, "absolutePath"),
	Execute: func(ctx Context, args json.RawMessage) types.ToolResult {
		var a struct {
			AbsolutePath string `json:"absolutePath"`
		}
		if err := json.Unmarshal(args, &a); err != nil {
			return fail("Invalid arguments", err.Error())
		}
		resolved, v := ResolveWritePath(a.AbsolutePath, ctx)
		if v != "" {
			return fail("Invalid path", v)
		}
		rel := toRelative(resolved, ctx.RootDir)
		argv := formatFileArgv(rel)
		res := execArgv(ctx.Context(), ctx.RootDir, 0, argv[0], argv[1:]...)
		display := "File formatted"
		if !res.success {
			display = "Formatting failed"
		}
		return types.ToolResult{Display: display, LLMResult: res.combined(), Failed: !res.success}
	},
}

var addScriptTool = Tool{
	Name:        "add_script",
	Description: "Adds a script to package.json.",
	Mutating:    true,
	Schema: obj(map[string]any{
		"name":      strProp("Script name."),
		"command":   strProp("Script command."),
		"overwrite": boolProp("If true, overwrite an existing script."),
	}, "name", "command"),
	Execute: func(ctx Context, args json.RawMessage) types.ToolResult {
		var a struct {
			Name      string `json:"name"`
			Command   string `json:"command"`
			Overwrite bool   `json:"overwrite"`
		}
		if err := json.Unmarshal(args, &a); err != nil {
			return fail("Invalid arguments", err.Error())
		}
		if a.Name == "" || a.Command == "" {
			return fail("Invalid arguments", "name and command are required.")
		}
		packagePath := filepath.Join(ctx.RootDir, "package.json")
		if reason := EnsureAbsoluteWithinRoots(packagePath, ctx.Roots()); reason != "" {
			return fail("package.json blocked", reason)
		}
		pkg, ok := readPackageJSON(ctx.RootDir)
		if !ok {
			return fail("package.json not found", "package.json not found or invalid.")
		}
		scripts, ok := pkg["scripts"].(map[string]any)
		if !ok {
			scripts = map[string]any{}
			pkg["scripts"] = scripts
		}
		if _, exists := scripts[a.Name]; exists && !a.Overwrite {
			return fail("Script already exists", "Script \""+a.Name+"\" already exists.")
		}
		scripts[a.Name] = a.Command
		if err := writePackageJSON(ctx.RootDir, pkg); err != nil {
			return fail("package.json update failed", err.Error())
		}
		return types.ToolResult{Display: "Script added", LLMResult: "Added script \"" + a.Name + "\"."}
	},
}
