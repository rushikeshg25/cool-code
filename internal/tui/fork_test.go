package tui

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/rushikeshg25/cool-code/internal/llm"
	"github.com/rushikeshg25/cool-code/internal/session"
	"github.com/rushikeshg25/cool-code/internal/types"
)

func TestForkPreservesOriginalAndSwitchesChat(t *testing.T) {
	m := newTestModel(t)
	m.sessionID = session.NewID()
	parent := m.sessionID
	m.proc.Restore([]llm.Message{{Role: llm.RoleUser, Text: "Original"}}, "summary", nil, types.ModeAsk)
	m.forkSession()
	child := m.sessionID
	if child == parent {
		t.Fatal("did not fork")
	}
	original, copy := session.Load(parent), session.Load(child)
	if original == nil || copy == nil || string(original.Messages) != string(copy.Messages) || copy.Summary != "summary" || copy.Mode != "ask" {
		t.Fatal("fork lost state")
	}
	m.proc.Restore([]llm.Message{{Role: llm.RoleUser, Text: "Child only"}}, "summary", nil, types.ModeAsk)
	m.sessionList = []session.Data{*original}
	m.sessionIdx = 0
	m.resumeSelectedSession()
	if m.sessionID != parent {
		t.Fatal("could not return to original")
	}
	messages, _, _, _, _ := m.proc.Snapshot()
	if messages[0].Text != "Original" {
		t.Fatal(messages)
	}
	saved := session.Load(child)
	if saved == nil || string(saved.Messages) == string(original.Messages) {
		t.Fatal("child edits not saved")
	}
	if string(session.Load(parent).Messages) != string(original.Messages) {
		t.Fatal("parent overwritten")
	}
}
func TestForkFailureKeepsCurrentChat(t *testing.T) {
	m := newTestModel(t)
	m.sessionID = session.NewID()
	id := m.sessionID
	if err := os.WriteFile(filepath.Join(os.Getenv("HOME"), ".coolcode"), []byte("blocking file"), 0600); err != nil {
		t.Fatal(err)
	}
	m.forkSession()
	if m.sessionID != id {
		t.Fatal("switched after failed save")
	}
	if session.Load(id) != nil {
		t.Fatal("unexpected saved state")
	}
}
