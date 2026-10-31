package tui

import (
	"context"
	tea "github.com/charmbracelet/bubbletea"
	"strings"
	"testing"
)

func TestSendNowWaitsForCancelledTurn(t *testing.T) {
	m := newTestModel(t)
	m.processing = true
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	m.cancelTurn = cancel
	m.proc.EnqueueMessage("first")
	m.appendSystem("(queued) first")
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'s'}, Alt: true})
	if ctx.Err() == nil || cmd != nil || !m.processing {
		t.Fatal("send-now did not wait for cancellation cleanup")
	}
	if m.proc.PendingMessageCount() != 0 || len(m.sendNowMessages) != 1 {
		t.Fatal("pending message was not captured atomically")
	}
	m.proc.EnqueueMessage("second")
	m.appendSystem("(queued) second")
	_, cmd = m.Update(doneMsg{err: context.Canceled})
	if cmd == nil || !m.processing || len(m.sendNowMessages) != 0 {
		t.Fatal("next turn not started after cleanup")
	}
	users := 0
	for _, e := range m.history {
		if strings.HasPrefix(e.raw, "(queued)") {
			t.Fatal("stale queue label")
		}
		if e.kind == entryUser {
			users++
			if e.raw != "first\nsecond" {
				t.Fatalf("sent %q", e.raw)
			}
		}
	}
	if users != 1 {
		t.Fatalf("started %d user turns", users)
	}
}
func TestManualCancelPreservesPendingMessages(t *testing.T) {
	m := newTestModel(t)
	m.processing = true
	m.proc.EnqueueMessage("later")
	_, cmd := m.Update(doneMsg{err: context.Canceled})
	if cmd != nil || m.processing || m.proc.PendingMessageCount() != 1 {
		t.Fatal("manual cancellation unexpectedly sent/lost follow-up")
	}
	_, cmd = m.dispatchCommand("/send")
	if cmd == nil || !m.processing || m.proc.PendingMessageCount() != 0 {
		t.Fatal("/send did not deliver pending follow-up")
	}
}
func TestCompletionDrainsLateFollowup(t *testing.T) {
	m := newTestModel(t)
	m.processing = true
	m.proc.EnqueueMessage("late")
	_, cmd := m.Update(doneMsg{final: "finished"})
	if cmd == nil || !m.processing || m.proc.PendingMessageCount() != 0 {
		t.Fatal("follow-up arriving after worker return was stranded")
	}
}
func TestQueueDeliveryUpdatesOnlyMatchingLabels(t *testing.T) {
	m := newTestModel(t)
	m.appendSystem("(queued) first")
	m.appendSystem("(queued) second")
	m.Update(queueDeliveredMsg{"first"})
	if m.history[0].raw != "(sent follow-up) first" || m.history[1].raw != "(queued) second" {
		t.Fatal("delivery labels not matched to their messages")
	}
}

func TestSendNowDeclinesLateApproval(t *testing.T) {
	m := newTestModel(t)
	m.processing = true
	m.sendNowMessages = []string{"next"}
	response := make(chan bool, 1)
	m.Update(confirmReqMsg{message: "pending tool", resp: response})
	select {
	case approved := <-response:
		if approved {
			t.Fatal("cancelled tool approved")
		}
	default:
		t.Fatal("late approval left worker blocked")
	}
	if m.confirmMsg != "" {
		t.Fatal("cancelled turn reopened confirmation")
	}
}
func TestNewPromptKeepsPendingFollowupsInOrder(t *testing.T) {
	m := newTestModel(t)
	m.proc.EnqueueMessage("old")
	m = typeLine(t, m, "new")
	if got := m.history[len(m.history)-1].raw; got != "old\nnew" {
		t.Fatalf("prompt order: %q", got)
	}
}
