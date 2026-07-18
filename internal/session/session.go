// Package session persists conversation snapshots under ~/.coolcode/sessions.
package session

import (
	"encoding/json"
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

func pathFor(id string) string {
	return filepath.Join(dir(), id+".json")
}

// NewID returns a fresh session id.
func NewID() string {
	return uuid.NewString()
}

// Save writes the session to disk.
func Save(data Data) error {
	if err := os.MkdirAll(dir(), 0o755); err != nil {
		return err
	}
	raw, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(pathFor(data.ID), raw, 0o644)
}

// Load reads a session by id, returning (nil) when missing/corrupt.
func Load(id string) *Data {
	raw, err := os.ReadFile(pathFor(id))
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
	entries, err := os.ReadDir(dir())
	if err != nil {
		return nil
	}
	var out []Data
	for _, e := range entries {
		if filepath.Ext(e.Name()) != ".json" {
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

// Latest returns the most recent session for cwd, or nil.
func Latest(cwd string) *Data {
	list := List(cwd)
	if len(list) == 0 {
		return nil
	}
	return &list[0]
}
