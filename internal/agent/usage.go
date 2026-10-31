package agent

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/rushikeshg25/cool-code/internal/llm"
	"github.com/rushikeshg25/cool-code/internal/security"
	"github.com/rushikeshg25/cool-code/internal/tools"
	"github.com/rushikeshg25/cool-code/internal/types"
)

// usageEvent contains metadata only: never prompts, arguments, results, or errors.
// A request spans the provider call, including any transport retries. Its usage
// is only what the provider reported, not an estimate of invisible retry costs.
type usageEvent struct {
	Type       string     `json:"type"`
	ID         string     `json:"id"`
	RequestID  string     `json:"request_id,omitempty"`
	Kind       string     `json:"kind"`
	Provider   string     `json:"provider,omitempty"`
	Model      string     `json:"model,omitempty"`
	Started    time.Time  `json:"started_at"`
	Finished   time.Time  `json:"finished_at"`
	DurationMS int64      `json:"duration_ms"`
	Status     string     `json:"status"`
	Usage      *llm.Usage `json:"usage,omitempty"`
	UsageKnown bool       `json:"usage_known"`
	Tools      []string   `json:"requested_tools,omitempty"`
	InputTools []string   `json:"input_tool_results,omitempty"`
	Tool       string     `json:"tool,omitempty"`
	ToolIndex  *int       `json:"tool_index,omitempty"`
}

type usageTotal struct{ requests, unknown, input, output int }
type usageMeter struct {
	mu       sync.Mutex
	filename string
	created  bool
	writeErr error
	recent   []usageEvent
	totals   map[string]usageTotal
}

func meterToolName(name string) string {
	if _, ok := tools.Lookup(name); ok || name == spawnAgentTool || name == updateTaskListTool {
		return name
	}
	return "unknown_tool"
}

func (p *Processor) recordUsage(e usageEvent) {
	m := &p.meter
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.filename == "" {
		m.filename = "token-usage-dev-" + uuid.NewString() + ".jsonl"
	}
	if e.Type == "request" {
		if m.totals == nil {
			m.totals = make(map[string]usageTotal)
		}
		t := m.totals[e.Kind]
		t.requests++
		if e.UsageKnown {
			t.input += e.Usage.Input
			t.output += e.Usage.Output
		} else {
			t.unknown++
		}
		m.totals[e.Kind] = t
	}
	m.recent = append(m.recent, e)
	if len(m.recent) > 20 {
		m.recent = append([]usageEvent(nil), m.recent[len(m.recent)-20:]...)
	}
	// Fail visibly but do not fail a model/tool action if logging is unavailable.
	if m.writeErr == nil {
		m.writeErr = m.append(p.rootDir, e)
	}
}

func (m *usageMeter) append(rootDir string, e usageEvent) error {
	root, err := os.OpenRoot(rootDir)
	if err != nil {
		return err
	}
	defer root.Close()
	if err = root.Mkdir("logs", 0o700); err != nil && !errors.Is(err, os.ErrExist) {
		return err
	}
	// OpenRoot prevents a logs symlink from redirecting the ledger outside the project.
	flags := os.O_WRONLY | os.O_APPEND
	if !m.created {
		flags |= os.O_CREATE | os.O_EXCL
	}
	f, err := root.OpenFile(filepath.Join("logs", m.filename), flags, 0o600)
	if err != nil {
		return err
	}
	m.created = true
	err = json.NewEncoder(f).Encode(e)
	return errors.Join(err, f.Close())
}

func (p *Processor) completeMetered(ctx context.Context, kind string, provider llm.Provider, req llm.Request, onDelta func(string)) (llm.Message, string, error) {
	start := time.Now()
	id := uuid.NewString()
	var resp llm.Message
	var err error
	if streamer, ok := provider.(llm.Streamer); ok && onDelta != nil {
		resp, err = streamer.Stream(ctx, req, onDelta)
	} else {
		resp, err = provider.Complete(ctx, req)
	}
	end := time.Now()
	status := "completed"
	if err != nil {
		status = "failed"
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			status = "cancelled"
		}
	}
	known := resp.Usage.Reported || resp.Usage.Input > 0 || resp.Usage.Output > 0
	e := usageEvent{Type: "request", ID: id, Kind: kind, Provider: security.Redact(provider.Name()), Model: security.Redact(provider.Model()), Started: start.UTC(), Finished: end.UTC(), DurationMS: end.Sub(start).Milliseconds(), Status: status, UsageKnown: known}
	if known {
		u := resp.Usage
		e.Usage = &u
	}
	for _, call := range resp.ToolCalls {
		e.Tools = append(e.Tools, meterToolName(call.Name))
	}
	for _, msg := range req.Messages {
		if msg.Role == llm.RoleTool {
			e.InputTools = append(e.InputTools, meterToolName(msg.ToolName))
		}
	}
	p.recordUsage(e)
	if known {
		p.mu.Lock()
		if kind == "main" {
			p.lastUsage = resp.Usage
		}
		p.totalUsage.Input += resp.Usage.Input
		p.totalUsage.Output += resp.Usage.Output
		p.mu.Unlock()
	}
	return resp, id, err
}

