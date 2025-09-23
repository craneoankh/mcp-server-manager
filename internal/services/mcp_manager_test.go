package services

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/vlazic/mcp-server-manager/internal/models"
)

func TestAddServer(t *testing.T) {
	// Create a temporary config file for testing
	tempDir, err := os.MkdirTemp("", "mcp_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	configPath := filepath.Join(tempDir, "config.yaml")

	// Create initial config
	config := &models.Config{
		ServerPort: 6543,
		MCPServers: []models.MCPServer{
			{
				Name:            "existing-server",
				Command:         "echo",
				Clients:         map[string]bool{"claude_code": false},
			},
		},
		Clients: []models.Client{
			{Name: "claude_code", ConfigPath: "~/.claude.json"},
			{Name: "gemini_cli", ConfigPath: "~/.gemini/settings.json"},
		},
	}

	manager := NewMCPManagerService(config, configPath)

	tests := []struct {
		name    string
		server  models.MCPServer
		wantErr bool
		errMsg  string
	}{
		{
			name: "Valid STDIO server",
			server: models.MCPServer{
				Name:            "new-stdio-server",
				Command:         "echo",
				Args:            []string{"test"},
				Env:             map[string]string{"TEST": "value"},
				Timeout:         30000,
			},
			wantErr: false,
		},
		{
			name: "Valid HTTP server",
			server: models.MCPServer{
				Name:            "new-http-server",
				HttpURL:         "https://example.com/mcp",
				Headers:         map[string]string{"Authorization": "Bearer token"},
				Timeout:         15000,
			},
			wantErr: false,
		},
		{
			name: "Valid SSE server",
			server: models.MCPServer{
				Name:            "new-sse-server",
				URL:             "http://localhost:8080/sse",
				Headers:         map[string]string{"X-API-Key": "key123"},
				Timeout:         10000,
			},
			wantErr: false,
		},
		{
			name: "Duplicate server name",
			server: models.MCPServer{
				Name:            "existing-server",
				Command:         "echo",
			},
			wantErr: true,
			errMsg:  "server with name 'existing-server' already exists",
		},
		{
			name: "Invalid server - no transport",
			server: models.MCPServer{
				Name:            "invalid-server",
			},
			wantErr: true,
			errMsg:  "server validation failed",
		},
		{
			name: "Invalid server - empty name",
			server: models.MCPServer{
				Command:         "echo",
			},
			wantErr: true,
			errMsg:  "server validation failed",
		},
		{
			name: "Invalid server - multiple transports",
			server: models.MCPServer{
				Name:            "multi-transport",
				Command:         "echo",
				HttpURL:         "https://example.com",
			},
			wantErr: true,
			errMsg:  "server validation failed",
		},
		{
			name: "Invalid command",
			server: models.MCPServer{
				Name:            "invalid-command-server",
				Command:         "non_existent_command_12345",
			},
			wantErr: true,
			errMsg:  "server validation failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset the config to initial state for each test
			manager.config.MCPServers = []models.MCPServer{
				{
					Name:            "existing-server",
					Command:         "echo",
						Clients:         map[string]bool{"claude_code": false},
				},
			}

			err := manager.AddServer(&tt.server)

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
					return
				}

				// Verify the server was added
				found := false
				for _, server := range manager.config.MCPServers {
					if server.Name == tt.server.Name {
						found = true

						// Verify clients map was initialized
						if server.Clients == nil {
							t.Error("Expected clients map to be initialized")
						} else {
							// Check that all existing clients are set to false by default
							for _, client := range manager.config.Clients {
								if enabled, exists := server.Clients[client.Name]; !exists || enabled {
									t.Errorf("Expected client %s to be false by default, got %v", client.Name, enabled)
								}
							}
						}
						break
					}
				}

				if !found {
					t.Errorf("Server %s was not added to the config", tt.server.Name)
				}
			}
		})
	}
}

func TestAddServerClientInitialization(t *testing.T) {
	// Create a temporary config file for testing
	tempDir, err := os.MkdirTemp("", "mcp_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	configPath := filepath.Join(tempDir, "config.yaml")

	// Create config with multiple clients
	config := &models.Config{
		ServerPort: 6543,
		MCPServers: []models.MCPServer{},
		Clients: []models.Client{
			{Name: "claude_code", ConfigPath: "~/.claude.json"},
			{Name: "gemini_cli", ConfigPath: "~/.gemini/settings.json"},
			{Name: "custom_client", ConfigPath: "~/.custom/settings.json"},
		},
	}

	manager := NewMCPManagerService(config, configPath)

	// Test adding server without clients map
	server := models.MCPServer{
		Name:            "test-server",
		Command:         "echo",
		// Clients map is nil
	}

	err = manager.AddServer(&server)
	if err != nil {
		t.Fatalf("Failed to add server: %v", err)
	}

	// Verify all clients were initialized to false
	addedServer := manager.config.MCPServers[0]
	if addedServer.Clients == nil {
		t.Fatal("Clients map was not initialized")
	}

	expectedClients := []string{"claude_code", "gemini_cli", "custom_client"}
	for _, clientName := range expectedClients {
		if enabled, exists := addedServer.Clients[clientName]; !exists {
			t.Errorf("Client %s was not initialized", clientName)
		} else if enabled {
			t.Errorf("Client %s should be false by default, got true", clientName)
		}
	}
}

