package agent

import (
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/rushikeshg25/cool-code/internal/config"
	"github.com/rushikeshg25/cool-code/internal/llm"
	"github.com/rushikeshg25/cool-code/internal/memory"
	"github.com/rushikeshg25/cool-code/internal/project"
	"github.com/rushikeshg25/cool-code/internal/security"
	"github.com/rushikeshg25/cool-code/internal/skills"
	"github.com/rushikeshg25/cool-code/internal/tools"
	"github.com/rushikeshg25/cool-code/internal/types"
)

// contextManager holds the conversation and builds the system prompt + windowed
// message list for each turn. mu guards all mutable state: the processor
// goroutine appends messages while the TUI goroutine reads stats/snapshots.
type contextManager struct {
	rootDir             string
	gitIgnore           project.GitIgnoreChecker
	cfg                 config.Config
	maxTokens           int
	projectInstructions string

	mu            sync.Mutex
	mode          types.AgentMode
	messages      []llm.Message
	summary       string
	fileTree      string
	skillsCatalog string
	pinned        []string
	extraDirs     []string // additional /add-dir roots
	extraTrees    []string // rendered tree per extra dir, parallel to extraDirs
}

func newContextManager(rootDir string, cfg config.Config, checker project.GitIgnoreChecker) *contextManager {
	maxDepth := -1
	if cfg.Features.FileTreeMaxDepth != nil {
		maxDepth = *cfg.Features.FileTreeMaxDepth
	}
	return &contextManager{
		rootDir:             rootDir,
		gitIgnore:           checker,
		cfg:                 cfg,
		mode:                types.ModeAgent,
		maxTokens:           cfg.MaxContextTokens(),
		fileTree:            project.FolderStructure(rootDir, privacyChecker(rootDir, cfg, checker), maxDepth),
		projectInstructions: memory.LoadProjectInstructions(rootDir),
		skillsCatalog:       skills.Catalog(rootDir),
	}
}

func estimateTokens(s string) int { return (len(s) + 3) / 4 }

func (c *contextManager) addUser(text string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.messages = append(c.messages, llm.Message{Role: llm.RoleUser, Text: text})
}

func (c *contextManager) addAssistant(m llm.Message) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.messages = append(c.messages, m)
}

func (c *contextManager) addToolResult(call llm.ToolCall, result string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.messages = append(c.messages, llm.Message{
		Role:       llm.RoleTool,
		Text:       result,
		ToolCallID: call.ID,
		ToolName:   call.Name,
	})
}

// closeOpenToolCalls appends a synthetic result for every tool call in the
// last assistant message that has no matching result yet, keeping the
// tool_use/tool_result pairing valid after an aborted turn. No-op when the
// history is already consistent.
func (c *contextManager) closeOpenToolCalls(note string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	last := -1
	for i := len(c.messages) - 1; i >= 0; i-- {
		if c.messages[i].Role == llm.RoleAssistant {
			last = i
			break
		}
	}
	if last == -1 || len(c.messages[last].ToolCalls) == 0 {
		return
	}
	answered := map[string]bool{}
	for _, m := range c.messages[last+1:] {
		if m.Role == llm.RoleTool {
			answered[m.ToolCallID] = true
		}
	}
	for _, call := range c.messages[last].ToolCalls {
		if !answered[call.ID] {
			c.messages = append(c.messages, llm.Message{
				Role:       llm.RoleTool,
				Text:       note,
				ToolCallID: call.ID,
				ToolName:   call.Name,
			})
		}
	}
}

func (c *contextManager) setMode(mode types.AgentMode) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.mode = mode
}

func (c *contextManager) buildSystem() string {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.buildSystemLocked()
}

func (c *contextManager) buildSystemLocked() string {
	parts := []string{basePrompt, modePrompts[c.mode]}
	if c.projectInstructions != "" {
		parts = append(parts, "--- Project Instructions (COOLCODE.md) ---\n"+security.Redact(c.projectInstructions))
	}
	if c.skillsCatalog != "" {
		parts = append(parts, security.Redact(c.skillsCatalog))
	}

	var state strings.Builder
	state.WriteString("--- Project State ---\n")
	state.WriteString("CWD: " + c.rootDir + "\n")
	state.WriteString("File Tree:\n" + c.fileTree)
	if len(c.extraDirs) > 0 {
		state.WriteString("\n--- Additional Directories ---\n")
		state.WriteString("Read/write allowed; use absolute paths. Search tools cover only the primary root.\n")
		for i, dir := range c.extraDirs {
			state.WriteString(dir + "\n" + c.extraTrees[i])
		}
	}
	if len(c.pinned) > 0 {
		state.WriteString("\n--- Pinned Files ---\n")
		toolCtx := tools.Context{RootDir: c.rootDir, ExtraDirs: append([]string(nil), c.extraDirs...), Config: c.cfg, GitIgnore: c.gitIgnore}
		for _, p := range c.pinned {
			rel, _ := filepath.Rel(c.rootDir, p)
			if reason := tools.ValidateReadPath(p, toolCtx); reason != "" {
				state.WriteString("File: " + rel + " (blocked by guardrails)\n")
				continue
			}
			raw, err := os.ReadFile(p)
			if err != nil {
				state.WriteString("File: " + rel + " (error reading)\n")
				continue
			}
			state.WriteString("File: " + rel + "\n```\n" + security.Redact(string(raw)) + "\n```\n")
		}
	}
	parts = append(parts, state.String())

	if c.summary != "" {
		parts = append(parts, "--- Summary of Earlier Conversation ---\n"+security.Redact(c.summary))
	}
	return strings.Join(parts, "\n\n")
}

