package tools

import (
	"context"
	"net"
	"net/url"
	"os"
	"path/filepath"
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

// TestExecArgvDoesNotInterpretShellMetacharacters covers the injection that
// shellEscapeSingleQuotes used to allow. Its replacement emitted '\"'\"' where
// POSIX requires '\”, which left the shell parser unquoted, so any value
// containing a single quote could break out and run arbitrary commands.
func TestExecArgvDoesNotInterpretShellMetacharacters(t *testing.T) {
	dir := t.TempDir()
	marker := filepath.Join(dir, "pwned")
	payload := "handler'; touch " + marker + "; echo INJECTED #"

	// Mirrors the find_symbol argv: the payload is the search pattern.
	res := execArgv(context.Background(), dir, 0, "grep", "-r", "--", payload, ".")

	if strings.Contains(res.stdout, "INJECTED") {
		t.Fatalf("payload was interpreted by a shell: %q", res.stdout)
	}
	if _, err := os.Stat(marker); err == nil {
		t.Fatal("payload executed: marker file was created")
	}
}

// TestPathArgNeutralizesLeadingDash keeps a relative path from being parsed as
// an option by formatters that do not support a "--" separator.
func TestPathArgNeutralizesLeadingDash(t *testing.T) {
	if got := pathArg("-rf"); got != "./-rf" {
		t.Errorf("pathArg(-rf) = %q", got)
	}
	if got := pathArg("src/main.go"); got != "./src/main.go" {
		t.Errorf("pathArg(src/main.go) = %q", got)
	}
	if got := pathArg("/abs/main.go"); got != "/abs/main.go" {
		t.Errorf("pathArg absolute = %q", got)
	}
	if got := pathArg(""); got != "" {
		t.Errorf("pathArg empty = %q", got)
	}
}
