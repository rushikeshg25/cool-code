package tui

import (
	"strings"
	"testing"
)

func TestUsageDuringTurn(t *testing.T) {
	m := newTestModel(t)
	m.processing = true
	m = typeLine(t, m, "/usage")
	if m.proc.PendingMessageCount() != 0 {
		t.Fatal("usage command queued")
	}
	found := false
	for _, e := range m.history {
		if strings.Contains(e.raw, "Dev token usage") {
			found = true
		}
	}
	if !found {
		t.Fatal("usage report missing")
	}
}
