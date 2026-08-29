package tui

import (
	"strings"
	"sync"

	"github.com/charmbracelet/glamour"
	glamouransi "github.com/charmbracelet/glamour/ansi"
	"github.com/charmbracelet/glamour/styles"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
)

// renderers are cached per width. A single cached renderer thrashed, because
// assistant entries render at the content width while plan cards render three
// columns narrower, so alternating between them rebuilt the renderer every time.
var (
	rendererMu sync.Mutex
	renderers  = map[int]*glamour.TermRenderer{}
)

// renderMarkdown renders markdown to styled terminal text at the given width,
// caching a renderer per width. On any failure it returns the raw text.
func renderMarkdown(md string, width int) string {
	if width < 20 {
		width = 80
	}
	rendererMu.Lock()
	r, ok := renderers[width]
	if !ok {
		var err error
		r, err = glamour.NewTermRenderer(
			glamour.WithStyles(markdownStyle()),
			glamour.WithWordWrap(width),
		)
		if err != nil {
			rendererMu.Unlock()
			return strings.TrimRight(md, "\n")
		}
		// Bound the cache: widths change on resize, not per frame.
		if len(renderers) > 8 {
			clear(renderers)
		}
		renderers[width] = r
	}
	rendererMu.Unlock()

	out, err := r.Render(md)
	if err != nil {
		return strings.TrimRight(md, "\n")
	}
	return trimMarkdownPadding(out)
}

// paletteHex resolves an adaptive palette colour to the hex string Glamour
// wants, choosing the variant that matches the terminal.
func paletteHex(c lipgloss.AdaptiveColor) string {
	if lipgloss.HasDarkBackground() {
		return c.Dark
	}
	return c.Light
}

// markdownStyle drives Glamour from this program's own palette, and gives
// fenced code blocks real syntax highlighting.
//
// The base was PinkStyleConfig, which defines no CodeBlock block at all, so
// Chroma was nil and fenced code rendered as flat body text with no
// highlighting, background or frame. For a coding agent that was the most
// conspicuous gap in the interface. PinkStyleConfig also left HorizontalRule
// at "212", a hot pink that belongs to no part of this palette, and headings
// and inline code were being set to the ANSI-256 literal "99", a different
// violet from the accent and not adaptive at all.
func markdownStyle() glamouransi.StyleConfig {
	base := styles.LightStyleConfig
	if lipgloss.HasDarkBackground() {
		base = styles.DarkStyleConfig
	}
	style := base

	zero := uint(0)
	style.Document.Margin = &zero
	style.Document.BlockPrefix = ""
	style.Document.BlockSuffix = ""

	accentHex := paletteHex(accent)
	faintHex := paletteHex(faint)
	mutedHex := paletteHex(muted)

	style.Heading.Color = &accentHex
	style.Heading.Bold = boolPtr(true)
	style.H1.Prefix = "◆ "
	style.H1.Suffix = ""
	style.H1.BackgroundColor = nil
	style.H2.Prefix = "▌ "
	style.H3.Prefix = "› "
	style.H4.Prefix = "  "
	style.H5.Prefix = "  "
	style.H6.Prefix = "  "

	style.Code.BackgroundColor = nil
	style.Code.Color = &accentHex
	style.Link.Color = &accentHex
	style.LinkText.Color = &mutedHex
	style.HorizontalRule.Color = &faintHex
	style.HorizontalRule.Format = "\n─────\n"

	// Keep the inherited Chroma block, which is what performs the
	// highlighting, but drop the margin and any block background so
	// trimMarkdownPadding still sees plain trailing spaces to strip.
	style.CodeBlock.Margin = &zero
	style.CodeBlock.BackgroundColor = nil
	if style.CodeBlock.Chroma != nil {
		chroma := *style.CodeBlock.Chroma
		chroma.Background = glamouransi.StylePrimitive{}
		style.CodeBlock.Chroma = &chroma
	}

	return style
}

func boolPtr(b bool) *bool { return &b }

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
