//go:build linux || darwin

package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func waitProcess(t *testing.T, m *BackgroundProcesses, id string, want func(processStatus) bool) processStatus {
	t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		s, err := m.inspect(backgroundArgs{Action: "logs", ID: id})
		if err != nil {
			t.Fatal(err)
		}
		if want(s) {
			return s
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatal("process did not reach expected state")
	return processStatus{}
}
func TestBackgroundHelperProcess(t *testing.T) {
	if len(os.Args) == 0 || os.Args[len(os.Args)-1] != "--coolcode-background-fixture" {
		return
	}
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		os.Exit(2)
	}
	fmt.Println("LISTEN=" + listener.Addr().String())
	server := http.Server{ReadHeaderTimeout: time.Second, Handler: http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { _, _ = io.WriteString(w, "ready") })}
	_ = server.Serve(listener)
	os.Exit(0)
}
func TestBackgroundServerSurvivesTurnAndStops(t *testing.T) {
	m := &BackgroundProcesses{}
	defer m.Close()
	executable, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	command := "'" + strings.ReplaceAll(executable, "'", "'\\''") + "' -test.run=^TestBackgroundHelperProcess$ -- --coolcode-background-fixture"
	ctx, cancel := context.WithCancel(context.Background())
	job, err := m.start(ctx, backgroundArgs{Command: command, Directory: t.TempDir()})
	if err != nil {
		t.Fatal(err)
	}
	cancel() // End of the originating turn must not stop an acknowledged job.
	status := waitProcess(t, m, job.ID, func(s processStatus) bool { return strings.Contains(s.Logs, "LISTEN=") })
	_, address, _ := strings.Cut(status.Logs, "LISTEN=")
	address = strings.Fields(address)[0]
	client := http.Client{Timeout: time.Second, Transport: &http.Transport{Proxy: nil}}
	response, err := client.Get("http://" + address)
	if err != nil {
		t.Fatal(err)
	}
	data, _ := io.ReadAll(response.Body)
	response.Body.Close()
	if string(data) != "ready" {
		t.Fatal(string(data))
	}
	stopped, err := m.inspect(backgroundArgs{Action: "stop", ID: job.ID})
	if err != nil || stopped.State != "stopped" {
		t.Fatalf("%+v %v", stopped, err)
	}
	if _, err := client.Get("http://" + address); err == nil {
		t.Fatal("server still reachable after stop")
	}
	if _, err := m.inspect(backgroundArgs{Action: "status", ID: "1"}); err == nil {
		t.Fatal("accepted arbitrary PID")
	}
	if _, err := m.inspect(backgroundArgs{Action: "forget", ID: job.ID}); err != nil {
		t.Fatal(err)
	}
	if len(m.list()) != 0 {
		t.Fatal("completed job not forgotten")
	}
}
func TestBackgroundTimeoutExitAndClose(t *testing.T) {
	m := &BackgroundProcesses{}
	defer m.Close()
	dir := t.TempDir()
	done, err := m.start(context.Background(), backgroundArgs{Command: "printf finished; exit 7", Directory: dir})
	if err != nil {
		t.Fatal(err)
	}
	result := waitProcess(t, m, done.ID, func(s processStatus) bool { return s.State != "running" })
	if result.State != "failed" || result.ExitCode == nil || *result.ExitCode != 7 || result.Logs != "finished" {
		t.Fatalf("%+v", result)
	}
	timed, err := m.start(context.Background(), backgroundArgs{Command: "sleep 30", Directory: dir, Timeout: 1})
	if err != nil {
		t.Fatal(err)
	}
	result = waitProcess(t, m, timed.ID, func(s processStatus) bool { return s.State != "running" })
	if result.State != "timed_out" {
		t.Fatal(result)
	}
	active, err := m.start(context.Background(), backgroundArgs{Command: "sleep 30", Directory: dir})
	if err != nil {
		t.Fatal(err)
	}
	m.Close()
	result, _ = m.inspect(backgroundArgs{Action: "status", ID: active.ID})
	if result.State != "stopped" {
		t.Fatal(result)
	}
	if _, err := m.start(context.Background(), backgroundArgs{Command: "true", Directory: dir}); err == nil {
		t.Fatal("start after close")
	}
}
func TestBackgroundLogsBoundedAndRedacted(t *testing.T) {
	var log processLog
	_, _ = log.Write([]byte("password=split"))
	if got := log.read(100); got != "" {
		t.Fatalf("exposed partial line: %s", got)
	}
	_, _ = log.Write([]byte("-secret\n"))
	if got := log.read(100); strings.Contains(got, "split-secret") || !strings.Contains(got, "[REDACTED]") {
		t.Fatal(got)
	}
	_, _ = log.Write([]byte(strings.Repeat("x", backgroundLogLimit*3) + "\nlast\n"))
	if len(log.complete) > backgroundLogLimit || log.read(1) != "[Earlier output omitted]\nlast" {
		t.Fatal("unbounded log or wrong tail")
	}
}
func TestBackgroundApprovalAndPathValidation(t *testing.T) {
	root := t.TempDir()
	m := &BackgroundProcesses{}
	defer m.Close()
	ctx := Context{RootDir: root, Processes: m}
	raw := json.RawMessage(`{"action":"start","command":"printf ok","directory":"."}`)
	args, err := PrepareCommand(ctx, "background_process", raw)
	if err != nil {
		t.Fatal(err)
	}
	var a backgroundArgs
	_ = json.Unmarshal(args, &a)
	if a.Directory != root || DangerReason("background_process", args) == "" {
		t.Fatal("approval not bound to directory")
	}
	for _, bad := range []string{`{"action":"start","command":"echo password=literal-secret"}`, `{"action":"start","command":"true","directory":".."}`} {
		if _, err := PrepareCommand(ctx, "background_process", json.RawMessage(bad)); err == nil {
			t.Fatal("accepted unsafe start")
		}
	}
	if !RunReadOnly(ctx, "background_process", raw).Failed {
		t.Fatal("read-only bypass")
	}
	if result := Run(ctx, "background_process", json.RawMessage(`{"action":"start","command":"true","timeout":86401}`)); !result.Failed {
		t.Fatal("unbounded lifetime")
	}
	t.Setenv("TEST_PRIVATE_TOKEN", "a-long-private-token")
	job, err := m.start(context.Background(), backgroundArgs{Command: "printenv TEST_PRIVATE_TOKEN", Directory: root})
	if err != nil {
		t.Fatal(err)
	}
	status := waitProcess(t, m, job.ID, func(s processStatus) bool { return s.State != "running" })
	if status.ExitCode == nil || *status.ExitCode != 1 || status.Logs != "" {
		t.Fatal("secret environment inherited")
	}
	// A non-directory path must fail before a job is registered.
	file := filepath.Join(root, "file")
	_ = os.WriteFile(file, nil, 0600)
	if _, err := m.start(context.Background(), backgroundArgs{Command: "true", Directory: file}); err == nil {
		t.Fatal("accepted file as cwd")
	}
}
func TestBackgroundStopKillsProcessGroup(t *testing.T) {
	m := &BackgroundProcesses{}
	defer m.Close()
	job, err := m.start(context.Background(), backgroundArgs{Command: "sleep 30 & echo CHILD=$!; wait", Directory: t.TempDir()})
	if err != nil {
		t.Fatal(err)
	}
	status := waitProcess(t, m, job.ID, func(s processStatus) bool { return strings.Contains(s.Logs, "CHILD=") })
	_, pid, _ := strings.Cut(status.Logs, "CHILD=")
	pid = strings.TrimSpace(pid)
	if _, err := m.inspect(backgroundArgs{Action: "stop", ID: job.ID}); err != nil {
		t.Fatal(err)
	}
	// Killed children may briefly be zombies pending init reaping; they are no longer running.
	out, _ := exec.Command("ps", "-o", "stat=", "-p", pid).Output()
	if state := strings.TrimSpace(string(out)); state != "" && !strings.HasPrefix(state, "Z") {
		t.Fatalf("child remains active: %s", state)
	}
}
