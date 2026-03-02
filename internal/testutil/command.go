package testutil

import (
	"os/exec"
	"strings"
)

// RunCommand executes a shell command and returns the output and error.
func RunCommand(command string) (string, error) {
	parts := strings.Fields(command)
	if len(parts) == 0 {
		return "", nil
	}

	cmd := exec.Command(parts[0], parts[1:]...)
	output, err := cmd.CombinedOutput()
	return string(output), err
}
