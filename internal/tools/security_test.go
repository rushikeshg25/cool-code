package tools

import (
	"context"
	"net"
	"net/url"
	"strings"
	"testing"
)

func TestSafeCommandEnvDropsSecrets(t *testing.T) {
	t.Setenv("COOLCODE_TEST_API_KEY", "must-not-reach-child")
	t.Setenv("AWS_SECRET_ACCESS_KEY", "must-not-reach-child-either")
	result := execCommand(context.Background(), `printf '%s|%s' "$COOLCODE_TEST_API_KEY" "$AWS_SECRET_ACCESS_KEY"`, t.TempDir(), 0)
	if result.stdout != "|" {
		t.Fatalf("child inherited secrets: %q", result.stdout)
	}
}

func TestValidateWebURLBlocksSSRF(t *testing.T) {
	for _, raw := range []string{
		"http://example.com",
		"https://127.0.0.1/admin",
		"https://169.254.169.254/latest/meta-data",
		"https://10.0.0.1/internal",
		"https://[::1]/admin",
	} {
		u, err := url.Parse(raw)
		if err != nil {
			t.Fatal(err)
		}
		if err := validateWebURL(context.Background(), u); err == nil {
			t.Errorf("URL should be blocked: %s", raw)
		}
	}
}

func TestBlockedNetworkIP(t *testing.T) {
	for _, raw := range []string{"127.0.0.1", "10.0.0.1", "169.254.169.254", "::1", "fc00::1"} {
		if !blockedNetworkIP(net.ParseIP(raw)) {
			t.Errorf("IP should be blocked: %s", raw)
		}
	}
	if blockedNetworkIP(net.ParseIP("8.8.8.8")) {
		t.Fatal("public IP was blocked")
	}
}

func TestShellConfirmationIncludesSanitizedCommand(t *testing.T) {
	t.Setenv("COOLCODE_TEST_API_KEY", "never-show-this-value")
	reason := DangerReason("shell_command", args(t, map[string]any{"command": "echo never-show-this-value"}))
	if !strings.Contains(reason, "shell command") || strings.Contains(reason, "never-show-this-value") {
		t.Fatalf("unsafe confirmation text: %s", reason)
	}
}
