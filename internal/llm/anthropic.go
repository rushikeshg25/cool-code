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

// buildBody converts a Request into the Anthropic messages-API body + headers.
func (p *anthropicProvider) buildBody(req Request) (map[string]any, map[string]string) {
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
	return body, headers
}

func (p *anthropicProvider) Complete(ctx context.Context, req Request) (Message, error) {
	body, headers := p.buildBody(req)

	var resp struct {
		Content []anthropicContentBlock `json:"content"`
		Usage   struct {
			InputTokens  int `json:"input_tokens"`
			OutputTokens int `json:"output_tokens"`
		} `json:"usage"`
	}
	if err := postJSON(ctx, anthropicURL, headers, body, &resp); err != nil {
		return Message{}, err
	}

	out := Message{Role: RoleAssistant, Usage: Usage{Input: resp.Usage.InputTokens, Output: resp.Usage.OutputTokens}}
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

// Stream implements Streamer using the Anthropic SSE protocol.
func (p *anthropicProvider) Stream(ctx context.Context, req Request, onDelta func(string)) (Message, error) {
	body, headers := p.buildBody(req)
	body["stream"] = true

	out := Message{Role: RoleAssistant}
	type blockState struct {
		kind string // "text" or "tool_use"
		id   string
		name string
		args []byte
	}
	blocks := map[int]*blockState{}

	err := streamSSE(ctx, anthropicURL, headers, body, func(event string, data []byte) {
		var ev struct {
			Type  string `json:"type"`
			Index int    `json:"index"`
			// message_start
			Message struct {
				Usage struct {
					InputTokens int `json:"input_tokens"`
				} `json:"usage"`
			} `json:"message"`
			// content_block_start
			ContentBlock struct {
				Type string `json:"type"`
				ID   string `json:"id"`
				Name string `json:"name"`
			} `json:"content_block"`
			// content_block_delta
			Delta struct {
				Type        string `json:"type"`
				Text        string `json:"text"`
				PartialJSON string `json:"partial_json"`
			} `json:"delta"`
			// message_delta
			Usage struct {
				OutputTokens int `json:"output_tokens"`
			} `json:"usage"`
		}
		if json.Unmarshal(data, &ev) != nil {
			return
		}
		switch ev.Type {
		case "message_start":
			out.Usage.Input = ev.Message.Usage.InputTokens
		case "content_block_start":
			blocks[ev.Index] = &blockState{kind: ev.ContentBlock.Type, id: ev.ContentBlock.ID, name: ev.ContentBlock.Name}
		case "content_block_delta":
			b := blocks[ev.Index]
			if b == nil {
				return
			}
			switch ev.Delta.Type {
			case "text_delta":
				out.Text += ev.Delta.Text
				onDelta(ev.Delta.Text)
			case "input_json_delta":
				b.args = append(b.args, ev.Delta.PartialJSON...)
			}
		case "content_block_stop":
			b := blocks[ev.Index]
			if b != nil && b.kind == "tool_use" {
				args := json.RawMessage(b.args)
				if len(args) == 0 {
					args = json.RawMessage("{}")
				}
				out.ToolCalls = append(out.ToolCalls, ToolCall{ID: b.id, Name: b.name, Arguments: args})
			}
		case "message_delta":
			out.Usage.Output = ev.Usage.OutputTokens
		}
	})
	if err != nil {
		return Message{}, err
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
