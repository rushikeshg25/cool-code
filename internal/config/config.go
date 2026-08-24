// Package config loads and persists the project .coolcode.json file and
// provides dotted-path get/set used by the `config` subcommand.
package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// LLM holds model/provider settings.
type LLM struct {
	Model             string   `json:"model"`
	Provider          string   `json:"provider,omitempty"`
	BaseURL           string   `json:"baseUrl,omitempty"`
	APIKeyEnv         string   `json:"apiKeyEnv,omitempty"`
	AllowInsecureHTTP *bool    `json:"allowInsecureHttp,omitempty"`
	ReasoningEffort   string   `json:"reasoningEffort,omitempty"`
	Temperature       *float64 `json:"temperature,omitempty"`
	MaxTokens         *int     `json:"maxTokens,omitempty"`
}

// Features holds behavioural toggles.
type Features struct {
	FileTreeMaxDepth *int  `json:"fileTreeMaxDepth,omitempty"`
	ScanCache        *bool `json:"scanCache,omitempty"`
	AllowDangerous   *bool `json:"allowDangerous,omitempty"`
	ConfirmEdits     *bool `json:"confirmEdits,omitempty"`
	MaxContextTokens *int  `json:"maxContextTokens,omitempty"`
}

// Guardrails holds read-blocking patterns.
type Guardrails struct {
	BlockReadPatterns []string `json:"blockReadPatterns,omitempty"`
}

// Config is the full .coolcode.json shape.
type Config struct {
	LLM        LLM        `json:"llm"`
	Features   Features   `json:"features,omitempty"`
	Guardrails Guardrails `json:"guardrails,omitempty"`
}

func boolPtr(b bool) *bool { return &b }
func intPtr(i int) *int    { return &i }

// Default returns the built-in default configuration.
func Default() Config {
	return Config{
		LLM: LLM{Model: "gemini-2.5-flash"},
		Features: Features{
			ScanCache:        boolPtr(true),
			AllowDangerous:   boolPtr(false),
			ConfirmEdits:     boolPtr(false),
			MaxContextTokens: intPtr(20000),
		},
		Guardrails: Guardrails{
			BlockReadPatterns: []string{
				".env", ".env.*", "*.pem", "*.key", "*.p12", "*.pfx",
				"id_rsa", "id_ed25519", ".npmrc", ".netrc", ".git-credentials",
				"**/.aws/credentials", "**/.kube/config", "**/.docker/config.json",
			},
		},
	}
}

// Path returns the config file path for a project root.
func Path(rootDir string) string {
	return filepath.Join(rootDir, ".coolcode.json")
}

// GlobalPath returns the user-level settings file (~/.coolcode/settings.json).
func GlobalPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".coolcode", "settings.json")
}

// Load returns defaults merged with trusted user settings and then safe
// project-local preferences. Security-sensitive settings are deliberately
// ignored in .coolcode.json so an untrusted repository cannot redirect
// credentials, disable guardrails, or silently bypass confirmations.
func Load(rootDir string) Config {
	cfg := Default()
	if p := GlobalPath(); p != "" {
		cfg = mergeFile(cfg, p)
	}
	raw, err := readRegularFile(Path(rootDir))
	if err != nil {
		return cfg
	}
	var project Config
	if json.Unmarshal(raw, &project) != nil {
		return cfg
	}
	return mergeProject(cfg, project)
}

func mergeFile(base Config, path string) Config {
	raw, err := readRegularFile(path)
	if err != nil {
		return base
	}
	var over Config
	if err := json.Unmarshal(raw, &over); err != nil {
		return base
	}
	return merge(base, over)
}

// SetGlobalLLM persists model/provider defaults to the global settings file,
// preserving any other settings already stored there.
func SetGlobalLLM(llm LLM) error {
	p := GlobalPath()
	if p == "" {
		return os.ErrNotExist
	}
	var current Config
	if raw, err := readRegularFile(p); err == nil {
		_ = json.Unmarshal(raw, &current)
	}
	current = merge(current, Config{LLM: llm})
	if err := rejectSymlink(p); err != nil {
		return err
	}
	if err := rejectSymlink(filepath.Dir(p)); err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(p), 0o700); err != nil {
		return err
	}
	if err := os.Chmod(filepath.Dir(p), 0o700); err != nil {
		return err
	}
	data, err := json.MarshalIndent(current, "", "  ")
	if err != nil {
		return err
	}
	if err := os.WriteFile(p, append(data, '\n'), 0o600); err != nil {
		return err
	}
	return os.Chmod(p, 0o600)
}