func (p *Processor) runMeteredTool(toolCtx tools.Context, kind, requestID string, index int, call llm.ToolCall, readOnly bool) types.ToolResult {
	start := time.Now()
	var result types.ToolResult
	if readOnly {
		result = tools.RunReadOnly(toolCtx, call.Name, call.Arguments)
	} else {
		result = tools.Run(toolCtx, call.Name, call.Arguments)
	}
	status := "completed"
	if result.Failed {
		status = "failed"
	}
	if toolCtx.Ctx != nil && toolCtx.Ctx.Err() != nil {
		status = "cancelled"
	}
	p.recordTool(kind, requestID, index, call.Name, start, status)
	return result
}

func (p *Processor) recordTool(kind, requestID string, index int, name string, start time.Time, status string) {
	end := time.Now()
	p.recordUsage(usageEvent{Type: "tool", ID: uuid.NewString(), RequestID: requestID, Kind: kind, Started: start.UTC(), Finished: end.UTC(), DurationMS: end.Sub(start).Milliseconds(), Status: status, Tool: meterToolName(name), ToolIndex: &index})
}

// UsageReport covers this processor's lifetime, including /fork and restored chats.
func (p *Processor) UsageReport() string {
	m := &p.meter
	m.mu.Lock()
	defer m.mu.Unlock()
	var b strings.Builder
	b.WriteString("Dev token usage for this running instance (all chats)\n")
	for _, kind := range []string{"main", "subagent", "btw", "compaction"} {
		if t, ok := m.totals[kind]; ok {
			fmt.Fprintf(&b, "\n%s: %d requests · %d input · %d output · %d usage unavailable", kind, t.requests, t.input, t.output, t.unknown)
		}
	}
	b.WriteString("\n\nRecent events (UTC):")
	for _, e := range m.recent {
		fmt.Fprintf(&b, "\n%s %s %s · %d ms · %s", e.Started.Format(time.RFC3339Nano), e.Kind, e.Type, e.DurationMS, e.Status)
		if e.Type == "tool" {
			fmt.Fprintf(&b, " · %s", e.Tool)
			continue
		}
		fmt.Fprintf(&b, " · %s/%s", e.Provider, e.Model)
		if e.UsageKnown {
			fmt.Fprintf(&b, " · input %d / output %d", e.Usage.Input, e.Usage.Output)
			if e.Usage.CachedInput != nil {
				fmt.Fprintf(&b, " · cached input %d", *e.Usage.CachedInput)
			}
			if e.Usage.Reasoning != nil {
				fmt.Fprintf(&b, " · reasoning %d", *e.Usage.Reasoning)
			}
		} else {
			b.WriteString(" · usage unavailable")
		}
		if len(e.Tools) > 0 {
			fmt.Fprintf(&b, " · tools: %s", strings.Join(e.Tools, ", "))
		}
	}
	if m.filename != "" {
		fmt.Fprintf(&b, "\n\nLog: %s", filepath.Join(p.rootDir, "logs", m.filename))
	} else {
		b.WriteString("\nNo requests recorded yet.")
	}
	if m.writeErr != nil {
		fmt.Fprintf(&b, "\nLog write failed; subsequent records are only kept in memory (latest 20): %s", security.Redact(m.writeErr.Error()))
	}
	b.WriteString("\n\nCounts are provider-reported per request. Cached input and reasoning are subsets, not extra tokens. Exact per-tool tokens, hidden retries, billing, and remaining subscription quota are unavailable.")
	return security.SanitizeTerminal(b.String())
}
