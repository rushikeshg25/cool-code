package llm

import (
	"context"
	"encoding/json"
)

type openaiProvider struct{ baseProvider }

func (p *openaiProvider) Name() string { return "openai" }

type openaiToolCall struct {
	ID       string `json:"id"`
	Type     string `json:"type"`
	Function struct {
		Name      string `json:"name"`
		Arguments string `json:"arguments"`
	} `json:"function"`
}

type openaiMessage struct {
	Role       string           `json:"role"`
	Content    string           `json:"content,omitempty"`
	ToolCalls  []openaiToolCall `json:"tool_calls,omitempty"`
	ToolCallID string           `json:"tool_call_id,omitempty"`
}

func (p *openaiProvider) Complete(ctx context.Context, req Request) (Message, error) {
	var messages []openaiMessage
	if req.System != "" {
		messages = append(messages, openaiMessage{Role: "system", Content: req.System})
	}
	for _, m := range req.Messages {
		switch m.Role {
		case RoleUser:
			messages = append(messages, openaiMessage{Role: "user", Content: m.Text})
		case RoleAssistant:
			om := openaiMessage{Role: "assistant", Content: m.Text}
			for _, tc := range m.ToolCalls {
				args := string(tc.Arguments)
				if args == "" {
					args = "{}"
				}
				var otc openaiToolCall
				otc.ID = tc.ID
				otc.Type = "function"
				otc.Function.Name = tc.Name
				otc.Function.Arguments = args
				om.ToolCalls = append(om.ToolCalls, otc)
			}
			messages = append(messages, om)
		case RoleTool:
			messages = append(messages, openaiMessage{Role: "tool", ToolCallID: m.ToolCallID, Content: m.Text})
		}
	}

	var tools []map[string]any
	for _, t := range req.Tools {
		fn := map[string]any{
			"name":        t.Name,
			"description": t.Description,
			"parameters":  t.Parameters,
		}
		tools = append(tools, map[string]any{"type": "function", "function": fn})
	}

	body := map[string]any{
		"model":    p.model,
		"messages": messages,
	}
	if len(tools) > 0 {
		body["tools"] = tools
	}
	if req.Temperature != nil {
		body["temperature"] = *req.Temperature
	}
	if p.maxTokens > 0 {
		body["max_completion_tokens"] = p.maxTokens
	}

	headers := map[string]string{"Authorization": "Bearer " + p.apiKey}

	var resp struct {
		Choices []struct {
			Message struct {
				Content   string           `json:"content"`
				ToolCalls []openaiToolCall `json:"tool_calls"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := postJSON(ctx, "https://api.openai.com/v1/chat/completions", headers, body, &resp); err != nil {
		return Message{}, err
	}
	if len(resp.Choices) == 0 {
		return Message{Role: RoleAssistant}, nil
	}
	msg := resp.Choices[0].Message
	out := Message{Role: RoleAssistant, Text: msg.Content}
	for _, tc := range msg.ToolCalls {
		args := tc.Function.Arguments
		if args == "" {
			args = "{}"
		}
		out.ToolCalls = append(out.ToolCalls, ToolCall{
			ID:        tc.ID,
			Name:      tc.Function.Name,
			Arguments: json.RawMessage(args),
		})
	}
	return out, nil
}
