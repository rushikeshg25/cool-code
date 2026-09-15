package agent

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/rushikeshg25/cool-code/internal/llm"
	"github.com/rushikeshg25/cool-code/internal/types"
)

func TestSubagentEnforcesReadOnlyTools(t *testing.T) {
	for _, mode := range []types.AgentMode{types.ModeAgent, types.ModeAsk, types.ModePlan} {
		t.Run(string(mode), func(t *testing.T) {
			root := t.TempDir()
			marker := filepath.Join(root, "marker")
			source := filepath.Join(root, "hello.txt")
			if err := os.WriteFile(source, []byte("hello control"), 0600); err != nil {
				t.Fatal(err)
			}
			provider := &fakeProvider{responses: []llm.Message{{Role: llm.RoleAssistant, ToolCalls: []llm.ToolCall{
				toolCall("1", "shell_command", map[string]any{"command": "touch marker"}),
				toolCall("2", "new_file", map[string]any{"filePath": marker, "content": "bad"}),
				toolCall("3", "git_commit", map[string]any{"message": "bad", "all": true}),
				toolCall("4", "unknown", map[string]any{}),
				toolCall("5", "read_file", map[string]any{"absolutePath": source}),
			}}}}
			p := newTestProcessor(t, root, provider, mode)
			p.runSubagent(context.Background(), "inspect", func(string) {})
			if _, err := os.Stat(marker); !os.IsNotExist(err) {
				t.Fatalf("unexpected mutation: %v", err)
			}
			results := map[string]string{}
			for _, m := range provider.lastReq.Messages {
				if m.Role == llm.RoleTool {
					results[m.ToolCallID] = m.Text
				}
			}
			for _, id := range []string{"1", "2", "3", "4"} {
				if !strings.Contains(results[id], "read-only") {
					t.Errorf("call %s not refused: %q", id, results[id])
				}
			}
			if !strings.Contains(results["5"], "hello control") {
				t.Errorf("read failed: %q", results["5"])
			}
		})
	}
}
