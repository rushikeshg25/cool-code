package tools

import (
	"testing"
	"time"
)

func TestCommandTimeoutCannotOverflow(t *testing.T) {
	for _, seconds := range []int{1801, int(^uint(0) >> 1)} {
		if got := commandTimeout(seconds); got != maxCommandTimeout {
			t.Fatalf("%d seconds became %v", seconds, got)
		}
	}
	if commandTimeout(1) != time.Second || commandTimeout(0) != defaultCommandTimeout {
		t.Fatal("normal timeouts changed")
	}
}
