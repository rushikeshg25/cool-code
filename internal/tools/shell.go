package tools

import (
	"bytes"
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/rushikeshg25/cool-code/internal/security"
)

type shellResult struct {
	stdout   string
	stderr   string
	exitCode int
	success  bool
	errMsg   string
}

// execCommand runs a shell command via `bash -c` in dir (or root) with a
// timeout, capturing stdout/stderr. The parent context cancels the command
// early when the turn is aborted.
//
// The command string is interpreted by bash, so this may only be used for
// commands the user has confirmed. Anything assembled from model- or
// repository-controlled values must go through execArgv instead.
func execCommand(parent context.Context, command, dir string, timeout time.Duration) shellResult {
	return runCommand(parent, dir, timeout, "bash", "-c", command)
}

// execArgv runs a program directly, with no shell in between, so caller
// supplied arguments can never be reinterpreted as shell syntax.
func execArgv(parent context.Context, dir string, timeout time.Duration, name string, args ...string) shellResult {
	return runCommand(parent, dir, timeout, name, args...)
}

func runCommand(parent context.Context, dir string, timeout time.Duration, name string, args ...string) shellResult {
	if timeout <= 0 {
		timeout = defaultCommandTimeout
	}
	if parent == nil {
		parent = context.Background()
	}
	ctx, cancel := context.WithTimeout(parent, timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Env = safeCommandEnv()
	if dir != "" {
		cmd.Dir, _ = filepath.Abs(dir)
	}
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	res := shellResult{
		stdout: security.Redact(stdout.String()),
		stderr: security.Redact(stderr.String()),
	}
	if ctx.Err() == context.DeadlineExceeded {
		res.exitCode = -1
		res.success = false
		res.errMsg = "Command timed out after " + timeout.String()
		return res
	}
	if ctx.Err() == context.Canceled {
		res.exitCode = -1
		res.success = false
		res.errMsg = "Command cancelled"
		return res
	}
	if err != nil {
		res.success = false
		if exitErr, ok := err.(*exec.ExitError); ok {
			res.exitCode = exitErr.ExitCode()
		} else {
			res.exitCode = -1
		}
		res.errMsg = err.Error()
		return res
	}
	res.success = true
	res.exitCode = 0
	return res
}

func safeCommandEnv() []string {
	allowed := map[string]bool{
		"PATH": true, "HOME": true, "TMPDIR": true, "TMP": true, "TEMP": true,
		"USER": true, "LOGNAME": true, "SHELL": true, "LANG": true, "TERM": true,
		"COLORTERM": true, "CI": true, "GOPATH": true, "GOROOT": true,
		"GOMODCACHE": true, "GOCACHE": true, "GOENV": true, "CARGO_HOME": true,
		"RUSTUP_HOME": true, "PNPM_HOME": true, "BUN_INSTALL": true,
		"PYENV_ROOT": true, "VIRTUAL_ENV": true, "GIT_AUTHOR_NAME": true,
		"GIT_AUTHOR_EMAIL": true, "GIT_COMMITTER_NAME": true, "GIT_COMMITTER_EMAIL": true,
	}
	var env []string
	for _, item := range os.Environ() {
		name, _, _ := strings.Cut(item, "=")
		if allowed[name] || strings.HasPrefix(name, "LC_") {
			env = append(env, item)
		}
	}
	return env
}

// combined merges stdout with a STDERR section, matching the TS git tools.
func (r shellResult) combined() string {
	out := r.stdout
	if r.stderr != "" {
		out += "\nSTDERR:\n" + r.stderr
	}
	return out
}

// Command timeouts. The default stays short so a hung command does not stall a
// turn, but it used to be the only option: shell_command always passed 0, so
// 30 seconds was a hard ceiling and builds, installs and long test runs were
// simply impossible.
const (
	defaultCommandTimeout = 30 * time.Second
	maxCommandTimeout     = 30 * time.Minute
)

// commandTimeout converts a caller-supplied number of seconds into a bounded
// duration. Zero or less selects the default.
func commandTimeout(seconds int) time.Duration {
	if seconds <= 0 {
		return defaultCommandTimeout
	}
	d := time.Duration(seconds) * time.Second
	if d > maxCommandTimeout {
		return maxCommandTimeout
	}
	return d
}
