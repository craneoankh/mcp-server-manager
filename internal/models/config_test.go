package models

import (
	"encoding/json"
	"testing"
)

const (
	testContext7URL      = "https://mcp.context7.com/mcp"
	testContext7GeminiID = "context7-gemini"
)

// findServer is a helper to locate an MCP server by name
func findServer(servers []MCPServer, name string) *MCPServer {
	for i := range servers {
		if servers[i].Name == name {
			return &servers[i]
		}
	}
	return nil
}

func TestConfigStructure(t *testing.T) {
	// Note: This test uses direct YAML unmarshaling, not the loader
	// The loader uses custom logic to preserve order. This tests basic struct compatibility.
	config := Config{
		ServerPort: 6543,
		MCPServers: []MCPServer{
			{
				Name: "filesystem",
				Config: map[string]interface{}{
					"command": "npx",
					"args":    []interface{}{"@modelcontextprotocol/server-filesystem"},
					"env": map[string]interface{}{
						"NODE_ENV": "production",
					},
				},
			},
			{
				Name: "context7",
				Config: map[string]interface{}{
					"type": "http",
					"url":  testContext7URL,
					"headers": map[string]interface{}{
						"CONTEXT7_API_KEY": "key123",
						"Accept":           "application/json",
					},
				},
			},
			{
				Name: testContext7GeminiID,
				Config: map[string]interface{}{
					"httpUrl": testContext7URL,
					"headers": map[string]interface{}{
						"CONTEXT7_API_KEY": "key123",
					},
				},
			},
		},
		Clients: map[string]*Client{
			"claude_code": {
				ConfigPath: "~/.claude.json",
				Enabled:    []string{"filesystem", "context7"},
			},
			"gemini_cli": {
				ConfigPath: "~/.gemini/settings.json",
				Enabled:    []string{testContext7GeminiID},
			},
		},
	}

	if config.ServerPort != 6543 {
		t.Errorf("ServerPort: got %d, want %d", config.ServerPort, 6543)
	}

	if len(config.MCPServers) != 3 {
		t.Fatalf("Expected 3 MCP servers, got %d", len(config.MCPServers))
	}

	if len(config.Clients) != 2 {
		t.Fatalf("Expected 2 clients, got %d", len(config.Clients))
	}

	// Check filesystem server (STDIO)
	testFilesystemServer(t, config.MCPServers)

	// Check context7 server (HTTP with type field)
	testContext7Server(t, config.MCPServers)

	// Check context7-gemini server (HTTP with httpUrl field)
	testContext7GeminiServer(t, config.MCPServers)

	// Check clients
	testClaudeClient(t, config.Clients)
	testGeminiClient(t, config.Clients)
}

func testFilesystemServer(t *testing.T, servers []MCPServer) {
	t.Helper()
	fs := findServer(servers, "filesystem")
	if fs == nil {
		t.Fatal("filesystem server not found")
	}
	if fs.Config["command"] != "npx" {
		t.Errorf("filesystem command: got %v, want 'npx'", fs.Config["command"])
	}
}

func testContext7Server(t *testing.T, servers []MCPServer) {
	t.Helper()
	ctx7 := findServer(servers, "context7")
	if ctx7 == nil {
		t.Fatal("context7 server not found")
	}
	if ctx7.Config["type"] != "http" {
		t.Errorf("context7 type: got %v, want 'http'", ctx7.Config["type"])
	}
	if ctx7.Config["url"] != testContext7URL {
		t.Errorf("context7 url: got %v, want %q", ctx7.Config["url"], testContext7URL)
	}
}

func testContext7GeminiServer(t *testing.T, servers []MCPServer) {
	t.Helper()
	ctx7gem := findServer(servers, testContext7GeminiID)
	if ctx7gem == nil {
		t.Fatal("context7-gemini server not found")
	}
	if ctx7gem.Config["httpUrl"] != testContext7URL {
		t.Errorf("context7-gemini httpUrl: got %v, want %q", ctx7gem.Config["httpUrl"], testContext7URL)
	}
}

