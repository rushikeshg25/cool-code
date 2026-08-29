package tui

// Layout is computed once per frame from the terminal size and the current
// overlay state, and every region measures itself from it.
//
// The previous arrangement had View call footer, and viewportHeight call footer
// again to measure it, so the whole footer was built twice per frame. Worse,
// building it mutated the model: renderConfirmation clamped m.confirmOff as a
// side effect of being rendered. Deciding the geometry up front removes both.

// Width breakpoints. The sidebar is a function of width, not a mode the user
// selects, so a narrow terminal degrades to the single column it can afford.
const (
	// sidebarMinWidth is the total width at which a sidebar starts to pay for
	// itself. Below it the transcript needs every column.
	sidebarMinWidth = 100
	// taskListMinWidth is where a full task list still fits under the
	// transcript, even though a side-by-side split does not.
	taskListMinWidth = 72
	// sidebarWidth is the sidebar's own column count, excluding its rule.
	sidebarWidth = 26
	// sidebarGutter is the rule plus the space either side of it.
	sidebarGutter = 3
)

type layout struct {
	width, height int

	// showHeader is false only when the terminal is too short to spare a row.
	showHeader bool
	// showSidebar puts tasks and subagents beside the transcript.
	showSidebar bool
	// showTaskList puts the full task list below the transcript instead.
	showTaskList bool

	// transcriptWidth is the usable width of the transcript column.
	transcriptWidth int
	// sidebarWidth is zero when no sidebar is drawn.
	sidebarWidth int
	// transcriptHeight is what remains after the header and every footer row.
	transcriptHeight int
}

// computeLayout decides the frame's geometry. footerHeight is the measured
// height of the footer stack, passed in so the footer is rendered exactly once.
// hasSidebarContent gates the sidebar on there being something to put in it, so
// an idle session does not give up a column to say "No active tasks".
func computeLayout(width, height, footerHeight int, hasOverlay, hasSidebarContent bool) layout {
	l := layout{width: width, height: height}

	// A header costs a row, and on a very short terminal the transcript needs
	// it more. An overlay is transient, so the header stays put under it.
	l.showHeader = height >= 12

	switch {
	case width >= sidebarMinWidth && !hasOverlay && hasSidebarContent:
		l.showSidebar = true
		l.sidebarWidth = sidebarWidth
		l.transcriptWidth = width - sidebarWidth - sidebarGutter
	case width >= taskListMinWidth:
		l.showTaskList = true
		l.transcriptWidth = width
	default:
		l.transcriptWidth = width
	}
	if l.transcriptWidth < 20 {
		l.transcriptWidth = maxInt(1, width)
		l.showSidebar = false
		l.sidebarWidth = 0
	}

	used := footerHeight
	if l.showHeader {
		used++ // header row
	}
	used++ // the newline View emits between the body and the footer

	l.transcriptHeight = height - used
	if l.transcriptHeight < 1 {
		l.transcriptHeight = 1
	}
	return l
}

// contentWidth is the width markdown and wrapping should target.
func (l layout) contentWidth() int {
	if l.transcriptWidth < 10 {
		return 80
	}
	return l.transcriptWidth
}
