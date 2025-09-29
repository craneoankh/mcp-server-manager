package services

import (
	"testing"

	"github.com/vlazic/mcp-server-manager/internal/models"
)

// TestValidateMCPServerConfig provides comprehensive coverage of server configuration validation
// including all transport types (command, url, httpUrl), environment variables, timeouts,
// and error conditions. This test ensures the validator correctly identifies:
// - Invalid server names (empty, whitespace)
// - Missing or multiple transport types
// - Invalid URLs (missing scheme/host, unsupported protocols)
// - Invalid environment variables (empty keys/values, keys with '=')
// - Negative timeouts
// - Commands not found in PATH
func TestValidateMCPServerConfig(t *testing.T) {
	validator := NewValidatorService()

	tests := []struct {
		name        string
		serverName  string
		config      map[string]interface{}
		wantErr     bool
		errContains string
	}{
		// Valid configurations
		{
			name:       "Valid STDIO server",
			serverName: "filesystem",
			config: map[string]interface{}{
				"command": "echo",
				"args":    []interface{}{"test"},
			},
			wantErr: false,
		},
		{
			name:       "Valid HTTP server with type",
			serverName: "context7",
			config: map[string]interface{}{
				"type": "http",
				"url":  "https://mcp.context7.com/mcp",
			},
			wantErr: false,
		},
		{
			name:       "Valid HTTP server with httpUrl",
			serverName: "context7-gemini",
			config: map[string]interface{}{
				"httpUrl": "https://mcp.context7.com/mcp",
			},
			wantErr: false,
		},
		{
			name:       "Valid server with environment variables",
			serverName: "with-env",
			config: map[string]interface{}{
				"command": "echo",
				"env": map[string]interface{}{
					"NODE_ENV": "production",
					"API_KEY":  "secret",
				},
			},
			wantErr: false,
		},
		{
			name:       "Valid server with timeout",
			serverName: "with-timeout",
			config: map[string]interface{}{
				"command": "echo",
				"timeout": 30000,
			},
			wantErr: false,
		},
		{
			name:       "Valid HTTP server with http scheme",
			serverName: "http-server",
			config: map[string]interface{}{
				"url": "http://localhost:8080/mcp",
			},
			wantErr: false,
		},

		// Invalid server names
		{
			name:        "Empty server name",
			serverName:  "",
			config:      map[string]interface{}{"command": "echo"},
			wantErr:     true,
			errContains: "name cannot be empty",
		},
		{
			name:        "Whitespace-only server name",
			serverName:  "   ",
			config:      map[string]interface{}{"command": "echo"},
			wantErr:     true,
			errContains: "name cannot be empty",
		},

		// Invalid transport types
		{
			name:        "No transport type",
			serverName:  "invalid",
			config:      map[string]interface{}{},
			wantErr:     true,
			errContains: "exactly one transport type",
		},
		{
			name:       "Multiple transports - command and url",
			serverName: "multi-transport",
			config: map[string]interface{}{
				"command": "echo",
				"url":     "https://example.com",
			},
			wantErr:     true,
			errContains: "exactly one transport type",
		},
		{
			name:       "Multiple transports - command and httpUrl",
			serverName: "multi-transport",
			config: map[string]interface{}{
				"command": "echo",
				"httpUrl": "https://example.com",
			},
			wantErr:     true,
			errContains: "exactly one transport type",
		},
		{
			name:       "Multiple transports - url and httpUrl",
			serverName: "multi-transport",
			config: map[string]interface{}{
				"url":     "https://example.com",
				"httpUrl": "https://example.com",
			},
			wantErr:     true,
			errContains: "exactly one transport type",
		},
		{
			name:       "All three transports",
			serverName: "all-transports",
			config: map[string]interface{}{
				"command": "echo",
				"url":     "https://example.com",
				"httpUrl": "https://example.com",
			},
			wantErr:     true,
			errContains: "exactly one transport type",
		},
		{
			name:        "Empty command string",
			serverName:  "empty-cmd",
			config:      map[string]interface{}{"command": ""},
			wantErr:     true,
			errContains: "exactly one transport type",
		},
		{
			name:        "Whitespace-only command",
			serverName:  "whitespace-cmd",
			config:      map[string]interface{}{"command": "   "},
			wantErr:     true,
			errContains: "exactly one transport type",
		},
		{
			name:        "Command not in PATH",
			serverName:  "invalid-cmd",
			config:      map[string]interface{}{"command": "nonexistent-command-xyz123"},
			wantErr:     true,
			errContains: "not found in PATH",
		},

		// Invalid URLs
		{
			name:        "URL without scheme",
			serverName:  "no-scheme",
			config:      map[string]interface{}{"url": "example.com/path"},
			wantErr:     true,
			errContains: "missing scheme",
		},
		{
			name:        "URL without host",
			serverName:  "no-host",
			config:      map[string]interface{}{"url": "https://"},
			wantErr:     true,
			errContains: "missing host",
		},
		{
			name:        "URL with invalid scheme",
			serverName:  "bad-scheme",
			config:      map[string]interface{}{"url": "ftp://example.com"},
			wantErr:     true,
			errContains: "must be http or https",
		},
		{
			name:        "httpUrl without scheme",
			serverName:  "httpurl-no-scheme",
			config:      map[string]interface{}{"httpUrl": "example.com"},
			wantErr:     true,
			errContains: "missing scheme",
		},
		{
			name:        "httpUrl with invalid scheme",
			serverName:  "httpurl-bad-scheme",
			config:      map[string]interface{}{"httpUrl": "ws://example.com"},
			wantErr:     true,
			errContains: "must be http or https",
		},
		{
			name:        "Empty URL",
			serverName:  "empty-url",
			config:      map[string]interface{}{"url": ""},
			wantErr:     true,
			errContains: "exactly one transport type",
		},

		// Invalid timeout
		{
			name:       "Negative timeout",
			serverName: "neg-timeout",
			config: map[string]interface{}{
				"command": "echo",
				"timeout": -1000,
			},
			wantErr:     true,
			errContains: "timeout cannot be negative",
		},

		// Invalid environment variables
		{
			name:       "Empty env key",
			serverName: "empty-env-key",
			config: map[string]interface{}{
				"command": "echo",
				"env": map[string]interface{}{
					"": "value",
				},
			},
			wantErr:     true,
			errContains: "key cannot be empty",
		},
		{
			name:       "Env key with equals sign",
			serverName: "env-key-equals",
			config: map[string]interface{}{
				"command": "echo",
				"env": map[string]interface{}{
					"KEY=VALUE": "test",
				},
			},
			wantErr:     true,
			errContains: "cannot contain '='",
		},
		{
			name:       "Empty env value",
			serverName: "empty-env-value",
			config: map[string]interface{}{
				"command": "echo",
				"env": map[string]interface{}{
					"KEY": "",
				},
			},
			wantErr:     true,
			errContains: "value for 'KEY' cannot be empty",
		},
		{
			name:       "Whitespace-only env value",
			serverName: "whitespace-env-value",
			config: map[string]interface{}{
				"command": "echo",
				"env": map[string]interface{}{
					"KEY": "   ",
				},
			},
			wantErr:     true,
			errContains: "value for 'KEY' cannot be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateMCPServerConfig(tt.serverName, tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateMCPServerConfig() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr && tt.errContains != "" && err != nil {
				if !containsSubstring(err.Error(), tt.errContains) {
					t.Errorf("Expected error containing '%s', got '%s'", tt.errContains, err.Error())
				}
			}
		})
	}
}

