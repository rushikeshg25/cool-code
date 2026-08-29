package tools

import (
	"os"
	"path/filepath"
	"strings"
)

// projectKind identifies a project's toolchain by its marker files.
type projectKind int

const (
	projUnknown projectKind = iota
	projGo
	projRust
	projPython
	projNode
)

func rootFileExists(root, name string) bool {
	_, err := os.Stat(filepath.Join(root, name))
	return err == nil
}

// detectProject infers the toolchain from marker files at the project root.
func detectProject(root string) projectKind {
	switch {
	case rootFileExists(root, "go.mod"):
		return projGo
	case rootFileExists(root, "Cargo.toml"):
		return projRust
	case rootFileExists(root, "package.json"):
		return projNode
	case rootFileExists(root, "pyproject.toml"),
		rootFileExists(root, "requirements.txt"),
		rootFileExists(root, "setup.py"):
		return projPython
	}
	return projUnknown
}

// nodeScriptCommand returns "npm run <name>" when package.json defines it.
func nodeScriptCommand(root string, names ...string) string {
	pkg, ok := readPackageJSON(root)
	if !ok {
		return ""
	}
	scripts, ok := pkg["scripts"].(map[string]any)
	if !ok {
		return ""
	}
	for _, n := range names {
		if _, ok := scripts[n]; ok {
			return "npm run " + n
		}
	}
	return ""
}

// defaultTestCommand picks a test command for the detected toolchain.
func defaultTestCommand(root string) string {
	switch detectProject(root) {
	case projGo:
		return "go test ./..."
	case projRust:
		return "cargo test"
	case projPython:
		return "pytest"
	case projNode:
		return nodeScriptCommand(root, "test")
	}
	return ""
}

// defaultLintFixCommand picks a lint/format-with-fix command for the toolchain.
func defaultLintFixCommand(root string) string {
	switch detectProject(root) {
	case projGo:
		return "gofmt -w . && go vet ./..."
	case projRust:
		return "cargo fmt && cargo clippy --fix --allow-dirty"
	case projPython:
		return "ruff check --fix . && ruff format ."
	case projNode:
		if cmd := nodeScriptCommand(root, "lint:fix"); cmd != "" {
			return cmd
		}
		if cmd := nodeScriptCommand(root, "lint"); cmd != "" {
			return cmd + " -- --fix"
		}
		return nodeScriptCommand(root, "format")
	}
	return ""
}

// formatFileArgv picks a formatter for a single file by extension, returning
// the program and its arguments so no shell is involved.
func formatFileArgv(rel string) []string {
	q := pathArg(rel)
	switch strings.ToLower(filepath.Ext(rel)) {
	case ".go":
		return []string{"gofmt", "-w", q}
	case ".rs":
		return []string{"rustfmt", q}
	case ".py":
		return []string{"ruff", "format", q}
	default:
		return []string{"npx", "prettier", "--write", q}
	}
}
