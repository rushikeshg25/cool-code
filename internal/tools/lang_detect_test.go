package tools

import (
	"os"
	"path/filepath"
	"slices"
	"testing"
)

func TestDetectProjectAndDefaults(t *testing.T) {
	cases := []struct {
		marker  string
		kind    projectKind
		testCmd string
		fmtFile string
		fmtCmd  []string
	}{
		{"go.mod", projGo, "go test ./...", "main.go", []string{"gofmt", "-w", "./main.go"}},
		{"Cargo.toml", projRust, "cargo test", "main.rs", []string{"rustfmt", "./main.rs"}},
		{"pyproject.toml", projPython, "pytest", "app.py", []string{"ruff", "format", "./app.py"}},
	}
	for _, c := range cases {
		dir := t.TempDir()
		if err := os.WriteFile(filepath.Join(dir, c.marker), []byte("x"), 0o644); err != nil {
			t.Fatal(err)
		}
		if got := detectProject(dir); got != c.kind {
			t.Errorf("%s: detectProject = %d, want %d", c.marker, got, c.kind)
		}
		if got := defaultTestCommand(dir); got != c.testCmd {
			t.Errorf("%s: defaultTestCommand = %q, want %q", c.marker, got, c.testCmd)
		}
		if got := formatFileArgv(c.fmtFile); !slices.Equal(got, c.fmtCmd) {
			t.Errorf("%s: formatFileArgv = %q, want %q", c.marker, got, c.fmtCmd)
		}
	}

	// Unknown project yields no default test command.
	if got := defaultTestCommand(t.TempDir()); got != "" {
		t.Errorf("empty project: defaultTestCommand = %q, want \"\"", got)
	}
	// Unknown extension falls back to prettier.
	want := []string{"npx", "prettier", "--write", "./index.html"}
	if got := formatFileArgv("index.html"); !slices.Equal(got, want) {
		t.Errorf("html: formatFileArgv = %q", got)
	}
}
