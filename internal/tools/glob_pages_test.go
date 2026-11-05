package tools

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGlobPages(t *testing.T) {
	root := t.TempDir()
	for _, name := range []string{"c.txt", "a.txt", "b.txt", ".env"} {
		if err := os.WriteFile(filepath.Join(root, name), []byte("x"), 0600); err != nil {
			t.Fatal(err)
		}
	}
	ctx := testCtx(root)
	r := globTool.Execute(ctx, args(t, map[string]any{"pattern": "*", "limit": 2}))
	if r.Failed || !strings.Contains(r.LLMResult, "offset=2") || strings.Contains(r.LLMResult, "c.txt") || strings.Contains(r.LLMResult, ".env") {
		t.Fatal(r)
	}
	r = globTool.Execute(ctx, args(t, map[string]any{"pattern": "*", "limit": 2, "offset": 2}))
	if r.Failed || strings.TrimSpace(r.LLMResult) != filepath.Join(root, "c.txt") {
		t.Fatal(r)
	}
	for _, offset := range []int{3, int(^uint(0) >> 1)} {
		r = globTool.Execute(ctx, args(t, map[string]any{"pattern": "*", "offset": offset}))
		if r.Failed || r.LLMResult != "" {
			t.Fatal("out of range offset failed")
		}
	}
	for _, params := range []map[string]any{{"pattern": "*", "offset": -1}, {"pattern": "*", "limit": 0}, {"pattern": "*", "limit": 501}} {
		if !globTool.Execute(ctx, args(t, params)).Failed {
			t.Fatal("invalid pagination accepted")
		}
	}
}
