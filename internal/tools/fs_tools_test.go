package tools

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/rushikeshg25/cool-code/internal/config"
	"github.com/rushikeshg25/cool-code/internal/types"
)

func testCtx(root string) Context {
	return Context{RootDir: root, Config: config.Default(), GitIgnore: func(string) bool { return false }}
}

func args(t *testing.T, m map[string]any) json.RawMessage {
	t.Helper()
	raw, err := json.Marshal(m)
	if err != nil {
		t.Fatal(err)
	}
	return raw
}

func TestNewReadEditFile(t *testing.T) {
	root := t.TempDir()
	ctx := testCtx(root)
	file := filepath.Join(root, "src", "a.txt")

	res := newFileTool.Execute(ctx, args(t, map[string]any{"filePath": file, "content": "hello world\nsecond line\n"}))
	if !strings.Contains(res.LLMResult, "created successfully") {
		t.Fatalf("new_file: %s", res.LLMResult)
	}

	res = readFileTool.Execute(ctx, args(t, map[string]any{"absolutePath": file}))
	if !strings.Contains(res.LLMResult, "hello world") {
		t.Fatalf("read_file: %s", res.LLMResult)
	}

	// edit_file reports the replacement count, not the file, so an edit is not
	// also a full disclosure of the file it touched.
	res = editFileTool.Execute(ctx, args(t, map[string]any{"filePath": file, "oldString": "hello world", "newString": "goodbye"}))
	if !strings.Contains(res.LLMResult, "Replaced 1 occurrence(s) in src/a.txt") {
		t.Fatalf("edit_file: %s", res.LLMResult)
	}
	if strings.Contains(res.LLMResult, "second line") {
		t.Fatalf("edit_file echoed file contents: %s", res.LLMResult)
	}
	raw, _ := os.ReadFile(file)
	if !strings.Contains(string(raw), "goodbye") {
		t.Fatalf("file not edited: %s", raw)
	}
}

func TestNewFileOutsideRootRejected(t *testing.T) {
	root := t.TempDir()
	ctx := testCtx(root)
	res := newFileTool.Execute(ctx, args(t, map[string]any{"filePath": "/etc/evil.txt", "content": "x"}))
	if !strings.Contains(res.LLMResult, "within project root") {
		t.Fatalf("expected containment error, got: %s", res.LLMResult)
	}
}

func TestEditFileOutsideRootRejected(t *testing.T) {
	root := t.TempDir()
	ctx := testCtx(root)
	outside := filepath.Join(t.TempDir(), "outside.txt")
	_ = os.WriteFile(outside, []byte("original"), 0o644)
	res := editFileTool.Execute(ctx, args(t, map[string]any{"filePath": outside, "oldString": "original", "newString": "hacked"}))
	if !strings.Contains(res.LLMResult, "within project root") {
		t.Fatalf("expected containment error, got: %s", res.LLMResult)
	}
	raw, _ := os.ReadFile(outside)
	if string(raw) != "original" {
		t.Fatalf("file outside root was modified: %s", raw)
	}
}

func TestReadAndSearchOutsideRootRejected(t *testing.T) {
	root := t.TempDir()
	ctx := testCtx(root)
	outside := filepath.Join(t.TempDir(), "outside.txt")
	_ = os.WriteFile(outside, []byte("company secret"), 0o644)

	for name, result := range map[string]types.ToolResult{
		"read": readFileTool.Execute(ctx, args(t, map[string]any{"absolutePath": outside})),
		"grep": grepTool.Execute(ctx, args(t, map[string]any{"pattern": "secret", "path": outside})),
		"find": findSymbolTool.Execute(ctx, args(t, map[string]any{"pattern": "secret", "path": outside})),
	} {
		if !strings.Contains(result.LLMResult, "within project root") {
			t.Errorf("%s did not enforce root jail: %s", name, result.LLMResult)
		}
	}
}

func TestSymlinkEscapeRejectedForReadAndWrite(t *testing.T) {
	root := t.TempDir()
	ctx := testCtx(root)
	outside := filepath.Join(t.TempDir(), "outside.txt")
	_ = os.WriteFile(outside, []byte("original"), 0o644)
	link := filepath.Join(root, "link.txt")
	if err := os.Symlink(outside, link); err != nil {
		t.Skipf("symlinks unavailable: %v", err)
	}

	read := readFileTool.Execute(ctx, args(t, map[string]any{"absolutePath": link}))
	if !strings.Contains(read.LLMResult, "within project root") {
		t.Fatalf("symlink read escaped root: %s", read.LLMResult)
	}
	editFileTool.Execute(ctx, args(t, map[string]any{"filePath": link, "oldString": "original", "newString": "changed"}))
	raw, _ := os.ReadFile(outside)
	if string(raw) != "original" {
		t.Fatalf("symlink edit escaped root: %s", raw)
	}
}

func TestReadBlockedByGuardrail(t *testing.T) {
	root := t.TempDir()
	ctx := testCtx(root)
	env := filepath.Join(root, ".env")
	_ = os.WriteFile(env, []byte("SECRET=1"), 0o644)
	res := readFileTool.Execute(ctx, args(t, map[string]any{"absolutePath": env}))
	if !strings.Contains(res.LLMResult, "blocked") {
		t.Fatalf("expected guardrail block, got: %s", res.LLMResult)
	}
}

func TestGlobAndGrep(t *testing.T) {
	root := t.TempDir()
	ctx := testCtx(root)
	_ = os.WriteFile(filepath.Join(root, "a.go"), []byte("package main\nfunc Hello() {}\n"), 0o644)
	_ = os.WriteFile(filepath.Join(root, "b.txt"), []byte("nothing here\n"), 0o644)

	res := globTool.Execute(ctx, args(t, map[string]any{"pattern": "*.go"}))
	if !strings.Contains(res.LLMResult, "a.go") {
		t.Fatalf("glob: %s", res.LLMResult)
	}

	res = grepTool.Execute(ctx, args(t, map[string]any{"pattern": "func Hello"}))
	if !strings.Contains(res.LLMResult, "a.go:2") {
		t.Fatalf("grep: %s", res.LLMResult)
	}
}

func TestReplaceInFilesDryRunDefault(t *testing.T) {
	root := t.TempDir()
	ctx := testCtx(root)
	file := filepath.Join(root, "x.txt")
	_ = os.WriteFile(file, []byte("foo foo foo"), 0o644)

	// Default dryRun: file must be unchanged.
	res := replaceInFilesTool.Execute(ctx, args(t, map[string]any{"pattern": "foo", "replacement": "bar"}))
	if !strings.Contains(res.Display, "Dry run") {
		t.Fatalf("expected dry run, got: %s", res.Display)
	}
	raw, _ := os.ReadFile(file)
	if string(raw) != "foo foo foo" {
		t.Fatalf("dry run mutated file: %s", raw)
	}

	// Explicit write.
	replaceInFilesTool.Execute(ctx, args(t, map[string]any{"pattern": "foo", "replacement": "bar", "dryRun": false}))
	raw, _ = os.ReadFile(file)
	if string(raw) != "bar bar bar" {
		t.Fatalf("write replace failed: %s", raw)
	}
}
