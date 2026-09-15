package tools

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestFindSymbolChecksEveryFileThroughDirectoryAliases(t *testing.T) {
	if _, err := exec.LookPath("rg"); err != nil {
		t.Skip("ripgrep unavailable")
	}
	ctx, root := guardCtx(t)
	ctx.Config.Guardrails.BlockReadPatterns = append(ctx.Config.Guardrails.BlockReadPatterns, "guarded/**", "brace/{a,b}.txt")
	for _, file := range []string{"guarded/hidden.txt", "brace/a.txt", "KEY.PEM", "allowed/visible.txt"} {
		path := filepath.Join(root, file)
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			t.Fatal(err)
		}
		body := "SEARCH_SENTINEL blocked\n"
		if strings.Contains(file, "visible") {
			body = "SEARCH_SENTINEL allowed\n"
		}
		if err := os.WriteFile(path, []byte(body), 0600); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.Symlink(filepath.Join(root, "guarded"), filepath.Join(root, "alias")); err != nil {
		t.Fatal(err)
	}
	for _, path := range []string{root, filepath.Join(root, "guarded"), filepath.Join(root, "alias"), filepath.Join(root, "brace"), filepath.Join(root, "allowed")} {
		res := findSymbolTool.Execute(ctx, args(t, map[string]any{"path": path, "pattern": "SEARCH_SENTINEL"}))
		if strings.Contains(res.LLMResult, "SEARCH_SENTINEL blocked") {
			t.Fatalf("path %s: %+v", path, res)
		}
		if (path == root || path == filepath.Join(root, "allowed")) && !strings.Contains(res.LLMResult, "SEARCH_SENTINEL allowed") {
			t.Fatalf("ordinary file omitted: %+v", res)
		}
	}
	// The original alias spelling must remain subject to policy too.
	if err := os.Symlink(filepath.Join(root, "allowed"), filepath.Join(root, "blocked-alias")); err != nil {
		t.Fatal(err)
	}
	ctx.Config.Guardrails.BlockReadPatterns = append(ctx.Config.Guardrails.BlockReadPatterns, "blocked-alias/**")
	res := findSymbolTool.Execute(ctx, args(t, map[string]any{"path": filepath.Join(root, "blocked-alias"), "pattern": "SEARCH_SENTINEL"}))
	if strings.Contains(res.LLMResult, "SEARCH_SENTINEL") {
		t.Fatal("blocked alias disclosed content")
	}
}
