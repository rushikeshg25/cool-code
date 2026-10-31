package tui

import "testing"

func TestYOLOCommandDuringTurn(t *testing.T) {
	m := newTestModel(t)
	m.processing = true
	for _, command := range []string{"/yolo on", "/yolo off", "/yolo"} {
		m = typeLine(t, m, command)
		want := command != "/yolo off"
		if m.proc.YOLO() != want {
			t.Fatalf("%s: YOLO=%v", command, m.proc.YOLO())
		}
	}
	m = typeLine(t, m, "/yolo invalid")
	if !m.proc.YOLO() {
		t.Fatal("invalid argument changed approvals")
	}
	for _, e := range m.history {
		if e.raw == "(queued) /yolo on" {
			t.Fatal("local command was queued to model")
		}
	}
}
