package agent

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"

	"github.com/rushikeshg25/cool-code/internal/config"
	"github.com/rushikeshg25/cool-code/internal/llm"
	"github.com/rushikeshg25/cool-code/internal/project"
	"github.com/rushikeshg25/cool-code/internal/security"
	"github.com/rushikeshg25/cool-code/internal/tools"
	"github.com/rushikeshg25/cool-code/internal/types"
)

// Reporter receives live updates during a turn. All methods may be called from
// the processor goroutine or, for Subagents, from subagent worker goroutines.
type Reporter interface {
	Status(text string)                     // spinner / progress text
	AssistantDelta(text string)             // a streamed fragment of the current reply
	Assistant(markdown string)              // final or intermediate model text
	Tool(name, display string, failed bool) // a tool finished with this display line
	AssistantDiscard()                      // drop streamed text that turned out to be scratch
	Compacted(note string)                  // the conversation was summarized and trimmed
	Tasks(list *types.TaskList)             // the task list changed
	Subagents(lines []string)               // live subagent status lines; nil clears them
}

// Options configure a Processor.
type Options struct {
	Mode           types.AgentMode
	Quiet          bool
	AllowDangerous bool
	// AllowMissingKey builds the Processor without a provider when no API key
	// is configured; ProcessQuery then directs the user to /connect.
	AllowMissingKey bool
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

	mu         sync.Mutex
	mode       types.AgentMode
	taskList   *types.TaskList
	queue      []string
	lastUsage  llm.Usage
	totalUsage llm.Usage // cumulative across the session, for cost

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
		var mk *llm.MissingKeyError
		if !(opts.AllowMissingKey && errors.As(err, &mk)) {
			return nil, err
		}
		provider = nil
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
	defs := make([]llm.ToolDef, 0, len(tools.All)+2)
	for _, t := range tools.All {
		defs = append(defs, llm.ToolDef{Name: t.Name, Description: t.Description, Parameters: t.Schema})
	}
	defs = append(defs, spawnAgentDef)
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
	if p.provider == nil {
		return "", errors.New("no API key configured - run /connect to set one up")
	}
	p.ctxMgr.addUser(security.Redact(query))
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
		var resp llm.Message
		var err error
		streamed := false
		if streamer, ok := p.provider.(llm.Streamer); ok && reporter != nil {
			sink := newStreamSink(reporter)
			resp, err = streamer.Stream(ctx, req, sink.write)
			streamed = sink.flush()
		} else {
			resp, err = p.provider.Complete(ctx, req)
		}
		if err != nil {
			return finalText, err
		}
		if resp.Usage.Input > 0 || resp.Usage.Output > 0 {
			p.mu.Lock()
			p.lastUsage = resp.Usage
			p.totalUsage.Input += resp.Usage.Input
			p.totalUsage.Output += resp.Usage.Output
			p.mu.Unlock()
		}
		p.ctxMgr.addAssistant(redactMessage(resp))

		if len(resp.ToolCalls) == 0 {
			finalText = security.Redact(resp.Text)
			if finalText != "" && reporter != nil {
				// Replaces the streamed text with its markdown rendering.
				reporter.Assistant(finalText)
			} else if streamed && reporter != nil {
				reporter.AssistantDiscard()
			}
			break
		}

		// The turn continues, so whatever was streamed was the model talking
		// to itself before deciding on a tool. Drop it: the tool lines that
		// follow say what actually happened.
		if streamed && reporter != nil {
			reporter.AssistantDiscard()
		}

		toolCtx := p.toolCtx
		toolCtx.Ctx = ctx
		treeDirty := false

		// Sequential pre-pass in call order: resolve the task list, ask-mode
		// blocks, and confirmation prompts (only one overlay can exist), and
		// partition the surviving calls into read-only and mutating sets.
		n := len(resp.ToolCalls)
		results := make([]*types.ToolResult, n)
		executed := make([]bool, n) // true when a real tool ran (fires reporter.Tool)
		var readIdx, mutIdx, spawnIdx []int
		for i, call := range resp.ToolCalls {
			switch {
			case call.Name == updateTaskListTool:
				results[i] = &types.ToolResult{LLMResult: p.handleTaskList(call, reporter)}
			case call.Name == spawnAgentTool:
				spawnIdx = append(spawnIdx, i)
			case p.getMode() != types.ModeAgent && tools.IsMutating(call.Name):
				results[i] = &types.ToolResult{LLMResult: "[READ-ONLY MODE] Cannot execute tool '" + call.Name + "'. Switch to Agent mode to make changes or execute project code."}
			default:
				if declined, msg := p.gate(call); declined {
					results[i] = &types.ToolResult{LLMResult: msg}
				} else if tools.IsReadOnly(call.Name) {
					readIdx = append(readIdx, i)
				} else {
					mutIdx = append(mutIdx, i)
				}
			}
		}

		// Read-only calls and subagents run concurrently under one WaitGroup;
		// each worker writes only results[i].
		if len(readIdx)+len(spawnIdx) > 0 {
			if reporter != nil {
				switch {
				case len(spawnIdx) > 0:
					reporter.Status("Running " + strconv.Itoa(len(spawnIdx)) + " subagent(s)…")
				case len(readIdx) == 1:
					reporter.Status(toolStatus(resp.ToolCalls[readIdx[0]].Name))
				default:
					reporter.Status("Running " + strconv.Itoa(len(readIdx)) + " tools in parallel…")
				}
			}
			var wg sync.WaitGroup
			for _, i := range readIdx {
				wg.Add(1)
				go func(i int, call llm.ToolCall) {
					defer wg.Done()
					r := tools.Run(toolCtx, call.Name, call.Arguments)
					results[i] = &r
				}(i, resp.ToolCalls[i])
			}
			if len(spawnIdx) > 0 {
				lines := make([]string, len(spawnIdx))
				var linesMu sync.Mutex
				push := func() {
					if reporter == nil {
						return
					}
					linesMu.Lock()
					cp := append([]string(nil), lines...)
					linesMu.Unlock()
					reporter.Subagents(cp)
				}
				for k, i := range spawnIdx {
					task, ok := parseSpawnTask(resp.ToolCalls[i].Arguments)
					if !ok {
						results[i] = &types.ToolResult{Display: "Subagent failed", LLMResult: "Invalid arguments: spawn_agent requires a non-empty task."}
						continue
					}
					label := "agent " + strconv.Itoa(k+1) + ": " + truncateTask(security.Redact(task))
					setLine := func(k int, state string) {
						linesMu.Lock()
						lines[k] = label + " - " + state
						linesMu.Unlock()
						push()
					}
					setLine(k, "starting")
					wg.Add(1)
					go func(k, i int, task string) {
						defer wg.Done()
						text := p.runSubagent(ctx, task, func(state string) { setLine(k, state) })
						results[i] = &types.ToolResult{Display: "Subagent: " + truncateTask(task), LLMResult: text}
						setLine(k, "done")
					}(k, i, task)
				}
			}
			wg.Wait()
			if len(spawnIdx) > 0 && reporter != nil {
				reporter.Subagents(nil)
			}
			for _, i := range readIdx {
				executed[i] = true
			}
			for _, i := range spawnIdx {
				executed[i] = true
			}
		}

		// Mutating calls run sequentially, in original order.
		var cancelErr error
		for _, i := range mutIdx {
			if err := ctx.Err(); err != nil {
				cancelErr = err
				break
			}
			call := resp.ToolCalls[i]
			if reporter != nil {
				reporter.Status(toolStatus(call.Name))
			}
			r := tools.Run(toolCtx, call.Name, call.Arguments)
			results[i] = &r
			executed[i] = true
			if structuralTools[call.Name] {
				treeDirty = true
			}
		}

		// Append results in original call order so tool_call/result pairing
		// stays deterministic. Calls skipped by cancellation stay nil and are
		// closed by the deferred closeOpenToolCalls.
		for i, call := range resp.ToolCalls {
			if results[i] == nil {
				continue
			}
			results[i].LLMResult = security.Redact(results[i].LLMResult)
			p.ctxMgr.addToolResult(call, results[i].LLMResult)
			if executed[i] && reporter != nil {
				reporter.Tool(call.Name, results[i].Display, results[i].Failed)
			}
		}

		if treeDirty {
			p.ctxMgr.refreshTree()
		}
		if cancelErr != nil {
			return finalText, cancelErr
		}
		if msg := p.drainQueue(); msg != "" {
			p.ctxMgr.addUser("[SYSTEM: USER INTERRUPTED WITH NEW MESSAGE]\n" + msg)
			if reporter != nil {
				reporter.Status(randomThinking())
			}
		}
	}

	// Compaction is destructive and used to happen with no signal at all, so
	// a session silently lost most of its history. Say so when it runs.
	if note := p.summarize(ctx); note != "" && reporter != nil {
		reporter.Status("")
		reporter.Compacted(note)
	}
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

