package cmd

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/rushikeshg25/cool-code/internal/agent"
)

func TestPlainOutputSanitizesUntrustedControls(t *testing.T) {
	payload := "visible Ω\x1b]52;c;BAD\x07\x1b[2K\r\u009b2J"
	plan := &agent.TaskPlan{Goal: payload, Steps: []agent.TaskPlanStep{{Title: payload, Detail: payload}}, Assumptions: []string{payload}, Risks: []string{payload}}
	var out writerRecorder
	writePlainText(&out, payload)
	if err := printTaskPlan(&out, plan, false); err != nil {
		t.Fatal(err)
	}
	r := &printReporter{verbose: true, out: &out}
	r.Tool(payload, payload, false)
	r.Compacted(payload)
	r.Subagents([]string{payload})
	for _, bad := range []string{"\x1b", "\x07", "\r", "\u009b", "BAD"} {
		if strings.Contains(out.String(), bad) {
			t.Errorf("unsafe output %q", out.String())
		}
	}
	if !strings.Contains(out.String(), "visible Ω") {
		t.Fatal("ordinary Unicode lost")
	}
	var encoded writerRecorder
	if err := printTaskPlan(&encoded, plan, true); err != nil {
		t.Fatal(err)
	}
	var decoded agent.TaskPlan
	if err := json.Unmarshal(encoded.b, &decoded); err != nil {
		t.Fatal(err)
	}
	if decoded.Goal != payload {
		t.Fatal("JSON output changed")
	}
	if strings.Contains(encoded.String(), "\x1b") {
		t.Fatal("JSON contains raw escape")
	}
}
