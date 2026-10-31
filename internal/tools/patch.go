package tools

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/rushikeshg25/cool-code/internal/types"
)

var applyPatchTool = Tool{
	Name: "apply_patch", Mutating: true,
	Description: "Apply a multi-file text patch. Format: *** Begin Patch, then *** Add File: path (lines prefixed +), *** Delete File: path, or *** Update File: path (one or more @@ hunks with space/context, -/removed, +/added lines), then *** End Patch. Paths are relative to the project or absolute within granted roots. Context must match exactly and uniquely; no fuzzy matching. All files are validated before writing. Use edit_file for individual replacements. Move/rename headers and binary files are unsupported.",
	Schema:      obj(map[string]any{"patch": strProp("Complete patch in the documented format.")}, "patch"),
	Execute: func(ctx Context, args json.RawMessage) types.ToolResult {
		var a struct {
			Patch string `json:"patch"`
		}
		if err := json.Unmarshal(args, &a); err != nil {
			return fail("Invalid patch", err.Error())
		}
		changes, err := preparePatch(ctx, a.Patch)
		if err != nil {
			return fail("Patch rejected", err.Error())
		}
		if err = writePatch(ctx, changes); err != nil {
			return fail("Patch failed", err.Error())
		}
		var paths []string
		for _, c := range changes {
			paths = append(paths, c.kind+" "+c.path)
		}
		return types.ToolResult{Display: fmt.Sprintf("Patched %d file(s)", len(changes)), LLMResult: strings.Join(paths, "\n")}
	},
}

type patchChange struct {
	kind, path    string
	before, after []byte
	mode          os.FileMode
}

func preparePatch(ctx Context, patch string) ([]patchChange, error) {
	if len(patch) > 4*1024*1024 {
		return nil, errors.New("patch exceeds 4 MiB")
	}
	lines := strings.Split(strings.TrimSuffix(strings.ReplaceAll(patch, "\r\n", "\n"), "\n"), "\n")
	if len(lines) < 3 || lines[0] != "*** Begin Patch" || lines[len(lines)-1] != "*** End Patch" {
		return nil, errors.New("expected Begin Patch and End Patch markers")
	}
	var changes []patchChange
	seen := map[string]bool{}
	for i := 1; i < len(lines)-1; {
		header := lines[i]
		i++
		kind, path, ok := strings.Cut(header, ": ")
		if !ok || (kind != "*** Add File" && kind != "*** Update File" && kind != "*** Delete File") || path == "" {
			return nil, fmt.Errorf("invalid file header %q", header)
		}
		if !filepath.IsAbs(path) {
			path = filepath.Join(ctx.RootDir, path)
		}
		path, reason := ResolveWritePath(path, ctx)
		if reason != "" {
			return nil, fmt.Errorf("patch path blocked: %s", reason)
		}
		if seen[path] {
			return nil, fmt.Errorf("duplicate patch target: %s", path)
		}
		seen[path] = true
		c := patchChange{kind: strings.TrimPrefix(kind, "*** "), path: path, mode: 0644}
		stat, err := os.Stat(path)
		if kind == "*** Add File" {
			if !os.IsNotExist(err) {
				return nil, fmt.Errorf("add target must not exist: %s", path)
			}
		} else {
			if err != nil {
				return nil, err
			}
			if !stat.Mode().IsRegular() || stat.Size() > 4*1024*1024 {
				return nil, fmt.Errorf("expected regular text file under 4 MiB: %s", path)
			}
			c.mode = stat.Mode().Perm()
			c.before, err = os.ReadFile(path)
			if err != nil {
				return nil, err
			}
			if bytes.ContainsRune(c.before, 0) {
				return nil, errors.New("binary patches are unsupported")
			}
		}
		start := i
		for i < len(lines)-1 && !strings.HasPrefix(lines[i], "*** ") {
			i++
		}
		body := lines[start:i]
		switch kind {
		case "*** Add File":
			var b strings.Builder
			for _, line := range body {
				if !strings.HasPrefix(line, "+") {
					return nil, errors.New("added file lines must start with +")
				}
				b.WriteString(line[1:])
				b.WriteByte('\n')
			}
			c.after = []byte(b.String())
		case "*** Delete File":
			if len(body) != 0 {
				return nil, errors.New("delete header must not have a body")
			}
		case "*** Update File":
			c.after, err = applyPatchHunks(c.before, body)
			if err != nil {
				return nil, fmt.Errorf("%s: %w", path, err)
			}
		}
		if bytes.ContainsRune(c.after, 0) {
			return nil, errors.New("binary patches are unsupported")
		}
		changes = append(changes, c)
		if len(changes) > 100 {
			return nil, errors.New("patch exceeds 100 files")
		}
	}
	if len(changes) == 0 {
		return nil, errors.New("patch contains no files")
	}
	// Refuse overlapping parent/child targets before any filesystem changes.
	for _, a := range changes {
		for _, b := range changes {
			if a.path != b.path && strings.HasPrefix(b.path, a.path+string(filepath.Separator)) {
				return nil, errors.New("patch targets overlap")
			}
		}
	}
	return changes, nil
}
func applyPatchHunks(before []byte, body []string) ([]byte, error) {
	if len(body) == 0 {
		return nil, errors.New("update requires an @@ hunk")
	}
	text := string(before)
	newline := strings.HasSuffix(text, "\n")
	lines := strings.Split(strings.TrimSuffix(text, "\n"), "\n")
	if len(before) == 0 {
		lines = nil
	}
	cursor := 0
	for i := 0; i < len(body); {
		if body[i] != "@@" && !strings.HasPrefix(body[i], "@@ ") {
			return nil, errors.New("expected @@ hunk")
		}
		i++
		var old, replacement []string
		for i < len(body) && body[i] != "@@" && !strings.HasPrefix(body[i], "@@ ") {
			line := body[i]
			i++
			if len(line) == 0 {
				return nil, errors.New("hunk lines need a space, +, or - prefix")
			}
			switch line[0] {
			case ' ':
				old = append(old, line[1:])
				replacement = append(replacement, line[1:])
			case '-':
				old = append(old, line[1:])
			case '+':
				replacement = append(replacement, line[1:])
			default:
				return nil, errors.New("unsupported hunk line")
			}
		}
		if len(old) == 0 {
			return nil, errors.New("update hunk needs context or removed lines; use Add File for a new file")
		}
		match := -1
		for at := cursor; at+len(old) <= len(lines); at++ {
			equal := true
			for j := range old {
				if lines[at+j] != old[j] {
					equal = false
					break
				}
			}
			if equal {
				if match >= 0 {
					return nil, errors.New("ambiguous hunk; include more context")
				}
				match = at
			}
		}
		if match < 0 {
			return nil, errors.New("hunk does not match current file")
		}
		updated := append([]string(nil), lines[:match]...)
		updated = append(updated, replacement...)
		updated = append(updated, lines[match+len(old):]...)
		lines = updated
		cursor = match + len(replacement)
	}
	result := strings.Join(lines, "\n")
	if newline && len(lines) > 0 {
		result += "\n"
	}
	return []byte(result), nil
}

