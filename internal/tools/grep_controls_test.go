package tools

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGrepControls(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "a.txt")
	if err := os.WriteFile(path, []byte("A[B]\naXb\na[b]\n"), 0600); err != nil {
		t.Fatal(err)
	}
	ctx := testCtx(root)
	r := grepTool.Execute(ctx, args(t, map[string]any{"pattern": "a[b]", "literal": true, "ignoreCase": true, "limit": 1}))
	if r.Failed || !strings.Contains(r.LLMResult, "A[B]") || !strings.Contains(r.LLMResult, "Results truncated") {
		t.Fatal(r)
	}
	r = grepTool.Execute(ctx, args(t, map[string]any{"pattern": "a[b]", "literal": true}))
	if r.Failed || !strings.Contains(r.LLMResult, ":3: a[b]") || strings.Contains(r.LLMResult, "A[B]") {
		t.Fatal(r)
	}
	for _, limit := range []int{0, 201} {
		if !grepTool.Execute(ctx, args(t, map[string]any{"pattern": "a", "limit": limit})).Failed {
			t.Fatal("invalid limit accepted")
		}
	}
	c, cancel := context.WithCancel(context.Background())
	cancel()
	ctx.Ctx = c
	if !grepTool.Execute(ctx, args(t, map[string]any{"pattern": "a"})).Failed {
		t.Fatal("cancelled search succeeded")
	}
}
func TestGrepBoundsLongLine(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "a.txt")
	if err := os.WriteFile(path, []byte(strings.Repeat("界", 10000)), 0600); err != nil {
		t.Fatal(err)
	}
	r := grepTool.Execute(testCtx(root), args(t, map[string]any{"pattern": "界"}))
	if r.Failed || len(r.LLMResult) > 1500 || !strings.Contains(r.LLMResult, "line truncated") {
		t.Fatal("unbounded output")
	}
}
