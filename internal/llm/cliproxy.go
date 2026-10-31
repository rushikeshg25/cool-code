package llm

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"sort"
	"strings"

	"github.com/rushikeshg25/cool-code/internal/config"
)

// ProxyModels reads the existing proxy's public model catalog. It never calls
// management, login, or configuration endpoints.
func ProxyModels(ctx context.Context, cfg config.LLM) ([]string, error) {
	if ResolveProvider(cfg) != "cliproxy" {
		return nil, errors.New("model discovery requires the cliproxy provider; select it with /connect")
	}
	provider, err := New(cfg)
	if err != nil {
		return nil, err
	}
	p := provider.(*openaiProvider)
	endpoint := strings.TrimSuffix(providerEndpoint("cliproxy", p.baseURL), "/chat/completions") + "/models"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, errors.New("invalid CLI proxy model endpoint")
	}
	req.Header.Set("Authorization", "Bearer "+p.apiKey)
	resp, err := httpClient.Do(req)
	if err != nil {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		return nil, errors.New("could not reach CLI proxy; check that its existing endpoint is available")
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(io.LimitReader(resp.Body, maxResponseBytes+1))
	if err != nil {
		return nil, errors.New("could not read CLI proxy model catalog")
	}
	if len(raw) > maxResponseBytes {
		return nil, errors.New("CLI proxy model catalog exceeds size limit")
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, httpStatusError(resp.StatusCode, raw)
	}
	var body struct {
		Data []struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	if json.Unmarshal(raw, &body) != nil {
		return nil, errors.New("invalid CLI proxy model catalog")
	}
	ids := make([]string, 0, len(body.Data))
	seen := map[string]bool{}
	for _, model := range body.Data {
		if model.ID != "" && !seen[model.ID] {
			ids = append(ids, model.ID)
			seen[model.ID] = true
		}
	}
	sort.Strings(ids)
	return ids, nil
}
