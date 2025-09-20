package services

import (
	"testing"

	"github.com/vlazic/mcp-server-manager/internal/models"
)

func TestValidateMCPServer(t *testing.T) {
	validator := NewValidatorService()

	tests := []struct {
		name    string
		server  models.MCPServer
		wantErr bool
		errMsg  string
	}{
		{
			name: "Valid STDIO server",
			server: models.MCPServer{
				Name:            "filesystem",
				Command:         "echo", // Using echo since it's always available
				Args:            []string{"test"},
				Env:             map[string]string{"TEST": "value"},
				Timeout:         5000,
				Clients:         map[string]bool{"claude": true},
			},
			wantErr: false,
		},
		{
			name: "Valid HTTP server",
			server: models.MCPServer{
				Name:            "context7",
				HttpURL:         "https://mcp.context7.com/mcp",
				Headers:         map[string]string{"Authorization": "Bearer token"},
				Timeout:         5000,
				Clients:         map[string]bool{"gemini": true},
			},
			wantErr: false,
		},
		{
			name: "Valid SSE server",
			server: models.MCPServer{
				Name:            "sse_server",
				URL:             "http://localhost:8080/sse",
				Headers:         map[string]string{"X-API-Key": "key123"},
				Timeout:         10000,
				Clients:         map[string]bool{"claude": true},
			},
			wantErr: false,
		},
		{
			name: "Empty name",
			server: models.MCPServer{
				Command: "echo",
			},
			wantErr: true,
			errMsg:  "server name cannot be empty",
		},
		{
			name: "No transport type",
			server: models.MCPServer{
				Name: "invalid",
			},
			wantErr: true,
			errMsg:  "server must have exactly one transport type",
		},
		{
			name: "Multiple transport types",
			server: models.MCPServer{
				Name:    "invalid",
				Command: "echo",
				HttpURL: "https://example.com",
			},
			wantErr: true,
			errMsg:  "server must have exactly one transport type, found 2",
		},
		{
			name: "Invalid command",
			server: models.MCPServer{
				Name:    "invalid",
				Command: "non_existent_command_12345",
			},
			wantErr: true,
			errMsg:  "command 'non_existent_command_12345' not found in PATH",
		},
		{
			name: "Invalid URL",
			server: models.MCPServer{
				Name: "invalid",
				URL:  "not-a-valid-url",
			},
			wantErr: true,
			errMsg:  "invalid SSE URL",
		},
		{
			name: "Invalid HTTP URL",
			server: models.MCPServer{
				Name:    "invalid",
				HttpURL: "not-a-valid-url",
			},
			wantErr: true,
			errMsg:  "invalid HTTP URL",
		},
		{
			name: "Negative timeout",
			server: models.MCPServer{
				Name:    "invalid",
				Command: "echo",
				Timeout: -1,
			},
			wantErr: true,
			errMsg:  "timeout cannot be negative",
		},
		{
			name: "Empty env key",
			server: models.MCPServer{
				Name:    "invalid",
				Command: "echo",
				Env:     map[string]string{"": "value"},
			},
			wantErr: true,
			errMsg:  "environment variable key cannot be empty",
		},
		{
			name: "Env key with equals",
			server: models.MCPServer{
				Name:    "invalid",
				Command: "echo",
				Env:     map[string]string{"KEY=INVALID": "value"},
			},
			wantErr: true,
			errMsg:  "environment variable key cannot contain '='",
		},
		{
			name: "Empty env value",
			server: models.MCPServer{
				Name:    "invalid",
				Command: "echo",
				Env:     map[string]string{"KEY": ""},
			},
			wantErr: true,
			errMsg:  "environment variable value for 'KEY' cannot be empty",
		},
		{
			name: "Empty include tool",
			server: models.MCPServer{
				Name:         "invalid",
				Command:      "echo",
				IncludeTools: []string{"valid_tool", ""},
			},
			wantErr: true,
			errMsg:  "include_tools cannot contain empty tool names",
		},
		{
			name: "Empty exclude tool",
			server: models.MCPServer{
				Name:         "invalid",
				Command:      "echo",
				ExcludeTools: []string{"valid_tool", ""},
			},
			wantErr: true,
			errMsg:  "exclude_tools cannot contain empty tool names",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateMCPServer(&tt.server)

			if tt.wantErr {
				if err == nil {
					t.Errorf("Expected error but got none")
					return
				}
				if tt.errMsg != "" && !containsString(err.Error(), tt.errMsg) {
					t.Errorf("Expected error to contain %q, got %q", tt.errMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
				}
			}
		})
	}
}

