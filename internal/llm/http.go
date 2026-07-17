package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"strconv"
	"time"
)

var httpClient = &http.Client{Timeout: 5 * time.Minute}

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

func truncateErr(s string) string {
	if len(s) > 800 {
		return s[:800] + "…"
	}
	return s
}
