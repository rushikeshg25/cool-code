package tools

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestApplyPatchMultipleFiles(t *testing.T) {
	root := t.TempDir()
	a := filepath.Join(root, "a.go")
	gone := filepath.Join(root, "gone")
	if err := os.WriteFile(a, []byte("first\nold\nlast\n"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(gone, []byte("bye"), 0600); err != nil {
		t.Fatal(err)
	}
	patch := "*** Begin Patch\n*** Update File: a.go\n@@\n first\n-old\n+new\n last\n*** Add File: sub/new.txt\n+hello\n*** Delete File: gone\n*** End Patch"
	args, _ := json.Marshal(map[string]string{"patch": patch})
	r := Run(Context{RootDir: root}, "apply_patch", args)
	if r.Failed {
		t.Fatal(r.LLMResult)
	}
	data, _ := os.ReadFile(a)
	if string(data) != "first\nnew\nlast\n" {
		t.Fatal(string(data))
	}
	info, _ := os.Stat(a)
	if info.Mode().Perm() != 0755 {
		t.Fatal("lost mode")
	}
	if _, err := os.Stat(gone); !os.IsNotExist(err) {
		t.Fatal("not deleted")
	}
	data, _ = os.ReadFile(filepath.Join(root, "sub/new.txt"))
	if string(data) != "hello\n" {
		t.Fatal(string(data))
	}
	if !RunReadOnly(Context{RootDir: root}, "apply_patch", args).Failed {
		t.Fatal("read-only bypass")
	}
	if EditPreview("apply_patch", args) != patch {
		t.Fatal("preview differs")
	}
}
func TestApplyPatchRejectsBeforeWriting(t *testing.T) {
	for _, tail := range []string{
		"*** Update File: missing\n@@\n-old\n+new",
		"*** Add File: .git/config\n+evil",
		"*** Add File: ../outside\n+evil",
		"*** Add File: a\n+overwrite",
		"*** Update File: a\n@@\n-missing\n+new",
	} {
		t.Run(strings.Split(tail, "\n")[0], func(t *testing.T) {
			root := t.TempDir()
			_ = os.WriteFile(filepath.Join(root, "a"), []byte("old\n"), 0600)
			patch := "*** Begin Patch\n*** Add File: first\n+new\n" + tail + "\n*** End Patch"
			args, _ := json.Marshal(map[string]string{"patch": patch})
			r := Run(Context{RootDir: root}, "apply_patch", args)
			if !r.Failed {
				t.Fatal("accepted invalid patch")
			}
			if _, err := os.Stat(filepath.Join(root, "first")); !os.IsNotExist(err) {
				t.Fatal("partial write")
			}
			data, _ := os.ReadFile(filepath.Join(root, "a"))
			if string(data) != "old\n" {
				t.Fatal("changed original")
			}
		})
	}
}
func TestPatchRejectsSymlinkEscapeAndAmbiguity(t *testing.T) {
	root := t.TempDir()
	outside := t.TempDir()
	if err := os.Symlink(outside, filepath.Join(root, "link")); err != nil {
		t.Skip(err)
	}
	if _, err := preparePatch(Context{RootDir: root}, "*** Begin Patch\n*** Add File: link/leak\n+bad\n*** End Patch"); err == nil {
		t.Fatal("escaped root")
	}
	if _, err := applyPatchHunks([]byte("a\na\n"), []string{"@@", "-a", "+b"}); err == nil {
		t.Fatal("ambiguous hunk accepted")
	}
	got, err := applyPatchHunks([]byte("a\nb"), []string{"@@", " a", "-b", "+c"})
	if err != nil || string(got) != "a\nc" {
		t.Fatalf("newline changed: %q %v", got, err)
	}
}
