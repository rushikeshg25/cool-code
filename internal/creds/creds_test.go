package creds

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSetAndGetAPIKey(t *testing.T) {
	tmp := filepath.Join(t.TempDir(), "credentials.json")
	old := path
	path = func() string { return tmp }
	defer func() { path = old }()

	if got := APIKey("anthropic"); got != "" {
		t.Fatalf("expected no key, got %q", got)
	}
	if err := SetAPIKey("anthropic", "sk-test-123"); err != nil {
		t.Fatal(err)
	}
	if got := APIKey("anthropic"); got != "sk-test-123" {
		t.Fatalf("key = %q", got)
	}
	// A second provider must not clobber the first.
	if err := SetAPIKey("openai", "sk-oa"); err != nil {
		t.Fatal(err)
	}
	if APIKey("anthropic") != "sk-test-123" || APIKey("openai") != "sk-oa" {
		t.Fatal("keys clobbered on second SetAPIKey")
	}
	info, err := os.Stat(tmp)
	if err != nil {
		t.Fatal(err)
	}
	if perm := info.Mode().Perm(); perm != 0o600 {
		t.Fatalf("credentials file mode = %o, want 600", perm)
	}
}