func TestValidateClient(t *testing.T) {
	validator := NewValidatorService()

	tests := []struct {
		name        string
		clientName  string
		client      *models.Client
		wantErr     bool
		errContains string
	}{
		{
			name:       "Valid client",
			clientName: "claude_code",
			client: &models.Client{
				ConfigPath: "~/.claude.json",
				Enabled:    []string{},
			},
			wantErr: false,
		},
		{
			name:        "Empty client name",
			clientName:  "",
			client:      &models.Client{ConfigPath: "~/.test.json"},
			wantErr:     true,
			errContains: "name cannot be empty",
		},
		{
			name:        "Whitespace-only client name",
			clientName:  "   ",
			client:      &models.Client{ConfigPath: "~/.test.json"},
			wantErr:     true,
			errContains: "name cannot be empty",
		},
		{
			name:        "Empty config path",
			clientName:  "test",
			client:      &models.Client{ConfigPath: ""},
			wantErr:     true,
			errContains: "config path cannot be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateClient(tt.clientName, tt.client)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateClient() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr && tt.errContains != "" && err != nil {
				if !containsSubstring(err.Error(), tt.errContains) {
					t.Errorf("Expected error containing '%s', got '%s'", tt.errContains, err.Error())
				}
			}
		})
	}
}

