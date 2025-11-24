// Package testutil contains test helpers.
package testutil

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

const defaultFilePermissions = 0o600

// WriteYAMLFile writes the given YAML content to a temporary file and returns its path.
// The file will be automatically cleaned up when the test completes.
// Returns the absolute path to the created file.
func WriteYAMLFile(t *testing.T, content string) string {
	t.Helper()

	tmpDir := t.TempDir()
	// Use test name to create unique filename for parallel test execution
	filename := fmt.Sprintf("config-%s.yaml", t.Name())
	configPath := filepath.Join(tmpDir, filename)

	err := os.WriteFile(configPath, []byte(content), defaultFilePermissions)
	if err != nil {
		t.Fatalf("Failed to write YAML file: %v", err)
	}

	return configPath
}
