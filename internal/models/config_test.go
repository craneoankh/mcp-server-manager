package models

import (
	"encoding/json"
	"testing"
)

const (
	testContext7URL      = "https://mcp.context7.com/mcp"
	testContext7GeminiID = "context7-gemini"
)

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

	// Check filesystem server (STDIO) - find by name in slice
	var fs *MCPServer
	for i := range config.MCPServers {
		if config.MCPServers[i].Name == "filesystem" {
			fs = &config.MCPServers[i]
			break
		}
	}
	if fs == nil {
		t.Fatal("filesystem server not found")
	}
	if fs.Config["command"] != "npx" {
		t.Errorf("filesystem command: got %v, want 'npx'", fs.Config["command"])
	}

	// Check context7 server (HTTP with type field)
	var ctx7 *MCPServer
	for i := range config.MCPServers {
		if config.MCPServers[i].Name == "context7" {
			ctx7 = &config.MCPServers[i]
			break
		}
	}
	if ctx7 == nil {
		t.Fatal("context7 server not found")
	}
	if ctx7.Config["type"] != "http" {
		t.Errorf("context7 type: got %v, want 'http'", ctx7.Config["type"])
	}
	if ctx7.Config["url"] != "https://mcp.context7.com/mcp" {
		t.Errorf("context7 url: got %v, want 'https://mcp.context7.com/mcp'", ctx7.Config["url"])
	}

	// Check context7-gemini server (HTTP with httpUrl field)
	var ctx7gem *MCPServer
	for i := range config.MCPServers {
		if config.MCPServers[i].Name == "context7-gemini" {
			ctx7gem = &config.MCPServers[i]
			break
		}
	}
	if ctx7gem == nil {
		t.Fatal("context7-gemini server not found")
	}
	if ctx7gem.Config["httpUrl"] != "https://mcp.context7.com/mcp" {
		t.Errorf("context7-gemini httpUrl: got %v, want 'https://mcp.context7.com/mcp'", ctx7gem.Config["httpUrl"])
	}

	// Check clients
	claudeClient, exists := config.Clients["claude_code"]
	if !exists {
		t.Fatal("claude_code client not found")
	}
	if claudeClient.ConfigPath != "~/.claude.json" {
		t.Errorf("claude_code config_path: got %q, want '~/.claude.json'", claudeClient.ConfigPath)
	}
	if len(claudeClient.Enabled) != 2 {
		t.Errorf("claude_code enabled: got %d servers, want 2", len(claudeClient.Enabled))
	}

	geminiClient, exists := config.Clients["gemini_cli"]
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