package tui

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func lastSystemEntry(m *model) string {
	for i := len(m.history) - 1; i >= 0; i-- {
		if m.history[i].kind == entrySystem {
			return m.history[i].raw
		}
	}
	return ""
}

func TestAddDirCommand(t *testing.T) {
	m := newTestModel(t)

	// No arg, nothing added yet.
	m = typeLine(t, m, "/add-dir")
	if !strings.Contains(lastSystemEntry(m), "No additional directories") {
		t.Fatalf("unexpected message: %q", lastSystemEntry(m))
	}

	// Add a valid directory.
	extra := t.TempDir()
	if err := os.WriteFile(filepath.Join(extra, "note.md"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	m = typeLine(t, m, "/add-dir "+extra)
	if !strings.Contains(lastSystemEntry(m), "Added directory: "+extra) {
		t.Fatalf("unexpected message: %q", lastSystemEntry(m))
	}
	if dirs := m.proc.ExtraDirs(); len(dirs) != 1 || dirs[0] != extra {
		t.Fatalf("ExtraDirs = %v", dirs)
	}

	// No arg now lists it.
	m = typeLine(t, m, "/add-dir")
	if !strings.Contains(lastSystemEntry(m), extra) {
		t.Fatalf("listing missing dir: %q", lastSystemEntry(m))
	}

	// Invalid path errors.
	m = typeLine(t, m, "/add-dir /nonexistent-path-xyz")
	if !strings.Contains(lastSystemEntry(m), "add-dir failed") {
		t.Fatalf("unexpected message: %q", lastSystemEntry(m))
	}

	// Extra-dir files are @-completable as absolute paths.
	found := false
	for _, f := range m.projectFiles() {
		if f == filepath.Join(extra, "note.md") {
			found = true
		}
	}
	if !found {
		t.Errorf("extra-dir file missing from completion list: %v", m.projectFiles())
	}
}
