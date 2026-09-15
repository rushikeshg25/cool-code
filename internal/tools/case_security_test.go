package tools

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestCaseAliasesRespectGuards(t *testing.T) {
	ctx, root := guardCtx(t)
	for _, name := range []string{".GIT/config", ".CoolCode/skills/demo/SKILL.md", ".COOLCODE.JSON", ".ENV", "cert.PEM"} {
		if _, reason := ResolveWritePath(filepath.Join(root, name), ctx); reason == "" {
			t.Errorf("allowed write: %s", name)
		}
	}
	for _, name := range []string{".ENV", ".Env.LOCAL", "ID_RSA", "cert.PEM"} {
		path := filepath.Join(root, name)
		if err := os.WriteFile(path, []byte("CASE_SENTINEL"), 0600); err != nil {
			t.Fatal(err)
		}
		if _, reason := ResolveReadPath(path, ctx); reason == "" {
			t.Errorf("allowed read: %s", name)
		}
	}
	if res := grepTool.Execute(ctx, args(t, map[string]any{"pattern": "CASE_SENTINEL"})); strings.Contains(res.LLMResult, "CASE_SENTINEL") {
		t.Fatal(res.LLMResult)
	}
	if _, err := exec.LookPath("rg"); err == nil {
		if res := findSymbolTool.Execute(ctx, args(t, map[string]any{"pattern": "CASE_SENTINEL"})); strings.Contains(res.LLMResult, "CASE_SENTINEL") {
			t.Fatal(res.LLMResult)
		}
	}
	if _, reason := ResolveWritePath(filepath.Join(root, "README.md"), ctx); reason != "" {
		t.Fatal(reason)
	}
}

func TestGitDiffExcludesCaseAliases(t *testing.T) {
	ctx, root := gitRepo(t)
	path := filepath.Join(root, "KEY.PEM")
	for _, body := range []string{"before\n", "CASE_SENTINEL\n"} {
		if err := os.WriteFile(path, []byte(body), 0600); err != nil {
			t.Fatal(err)
		}
		if body == "before\n" {
			cmd := exec.Command("git", "add", "KEY.PEM")
			cmd.Dir = root
			if out, err := cmd.CombinedOutput(); err != nil {
				t.Fatalf("%v %s", err, out)
			}
		}
	}
	res := gitDiffTool.Execute(ctx, args(t, map[string]any{}))
	if strings.Contains(res.LLMResult, "CASE_SENTINEL") || !strings.Contains(res.LLMResult, "func main") {
		t.Fatal(res.LLMResult)
	}
}
