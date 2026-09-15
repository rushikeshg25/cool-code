package tools

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRenameRejectsDirectoryLaundering(t *testing.T) {
	for _, rel := range []string{".aws/credentials", "parent/.git/config", "private/nested/secret.txt", "ordinary/file.txt"} {
		t.Run(rel, func(t *testing.T) {
			ctx, root := guardCtx(t)
			ctx.Config.Guardrails.BlockReadPatterns = append(ctx.Config.Guardrails.BlockReadPatterns, "private/**")
			file := filepath.Join(root, rel)
			if err := os.MkdirAll(filepath.Dir(file), 0755); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(file, []byte("unchanged"), 0600); err != nil {
				t.Fatal(err)
			}
			from := filepath.Join(root, filepath.ToSlash(rel[:len(rel)-len(filepath.Base(rel))-1]))
			// Move the top-level directory, including any protected descendants.
			for filepath.Dir(from) != root {
				from = filepath.Dir(from)
			}
			to := filepath.Join(root, "public")
			res := renameFileTool.Execute(ctx, args(t, map[string]any{"fromPath": from, "toPath": to}))
			if !res.Failed {
				t.Fatalf("directory rename accepted: %+v", res)
			}
			data, err := os.ReadFile(file)
			if err != nil || string(data) != "unchanged" {
				t.Fatalf("source changed: %q %v", data, err)
			}
			if _, err := os.Stat(to); !os.IsNotExist(err) {
				t.Fatal("destination created")
			}
		})
	}
}

func TestRenameRegularFileStillWorks(t *testing.T) {
	ctx, root := guardCtx(t)
	from := filepath.Join(root, "old.txt")
	to := filepath.Join(root, "nested", "new.txt")
	if err := os.WriteFile(from, []byte("ordinary"), 0600); err != nil {
		t.Fatal(err)
	}
	res := renameFileTool.Execute(ctx, args(t, map[string]any{"fromPath": from, "toPath": to}))
	if res.Failed {
		t.Fatal(res.LLMResult)
	}
	data, err := os.ReadFile(to)
	if err != nil || string(data) != "ordinary" {
		t.Fatalf("%q %v", data, err)
	}
	if _, err := os.Stat(from); !os.IsNotExist(err) {
		t.Fatal("source still exists")
	}
}