// handleTaskList records the model's task list and returns the tool result.
func (p *Processor) handleTaskList(call llm.ToolCall, reporter Reporter) string {
	var list types.TaskList
	if err := json.Unmarshal([]byte(security.Redact(string(call.Arguments))), &list); err == nil {
		p.mu.Lock()
		p.taskList = &list
		p.mu.Unlock()
		if reporter != nil {
			reporter.Tasks(&list)
		}
	}
	return "Task list updated."
}

// summarize condenses the conversation once it grows past the configured
// threshold, then trims the history it summarized.
//
// The request used to concatenate every message with no budget at all, so on a
// long conversation the compaction call could itself exceed the context window,
// and it ran on every turn past the threshold rather than only when the history
// had grown since the last one.
func (p *Processor) summarize(ctx context.Context) string {
	threshold := p.cfg.CompactAfter()
	if threshold <= 0 {
		return ""
	}
	count, _ := p.ctxMgr.stats()
	if count < threshold {
		return ""
	}
	return p.compact(ctx, false)
}

// compact summarizes and trims the conversation. When manual, it runs
// regardless of the threshold and reports the outcome.
func (p *Processor) compact(ctx context.Context, manual bool) string {
	before, _ := p.ctxMgr.stats()
	if before <= keepRecentMessages {
		return "Nothing to compact yet."
	}

	var b strings.Builder
	b.WriteString("Summarize the key technical objectives and progress in the following conversation in under 200 words. Focus on specific code changes and design decisions. Avoid filler.\n\n")
	if prior := p.ctxMgr.currentSummary(); prior != "" {
		b.WriteString("Summary of the conversation before this excerpt:\n" + prior + "\n\n")
	}
	b.WriteString(p.ctxMgr.transcriptForSummary(summaryInputBudget))

	resp, err := p.provider.Complete(ctx, llm.Request{
		Messages: []llm.Message{{Role: llm.RoleUser, Text: b.String()}},
	})
	if err != nil {
		if manual {
			return "Compaction failed: " + err.Error()
		}
		return ""
	}
	if resp.Text == "" {
		if manual {
			return "Compaction produced no summary; history left as is."
		}
		return ""
	}
	p.ctxMgr.applySummary(security.Redact(resp.Text))
	after, _ := p.ctxMgr.stats()
	return fmt.Sprintf("Compacted %d messages into a summary; %d kept.", before-after, after)
}

