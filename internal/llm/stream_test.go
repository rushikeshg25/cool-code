package llm

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func collectSSE(t *testing.T, transcript string) (events []string, datas []string) {
	t.Helper()
	err := scanSSE(strings.NewReader(transcript), func(event string, data []byte) {
		events = append(events, event)
		datas = append(datas, string(data))
	})
	if err != nil {
		t.Fatal(err)
	}
	return events, datas
}

func TestScanSSE(t *testing.T) {
	transcript := "event: message_start\ndata: {\"a\":1}\n\ndata: {\"b\":2}\n\ndata: [DONE]\ndata: {\"never\":true}\n"
	events, datas := collectSSE(t, transcript)
	if len(datas) != 2 {
		t.Fatalf("datas = %v", datas)
	}
	if events[0] != "message_start" || datas[0] != `{"a":1}` {
		t.Fatalf("first event = %q %q", events[0], datas[0])
	}
	if events[1] != "" || datas[1] != `{"b":2}` {
		t.Fatalf("second event = %q %q (event name must reset after blank line)", events[1], datas[1])
	}
}

// runProviderStream serves transcript from a local SSE server, points the
// named provider's endpoint at it, and returns the assembled message.
func runProviderStream(t *testing.T, transcript, provider string) Message {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		_, _ = io.WriteString(w, transcript)
	}))
	defer srv.Close()

	base := baseProvider{model: "test-model", apiKey: "test-key", maxTokens: 4096}
	var s Streamer
	switch provider {
	case "anthropic":
		old := anthropicURL
		anthropicURL = srv.URL
		defer func() { anthropicURL = old }()
		s = &anthropicProvider{base}
	case "openai":
		old := openaiURL
		openaiURL = srv.URL
		defer func() { openaiURL = old }()
		s = &openaiProvider{base}
	case "google":
		old := googleBaseURL
		googleBaseURL = srv.URL + "/"
		defer func() { googleBaseURL = old }()
		s = &googleProvider{base}
	default:
		t.Fatalf("unknown provider %q", provider)
	}

	var streamed strings.Builder
	msg, err := s.Stream(context.Background(), Request{Messages: []Message{{Role: RoleUser, Text: "hi"}}},
		func(d string) { streamed.WriteString(d) })
	if err != nil {
		t.Fatal(err)
	}
	if streamed.String() != msg.Text {
		t.Fatalf("streamed deltas %q != assembled text %q", streamed.String(), msg.Text)
	}
	return msg
}

func TestAnthropicStreamAssembly(t *testing.T) {
	transcript := strings.Join([]string{
		`event: message_start`,
		`data: {"type":"message_start","message":{"usage":{"input_tokens":42}}}`,
		``,
		`event: content_block_start`,
		`data: {"type":"content_block_start","index":0,"content_block":{"type":"text"}}`,
		``,
		`event: content_block_delta`,
		`data: {"type":"content_block_delta","index":0,"delta":{"type":"text_delta","text":"Hello "}}`,
		``,
		`event: content_block_delta`,
		`data: {"type":"content_block_delta","index":0,"delta":{"type":"text_delta","text":"world"}}`,
		``,
		`event: content_block_stop`,
		`data: {"type":"content_block_stop","index":0}`,
		``,
		`event: content_block_start`,
		`data: {"type":"content_block_start","index":1,"content_block":{"type":"tool_use","id":"tu_1","name":"read_file"}}`,
		``,
		`event: content_block_delta`,
		`data: {"type":"content_block_delta","index":1,"delta":{"type":"input_json_delta","partial_json":"{\"absolutePath\":"}}`,
		``,
		`event: content_block_delta`,
		`data: {"type":"content_block_delta","index":1,"delta":{"type":"input_json_delta","partial_json":"\"/tmp/x\"}"}}`,
		``,
		`event: content_block_stop`,
		`data: {"type":"content_block_stop","index":1}`,
		``,
		`event: message_delta`,
		`data: {"type":"message_delta","usage":{"output_tokens":7}}`,
		``,
	}, "\n") + "\n"

	msg := runProviderStream(t, transcript, "anthropic")
	if msg.Text != "Hello world" {
		t.Fatalf("text = %q", msg.Text)
	}
	if len(msg.ToolCalls) != 1 || msg.ToolCalls[0].ID != "tu_1" || msg.ToolCalls[0].Name != "read_file" ||
		string(msg.ToolCalls[0].Arguments) != `{"absolutePath":"/tmp/x"}` {
		t.Fatalf("tool calls = %+v", msg.ToolCalls)
	}
	if msg.Usage.Input != 42 || msg.Usage.Output != 7 {
		t.Fatalf("usage = %+v", msg.Usage)
	}
}