// window returns the trailing slice of messages that fits within the token
// budget, starting at a real user turn so tool-call/result pairs stay intact.
func (c *contextManager) window() []llm.Message {
	c.mu.Lock()
	defer c.mu.Unlock()
	budget := c.maxTokens - estimateTokens(c.buildSystemLocked())
	if budget < 2000 {
		budget = 2000
	}
	start := len(c.messages)
	used := 0
	for i := len(c.messages) - 1; i >= 0; i-- {
		t := estimateTokens(c.messages[i].Text)
		if used+t > budget {
			break
		}
		used += t
		start = i
	}
	// Advance to the first genuine user message so we never start mid-round.
	for start < len(c.messages) && c.messages[start].Role != llm.RoleUser {
		start++
	}
	if start >= len(c.messages) {
		// Fallback: last user message onward.
		for i := len(c.messages) - 1; i >= 0; i-- {
			if c.messages[i].Role == llm.RoleUser {
				return c.messages[i:]
			}
		}
		return c.messages
	}
	return c.messages[start:]
}

func (c *contextManager) refreshTree() {
	maxDepth := -1
	if c.cfg.Features.FileTreeMaxDepth != nil {
		maxDepth = *c.cfg.Features.FileTreeMaxDepth
	}
	tree := project.FolderStructure(c.rootDir, privacyChecker(c.rootDir, c.cfg, c.gitIgnore), maxDepth)
	c.mu.Lock()
	c.fileTree = tree
	c.mu.Unlock()
}

// extraDirTreeDepth caps additional-directory trees so a large extra dir
// cannot blow the prompt token budget.
const extraDirTreeDepth = 3

// addExtraDir registers an /add-dir root and renders its (depth-capped) tree
// for the system prompt. The tree is built outside the lock like refreshTree.
func (c *contextManager) addExtraDir(dir string) {
	checker := project.NewGitIgnoreChecker(dir)
	tree := project.FolderStructure(dir, privacyChecker(dir, c.cfg, checker), extraDirTreeDepth)
	c.mu.Lock()
	c.extraDirs = append(c.extraDirs, dir)
	c.extraTrees = append(c.extraTrees, tree)
	c.mu.Unlock()
}

func privacyChecker(root string, cfg config.Config, checker project.GitIgnoreChecker) project.GitIgnoreChecker {
	return func(rel string) bool {
		return (checker != nil && checker(rel)) || tools.BlockedPath(filepath.Join(root, rel), cfg) != ""
	}
}

// extraDirList returns a copy of the registered /add-dir roots.
func (c *contextManager) extraDirList() []string {
	c.mu.Lock()
	defer c.mu.Unlock()
	return append([]string(nil), c.extraDirs...)
}

func (c *contextManager) reloadSkills() {
	catalog := skills.Catalog(c.rootDir)
	c.mu.Lock()
	c.skillsCatalog = catalog
	c.mu.Unlock()
}

func (c *contextManager) applySummary(summary string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.summary = summary
	const keepRecent = 6
	if len(c.messages) > keepRecent {
		// Trim to a suffix beginning at a real user message.
		trimmed := c.messages[len(c.messages)-keepRecent:]
		for len(trimmed) > 0 && trimmed[0].Role != llm.RoleUser {
			trimmed = trimmed[1:]
		}
		if len(trimmed) > 0 {
			c.messages = trimmed
		}
	}
}

func (c *contextManager) stats() (messageCount, totalTokens int) {
	c.mu.Lock()
	defer c.mu.Unlock()
	for _, m := range c.messages {
		totalTokens += estimateTokens(m.Text)
	}
	return len(c.messages), totalTokens
}

// tree returns the current file tree text.
func (c *contextManager) tree() string {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.fileTree
}

// snapshotMessages returns a copy of the message history.
func (c *contextManager) snapshotMessages() []llm.Message {
	c.mu.Lock()
	defer c.mu.Unlock()
	return append([]llm.Message(nil), c.messages...)
}

// sessionState returns copies of the persistable state under one lock.
func (c *contextManager) sessionState() (messages []llm.Message, summary string, pinned []string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	return append([]llm.Message(nil), c.messages...), c.summary, append([]string(nil), c.pinned...)
}

func (c *contextManager) pin(path string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	for _, e := range c.pinned {
		if e == path {
			return
		}
	}
	c.pinned = append(c.pinned, path)
}

func (c *contextManager) unpin(path string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	out := c.pinned[:0]
	for _, e := range c.pinned {
		if e != path {
			out = append(out, e)
		}
	}
	c.pinned = out
}

func (c *contextManager) pinnedFiles() []string {
	c.mu.Lock()
	defer c.mu.Unlock()
	return append([]string{}, c.pinned...)
}

func (c *contextManager) restore(messages []llm.Message, summary string, pinned []string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if len(messages) > 0 {
		c.messages = messages
	}
	c.summary = summary
	if pinned != nil {
		c.pinned = pinned
	}
}
