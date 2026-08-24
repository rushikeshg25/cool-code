package tools

import (
	"encoding/json"
	"path/filepath"
	"testing"

	"github.com/rushikeshg25/cool-code/internal/config"
)

func TestBlockedPath(t *testing.T) {
	cfg := config.Default()
	cases := map[string]bool{
		".env":         true,
		".env.local":   true,
		"id_rsa":       true,
		"server.key":   true,
		"cert.pem":     true,
		"main.go":      false,
		"package.json": false,
	}
	for name, want := range cases {
		got := BlockedPath(filepath.Join("/proj", name), cfg) != ""
		if got != want {
			t.Errorf("BlockedPath(%q) = %v, want %v", name, got, want)
		}
	}
}

func TestEnsureAbsoluteWithinRoot(t *testing.T) {
	root := "/home/user/proj"
	if EnsureAbsoluteWithinRoot("/home/user/proj/src/a.go", root) != "" {
		t.Error("expected in-root path to pass")
	}
	if EnsureAbsoluteWithinRoot("relative/path", root) == "" {
		t.Error("expected relative path to fail")
	}
	if EnsureAbsoluteWithinRoot("/home/user/other/a.go", root) == "" {
		t.Error("expected out-of-root path to fail")
	}
	if EnsureAbsoluteWithinRoot("/home/user/proj/../secret", root) == "" {
		t.Error("expected traversal escape to fail")
	}
}

func TestEnsureAbsoluteWithinRoots(t *testing.T) {
	roots := []string{"/home/user/proj", "/home/user/lib"}
	if EnsureAbsoluteWithinRoots("/home/user/proj/src/a.go", roots) != "" {
		t.Error("expected primary-root path to pass")
	}
	if EnsureAbsoluteWithinRoots("/home/user/lib/b.go", roots) != "" {
		t.Error("expected extra-dir path to pass")
	}
	if EnsureAbsoluteWithinRoots("/home/user/other/a.go", roots) == "" {
		t.Error("expected out-of-all-roots path to fail")
	}
	if EnsureAbsoluteWithinRoots("relative/path", roots) == "" {
		t.Error("expected relative path to fail")
	}
	if EnsureAbsoluteWithinRoots("/home/user/lib/../secret", roots) == "" {
		t.Error("expected traversal escape to fail")
	}
	if EnsureAbsoluteWithinRoots("/home/user/proj/a.go", roots[:1]) != "" {
		t.Error("expected single-root behavior with no extras")
	}
}

func TestDangerReason(t *testing.T) {
	rm, _ := json.Marshal(map[string]any{"command": "rm -rf /tmp/x"})
	if DangerReason("shell_command", rm) == "" {
		t.Error("rm -rf should be dangerous")
	}
	safe, _ := json.Marshal(map[string]any{"command": "ls -la"})
	if DangerReason("shell_command", safe) == "" {
		t.Error("all arbitrary shell commands should require confirmation")
	}
	write, _ := json.Marshal(map[string]any{"dryRun": false})
	if DangerReason("replace_in_files", write) == "" {
		t.Error("non-dry-run replace should be dangerous")
	}
	overwrite, _ := json.Marshal(map[string]any{"overwrite": true})
	if DangerReason("rename_file", overwrite) == "" {
		t.Error("overwrite rename should be dangerous")
	}
}

func TestEditPreviewAndMutating(t *testing.T) {
	args, _ := json.Marshal(map[string]any{"filePath": "/a.go", "oldString": "x", "newString": "y"})
	if EditPreview("edit_file", args) == "" {
		t.Error("edit_file should produce a preview")
	}
	if !IsMutating("edit_file") || IsMutating("read_file") {
		t.Error("mutating classification wrong")
	}
}

func TestToPascalCase(t *testing.T) {
	cases := map[string]string{
		"user-profile": "UserProfile",
		"my_module":    "MyModule",
		"foo bar baz":  "FooBarBaz",
	}
	for in, want := range cases {
		if got := toPascalCase(in); got != want {
			t.Errorf("toPascalCase(%q) = %q, want %q", in, got, want)
		}
	}
}