// Set persists one setting in the appropriate trust scope. Endpoint,
// provider identity, credential, confirmation, and guardrail settings are
// global-only; generation and UI preferences remain project-local.
func Set(rootDir, path string, value any) (string, error) {
	dest := Path(rootDir)
	mode := os.FileMode(0o644)
	if globalOnly(path) {
		dest = GlobalPath()
		mode = 0o600
	}
	if dest == "" {
		return "", os.ErrNotExist
	}
	var current Config
	if raw, err := readRegularFile(dest); err == nil {
		_ = json.Unmarshal(raw, &current)
	}
	if err := rejectSymlink(dest); err != nil {
		return "", err
	}
	if err := rejectSymlink(filepath.Dir(dest)); err != nil {
		return "", err
	}
	if err := SetByPath(&current, path, value); err != nil {
		return "", err
	}
	data, err := json.MarshalIndent(current, "", "  ")
	if err != nil {
		return "", err
	}
	if err := os.MkdirAll(filepath.Dir(dest), 0o700); err != nil {
		return "", err
	}
	if mode == 0o600 {
		if err := os.Chmod(filepath.Dir(dest), 0o700); err != nil {
			return "", err
		}
	}
	if err := os.WriteFile(dest, append(data, '\n'), mode); err != nil {
		return "", err
	}
	if err := os.Chmod(dest, mode); err != nil {
		return "", err
	}
	return dest, nil
}

func globalOnly(path string) bool {
	switch path {
	case "llm.model", "llm.provider", "llm.baseUrl", "llm.apiKeyEnv", "llm.allowInsecureHttp", "features.allowDangerous", "features.confirmEdits":
		return true
	}
	return path == "guardrails" || strings.HasPrefix(path, "guardrails.")
}

func merge(base, over Config) Config {
	m := base
	if over.LLM.Model != "" {
		m.LLM.Model = over.LLM.Model
	}
	if over.LLM.Provider != "" {
		m.LLM.Provider = over.LLM.Provider
	}
	if over.LLM.BaseURL != "" {
		m.LLM.BaseURL = over.LLM.BaseURL
	}
	if over.LLM.APIKeyEnv != "" {
		m.LLM.APIKeyEnv = over.LLM.APIKeyEnv
	}
	if over.LLM.AllowInsecureHTTP != nil {
		m.LLM.AllowInsecureHTTP = over.LLM.AllowInsecureHTTP
	}
	if over.LLM.ReasoningEffort != "" {
		m.LLM.ReasoningEffort = over.LLM.ReasoningEffort
	}
	if over.LLM.Temperature != nil {
		m.LLM.Temperature = over.LLM.Temperature
	}
	if over.LLM.MaxTokens != nil {
		m.LLM.MaxTokens = over.LLM.MaxTokens
	}
	if over.Features.FileTreeMaxDepth != nil {
		m.Features.FileTreeMaxDepth = over.Features.FileTreeMaxDepth
	}
	if over.Features.ScanCache != nil {
		m.Features.ScanCache = over.Features.ScanCache
	}
	if over.Features.AllowDangerous != nil {
		m.Features.AllowDangerous = over.Features.AllowDangerous
	}
	if over.Features.ConfirmEdits != nil {
		m.Features.ConfirmEdits = over.Features.ConfirmEdits
	}
	if over.Features.MaxContextTokens != nil {
		m.Features.MaxContextTokens = over.Features.MaxContextTokens
	}
	if over.Guardrails.BlockReadPatterns != nil {
		m.Guardrails.BlockReadPatterns = over.Guardrails.BlockReadPatterns
	}
	return m
}

func readRegularFile(path string) ([]byte, error) {
	parent, err := os.Lstat(filepath.Dir(path))
	if err != nil || !parent.IsDir() {
		return nil, os.ErrPermission
	}
	info, err := os.Lstat(path)
	if err != nil {
		return nil, err
	}
	if !info.Mode().IsRegular() {
		return nil, os.ErrPermission
	}
	return os.ReadFile(path)
}

