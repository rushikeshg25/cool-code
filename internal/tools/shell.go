package tools

import (
	"bytes"
	"context"
	"os/exec"
	"path/filepath"
	"time"
)

type shellResult struct {
	stdout   string
	stderr   string
	exitCode int
	success  bool
	errMsg   string
}

// execCommand runs a shell command via `bash -c` in dir (or root) with a
// timeout, capturing stdout/stderr.
func execCommand(command, dir string, timeout time.Duration) shellResult {
	if timeout == 0 {
		timeout = 30 * time.Second
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, "bash", "-c", command)
	if dir != "" {
		cmd.Dir, _ = filepath.Abs(dir)
	}
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	res := shellResult{
		stdout: stdout.String(),
		stderr: stderr.String(),
	}
	if ctx.Err() == context.DeadlineExceeded {
		res.exitCode = -1
		res.success = false
		res.errMsg = "Command timed out after " + timeout.String()
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

// combined merges stdout with a STDERR section, matching the TS git tools.
func (r shellResult) combined() string {
	out := r.stdout
	if r.stderr != "" {
		out += "\nSTDERR:\n" + r.stderr
	}
	return out
}
