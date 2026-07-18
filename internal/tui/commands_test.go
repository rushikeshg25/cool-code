package tui

import "testing"

func TestAtToken(t *testing.T) {
	cases := []struct {
		in    string
		token string
		at    int
		ok    bool
	}{
		{"@main", "main", 0, true},
		{"see @internal/tui", "internal/tui", 4, true},
		{"@", "", 0, true},
		{"no mention here", "", 0, false},
		{"email me@example.com", "", 0, false}, // @ not at a word boundary
		{"@done already", "", 0, false},        // whitespace after token
	}
	for _, c := range cases {
		token, at, ok := atToken(c.in)
		if ok != c.ok || (ok && (token != c.token || at != c.at)) {
			t.Errorf("atToken(%q) = (%q,%d,%v), want (%q,%d,%v)",
				c.in, token, at, ok, c.token, c.at, c.ok)
		}
	}
}

func TestMatchFiles(t *testing.T) {
	files := []string{"internal/tui/app.go", "internal/agent/processor.go", "README.md"}

	// Basename match ranks ahead of a path-only match.
	got := matchFiles(files, "app")
	if len(got) != 1 || got[0].name != "internal/tui/app.go" {
		t.Fatalf("matchFiles app = %v", got)
	}

	// Empty token returns everything (capped).
	if got := matchFiles(files, ""); len(got) != len(files) {
		t.Fatalf("matchFiles empty = %v", got)
	}

	// Path substring that is not in any basename still matches.
	got = matchFiles(files, "agent")
	if len(got) != 1 || got[0].name != "internal/agent/processor.go" {
		t.Fatalf("matchFiles agent = %v", got)
	}
}
