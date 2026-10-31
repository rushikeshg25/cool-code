package tools

import (
	"context"
	"encoding/json"
	"errors"
	"os/exec"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/rushikeshg25/cool-code/internal/security"
	"github.com/rushikeshg25/cool-code/internal/types"
)

const backgroundLogLimit = 64 * 1024

// BackgroundProcesses owns commands for this Coolcode instance, across turns.
// IDs are opaque and never interpreted as operating-system PIDs.
type BackgroundProcesses struct {
	mu     sync.Mutex
	jobs   map[string]*backgroundJob
	closed bool
}
type backgroundJob struct {
	mu                            sync.Mutex
	id, command, directory, state string
	started                       time.Time
	exitCode                      *int
	failure                       string
	cancel                        context.CancelFunc
	done                          chan struct{}
	log                           processLog
}

type backgroundArgs struct {
	Action    string `json:"action"`
	Command   string `json:"command"`
	Directory string `json:"directory"`
	ID        string `json:"id"`
	Timeout   int    `json:"timeout"`
	Tail      int    `json:"tail"`
}
type processStatus struct {
	ID        string `json:"id"`
	Command   string `json:"command"`
	Directory string `json:"directory"`
	State     string `json:"state"`
	StartedAt string `json:"startedAt"`
	ExitCode  *int   `json:"exitCode,omitempty"`
	Error     string `json:"error,omitempty"`
	Logs      string `json:"logs,omitempty"`
}

var backgroundProcessTool = Tool{
	Name: "background_process", Mutating: true,
	Description: "Manage long-running commands across turns: start, list, status, logs, stop, or forget completed jobs. Start returns an opaque process ID immediately, not readiness. Read logs and check the app before claiming success. Run foreground commands (no &, nohup, daemonizing or detaching); Coolcode manages the background lifetime. Commands stop on explicit stop, timeout, or Coolcode exit. Only this instance's jobs are accessible. Linux/macOS only. Requires implementation mode; start/stop use normal action approvals.",
	Schema: obj(map[string]any{
		"action":  map[string]any{"type": "string", "enum": []string{"start", "list", "status", "logs", "stop", "forget"}},
		"command": strProp("Exact bash command for start."), "directory": strProp("Working directory for start, relative to project root or absolute within granted roots."),
		"id":      strProp("Opaque process ID returned by start, required except for start/list."),
		"timeout": numProp("Lifetime in seconds for start; default 1800, maximum 86400."),
		"tail":    numProp("Lines to return for logs, default 100, maximum 500. Only complete lines are shown until exit; the latest 64 KiB are retained."),
	}, "action"),
	Execute: func(ctx Context, raw json.RawMessage) types.ToolResult {
		var a backgroundArgs
		if err := json.Unmarshal(raw, &a); err != nil {
			return fail("Invalid process arguments", err.Error())
		}
		if ctx.Processes == nil {
			return fail("Background processes unavailable", "This agent has no background process manager.")
		}
		var result any
		var err error
		switch a.Action {
		case "start":
			if a.Timeout < 0 || a.Timeout > 86400 {
				return fail("Invalid lifetime", "timeout must be between 0 and 86400 seconds")
			}
			prepared, e := PrepareCommand(ctx, "background_process", raw)
			if e != nil {
				return fail("Command blocked", e.Error())
			}
			if e = json.Unmarshal(prepared, &a); e != nil {
				return fail("Invalid command", e.Error())
			}
			result, err = ctx.Processes.start(ctx.Context(), a)
		case "list":
			result = ctx.Processes.list()
		case "status", "logs", "stop", "forget":
			if a.ID == "" {
				return fail("Missing process ID", "Use an ID returned by start or list.")
			}
			if a.Tail < 0 || a.Tail > 500 {
				return fail("Invalid tail", "tail must be between 0 and 500")
			}
			result, err = ctx.Processes.inspect(a)
		default:
			return fail("Invalid action", "Use start, list, status, logs, stop, or forget.")
		}
		if err != nil {
			return fail("Background process failed", security.Redact(err.Error()))
		}
		data, err := json.MarshalIndent(result, "", "  ")
		if err != nil {
			return fail("Process result failed", err.Error())
		}
		return types.ToolResult{Display: "Background process: " + a.Action, LLMResult: Untrusted("background process output", security.Redact(string(data)))}
	},
}

