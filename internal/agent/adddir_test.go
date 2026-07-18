package agent

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/rushikeshg25/cool-code/internal/types"
)

func TestAddDirValidation(t *testing.T) {
	root := t.TempDir()
	extra := t.TempDir()
	p := newTestProcessor(t, root, nil, types.ModeAgent)

	// Valid absolute dir.
	abs, err := p.AddDir(extra)
	if err != nil || abs != extra {
		t.Fatalf("AddDir(%q) = %q, %v", extra, abs, err)
	}
	if dirs := p.ExtraDirs(); len(dirs) != 1 || dirs[0] != extra {
		t.Fatalf("ExtraDirs = %v", dirs)
	}

	// Duplicate rejected.
	if _, err := p.AddDir(extra); err == nil {
		t.Error("expected duplicate to fail")
	}
	// Inside primary root rejected.
	sub := filepath.Join(root, "sub")
	if err := os.Mkdir(sub, 0o755); err != nil {
		t.Fatal(err)
	}
	if _, err := p.AddDir(sub); err == nil {
		t.Error("expected dir inside root to fail")
	}
	// Inside an existing extra dir rejected.
	nested := filepath.Join(extra, "nested")
	if err := os.Mkdir(nested, 0o755); err != nil {
		t.Fatal(err)
	}
	if _, err := p.AddDir(nested); err == nil {
		t.Error("expected dir inside extra dir to fail")
	}
	// Nonexistent rejected.
	if _, err := p.AddDir(filepath.Join(extra, "missing")); err == nil {
		t.Error("expected missing dir to fail")
	}
	// A file (not dir) rejected.
	file := filepath.Join(extra, "f.txt")
	if err := os.WriteFile(file, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := p.AddDir(file); err == nil {
		t.Error("expected file to fail")
	}
	// Empty arg rejected.
	if _, err := p.AddDir(""); err == nil {
		t.Error("expected empty arg to fail")
	}

	// Relative path resolves against the root.
	sibling := filepath.Join(filepath.Dir(root), filepath.Base(root)+"-sib")
	if err := os.Mkdir(sibling, 0o755); err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(sibling)
	abs, err = p.AddDir("../" + filepath.Base(sibling))
	if err != nil || abs != sibling {
		t.Fatalf("relative AddDir = %q, %v", abs, err)
	}

	// Tool context sees every added root.
	if got := len(p.toolCtx.Roots()); got != 3 {
		t.Errorf("toolCtx.Roots() has %d entries, want 3", got)
	}
}

func TestAddDirSystemPrompt(t *testing.T) {
	root := t.TempDir()
	extra := t.TempDir()
	if err := os.WriteFile(filepath.Join(extra, "lib.go"), []byte("package lib"), 0o644); err != nil {
		t.Fatal(err)
	}
	p := newTestProcessor(t, root, nil, types.ModeAgent)

	before := p.ctxMgr.buildSystem()
	if strings.Contains(before, "Additional Directories") {
		t.Fatal("section should be absent before AddDir")
	}
	if _, err := p.AddDir(extra); err != nil {
		t.Fatal(err)
	}
	after := p.ctxMgr.buildSystem()
	if !strings.Contains(after, "--- Additional Directories ---") || !strings.Contains(after, extra) {
		t.Error("system prompt missing additional-directories section")
	}
	if !strings.Contains(after, "lib.go") {
		t.Error("system prompt missing extra dir tree contents")
	}
}
