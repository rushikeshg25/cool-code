package agent

import (
	"context"
	"encoding/json"
	"strings"
	"sync"

	"github.com/rushikeshg25/cool-code/internal/config"
	"github.com/rushikeshg25/cool-code/internal/llm"
	"github.com/rushikeshg25/cool-code/internal/project"
	"github.com/rushikeshg25/cool-code/internal/tools"
	"github.com/rushikeshg25/cool-code/internal/types"
)

// Reporter receives live updates during a turn. All methods may be called from
// the processor goroutine.
type Reporter interface {
	Status(text string)         // spinner / progress text
	Assistant(markdown string)  // final or intermediate model text
	Tool(name, display string)  // a tool finished with this display line
	Tasks(list *types.TaskList) // the task list changed
}

// Options configure a Processor.
type Options struct {
	Mode           types.AgentMode
	Quiet          bool
	AllowDangerous bool
	// Confirm is called for dangerous actions; returns true to proceed.
	Confirm func(message string) bool
	// ConfirmEdit is called for edits when confirmEdits is enabled.
	ConfirmEdit func(message, preview string) bool
}

// Processor drives the agentic loop.
type Processor struct {
	rootDir        string
	cfg            config.Config
	provider       llm.Provider
	ctxMgr         *contextManager
	toolDefs       []llm.ToolDef
	toolCtx        tools.Context
	allowDangerous bool
	confirmEdits   bool

	mu       sync.Mutex
	mode     types.AgentMode
	taskList *types.TaskList
	queue    []string

	confirm     func(string) bool
	confirmEdit func(string, string) bool
}

var structuralTools = map[string]bool{"new_file": true, "rename_file": true, "new_module": true}

const updateTaskListTool = "update_task_list"

// New constructs a Processor, returning an error (e.g. *llm.MissingKeyError)
// when the provider can't be initialised.
func New(rootDir string, cfg config.Config, opts Options) (*Processor, error) {
	provider, err := llm.New(cfg.LLM)
	if err != nil {
		return nil, err
	}
	checker := project.NewGitIgnoreChecker(rootDir)
	cm := newContextManager(rootDir, cfg, checker)
	mode := opts.Mode
	if !mode.Valid() {
		mode = types.ModeAgent
	}
	cm.mode = mode

	p := &Processor{
		rootDir:        rootDir,
		cfg:            cfg,
		provider:       provider,
		ctxMgr:         cm,
		toolCtx:        tools.Context{RootDir: rootDir, Config: cfg, GitIgnore: checker},
		allowDangerous: opts.AllowDangerous || cfg.AllowDangerous(),
		confirmEdits:   cfg.ConfirmEdits(),
		mode:           mode,
		confirm:        opts.Confirm,
		confirmEdit:    opts.ConfirmEdit,
	}
	p.buildToolDefs()
	return p, nil
}

func (p *Processor) buildToolDefs() {
	defs := make([]llm.ToolDef, 0, len(tools.All)+1)
	for _, t := range tools.All {
		defs = append(defs, llm.ToolDef{Name: t.Name, Description: t.Description, Parameters: t.Schema})
	}
	defs = append(defs, llm.ToolDef{
		Name:        updateTaskListTool,
		Description: "Records or updates the task list shown to the user. Use it to track multi-step work and mark progress.",
		Parameters: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"goal": map[string]any{"type": "string", "description": "The overall goal."},
				"items": map[string]any{
					"type":        "array",
					"description": "Ordered task items.",
					"items": map[string]any{
						"type": "object",
						"properties": map[string]any{
							"id":     map[string]any{"type": "string"},
							"title":  map[string]any{"type": "string"},
							"status": map[string]any{"type": "string", "enum": []string{"todo", "in-progress", "done", "failed"}},
							"detail": map[string]any{"type": "string"},
						},
						"required": []string{"id", "title", "status"},
					},
				},
			},
			"required": []string{"goal", "items"},
		},
	})
	p.toolDefs = defs
}

