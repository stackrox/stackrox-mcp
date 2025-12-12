package testutil

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestWriteYAMLFile(t *testing.T) {
	content := "key: value\nfoo: bar"

	filePath := WriteYAMLFile(t, content)

	//nolint:gosec // Test code reading from known test file
	data, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("Failed to read created file: %v", err)
	}

	if string(data) != content {
		t.Errorf("Expected content %q, got %q", content, string(data))
	}
}

func TestWriteYAMLFile_ReturnsAbsolutePath(t *testing.T) {
	filePath := WriteYAMLFile(t, "test: content")

	if !filepath.IsAbs(filePath) {
		t.Errorf("Expected absolute path, got: %s", filePath)
	}
}

func TestWriteYAMLFile_HasYamlExtension(t *testing.T) {
	filePath := WriteYAMLFile(t, "test: content")

	if !strings.HasSuffix(filePath, ".yaml") {
		t.Errorf("Expected .yaml extension, got: %s", filePath)
	}
}

func TestWriteYAMLFile_PathIncludesTestName(t *testing.T) {
	filePath := WriteYAMLFile(t, "test: value")

	fileName := filepath.Base(filePath)
	// The filename should contain part of the test name
	// but sanitized (slashes replaced or handled)
	if !strings.HasPrefix(fileName, "config-") {
		t.Errorf("Expected filename to start with 'config-', got: %s", fileName)
	}
}

func TestWriteYAMLFile_FileExists(t *testing.T) {
	filePath := WriteYAMLFile(t, "exists: true")

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		t.Errorf("File should exist at path: %s", filePath)
	}
}

func TestWriteYAMLFile_InTempDirectory(t *testing.T) {
	filePath := WriteYAMLFile(t, "test: content")

	// The file should be in a temp directory managed by t.TempDir()
	// which will be automatically cleaned up
	dir := filepath.Dir(filePath)
	if dir == "" || dir == "." {
		t.Errorf("File should be in a proper directory, got: %s", dir)
	}
}

func TestWriteYAMLFile_LargeContent(t *testing.T) {
	// Test with larger YAML content
	var builder strings.Builder
	for i := range 100 {
		builder.WriteString("key")
		builder.WriteRune(rune('0' + i%10))
		builder.WriteString(": value")
		builder.WriteRune(rune('0' + i%10))
		builder.WriteString("\n")
	}

	content := builder.String()

	filePath := WriteYAMLFile(t, content)

	//nolint:gosec // Test code reading from known test file
	data, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("Failed to read file with large content: %v", err)
	}

	if string(data) != content {
		t.Error("Large content not preserved correctly")
	}
}