// Compact runs compaction on demand, for the /compact command.
func (p *Processor) Compact(ctx context.Context) string {
	return p.compact(ctx, true)
}

func redactMessage(message llm.Message) llm.Message {
	message.Text = security.Redact(message.Text)
	message.ToolCalls = append([]llm.ToolCall(nil), message.ToolCalls...)
	for i := range message.ToolCalls {
		message.ToolCalls[i].Arguments = json.RawMessage(security.Redact(string(message.ToolCalls[i].Arguments)))
	}
	return message
}

// --- accessors used by the TUI/CLI ---

// Status is a snapshot for the footer.
type Status struct {
	Model        string
	Effort       string
	Mode         types.AgentMode
	MessageCount int
	// TotalTokens is the context size: the last request's real input tokens
	// when the provider reported usage, otherwise a len/4 estimate.
	TotalTokens int
	// Estimated is true when TotalTokens is the fallback estimate.
	Estimated bool
	// MaxTokens is the configured context window, so callers can report how
	// full it is rather than an absolute token count that means little on its
	// own.
	MaxTokens int
	// SessionCost is the running spend in US dollars, summed across every
	// request this session. Zero when the model's rate is not known.
	SessionCost float64
	// CostKnown is false for a model with no published rate, including any
	// custom endpoint, so the UI can stay silent rather than show $0.00.
	CostKnown bool
}

