package testutil

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWriteYAMLFile(t *testing.T) {
	content := "key: value\nfoo: bar"

	filePath := WriteYAMLFile(t, content)

	//nolint:gosec // Test code reading from known test file
	data, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("Failed to read created file: %v", err)
	}

	assert.Equal(t, content, string(data))
}
