package tui

import (
	"bytes"
	"fmt"
	"io"
	"os"

	tea "github.com/charmbracelet/bubbletea"
)

// Bubble Tea v1 reports unfamiliar CSI events through fmt.Stringer rather than
// an exported key type. Keep this adapter confined to exact known events, and
// test it through the real terminal decoder when upgrading Bubble Tea. Normal
// KeyMsgs (especially bracketed paste) must never enter this lookup.
func enhancedKey(msg tea.Msg) tea.Msg {
	if _, ok := msg.(tea.KeyMsg); ok {
		return msg
	}
	if s, ok := msg.(fmt.Stringer); ok {
		if key, found := enhancedKeys[s.String()]; found {
			return key
		}
	}
	return msg
}

var enhancedKeys = func() map[string]tea.Msg {
	keys := map[string]tea.Msg{}
	add := func(sequence string, key tea.Msg) { keys[fmt.Sprintf("?CSI%+v?", []byte(sequence))] = key }
	for code := 9; code <= 127; code++ {
		for mod := 1; mod <= 8; mod++ {
			key, ok := enhancedASCII(code, mod)
			if !ok {
				continue
			}
			for _, locks := range []int{0, 64, 128, 192} {
				encodedMod := mod + locks
				add(fmt.Sprintf("%d;%du", code, encodedMod), key)
				add(fmt.Sprintf("%d;%d:1u", code, encodedMod), key)
				add(fmt.Sprintf("%d;%d:2u", code, encodedMod), key)
				add(fmt.Sprintf("%d;%d:3u", code, encodedMod), nil)
			}
			add(fmt.Sprintf("27;%d;%d~", mod, code), key) // xterm modifyOtherKeys
			if mod == 1 {
				add(fmt.Sprintf("%du", code), key)
			}
		}
	}
	// Navigation remains usable with Caps Lock/Num Lock set as well.
	for _, nav := range []struct {
		code, end      string
		plain, control tea.KeyType
	}{
		{"1", "A", tea.KeyUp, tea.KeyCtrlUp}, {"1", "B", tea.KeyDown, tea.KeyCtrlDown},
		{"1", "C", tea.KeyRight, tea.KeyCtrlRight}, {"1", "D", tea.KeyLeft, tea.KeyCtrlLeft},
		{"1", "H", tea.KeyHome, tea.KeyCtrlHome}, {"1", "F", tea.KeyEnd, tea.KeyCtrlEnd},
		{"5", "~", tea.KeyPgUp, tea.KeyCtrlPgUp}, {"6", "~", tea.KeyPgDown, tea.KeyCtrlPgDown},
		{"3", "~", tea.KeyDelete, tea.KeyDelete},
	} {
		for mod := 1; mod <= 8; mod++ {
			key := tea.KeyMsg{Type: nav.plain, Alt: (mod-1)&2 != 0}
			if (mod-1)&4 != 0 {
				key.Type = nav.control
			}
			if nav.plain == tea.KeyDelete && (mod-1)&4 != 0 {
				key = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}, Alt: true}
			}
			for _, locks := range []int{0, 64, 128, 192} {
				add(fmt.Sprintf("%s;%d%s", nav.code, mod+locks, nav.end), key)
			}
		}
	}
	return keys
}()

func enhancedASCII(code, mod int) (tea.KeyMsg, bool) {
	bits := mod - 1
	alt, ctrl := bits&2 != 0, bits&4 != 0
	key := tea.KeyMsg{Alt: alt}
	switch code {
	case 13:
		if bits != 0 {
			return tea.KeyMsg{Type: tea.KeyCtrlJ}, true
		}
		key.Type = tea.KeyEnter
	case 9:
		key.Type = tea.KeyTab
		if bits&1 != 0 {
			key.Type = tea.KeyShiftTab
		}
	case 27:
		key.Type = tea.KeyEsc
	case 127:
		key.Type = tea.KeyBackspace
		if ctrl {
			key.Type = tea.KeyCtrlW
			key.Alt = false
		}
	default:
		if code < 32 {
			return key, false
		}
		if ctrl {
			switch {
			case code == 32 || code == 64:
				key.Type = tea.KeyCtrlAt
			case code >= 97 && code <= 122:
				key.Type = tea.KeyType(code - 96)
			case code >= 65 && code <= 95:
				key.Type = tea.KeyType(code - 64)
			case code == 47:
				key.Type = tea.KeyCtrlUnderscore
			default:
				return key, false
			}
		} else {
			key.Type = tea.KeyRunes
			key.Runes = []rune{rune(code)}
		}
	}
	return key, true
}

// Scope keyboard negotiation to the alternate screen. Decorating Bubble Tea's
// output keeps writes serialized with rendering and also covers suspend/resume
// and error cleanup. Push/pop restores the terminal's previous keyboard mode.
// See https://sw.kovidgoyal.net/kitty/keyboard-protocol/#progressive-enhancement
// These screen transitions are emitted as complete writes by Bubble Tea v1.
type keyboardOutput struct{ *os.File }

func keyboardScreenModes(p []byte) []byte {
	p = bytes.ReplaceAll(p, []byte("\x1b[?1049h"), []byte("\x1b[?1049h\x1b[>1u"))
	return bytes.ReplaceAll(p, []byte("\x1b[?1049l"), []byte("\x1b[<u\x1b[?1049l"))
}
func (w keyboardOutput) Write(p []byte) (int, error) {
	data := keyboardScreenModes(p)
	n, err := w.File.Write(data)
	if err != nil {
		return 0, err
	}
	if n != len(data) {
		return 0, io.ErrShortWrite
	}
	return len(p), nil
}

// io.WriteString must also pass through the decorator, not os.File.WriteString.
func (w keyboardOutput) WriteString(s string) (int, error) { return w.Write([]byte(s)) }