func TestValidateConfig(t *testing.T) {
	validator := NewValidatorService()

	t.Run("Valid config", func(t *testing.T) {
		cfg := &models.Config{
			ServerPort: 6543,
			MCPServers: []models.MCPServer{
				{Name: "test-server", Config: map[string]interface{}{"command": "echo"}},
			},
			Clients: map[string]*models.Client{
				"test_client": {ConfigPath: "~/.test.json", Enabled: []string{"test-server"}},
			},
		}

		if err := validator.ValidateConfig(cfg); err != nil {
			t.Errorf("ValidateConfig() unexpected error: %v", err)
		}
	})

	t.Run("Invalid port - too low", func(t *testing.T) {
		cfg := &models.Config{
			ServerPort: 0,
			MCPServers: []models.MCPServer{{Name: "test", Config: map[string]interface{}{"command": "echo"}}},
			Clients:    map[string]*models.Client{"test": {ConfigPath: "~/.test.json"}},
		}

		err := validator.ValidateConfig(cfg)
		if err == nil {
			t.Error("Expected error for port 0")
		}
		if !containsSubstring(err.Error(), "invalid server port") {
			t.Errorf("Expected 'invalid server port' error, got: %v", err)
		}
	})

	t.Run("Invalid port - too high", func(t *testing.T) {
		cfg := &models.Config{
			ServerPort: 70000,
			MCPServers: []models.MCPServer{{Name: "test", Config: map[string]interface{}{"command": "echo"}}},
			Clients:    map[string]*models.Client{"test": {ConfigPath: "~/.test.json"}},
		}

		err := validator.ValidateConfig(cfg)
		if err == nil {
			t.Error("Expected error for port 70000")
		}
	})

	t.Run("No servers configured", func(t *testing.T) {
		cfg := &models.Config{
			ServerPort: 6543,
			MCPServers: []models.MCPServer{},
			Clients:    map[string]*models.Client{"test": {ConfigPath: "~/.test.json"}},
		}

		err := validator.ValidateConfig(cfg)
		if err == nil {
			t.Error("Expected error for no servers")
		}
		if !containsSubstring(err.Error(), "no MCP servers configured") {
			t.Errorf("Expected 'no MCP servers' error, got: %v", err)
		}
	})

	t.Run("No clients configured", func(t *testing.T) {
		cfg := &models.Config{
			ServerPort: 6543,
			MCPServers: []models.MCPServer{{Name: "test", Config: map[string]interface{}{"command": "echo"}}},
			Clients:    map[string]*models.Client{},
		}

		err := validator.ValidateConfig(cfg)
		if err == nil {
			t.Error("Expected error for no clients")
		}
		if !containsSubstring(err.Error(), "no clients configured") {
			t.Errorf("Expected 'no clients' error, got: %v", err)
		}
	})

	t.Run("Client references non-existent server", func(t *testing.T) {
		cfg := &models.Config{
			ServerPort: 6543,
			MCPServers: []models.MCPServer{{Name: "server1", Config: map[string]interface{}{"command": "echo"}}},
			Clients: map[string]*models.Client{
				"test": {ConfigPath: "~/.test.json", Enabled: []string{"nonexistent"}},
			},
		}

		err := validator.ValidateConfig(cfg)
		if err == nil {
			t.Error("Expected error for non-existent server reference")
		}
		if !containsSubstring(err.Error(), "references non-existent server") {
			t.Errorf("Expected 'references non-existent server' error, got: %v", err)
		}
	})

	t.Run("Invalid server in config", func(t *testing.T) {
		cfg := &models.Config{
			ServerPort: 6543,
			MCPServers: []models.MCPServer{
				{Name: "", Config: map[string]interface{}{"command": "echo"}}, // Empty name
			},
			Clients: map[string]*models.Client{"test": {ConfigPath: "~/.test.json"}},
		}

		err := validator.ValidateConfig(cfg)
		if err == nil {
			t.Error("Expected error for invalid server")
		}
	})
}

