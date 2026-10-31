package tui

import (
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"
)

func TestStreamingFormatsActiveParagraph(t *testing.T) {
	m := resizeModel(t, newTestModel(t), 76, 24)
	m.appendDelta("Read **this text** and `go test` while the answer is still arriving")
	got := ansi.Strip(m.history[m.streamIdx].rendered)
	if strings.Contains(got, "**this text**") || strings.Contains(got, "`go test`") {
		t.Fatalf("active paragraph exposes Markdown before completion: %s", got)
	}
	if !strings.Contains(got, "this text") || !strings.Contains(got, "go test") {
		t.Fatalf("text missing: %s", got)
	}
}

func TestStreamingPreservesMarkdownAcrossBlankLines(t *testing.T) {
	for _, raw := range []string{
		"```go\nfunc main() {\n\n    println(\"hello\")\n}\n```\n",
		"1. **First**\n\n   A paragraph inside the first item.\n\n2. Second\n",
		"| Name | Value |\n| --- | --- |\n| a | **bold** |\n",
	} {
		m := resizeModel(t, newTestModel(t), 76, 24)
		for _, chunk := range strings.SplitAfter(raw, "\n") {
			m.appendDelta(chunk)
		}
		streamed := m.history[m.streamIdx].rendered
		m.finishStream(raw)
		if streamed != m.history[len(m.history)-1].rendered {
			t.Fatalf("stream changes formatting at completion for %q", raw)
		}
	}
}
func BenchmarkStreamingMarkdown(b *testing.B) {
	m := &model{width: 100}
	raw := strings.Repeat("A paragraph with **bold** and `code`.\n\n", 100)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m.renderStreaming(raw)
	}
}
