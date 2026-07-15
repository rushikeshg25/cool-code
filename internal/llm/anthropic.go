package llm

import (
	"context"
	"encoding/json"
)

type anthropicProvider struct{ baseProvider }

func (p *anthropicProvider) Name() string { return "anthropic" }

type anthropicContentBlock struct {
	Type string `json:"type"`
	// text block
	Text string `json:"text,omitempty"`
	// tool_use block
	ID    string          `json:"id,omitempty"`
	Name  string          `json:"name,omitempty"`
	Input json.RawMessage `json:"input,omitempty"`
	// tool_result block
	ToolUseID string `json:"tool_use_id,omitempty"`
	Content   string `json:"content,omitempty"`
}

type anthropicMessage struct {
	Role    string                  `json:"role"`
	Content []anthropicContentBlock `json:"content"`
}

func (p *anthropicProvider) Complete(ctx context.Context, req Request) (Message, error) {
	var messages []anthropicMessage
	for _, m := range req.Messages {
		switch m.Role {
		case RoleUser:
			messages = append(messages, anthropicMessage{
				Role:    "user",
				Content: []anthropicContentBlock{{Type: "text", Text: m.Text}},
			})
		case RoleAssistant:
			var blocks []anthropicContentBlock
			if m.Text != "" {
				blocks = append(blocks, anthropicContentBlock{Type: "text", Text: m.Text})
			}
			for _, tc := range m.ToolCalls {
				input := tc.Arguments
				if len(input) == 0 {
					input = json.RawMessage("{}")
				}
				blocks = append(blocks, anthropicContentBlock{Type: "tool_use", ID: tc.ID, Name: tc.Name, Input: input})
			}
			messages = append(messages, anthropicMessage{Role: "assistant", Content: blocks})
		case RoleTool:
			block := anthropicContentBlock{Type: "tool_result", ToolUseID: m.ToolCallID, Content: m.Text}
			// Merge consecutive tool results into the trailing user message.
			if n := len(messages); n > 0 && messages[n-1].Role == "user" && isToolResultTurn(messages[n-1]) {
				messages[n-1].Content = append(messages[n-1].Content, block)
			} else {
				messages = append(messages, anthropicMessage{Role: "user", Content: []anthropicContentBlock{block}})
			}
		}
	}

	var tools []map[string]any
	for _, t := range req.Tools {
		tools = append(tools, map[string]any{
			"name":         t.Name,
			"description":  t.Description,
			"input_schema": t.Parameters,
		})
	}

	body := map[string]any{
		"model":      p.model,
		"max_tokens": p.maxTokens,
		"messages":   messages,
	}
	if req.System != "" {
		body["system"] = req.System
	}
	if len(tools) > 0 {
		body["tools"] = tools
	}
	if req.Temperature != nil {
		body["temperature"] = *req.Temperature
	}

	headers := map[string]string{
		"x-api-key":         p.apiKey,
		"anthropic-version": "2023-06-01",
	}

	var resp struct {
		Content []anthropicContentBlock `json:"content"`
	}
	if err := postJSON(ctx, "https://api.anthropic.com/v1/messages", headers, body, &resp); err != nil {
		return Message{}, err
	}

	out := Message{Role: RoleAssistant}
	for _, block := range resp.Content {
		switch block.Type {
		case "text":
			out.Text += block.Text
		case "tool_use":
			args := block.Input
			if len(args) == 0 {
				args = json.RawMessage("{}")
			}
			out.ToolCalls = append(out.ToolCalls, ToolCall{ID: block.ID, Name: block.Name, Arguments: args})
		}
	}
	return out, nil
}

func isToolResultTurn(m anthropicMessage) bool {
	for _, b := range m.Content {
		if b.Type != "tool_result" {
			return false
		}
	}
	return len(m.Content) > 0
}
