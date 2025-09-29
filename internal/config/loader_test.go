package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/vlazic/mcp-server-manager/internal/models"
	"github.com/vlazic/mcp-server-manager/internal/services/testutil"
)

// TestLoadConfig_OrderPreservation verifies that server order defined in YAML is preserved
// when loading configuration. This is critical for v2.0 architecture which uses yaml.v3 Node
// parsing to maintain declaration order. The test creates a YAML file with specific order
// (server-b, server-a, server-c) and verifies the loaded MCPServers slice maintains that order.
//
// IMPORTANT: This only tests LoadConfig. SaveConfig has a known limitation where it uses
// map[string]interface{} which loses order. See TestOrderPreservation_MultipleServers for
// documentation of that limitation.
func TestLoadConfig_OrderPreservation(t *testing.T) {
	// Create a temporary config file with specific server order
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, testutil.TestConfigYAML)

	yamlContent := `server_port: 6543

mcpServers:
  server-b:
    command: "echo"
    args: ["b"]
  server-a:
    command: "echo"
    args: ["a"]
  server-c:
    command: "echo"
    args: ["c"]

clients:
  test_client:
    config_path: "~/.test.json"
    enabled:
      - server-a
`

	if err := os.WriteFile(configPath, []byte(yamlContent), 0644); err != nil {
		t.Fatalf(testutil.ErrWriteConfigFailedFmt, err)
	}

	// Load the config
	cfg, actualPath, err := LoadConfig(configPath)
	if err != nil {
		t.Fatalf(testutil.ErrLoadConfigFailedFmt, err)
	}

	if actualPath != configPath {
		t.Errorf("Expected path %s, got %s", configPath, actualPath)
	}

	// Verify order is preserved: server-b, server-a, server-c
	if len(cfg.MCPServers) != 3 {
		t.Fatalf("Expected 3 servers, got %d", len(cfg.MCPServers))
	}

	expectedOrder := []string{"server-b", "server-a", "server-c"}
	for i, expected := range expectedOrder {
		if cfg.MCPServers[i].Name != expected {
			t.Errorf("Server[%d]: expected %s, got %s", i, expected, cfg.MCPServers[i].Name)
		}
	}
}

func TestLoadConfig_DefaultPort(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, testutil.TestConfigYAML)

	// Config without server_port specified
	yamlContent := `mcpServers:
  test:
    command: "echo"

clients:
  test_client:
    config_path: "~/.test.json"
    enabled: []
`

	if err := os.WriteFile(configPath, []byte(yamlContent), 0644); err != nil {
		t.Fatalf(testutil.ErrWriteConfigFailedFmt, err)
	}

	cfg, _, err := LoadConfig(configPath)
	if err != nil {
		t.Fatalf(testutil.ErrLoadConfigFailedFmt, err)
	}

	if cfg.ServerPort != 6543 {
		t.Errorf("Expected default port 6543, got %d", cfg.ServerPort)
	}
}

func TestLoadConfig_InvalidYAML(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, testutil.TestConfigYAML)

	// Invalid YAML syntax
	yamlContent := `mcpServers:
  test
    command: "echo"
`

	if err := os.WriteFile(configPath, []byte(yamlContent), 0644); err != nil {
		t.Fatalf(testutil.ErrWriteConfigFailedFmt, err)
	}

	_, _, err := LoadConfig(configPath)
	if err == nil {
		t.Error("Expected error for invalid YAML, got nil")
	}
}

func TestLoadConfig_FileNotFound(t *testing.T) {
	// Note: LoadConfig actually creates a default config if explicit path doesn't exist
	// This tests that behavior
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "nonexistent", testutil.TestConfigYAML)

	cfg, actualPath, err := LoadConfig(configPath)
	if err != nil {
		t.Fatalf(testutil.ErrLoadConfigFailedFmt, err)
	}

	// Should have created default config
	if actualPath != configPath {
		t.Errorf("Expected path %s, got %s", configPath, actualPath)
	}

	// Verify default config was created
	if cfg.ServerPort != 6543 {
		t.Errorf("Expected default port 6543, got %d", cfg.ServerPort)
	}

	// Verify file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Error("Config file was not auto-created")
	}
}

