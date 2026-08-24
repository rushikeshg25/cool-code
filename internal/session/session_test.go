package session

import (
	"os"
	"runtime"
	"testing"
)

func TestSessionFilesArePrivateAndIDsCannotTraverse(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	id := NewID()
	if err := Save(Data{ID: id, Cwd: "/workspace"}); err != nil {
		t.Fatal(err)
	}
	if runtime.GOOS != "windows" {
		dirInfo, err := os.Stat(dir())
		if err != nil {
			t.Fatal(err)
		}
		if got := dirInfo.Mode().Perm(); got != 0o700 {
			t.Fatalf("session directory mode = %o", got)
		}
		path, _ := pathFor(id)
		fileInfo, err := os.Stat(path)
		if err != nil {
			t.Fatal(err)
		}
		if got := fileInfo.Mode().Perm(); got != 0o600 {
			t.Fatalf("session file mode = %o", got)
		}
	}
	if Load("../../credentials") != nil {
		t.Fatal("path traversal session ID was accepted")
	}
	if err := Save(Data{ID: "../../credentials"}); err == nil {
		t.Fatal("invalid session ID was saved")
	}
}