func (m *BackgroundProcesses) start(caller context.Context, a backgroundArgs) (processStatus, error) {
	if err := caller.Err(); err != nil {
		return processStatus{}, err
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.closed {
		return processStatus{}, errors.New("process manager is closed")
	}
	if m.jobs == nil {
		m.jobs = make(map[string]*backgroundJob)
	}
	if len(m.jobs) >= 64 {
		return processStatus{}, errors.New("64 jobs are tracked; forget completed jobs before starting another")
	}
	timeout := a.Timeout
	if timeout == 0 {
		timeout = 1800
	}
	// Deliberately independent of the current turn: a server must survive its
	// start call, subsequent inference requests and turn cancellation.
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
	cmd := exec.CommandContext(ctx, "bash", "-c", a.Command)
	cmd.Dir = a.Directory
	cmd.Env = safeCommandEnv()
	if err := configureBackgroundProcess(cmd); err != nil {
		cancel()
		return processStatus{}, err
	}
	// On cancellation, kill the owned process group before Wait closes pipes.
	cmd.WaitDelay = time.Second
	job := &backgroundJob{id: uuid.NewString(), command: a.Command, directory: a.Directory, state: "running", started: time.Now().UTC(), cancel: cancel, done: make(chan struct{})}
	cmd.Stdout = &job.log
	cmd.Stderr = &job.log
	if err := cmd.Start(); err != nil {
		cancel()
		return processStatus{}, err
	}
	m.jobs[job.id] = job
	go func() {
		err := cmd.Wait()
		if errors.Is(err, exec.ErrWaitDelay) {
			_ = cmd.Cancel()
		}
		job.log.finish()
		job.mu.Lock()
		code := cmd.ProcessState.ExitCode()
		job.exitCode = &code
		switch {
		case errors.Is(ctx.Err(), context.DeadlineExceeded):
			job.state = "timed_out"
		case ctx.Err() != nil:
			job.state = "stopped"
		case err != nil:
			job.state = "failed"
		default:
			job.state = "exited"
		}
		if err != nil {
			job.failure = err.Error()
		}
		job.mu.Unlock()
		cancel()
		close(job.done)
	}()
	if err := caller.Err(); err != nil {
		cancel()
		<-job.done
		return processStatus{}, err
	}
	return job.status(false, 0), nil
}
func (j *backgroundJob) status(logs bool, tail int) processStatus {
	j.mu.Lock()
	status := processStatus{ID: j.id, Command: j.command, Directory: j.directory, State: j.state, StartedAt: j.started.Format(time.RFC3339Nano), ExitCode: j.exitCode, Error: j.failure}
	j.mu.Unlock()
	if logs {
		status.Logs = j.log.read(tail)
	}
	return status
}
func (m *BackgroundProcesses) list() []processStatus {
	m.mu.Lock()
	defer m.mu.Unlock()
	result := make([]processStatus, 0, len(m.jobs))
	for _, j := range m.jobs {
		result = append(result, j.status(false, 0))
	}
	sort.Slice(result, func(i, j int) bool { return result[i].StartedAt < result[j].StartedAt })
	return result
}
func (m *BackgroundProcesses) inspect(a backgroundArgs) (processStatus, error) {
	m.mu.Lock()
	job := m.jobs[a.ID]
	m.mu.Unlock()
	if job == nil {
		return processStatus{}, errors.New("unknown process ID; only jobs started by this Coolcode instance can be managed")
	}
	switch a.Action {
	case "stop":
		job.cancel()
		<-job.done
	case "forget":
		select {
		case <-job.done:
			m.mu.Lock()
			delete(m.jobs, a.ID)
			m.mu.Unlock()
		default:
			return processStatus{}, errors.New("stop the running job before forgetting it")
		}
	}
	return job.status(a.Action == "logs", a.Tail), nil
}
func (m *BackgroundProcesses) Close() {
	if m == nil {
		return
	}
	m.mu.Lock()
	m.closed = true
	var jobs []*backgroundJob
	for _, job := range m.jobs {
		job.cancel()
		jobs = append(jobs, job)
	}
	m.mu.Unlock()
	for _, job := range jobs {
		<-job.done
	}
}

// processLog retains complete lines, so neither chunk boundaries nor rotation
// can cut off a secret's label before redaction. Oversized lines are omitted.
type processLog struct {
	mu                  sync.Mutex
	complete, partial   string
	dropping, truncated bool
}

func (l *processLog) Write(p []byte) (int, error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	text := string(p)
	for len(text) > 0 {
		part, rest, newline := strings.Cut(text, "\n")
		if !l.dropping {
			if len(l.partial)+len(part) > backgroundLogLimit {
				l.partial = ""
				l.dropping = true
			} else {
				l.partial += part
			}
		}
		if newline {
			if l.dropping {
				l.appendLine("[Oversized output line omitted]\n")
			} else {
				l.appendLine(l.partial + "\n")
			}
			l.partial = ""
			l.dropping = false
		}
		if !newline {
			break
		}
		text = rest
	}
	return len(p), nil
}
func (l *processLog) appendLine(line string) {
	l.complete += line
	for len(l.complete) > backgroundLogLimit {
		at := strings.IndexByte(l.complete, '\n')
		if at < 0 {
			l.complete = ""
		} else {
			l.complete = l.complete[at+1:]
		}
		l.truncated = true
	}
}
func (l *processLog) finish() {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.dropping {
		l.appendLine("[Oversized output line omitted]\n")
	} else if l.partial != "" {
		l.appendLine(l.partial + "\n")
	}
	l.partial = ""
	l.dropping = false
}
func (l *processLog) read(tail int) string {
	l.mu.Lock()
	defer l.mu.Unlock()
	if tail == 0 {
		tail = 100
	}
	lines := strings.Split(strings.TrimSuffix(l.complete, "\n"), "\n")
	omitted := l.truncated
	if len(lines) > tail {
		lines = lines[len(lines)-tail:]
		omitted = true
	}
	text := security.Redact(strings.Join(lines, "\n"))
	if omitted {
		text = "[Earlier output omitted]\n" + text
	}
	if len(text) > webMaxChars {
		text = truncateRunes(text, webMaxChars) + "\n[Output truncated; request fewer tail lines.]"
	}
	return text
}
