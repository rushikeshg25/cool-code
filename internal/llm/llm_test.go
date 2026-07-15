package llm

import "testing"

func TestInferProvider(t *testing.T) {
	cases := map[string]string{
		"gemini-2.5-flash":   "google",
		"gpt-4o":             "openai",
		"o1-preview":         "openai",
		"o3-mini":            "openai",
		"claude-sonnet-5":    "anthropic",
		"claude-opus-4-8":    "anthropic",
		"some-unknown-model": "google",
	}
	for model, want := range cases {
		if got := InferProvider(model); got != want {
			t.Errorf("InferProvider(%q) = %q, want %q", model, got, want)
		}
	}
}

func TestConvertSchemaForGemini(t *testing.T) {
	in := map[string]any{
		"type": "object",
		"properties": map[string]any{
			"name": map[string]any{"type": "string"},
			"tags": map[string]any{"type": "array", "items": map[string]any{"type": "string"}},
		},
		"required": []string{"name"},
	}
	out := convertSchemaForGemini(in)
	if out["type"] != "OBJECT" {
		t.Errorf("root type = %v", out["type"])
	}
	props := out["properties"].(map[string]any)
	if props["name"].(map[string]any)["type"] != "STRING" {
		t.Errorf("name type = %v", props["name"])
	}
	tags := props["tags"].(map[string]any)
	if tags["type"] != "ARRAY" {
		t.Errorf("tags type = %v", tags["type"])
	}
	if tags["items"].(map[string]any)["type"] != "STRING" {
		t.Errorf("items type = %v", tags["items"])
	}
}

func TestMissingKeyError(t *testing.T) {
	err := &MissingKeyError{Provider: "google", EnvKey: "GOOGLE_GENERATIVE_AI_API_KEY"}
	if err.Error() == "" {
		t.Error("expected non-empty error message")
	}
}
