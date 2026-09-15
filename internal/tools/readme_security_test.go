package tools

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestReadmeRespectsWritePolicy(t *testing.T) {
	for _, target := range []string{".git/config", ".coolcode/settings.json", ".env", "outside", "notes.md"} {
		t.Run(target, func(t *testing.T) {
			ctx, root := guardCtx(t)
			dest := filepath.Join(root, target)
			if target == "outside" {
				dest = filepath.Join(t.TempDir(), "sentinel")
			}
			if err := os.MkdirAll(filepath.Dir(dest), 0755); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(dest, []byte("original\n"), 0600); err != nil {
				t.Fatal(err)
			}
			if err := os.Symlink(dest, filepath.Join(root, "README.md")); err != nil {
				t.Fatal(err)
			}
			res := generateReadmeSectionTool.Execute(ctx, args(t, map[string]any{"title": "Usage", "content": "example"}))
			data, _ := os.ReadFile(dest)
			if target == "notes.md" {
				if res.Failed || !strings.Contains(string(data), "## Usage") {
					t.Fatalf("safe README link failed: %+v", res)
				}
			} else if !res.Failed || string(data) != "original\n" {
				t.Fatalf("protected target modified: %+v %q", res, data)
			}
		})
	}
}

func TestReadmeCreateAndAppend(t *testing.T) {
	ctx, root := guardCtx(t)
	for _, title := range []string{"First", "Second"} {
		if res := generateReadmeSectionTool.Execute(ctx, args(t, map[string]any{"title": title, "content": "example"})); res.Failed {
			t.Fatal(res.LLMResult)
		}
	}
	data, _ := os.ReadFile(filepath.Join(root, "README.md"))
	if !strings.Contains(string(data), "## First") || !strings.Contains(string(data), "## Second") {
		t.Fatal(string(data))
	}
}