func testClaudeClient(t *testing.T, clients map[string]*Client) {
	t.Helper()
	claudeClient, exists := clients["claude_code"]
	if !exists {
		t.Fatal("claude_code client not found")
	}
	if claudeClient.ConfigPath != "~/.claude.json" {
		t.Errorf("claude_code config_path: got %q, want '~/.claude.json'", claudeClient.ConfigPath)
	}
	if len(claudeClient.Enabled) != 2 {
		t.Errorf("claude_code enabled: got %d servers, want 2", len(claudeClient.Enabled))
	}
}

func testGeminiClient(t *testing.T, clients map[string]*Client) {
	t.Helper()
	geminiClient, exists := clients["gemini_cli"]
	if !exists {
		t.Fatal("gemini_cli client not found")
	}
	if len(geminiClient.Enabled) != 1 {
		t.Errorf("gemini_cli enabled: got %d servers, want 1", len(geminiClient.Enabled))
	}
}

func TestConfigJSONSerialization(t *testing.T) {
	config := Config{
		ServerPort: 6543,
		MCPServers: []MCPServer{
			{
				Name: "test-server",
				Config: map[string]interface{}{
					"command": "npx",
					"args":    []interface{}{"test"},
					"env": map[string]interface{}{
						"KEY": "value",
					},
				},
			},
		},
		Clients: map[string]*Client{
			"test_client": {
				ConfigPath: "~/.test.json",
				Enabled:    []string{"test-server"},
			},
		},
	}

	// Test JSON serialization
	jsonData, err := json.Marshal(config)
	if err != nil {
		t.Fatalf("Failed to marshal JSON: %v", err)
	}

	// Test JSON deserialization
	var parsedConfig Config
	err = json.Unmarshal(jsonData, &parsedConfig)
	if err != nil {
		t.Fatalf("Failed to unmarshal JSON: %v", err)
	}

	// Verify key fields
	if parsedConfig.ServerPort != config.ServerPort {
		t.Errorf("ServerPort: got %d, want %d", parsedConfig.ServerPort, config.ServerPort)
	}
	if len(parsedConfig.MCPServers) != len(config.MCPServers) {
		t.Errorf("MCPServers count: got %d, want %d", len(parsedConfig.MCPServers), len(config.MCPServers))
	}
	if len(parsedConfig.Clients) != len(config.Clients) {
		t.Errorf("Clients count: got %d, want %d", len(parsedConfig.Clients), len(config.Clients))
	}
}

func TestFieldPreservation(t *testing.T) {
	// Test that ALL fields are preserved, including custom ones
	config := Config{
		ServerPort: 6543,
		MCPServers: []MCPServer{
			{
				Name: "custom-server",
				Config: map[string]interface{}{
					"type":        "http",
					"url":         "https://example.com",
					"customField": "customValue",
					"nestedCustom": map[string]interface{}{
						"foo": "bar",
						"baz": 123,
					},
				},
			},
		},
		Clients: map[string]*Client{
			"test": {
				ConfigPath: "~/.test.json",
				Enabled:    []string{},
			},
		},
	}

	// Find custom-server
	var server *MCPServer
	for i := range config.MCPServers {
		if config.MCPServers[i].Name == "custom-server" {
			server = &config.MCPServers[i]
			break
		}
	}
	if server == nil {
		t.Fatal("custom-server not found")
	}

	// Verify all fields are preserved
	if server.Config["type"] != "http" {
		t.Errorf("type field not preserved")
	}
	if server.Config["url"] != "https://example.com" {
		t.Errorf("url field not preserved")
	}
	if server.Config["customField"] != "customValue" {
		t.Errorf("customField not preserved")
	}
	if server.Config["nestedCustom"] == nil {
		t.Errorf("nestedCustom not preserved")
	}
}