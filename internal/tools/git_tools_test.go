package tools

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// gitRepo builds a throwaway repository with one tracked source file and one
// tracked .env, both with local modifications.
func gitRepo(t *testing.T) (Context, string) {
	t.Helper()
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available")
	}
	ctx, root := guardCtx(t)

	run := func(args ...string) {
		t.Helper()
		cmd := exec.Command("git", args...)
		cmd.Dir = root
		cmd.Env = append(os.Environ(),
			"GIT_AUTHOR_NAME=t", "GIT_AUTHOR_EMAIL=t@example.com",
			"GIT_COMMITTER_NAME=t", "GIT_COMMITTER_EMAIL=t@example.com",
		)
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %s: %v\n%s", strings.Join(args, " "), err, out)
		}
	}

	write := func(name, body string) {
		t.Helper()
		if err := os.WriteFile(filepath.Join(root, name), []byte(body), 0o600); err != nil {
			t.Fatal(err)
		}
	}

	run("init", "-q")
	write("app.go", "package main\n")
	write(".env", "SENTRY_DSN=https://abc@sentry.io/1\n")
	run("add", "-A")
	run("commit", "-q", "-m", "initial")

	// Modify both so a bare diff would have something to print for each.
	write("app.go", "package main\n\nfunc main() {}\n")
	write(".env", "SENTRY_DSN=https://abc@sentry.io/1\nSTRIPE=sk_live_deadbeef\n")
	return ctx, root
}

// TestGitDiffExcludesGuardrailedFiles covers the bypass where a bare git_diff
// printed the working tree diff of every tracked file, .env included, without
// consulting blockReadPatterns at all.
func TestGitDiffExcludesGuardrailedFiles(t *testing.T) {
	ctx, _ := gitRepo(t)

	res := gitDiffTool.Execute(ctx, mustArgs(t, map[string]any{}))
	if strings.Contains(res.LLMResult, "sentry.io") || strings.Contains(res.LLMResult, "sk_live") {
		t.Errorf("bare git_diff printed a guardrailed file:\n%s", res.LLMResult)
	}
	if !strings.Contains(res.LLMResult, "func main") {
		t.Errorf("git_diff missed an ordinary change:\n%s", res.LLMResult)
	}
}

// TestGitDiffRefusesGuardrailedPath covers the targeted form.
func TestGitDiffRefusesGuardrailedPath(t *testing.T) {
	ctx, root := gitRepo(t)

	res := gitDiffTool.Execute(ctx, mustArgs(t, map[string]any{
		"filePath": filepath.Join(root, ".env"),
	}))
	if strings.Contains(res.LLMResult, "sentry.io") {
		t.Errorf("git_diff returned a guardrailed file:\n%s", res.LLMResult)
	}
}

// TestGitToolsDoNotUseAShell covers the injection that reached git through a
// repository filename. The tools build argv now, so metacharacters in a path
// are inert.
func TestGitToolsDoNotUseAShell(t *testing.T) {
	ctx, root := gitRepo(t)
	// The marker is relative: the command would run with the repository as its
	// working directory, and a filename cannot contain a separator.
	marker := filepath.Join(root, "pwned")
	nasty := filepath.Join(root, "x'; touch pwned; echo INJECTED #.txt")
	if err := os.WriteFile(nasty, []byte("hi\n"), 0o600); err != nil {
		t.Skipf("filesystem rejected the filename: %v", err)
	}

	res := gitDiffTool.Execute(ctx, mustArgs(t, map[string]any{"filePath": nasty}))
	if strings.Contains(res.LLMResult, "INJECTED") {
		t.Errorf("filename was interpreted by a shell:\n%s", res.LLMResult)
	}
	if _, err := os.Stat(marker); err == nil {
		t.Error("filename injection executed a command")
	}
}

func TestGitStatusReportsChanges(t *testing.T) {
	ctx, _ := gitRepo(t)
	res := gitStatusTool.Execute(ctx, mustArgs(t, map[string]any{}))
	if !strings.Contains(res.LLMResult, "app.go") {
		t.Errorf("git_status missed a modified file:\n%s", res.LLMResult)
	}
}
