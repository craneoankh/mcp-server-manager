package testutil

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/vlazic/mcp-server-manager/internal/models"
)

// CreateTempDir creates a temporary directory for testing
func CreateTempDir(t *testing.T, pattern string) string {
	t.Helper()
	tempDir, err := os.MkdirTemp("", pattern)
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	t.Cleanup(func() {
		os.RemoveAll(tempDir)
	})
	return tempDir
}

// WriteTestFile writes content to a file in the test directory
func WriteTestFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatalf("Failed to create directory: %v", err)
	}
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}
}

// CreateTestServer returns a standard test server config
func CreateTestServer() map[string]interface{} {
	return map[string]interface{}{
		"command": "npx",
		"args":    []interface{}{"@modelcontextprotocol/server-filesystem"},
	}
}

// CreateTestHTTPServer returns a test HTTP server config
func CreateTestHTTPServer() map[string]interface{} {
	return map[string]interface{}{
		"url": TestExampleURL,
	}
}

// CreateTestClient returns a standard test client
func CreateTestClient(configPath string, enabled []string) *models.Client {
	return &models.Client{
		ConfigPath: configPath,
		Enabled:    enabled,
	}
}

// AssertErrorContains checks if error contains substring
func AssertErrorContains(t *testing.T, err error, substr string) {
	t.Helper()
	if err == nil {
		t.Fatalf("Expected error containing '%s', got nil", substr)
	}
	if !containsSubstring(err.Error(), substr) {
		t.Errorf("Expected error containing '%s', got '%s'", substr, err.Error())
	}
}

// containsSubstring checks if s contains substr
func containsSubstring(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > 0 && len(substr) > 0 && stringContains(s, substr)))
}

func stringContains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}