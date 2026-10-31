package llm

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/rushikeshg25/cool-code/internal/config"
	"github.com/rushikeshg25/cool-code/internal/creds"
)

func TestCLIProxyCredentialBinding(t *testing.T) {
	isolateProviderEnv(t)
	t.Setenv("CLIPROXY_API_KEY", "")
	t.Setenv("CLIPROXY_BASE_URL", "")
	t.Setenv("OPENAI_API_KEY", "first-party-key")
	cfg := config.LLM{Provider: "cliproxy", Model: "gpt-6.1-sol"}
	_, err := New(cfg)
	var missing *MissingKeyError
	if !errors.As(err, &missing) || missing.EnvKey != "CLIPROXY_API_KEY" {
		t.Fatalf("expected proxy-specific missing key: %v", err)
	}
	if err := creds.SetAPIKey("cliproxy", "proxy-key"); err != nil {
		t.Fatal(err)
	}
	p, err := New(cfg)
	if err != nil || p.Name() != "cliproxy" {
		t.Fatalf("provider = %v, error = %v", p, err)
	}
	if got := providerEndpoint(p.Name(), ""); got != "http://127.0.0.1:8317/v1/chat/completions" {
		t.Fatalf("endpoint = %q", got)
	}
	cfg.Provider = "openai-subscription"
	if p, err := New(cfg); err != nil || p.Name() != "cliproxy" {
		t.Fatalf("legacy alias: %v", err)
	}
	cfg.BaseURL = "https://different.example/v1"
	if _, err := New(cfg); !errors.As(err, &missing) {
		t.Fatal("stored proxy key must not go to an endpoint override")
	}
	t.Setenv("TEST_PROXY_API_KEY", "explicit-proxy-key")
	cfg.APIKeyEnv = "TEST_PROXY_API_KEY"
	if _, err := New(cfg); err != nil {
		t.Fatal(err)
	}
	cfg.BaseURL = "http://remote.example/v1"
	if _, err := New(cfg); err == nil {
		t.Fatal("remote plaintext endpoint accepted")
	}
}

func TestCLIProxyModelCatalog(t *testing.T) {
	isolateProviderEnv(t)
	t.Setenv("CLIPROXY_BASE_URL", "")
	t.Setenv("TEST_PROXY_API_KEY", "proxy-key")
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" || r.URL.Path != "/v1/models" || r.Header.Get("Authorization") != "Bearer proxy-key" {
			t.Error("incorrect model request")
		}
		fmt.Fprint(w, `{"data":[{"id":"gpt-b"},{"id":"gpt-a"},{"id":"gpt-b"}]}`)
	}))
	defer srv.Close()
	ids, err := ProxyModels(context.Background(), config.LLM{Provider: "cliproxy", BaseURL: srv.URL + "/v1/chat/completions", APIKeyEnv: "TEST_PROXY_API_KEY"})
	if err != nil || strings.Join(ids, ",") != "gpt-a,gpt-b" {
		t.Fatalf("catalog = %v, error = %v", ids, err)
	}
}

func TestCLIProxyModelCatalogRejectsRedirects(t *testing.T) {
	isolateProviderEnv(t)
	t.Setenv("TEST_PROXY_API_KEY", "proxy-key")
	forwarded := false
	target := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { forwarded = true }))
	defer target.Close()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { http.Redirect(w, r, target.URL, http.StatusFound) }))
	defer srv.Close()
	_, err := ProxyModels(context.Background(), config.LLM{Provider: "cliproxy", BaseURL: srv.URL, APIKeyEnv: "TEST_PROXY_API_KEY"})
	if err == nil || forwarded {
		t.Fatal("model discovery followed a redirect")
	}
}

func TestCLIProxyStreamingToolCalls(t *testing.T) {
	isolateProviderEnv(t)
	t.Setenv("TEST_PROXY_API_KEY", "proxy-key")
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/chat/completions" || r.Header.Get("Authorization") != "Bearer proxy-key" {
			t.Error("incorrect proxy request")
		}
		w.Header().Set("Content-Type", "text/event-stream")
		fmt.Fprint(w, "data: {\"choices\":[{\"delta\":{\"tool_calls\":[{\"index\":0,\"id\":\"call_1\",\"function\":{\"name\":\"read_file\",\"arguments\":\"{}\"}}]}}]}\n\ndata: {\"choices\":[],\"usage\":{\"prompt_tokens\":11,\"completion_tokens\":2}}\n\ndata: [DONE]\n\n")
	}))
	defer srv.Close()
	p, err := New(config.LLM{Provider: "cliproxy", Model: "gpt-6.1-sol", BaseURL: srv.URL + "/v1", APIKeyEnv: "TEST_PROXY_API_KEY"})
	if err != nil {
		t.Fatal(err)
	}
	msg, err := p.(Streamer).Stream(context.Background(), Request{Messages: []Message{{Role: RoleUser, Text: "read"}}}, func(string) {})
	if err != nil || len(msg.ToolCalls) != 1 || msg.ToolCalls[0].Name != "read_file" || msg.Usage.Input != 11 {
		t.Fatalf("message = %+v, error = %v", msg, err)
	}
}
