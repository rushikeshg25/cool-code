package agent

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/rushikeshg25/cool-code/internal/config"
	"github.com/rushikeshg25/cool-code/internal/llm"
	"github.com/rushikeshg25/cool-code/internal/project"
	"github.com/rushikeshg25/cool-code/internal/tools"
	"github.com/rushikeshg25/cool-code/internal/types"
)

// fakeProvider scripts a sequence of responses for the agent loop.
type fakeProvider struct {
	responses []llm.Message
	calls     int
	lastReq   llm.Request
}

func (f *fakeProvider) Name() string  { return "fake" }
func (f *fakeProvider) Model() string { return "fake-model" }
func (f *fakeProvider) Complete(_ context.Context, req llm.Request) (llm.Message, error) {
	f.lastReq = req
	i := f.calls
	f.calls++
	if i < len(f.responses) {
		return f.responses[i], nil
	}
	return llm.Message{Role: llm.RoleAssistant, Text: "done"}, nil
}

type captureReporter struct {
	tools []string
	texts []string
	tasks *types.TaskList
}

func (c *captureReporter) Status(string)           {}
func (c *captureReporter) Assistant(md string)     { c.texts = append(c.texts, md) }
func (c *captureReporter) Tool(name, _ string)     { c.tools = append(c.tools, name) }
func (c *captureReporter) Tasks(l *types.TaskList) { c.tasks = l }

func newTestProcessor(t *testing.T, root string, provider llm.Provider, mode types.AgentMode) *Processor {
	t.Helper()
	cfg := config.Default()
	checker := project.NewGitIgnoreChecker(root)
	cm := newContextManager(root, cfg, checker)
	cm.mode = mode
	p := &Processor{
		rootDir:        root,
		cfg:            cfg,
		provider:       provider,
		ctxMgr:         cm,
		toolCtx:        tools.Context{RootDir: root, Config: cfg, GitIgnore: checker},
		allowDangerous: true,
		mode:           mode,
	}
	p.buildToolDefs()
	return p
}

func toolCall(id, name string, args map[string]any) llm.ToolCall {
	raw, _ := json.Marshal(args)
	return llm.ToolCall{ID: id, Name: name, Arguments: raw}
}

func TestProcessQueryToolLoop(t *testing.T) {
	root := t.TempDir()
	file := filepath.Join(root, "hello.txt")
	if err := os.WriteFile(file, []byte("secret contents"), 0o644); err != nil {
		t.Fatal(err)
	}

	provider := &fakeProvider{responses: []llm.Message{
		{Role: llm.RoleAssistant, ToolCalls: []llm.ToolCall{toolCall("1", "read_file", map[string]any{"absolutePath": file})}},
		{Role: llm.RoleAssistant, Text: "The file says: secret contents"},
	}}
	p := newTestProcessor(t, root, provider, types.ModeAgent)
	rep := &captureReporter{}

	final, err := p.ProcessQuery(context.Background(), "what does hello.txt say?", rep)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(final, "secret contents") {
		t.Fatalf("final = %q", final)
	}
	if len(rep.tools) != 1 || rep.tools[0] != "read_file" {
		t.Fatalf("tool events = %v", rep.tools)
	}
	if provider.calls != 2 {
		t.Fatalf("expected 2 provider calls, got %d", provider.calls)
	}
	// The tool result must have been fed back into the message history.
	var sawResult bool
	for _, m := range p.ctxMgr.messages {
		if m.Role == llm.RoleTool && strings.Contains(m.Text, "secret contents") {
			sawResult = true
		}
	}
	if !sawResult {
		t.Fatal("tool result not fed back into context")
	}
}

func TestAskModeBlocksMutating(t *testing.T) {
	root := t.TempDir()
	target := filepath.Join(root, "x.txt")
	_ = os.WriteFile(target, []byte("original"), 0o644)

	provider := &fakeProvider{responses: []llm.Message{
		{Role: llm.RoleAssistant, ToolCalls: []llm.ToolCall{
			toolCall("1", "new_file", map[string]any{"filePath": target, "content": "overwritten"}),
		}},
		{Role: llm.RoleAssistant, Text: "I cannot edit in ask mode."},
	}}
	p := newTestProcessor(t, root, provider, types.ModeAsk)

	if _, err := p.ProcessQuery(context.Background(), "please overwrite x.txt", nil); err != nil {
		t.Fatal(err)
	}
	raw, _ := os.ReadFile(target)
	if string(raw) != "original" {
		t.Fatalf("ask mode should have blocked the write, file = %q", raw)
	}
	var blocked bool
	for _, m := range p.ctxMgr.messages {
		if m.Role == llm.RoleTool && strings.Contains(m.Text, "ASK MODE") {
			blocked = true
		}
	}
	if !blocked {
		t.Fatal("expected an ASK MODE refusal in context")
	}
}

func TestProcessQueryTaskList(t *testing.T) {
	root := t.TempDir()
	provider := &fakeProvider{responses: []llm.Message{
		{Role: llm.RoleAssistant, ToolCalls: []llm.ToolCall{
			toolCall("1", "update_task_list", map[string]any{
				"goal":  "Ship it",
				"items": []map[string]any{{"id": "1", "title": "Do the thing", "status": "todo"}},
			}),
		}},
		{Role: llm.RoleAssistant, Text: "planned"},
	}}
	p := newTestProcessor(t, root, provider, types.ModeAgent)
	rep := &captureReporter{}
	if _, err := p.ProcessQuery(context.Background(), "plan it", rep); err != nil {
		t.Fatal(err)
	}
	list := p.TaskList()
	if list == nil || list.Goal != "Ship it" || len(list.Items) != 1 {
		t.Fatalf("task list = %+v", list)
	}
	if rep.tasks == nil {
		t.Fatal("reporter did not receive task update")
	}
}
