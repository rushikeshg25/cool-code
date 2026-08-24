package llm

import (
	"context"
	"encoding/json"
	"strings"
)

type googleProvider struct{ baseProvider }

func (p *googleProvider) Name() string { return "google" }

type geminiPart struct {
	Text         string          `json:"text,omitempty"`
	FunctionCall *geminiFuncCall `json:"functionCall,omitempty"`
	FunctionResp *geminiFuncResp `json:"functionResponse,omitempty"`
}

type geminiFuncCall struct {
	Name string          `json:"name"`
	Args json.RawMessage `json:"args,omitempty"`
}

type geminiFuncResp struct {
	Name     string         `json:"name"`
	Response map[string]any `json:"response"`
}

type geminiContent struct {
	Role  string       `json:"role"`
	Parts []geminiPart `json:"parts"`
}

// buildBody converts a Request into the generateContent body.
func (p *googleProvider) buildBody(req Request) map[string]any {
	var contents []geminiContent
	for _, m := range req.Messages {
		switch m.Role {
		case RoleUser:
			contents = append(contents, geminiContent{Role: "user", Parts: []geminiPart{{Text: m.Text}}})
		case RoleAssistant:
			var parts []geminiPart
			if m.Text != "" {
				parts = append(parts, geminiPart{Text: m.Text})
			}
			for _, tc := range m.ToolCalls {
				args := tc.Arguments
				if len(args) == 0 {
					args = json.RawMessage("{}")
				}
				parts = append(parts, geminiPart{FunctionCall: &geminiFuncCall{Name: tc.Name, Args: args}})
			}
			contents = append(contents, geminiContent{Role: "model", Parts: parts})
		case RoleTool:
			part := geminiPart{FunctionResp: &geminiFuncResp{
				Name:     m.ToolName,
				Response: map[string]any{"result": m.Text},
			}}
			// Merge consecutive tool results into one user turn.
			if n := len(contents); n > 0 && contents[n-1].Role == "user" && isFuncRespTurn(contents[n-1]) {
				contents[n-1].Parts = append(contents[n-1].Parts, part)
			} else {
				contents = append(contents, geminiContent{Role: "user", Parts: []geminiPart{part}})
			}
		}
	}

	body := map[string]any{"contents": contents}
	if req.System != "" {
		body["systemInstruction"] = map[string]any{"parts": []map[string]any{{"text": req.System}}}
	}
	if len(req.Tools) > 0 {
		var decls []map[string]any
		for _, t := range req.Tools {
			decl := map[string]any{"name": t.Name, "description": t.Description}
			if params := convertSchemaForGemini(t.Parameters); hasProperties(params) {
				decl["parameters"] = params
			}
			decls = append(decls, decl)
		}
		body["tools"] = []map[string]any{{"functionDeclarations": decls}}
	}
	genConfig := map[string]any{}
	if req.Temperature != nil {
		genConfig["temperature"] = *req.Temperature
	}
	if p.maxTokens > 0 {
		genConfig["maxOutputTokens"] = p.maxTokens
	}
	if len(genConfig) > 0 {
		body["generationConfig"] = genConfig
	}
	return body
}

func (p *googleProvider) Complete(ctx context.Context, req Request) (Message, error) {
	body := p.buildBody(req)
	url := providerEndpoint("google", p.baseURL) + p.model + ":generateContent?key=" + p.apiKey

	var resp struct {
		Candidates []struct {
			Content geminiContent `json:"content"`
		} `json:"candidates"`
		UsageMetadata struct {
			PromptTokenCount     int `json:"promptTokenCount"`
			CandidatesTokenCount int `json:"candidatesTokenCount"`
		} `json:"usageMetadata"`
	}
	if err := postJSON(ctx, url, nil, body, &resp); err != nil {
		return Message{}, err
	}
	out := Message{Role: RoleAssistant, Usage: Usage{
		Input:  resp.UsageMetadata.PromptTokenCount,
		Output: resp.UsageMetadata.CandidatesTokenCount,
	}}
	if len(resp.Candidates) == 0 {
		return out, nil
	}
	for _, part := range resp.Candidates[0].Content.Parts {
		if part.Text != "" {
			out.Text += part.Text
		}
		if part.FunctionCall != nil {
			args := part.FunctionCall.Args
			if len(args) == 0 {
				args = json.RawMessage("{}")
			}
			out.ToolCalls = append(out.ToolCalls, ToolCall{
				ID:        part.FunctionCall.Name,
				Name:      part.FunctionCall.Name,
				Arguments: args,
			})
		}
	}
	return out, nil
}

