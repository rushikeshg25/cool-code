package tools

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestApprovalShowsFullCommandsAndControls(t *testing.T) {
	ctx, root := guardCtx(t)
	for _, name := range []string{"shell_command", "run_tests", "lint_fix"} {
		command := "echo " + strings.Repeat("x", 200) + "\x1b]0; touch marker; #\x07\n echo tail"
		prepared, err := PrepareCommand(ctx, name, args(t, map[string]any{"command": command}))
		if err != nil {
			t.Fatal(err)
		}
		reason := DangerReason(name, prepared)
		for _, want := range []string{strings.Repeat("x", 200), "touch marker", `\x1b`, "echo tail", root} {
			if !strings.Contains(reason, want) {
				t.Errorf("%s preview hides %q: %q", name, want, reason)
			}
		}
		if strings.ContainsAny(reason, "\x1b\x07") {
			t.Fatal("raw terminal controls")
		}
	}
}

func TestPrepareCommandFreezesDetectedCommand(t *testing.T) {
	ctx, root := guardCtx(t)
	if err := os.WriteFile(filepath.Join(root, "package.json"), []byte(`{"scripts":{"test":"echo ok"}}`), 0600); err != nil {
		t.Fatal(err)
	}
	prepared, err := PrepareCommand(ctx, "run_tests", args(t, map[string]any{}))
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "go.mod"), []byte("module example"), 0600); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(DangerReason("run_tests", prepared), "npm run test") {
		t.Fatal("inferred command was not fixed before approval")
	}
}

func TestPrepareCommandRefusesRedactionHiddenSyntax(t *testing.T) {
	ctx, _ := guardCtx(t)
	t.Setenv("COOLCODE_TEST_API_KEY", "literal-secret-value")
	for _, command := range []string{"password=$(>hidden-marker)", "echo literal-secret-value", "api_key=`touch hidden-marker`"} {
		for _, name := range []string{"shell_command", "run_tests", "lint_fix"} {
			if _, err := PrepareCommand(ctx, name, args(t, map[string]any{"command": command})); err == nil {
				t.Fatalf("hidden command accepted: %s", name)
			}
		}
	}
}
