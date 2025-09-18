package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/vlazic/mcp-server-manager/internal/models"
	"github.com/vlazic/mcp-server-manager/internal/services"
)

func TestAPIHandler_AddServer(t *testing.T) {
	// Set gin to test mode
	gin.SetMode(gin.TestMode)

	// Create a temporary config file for testing
	tempDir, err := os.MkdirTemp("", "api_test")
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
				EnabledGlobally: false,
				Clients:         map[string]bool{"claude_code": false},
			},
		},
		Clients: []models.Client{
			{Name: "claude_code", ConfigPath: "~/.claude/settings.json"},
			{Name: "gemini_cli", ConfigPath: "~/.gemini/settings.json"},
		},
	}

	manager := services.NewMCPManagerService(config, configPath)
	handler := NewAPIHandler(manager)

	tests := []struct {
		name           string
		requestBody    interface{}
		expectedStatus int
		expectError    bool
		errorContains  string
	}{
		{
			name: "Valid STDIO server",
			requestBody: models.MCPServer{
				Name:            "new-stdio-server",
				Command:         "echo",
				Args:            []string{"test"},
				Env:             map[string]string{"TEST": "value"},
				Timeout:         30000,
				EnabledGlobally: false,
			},
			expectedStatus: http.StatusOK,
			expectError:    false,
		},
		{
			name: "Valid HTTP server",
			requestBody: models.MCPServer{
				Name:            "new-http-server",
				HttpURL:         "https://example.com/mcp",
				Headers:         map[string]string{"Authorization": "Bearer token"},
				Timeout:         15000,
				EnabledGlobally: false,
			},
			expectedStatus: http.StatusOK,
			expectError:    false,
		},
		{
			name: "Valid SSE server",
			requestBody: models.MCPServer{
				Name:            "new-sse-server",
				URL:             "http://localhost:8080/sse",
				Headers:         map[string]string{"X-API-Key": "key123"},
				Timeout:         10000,
				EnabledGlobally: false,
			},
			expectedStatus: http.StatusOK,
			expectError:    false,
		},
		{
			name: "Duplicate server name",
			requestBody: models.MCPServer{
				Name:            "existing-server",
				Command:         "echo",
				EnabledGlobally: false,
			},
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
			errorContains:  "server with name 'existing-server' already exists",
		},
		{
			name: "Invalid JSON - missing required field",
			requestBody: map[string]interface{}{
				"command":          "echo",
				"enabled_globally": false,
				// missing name
			},
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
			errorContains:  "server validation failed",
		},
		{
			name: "Invalid server - no transport",
			requestBody: models.MCPServer{
				Name:            "invalid-server",
				EnabledGlobally: false,
				// no transport type specified
			},
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
			errorContains:  "server validation failed",
		},
		{
			name: "Invalid server - multiple transports",
			requestBody: models.MCPServer{
				Name:            "multi-transport",
				Command:         "echo",
				HttpURL:         "https://example.com",
				EnabledGlobally: false,
			},
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
			errorContains:  "server validation failed",
		},
		{
			name:           "Invalid JSON syntax",
			requestBody:    "{invalid json",
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
			errorContains:  "Invalid JSON",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset the config to initial state for each test
			manager.GetConfig().MCPServers = []models.MCPServer{
				{
					Name:            "existing-server",
					Command:         "echo",
					EnabledGlobally: false,
					Clients:         map[string]bool{"claude_code": false},
				},
			}

			// Prepare request body
			var bodyBytes []byte
			if str, ok := tt.requestBody.(string); ok {
				// For invalid JSON test
				bodyBytes = []byte(str)
			} else {
				bodyBytes, err = json.Marshal(tt.requestBody)
				if err != nil {
					t.Fatalf("Failed to marshal request body: %v", err)
				}
			}

			// Create request
			req, err := http.NewRequest("POST", "/api/servers", bytes.NewReader(bodyBytes))
			if err != nil {
				t.Fatalf("Failed to create request: %v", err)
			}
			req.Header.Set("Content-Type", "application/json")

			// Create response recorder
			w := httptest.NewRecorder()

			// Create gin context
			router := gin.New()
			router.POST("/api/servers", handler.AddServer)

			// Perform request
			router.ServeHTTP(w, req)

			// Check status code
			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			// Parse response
			var response map[string]interface{}
			err = json.Unmarshal(w.Body.Bytes(), &response)
			if err != nil {
				t.Fatalf("Failed to parse response: %v", err)
			}

			if tt.expectError {
				// Check that error is present
				if errorMsg, exists := response["error"]; !exists {
					t.Error("Expected error in response but none found")
				} else if tt.errorContains != "" {
					errorStr := errorMsg.(string)
					if !containsString(errorStr, tt.errorContains) {
						t.Errorf("Expected error to contain %q, got %q", tt.errorContains, errorStr)
					}
				}
			} else {
				// Check that success is true
				if success, exists := response["success"]; !exists || success != true {
					t.Errorf("Expected success=true in response, got %v", response)
				}

				// Check that server object is returned
				if _, exists := response["server"]; !exists {
					t.Error("Expected server object in response")
				}
			}
		})
	}
}

func TestAPIHandler_AddServerClientInitialization(t *testing.T) {
	// Set gin to test mode
	gin.SetMode(gin.TestMode)

	// Create a temporary config file for testing
	tempDir, err := os.MkdirTemp("", "api_test")
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
			{Name: "claude_code", ConfigPath: "~/.claude/settings.json"},
			{Name: "gemini_cli", ConfigPath: "~/.gemini/settings.json"},
			{Name: "custom_client", ConfigPath: "~/.custom/settings.json"},
		},
	}

	manager := services.NewMCPManagerService(config, configPath)
	handler := NewAPIHandler(manager)

	// Create server without clients map
	server := models.MCPServer{
		Name:            "test-server",
		Command:         "echo",
		EnabledGlobally: false,
		// Clients map is nil
	}

	bodyBytes, err := json.Marshal(server)
	if err != nil {
		t.Fatalf("Failed to marshal request body: %v", err)
	}

	// Create request
	req, err := http.NewRequest("POST", "/api/servers", bytes.NewReader(bodyBytes))
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	// Create response recorder
	w := httptest.NewRecorder()

	// Create gin context
	router := gin.New()
	router.POST("/api/servers", handler.AddServer)

	// Perform request
	router.ServeHTTP(w, req)

	// Check status code
	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	// Verify clients were initialized
	addedServer := manager.GetConfig().MCPServers[0]
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

// Helper function for testing
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