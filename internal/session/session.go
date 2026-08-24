// Package session persists conversation snapshots under ~/.coolcode/sessions.
package session

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sort"

	"github.com/google/uuid"
)

// Data is a persisted session snapshot. Messages is the opaque, agent-owned
// serialized conversation.
type Data struct {
	ID           string          `json:"id"`
	Cwd          string          `json:"cwd"`
	UpdatedAt    string          `json:"updatedAt"`
	Mode         string          `json:"mode"`
	Messages     json.RawMessage `json:"messages"`
	Summary      string          `json:"summary"`
	PinnedFiles  []string        `json:"pinnedFiles"`
	ExtraDirs    []string        `json:"extraDirs,omitempty"`
	MessageCount int             `json:"messageCount"`
}

func dir() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".coolcode", "sessions")
}

func pathFor(id string) (string, bool) {
	if _, err := uuid.Parse(id); err != nil {
		return "", false
	}
	return filepath.Join(dir(), id+".json"), true
}

// NewID returns a fresh session id.
func NewID() string {
	return uuid.NewString()
}

// Save writes the session to disk.
func Save(data Data) error {
	path, ok := pathFor(data.ID)
	if !ok {
		return errors.New("invalid session id")
	}
	if info, err := os.Lstat(filepath.Dir(dir())); err == nil && info.Mode()&os.ModeSymlink != 0 {
		return errors.New("session parent directory must not be a symlink")
	}
	if info, err := os.Lstat(dir()); err == nil && info.Mode()&os.ModeSymlink != 0 {
		return errors.New("session directory must not be a symlink")
	}
	if err := os.MkdirAll(dir(), 0o700); err != nil {
		return err
	}
	if err := os.Chmod(dir(), 0o700); err != nil {
		return err
	}
	raw, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}
	if info, err := os.Lstat(path); err == nil && info.Mode()&os.ModeSymlink != 0 {
		return errors.New("session file must not be a symlink")
	}
	if err := os.WriteFile(path, raw, 0o600); err != nil {
		return err
	}
	return os.Chmod(path, 0o600)
}

// Load reads a session by id, returning (nil) when missing/corrupt.
func Load(id string) *Data {
	path, ok := pathFor(id)
	if !ok {
		return nil
	}
	if !safeSessionDir() {
		return nil
	}
	info, err := os.Lstat(path)
	if err != nil || !info.Mode().IsRegular() {
		return nil
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil
	}
	var data Data
	if err := json.Unmarshal(raw, &data); err != nil {
		return nil
	}
	return &data
}

// List returns sessions recorded for cwd, newest first.
func List(cwd string) []Data {
	if !safeSessionDir() {
		return nil
	}
	entries, err := os.ReadDir(dir())
	if err != nil {
		return nil
	}
	var out []Data
	for _, e := range entries {
		if filepath.Ext(e.Name()) != ".json" || e.Type()&os.ModeSymlink != 0 || !e.Type().IsRegular() {
			continue
		}
		raw, err := os.ReadFile(filepath.Join(dir(), e.Name()))
		if err != nil {
			continue
		}
		var data Data
		if err := json.Unmarshal(raw, &data); err != nil {
			continue
		}
		if data.Cwd == cwd {
			out = append(out, data)
		}
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i].UpdatedAt > out[j].UpdatedAt
	})
	return out
}

func safeSessionDir() bool {
	parent, err := os.Lstat(filepath.Dir(dir()))
	if err != nil || !parent.IsDir() {
		return false
	}
	info, err := os.Lstat(dir())
	return err == nil && info.IsDir()
}

// Latest returns the most recent session for cwd, or nil.
func Latest(cwd string) *Data {
	list := List(cwd)
	if len(list) == 0 {
		return nil
	}
	return &list[0]
}