func TestOpenAIStreamAssembly(t *testing.T) {
	transcript := strings.Join([]string{
		`data: {"choices":[{"delta":{"content":"Hi"}}]}`,
		``,
		`data: {"choices":[{"delta":{"content":" there"}}]}`,
		``,
		`data: {"choices":[{"delta":{"tool_calls":[{"index":0,"id":"call_1","function":{"name":"grep","arguments":"{\"pat"}}]}}]}`,
		``,
		`data: {"choices":[{"delta":{"tool_calls":[{"index":0,"function":{"arguments":"tern\":\"x\"}"}}]}}]}`,
		``,
		`data: {"choices":[],"usage":{"prompt_tokens":11,"completion_tokens":3}}`,
		``,
		`data: [DONE]`,
	}, "\n") + "\n"

	msg := runProviderStream(t, transcript, "openai")
	if msg.Text != "Hi there" {
		t.Fatalf("text = %q", msg.Text)
	}
	if len(msg.ToolCalls) != 1 || msg.ToolCalls[0].ID != "call_1" || msg.ToolCalls[0].Name != "grep" ||
		string(msg.ToolCalls[0].Arguments) != `{"pattern":"x"}` {
		t.Fatalf("tool calls = %+v", msg.ToolCalls)
	}
	if msg.Usage.Input != 11 || msg.Usage.Output != 3 {
		t.Fatalf("usage = %+v", msg.Usage)
	}
}

func TestGeminiStreamAssembly(t *testing.T) {
	transcript := strings.Join([]string{
		`data: {"candidates":[{"content":{"role":"model","parts":[{"text":"Once "}]}}]}`,
		``,
		`data: {"candidates":[{"content":{"role":"model","parts":[{"text":"upon"},{"functionCall":{"name":"glob","args":{"pattern":"*.go"}}}]}}],"usageMetadata":{"promptTokenCount":5,"candidatesTokenCount":2}}`,
		``,
	}, "\n") + "\n"

	msg := runProviderStream(t, transcript, "google")
	if msg.Text != "Once upon" {
		t.Fatalf("text = %q", msg.Text)
	}
	if len(msg.ToolCalls) != 1 || msg.ToolCalls[0].Name != "glob" ||
		string(msg.ToolCalls[0].Arguments) != `{"pattern":"*.go"}` {
		t.Fatalf("tool calls = %+v", msg.ToolCalls)
	}
	if msg.Usage.Input != 5 || msg.Usage.Output != 2 {
		t.Fatalf("usage = %+v", msg.Usage)
	}
}

func TestOpenAICompatibleProxyStream(t *testing.T) {
	var requestPath string
	var authorization string
	var streamed bool
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestPath = r.URL.Path
		authorization = r.Header.Get("Authorization")
		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Errorf("decode request: %v", err)
		}
		streamed, _ = body["stream"].(bool)
		w.Header().Set("Content-Type", "text/event-stream")
		_, _ = io.WriteString(w, "data: {\"choices\":[{\"delta\":{\"content\":\"proxy ok\"}}]}\n\ndata: [DONE]\n\n")
	}))
	defer srv.Close()

	provider := &openaiProvider{baseProvider{
		model:     "proxy-model",
		apiKey:    "proxy-key",
		baseURL:   srv.URL + "/v1",
		maxTokens: 32,
	}}
	var deltas strings.Builder
	msg, err := provider.Stream(context.Background(), Request{
		Messages: []Message{{Role: RoleUser, Text: "say proxy ok"}},
	}, func(delta string) { deltas.WriteString(delta) })
	if err != nil {
		t.Fatal(err)
	}
	if requestPath != "/v1/chat/completions" {
		t.Fatalf("request path = %q", requestPath)
	}
	if authorization != "Bearer proxy-key" {
		t.Fatalf("authorization header = %q", authorization)
	}
	if !streamed {
		t.Fatal("proxy request did not enable streaming")
	}
	if msg.Text != "proxy ok" || deltas.String() != "proxy ok" {
		t.Fatalf("proxy response = %q, deltas = %q", msg.Text, deltas.String())
	}
}
