package agent

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

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
	mu            sync.Mutex
	tools         []string
	texts         []string
	tasks         *types.TaskList
	subagentLines []string
}

func (c *captureReporter) Status(string)         {}
func (c *captureReporter) AssistantDelta(string) {}
func (c *captureReporter) Assistant(md string)   { c.texts = append(c.texts, md) }
func (c *captureReporter) Tool(name, _ string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.tools = append(c.tools, name)
}
func (c *captureReporter) Tasks(l *types.TaskList) { c.tasks = l }
func (c *captureReporter) Subagents(lines []string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if len(lines) > 0 {
		c.subagentLines = append(c.subagentLines, lines...)
	}
}

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

// blockingProvider blocks until the context is cancelled.
type blockingProvider struct{ started chan struct{} }

func (b *blockingProvider) Name() string  { return "blocking" }
func (b *blockingProvider) Model() string { return "blocking-model" }
func (b *blockingProvider) Complete(ctx context.Context, _ llm.Request) (llm.Message, error) {
	close(b.started)
	<-ctx.Done()
	return llm.Message{}, ctx.Err()
}

func TestProcessQueryCancellation(t *testing.T) {
	root := t.TempDir()
	provider := &blockingProvider{started: make(chan struct{})}
	p := newTestProcessor(t, root, provider, types.ModeAgent)

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() {
		_, err := p.ProcessQuery(ctx, "hang forever", nil)
		done <- err
	}()
	<-provider.started
	cancel()
	select {
	case err := <-done:
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("expected context.Canceled, got %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("ProcessQuery did not return after cancellation")
	}
}

func TestCloseOpenToolCalls(t *testing.T) {
	root := t.TempDir()
	p := newTestProcessor(t, root, &fakeProvider{}, types.ModeAgent)
	cm := p.ctxMgr
	cm.addUser("do things")
	cm.addAssistant(llm.Message{Role: llm.RoleAssistant, ToolCalls: []llm.ToolCall{
		toolCall("a", "read_file", nil), toolCall("b", "grep", nil),
	}})
	cm.addToolResult(llm.ToolCall{ID: "a", Name: "read_file"}, "result a")
	// Tool call "b" has no result — as after a mid-turn cancellation.
	cm.closeOpenToolCalls("[interrupted]")

	last := cm.messages[len(cm.messages)-1]
	if last.Role != llm.RoleTool || last.ToolCallID != "b" || last.Text != "[interrupted]" {
		t.Fatalf("dangling call not closed, last = %+v", last)
	}
	// Idempotent: a second pass adds nothing.
	n := len(cm.messages)
	cm.closeOpenToolCalls("[interrupted]")
	if len(cm.messages) != n {
		t.Fatal("closeOpenToolCalls is not idempotent")
	}
}

func TestConfirmEditsIndependentOfAllowDangerous(t *testing.T) {
	root := t.TempDir()
	target := filepath.Join(root, "x.txt")
	_ = os.WriteFile(target, []byte("original"), 0o644)

	provider := &fakeProvider{responses: []llm.Message{
		{Role: llm.RoleAssistant, ToolCalls: []llm.ToolCall{
			toolCall("1", "edit_file", map[string]any{"filePath": target, "oldString": "original", "newString": "changed"}),
		}},
		{Role: llm.RoleAssistant, Text: "done"},
	}}
	p := newTestProcessor(t, root, provider, types.ModeAgent)
	p.confirmEdits = true // allowDangerous is already true in newTestProcessor
	confirmCalled := false
	p.confirmEdit = func(_, _ string) bool {
		confirmCalled = true
		return false // decline the edit
	}

	if _, err := p.ProcessQuery(context.Background(), "edit x.txt", nil); err != nil {
		t.Fatal(err)
	}
	if !confirmCalled {
		t.Fatal("confirmEdit was not called despite confirmEdits=true (allowDangerous must not skip it)")
	}
	raw, _ := os.ReadFile(target)
	if string(raw) != "original" {
		t.Fatalf("declined edit was applied anyway: %s", raw)
	}
}

func TestParallelToolResultsKeepCallOrder(t *testing.T) {
	root := t.TempDir()
	for _, name := range []string{"a.txt", "b.txt", "c.txt"} {
		if err := os.WriteFile(filepath.Join(root, name), []byte("content of "+name), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	target := filepath.Join(root, "new.txt")
	provider := &fakeProvider{responses: []llm.Message{
		{Role: llm.RoleAssistant, ToolCalls: []llm.ToolCall{
			toolCall("1", "read_file", map[string]any{"absolutePath": filepath.Join(root, "a.txt")}),
			toolCall("2", "read_file", map[string]any{"absolutePath": filepath.Join(root, "b.txt")}),
			toolCall("3", "new_file", map[string]any{"filePath": target, "content": "made"}),
			toolCall("4", "read_file", map[string]any{"absolutePath": filepath.Join(root, "c.txt")}),
		}},
		{Role: llm.RoleAssistant, Text: "done"},
	}}
	p := newTestProcessor(t, root, provider, types.ModeAgent)
	rep := &captureReporter{}

	if _, err := p.ProcessQuery(context.Background(), "read and create", rep); err != nil {
		t.Fatal(err)
	}
	// Tool results must appear in original call order regardless of the
	// read-only calls running concurrently.
	var ids []string
	for _, m := range p.ctxMgr.messages {
		if m.Role == llm.RoleTool {
			ids = append(ids, m.ToolCallID)
		}
	}
	want := []string{"1", "2", "3", "4"}
	if len(ids) != len(want) {
		t.Fatalf("tool results = %v", ids)
	}
	for i := range want {
		if ids[i] != want[i] {
			t.Fatalf("result order = %v, want %v", ids, want)
		}
	}
	if raw, _ := os.ReadFile(target); string(raw) != "made" {
		t.Fatalf("mutating tool did not run: %q", raw)
	}
}

// subagentAwareProvider serves the main loop (scripted) and any number of
// concurrent subagent conversations (stateless: grep first, then report).
type subagentAwareProvider struct {
	mu        sync.Mutex
	mainCalls int
}

func (s *subagentAwareProvider) Name() string  { return "fake" }
func (s *subagentAwareProvider) Model() string { return "fake-model" }
func (s *subagentAwareProvider) Complete(_ context.Context, req llm.Request) (llm.Message, error) {
	if strings.Contains(req.System, "explore subagent") {
		last := req.Messages[len(req.Messages)-1]
		if last.Role == llm.RoleTool {
			return llm.Message{Role: llm.RoleAssistant, Text: "findings for: " + req.Messages[0].Text}, nil
		}
		return llm.Message{Role: llm.RoleAssistant, ToolCalls: []llm.ToolCall{
			{ID: "s1", Name: "glob", Arguments: json.RawMessage(`{"pattern":"*.txt"}`)},
		}}, nil
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.mainCalls++
	if s.mainCalls == 1 {
		return llm.Message{Role: llm.RoleAssistant, ToolCalls: []llm.ToolCall{
			{ID: "1", Name: "spawn_agent", Arguments: json.RawMessage(`{"task":"explore area A"}`)},
			{ID: "2", Name: "spawn_agent", Arguments: json.RawMessage(`{"task":"explore area B"}`)},
		}}, nil
	}
	return llm.Message{Role: llm.RoleAssistant, Text: "combined report"}, nil
}

func TestSpawnAgentFanOut(t *testing.T) {
	root := t.TempDir()
	_ = os.WriteFile(filepath.Join(root, "x.txt"), []byte("hi"), 0o644)
	provider := &subagentAwareProvider{}
	p := newTestProcessor(t, root, provider, types.ModeAgent)
	rep := &captureReporter{}

	final, err := p.ProcessQuery(context.Background(), "explore everything", rep)
	if err != nil {
		t.Fatal(err)
	}
	if final != "combined report" {
		t.Fatalf("final = %q", final)
	}
	// Both subagent results must be present, in call order.
	var got []string
	for _, m := range p.ctxMgr.messages {
		if m.Role == llm.RoleTool && m.ToolName == "spawn_agent" {
			got = append(got, m.Text)
		}
	}
	if len(got) != 2 ||
		!strings.Contains(got[0], "explore area A") ||
		!strings.Contains(got[1], "explore area B") {
		t.Fatalf("subagent results = %v", got)
	}
	rep.mu.Lock()
	defer rep.mu.Unlock()
	if len(rep.subagentLines) == 0 {
		t.Fatal("no subagent status lines reported")
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
