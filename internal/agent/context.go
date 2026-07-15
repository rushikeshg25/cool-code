package agent

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/rushikeshg25/cool-code/internal/config"
	"github.com/rushikeshg25/cool-code/internal/llm"
	"github.com/rushikeshg25/cool-code/internal/memory"
	"github.com/rushikeshg25/cool-code/internal/project"
	"github.com/rushikeshg25/cool-code/internal/skills"
	"github.com/rushikeshg25/cool-code/internal/tools"
	"github.com/rushikeshg25/cool-code/internal/types"
)

// contextManager holds the conversation and builds the system prompt + windowed
// message list for each turn.
type contextManager struct {
	rootDir             string
	gitIgnore           project.GitIgnoreChecker
	cfg                 config.Config
	mode                types.AgentMode
	messages            []llm.Message
	summary             string
	maxTokens           int
	fileTree            string
	projectInstructions string
	skillsCatalog       string
	pinned              []string
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
		fileTree:            project.FolderStructure(rootDir, checker, maxDepth),
		projectInstructions: memory.LoadProjectInstructions(rootDir),
		skillsCatalog:       skills.Catalog(rootDir),
	}
}

func estimateTokens(s string) int { return (len(s) + 3) / 4 }

func (c *contextManager) addUser(text string) {
	c.messages = append(c.messages, llm.Message{Role: llm.RoleUser, Text: text})
}

func (c *contextManager) addAssistant(m llm.Message) {
	c.messages = append(c.messages, m)
}

func (c *contextManager) addToolResult(call llm.ToolCall, result string) {
	c.messages = append(c.messages, llm.Message{
		Role:       llm.RoleTool,
		Text:       result,
		ToolCallID: call.ID,
		ToolName:   call.Name,
	})
}

func (c *contextManager) buildSystem() string {
	parts := []string{basePrompt, modePrompts[c.mode]}
	if c.projectInstructions != "" {
		parts = append(parts, "--- Project Instructions (COOLCODE.md) ---\n"+c.projectInstructions)
	}
	if c.skillsCatalog != "" {
		parts = append(parts, c.skillsCatalog)
	}

	var state strings.Builder
	state.WriteString("--- Project State ---\n")
	state.WriteString("CWD: " + c.rootDir + "\n")
	state.WriteString("File Tree:\n" + c.fileTree)
	if len(c.pinned) > 0 {
		state.WriteString("\n--- Pinned Files ---\n")
		for _, p := range c.pinned {
			rel, _ := filepath.Rel(c.rootDir, p)
			if reason := tools.BlockedPath(p, c.cfg); reason != "" {
				state.WriteString("File: " + rel + " (blocked by guardrails)\n")
				continue
			}
			raw, err := os.ReadFile(p)
			if err != nil {
				state.WriteString("File: " + rel + " (error reading)\n")
				continue
			}
			state.WriteString("File: " + rel + "\n```\n" + string(raw) + "\n```\n")
		}
	}
	parts = append(parts, state.String())

	if c.summary != "" {
		parts = append(parts, "--- Summary of Earlier Conversation ---\n"+c.summary)
	}
	return strings.Join(parts, "\n\n")
}

// window returns the trailing slice of messages that fits within the token
// budget, starting at a real user turn so tool-call/result pairs stay intact.
func (c *contextManager) window() []llm.Message {
	budget := c.maxTokens - estimateTokens(c.buildSystem())
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
	c.fileTree = project.FolderStructure(c.rootDir, c.gitIgnore, maxDepth)
}

func (c *contextManager) reloadSkills() {
	c.skillsCatalog = skills.Catalog(c.rootDir)
}

func (c *contextManager) applySummary(summary string) {
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
	for _, m := range c.messages {
		totalTokens += estimateTokens(m.Text)
	}
	return len(c.messages), totalTokens
}