func TestValidateConfig(t *testing.T) {
	validator := NewValidatorService()

	tests := []struct {
		name    string
		config  models.Config
		wantErr bool
		errMsg  string
	}{
		{
			name: "Valid config",
			config: models.Config{
				ServerPort: 6543,
				MCPServers: []models.MCPServer{
					{
						Name:            "filesystem",
						Command:         "echo",
								Clients:         map[string]bool{"claude": true},
					},
				},
				Clients: []models.Client{
					{
						Name:       "claude",
						ConfigPath: "~/.claude/settings.json",
					},
				},
			},
			wantErr: false,
		},
		{
			name: "Invalid port - too low",
			config: models.Config{
				ServerPort: 0,
				MCPServers: []models.MCPServer{
					{
						Name:    "test",
						Command: "echo",
					},
				},
				Clients: []models.Client{
					{Name: "claude", ConfigPath: "~/.claude/settings.json"},
				},
			},
			wantErr: true,
			errMsg:  "invalid server port: 0",
		},
		{
			name: "Invalid port - too high",
			config: models.Config{
				ServerPort: 70000,
				MCPServers: []models.MCPServer{
					{
						Name:    "test",
						Command: "echo",
					},
				},
				Clients: []models.Client{
					{Name: "claude", ConfigPath: "~/.claude/settings.json"},
				},
			},
			wantErr: true,
			errMsg:  "invalid server port: 70000",
		},
		{
			name: "No MCP servers",
			config: models.Config{
				ServerPort: 6543,
				MCPServers: []models.MCPServer{},
				Clients: []models.Client{
					{Name: "claude", ConfigPath: "~/.claude/settings.json"},
				},
			},
			wantErr: true,
			errMsg:  "no MCP servers configured",
		},
		{
			name: "No clients",
			config: models.Config{
				ServerPort: 6543,
				MCPServers: []models.MCPServer{
					{
						Name:    "test",
						Command: "echo",
					},
				},
				Clients: []models.Client{},
			},
			wantErr: true,
			errMsg:  "no clients configured",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateConfig(&tt.config)

			if tt.wantErr {
				if err == nil {
					t.Errorf("Expected error but got none")
					return
				}
				if tt.errMsg != "" && !containsString(err.Error(), tt.errMsg) {
					t.Errorf("Expected error to contain %q, got %q", tt.errMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
				}
			}
		})
	}
}

func TestValidateClient(t *testing.T) {
	validator := NewValidatorService()

	tests := []struct {
		name    string
		client  models.Client
		wantErr bool
		errMsg  string
	}{
		{
			name: "Valid client",
			client: models.Client{
				Name:       "claude",
				ConfigPath: "~/.claude/settings.json",
			},
			wantErr: false,
		},
		{
			name: "Empty name",
			client: models.Client{
				ConfigPath: "~/.claude/settings.json",
			},
			wantErr: true,
			errMsg:  "client name cannot be empty",
		},
		{
			name: "Empty config path",
			client: models.Client{
				Name: "claude",
			},
			wantErr: true,
			errMsg:  "client config path cannot be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateClient(&tt.client)

			if tt.wantErr {
				if err == nil {
					t.Errorf("Expected error but got none")
					return
				}
				if tt.errMsg != "" && !containsString(err.Error(), tt.errMsg) {
					t.Errorf("Expected error to contain %q, got %q", tt.errMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
				}
			}
		})
	}
}

// Helper function to check if a string contains a substring
func containsString(s, substr string) bool {
	return len(substr) == 0 || (len(s) >= len(substr) &&
		func() bool {
			for i := 0; i <= len(s)-len(substr); i++ {
				if s[i:i+len(substr)] == substr {
					return true
				}
			}
			return false
		}())
}

func TestValidateClientConfig_Context7(t *testing.T) {
	validator := NewValidatorService()

	// Test Context7's official SSE configuration format
	context7Config := &models.ClientConfig{
		MCPServers: map[string]interface{}{
			"context7": map[string]interface{}{
				"url": "https://mcp.context7.com/mcp",
				"headers": map[string]interface{}{
					"CONTEXT7_API_KEY": "YOUR_API_KEY",
				},
			},
		},
	}

	err := validator.ValidateClientConfig(context7Config)
	if err != nil {
		t.Errorf("Context7 configuration should be valid, got error: %v", err)
	}
}

func TestValidateClientConfig_AllTransportTypes(t *testing.T) {
	validator := NewValidatorService()

	tests := []struct {
		name   string
		config *models.ClientConfig
		valid  bool
	}{
		{
			name: "STDIO server",
			config: &models.ClientConfig{
				MCPServers: map[string]interface{}{
					"stdio-server": map[string]interface{}{
						"command": "npx",
						"args":    []interface{}{"@my/server"},
					},
				},
			},
			valid: true,
		},
		{
			name: "HTTP server",
			config: &models.ClientConfig{
				MCPServers: map[string]interface{}{
					"http-server": map[string]interface{}{
						"httpUrl": "https://api.example.com/mcp",
						"headers": map[string]interface{}{
							"Authorization": "Bearer token",
						},
					},
				},
			},
			valid: true,
		},
		{
			name: "SSE server",
			config: &models.ClientConfig{
				MCPServers: map[string]interface{}{
					"sse-server": map[string]interface{}{
						"url": "http://localhost:8080/sse",
						"headers": map[string]interface{}{
							"X-API-Key": "key123",
						},
					},
				},
			},
			valid: true,
		},
		{
			name: "Server with multiple transports (invalid)",
			config: &models.ClientConfig{
				MCPServers: map[string]interface{}{
					"multi-transport": map[string]interface{}{
						"command": "npx",
						"httpUrl": "https://example.com",
					},
				},
			},
			valid: false,
		},
		{
			name: "Server with no transport (invalid)",
			config: &models.ClientConfig{
				MCPServers: map[string]interface{}{
					"no-transport": map[string]interface{}{
						"headers": map[string]interface{}{
							"X-Key": "value",
						},
					},
				},
			},
			valid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateClientConfig(tt.config)
			if tt.valid && err != nil {
				t.Errorf("Expected valid config but got error: %v", err)
			}
			if !tt.valid && err == nil {
				t.Errorf("Expected invalid config but validation passed")
			}
		})
	}
}