func TestValidateClientConfig(t *testing.T) {
	validator := NewValidatorService()

	t.Run("Valid client config with STDIO server", func(t *testing.T) {
		clientCfg := &models.ClientConfig{
			MCPServers: map[string]interface{}{
				"filesystem": map[string]interface{}{
					"command": "npx",
					"args":    []interface{}{"test"},
				},
			},
		}

		if err := validator.ValidateClientConfig(clientCfg); err != nil {
			t.Errorf("ValidateClientConfig() unexpected error: %v", err)
		}
	})

	t.Run("Valid client config with HTTP server", func(t *testing.T) {
		clientCfg := &models.ClientConfig{
			MCPServers: map[string]interface{}{
				"context7": map[string]interface{}{
					"httpUrl": "https://example.com",
				},
			},
		}

		if err := validator.ValidateClientConfig(clientCfg); err != nil {
			t.Errorf("ValidateClientConfig() unexpected error: %v", err)
		}
	})

	t.Run("Nil mcpServers", func(t *testing.T) {
		clientCfg := &models.ClientConfig{
			MCPServers: nil,
		}

		if err := validator.ValidateClientConfig(clientCfg); err != nil {
			t.Errorf("ValidateClientConfig() should accept nil mcpServers, got error: %v", err)
		}
	})

	t.Run("Empty server name", func(t *testing.T) {
		clientCfg := &models.ClientConfig{
			MCPServers: map[string]interface{}{
				"": map[string]interface{}{"command": "echo"},
			},
		}

		err := validator.ValidateClientConfig(clientCfg)
		if err == nil {
			t.Error("Expected error for empty server name")
		}
	})

	t.Run("Server with no transport type", func(t *testing.T) {
		clientCfg := &models.ClientConfig{
			MCPServers: map[string]interface{}{
				"invalid": map[string]interface{}{
					"args": []interface{}{"test"},
				},
			},
		}

		err := validator.ValidateClientConfig(clientCfg)
		if err == nil {
			t.Error("Expected error for missing transport type")
		}
	})

	t.Run("Server with multiple transport types", func(t *testing.T) {
		clientCfg := &models.ClientConfig{
			MCPServers: map[string]interface{}{
				"multi": map[string]interface{}{
					"command": "echo",
					"httpUrl": "https://example.com",
				},
			},
		}

		err := validator.ValidateClientConfig(clientCfg)
		if err == nil {
			t.Error("Expected error for multiple transport types")
		}
	})

	t.Run("Empty command string", func(t *testing.T) {
		clientCfg := &models.ClientConfig{
			MCPServers: map[string]interface{}{
				"empty-cmd": map[string]interface{}{
					"command": "",
				},
			},
		}

		err := validator.ValidateClientConfig(clientCfg)
		if err == nil {
			t.Error("Expected error for empty command")
		}
	})
}

func TestIsCommandAvailable(t *testing.T) {
	validator := NewValidatorService()

	t.Run("Available command", func(t *testing.T) {
		if !validator.IsCommandAvailable("echo") {
			t.Error("Expected 'echo' to be available")
		}
	})

	t.Run("Unavailable command", func(t *testing.T) {
		if validator.IsCommandAvailable("nonexistent-cmd-xyz123") {
			t.Error("Expected nonexistent command to be unavailable")
		}
	})
}

// Helper function to check if a string contains a substring
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