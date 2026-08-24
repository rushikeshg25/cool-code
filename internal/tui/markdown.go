package tui

import (
	"strings"
	"sync"

	"github.com/charmbracelet/glamour"
	glamouransi "github.com/charmbracelet/glamour/ansi"
	"github.com/charmbracelet/glamour/styles"
	"github.com/charmbracelet/x/ansi"
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
		style := compactMarkdownStyle()
		r, err := glamour.NewTermRenderer(
			glamour.WithStyles(style),
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
	return trimMarkdownPadding(out)
}

// compactMarkdownStyle keeps the semantic formatting from Glamour without
// its default two-column document margin or literal Markdown heading markers.
func compactMarkdownStyle() glamouransi.StyleConfig {
	style := styles.PinkStyleConfig
	zero := uint(0)
	accentColor := "99"
	style.Document.Margin = &zero
	style.CodeBlock.Margin = &zero
	style.Heading.Color = &accentColor
	style.H1.Prefix = "◆ "
	style.H2.Prefix = "▌ "
	style.H3.Prefix = "› "
	style.H4.Prefix = "  "
	style.H5.Prefix = "  "
	style.H6.Prefix = "  "
	style.Code.BackgroundColor = nil
	style.Code.Color = &accentColor
	return style
}

// trimMarkdownPadding removes the full-width cell padding emitted by the
// renderer while preserving ANSI styles on the visible part of each line.
func trimMarkdownPadding(out string) string {
	lines := strings.Split(strings.Trim(out, "\n"), "\n")
	for i, line := range lines {
		plain := ansi.Strip(line)
		trimmed := strings.TrimRight(plain, " \t")
		if trimmed == "" {
			lines[i] = ""
			continue
		}
		lines[i] = ansi.Truncate(line, ansi.StringWidth(trimmed), "")
	}
	return strings.TrimRight(strings.Join(lines, "\n"), "\n")
}
