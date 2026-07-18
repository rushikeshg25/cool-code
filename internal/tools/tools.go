// Package tools implements the agent's tool surface: filesystem, shell, git,
// search, and web operations, each exposed with a JSON-schema definition for
// native provider tool-calling.
package tools

import (
	"context"
	"encoding/json"
	"strconv"

	"github.com/rushikeshg25/cool-code/internal/config"
	"github.com/rushikeshg25/cool-code/internal/project"
	"github.com/rushikeshg25/cool-code/internal/types"
)

// Context is the ambient state every tool receives.
type Context struct {
	// Ctx cancels long-running tool work (shell commands, web requests) when
	// the turn is aborted. May be nil; use Context() for a non-nil value.
	Ctx     context.Context
	RootDir string
	// ExtraDirs are additional directories (added via /add-dir) that path-jailed
	// tools may read and write alongside RootDir.
	ExtraDirs []string
	Config    config.Config
	GitIgnore project.GitIgnoreChecker
}

// Roots returns every directory tools may operate in: RootDir plus ExtraDirs.
func (c Context) Roots() []string {
	return append([]string{c.RootDir}, c.ExtraDirs...)
}

// Context returns the cancellation context, defaulting to context.Background().
func (c Context) Context() context.Context {
	if c.Ctx != nil {
		return c.Ctx
	}
	return context.Background()
}

// Tool is a single callable capability exposed to the model.
type Tool struct {
	Name        string
	Description string
	// Schema is a JSON-schema object ({"type":"object","properties":{…},"required":[…]}).
	Schema map[string]any
	// Mutating marks tools blocked in ask mode.
	Mutating bool
	// ReadOnly marks side-effect-free tools that may run concurrently and are
	// available to explore subagents.
	ReadOnly bool
	// Execute runs the tool with raw JSON arguments and the ambient context.
	Execute func(ctx Context, args json.RawMessage) types.ToolResult
}

func obj(props map[string]any, required ...string) map[string]any {
	schema := map[string]any{"type": "object", "properties": props}
	if required == nil {
		required = []string{}
	}
	schema["required"] = required
	return schema
}

func strProp(desc string) map[string]any {
	return map[string]any{"type": "string", "description": desc}
}
func numProp(desc string) map[string]any {
	return map[string]any{"type": "number", "description": desc}
}
func boolProp(desc string) map[string]any {
	return map[string]any{"type": "boolean", "description": desc}
}
func arrProp(desc string) map[string]any {
	return map[string]any{"type": "array", "items": map[string]any{"type": "string"}, "description": desc}
}

func fail(display, msg string) types.ToolResult {
	return types.ToolResult{Display: display, LLMResult: msg}
}

func itoa(n int) string { return strconv.Itoa(n) }
