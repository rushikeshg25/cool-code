package security

import "strings"

// SanitizeTerminal removes escape sequences and control characters from text
// that came from a model, a tool result or a repository before it is drawn to
// the terminal. Only newline and tab survive.
//
// Without this, model output reaches the TTY verbatim: OSC 52 writes the
// user's clipboard, OSC 8 forges hyperlinks, CSI sequences move the cursor and
// rewrite lines that were already drawn, and a lone carriage return overwrites
// the line it sits on. That last one matters most in the confirmation prompt,
// where the text being hidden is the command the user is agreeing to run.
func SanitizeTerminal(text string) string {
	if text == "" {
		return ""
	}
	var b strings.Builder
	b.Grow(len(text))

	runes := []rune(text)
	for i := 0; i < len(runes); i++ {
		r := runes[i]
		if r == 0x1B { // ESC
			i = skipEscapeSequence(runes, i)
			continue
		}
		switch {
		case r == '\n' || r == '\t':
			b.WriteRune(r)
		case r < 0x20 || r == 0x7F:
			// Other C0 controls, including carriage return, are dropped.
		case r >= 0x80 && r <= 0x9F:
			// C1 controls, which some terminals treat as escape introducers.
		default:
			b.WriteRune(r)
		}
	}
	return b.String()
}

// skipEscapeSequence returns the index of the last rune of the escape sequence
// beginning at start, so the caller's loop increment lands past it.
func skipEscapeSequence(runes []rune, start int) int {
	i := start + 1
	if i >= len(runes) {
		return i
	}
	switch runes[i] {
	case '[': // CSI: parameters, then a final byte in 0x40-0x7E
		for i++; i < len(runes); i++ {
			if runes[i] >= 0x40 && runes[i] <= 0x7E {
				return i
			}
		}
		return len(runes)
	case ']': // OSC: terminated by BEL or by ST (ESC backslash)
		for i++; i < len(runes); i++ {
			if runes[i] == 0x07 {
				return i
			}
			if runes[i] == 0x1B && i+1 < len(runes) && runes[i+1] == '\\' {
				return i + 1
			}
		}
		return len(runes)
	case 'P', 'X', '^', '_': // DCS, SOS, PM, APC: terminated by ST
		for i++; i < len(runes); i++ {
			if runes[i] == 0x1B && i+1 < len(runes) && runes[i+1] == '\\' {
				return i + 1
			}
			if runes[i] == 0x07 {
				return i
			}
		}
		return len(runes)
	default:
		// Two-character escape, such as ESC c (full reset).
		return i
	}
}

// SanitizeLine is SanitizeTerminal for text that must stay on one line, such
// as the command shown in a confirmation prompt. Newlines and tabs become
// visible markers rather than disappearing, so a payload cannot be pushed out
// of view by padding the string with blank lines.
func SanitizeLine(text string) string {
	cleaned := SanitizeTerminal(text)
	cleaned = strings.ReplaceAll(cleaned, "\t", " ")
	cleaned = strings.ReplaceAll(cleaned, "\n", " ⏎ ")
	return strings.Join(strings.Fields(cleaned), " ")
}
