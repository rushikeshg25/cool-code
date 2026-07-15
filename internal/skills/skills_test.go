package skills

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseFileFrontmatter(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "SKILL.md")
	content := "---\nname: greeter\ndescription: Greets people\n---\n\nAlways greet warmly.\n"
	if err := os.WriteFile(file, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	skill, err := ParseFile(file, "fallback")
	if err != nil {
		t.Fatal(err)
	}
	if skill.Name != "greeter" {
		t.Errorf("name = %q", skill.Name)
	}
	if skill.Description != "Greets people" {
		t.Errorf("description = %q", skill.Description)
	}
	if skill.Body != "Always greet warmly." {
		t.Errorf("body = %q", skill.Body)
	}
}

func TestParseFileNoFrontmatter(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "SKILL.md")
	if err := os.WriteFile(file, []byte("First line describes it.\nMore text.\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	skill, err := ParseFile(file, "mydir")
	if err != nil {
		t.Fatal(err)
	}
	if skill.Name != "mydir" {
		t.Errorf("name = %q, want fallback dir", skill.Name)
	}
	if skill.Description != "First line describes it." {
		t.Errorf("description = %q", skill.Description)
	}
}

func TestDiscoverAndInstall(t *testing.T) {
	project := t.TempDir()
	src := t.TempDir()
	if err := os.WriteFile(filepath.Join(src, "SKILL.md"), []byte("---\nname: demo\ndescription: A demo\n---\nBody\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	res := Install(src, false, project)
	if res.Error != "" {
		t.Fatalf("install error: %s", res.Error)
	}
	if len(res.Installed) != 1 || res.Installed[0] != "demo" {
		t.Fatalf("installed = %v", res.Installed)
	}
	// Discover also scans the global skills dir, so assert membership.
	found := false
	for _, n := range Names(project) {
		if n == "demo" {
			found = true
		}
	}
	if !found {
		t.Fatalf("discover did not include demo: %v", Names(project))
	}
	if body, ok := Body(project, "demo"); !ok || body != "Body" {
		t.Fatalf("body = %q ok=%v", body, ok)
	}
}
