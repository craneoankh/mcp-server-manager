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
	"github.com/vlazic/mcp-server-manager/internal/config"
	"github.com/vlazic/mcp-server-manager/internal/models"
	"github.com/vlazic/mcp-server-manager/internal/services"
)

// setupTestAPIHandler creates a test API handler with a temporary config file
func setupTestAPIHandler(t *testing.T) (*APIHandler, string, func()) {
	t.Helper()

	// Create temp directory
	tempDir, err := os.MkdirTemp("", "api_test_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	// Create config file path
	configPath := filepath.Join(tempDir, "config.yaml")

	// Create initial config
	cfg := &models.Config{
		ServerPort: 6543,
		MCPServers: []models.MCPServer{
			{
				Name: "test-server",
				Config: map[string]interface{}{
					"command": "npx",
					"args":    []interface{}{"-y", "@modelcontextprotocol/server-filesystem", "/tmp"},
				},
			},
		},
		Clients: map[string]*models.Client{
			"test-client": {
				ConfigPath: filepath.Join(tempDir, "client.json"),
				Enabled:    []string{"test-server"},
			},
		},
	}

	// Save config
	if err := config.SaveConfig(cfg, configPath); err != nil {
		os.RemoveAll(tempDir)
		t.Fatalf("Failed to save config: %v", err)
	}

	// Create MCP manager
	mcpManager := services.NewMCPManagerService(cfg, configPath)

	// Create handler
	handler := NewAPIHandler(mcpManager)

	// Cleanup function
	cleanup := func() {
		os.RemoveAll(tempDir)
	}

	return handler, tempDir, cleanup
}

// TestAddServer_Success tests successful server addition with v2.0 format
func TestAddServer_Success(t *testing.T) {
	handler, _, cleanup := setupTestAPIHandler(t)
	defer cleanup()

	// Create test router
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/api/servers", handler.AddServer)

	// Prepare request body in v2.0 format
	requestBody := map[string]interface{}{
		"mcpServers": map[string]interface{}{
			"cloudflare": map[string]interface{}{
				"command": "npx",
				"args":    []string{"mcp-remote", "https://docs.mcp.cloudflare.com/sse"},
			},
		},
	}

	jsonData, _ := json.Marshal(requestBody)

	// Make request
	req, _ := http.NewRequest("POST", "/api/servers", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert response
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d. Body: %s", w.Code, w.Body.String())
	}

	var response map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if success, ok := response["success"].(bool); !ok || !success {
		t.Error("Expected success=true in response")
	}

	if server, ok := response["server"].(map[string]interface{}); !ok {
		t.Error("Expected server object in response")
	} else {
		if name, ok := server["name"].(string); !ok || name != "cloudflare" {
			t.Errorf("Expected server name 'cloudflare', got %v", server["name"])
		}
	}
}

// TestAddServer_InvalidJSON tests handling of malformed JSON
func TestAddServer_InvalidJSON(t *testing.T) {
	handler, _, cleanup := setupTestAPIHandler(t)
	defer cleanup()

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/api/servers", handler.AddServer)

	// Send invalid JSON
	req, _ := http.NewRequest("POST", "/api/servers", bytes.NewBufferString("{invalid json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	if err, ok := response["error"].(string); !ok || err == "" {
		t.Error("Expected error message in response")
	}
}

// TestAddServer_MissingMCPServers tests missing mcpServers key
func TestAddServer_MissingMCPServers(t *testing.T) {
	handler, _, cleanup := setupTestAPIHandler(t)
	defer cleanup()

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/api/servers", handler.AddServer)

	// Send request without mcpServers key
	requestBody := map[string]interface{}{
		"name":    "cloudflare",
		"command": "npx",
	}
	jsonData, _ := json.Marshal(requestBody)

	req, _ := http.NewRequest("POST", "/api/servers", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	errorMsg, _ := response["error"].(string)
	if errorMsg != "Must provide exactly one server in mcpServers" {
		t.Errorf("Expected specific error message, got: %s", errorMsg)
	}
}

// TestAddServer_EmptyMCPServers tests empty mcpServers object
func TestAddServer_EmptyMCPServers(t *testing.T) {
	handler, _, cleanup := setupTestAPIHandler(t)
	defer cleanup()

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/api/servers", handler.AddServer)

	// Send request with empty mcpServers
	requestBody := map[string]interface{}{
		"mcpServers": map[string]interface{}{},
	}
	jsonData, _ := json.Marshal(requestBody)

	req, _ := http.NewRequest("POST", "/api/servers", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	errorMsg, _ := response["error"].(string)
	if errorMsg != "Must provide exactly one server in mcpServers" {
		t.Errorf("Expected 'exactly one server' error, got: %s", errorMsg)
	}
}

// TestAddServer_MultipleServers tests rejection of multiple servers
func TestAddServer_MultipleServers(t *testing.T) {
	handler, _, cleanup := setupTestAPIHandler(t)
	defer cleanup()

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/api/servers", handler.AddServer)

	// Send request with multiple servers (should be rejected)
	requestBody := map[string]interface{}{
		"mcpServers": map[string]interface{}{
			"cloudflare": map[string]interface{}{
				"command": "npx",
				"args":    []string{"mcp-remote", "https://docs.mcp.cloudflare.com/sse"},
			},
			"another-server": map[string]interface{}{
				"command": "echo",
			},
		},
	}
	jsonData, _ := json.Marshal(requestBody)

	req, _ := http.NewRequest("POST", "/api/servers", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	errorMsg, _ := response["error"].(string)
	if errorMsg != "Must provide exactly one server in mcpServers" {
		t.Errorf("Expected 'exactly one server' error, got: %s", errorMsg)
	}
}

// TestAddServer_InvalidServerConfig tests server with invalid configuration
func TestAddServer_InvalidServerConfig(t *testing.T) {
	handler, _, cleanup := setupTestAPIHandler(t)
	defer cleanup()

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/api/servers", handler.AddServer)

	// Send server without transport type (command, httpUrl, or url)
	requestBody := map[string]interface{}{
		"mcpServers": map[string]interface{}{
			"invalid-server": map[string]interface{}{
				"description": "Missing transport type",
			},
		},
	}
	jsonData, _ := json.Marshal(requestBody)

	req, _ := http.NewRequest("POST", "/api/servers", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	if err, ok := response["error"].(string); !ok || err == "" {
		t.Error("Expected error message for invalid server config")
	}
}

// TestAddServer_DuplicateServer tests adding a server that already exists
func TestAddServer_DuplicateServer(t *testing.T) {
	handler, _, cleanup := setupTestAPIHandler(t)
	defer cleanup()

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/api/servers", handler.AddServer)

	// Try to add server with name that already exists ("test-server" from setup)
	requestBody := map[string]interface{}{
		"mcpServers": map[string]interface{}{
			"test-server": map[string]interface{}{
				"command": "npx",
				"args":    []string{"-y", "@modelcontextprotocol/server-filesystem"},
			},
		},
	}
	jsonData, _ := json.Marshal(requestBody)

	req, _ := http.NewRequest("POST", "/api/servers", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400 for duplicate server, got %d", w.Code)
	}

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	errorMsg, _ := response["error"].(string)
	if errorMsg == "" {
		t.Error("Expected error message for duplicate server")
	}
}

// TestAddServer_HTTPServer tests adding an HTTP-based server
func TestAddServer_HTTPServer(t *testing.T) {
	handler, _, cleanup := setupTestAPIHandler(t)
	defer cleanup()

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/api/servers", handler.AddServer)

	// Add HTTP server with httpUrl
	requestBody := map[string]interface{}{
		"mcpServers": map[string]interface{}{
			"http-server": map[string]interface{}{
				"httpUrl": "http://localhost:3000/mcp",
				"headers": map[string]string{
					"Authorization": "Bearer token123",
				},
			},
		},
	}
	jsonData, _ := json.Marshal(requestBody)

	req, _ := http.NewRequest("POST", "/api/servers", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d. Body: %s", w.Code, w.Body.String())
	}

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	if success, _ := response["success"].(bool); !success {
		t.Error("Expected successful addition of HTTP server")
	}
}

// TestAddServer_SSEServer tests adding an SSE-based server
func TestAddServer_SSEServer(t *testing.T) {
	handler, _, cleanup := setupTestAPIHandler(t)
	defer cleanup()

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/api/servers", handler.AddServer)

	// Add SSE server with url field
	requestBody := map[string]interface{}{
		"mcpServers": map[string]interface{}{
			"sse-server": map[string]interface{}{
				"url": "https://example.com/sse",
			},
		},
	}
	jsonData, _ := json.Marshal(requestBody)

	req, _ := http.NewRequest("POST", "/api/servers", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d. Body: %s", w.Code, w.Body.String())
	}

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	if success, _ := response["success"].(bool); !success {
		t.Error("Expected successful addition of SSE server")
	}
}

// TestToggleClientServer_Enable tests enabling a server for a client
func TestToggleClientServer_Enable(t *testing.T) {
	handler, tempDir, cleanup := setupTestAPIHandler(t)
	defer cleanup()

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/api/clients/:client/servers/:server/toggle", handler.ToggleClientServer)

	// Create client config file first
	clientPath := filepath.Join(tempDir, "client.json")
	clientData := map[string]interface{}{
		"mcpServers": map[string]interface{}{},
	}
	jsonData, _ := json.Marshal(clientData)
	os.WriteFile(clientPath, jsonData, 0644)

	// Toggle test-server to enabled for test-client
	req, _ := http.NewRequest("POST", "/api/clients/test-client/servers/test-server/toggle", nil)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.PostForm = map[string][]string{
		"enabled": {"true"},
	}
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d. Body: %s", w.Code, w.Body.String())
	}

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	if success, _ := response["success"].(bool); !success {
		t.Error("Expected success=true")
	}
}

// TestToggleClientServer_Disable tests disabling a server for a client
func TestToggleClientServer_Disable(t *testing.T) {
	handler, tempDir, cleanup := setupTestAPIHandler(t)
	defer cleanup()

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/api/clients/:client/servers/:server/toggle", handler.ToggleClientServer)

	// Create client config file with server enabled
	clientPath := filepath.Join(tempDir, "client.json")
	clientData := map[string]interface{}{
		"mcpServers": map[string]interface{}{
			"test-server": map[string]interface{}{
				"command": "npx",
				"args":    []string{"-y", "@modelcontextprotocol/server-filesystem"},
			},
		},
	}
	jsonData, _ := json.Marshal(clientData)
	os.WriteFile(clientPath, jsonData, 0644)

	// Toggle test-server to disabled for test-client
	req, _ := http.NewRequest("POST", "/api/clients/test-client/servers/test-server/toggle", nil)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.PostForm = map[string][]string{
		"enabled": {"false"},
	}
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d. Body: %s", w.Code, w.Body.String())
	}

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	if success, _ := response["success"].(bool); !success {
		t.Error("Expected success=true")
	}
}

// TestToggleClientServer_InvalidEnabledValue tests invalid enabled parameter
func TestToggleClientServer_InvalidEnabledValue(t *testing.T) {
	handler, _, cleanup := setupTestAPIHandler(t)
	defer cleanup()

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/api/clients/:client/servers/:server/toggle", handler.ToggleClientServer)

	// Send invalid enabled value
	req, _ := http.NewRequest("POST", "/api/clients/test-client/servers/test-server/toggle", nil)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.PostForm = map[string][]string{
		"enabled": {"invalid"},
	}
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	errorMsg, _ := response["error"].(string)
	if errorMsg != "Invalid enabled value" {
		t.Errorf("Expected 'Invalid enabled value' error, got: %s", errorMsg)
	}
}

// TestToggleClientServer_NonExistentServer tests toggling non-existent server
func TestToggleClientServer_NonExistentServer(t *testing.T) {
	handler, tempDir, cleanup := setupTestAPIHandler(t)
	defer cleanup()

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/api/clients/:client/servers/:server/toggle", handler.ToggleClientServer)

	// Create client config file
	clientPath := filepath.Join(tempDir, "client.json")
	clientData := map[string]interface{}{
		"mcpServers": map[string]interface{}{},
	}
	jsonData, _ := json.Marshal(clientData)
	os.WriteFile(clientPath, jsonData, 0644)

	// Try to toggle non-existent server
	req, _ := http.NewRequest("POST", "/api/clients/test-client/servers/non-existent/toggle", nil)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.PostForm = map[string][]string{
		"enabled": {"true"},
	}
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status 500, got %d", w.Code)
	}

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	if err, ok := response["error"].(string); !ok || err == "" {
		t.Error("Expected error message for non-existent server")
	}
}

// TestToggleClientServer_NonExistentClient tests toggling for non-existent client
func TestToggleClientServer_NonExistentClient(t *testing.T) {
	handler, _, cleanup := setupTestAPIHandler(t)
	defer cleanup()

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/api/clients/:client/servers/:server/toggle", handler.ToggleClientServer)

	// Try to toggle for non-existent client
	req, _ := http.NewRequest("POST", "/api/clients/non-existent/servers/test-server/toggle", nil)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.PostForm = map[string][]string{
		"enabled": {"true"},
	}
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status 500, got %d", w.Code)
	}

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	if err, ok := response["error"].(string); !ok || err == "" {
		t.Error("Expected error message for non-existent client")
	}
}