// ProcessQuery runs one user query to completion, returning the final assistant
// text. Reporter may be nil.
func (p *Processor) ProcessQuery(ctx context.Context, query string, reporter Reporter) (string, error) {
	p.ctxMgr.addUser(query)
	if reporter != nil {
		reporter.Status(randomThinking())
	}
	// Whatever way this turn ends, never leave a tool call without a result:
	// providers reject histories with dangling tool_use blocks.
	defer p.ctxMgr.closeOpenToolCalls("[Turn interrupted before this tool ran]")

	var finalText string
	for {
		if err := ctx.Err(); err != nil {
			return finalText, err
		}
		req := llm.Request{
			System:      p.ctxMgr.buildSystem(),
			Messages:    p.ctxMgr.window(),
			Tools:       p.toolDefs,
			Temperature: p.cfg.LLM.Temperature,
		}
		resp, err := p.provider.Complete(ctx, req)
		if err != nil {
			return finalText, err
		}
		p.ctxMgr.addAssistant(resp)

		if resp.Text != "" {
			finalText = resp.Text
			if reporter != nil {
				reporter.Assistant(resp.Text)
			}
		}

		if len(resp.ToolCalls) == 0 {
			break
		}

		toolCtx := p.toolCtx
		toolCtx.Ctx = ctx
		treeDirty := false
		interrupted := false
		for _, call := range resp.ToolCalls {
			if err := ctx.Err(); err != nil {
				return finalText, err
			}
			if call.Name == updateTaskListTool {
				p.handleTaskList(call, reporter)
				continue
			}

			// Ask mode blocks mutating tools.
			if p.getMode() == types.ModeAsk && tools.IsMutating(call.Name) {
				p.ctxMgr.addToolResult(call, "[ASK MODE] Cannot execute tool '"+call.Name+"' in Ask mode. Only reading and answering questions is allowed.")
				continue
			}

			if declined, msg := p.gate(call); declined {
				p.ctxMgr.addToolResult(call, msg)
				continue
			}

			if reporter != nil {
				reporter.Status(toolStatus(call.Name))
			}
			result := tools.Run(toolCtx, call.Name, call.Arguments)
			p.ctxMgr.addToolResult(call, result.LLMResult)
			if reporter != nil {
				reporter.Tool(call.Name, result.Display)
			}
			if structuralTools[call.Name] {
				treeDirty = true
			}

			if msg := p.drainQueue(); msg != "" {
				p.ctxMgr.addUser("[SYSTEM: USER INTERRUPTED WITH NEW MESSAGE]\n" + msg)
				interrupted = true
				break
			}
		}

		if treeDirty {
			p.ctxMgr.refreshTree()
		}
		if interrupted {
			if reporter != nil {
				reporter.Status(randomThinking())
			}
			continue
		}
	}

	p.summarize(ctx)
	return finalText, nil
}

// gate applies danger + edit confirmation. Returns (declined, message).
func (p *Processor) gate(call llm.ToolCall) (bool, string) {
	reason := tools.DangerReason(call.Name, call.Arguments)
	preview := tools.EditPreview(call.Name, call.Arguments)
	if reason != "" && !p.allowDangerous {
		if p.confirm == nil || !p.confirm("Allow potentially dangerous action ("+reason+")?") {
			return true, "User declined to run a potentially dangerous tool."
		}
	}
	if preview != "" && p.confirmEdits {
		if p.confirmEdit == nil || !p.confirmEdit("Apply this edit?", preview) {
			return true, "User declined to apply edits."
		}
	}
	return false, ""
}

func (p *Processor) handleTaskList(call llm.ToolCall, reporter Reporter) {
	var list types.TaskList
	if err := json.Unmarshal(call.Arguments, &list); err == nil {
		p.mu.Lock()
		p.taskList = &list
		p.mu.Unlock()
		if reporter != nil {
			reporter.Tasks(&list)
		}
	}
	p.ctxMgr.addToolResult(call, "Task list updated.")
}

