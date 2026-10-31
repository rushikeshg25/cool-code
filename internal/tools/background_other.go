//go:build !linux && !darwin

package tools

import (
	"errors"
	"os/exec"
)

func configureBackgroundProcess(_ *exec.Cmd) error {
	return errors.New("managed background processes currently require Linux or macOS")
}
