package tools

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestReadPagesBounded(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "a.txt")
	if err := os.WriteFile(path, []byte(strings.Repeat("line\n", 3000)), 0600); err != nil {
		t.Fatal(err)
	}
	r := readFileTool.Execute(testCtx(root), args(t, map[string]any{"absolutePath": path}))
	if r.Failed || !strings.Contains(r.LLMResult, "startLine=2001") || len(r.LLMResult) > 11000 {
		t.Fatalf("read not paginated: length %d", len(r.LLMResult))
	}
	r = readFileTool.Execute(testCtx(root), args(t, map[string]any{"absolutePath": path, "startLine": 2001, "endLine": 3000}))
	if r.Failed || strings.Count(r.LLMResult, "line") != 1000 {
		t.Fatal("continuation failed")
	}
}
func TestReadStopsAtRequestedLine(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "a.txt")
	if err := os.WriteFile(path, []byte("ok\n"+strings.Repeat("x", 11*1024*1024)), 0600); err != nil {
		t.Fatal(err)
	}
	r := readFileTool.Execute(testCtx(root), args(t, map[string]any{"absolutePath": path, "startLine": 1, "endLine": 2}))
	if !r.Failed {
		t.Fatal("scanner failure returned successful partial read")
	}
}
