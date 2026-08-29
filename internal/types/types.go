// Package types holds the small domain types shared across cool-code.
package types

// AgentMode is the operating mode of the agent.
type AgentMode string

const (
	ModePlan  AgentMode = "plan"
	ModeAgent AgentMode = "agent"
	ModeAsk   AgentMode = "ask"
)

// Valid reports whether m is a known mode.
func (m AgentMode) Valid() bool {
	return m == ModePlan || m == ModeAgent || m == ModeAsk
}

// ToolResult is what every tool returns: LLMResult is fed back to the model,
// Display is the short human-facing status line shown in the TUI.
type ToolResult struct {
	LLMResult string
	Display   string
	// Failed marks a call that did not do what it was asked. Without it the
	// UI drew a failure exactly like a success.
	Failed bool
}

// TaskStatus is the state of a single task item.
type TaskStatus string

const (
	TaskTodo       TaskStatus = "todo"
	TaskInProgress TaskStatus = "in-progress"
	TaskDone       TaskStatus = "done"
	TaskFailed     TaskStatus = "failed"
)

// TaskItem is one entry in the model-managed task list.
type TaskItem struct {
	ID     string     `json:"id"`
	Title  string     `json:"title"`
	Status TaskStatus `json:"status"`
	Detail string     `json:"detail,omitempty"`
}

// TaskList is the model's plan/progress tracker, surfaced in the TUI.
type TaskList struct {
	Goal  string     `json:"goal"`
	Items []TaskItem `json:"items"`
}
