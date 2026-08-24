package tools

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"

	"github.com/rushikeshg25/cool-code/internal/project"
	"github.com/rushikeshg25/cool-code/internal/skills"
	"github.com/rushikeshg25/cool-code/internal/types"
)

var projectSummaryTool = Tool{
	Name:        "project_summary",
	Description: "Summarizes the project (entrypoints, frameworks, scripts, languages).",
	ReadOnly:    true,
	Schema:      obj(map[string]any{}),
	Execute: func(ctx Context, _ json.RawMessage) types.ToolResult {
		packagePath := filepath.Join(ctx.RootDir, "package.json")
		if _, err := os.Lstat(packagePath); err == nil {
			if reason := ValidateReadPath(packagePath, ctx); reason != "" {
				return fail("Project summary blocked", reason)
			}
		}
		scan := project.ScanProject(ctx.RootDir)
		data, _ := json.MarshalIndent(scan, "", "  ")
		return types.ToolResult{Display: "Project summary generated", LLMResult: string(data)}
	},
}

var generateReadmeSectionTool = Tool{
	Name:        "generate_readme_section",
	Description: "Appends a section to README.md.",
	Mutating:    true,
	Schema: obj(map[string]any{
		"title":   strProp("Section title."),
		"bullets": arrProp("Optional bullet list."),
		"content": strProp("Optional raw content."),
	}, "title"),
	Execute: func(ctx Context, args json.RawMessage) types.ToolResult {
		var a struct {
			Title   string   `json:"title"`
			Bullets []string `json:"bullets"`
			Content string   `json:"content"`
		}
		if err := json.Unmarshal(args, &a); err != nil {
			return fail("Invalid arguments", err.Error())
		}
		if strings.TrimSpace(a.Title) == "" {
			return fail("Invalid arguments", "title is required.")
		}
		readmePath := filepath.Join(ctx.RootDir, "README.md")
		if reason := EnsureAbsoluteWithinRoots(readmePath, ctx.Roots()); reason != "" {
			return fail("README update blocked", reason)
		}
		heading := "## " + a.Title + "\n"
		var body string
		switch {
		case strings.TrimSpace(a.Content) != "":
			body = strings.TrimSpace(a.Content) + "\n"
		case len(a.Bullets) > 0:
			body = "- " + strings.Join(a.Bullets, "\n- ") + "\n"
		default:
			body = "- TODO: add details\n"
		}
		section := "\n" + heading + "\n" + body

		if raw, err := os.ReadFile(readmePath); err == nil {
			if strings.Contains(string(raw), strings.TrimSpace(heading)) {
				return fail("Section already exists", "README already contains section \""+a.Title+"\".")
			}
			f, err := os.OpenFile(readmePath, os.O_APPEND|os.O_WRONLY, 0o644)
			if err != nil {
				return fail("README update failed", err.Error())
			}
			_, _ = f.WriteString(section)
			_ = f.Close()
		} else {
			_ = os.WriteFile(readmePath, []byte("# "+filepath.Base(ctx.RootDir)+"\n"+section), 0o644)
		}
		return types.ToolResult{Display: "README updated", LLMResult: "Added section \"" + a.Title + "\" to README.md"}
	},
}

var useSkillTool = Tool{
	Name: "use_skill",
	Description: "Loads the full instructions for one of the available skills into context. " +
		"Call this when a skill is relevant before acting on it.",
	Schema: obj(map[string]any{
		"name": strProp("The exact name of the skill to load."),
	}, "name"),
	Execute: func(ctx Context, args json.RawMessage) types.ToolResult {
		var a struct {
			Name string `json:"name"`
		}
		if err := json.Unmarshal(args, &a); err != nil {
			return fail("Invalid arguments", err.Error())
		}
		if strings.TrimSpace(a.Name) == "" {
			return fail("Invalid arguments", "skill name is required.")
		}
		body, ok := skills.Body(ctx.RootDir, a.Name)
		if !ok {
			available := strings.Join(skills.Names(ctx.RootDir), ", ")
			if available == "" {
				available = "none"
			}
			return fail("Skill not found", "No skill named \""+a.Name+"\". Available skills: "+available+".")
		}
		return types.ToolResult{
			Display:   "Loaded skill: " + a.Name,
			LLMResult: "Instructions for skill \"" + a.Name + "\":\n" + body,
		}
	},
}
