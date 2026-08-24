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
func execCommand(parent context.Context, command, dir string, timeout time.Duration) shellResult {
	if timeout == 0 {
		timeout = 30 * time.Second
	}
	if parent == nil {
		parent = context.Background()
	}
	ctx, cancel := context.WithTimeout(parent, timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, "bash", "-c", command)
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
