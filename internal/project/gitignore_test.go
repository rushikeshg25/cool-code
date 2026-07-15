package project

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGitIgnoreChecker(t *testing.T) {
	root := t.TempDir()
	gitignore := "node_modules/\ndist/\n*.log\n/build\n"
	if err := os.WriteFile(filepath.Join(root, ".gitignore"), []byte(gitignore), 0o644); err != nil {
		t.Fatal(err)
	}
	check := NewGitIgnoreChecker(root)

	ignored := []string{"node_modules", "node_modules/react/index.js", "dist/app.js", "error.log", "build"}
	for _, p := range ignored {
		if !check(p) {
			t.Errorf("expected %q to be ignored", p)
		}
	}
	kept := []string{"src/index.ts", "main.go", "README.md"}
	for _, p := range kept {
		if check(p) {
			t.Errorf("expected %q to be kept", p)
		}
	}
}

func TestGitIgnoreMissingFile(t *testing.T) {
	check := NewGitIgnoreChecker(t.TempDir())
	if check("anything") {
		t.Error("no .gitignore should match nothing")
	}
}
