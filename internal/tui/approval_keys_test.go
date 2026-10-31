package tui

import (
	tea "github.com/charmbracelet/bubbletea"
	"testing"
)

func TestApprovalKeys(t *testing.T) {
	for _, tt := range []struct {
		name string
		key  tea.KeyMsg
		want bool
	}{
		{"enter", tea.KeyMsg{Type: tea.KeyEnter}, true},
		{"yes", tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'y'}}, true},
		{"no", tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}}, false},
		{"escape", tea.KeyMsg{Type: tea.KeyEsc}, false},
	} {
		t.Run(tt.name, func(t *testing.T) {
			m := newTestModel(t)
			response := make(chan bool, 1)
			m.confirmMsg = "Run npm install?"
			m.confirmResp = response
			m.Update(tt.key)
			select {
			case got := <-response:
				if got != tt.want {
					t.Fatalf("approved=%v, want %v", got, tt.want)
				}
			default:
				t.Fatal("no approval decision")
			}
		})
	}
}
