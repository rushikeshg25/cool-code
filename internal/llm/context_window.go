package llm

import "strings"

// contextWindows maps a model-id prefix to its context window in tokens.
// Prefixes are matched longest first, so a specific id wins over a family
// default, exactly as prices are matched.
//
// A model that is not listed reports no window rather than a guess. Callers
// show an absolute token count in that case: an invented denominator produces a
// percentage that looks authoritative and is wrong, which is worse than no
// percentage at all. A custom endpoint can declare its own with the
// llm.contextWindow setting.
var contextWindows = map[string]int{
	// Anthropic
	"claude-opus-4":     200_000,
	"claude-sonnet-4":   200_000,
	"claude-haiku-4":    200_000,
	"claude-3-5-sonnet": 200_000,
	"claude-3-5-haiku":  200_000,
	"claude-3-opus":     200_000,

	// OpenAI
	"gpt-4o":      128_000,
	"gpt-4o-mini": 128_000,
	"gpt-4.1":     1_047_576,
	"o3":          200_000,
	"o3-mini":     200_000,

	// Google
	"gemini-2.5-pro":   1_048_576,
	"gemini-2.5-flash": 1_048_576,
	"gemini-2.0-flash": 1_048_576,
}

// ContextWindow returns the context window for a model id and whether one is
// known.
func ContextWindow(model string) (int, bool) {
	id := strings.ToLower(strings.TrimSpace(model))
	best := ""
	for prefix := range contextWindows {
		if strings.HasPrefix(id, prefix) && len(prefix) > len(best) {
			best = prefix
		}
	}
	if best == "" {
		return 0, false
	}
	return contextWindows[best], true
}
