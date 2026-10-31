package agent

import (
	"github.com/rushikeshg25/cool-code/internal/config"
	"github.com/rushikeshg25/cool-code/internal/llm"
	"testing"
)

func TestYOLOSkipsBothApprovalGates(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	cfg := config.Default()
	enabled := true
	cfg.Features.ConfirmEdits = &enabled
	for _, yolo := range []bool{false, true} {
		p, err := New(t.TempDir(), cfg, Options{AllowMissingKey: true, YOLO: yolo, Confirm: func(string) bool { t.Error("unexpected prompt"); return false }, ConfirmEdit: func(string, string) bool { t.Error("unexpected prompt"); return false }})
		if err != nil {
			t.Fatal(err)
		}
		if !yolo {
			p.confirm = nil
			p.confirmEdit = nil
		}
		for _, call := range []llm.ToolCall{{Name: "shell_command", Arguments: []byte(`{"command":"echo hello"}`)}, {Name: "edit_file", Arguments: []byte(`{"filePath":"a.go","oldString":"a","newString":"b"}`)}} {
			declined, _ := p.gate(call)
			if declined == yolo {
				t.Fatalf("YOLO=%v tool=%s declined=%v", yolo, call.Name, declined)
			}
		}
	}
}