func TestSaveConfig(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "output.yaml")

	cfg := &models.Config{
		ServerPort: 8080,
		MCPServers: []models.MCPServer{
			{
				Name: testutil.TestServerName,
				Config: map[string]interface{}{
					"command": "npx",
					"args":    []interface{}{"test"},
				},
			},
		},
		Clients: map[string]*models.Client{
			"test_client": {
				ConfigPath: "~/.test.json",
				Enabled:    []string{testutil.TestServerName},
			},
		},
	}

	// Save config
	if err := SaveConfig(cfg, configPath); err != nil {
		t.Fatalf("SaveConfig failed: %v", err)
	}

	// Verify file was created
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Fatal("Config file was not created")
	}

	// Reload and verify
	loadedCfg, _, err := LoadConfig(configPath)
	if err != nil {
		t.Fatalf("Failed to reload config: %v", err)
	}

	if loadedCfg.ServerPort != 8080 {
		t.Errorf("ServerPort: expected 8080, got %d", loadedCfg.ServerPort)
	}

	if len(loadedCfg.MCPServers) != 1 {
		t.Fatalf("Expected 1 server, got %d", len(loadedCfg.MCPServers))
	}

	if loadedCfg.MCPServers[0].Name != testutil.TestServerName {
		t.Errorf("Server name: expected 'test-server', got '%s'", loadedCfg.MCPServers[0].Name)
	}
}

func TestExpandPath(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		expected string // empty means dynamic based on home dir
	}{
		{
			name:     "Home directory expansion",
			path:     "~/.config/test",
			expected: "", // will be computed
		},
		{
			name:     "Absolute path no expansion",
			path:     "/etc/config",
			expected: "/etc/config",
		},
		{
			name:     "Relative path no expansion",
			path:     testutil.TestConfigPath,
			expected: testutil.TestConfigPath,
		},
		{
			name:     "Empty path",
			path:     "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ExpandPath(tt.path)

			if tt.expected == "" && tt.path != "" && tt.path[0] == '~' {
				// Verify it expanded by checking it doesn't start with ~
				if result[0] == '~' {
					t.Errorf("Path was not expanded: %s", result)
				}
				// Verify it's an absolute path
				if !filepath.IsAbs(result) {
					t.Errorf("Expanded path is not absolute: %s", result)
				}
			} else if result != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestResolveConfigPath_ExplicitPath(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "explicit.yaml")

	// resolveConfigPath will create default config if file doesn't exist
	resolved, err := resolveConfigPath(configPath)
	if err != nil {
		t.Fatalf("resolveConfigPath failed: %v", err)
	}

	if resolved != configPath {
		t.Errorf("Expected %s, got %s", configPath, resolved)
	}

	// Verify file was created
	if _, err := os.Stat(resolved); os.IsNotExist(err) {
		t.Error("Config file was not auto-created")
	}
}

func TestResolveConfigPath_FindsExisting(t *testing.T) {
	tempDir := t.TempDir()

	// Change to temp directory for this test
	oldWd, _ := os.Getwd()
	defer os.Chdir(oldWd)
	os.Chdir(tempDir)

	// Create config in current directory
	configPath := testutil.TestConfigPath
	if err := os.WriteFile(configPath, []byte("server_port: 6543\nmcpServers: {}\nclients: {}"), 0644); err != nil {
		t.Fatalf("Failed to create test config: %v", err)
	}

	// Note: resolveConfigPath prioritizes ~/.config/mcp-server-manager/config.yaml over ./config.yaml
	// If the user config exists, it will be found first
	// This test verifies that explicit path works
	absConfigPath, _ := filepath.Abs(configPath)
	resolved, err := resolveConfigPath(absConfigPath)
	if err != nil {
		t.Fatalf("resolveConfigPath failed: %v", err)
	}

	if resolved != absConfigPath {
		t.Errorf("Expected to find %s, got %s", absConfigPath, resolved)
	}
}

func TestCreateDefaultConfig(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "subdir", testutil.TestConfigYAML)

	err := createDefaultConfig(configPath)
	if err != nil {
		t.Fatalf("createDefaultConfig failed: %v", err)
	}

	// Verify file was created
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Fatal("Default config file was not created")
	}

	// Verify it's valid YAML
	cfg, _, err := LoadConfig(configPath)
	if err != nil {
		t.Fatalf("Failed to load created config: %v", err)
	}

	// Verify default port
	if cfg.ServerPort != 6543 {
		t.Errorf("Default config port: expected 6543, got %d", cfg.ServerPort)
	}

	// Verify it has example servers
	if len(cfg.MCPServers) == 0 {
		t.Error("Default config should have example servers")
	}

	// Verify it has example clients
	if len(cfg.Clients) == 0 {
		t.Error("Default config should have example clients")
	}
}

