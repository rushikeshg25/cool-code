// Package tools implements the agent's tool surface: filesystem, shell, git,
// search, and web operations, each exposed with a JSON-schema definition for
// native provider tool-calling.
package tools

import (
	"encoding/json"
	"strconv"

	"github.com/rushikeshg25/cool-code/internal/config"
	"github.com/rushikeshg25/cool-code/internal/project"
	"github.com/rushikeshg25/cool-code/internal/types"
)

// Context is the ambient state every tool receives.
type Context struct {
	RootDir   string
	Config    config.Config
	GitIgnore project.GitIgnoreChecker
}

// Tool is a single callable capability exposed to the model.
type Tool struct {
	Name        string
	Description string
	// Schema is a JSON-schema object ({"type":"object","properties":{…},"required":[…]}).
	Schema map[string]any
	// Mutating marks tools blocked in ask mode.
	Mutating bool
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
