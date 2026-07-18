package tools

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDetectProjectAndDefaults(t *testing.T) {
	cases := []struct {
		marker  string
		kind    projectKind
		testCmd string
		fmtFile string
		fmtCmd  string
	}{
		{"go.mod", projGo, "go test ./...", "main.go", "gofmt -w 'main.go'"},
		{"Cargo.toml", projRust, "cargo test", "main.rs", "rustfmt 'main.rs'"},
		{"pyproject.toml", projPython, "pytest", "app.py", "ruff format 'app.py'"},
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
		if got := formatFileCommand(c.fmtFile); got != c.fmtCmd {
			t.Errorf("%s: formatFileCommand = %q, want %q", c.marker, got, c.fmtCmd)
		}
	}

	// Unknown project yields no default test command.
	if got := defaultTestCommand(t.TempDir()); got != "" {
		t.Errorf("empty project: defaultTestCommand = %q, want \"\"", got)
	}
	// Unknown extension falls back to prettier.
	if got := formatFileCommand("index.html"); got != "npx prettier --write 'index.html'" {
		t.Errorf("html: formatFileCommand = %q", got)
	}
}