func rejectSymlink(path string) error {
	info, err := os.Lstat(path)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return err
	}
	if info.Mode()&os.ModeSymlink != 0 {
		return os.ErrPermission
	}
	return nil
}

func mergeProject(base, over Config) Config {
	// Only preferences that cannot expand host access or redirect secrets are
	// accepted from a repository-controlled file.
	over.LLM.BaseURL = ""
	over.LLM.APIKeyEnv = ""
	over.LLM.AllowInsecureHTTP = nil
	over.LLM.Model = ""
	over.LLM.Provider = ""
	over.Features.AllowDangerous = nil
	over.Features.ConfirmEdits = nil
	over.Guardrails.BlockReadPatterns = nil
	return merge(base, over)
}

// ValidReasoningEffort reports whether effort is supported by OpenAI-style
// reasoning models. Empty leaves the provider default unchanged.
func ValidReasoningEffort(effort string) bool {
	switch effort {
	case "", "minimal", "low", "medium", "high", "xhigh":
		return true
	default:
		return false
	}
}

// AllowDangerous reports the effective danger toggle.
func (c Config) AllowDangerous() bool {
	return c.Features.AllowDangerous != nil && *c.Features.AllowDangerous
}

// ConfirmEdits reports whether edits require confirmation.
func (c Config) ConfirmEdits() bool {
	return c.Features.ConfirmEdits != nil && *c.Features.ConfirmEdits
}

// ScanCache reports whether scan results should be cached.
func (c Config) ScanCache() bool {
	return c.Features.ScanCache != nil && *c.Features.ScanCache
}

// MaxContextTokens returns the configured window (default 20000).
func (c Config) MaxContextTokens() int {
	if c.Features.MaxContextTokens != nil {
		return *c.Features.MaxContextTokens
	}
	return 20000
}

// GetByPath resolves a dotted key against the config, returning the value and
// whether it was found.
func GetByPath(c Config, path string) (any, bool) {
	m := toMap(c)
	var cur any = m
	for _, part := range splitPath(path) {
		obj, ok := cur.(map[string]any)
		if !ok {
			return nil, false
		}
		cur, ok = obj[part]
		if !ok {
			return nil, false
		}
	}
	return cur, true
}

// SetByPath sets a dotted key on the config by round-tripping through a map.
func SetByPath(c *Config, path string, value any) error {
	m := toMap(*c)
	parts := splitPath(path)
	if len(parts) == 0 {
		return nil
	}
	cur := m
	for _, part := range parts[:len(parts)-1] {
		next, ok := cur[part].(map[string]any)
		if !ok {
			next = map[string]any{}
			cur[part] = next
		}
		cur = next
	}
	cur[parts[len(parts)-1]] = value

	data, err := json.Marshal(m)
	if err != nil {
		return err
	}
	var merged Config
	if err := json.Unmarshal(data, &merged); err != nil {
		return err
	}
	*c = merged
	return nil
}

// ParseValue interprets a raw CLI value: JSON when it looks like JSON, a number
// when numeric, otherwise the literal string.
func ParseValue(raw string) any {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return ""
	}
	if trimmed == "true" || trimmed == "false" || trimmed == "null" ||
		strings.HasPrefix(trimmed, "{") || strings.HasPrefix(trimmed, "[") ||
		strings.HasPrefix(trimmed, "\"") || strings.HasPrefix(trimmed, "'") {
		normalized := trimmed
		if strings.HasPrefix(normalized, "'") && strings.HasSuffix(normalized, "'") {
			normalized = "\"" + strings.Trim(normalized, "'") + "\""
		}
		var v any
		if err := json.Unmarshal([]byte(normalized), &v); err == nil {
			return v
		}
		return raw
	}
	if n, err := strconv.ParseFloat(trimmed, 64); err == nil {
		return n
	}
	return raw
}

func splitPath(path string) []string {
	var out []string
	for _, p := range strings.Split(path, ".") {
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

func toMap(c Config) map[string]any {
	data, _ := json.Marshal(c)
	var m map[string]any
	_ = json.Unmarshal(data, &m)
	if m == nil {
		m = map[string]any{}
	}
	return m
}
