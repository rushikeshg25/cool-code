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
	Model       string   `json:"model"`
	Provider    string   `json:"provider,omitempty"`
	Temperature *float64 `json:"temperature,omitempty"`
	MaxTokens   *int     `json:"maxTokens,omitempty"`
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
				".env", ".env.*", "*.pem", "*.key", "id_rsa", "id_ed25519", ".npmrc",
			},
		},
	}
}

// Path returns the config file path for a project root.
func Path(rootDir string) string {
	return filepath.Join(rootDir, ".coolcode.json")
}

// Load reads .coolcode.json merged over defaults; missing/invalid falls back
// to defaults.
func Load(rootDir string) Config {
	def := Default()
	raw, err := os.ReadFile(Path(rootDir))
	if err != nil {
		return def
	}
	var over Config
	if err := json.Unmarshal(raw, &over); err != nil {
		return def
	}
	return merge(def, over)
}

// Save writes the config back to disk as pretty JSON.
func Save(rootDir string, c Config) error {
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(Path(rootDir), append(data, '\n'), 0o644)
}

func merge(base, over Config) Config {
	m := base
	if over.LLM.Model != "" {
		m.LLM.Model = over.LLM.Model
	}
	if over.LLM.Provider != "" {
		m.LLM.Provider = over.LLM.Provider
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
