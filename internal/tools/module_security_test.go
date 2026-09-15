package tools

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestModuleRejectsTraversalBeforeWriting(t *testing.T) {
	for _, name := range []string{"../../pwn", "../pwn", ".", "a/b", `a\b`, "x';bad"} {
		t.Run(name, func(t *testing.T) {
			ctx, root := guardCtx(t)
			result := newModuleTool.Execute(ctx, args(t, map[string]any{"moduleName": name, "baseDir": "src/nested"}))
			if !result.Failed {
				t.Fatalf("accepted %q: %+v", name, result)
			}
			entries, err := os.ReadDir(root)
			if err != nil {
				t.Fatal(err)
			}
			if len(entries) != 0 {
				t.Fatal("partial writes")
			}
		})
	}
}

func TestModuleValidatesEveryDestination(t *testing.T) {
	for _, dest := range []string{"module", "index", "rootIndex"} {
		t.Run(dest, func(t *testing.T) {
			ctx, root := guardCtx(t)
			dir := filepath.Join(root, "src", "demo")
			if err := os.MkdirAll(dir, 0755); err != nil {
				t.Fatal(err)
			}
			outside := filepath.Join(t.TempDir(), "sentinel")
			if err := os.WriteFile(outside, []byte("unchanged"), 0600); err != nil {
				t.Fatal(err)
			}
			link := map[string]string{"module": filepath.Join(dir, "demo.ts"), "index": filepath.Join(dir, "index.ts"), "rootIndex": filepath.Join(root, "src", "index.ts")}[dest]
			if err := os.Symlink(outside, link); err != nil {
				t.Fatal(err)
			}
			res := newModuleTool.Execute(ctx, args(t, map[string]any{"moduleName": "demo", "exportFromRootIndex": true}))
			if !res.Failed {
				t.Fatalf("allowed symlink: %+v", res)
			}
			data, _ := os.ReadFile(outside)
			if string(data) != "unchanged" {
				t.Fatal("target modified")
			}
			entries, _ := os.ReadDir(dir)
			want := 1
			if dest == "rootIndex" {
				want = 0
			}
			if len(entries) != want {
				t.Fatal("partial scaffold")
			}
		})
	}
}

func TestModuleScaffoldsValidName(t *testing.T) {
	ctx, root := guardCtx(t)
	res := newModuleTool.Execute(ctx, args(t, map[string]any{"moduleName": "user-profile", "exportFromRootIndex": true}))
	if res.Failed {
		t.Fatal(res.LLMResult)
	}
	for _, path := range []string{"src/user-profile/user-profile.ts", "src/user-profile/index.ts", "src/index.ts"} {
		if _, err := os.Stat(filepath.Join(root, path)); err != nil {
			t.Fatal(err)
		}
	}
}

func TestModuleRetainsFallbackExport(t *testing.T) {
	for _, name := range []string{"---", "___"} {
		ctx, root := guardCtx(t)
		res := newModuleTool.Execute(ctx, args(t, map[string]any{"moduleName": name}))
		if res.Failed {
			t.Fatal(res.LLMResult)
		}
		data, err := os.ReadFile(filepath.Join(root, "src", name, name+".ts"))
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(string(data), "export const NewModule") {
			t.Fatalf("invalid module: %s", data)
		}
	}
}