// GetStatus returns a footer snapshot.
func (p *Processor) GetStatus() Status {
	count, tokens := p.ctxMgr.stats()
	p.mu.Lock()
	usage := p.lastUsage
	total := p.totalUsage
	p.mu.Unlock()
	model := p.cfg.LLM.Model + " (not connected)"
	if p.provider != nil {
		model = p.provider.Model()
	}
	s := Status{
		Model:        model,
		Effort:       p.cfg.LLM.ReasoningEffort,
		Mode:         p.getMode(),
		MessageCount: count,
		TotalTokens:  tokens,
		MaxTokens:    p.cfg.MaxContextTokens(),
		Estimated:    true,
	}
	if usage.Input > 0 {
		s.TotalTokens = usage.Input
		s.Estimated = false
	}
	if total.Input > 0 || total.Output > 0 {
		if cost, ok := llm.Cost(model, total.Input, total.Output); ok {
			s.SessionCost = cost
			s.CostKnown = true
		}
	}
	return s
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

// ConfigureLLM rebuilds the provider from new LLM settings (after /connect).
// Only callable between turns: slash commands never run while processing.
func (p *Processor) ConfigureLLM(llmCfg config.LLM) error {
	provider, err := llm.New(llmCfg)
	if err != nil {
		return err
	}
	p.cfg.LLM = llmCfg
	p.provider = provider
	return nil
}

// LLMConfig returns the current provider settings, so callers can adjust one
// field and hand the whole thing back to ConfigureLLM.
func (p *Processor) LLMConfig() config.LLM {
	return p.cfg.LLM
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

// PinFile pins an allowed absolute file path into context.
func (p *Processor) PinFile(path string) error {
	p.mu.Lock()
	toolCtx := p.toolCtx
	p.mu.Unlock()
	if reason := tools.ValidateReadPath(path, toolCtx); reason != "" {
		return errors.New(reason)
	}
	info, err := os.Stat(path)
	if err != nil {
		return err
	}
	if info.IsDir() {
		return errors.New("path is not a file")
	}
	p.ctxMgr.pin(path)
	return nil
}

// UnpinFile removes a pinned file.
func (p *Processor) UnpinFile(path string) { p.ctxMgr.unpin(path) }

// PinnedFiles returns the pinned file paths.
func (p *Processor) PinnedFiles() []string { return p.ctxMgr.pinnedFiles() }

// ReloadSkills rebuilds the skills catalog after an install.
func (p *Processor) ReloadSkills() { p.ctxMgr.reloadSkills() }

// RootDir returns the project root.
func (p *Processor) RootDir() string { return p.rootDir }

// AddDir grants the session access to an additional directory (/add-dir). It
// resolves ~ and relative paths against the project root, validates the target,
// and returns the resolved absolute path.
func (p *Processor) AddDir(path string) (string, error) {
	if path == "" {
		return "", errors.New("usage: /add-dir <path>")
	}
	if path == "~" || strings.HasPrefix(path, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		path = filepath.Join(home, strings.TrimPrefix(path, "~"))
	}
	if !filepath.IsAbs(path) {
		path = filepath.Join(p.rootDir, path)
	}
	abs, err := filepath.Abs(filepath.Clean(path))
	if err != nil {
		return "", err
	}
	abs, err = filepath.EvalSymlinks(abs)
	if err != nil {
		return "", errors.New("directory not found: " + abs)
	}
	info, err := os.Stat(abs)
	if err != nil {
		return "", errors.New("directory not found: " + abs)
	}
	if !info.IsDir() {
		return "", errors.New("not a directory: " + abs)
	}

	p.mu.Lock()
	roots := p.toolCtx.Roots()
	p.mu.Unlock()
	for _, root := range roots {
		if tools.EnsureAbsoluteWithinRoot(abs, root) == "" {
			return "", errors.New(abs + " is already accessible (inside " + root + ")")
		}
	}

	p.mu.Lock()
	p.toolCtx.ExtraDirs = append(append([]string(nil), p.toolCtx.ExtraDirs...), abs)
	p.mu.Unlock()
	p.ctxMgr.addExtraDir(abs)
	return abs, nil
}

// ExtraDirs returns the directories added via AddDir.
func (p *Processor) ExtraDirs() []string { return p.ctxMgr.extraDirList() }

// Connected reports whether a provider is configured.
func (p *Processor) Connected() bool { return p.provider != nil }

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

// streamSink turns provider deltas into whole redacted lines.
//
// Redaction cannot run on a raw fragment, because a provider splits wherever
// it likes and a credential can straddle two of them. Buffering to the last
// newline keeps every emitted line complete, so security.Redact sees the same
// text it would have seen in a non-streamed response. The cost is that output
// appears a line at a time rather than a token at a time.
type streamSink struct {
	reporter Reporter
	buf      strings.Builder
	emitted  bool
}

func newStreamSink(reporter Reporter) *streamSink {
	return &streamSink{reporter: reporter}
}

func (s *streamSink) write(chunk string) {
	s.buf.WriteString(chunk)
	text := s.buf.String()
	cut := strings.LastIndexByte(text, '\n')
	if cut < 0 {
		return
	}
	complete, rest := text[:cut+1], text[cut+1:]
	s.buf.Reset()
	s.buf.WriteString(rest)
	if out := security.Redact(complete); out != "" {
		s.emitted = true
		s.reporter.AssistantDelta(out)
	}
}

// flush emits any trailing partial line and reports whether anything was shown.
func (s *streamSink) flush() bool {
	if rest := s.buf.String(); rest != "" {
		s.buf.Reset()
		if out := security.Redact(rest); out != "" {
			s.emitted = true
			s.reporter.AssistantDelta(out)
		}
	}
	return s.emitted
}
