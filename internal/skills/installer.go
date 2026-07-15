package skills

import (
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

// InstallResult reports the outcome of an install.
type InstallResult struct {
	Installed []string
	Dest      string
	Error     string
}

var gitURLRe = regexp.MustCompile(`^(https?://|git@|ssh://|git://)`)

// IsGitURL reports whether source looks like a git remote.
func IsGitURL(source string) bool {
	return gitURLRe.MatchString(source) || strings.HasSuffix(source, ".git")
}

var sanitizeRe = regexp.MustCompile(`[^a-zA-Z0-9_-]+`)

func sanitizeName(name string) string {
	s := sanitizeRe.ReplaceAllString(name, "-")
	s = strings.Trim(s, "-")
	s = strings.ToLower(s)
	if s == "" {
		return "skill"
	}
	return s
}

// collectSkillDirs finds directories containing a SKILL.md: root, its immediate
// subdirectories, and a conventional skills/<name> layout.
func collectSkillDirs(root string) []string {
	seen := map[string]bool{}
	var out []string
	check := func(dir string) {
		if _, err := os.Stat(filepath.Join(dir, skillFile)); err == nil {
			if !seen[dir] {
				seen[dir] = true
				out = append(out, dir)
			}
		}
	}
	check(root)
	for _, sub := range []string{root, filepath.Join(root, "skills")} {
		entries, err := os.ReadDir(sub)
		if err != nil {
			continue
		}
		for _, e := range entries {
			if e.IsDir() {
				check(filepath.Join(sub, e.Name()))
			}
		}
	}
	return out
}

func nameForSkill(skillMdPath string) string {
	dirName := filepath.Base(filepath.Dir(skillMdPath))
	skill, err := ParseFile(skillMdPath, dirName)
	if err != nil {
		return dirName
	}
	return skill.Name
}

type installItem struct {
	src      string
	name     string
	fileOnly bool
}

// Install installs one or more skills from a local path or git URL into the
// project (.coolcode/skills) or global (~/.coolcode/skills when global) dir.
func Install(source string, global bool, rootDir string) InstallResult {
	home, _ := os.UserHomeDir()
	destBase := filepath.Join(rootDir, ".coolcode", "skills")
	if global {
		destBase = filepath.Join(home, ".coolcode", "skills")
	}

	var tempDir string
	defer func() {
		if tempDir != "" {
			_ = os.RemoveAll(tempDir)
		}
	}()

	var items []installItem

	if IsGitURL(source) {
		var err error
		tempDir, err = os.MkdirTemp("", "coolcode-skill-")
		if err != nil {
			return InstallResult{Dest: destBase, Error: err.Error()}
		}
		cmd := exec.Command("git", "clone", "--depth", "1", source, tempDir)
		if out, err := cmd.CombinedOutput(); err != nil {
			return InstallResult{Dest: destBase, Error: "git clone failed: " + strings.TrimSpace(string(out))}
		}
		for _, dir := range collectSkillDirs(tempDir) {
			items = append(items, installItem{src: dir, name: nameForSkill(filepath.Join(dir, skillFile))})
		}
	} else {
		abs, _ := filepath.Abs(source)
		info, err := os.Stat(abs)
		if err != nil {
			return InstallResult{Dest: destBase, Error: "Source not found: " + abs}
		}
		if !info.IsDir() {
			if filepath.Base(abs) != skillFile {
				return InstallResult{Dest: destBase, Error: "Expected a SKILL.md file or a directory containing one."}
			}
			items = append(items, installItem{src: abs, name: nameForSkill(abs), fileOnly: true})
		} else {
			for _, dir := range collectSkillDirs(abs) {
				items = append(items, installItem{src: dir, name: nameForSkill(filepath.Join(dir, skillFile))})
			}
		}
	}

	if len(items) == 0 {
		return InstallResult{Dest: destBase, Error: "No SKILL.md found in the provided source."}
	}

	if err := os.MkdirAll(destBase, 0o755); err != nil {
		return InstallResult{Dest: destBase, Error: err.Error()}
	}
	var installed []string
	for _, item := range items {
		destDir := filepath.Join(destBase, sanitizeName(item.name))
		_ = os.RemoveAll(destDir)
		if item.fileOnly {
			if err := os.MkdirAll(destDir, 0o755); err != nil {
				return InstallResult{Dest: destBase, Error: err.Error()}
			}
			if err := copyFile(item.src, filepath.Join(destDir, skillFile)); err != nil {
				return InstallResult{Dest: destBase, Error: err.Error()}
			}
		} else {
			if err := copyDir(item.src, destDir); err != nil {
				return InstallResult{Dest: destBase, Error: err.Error()}
			}
		}
		installed = append(installed, item.name)
	}
	return InstallResult{Installed: installed, Dest: destBase}
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}
	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()
	_, err = io.Copy(out, in)
	return err
}

func copyDir(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		target := filepath.Join(dst, rel)
		if info.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		return copyFile(path, target)
	})
}
