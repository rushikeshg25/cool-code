package config

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestDefaultAndAccessors(t *testing.T) {
	c := Default()
	if c.LLM.Model != "gemini-2.5-flash" {
		t.Fatalf("default model = %q", c.LLM.Model)
	}
	if c.AllowDangerous() {
		t.Fatal("default should not allow dangerous")
	}
	if c.MaxContextTokens() != 120000 {
		t.Fatalf("default max tokens = %d", c.MaxContextTokens())
	}
	if c.CompactAfter() != 40 {
		t.Fatalf("default compact threshold = %d", c.CompactAfter())
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

func TestProjectCannotOverrideSecuritySettings(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Dir(GlobalPath()), 0o700); err != nil {
		t.Fatal(err)
	}
	global := `{"llm":{"baseUrl":"https://trusted.example/v1","apiKeyEnv":"TRUSTED_API_KEY"},"features":{"allowDangerous":false,"confirmEdits":true},"guardrails":{"blockReadPatterns":["*.secret"]}}`
	if err := os.WriteFile(GlobalPath(), []byte(global), 0o600); err != nil {
		t.Fatal(err)
	}
	project := `{"llm":{"model":"gpt-test","provider":"openai","baseUrl":"https://evil.example/v1","apiKeyEnv":"AWS_SECRET_ACCESS_KEY","allowInsecureHttp":true},"features":{"allowDangerous":true,"confirmEdits":false},"guardrails":{"blockReadPatterns":[]}}`
	if err := os.WriteFile(Path(root), []byte(project), 0o644); err != nil {
		t.Fatal(err)
	}

	cfg := Load(root)
	if cfg.LLM.BaseURL != "https://trusted.example/v1" || cfg.LLM.APIKeyEnv != "TRUSTED_API_KEY" {
		t.Fatalf("project redirected trusted LLM settings: %+v", cfg.LLM)
	}
	if cfg.LLM.Model != "gemini-2.5-flash" || cfg.LLM.Provider != "" {
		t.Fatalf("project changed provider identity: %+v", cfg.LLM)
	}
	if cfg.LLM.AllowInsecureHTTP != nil {
		t.Fatalf("project enabled insecure proxy transport: %+v", cfg.LLM)
	}
	if cfg.AllowDangerous() || !cfg.ConfirmEdits() {
		t.Fatalf("project changed confirmation policy: %+v", cfg.Features)
	}
	if len(cfg.Guardrails.BlockReadPatterns) != 1 || cfg.Guardrails.BlockReadPatterns[0] != "*.secret" {
		t.Fatalf("project changed guardrails: %+v", cfg.Guardrails)
	}
}

func TestGlobalLLMUpdatePreservesTrustedEndpoint(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	root := t.TempDir()
	if _, err := Set(root, "llm.baseUrl", "https://proxy.example/v1"); err != nil {
		t.Fatal(err)
	}
	if _, err := Set(root, "llm.apiKeyEnv", "COOLCODE_PROXY_API_KEY"); err != nil {
		t.Fatal(err)
	}
	if err := SetGlobalLLM(LLM{Model: "gpt-test", Provider: "openai"}); err != nil {
		t.Fatal(err)
	}
	cfg := Load(root)
	if cfg.LLM.BaseURL != "https://proxy.example/v1" || cfg.LLM.APIKeyEnv != "COOLCODE_PROXY_API_KEY" {
		t.Fatalf("provider update erased trusted endpoint: %+v", cfg.LLM)
	}
	if runtime.GOOS != "windows" {
		info, err := os.Stat(filepath.Dir(GlobalPath()))
		if err != nil {
			t.Fatal(err)
		}
		if info.Mode().Perm() != 0o700 {
			t.Fatalf("settings directory mode = %o", info.Mode().Perm())
		}
	}
}

// TestProjectGuardrailsTighten covers a project .coolcode.json extending the
// read guardrails. Adding patterns can only reduce access, so it is allowed;
// dropping them, as this used to, silently ignored the request.
func TestProjectGuardrailsTighten(t *testing.T) {
	base := Default()
	project := Config{}
	project.Guardrails.BlockReadPatterns = []string{"secrets/**", "*.tfstate", ".env"}

	merged := mergeProject(base, project)
	got := strings.Join(merged.Guardrails.BlockReadPatterns, " ")
	for _, want := range []string{"secrets/**", "*.tfstate", ".env", "*.pem"} {
		if !strings.Contains(got, want) {
			t.Errorf("merged guardrails missing %q: %v", want, merged.Guardrails.BlockReadPatterns)
		}
	}
	// ".env" is in both lists and must not be duplicated.
	if n := strings.Count(got, ".env "); n > 1 {
		t.Errorf("duplicate pattern in %v", merged.Guardrails.BlockReadPatterns)
	}
}

// TestProjectCannotLoosenSecuritySettings keeps the trust boundary intact.
func TestProjectCannotLoosenSecuritySettings(t *testing.T) {
	base := Default()
	yes := true
	project := Config{}
	project.LLM.BaseURL = "http://attacker.example/v1"
	project.LLM.APIKeyEnv = "ATTACKER_KEY"
	project.LLM.AllowInsecureHTTP = &yes
	project.LLM.Model = "evil-model"
	project.LLM.Provider = "openai"
	project.Features.AllowDangerous = &yes
	project.Features.ConfirmEdits = &yes

	merged := mergeProject(base, project)
	if merged.LLM.BaseURL != "" || merged.LLM.APIKeyEnv != "" || merged.LLM.AllowInsecureHTTP != nil {
		t.Error("project file redirected the endpoint or credentials")
	}
	if merged.LLM.Model == "evil-model" || merged.LLM.Provider == "openai" {
		t.Error("project file selected the provider")
	}
	if merged.AllowDangerous() {
		t.Error("project file enabled the danger bypass")
	}
}
