package llm

import "testing"

func TestContextWindowMatchesLongestPrefix(t *testing.T) {
	cases := []struct {
		model     string
		want      int
		wantKnown bool
	}{
		{"claude-sonnet-4-5", 200_000, true},
		{"gpt-4o-mini", 128_000, true},
		{"gpt-4o", 128_000, true},
		{"GEMINI-2.5-PRO", 1_048_576, true},
		{"o3-mini", 200_000, true},
		// A custom endpoint's model is unknown, and must stay unknown rather
		// than borrowing a number from an unrelated family.
		{"gpt-5.6-sol", 0, false},
		{"some-local-proxy-model", 0, false},
		{"", 0, false},
	}
	for _, c := range cases {
		got, ok := ContextWindow(c.model)
		if ok != c.wantKnown || got != c.want {
			t.Errorf("ContextWindow(%q) = %d, %v; want %d, %v", c.model, got, ok, c.want, c.wantKnown)
		}
	}
}
