package testutil

import (
	"context"
	"os/exec"
	"strings"
)

// RunCommand executes a shell command and returns the output and error.
func RunCommand(command string) (string, error) {
	parts := strings.Fields(command)
	if len(parts) == 0 {
		return "", nil
	}

	// #nosec G204 - This is a test utility function with controlled input
	cmd := exec.CommandContext(context.Background(), parts[0], parts[1:]...)
	output, err := cmd.CombinedOutput()

	return string(output), err
}
