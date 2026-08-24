package config

import "testing"

func TestDefaultAndAccessors(t *testing.T) {
	c := Default()
	if c.LLM.Model != "gemini-2.5-flash" {
		t.Fatalf("default model = %q", c.LLM.Model)
	}
	if c.AllowDangerous() {
		t.Fatal("default should not allow dangerous")
	}
	if c.MaxContextTokens() != 20000 {
		t.Fatalf("default max tokens = %d", c.MaxContextTokens())
	}
	if !c.ScanCache() {
		t.Fatal("scan cache should default on")
	}
}

func TestGetSetByPath(t *testing.T) {
	c := Default()
	if err := SetByPath(&c, "llm.model", "claude-sonnet-5"); err != nil {
		t.Fatal(err)
	}
	if c.LLM.Model != "claude-sonnet-5" {
		t.Fatalf("model = %q", c.LLM.Model)
	}
	if err := SetByPath(&c, "features.maxContextTokens", ParseValue("15000")); err != nil {
		t.Fatal(err)
	}
	if c.MaxContextTokens() != 15000 {
		t.Fatalf("max tokens = %d", c.MaxContextTokens())
	}
	v, ok := GetByPath(c, "llm.model")
	if !ok || v != "claude-sonnet-5" {
		t.Fatalf("get llm.model = %v, %v", v, ok)
	}
	if _, ok := GetByPath(c, "llm.missing"); ok {
		t.Fatal("expected missing key to be absent")
	}
	if err := SetByPath(&c, "llm.baseUrl", "http://localhost:8317/v1"); err != nil {
		t.Fatal(err)
	}
	if err := SetByPath(&c, "llm.apiKeyEnv", "CLIPROXY_API_KEY"); err != nil {
		t.Fatal(err)
	}
	if c.LLM.BaseURL != "http://localhost:8317/v1" || c.LLM.APIKeyEnv != "CLIPROXY_API_KEY" {
		t.Fatalf("proxy config = %+v", c.LLM)
	}
	if err := SetByPath(&c, "llm.reasoningEffort", "high"); err != nil {
		t.Fatal(err)
	}
	if c.LLM.ReasoningEffort != "high" {
		t.Fatalf("reasoning effort = %q", c.LLM.ReasoningEffort)
	}
}

func TestValidReasoningEffort(t *testing.T) {
	for _, effort := range []string{"", "minimal", "low", "medium", "high", "xhigh"} {
		if !ValidReasoningEffort(effort) {
			t.Errorf("expected %q to be valid", effort)
		}
	}
	if ValidReasoningEffort("maximum") {
		t.Fatal("maximum should not be accepted")
	}
}

func TestParseValue(t *testing.T) {
	cases := []struct {
		in   string
		want any
	}{
		{"true", true},
		{"false", false},
		{"1024", float64(1024)},
		{"hello", "hello"},
		{`"quoted"`, "quoted"},
	}
	for _, tc := range cases {
		if got := ParseValue(tc.in); got != tc.want {
			t.Errorf("ParseValue(%q) = %v (%T), want %v", tc.in, got, got, tc.want)
		}
	}
}
