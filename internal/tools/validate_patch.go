package tools

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/rushikeshg25/cool-code/internal/types"
)

var validatePatchTool = Tool{
	Name:        "validate_patch",
	Description: "Validate an apply_patch document without changing files or creating directories. Checks paths, existing content and unique hunk matches, then lists proposed operations. Available in read-only modes. Validation is not approval and does not guarantee a later apply will succeed if files change.",
	ReadOnly:    true,
	Schema:      obj(map[string]any{"patch": strProp("Complete apply_patch document.")}, "patch"),
	Execute: func(ctx Context, args json.RawMessage) types.ToolResult {
		var a struct {
			Patch string `json:"patch"`
		}
		if err := json.Unmarshal(args, &a); err != nil {
			return fail("Invalid patch", err.Error())
		}
		changes, err := preparePatchChecked(ctx, a.Patch, true)
		if err != nil {
			return fail("Patch validation failed", err.Error())
		}
		var lines []string
		for _, c := range changes {
			lines = append(lines, c.kind+" "+toRelative(c.path, ctx.RootDir))
		}
		return types.ToolResult{Display: fmt.Sprintf("Validated %d file(s); no changes made", len(changes)), LLMResult: "Patch valid against current files. No changes made.\n" + strings.Join(lines, "\n")}
	},
}
