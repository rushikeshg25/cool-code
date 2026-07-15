// Package skills discovers and installs model-invoked instruction modules
// stored as SKILL.md files (compatible with Claude Code skills).
package skills

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// Skill is a discovered instruction module.
type Skill struct {
	Name        string
	Description string
	Body        string
	Path        string
}

const skillFile = "SKILL.md"

func skillDirs(rootDir string) []string {
	home, _ := os.UserHomeDir()
	return []string{
		filepath.Join(rootDir, ".coolcode", "skills"),
		filepath.Join(home, ".coolcode", "skills"),
	}
}

var (
	frontmatterRe = regexp.MustCompile(`(?s)^---\s*\n(.*?)\n---\s*\n?`)
	nameRe        = regexp.MustCompile(`(?m)^\s*name:\s*(.+?)\s*$`)
	descRe        = regexp.MustCompile(`(?m)^\s*description:\s*(.+?)\s*$`)
)

// ParseFile parses a SKILL.md, reading `name`/`description` frontmatter and
// falling back to the directory name and first non-empty body line.
func ParseFile(filePath, dirName string) (Skill, error) {
	raw, err := os.ReadFile(filePath)
	if err != nil {
		return Skill{}, err
	}
	content := string(raw)
	name := dirName
	description := ""
	body := strings.TrimSpace(content)

	if fm := frontmatterRe.FindStringSubmatch(content); fm != nil {
		if m := nameRe.FindStringSubmatch(fm[1]); m != nil {
			name = strings.TrimSpace(m[1])
		}
		if m := descRe.FindStringSubmatch(fm[1]); m != nil {
			description = strings.TrimSpace(m[1])
		}
		body = strings.TrimSpace(content[len(fm[0]):])
	}
	if description == "" {
		for _, line := range strings.Split(body, "\n") {
			if strings.TrimSpace(line) != "" {
				description = strings.TrimSpace(line)
				break
			}
		}
	}
	return Skill{Name: name, Description: description, Body: body, Path: filePath}, nil
}

// Discover finds skills under project and global skills directories. Project
// skills win on name collision.
func Discover(rootDir string) []Skill {
	seen := map[string]bool{}
	var out []Skill
	for _, base := range skillDirs(rootDir) {
		entries, err := os.ReadDir(base)
		if err != nil {
			continue
		}
		for _, entry := range entries {
			if !entry.IsDir() {
				continue
			}
			file := filepath.Join(base, entry.Name(), skillFile)
			if _, err := os.Stat(file); err != nil {
				continue
			}
			skill, err := ParseFile(file, entry.Name())
			if err != nil {
				continue
			}
			if !seen[skill.Name] {
				seen[skill.Name] = true
				out = append(out, skill)
			}
		}
	}
	return out
}

// Catalog builds a compact `- name: description` list for the system prompt, or
// "" when no skills exist.
func Catalog(rootDir string) string {
	skills := Discover(rootDir)
	if len(skills) == 0 {
		return ""
	}
	var b strings.Builder
	b.WriteString("--- Available Skills ---\n")
	b.WriteString("Use the 'use_skill' tool with a skill's name to load its full instructions when relevant.\n")
	for _, s := range skills {
		b.WriteString("- " + s.Name + ": " + s.Description + "\n")
	}
	return strings.TrimRight(b.String(), "\n")
}

// Body returns the full SKILL.md body for a named skill, or ("", false).
func Body(rootDir, name string) (string, bool) {
	for _, s := range Discover(rootDir) {
		if s.Name == name {
			return s.Body, true
		}
	}
	return "", false
}

// Names returns the discovered skill names.
func Names(rootDir string) []string {
	var names []string
	for _, s := range Discover(rootDir) {
		names = append(names, s.Name)
	}
	return names
}
