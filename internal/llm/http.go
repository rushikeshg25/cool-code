package llm

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

var httpClient = &http.Client{Timeout: 5 * time.Minute}

// Provider endpoints, as vars so tests can point them at a local server.
var (
	anthropicURL  = "https://api.anthropic.com/v1/messages"
	openaiURL     = "https://api.openai.com/v1/chat/completions"
	googleBaseURL = "https://generativelanguage.googleapis.com/v1beta/models/"
)

// providerEndpoint turns a configured API base into the concrete endpoint
// used by a provider. Full endpoint URLs are accepted as well as conventional
// roots such as http://localhost:8317/v1.
func providerEndpoint(provider, base string) string {
	if base == "" {
		switch provider {
		case "anthropic":
			return anthropicURL
		case "openai":
			return openaiURL
		default:
			return googleBaseURL
		}
	}

	base = strings.TrimRight(base, "/")
	var suffix string
	switch provider {
	case "anthropic":
		if strings.HasSuffix(base, "/messages") {
			return base
		}
		suffix = "messages"
	case "openai":
		if strings.HasSuffix(base, "/chat/completions") {
			return base
		}
		suffix = "chat/completions"
	default:
		if strings.HasSuffix(base, "/models") {
			return base + "/"
		}
		suffix = "models/"
	}

	u, err := url.Parse(base + "/")
	if err != nil {
		return base + "/" + suffix
	}
	rel, err := url.Parse(suffix)
	if err != nil {
		return base + "/" + suffix
	}
	return u.ResolveReference(rel).String()
}

const maxAttempts = 3

// postJSON sends body as JSON to url with the given headers and decodes the
// response into out. Non-2xx responses return an error containing the body.
// Transient failures (transport errors, 429, 5xx) are retried with backoff.
func postJSON(ctx context.Context, url string, headers map[string]string, body any, out any) error {
	raw, err := json.Marshal(body)
	if err != nil {
		return err
	}
	var lastErr *httpError
	for attempt := 0; attempt < maxAttempts; attempt++ {
		if attempt > 0 {
			if err := sleepCtx(ctx, retryDelay(attempt, lastErr.retryAfter)); err != nil {
				return err
			}
		}
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(raw))
		if err != nil {
			return err
		}
		req.Header.Set("Content-Type", "application/json")
		for k, v := range headers {
			req.Header.Set(k, v)
		}
		resp, err := httpClient.Do(req)
		if err != nil {
			if ctx.Err() != nil {
				return err
			}
			lastErr = &httpError{err: err}
			continue
		}
		data, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			lastErr = &httpError{
				retryAfter: resp.Header.Get("Retry-After"),
				err:        fmt.Errorf("%s: HTTP %d: %s", url, resp.StatusCode, truncateErr(string(data))),
			}
			if !shouldRetry(resp.StatusCode) {
				return lastErr.err
			}
			continue
		}
		if err := json.Unmarshal(data, out); err != nil {
			return fmt.Errorf("decoding response: %w", err)
		}
		return nil
	}
	return lastErr.err
}

// httpError carries retry metadata for a failed attempt.
type httpError struct {
	retryAfter string
	err        error
}

func shouldRetry(status int) bool {
	return status == http.StatusTooManyRequests || status >= 500
}

var retryBaseDelay = time.Second

// retryDelay computes the backoff before the given attempt (1-based),
// honouring a parseable Retry-After header when present.
func retryDelay(attempt int, retryAfter string) time.Duration {
	if secs, err := strconv.Atoi(retryAfter); err == nil && secs > 0 {
		return time.Duration(secs) * time.Second
	}
	base := retryBaseDelay * time.Duration(1<<(attempt-1))
	return base + time.Duration(rand.Int63n(int64(base/2)))
}

func sleepCtx(ctx context.Context, d time.Duration) error {
	select {
	case <-time.After(d):
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// streamSSE posts body as JSON and parses the server-sent-event response,
// calling onEvent for every data line (event is "" when the server sends bare
// data lines, as OpenAI and Gemini do). Transient failures are retried only
// until the first event has been delivered.
func streamSSE(ctx context.Context, url string, headers map[string]string, body any, onEvent func(event string, data []byte)) error {
	raw, err := json.Marshal(body)
	if err != nil {
		return err
	}
	var lastErr *httpError
	for attempt := 0; attempt < maxAttempts; attempt++ {
		if attempt > 0 {
			if err := sleepCtx(ctx, retryDelay(attempt, lastErr.retryAfter)); err != nil {
				return err
			}
		}
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(raw))
		if err != nil {
			return err
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Accept", "text/event-stream")
		for k, v := range headers {
			req.Header.Set(k, v)
		}
		resp, err := httpClient.Do(req)
		if err != nil {
			if ctx.Err() != nil {
				return err
			}
			lastErr = &httpError{err: err}
			continue
		}
		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			data, _ := io.ReadAll(resp.Body)
			resp.Body.Close()
			lastErr = &httpError{
				retryAfter: resp.Header.Get("Retry-After"),
				err:        fmt.Errorf("%s: HTTP %d: %s", url, resp.StatusCode, truncateErr(string(data))),
			}
			if !shouldRetry(resp.StatusCode) {
				return lastErr.err
			}
			continue
		}
		err = scanSSE(resp.Body, onEvent)
		resp.Body.Close()
		return err
	}
	return lastErr.err
}

func scanSSE(r io.Reader, onEvent func(event string, data []byte)) error {
	scanner := bufio.NewScanner(r)
	scanner.Buffer(make([]byte, 0, 64*1024), 4*1024*1024)
	event := ""
	for scanner.Scan() {
		line := scanner.Text()
		switch {
		case strings.HasPrefix(line, "event:"):
			event = strings.TrimSpace(line[len("event:"):])
		case strings.HasPrefix(line, "data:"):
			data := strings.TrimSpace(line[len("data:"):])
			if data == "[DONE]" {
				return nil
			}
			onEvent(event, []byte(data))
		case line == "":
			event = ""
		}
	}
	return scanner.Err()
}

func truncateErr(s string) string {
	if len(s) > 800 {
		return s[:800] + "…"
	}
	return s
}