func writePatch(ctx Context, changes []patchChange) error {
	// Revalidate both paths and original content after preparation.
	for _, c := range changes {
		if err := ctx.Context().Err(); err != nil {
			return err
		}
		path, reason := ResolveWritePath(c.path, ctx)
		if reason != "" || path != c.path {
			return errors.New("patch path changed during preparation")
		}
		data, err := os.ReadFile(c.path)
		if c.kind == "Add File" {
			if !os.IsNotExist(err) {
				return errors.New("add target appeared during preparation")
			}
		} else if err != nil || !bytes.Equal(data, c.before) {
			return errors.New("file changed during patch preparation")
		}
	}
	applied := 0
	for _, c := range changes {
		var err error
		if err = ctx.Context().Err(); err == nil {
			if c.kind == "Delete File" {
				err = os.Remove(c.path)
			} else {
				err = replacePatchFile(c.path, c.after, c.mode)
			}
		}
		if err != nil {
			rollback := []error{fmt.Errorf("patch write failed: %w", err)}
			for j := applied - 1; j >= 0; j-- {
				prior := changes[j]
				var restore error
				if prior.kind == "Add File" {
					restore = os.Remove(prior.path)
				} else {
					restore = replacePatchFile(prior.path, prior.before, prior.mode)
				}
				if restore != nil {
					rollback = append(rollback, fmt.Errorf("rollback failed for %s: %w", prior.path, restore))
				}
			}
			return errors.Join(rollback...)
		}
		applied++
	}
	return nil
}
func replacePatchFile(path string, data []byte, mode os.FileMode) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	f, err := os.CreateTemp(filepath.Dir(path), ".coolcode-patch-*")
	if err != nil {
		return err
	}
	defer os.Remove(f.Name())
	if err = f.Chmod(mode); err == nil {
		_, err = f.Write(data)
	}
	closeErr := f.Close()
	if err != nil {
		return err
	}
	if closeErr != nil {
		return closeErr
	}
	return os.Rename(f.Name(), path)
}
