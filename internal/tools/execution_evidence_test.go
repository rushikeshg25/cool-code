package tools

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"
)

func TestExecutionEvidenceSilentFailure(t *testing.T) {
	for _, tool := range []Tool{shellCommandTool, runTestsTool, lintFixTool} {
		t.Run(tool.Name, func(t *testing.T) {
			root := t.TempDir()
			result := tool.Execute(testCtx(root), args(t, map[string]any{"command": "exit 7"}))
			var evidence struct {
				Success  bool
				ExitCode int `json:"exit_code"`
				CWD      string
			}
			if err := json.Unmarshal([]byte(result.LLMResult), &evidence); err != nil {
				t.Fatal(err)
			}
			if !result.Failed || evidence.Success || evidence.ExitCode != 7 || evidence.CWD != root {
				t.Fatalf("lost failure: %s", result.LLMResult)
			}
		})
	}
}

func TestExecutionEvidenceTerminationAndBounds(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	cancelled := execCommand(ctx, "true", t.TempDir(), time.Second)
	if !strings.Contains(cancelled.evidence("true", "."), `"cancelled":true`) {
		t.Fatal("missing cancellation")
	}
	timed := execCommand(context.Background(), "sleep 0.1", t.TempDir(), time.Millisecond)
	if !strings.Contains(timed.evidence("sleep 0.1", "."), `"timed_out":true`) {
		t.Fatal("missing timeout")
	}
	t.Setenv("COOLCODE_TEST_API_KEY", "test-secret-not-for-model")
	r := shellResult{stdout: strings.Repeat("x", 30000), stderr: "test-secret-not-for-model", success: true}
	got := r.evidence("echo test-secret-not-for-model", ".")
	if !json.Valid([]byte(got)) || !strings.Contains(got, `"output_truncated":true`) || strings.Contains(got, "test-secret-not-for-model") {
		t.Fatal("invalid, unbounded or unredacted evidence")
	}
}