func TestLoadConfig_EmptyMCPServers(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, testutil.TestConfigYAML)

	yamlContent := `server_port: 6543

mcpServers: {}

clients:
  test_client:
    config_path: "~/.test.json"
    enabled: []
`

	if err := os.WriteFile(configPath, []byte(yamlContent), 0644); err != nil {
		t.Fatalf(testutil.ErrWriteConfigFailedFmt, err)
	}

	cfg, _, err := LoadConfig(configPath)
	if err != nil {
		t.Fatalf(testutil.ErrLoadConfigFailedFmt, err)
	}

	if len(cfg.MCPServers) != 0 {
		t.Errorf("Expected 0 servers, got %d", len(cfg.MCPServers))
	}
}

func TestLoadConfig_MalformedYAML(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, testutil.TestConfigYAML)

	// Missing colon after key
	yamlContent := `server_port 6543
mcpServers:
  test
    command: "echo"
`

	if err := os.WriteFile(configPath, []byte(yamlContent), 0644); err != nil {
		t.Fatalf(testutil.ErrWriteConfigFailedFmt, err)
	}

	_, _, err := LoadConfig(configPath)
	if err == nil {
		t.Error("Expected error for malformed YAML")
	}
}

func TestLoadConfig_InvalidServerConfig(t *testing.T) {
	// Test that LoadConfig itself doesn't validate (validation happens separately)
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, testutil.TestConfigYAML)

	yamlContent := `server_port: 6543

mcpServers:
  invalid:
    badfield: "value"

clients:
  test:
    config_path: "~/.test.json"
    enabled: []
`

	if err := os.WriteFile(configPath, []byte(yamlContent), 0644); err != nil {
		t.Fatalf(testutil.ErrWriteConfigFailedFmt, err)
	}

	// LoadConfig should succeed (it doesn't validate)
	cfg, _, err := LoadConfig(configPath)
	if err != nil {
		t.Fatalf("LoadConfig should not validate, but got error: %v", err)
	}

	if len(cfg.MCPServers) != 1 {
		t.Errorf("Expected 1 server, got %d", len(cfg.MCPServers))
	}
}

func TestSaveConfig_InvalidPath(t *testing.T) {
	cfg := &models.Config{
		ServerPort: 6543,
		MCPServers: []models.MCPServer{
			{Name: "test", Config: map[string]interface{}{"command": "echo"}},
		},
		Clients: map[string]*models.Client{
			"test": {ConfigPath: "~/.test.json", Enabled: []string{}},
		},
	}

	// Try to save to an invalid path (non-existent parent directory with no write permission simulation)
	// Note: This test is platform-dependent and may not work on all systems
	invalidPath := "/root/nonexistent/config.yaml"
	err := SaveConfig(cfg, invalidPath)
	if err == nil {
		t.Error("Expected error when saving to invalid path")
	}
}

func TestExpandPath_EdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		checkFn  func(string) bool
		errorMsg string
	}{
		{
			name:  "Tilde only",
			input: "~",
			checkFn: func(result string) bool {
				return result != "~" && len(result) > 0
			},
			errorMsg: "Tilde-only path should be expanded",
		},
		{
			name:  "Tilde with trailing slash",
			input: "~/",
			checkFn: func(result string) bool {
				return result != "~/" && len(result) > 1
			},
			errorMsg: "Tilde with slash should be expanded",
		},
		{
			name:  "Path without tilde unchanged",
			input: "/absolute/path",
			checkFn: func(result string) bool {
				return result == "/absolute/path"
			},
			errorMsg: "Absolute path should remain unchanged",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ExpandPath(tt.input)
			if !tt.checkFn(result) {
				t.Errorf("%s: got '%s'", tt.errorMsg, result)
			}
		})
	}
}