package tools

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestValidatePatchReadOnly(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "a.txt")
	if err := os.WriteFile(path, []byte("old\n"), 0600); err != nil {
		t.Fatal(err)
	}
	patch := "*** Begin Patch\n*** Update File: a.txt\n@@\n-old\n+new\n*** Add File: nested/new.txt\n+created\n*** End Patch"
	r := RunReadOnly(testCtx(root), "validate_patch", args(t, map[string]any{"patch": patch}))
	if r.Failed || !strings.Contains(r.LLMResult, "No changes made") {
		t.Fatal(r)
	}
	raw, err := os.ReadFile(path)
	if err != nil || string(raw) != "old\n" {
		t.Fatal("validation changed source")
	}
	if _, err := os.Stat(filepath.Join(root, "nested")); !os.IsNotExist(err) {
		t.Fatal("validation created directory")
	}
	r = RunReadOnly(testCtx(root), "validate_patch", args(t, map[string]any{"patch": strings.Replace(patch, "-old", "-missing", 1)}))
	if !r.Failed {
		t.Fatal("invalid hunk accepted")
	}
}
func TestValidatePatchHonorsReadGuards(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "private.txt")
	if err := os.WriteFile(path, []byte("secret\n"), 0600); err != nil {
		t.Fatal(err)
	}
	ctx := testCtx(root)
	ctx.Config.Guardrails.BlockReadPatterns = append(ctx.Config.Guardrails.BlockReadPatterns, "private.txt")
	patch := "*** Begin Patch\n*** Delete File: private.txt\n*** End Patch"
	if !RunReadOnly(ctx, "validate_patch", args(t, map[string]any{"patch": patch})).Failed {
		t.Fatal("read guard bypass")
	}
	outside := "*** Begin Patch\n*** Add File: ../outside\n+x\n*** End Patch"
	if !RunReadOnly(ctx, "validate_patch", args(t, map[string]any{"patch": outside})).Failed {
		t.Fatal("path escape accepted")
	}
}
