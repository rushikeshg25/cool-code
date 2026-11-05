package tools

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNewFilePreservesExisting(t *testing.T) {
	root := t.TempDir()
	p := filepath.Join(root, "existing")
	if err := os.WriteFile(p, []byte("user work"), 0600); err != nil {
		t.Fatal(err)
	}
	r := newFileTool.Execute(testCtx(root), args(t, map[string]any{"filePath": p, "content": "replacement"}))
	raw, err := os.ReadFile(p)
	if err != nil {
		t.Fatal(err)
	}
	if !r.Failed || string(raw) != "user work" {
		t.Fatal("new_file overwrote existing work")
	}
}
