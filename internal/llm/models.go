package llm

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"slices"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/rushikeshg25/cool-code/internal/config"
	"github.com/rushikeshg25/cool-code/internal/creds"
	"github.com/rushikeshg25/cool-code/internal/security"
)

// ModelInfo keeps routing identity separate from the provider's model ID.
type ModelInfo struct{ Provider, ID, Name string }
type ModelCatalog struct {
	Models   []ModelInfo
	Warnings []string
}

// ConfigForModel preserves an active connection only when its provider matches.
func ConfigForModel(current config.LLM, provider, id string) (config.LLM, error) {
	if provider != ResolveProvider(current) {
		// Generic environment overrides have no provider identity. Never forward
		// their credential to a different provider during discovery or selection.
		if os.Getenv("COOLCODE_API_KEY") != "" || os.Getenv("COOLCODE_API_BASE_URL") != "" {
			return config.LLM{}, errors.New("generic COOLCODE_API overrides are active; use provider-specific credentials to switch providers")
		}
		current.BaseURL = ""
		current.APIKeyEnv = ""
		current.AllowInsecureHTTP = nil
		current.ContextWindow = nil
	}
	current.Provider = provider
	current.Model = id
	return current, nil
}

// DiscoverModels queries configured providers concurrently. One failed catalog
// does not hide models from other providers. No inference requests are made.
func DiscoverModels(ctx context.Context, current config.LLM) ModelCatalog {
	ctx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()
	var out ModelCatalog
	var mu sync.Mutex
	var wg sync.WaitGroup
	for name, info := range providers {
		if name != ResolveProvider(current) && creds.APIKey(name) == "" && os.Getenv(info.envKey) == "" {
			continue
		}
		wg.Add(1)
		go func(name string) {
			defer wg.Done()
			cfg, err := ConfigForModel(current, name, current.Model)
			var models []ModelInfo
			if err == nil {
				models, err = ListModels(ctx, cfg)
			}
			mu.Lock()
			defer mu.Unlock()
			if err != nil {
				out.Warnings = append(out.Warnings, name+": "+security.SanitizeLine(security.Redact(err.Error())))
			} else {
				out.Models = append(out.Models, models...)
				if len(models) == 0 {
					out.Warnings = append(out.Warnings, name+": no text models returned")
				}
			}
		}(name)
	}
	wg.Wait()
	sort.Slice(out.Models, func(i, j int) bool {
		a, b := out.Models[i], out.Models[j]
		if a.Provider != b.Provider {
			return a.Provider < b.Provider
		}
		return a.ID < b.ID
	})
	sort.Strings(out.Warnings)
	return out
}

