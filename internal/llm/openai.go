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

// buildBody converts a Request into the chat-completions body + headers.
func (p *openaiProvider) buildBody(req Request) (map[string]any, map[string]string) {
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
	return body, headers
}

func (p *openaiProvider) Complete(ctx context.Context, req Request) (Message, error) {
	body, headers := p.buildBody(req)

	var resp struct {
		Choices []struct {
			Message struct {
				Content   string           `json:"content"`
				ToolCalls []openaiToolCall `json:"tool_calls"`
			} `json:"message"`
		} `json:"choices"`
		Usage struct {
			PromptTokens     int `json:"prompt_tokens"`
			CompletionTokens int `json:"completion_tokens"`
		} `json:"usage"`
	}
	if err := postJSON(ctx, openaiURL, headers, body, &resp); err != nil {
		return Message{}, err
	}
	usage := Usage{Input: resp.Usage.PromptTokens, Output: resp.Usage.CompletionTokens}
	if len(resp.Choices) == 0 {
		return Message{Role: RoleAssistant, Usage: usage}, nil
	}
	msg := resp.Choices[0].Message
	out := Message{Role: RoleAssistant, Text: msg.Content, Usage: usage}
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

// Stream implements Streamer over chat-completions SSE chunks. Tool-call
// fragments arrive keyed by index (id/name once, arguments concatenated).
func (p *openaiProvider) Stream(ctx context.Context, req Request, onDelta func(string)) (Message, error) {
	body, headers := p.buildBody(req)
	body["stream"] = true
	body["stream_options"] = map[string]any{"include_usage": true}

	out := Message{Role: RoleAssistant}
	type callState struct {
		id   string
		name string
		args []byte
	}
	calls := map[int]*callState{}
	var order []int

	err := streamSSE(ctx, openaiURL, headers, body, func(_ string, data []byte) {
		var chunk struct {
			Choices []struct {
				Delta struct {
					Content   string `json:"content"`
					ToolCalls []struct {
						Index    int    `json:"index"`
						ID       string `json:"id"`
						Function struct {
							Name      string `json:"name"`
							Arguments string `json:"arguments"`
						} `json:"function"`
					} `json:"tool_calls"`
				} `json:"delta"`
			} `json:"choices"`
			Usage *struct {
				PromptTokens     int `json:"prompt_tokens"`
				CompletionTokens int `json:"completion_tokens"`
			} `json:"usage"`
		}
		if json.Unmarshal(data, &chunk) != nil {
			return
		}
		if chunk.Usage != nil {
			out.Usage = Usage{Input: chunk.Usage.PromptTokens, Output: chunk.Usage.CompletionTokens}
		}
		if len(chunk.Choices) == 0 {
			return
		}
		delta := chunk.Choices[0].Delta
		if delta.Content != "" {
			out.Text += delta.Content
			onDelta(delta.Content)
		}
		for _, tc := range delta.ToolCalls {
			c := calls[tc.Index]
			if c == nil {
				c = &callState{}
				calls[tc.Index] = c
				order = append(order, tc.Index)
			}
			if tc.ID != "" {
				c.id = tc.ID
			}
			if tc.Function.Name != "" {
				c.name = tc.Function.Name
			}
			c.args = append(c.args, tc.Function.Arguments...)
		}
	})
	if err != nil {
		return Message{}, err
	}
	for _, idx := range order {
		c := calls[idx]
		args := c.args
		if len(args) == 0 {
			args = []byte("{}")
		}
		out.ToolCalls = append(out.ToolCalls, ToolCall{ID: c.id, Name: c.name, Arguments: json.RawMessage(args)})
	}
	return out, nil
}