// Stream implements Streamer via streamGenerateContent with alt=sse. Each
// chunk is a partial generateContent response.
func (p *googleProvider) Stream(ctx context.Context, req Request, onDelta func(string)) (Message, error) {
	body := p.buildBody(req)
	url := providerEndpoint("google", p.baseURL) + p.model + ":streamGenerateContent?alt=sse&key=" + p.apiKey

	out := Message{Role: RoleAssistant}
	err := streamSSE(ctx, url, nil, body, func(_ string, data []byte) {
		var chunk struct {
			Candidates []struct {
				Content geminiContent `json:"content"`
			} `json:"candidates"`
			UsageMetadata *struct {
				PromptTokenCount     int `json:"promptTokenCount"`
				CandidatesTokenCount int `json:"candidatesTokenCount"`
			} `json:"usageMetadata"`
		}
		if json.Unmarshal(data, &chunk) != nil {
			return
		}
		if chunk.UsageMetadata != nil {
			out.Usage = Usage{Input: chunk.UsageMetadata.PromptTokenCount, Output: chunk.UsageMetadata.CandidatesTokenCount}
		}
		if len(chunk.Candidates) == 0 {
			return
		}
		for _, part := range chunk.Candidates[0].Content.Parts {
			if part.Text != "" {
				out.Text += part.Text
				onDelta(part.Text)
			}
			if part.FunctionCall != nil {
				args := part.FunctionCall.Args
				if len(args) == 0 {
					args = json.RawMessage("{}")
				}
				out.ToolCalls = append(out.ToolCalls, ToolCall{
					ID:        part.FunctionCall.Name,
					Name:      part.FunctionCall.Name,
					Arguments: args,
				})
			}
		}
	})
	if err != nil {
		return Message{}, err
	}
	return out, nil
}

func isFuncRespTurn(c geminiContent) bool {
	for _, p := range c.Parts {
		if p.FunctionResp == nil {
			return false
		}
	}
	return len(c.Parts) > 0
}

func hasProperties(schema map[string]any) bool {
	props, ok := schema["properties"].(map[string]any)
	return ok && len(props) > 0
}

// convertSchemaForGemini uppercases JSON-schema "type" values (Gemini expects
// the OpenAPI Type enum, e.g. STRING/OBJECT) recursively.
func convertSchemaForGemini(schema map[string]any) map[string]any {
	if schema == nil {
		return nil
	}
	out := make(map[string]any, len(schema))
	for k, v := range schema {
		switch k {
		case "type":
			if s, ok := v.(string); ok {
				out[k] = strings.ToUpper(s)
			} else {
				out[k] = v
			}
		case "properties":
			if props, ok := v.(map[string]any); ok {
				conv := make(map[string]any, len(props))
				for name, p := range props {
					if pm, ok := p.(map[string]any); ok {
						conv[name] = convertSchemaForGemini(pm)
					} else {
						conv[name] = p
					}
				}
				out[k] = conv
			}
		case "items":
			if im, ok := v.(map[string]any); ok {
				out[k] = convertSchemaForGemini(im)
			} else {
				out[k] = v
			}
		default:
			out[k] = v
		}
	}
	return out
}
