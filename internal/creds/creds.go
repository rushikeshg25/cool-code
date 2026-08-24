// Package creds stores provider API keys in ~/.coolcode/credentials.json so
// users can connect providers via /connect instead of exporting env vars.
package creds

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// entry holds the stored secret for one provider.
type entry struct {
	APIKey string `json:"apiKey"`
}

type file struct {
	Providers map[string]entry `json:"providers"`
}

// path returns the credentials file location; overridable for tests.
var path = func() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".coolcode", "credentials.json")
}

func load() file {
	f := file{Providers: map[string]entry{}}
	p := path()
	if p == "" {
		return f
	}
	parent, err := os.Lstat(filepath.Dir(p))
	if err != nil || !parent.IsDir() {
		return f
	}
	info, err := os.Lstat(p)
	if err != nil || !info.Mode().IsRegular() {
		return f
	}
	raw, err := os.ReadFile(p)
	if err != nil {
		return f
	}
	_ = json.Unmarshal(raw, &f)
	if f.Providers == nil {
		f.Providers = map[string]entry{}
	}
	return f
}

// APIKey returns the stored key for a provider, or "".
func APIKey(provider string) string {
	return load().Providers[provider].APIKey
}

// SetAPIKey stores a provider key, creating the file with 0600 permissions.
func SetAPIKey(provider, key string) error {
	p := path()
	if p == "" {
		return os.ErrNotExist
	}
	f := load()
	f.Providers[provider] = entry{APIKey: key}
	dir := filepath.Dir(p)
	if info, err := os.Lstat(dir); err == nil && info.Mode()&os.ModeSymlink != 0 {
		return os.ErrPermission
	}
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return err
	}
	if err := os.Chmod(dir, 0o700); err != nil {
		return err
	}
	data, err := json.MarshalIndent(f, "", "  ")
	if err != nil {
		return err
	}
	if info, err := os.Lstat(p); err == nil && info.Mode()&os.ModeSymlink != 0 {
		return os.ErrPermission
	}
	if err := os.WriteFile(p, append(data, '\n'), 0o600); err != nil {
		return err
	}
	return os.Chmod(p, 0o600)
}

// Path exposes the credentials file location for user-facing messages.
func Path() string { return path() }
