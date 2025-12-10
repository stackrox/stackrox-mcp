package testutil

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWriteYAMLFile(t *testing.T) {
	content := "test: value\n"

	path := WriteYAMLFile(t, content)

	filename := filepath.Base(path)
	assert.Contains(t, filename, t.Name(), "Filename should contain test name")

	info, err := os.Stat(path)
	require.NoError(t, err, "File should exist")
	assert.Equal(t, int64(len(content)), info.Size(), "File should not be empty")
}
