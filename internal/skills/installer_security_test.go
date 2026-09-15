package skills

import (
	"os"
	"path/filepath"
	"testing"
)

func TestInstallRejectsDestinationSymlinks(t *testing.T) {
	for _, global := range []bool{false, true} {
		for _, part := range []string{".coolcode", ".coolcode/skills", ".coolcode/skills/demo"} {
			t.Run(map[bool]string{false: "project/", true: "global/"}[global]+part, func(t *testing.T) {
				root := t.TempDir()
				home := t.TempDir()
				t.Setenv("HOME", home)
				base := root
				if global {
					base = home
				}
				outside := t.TempDir()
				sentinel := filepath.Join(outside, "sentinel")
				if err := os.WriteFile(sentinel, []byte("unchanged"), 0600); err != nil {
					t.Fatal(err)
				}
				link := filepath.Join(base, part)
				if err := os.MkdirAll(filepath.Dir(link), 0755); err != nil {
					t.Fatal(err)
				}
				if err := os.Symlink(outside, link); err != nil {
					t.Fatal(err)
				}
				src := t.TempDir()
				if err := os.WriteFile(filepath.Join(src, "SKILL.md"), []byte("---\nname: demo\n---\nbody\n"), 0600); err != nil {
					t.Fatal(err)
				}
				res := Install(src, global, root)
				if res.Error == "" {
					t.Fatal("symlink accepted")
				}
				data, err := os.ReadFile(sentinel)
				if err != nil || string(data) != "unchanged" {
					t.Fatalf("target changed: %q %v", data, err)
				}
				entries, _ := os.ReadDir(outside)
				if len(entries) != 1 {
					t.Fatal("wrote into symlink target")
				}
			})
		}
	}
}

func TestInstallReplacementPreservesOldSkillOnCopyFailure(t *testing.T) {
	root := t.TempDir()
	base, err := os.OpenRoot(root)
	if err != nil {
		t.Fatal(err)
	}
	defer base.Close()
	if err := base.Mkdir("demo", 0755); err != nil {
		t.Fatal(err)
	}
	if err := base.WriteFile("demo/SKILL.md", []byte("old"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := installItemInto(base, installItem{src: filepath.Join(t.TempDir(), "missing"), name: "demo", fileOnly: true}); err == nil {
		t.Fatal("missing source accepted")
	}
	data, _ := base.ReadFile("demo/SKILL.md")
	if string(data) != "old" {
		t.Fatal("previous skill destroyed")
	}
	src := filepath.Join(t.TempDir(), "SKILL.md")
	if err := os.WriteFile(src, []byte("new"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := installItemInto(base, installItem{src: src, name: "demo", fileOnly: true}); err != nil {
		t.Fatal(err)
	}
	data, _ = base.ReadFile("demo/SKILL.md")
	if string(data) != "new" {
		t.Fatal("replacement not installed")
	}
	entries, _ := os.ReadDir(root)
	if len(entries) != 1 {
		t.Fatal("staging/backup entries left behind")
	}
}
