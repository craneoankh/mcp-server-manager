package models

import (
	"encoding/json"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestConfigStructure(t *testing.T) {
	configYAML := `
server_port: 6543
mcpServers:
  filesystem:
    command: "npx"
    args: ["@modelcontextprotocol/server-filesystem"]
    env:
      NODE_ENV: "production"
  context7:
    type: "http"
    url: "https://mcp.context7.com/mcp"
    headers:
      CONTEXT7_API_KEY: "key123"
      Accept: "application/json"
  context7-gemini:
    httpUrl: "https://mcp.context7.com/mcp"
    headers:
      CONTEXT7_API_KEY: "key123"
clients:
  claude_code:
    config_path: "~/.claude.json"
    enabled:
      - filesystem
      - context7
  gemini_cli:
    config_path: "~/.gemini/settings.json"
    enabled:
      - context7-gemini
`

	var config Config
	err := yaml.Unmarshal([]byte(configYAML), &config)
	if err != nil {
		t.Fatalf("Failed to parse config YAML: %v", err)
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
	fs, exists := config.MCPServers["filesystem"]
	if !exists {
		t.Fatal("filesystem server not found")
	}
	if fs["command"] != "npx" {
		t.Errorf("filesystem command: got %v, want 'npx'", fs["command"])
	}

	// Check context7 server (HTTP with type field)
	ctx7, exists := config.MCPServers["context7"]
	if !exists {
		t.Fatal("context7 server not found")
	}
	if ctx7["type"] != "http" {
		t.Errorf("context7 type: got %v, want 'http'", ctx7["type"])
	}
	if ctx7["url"] != "https://mcp.context7.com/mcp" {
		t.Errorf("context7 url: got %v, want 'https://mcp.context7.com/mcp'", ctx7["url"])
	}

	// Check context7-gemini server (HTTP with httpUrl field)
	ctx7gem, exists := config.MCPServers["context7-gemini"]
	if !exists {
		t.Fatal("context7-gemini server not found")
	}
	if ctx7gem["httpUrl"] != "https://mcp.context7.com/mcp" {
		t.Errorf("context7-gemini httpUrl: got %v, want 'https://mcp.context7.com/mcp'", ctx7gem["httpUrl"])
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
		MCPServers: map[string]map[string]interface{}{
			"test-server": {
				"command": "npx",
				"args":    []interface{}{"test"},
				"env": map[string]interface{}{
					"KEY": "value",
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
	configYAML := `
mcpServers:
  custom-server:
    type: "http"
    url: "https://example.com"
    customField: "customValue"
    nestedCustom:
      foo: "bar"
      baz: 123
clients:
  test:
    config_path: "~/.test.json"
    enabled: []
server_port: 6543
`

	var config Config
	err := yaml.Unmarshal([]byte(configYAML), &config)
	if err != nil {
		t.Fatalf("Failed to parse YAML: %v", err)
	}

	server, exists := config.MCPServers["custom-server"]
	if !exists {
		t.Fatal("custom-server not found")
	}

	// Verify all fields are preserved
	if server["type"] != "http" {
		t.Errorf("type field not preserved")
	}
	if server["url"] != "https://example.com" {
		t.Errorf("url field not preserved")
	}
	if server["customField"] != "customValue" {
		t.Errorf("customField not preserved")
	}
	if server["nestedCustom"] == nil {
		t.Errorf("nestedCustom not preserved")
	}
}