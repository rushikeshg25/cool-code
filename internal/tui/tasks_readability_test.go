package tui

import (
	"github.com/charmbracelet/x/ansi"
	"github.com/rushikeshg25/cool-code/internal/types"
	"strings"
	"testing"
)

func TestLongTaskTitlesRemainReadable(t *testing.T) {
	m := newTestModel(t)
	title := "Initialize Git and TypeScript scaffolding with verification scripts"
	m.tasks = &types.TaskList{Goal: "Project setup", Items: []types.TaskItem{{ID: "1", Title: title, Detail: "Include the full validation commands", Status: types.TaskTodo}}}
	l := computeLayout(160, 40, 5, false, true)
	sidebar := strings.Join(strings.Fields(ansi.Strip(m.renderSidebar(l))), " ")
	if !strings.Contains(sidebar, title) {
		t.Fatalf("task title truncated: %s", sidebar)
	}
	m.processing = true
	m = typeLine(t, m, "/tasks")
	text := ansi.Strip(m.View())
	if !strings.Contains(text, "full validation commands") {
		t.Fatal("/tasks did not display full details during a turn")
	}
}
