package llm

import "testing"

func TestPriceForMatchesLongestPrefix(t *testing.T) {
	cases := []struct {
		model     string
		wantIn    float64
		wantKnown bool
	}{
		{"claude-sonnet-4-5", 3, true},
		{"claude-opus-4-1-20250805", 15, true},
		{"gpt-4o-mini", 0.15, true},
		{"gpt-4o", 2.50, true},
		{"gemini-2.5-flash", 0.30, true},
		{"GEMINI-2.5-PRO", 1.25, true},
		{"", 0, false},
		{"some-local-proxy-model", 0, false},
	}
	for _, c := range cases {
		got, ok := PriceFor(c.model)
		if ok != c.wantKnown {
			t.Errorf("PriceFor(%q) known = %v, want %v", c.model, ok, c.wantKnown)
			continue
		}
		if ok && got.Input != c.wantIn {
			t.Errorf("PriceFor(%q).Input = %v, want %v", c.model, got.Input, c.wantIn)
		}
	}
}

// TestCostIsSilentForUnknownModels keeps a custom endpoint from being given a
// made-up figure. Reporting $0.00 would read as free rather than as unknown.
func TestCostIsSilentForUnknownModels(t *testing.T) {
	if _, ok := Cost("some-local-proxy-model", 1_000_000, 1_000_000); ok {
		t.Error("unknown model reported a cost")
	}
	got, ok := Cost("claude-sonnet-4-5", 1_000_000, 1_000_000)
	if !ok {
		t.Fatal("known model reported no cost")
	}
	if want := 18.0; got != want {
		t.Errorf("Cost = %v, want %v", got, want)
	}
}
