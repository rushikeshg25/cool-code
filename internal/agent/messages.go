package agent

// waitingLabel is what the activity line says while a request is in flight.
//
// This used to pick at random from a list of flavour text: "Consulting the
// oracle…", "Scanning the matrix…", "Distilling digital wisdom…". None of it
// described anything, it changed every turn for no reason, and in a tool that
// works on code, "oracle" reads as the database. A status line should say what
// is happening; the caller adds elapsed time to it.
func waitingLabel() string { return "Thinking…" }

var toolLabels = map[string]string{
	"read_file":               "Reading file",
	"open_file_at":            "Reading file",
	"edit_file":               "Editing file",
	"new_file":                "Creating file",
	"rename_file":             "Renaming file",
	"list_recent_files":       "Listing recent files",
	"replace_in_files":        "Replacing across files",
	"new_module":              "Scaffolding module",
	"glob":                    "Finding files",
	"grep":                    "Searching contents",
	"find_symbol":             "Searching symbols",
	"shell_command":           "Running command",
	"run_tests":               "Running tests",
	"lint_fix":                "Running lint/format",
	"format_file":             "Formatting file",
	"add_script":              "Updating package.json",
	"git_status":              "Checking git status",
	"git_diff":                "Reading git diff",
	"git_commit":              "Creating commit",
	"project_summary":         "Summarizing project",
	"generate_readme_section": "Updating README",
	"use_skill":               "Loading skill",
	"web_fetch":               "Fetching page",
	"web_search":              "Searching the web",
}

func toolStatus(name string) string {
	if label, ok := toolLabels[name]; ok {
		return label + "…"
	}
	return "Working…"
}
