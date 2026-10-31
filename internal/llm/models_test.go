package llm

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/rushikeshg25/cool-code/internal/config"
	"github.com/rushikeshg25/cool-code/internal/creds"
)

func TestDiscoverModelsKeepsProvidersAndPartialResults(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	for _, key := range []string{"COOLCODE_API_KEY", "COOLCODE_API_BASE_URL", "OPENAI_API_KEY", "ANTHROPIC_API_KEY", "GOOGLE_GENERATIVE_AI_API_KEY", "OPENROUTER_API_KEY", "OPENAI_BASE_URL", "OPENROUTER_BASE_URL"} {
		t.Setenv(key, "")
	}
	if err := creds.SetAPIKey("openai", "test-openai"); err != nil {
		t.Fatal(err)
	}
	if err := creds.SetAPIKey("openrouter", "test-router"); err != nil {
		t.Fatal(err)
	}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/openai/models":
			if r.Header.Get("Authorization") != "Bearer test-openai" {
				t.Error("wrong OpenAI key")
			}
			fmt.Fprint(w, `{"data":[{"id":"gpt-test"},{"id":"gpt-image-1"},{"id":"gpt-test"},{"id":"text-embedding-3-small"}]}`)
		case "/router/models/user":
			if r.Header.Get("Authorization") != "Bearer test-router" {
				t.Error("wrong router key")
			}
			fmt.Fprint(w, `{"data":[{"id":"openai/gpt-test","name":"Test model","architecture":{"output_modalities":["text"]}},{"id":"image/model","architecture":{"output_modalities":["image"]}}]}`)
		default:
			w.WriteHeader(503)
		}
	}))
	defer srv.Close()
	oldAI, oldRouter, oldClaude := openaiURL, openrouterURL, anthropicURL
	openaiURL = srv.URL + "/openai/chat/completions"
	openrouterURL = srv.URL + "/router/chat/completions"
	anthropicURL = srv.URL + "/claude/messages"
	t.Cleanup(func() { openaiURL = oldAI; openrouterURL = oldRouter; anthropicURL = oldClaude })
	if err := creds.SetAPIKey("anthropic", "test-claude"); err != nil {
		t.Fatal(err)
	}
	got := DiscoverModels(context.Background(), config.LLM{Provider: "openai", Model: "gpt-test"})
	if len(got.Models) != 2 || got.Models[0].Provider != "openai" || got.Models[1].Provider != "openrouter" {
		t.Fatalf("models=%+v", got.Models)
	}
	if len(got.Warnings) != 1 || !strings.Contains(got.Warnings[0], "503") {
		t.Fatalf("warnings=%v", got.Warnings)
	}
}

func TestModelCatalogPagination(t *testing.T) {
	for _, provider := range []string{"google", "anthropic"} {
		t.Run(provider, func(t *testing.T) {
			t.Setenv("CATALOG_TEST_KEY", "secret")
			calls := 0
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				calls++
				if provider == "google" {
					if r.Header.Get("x-goog-api-key") != "secret" || r.URL.Query().Get("key") != "" {
						t.Error("wrong Google authentication")
					}
					if calls == 1 {
						fmt.Fprint(w, `{"models":[{"name":"models/gemini-pro","supportedGenerationMethods":["generateContent"]},{"name":"models/embed","supportedGenerationMethods":["embedContent"]}],"nextPageToken":"two"}`)
					} else {
						if r.URL.Query().Get("pageToken") != "two" {
							t.Error("missing next page")
						}
						fmt.Fprint(w, `{"models":[{"name":"models/gemini-flash","supportedGenerationMethods":["generateContent"]}]}`)
					}
				} else {
					if r.Header.Get("x-api-key") != "secret" || r.Header.Get("anthropic-version") == "" {
						t.Error("wrong Anthropic authentication")
					}
					if calls == 1 {
						fmt.Fprint(w, `{"data":[{"id":"claude-one"}],"has_more":true,"last_id":"claude-one"}`)
					} else {
						if r.URL.Query().Get("after_id") != "claude-one" {
							t.Error("missing next page")
						}
						fmt.Fprint(w, `{"data":[{"id":"claude-two"},{"id":"claude-retired","lifecycle":"retired"}]}`)
					}
				}
			}))
			defer srv.Close()
			models, err := ListModels(context.Background(), config.LLM{Provider: provider, Model: "test", BaseURL: srv.URL + "/v1", APIKeyEnv: "CATALOG_TEST_KEY"})
			if err != nil || len(models) != 2 || calls != 2 {
				t.Fatalf("models=%v calls=%d err=%v", models, calls, err)
			}
		})
	}
}

func TestModelDiscoveryRejectsRedirectsAndMalformedCatalogs(t *testing.T) {
	t.Setenv("CATALOG_TEST_KEY", "secret")
	for _, body := range []string{`{}`, `{"error":{"message":"secret"}}`, `not json`, `redirect`} {
		t.Run(body, func(t *testing.T) {
			hits := 0
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				hits++
				if body == "redirect" {
					http.Redirect(w, r, "/elsewhere", 302)
				} else {
					fmt.Fprint(w, body)
				}
			}))
			defer srv.Close()
			_, err := ListModels(context.Background(), config.LLM{Provider: "openai", Model: "gpt-test", BaseURL: srv.URL, APIKeyEnv: "CATALOG_TEST_KEY"})
			if err == nil || strings.Contains(err.Error(), "secret") || hits != 1 {
				t.Fatalf("hits=%d err=%v", hits, err)
			}
		})
	}
}

func TestModelSwitchDoesNotCarryConnectionOverrides(t *testing.T) {
	t.Setenv("COOLCODE_API_KEY", "")
	t.Setenv("COOLCODE_API_BASE_URL", "")
	cfg := config.LLM{Provider: "openai", BaseURL: "https://example.invalid/v1", APIKeyEnv: "CUSTOM_KEY", Model: "gpt-test"}
	got, err := ConfigForModel(cfg, "openrouter", "openai/gpt-test")
	if err != nil || got.BaseURL != "" || got.APIKeyEnv != "" || got.Provider != "openrouter" {
		t.Fatalf("cfg=%+v err=%v", got, err)
	}
	same, err := ConfigForModel(cfg, "openai", "gpt-next")
	if err != nil || same.BaseURL != cfg.BaseURL {
		t.Fatal("active connection lost")
	}
	t.Setenv("COOLCODE_API_KEY", "secret")
	if _, err := ConfigForModel(cfg, "openrouter", "openai/gpt-test"); err == nil {
		t.Fatal("generic credential must not cross providers")
	}
}
