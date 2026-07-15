package tui

import (
	"strings"
	"sync"

	"github.com/charmbracelet/glamour"
)

var (
	rendererMu    sync.Mutex
	renderer      *glamour.TermRenderer
	rendererWidth int
)

// renderMarkdown renders markdown to styled terminal text at the given width,
// caching the renderer per width. On any failure it returns the raw text.
func renderMarkdown(md string, width int) string {
	if width < 20 {
		width = 80
	}
	rendererMu.Lock()
	defer rendererMu.Unlock()
	if renderer == nil || rendererWidth != width {
		r, err := glamour.NewTermRenderer(
			glamour.WithAutoStyle(),
			glamour.WithWordWrap(width),
		)
		if err != nil {
			return strings.TrimRight(md, "\n")
		}
		renderer = r
		rendererWidth = width
	}
	out, err := renderer.Render(md)
	if err != nil {
		return strings.TrimRight(md, "\n")
	}
	return strings.TrimRight(out, "\n")
}
