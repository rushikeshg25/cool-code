package tui

import (
	"context"
	tea "github.com/charmbracelet/bubbletea"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// Exercise Bubble Tea's real terminal decoder as well as the composer handler.
type typingProbe struct{ *model }

func (p typingProbe) Init() tea.Cmd { return nil }
func (p typingProbe) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if key, ok := msg.(tea.KeyMsg); ok && key.String() == "~" {
		return p, tea.Quit
	}
	p.model.Update(msg)
	return p, nil
}
func (p typingProbe) View() string { return "" }
func TestTerminalTypingShortcuts(t *testing.T) {
	for _, tt := range []struct{ name, seq, want string }{
		{"ctrl_space", "\x00", "hello "},
		{"kitty_ctrl_space", "\x1b[32;5u", "hello "},
		{"kitty_ctrl_backspace", "\x1b[127;5u", "hello "},
		{"caps_lock_ctrl_enter", "\x1b[13;69u", "hello world\n"},
		{"num_lock_word_delete", "\x1b[1;133D\x1b[3;133~", "hello "},
		{"kitty_ctrl_enter", "\x1b[13;5u", "hello world\n"},
		{"kitty_shift_enter", "\x1b[13;2u", "hello world\n"},
		{"xterm_ctrl_enter", "\x1b[27;5;13~", "hello world\n"},
		{"ctrl_w", "\x17", "hello "},
		{"alt_enter", "\x1b\r", "hello world\n"},
		{"ctrl_j", "\n", "hello world\n"},
		{"word_forward_delete", "\x1b[1;5D\x1b[3;5~", "hello "},
		{"enhanced_line_start_newline", "\x1b[97;5u\x1b[13;5u", "\nhello world"},
		{"release_ignored", "\x1b[32;5:3u", "hello world"},
		{"press_and_repeat", "\x1b[13;5:1u\x1b[13;5:2u", "hello world\n\n"},
		{"paste", "\x1b[200~\nnext line\x1b[201~", "hello world\nnext line"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			m := newTestModel(t)
			m.ti.SetValue("hello world")
			m.ti.CursorEnd()
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()
			p := tea.NewProgram(typingProbe{m}, tea.WithContext(ctx), tea.WithInput(strings.NewReader(tt.seq+"~")), tea.WithOutput(io.Discard), tea.WithoutRenderer(), tea.WithoutSignalHandler())
			if _, err := p.Run(); err != nil {
				t.Fatal(err)
			}
			if got := m.ti.Value(); got != tt.want {
				t.Fatalf("composer = %q, want %q", got, tt.want)
			}
			if m.processing {
				t.Fatal("editing shortcut submitted the prompt")
			}
		})
	}
}

func TestKeyboardProtocolScreenLifecycle(t *testing.T) {
	file, err := os.Create(filepath.Join(t.TempDir(), "terminal"))
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	m := newTestModel(t)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	p := tea.NewProgram(typingProbe{m}, tea.WithContext(ctx), tea.WithInput(strings.NewReader("~")), tea.WithOutput(keyboardOutput{file}), tea.WithAltScreen(), tea.WithoutSignalHandler())
	if _, err = p.Run(); err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile(file.Name())
	if err != nil {
		t.Fatal(err)
	}
	text := string(raw)
	if strings.Count(text, "\x1b[?1049h\x1b[>1u") != 1 || strings.Count(text, "\x1b[<u\x1b[?1049l") != 1 {
		t.Fatalf("keyboard mode not scoped to alternate screen: %q", text)
	}
}
