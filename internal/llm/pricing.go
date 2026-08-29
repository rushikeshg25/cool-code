package llm

import "strings"

// Price is the cost of a million tokens, in US dollars.
type Price struct {
	Input  float64
	Output float64
}

// prices maps a model-id prefix to its published rate. Prefixes are matched
// longest first, so a specific id wins over a family default.
//
// These are list prices and will drift. They exist to give a running figure in
// the status bar, not to reconcile a bill: an unknown model reports no cost
// rather than a guess, and a custom endpoint may charge something else entirely.
var prices = map[string]Price{
	// Anthropic
	"claude-opus-4":     {Input: 15, Output: 75},
	"claude-sonnet-4":   {Input: 3, Output: 15},
	"claude-haiku-4":    {Input: 0.80, Output: 4},
	"claude-3-5-haiku":  {Input: 0.80, Output: 4},
	"claude-3-5-sonnet": {Input: 3, Output: 15},
	"claude-3-opus":     {Input: 15, Output: 75},

	// OpenAI
	"gpt-4o-mini": {Input: 0.15, Output: 0.60},
	"gpt-4o":      {Input: 2.50, Output: 10},
	"gpt-4.1":     {Input: 2, Output: 8},
	"o3-mini":     {Input: 1.10, Output: 4.40},
	"o3":          {Input: 2, Output: 8},

	// Google
	"gemini-2.5-pro":   {Input: 1.25, Output: 10},
	"gemini-2.5-flash": {Input: 0.30, Output: 2.50},
	"gemini-2.0-flash": {Input: 0.10, Output: 0.40},
}

// PriceFor returns the rate for a model id and whether one is known.
func PriceFor(model string) (Price, bool) {
	id := strings.ToLower(strings.TrimSpace(model))
	best := ""
	for prefix := range prices {
		if strings.HasPrefix(id, prefix) && len(prefix) > len(best) {
			best = prefix
		}
	}
	if best == "" {
		return Price{}, false
	}
	return prices[best], true
}

// Cost returns the dollar cost of a token count for a model, and whether the
// model's rate is known.
func Cost(model string, input, output int) (float64, bool) {
	p, ok := PriceFor(model)
	if !ok {
		return 0, false
	}
	const perMillion = 1_000_000
	return float64(input)/perMillion*p.Input + float64(output)/perMillion*p.Output, true
}