// ListModels uses the same endpoint validation and credential resolution as inference.
func ListModels(ctx context.Context, cfg config.LLM) ([]ModelInfo, error) {
	p, err := New(cfg)
	if err != nil {
		return nil, err
	}
	var base baseProvider
	switch p := p.(type) {
	case *openaiProvider:
		base = p.baseProvider
	case *googleProvider:
		base = p.baseProvider
	case *anthropicProvider:
		base = p.baseProvider
	}
	name := ResolveProvider(cfg)
	endpoint := strings.TrimRight(providerEndpoint(name, base.baseURL), "/")
	switch name {
	case "google":
	case "anthropic":
		endpoint = strings.TrimSuffix(endpoint, "/messages") + "/models"
	default:
		endpoint = strings.TrimSuffix(endpoint, "/chat/completions") + "/models"
	}
	if name == "openrouter" {
		endpoint += "/user"
	}
	var out []ModelInfo
	seen := map[string]bool{}
	pages := map[string]bool{}
	cursor := ""
	offset := 0
	for page := 0; page < 100; page++ {
		u, err := url.Parse(endpoint)
		if err != nil {
			return nil, errors.New("invalid model endpoint")
		}
		q := u.Query()
		switch name {
		case "google":
			q.Set("pageSize", "1000")
			if cursor != "" {
				q.Set("pageToken", cursor)
			}
		case "anthropic":
			q.Set("limit", "1000")
			if cursor != "" {
				q.Set("after_id", cursor)
			}
		case "openrouter":
			q.Set("output_modalities", "text")
			q.Set("limit", "1000")
			q.Set("offset", strconv.Itoa(offset))
		}
		u.RawQuery = q.Encode()
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
		if err != nil {
			return nil, errors.New("invalid model request")
		}
		switch name {
		case "google":
			req.Header.Set("x-goog-api-key", base.apiKey)
		case "anthropic":
			req.Header.Set("x-api-key", base.apiKey)
			req.Header.Set("anthropic-version", "2023-06-01")
		default:
			req.Header.Set("Authorization", "Bearer "+base.apiKey)
		}
		raw, err := readModelCatalog(req)
		if err != nil {
			return nil, err
		}
		var body struct {
			Data []struct {
				ID, Name     string
				DisplayName  string `json:"display_name"`
				Lifecycle    string
				Architecture struct {
					Output []string `json:"output_modalities"`
				}
				Endpoints []string
			} `json:"data"`
			Models []struct {
				Name, DisplayName string
				Methods           []string `json:"supportedGenerationMethods"`
			} `json:"models"`
			HasMore       bool   `json:"has_more"`
			LastID        string `json:"last_id"`
			NextPageToken string `json:"nextPageToken"`
		}
		if json.Unmarshal(raw, &body) != nil || (name == "google" && body.Models == nil) || (name != "google" && body.Data == nil) {
			return nil, errors.New("invalid model catalog JSON")
		}
		add := func(id, label string) {
			if id == "" || seen[id] || security.SanitizeLine(id) != id {
				return
			}
			seen[id] = true
			out = append(out, ModelInfo{Provider: name, ID: id, Name: security.SanitizeLine(label)})
		}
		for _, m := range body.Data {
			if m.Lifecycle == "retired" {
				continue
			}
			if len(m.Architecture.Output) > 0 && !slices.Contains(m.Architecture.Output, "text") {
				continue
			}
			if len(m.Endpoints) > 0 && name == "openrouter" && !slices.Contains(m.Endpoints, "/api/v1/chat/completions") {
				continue
			}
			if name == "cliproxy" && nonChatModelID(m.ID) {
				continue
			}
			if name == "openai" && !chatModelID(m.ID) {
				continue
			}
			if m.Name == "" {
				m.Name = m.DisplayName
			}
			add(m.ID, m.Name)
		}
		for _, m := range body.Models {
			if slices.Contains(m.Methods, "generateContent") && !nonChatModelID(m.Name) {
				add(strings.TrimPrefix(m.Name, "models/"), m.DisplayName)
			}
		}
		next := ""
		switch name {
		case "google":
			next = body.NextPageToken
		case "anthropic":
			if body.HasMore {
				next = body.LastID
				if next == "" {
					return nil, errors.New("model catalog omitted pagination cursor")
				}
			}
		case "openrouter":
			if len(body.Data) == 1000 {
				offset += 1000
				next = strconv.Itoa(offset)
			}
		}
		if next == "" {
			return out, nil
		}
		if pages[next] {
			return nil, errors.New("model catalog repeated pagination cursor")
		}
		pages[next] = true
		cursor = next
	}
	return nil, errors.New("model catalog pagination limit exceeded")
}

func readModelCatalog(req *http.Request) ([]byte, error) {
	resp, err := httpClient.Do(req)
	if err != nil {
		if req.Context().Err() != nil {
			return nil, req.Context().Err()
		}
		return nil, errors.New("could not reach model catalog")
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(io.LimitReader(resp.Body, maxResponseBytes+1))
	if err != nil {
		return nil, errors.New("could not read model catalog")
	}
	if len(raw) > maxResponseBytes {
		return nil, errors.New("model catalog exceeds size limit")
	}
	// Do not echo remote error bodies, which can contain credentials or URLs.
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("model catalog returned HTTP %d", resp.StatusCode)
	}
	var envelope struct{ Error json.RawMessage }
	if json.Unmarshal(raw, &envelope) == nil && len(envelope.Error) > 0 && string(envelope.Error) != "null" {
		return nil, errors.New("provider returned a model catalog error")
	}
	return raw, nil
}

// OpenAI's list API has no endpoint/capability metadata. Exclude known
// non-chat families; keep live IDs (including fine-tunes), never a static list.
func chatModelID(id string) bool {
	id = strings.ToLower(id)
	if nonChatModelID(id) || strings.Contains(id, "codex") || strings.HasSuffix(id, "-pro") || strings.Contains(id, "-pro-") {
		return false
	}
	return strings.HasPrefix(id, "gpt-") || strings.HasPrefix(id, "chatgpt-") || strings.HasPrefix(id, "o1") || strings.HasPrefix(id, "o3") || strings.HasPrefix(id, "o4") || strings.HasPrefix(id, "ft:")
}
func nonChatModelID(id string) bool {
	for _, kind := range []string{"embedding", "image", "audio", "tts", "transcrib", "realtime", "moderation", "deep-research"} {
		if strings.Contains(strings.ToLower(id), kind) {
			return true
		}
	}
	return false
}
