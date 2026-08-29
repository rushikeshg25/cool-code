package skills

import (
	"context"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"
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
		// Lstat, not Stat: a symlinked SKILL.md must not be accepted as one,
		// matching the check the skills loader applies at discovery time.
		info, err := os.Lstat(filepath.Join(dir, skillFile))
		if err != nil || !info.Mode().IsRegular() {
			return
		}
		if !seen[dir] {
			seen[dir] = true
			out = append(out, dir)
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
		ctx, cancel := context.WithTimeout(context.Background(), cloneTimeout)
		defer cancel()
		// "--" keeps a source like "--upload-pack=x.git" from being read as an
		// option, and the environment is trimmed so the clone cannot reach the
		// user's API keys, GITHUB_TOKEN, SSH agent or proxy settings. Every
		// other subprocess this program starts is already restricted this way.
		cmd := exec.CommandContext(ctx, "git", "clone", "--depth", "1", "--", source, tempDir)
		cmd.Env = cloneEnv()
		if out, err := cmd.CombinedOutput(); err != nil {
			if ctx.Err() == context.DeadlineExceeded {
				return InstallResult{Dest: destBase, Error: "git clone timed out"}
			}
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
		// filepath.Walk lstats, so a symlink arrives here as an entry rather
		// than as the thing it points at. Copying it would open the target and
		// write its bytes out as a regular file, which both escapes the source
		// tree and defeats the symlink checks in the skills loader, since what
		// lands on disk is no longer a link.
		if info.Mode()&os.ModeSymlink != 0 {
			return nil
		}
		if !info.IsDir() && !info.Mode().IsRegular() {
			return nil
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

// cloneTimeout bounds `git clone` so a hostile or unreachable remote cannot
// hang the install indefinitely.
const cloneTimeout = 2 * time.Minute

// cloneEnv is the allowlisted environment for `git clone`. It deliberately
// omits API keys, GITHUB_TOKEN, SSH_AUTH_SOCK and proxy variables.
func cloneEnv() []string {
	allowed := map[string]bool{
		"PATH": true, "HOME": true, "TMPDIR": true, "TMP": true, "TEMP": true,
		"USER": true, "LOGNAME": true, "LANG": true,
	}
	var env []string
	for _, item := range os.Environ() {
		name, _, _ := strings.Cut(item, "=")
		if allowed[name] || strings.HasPrefix(name, "LC_") {
			env = append(env, item)
		}
	}
	// Never prompt: a clone that needs credentials must fail, not block.
	return append(env, "GIT_TERMINAL_PROMPT=0", "GIT_ASKPASS=", "GCM_INTERACTIVE=never")
}
