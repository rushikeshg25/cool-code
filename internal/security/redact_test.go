package security

import (
	"strings"
	"testing"
)

func TestRedact(t *testing.T) {
	t.Setenv("COOLCODE_TEST_API_KEY", "environment-secret-value")
	input := strings.Join([]string{
		"key=environment-secret-value",
		"apiKey: hardcoded-secret-value",
		"token=sk-abcdefghijklmnopqrstuvwxyz",
		"github=ghp_abcdefghijklmnopqrstuvwxyz123456",
		"aws=AKIAIOSFODNN7EXAMPLE",
		"-----BEGIN PRIVATE KEY-----\nsecret\n-----END PRIVATE KEY-----",
	}, "\n")
	got := Redact(input)
	for _, secret := range []string{"environment-secret-value", "hardcoded-secret-value", "sk-abc", "ghp_abc", "AKIAIOSFODNN7EXAMPLE", "BEGIN PRIVATE KEY"} {
		if strings.Contains(got, secret) {
			t.Fatalf("secret %q was not redacted: %s", secret, got)
		}
	}
	if !strings.Contains(got, Replacement) {
		t.Fatalf("redaction marker missing: %s", got)
	}
}
