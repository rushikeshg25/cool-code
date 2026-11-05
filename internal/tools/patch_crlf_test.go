package tools

import (
	"strings"
	"testing"
)

func TestPatchPreservesCRLF(t *testing.T) {
	for _, ending := range []string{"\r\n", ""} {
		before := []byte("first\r\nold" + ending)
		got, err := applyPatchHunks(before, []string{"@@", " first", "-old", "+new"})
		if err != nil {
			t.Fatal(err)
		}
		if string(got) != "first\r\nnew"+ending {
			t.Fatalf("line endings changed: %q", got)
		}
	}
	if _, err := applyPatchHunks([]byte("first\r\nold\n"), []string{"@@", "-old", "+new"}); err == nil || !strings.Contains(err.Error(), "mixed") {
		t.Fatal("mixed endings must be rejected explicitly")
	}
}