func (p *Processor) summarize(ctx context.Context) {
	count, _ := p.ctxMgr.stats()
	if count < 20 {
		return
	}
	var b strings.Builder
	b.WriteString("Summarize the key technical objectives and progress in the following conversation in under 200 words. Focus on specific code changes and design decisions. Avoid filler.\n\n")
	for _, m := range p.ctxMgr.snapshotMessages() {
		if m.Text != "" {
			b.WriteString(string(m.Role) + ": " + m.Text + "\n")
		}
	}
	resp, err := p.provider.Complete(ctx, llm.Request{
		Messages: []llm.Message{{Role: llm.RoleUser, Text: b.String()}},
	})
	if err == nil && resp.Text != "" {
		p.ctxMgr.applySummary(resp.Text)
	}
}

// --- accessors used by the TUI/CLI ---

// Status is a snapshot for the footer.
type Status struct {
	Model        string
	Mode         types.AgentMode
	MessageCount int
	TotalTokens  int
}

// GetStatus returns a footer snapshot.
func (p *Processor) GetStatus() Status {
	count, tokens := p.ctxMgr.stats()
	return Status{
		Model:        p.provider.Model(),
		Mode:         p.getMode(),
		MessageCount: count,
		TotalTokens:  tokens,
	}
}

func (p *Processor) getMode() types.AgentMode {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.mode
}

// Mode returns the current mode.
func (p *Processor) Mode() types.AgentMode { return p.getMode() }

// SetMode changes the operating mode.
func (p *Processor) SetMode(mode types.AgentMode) {
	p.mu.Lock()
	p.mode = mode
	p.mu.Unlock()
	p.ctxMgr.setMode(mode)
}

// SetConfirmHandlers wires interactive confirmation callbacks.
func (p *Processor) SetConfirmHandlers(confirm func(string) bool, confirmEdit func(string, string) bool) {
	p.confirm = confirm
	p.confirmEdit = confirmEdit
}

// EnqueueMessage queues a message to interrupt the running turn.
func (p *Processor) EnqueueMessage(msg string) {
	p.mu.Lock()
	p.queue = append(p.queue, msg)
	p.mu.Unlock()
}

func (p *Processor) drainQueue() string {
	p.mu.Lock()
	defer p.mu.Unlock()
	if len(p.queue) == 0 {
		return ""
	}
	msg := strings.Join(p.queue, "\n")
	p.queue = nil
	return msg
}

// TaskList returns the current task list (may be nil).
func (p *Processor) TaskList() *types.TaskList {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.taskList
}

// PinFile pins an absolute file path into context.
func (p *Processor) PinFile(path string) { p.ctxMgr.pin(path) }

// UnpinFile removes a pinned file.
func (p *Processor) UnpinFile(path string) { p.ctxMgr.unpin(path) }

// PinnedFiles returns the pinned file paths.
func (p *Processor) PinnedFiles() []string { return p.ctxMgr.pinnedFiles() }

// ReloadSkills rebuilds the skills catalog after an install.
func (p *Processor) ReloadSkills() { p.ctxMgr.reloadSkills() }

// RootDir returns the project root.
func (p *Processor) RootDir() string { return p.rootDir }

// Snapshot returns copies of the serializable session state.
func (p *Processor) Snapshot() (messages []llm.Message, summary string, pinned []string, mode types.AgentMode, count int) {
	messages, summary, pinned = p.ctxMgr.sessionState()
	return messages, summary, pinned, p.getMode(), len(messages)
}

// Restore loads persisted session state.
func (p *Processor) Restore(messages []llm.Message, summary string, pinned []string, mode types.AgentMode) {
	p.ctxMgr.restore(messages, summary, pinned)
	// Guard against sessions persisted mid-turn with unanswered tool calls.
	p.ctxMgr.closeOpenToolCalls("[Session restored before this tool ran]")
	if mode.Valid() {
		p.SetMode(mode)
	}
}
