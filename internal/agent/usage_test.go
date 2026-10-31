package agent

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/rushikeshg25/cool-code/internal/llm"
	"github.com/rushikeshg25/cool-code/internal/types"
)

func usageEvents(t *testing.T, p *Processor) []usageEvent {
	t.Helper()
	path := filepath.Join(p.rootDir, "logs", p.meter.filename)
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(raw), "PRIVATE_CONTENT") {
		t.Fatal("content leaked into ledger")
	}
	stat, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if stat.Mode().Perm() != 0o600 {
		t.Fatalf("mode=%v", stat.Mode())
	}
	var events []usageEvent
	for _, line := range strings.Split(strings.TrimSpace(string(raw)), "\n") {
		var e usageEvent
		if err := json.Unmarshal([]byte(line), &e); err != nil {
			t.Fatal(err)
		}
		events = append(events, e)
	}
	return events
}

func TestUsageLedgerToolAttribution(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "input.txt")
	if err := os.WriteFile(path, []byte("PRIVATE_CONTENT"), 0o600); err != nil {
		t.Fatal(err)
	}
	provider := &fakeProvider{responses: []llm.Message{
		{Usage: llm.Usage{Input: 100, Output: 10, Reported: true}, ToolCalls: []llm.ToolCall{toolCall("PRIVATE_CONTENT", "read_file", map[string]any{"absolutePath": path})}},
		{Text: "PRIVATE_CONTENT", Usage: llm.Usage{Input: 120, Output: 5, Reported: true}},
	}}
	p := newTestProcessor(t, root, provider, types.ModeAgent)
	if _, err := p.ProcessQuery(context.Background(), "PRIVATE_CONTENT", nil); err != nil {
		t.Fatal(err)
	}
	events := usageEvents(t, p)
	if len(events) != 3 || events[0].Type != "request" || events[1].Type != "tool" || events[2].Type != "request" {
		t.Fatalf("events=%+v", events)
	}
	if events[1].RequestID != events[0].ID || events[1].Tool != "read_file" || events[1].Status != "completed" || events[1].Finished.Before(events[1].Started) {
		t.Fatalf("tool=%+v", events[1])
	}
	if len(events[2].InputTools) != 1 || events[2].InputTools[0] != "read_file" {
		t.Fatal("tool-result attribution missing")
	}
	if !strings.Contains(p.UsageReport(), "220 input · 15 output") {
		t.Fatal(p.UsageReport())
	}
}

type usageFailureProvider struct{}

func (usageFailureProvider) Name() string  { return "fake" }
func (usageFailureProvider) Model() string { return "model" }
func (usageFailureProvider) Complete(context.Context, llm.Request) (llm.Message, error) {
	return llm.Message{}, errors.New("PRIVATE_CONTENT")
}

func TestUsageMissingAndFailure(t *testing.T) {
	p := newTestProcessor(t, t.TempDir(), usageFailureProvider{}, types.ModeAgent)
	_, _, err := p.completeMetered(context.Background(), "main", p.provider, llm.Request{}, nil)
	if err == nil {
		t.Fatal("lost provider error")
	}
	events := usageEvents(t, p)
	if events[0].UsageKnown || events[0].Usage != nil || events[0].Status != "failed" {
		t.Fatalf("event=%+v", events[0])
	}
	if !strings.Contains(p.UsageReport(), "1 usage unavailable") {
		t.Fatal(p.UsageReport())
	}
}

func TestUsageConcurrentLedger(t *testing.T) {
	p := newTestProcessor(t, t.TempDir(), nil, types.ModeAgent)
	var wg sync.WaitGroup
	for i := 0; i < 40; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			p.recordUsage(usageEvent{Type: "request", Kind: "btw", Started: time.Now(), UsageKnown: true, Usage: &llm.Usage{Input: 1, Output: 2}})
			_ = p.UsageReport()
		}()
	}
	wg.Wait()
	events := usageEvents(t, p)
	if len(events) != 40 || len(p.meter.recent) != 20 || p.meter.totals["btw"].input != 40 {
		t.Fatal("lost concurrent usage")
	}
}

func TestUsageLogFailureDoesNotFailRequest(t *testing.T) {
	root := t.TempDir()
	outside := t.TempDir()
	if err := os.Symlink(outside, filepath.Join(root, "logs")); err != nil {
		t.Skip(err)
	}
	p := newTestProcessor(t, root, &fakeProvider{}, types.ModeAgent)
	if _, err := p.ProcessQuery(context.Background(), "hello", nil); err != nil {
		t.Fatal(err)
	}
	entries, err := os.ReadDir(outside)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 0 || !strings.Contains(p.UsageReport(), "Log write failed") {
		t.Fatal(p.UsageReport())
	}
}

func TestUsageSubagentAndCompaction(t *testing.T) {
	provider := &fakeProvider{responses: []llm.Message{{Text: "explored", Usage: llm.Usage{Input: 7, Output: 3}}, {Text: "summary", Usage: llm.Usage{Input: 8, Output: 4}}}}
	p := newTestProcessor(t, t.TempDir(), provider, types.ModeAgent)
	p.runSubagent(context.Background(), "explore", func(string) {})
	for i := 0; i < 40; i++ {
		p.ctxMgr.addUser("message")
	}
	p.Compact(context.Background())
	events := usageEvents(t, p)
	if len(events) != 2 || events[0].Kind != "subagent" || events[1].Kind != "compaction" || p.totalUsage.Input != 15 || p.lastUsage.Input != 0 {
		t.Fatalf("events=%+v total=%+v", events, p.totalUsage)
	}
}
