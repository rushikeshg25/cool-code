package llm

import (
	"context"
	"fmt"
	"github.com/rushikeshg25/cool-code/internal/config"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestProxyUsageDetails(t *testing.T) {
	isolateProviderEnv(t)
	t.Setenv("TEST_PROXY_API_KEY", "test")
	for _, stream := range []bool{false, true} {
		for _, tc := range []struct {
			name, usage string
			known       bool
			input       int
		}{
			{"absent", "null", false, 0}, {"empty", "{}", false, 0},
			{"zero", `{"prompt_tokens":0,"completion_tokens":0}`, true, 0},
			{"details", `{"prompt_tokens":11,"completion_tokens":7,"prompt_tokens_details":{"cached_tokens":3},"completion_tokens_details":{"reasoning_tokens":2}}`, true, 11},
		} {
			t.Run(fmt.Sprintf("%s/stream=%v", tc.name, stream), func(t *testing.T) {
				srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					if stream {
						w.Header().Set("Content-Type", "text/event-stream")
						fmt.Fprintf(w, "data: {\"choices\":[],\"usage\":%s}\n\ndata: [DONE]\n\n", tc.usage)
					} else {
						fmt.Fprintf(w, `{"choices":[{"message":{"content":"OK"}}],"usage":%s}`, tc.usage)
					}
				}))
				defer srv.Close()
				p, err := New(config.LLM{Provider: "cliproxy", Model: "test", BaseURL: srv.URL, APIKeyEnv: "TEST_PROXY_API_KEY"})
				if err != nil {
					t.Fatal(err)
				}
				var m Message
				if stream {
					m, err = p.(Streamer).Stream(context.Background(), Request{}, func(string) {})
				} else {
					m, err = p.Complete(context.Background(), Request{})
				}
				if err != nil || m.Usage.Reported != tc.known || m.Usage.Input != tc.input {
					t.Fatalf("usage=%+v err=%v", m.Usage, err)
				}
				if tc.name == "details" && (m.Usage.CachedInput == nil || *m.Usage.CachedInput != 3 || m.Usage.Reasoning == nil || *m.Usage.Reasoning != 2) {
					t.Fatalf("details=%+v", m.Usage)
				}
			})
		}
	}
}

func TestProxyExposedReasoning(t *testing.T) {
	isolateProviderEnv(t)
	t.Setenv("TEST_PROXY_API_KEY", "test")
	for _, stream := range []bool{false, true} {
		t.Run(fmt.Sprint(stream), func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if stream {
					w.Header().Set("Content-Type", "text/event-stream")
					fmt.Fprint(w, "data: {\"choices\":[{\"delta\":{\"content\":\"answer\",\"reasoning_content\":\"exposed reasoning\"}}]}\n\ndata: [DONE]\n\n")
				} else {
					fmt.Fprint(w, `{"choices":[{"message":{"content":"answer","reasoning_content":"exposed reasoning"}}]}`)
				}
			}))
			defer srv.Close()
			p, err := New(config.LLM{Provider: "cliproxy", Model: "test", BaseURL: srv.URL, APIKeyEnv: "TEST_PROXY_API_KEY"})
			if err != nil {
				t.Fatal(err)
			}
			var m Message
			var visible string
			if stream {
				m, err = p.(Streamer).Stream(context.Background(), Request{}, func(s string) { visible += s })
			} else {
				m, err = p.Complete(context.Background(), Request{})
			}
			if err != nil || m.Text != "answer" || m.Reasoning != "exposed reasoning" || (stream && visible != "answer") {
				t.Fatalf("reasoning separation failed: %+v %v", m, err)
			}
		})
	}
}
