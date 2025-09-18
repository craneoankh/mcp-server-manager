package models

import (
	"encoding/json"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestMCPServerParsing(t *testing.T) {
	tests := []struct {
		name     string
		yaml     string
		expected MCPServer
	}{
		{
			name: "STDIO Transport",
			yaml: `
name: "filesystem"
command: "npx"
args: ["@modelcontextprotocol/server-filesystem", "/path"]
env:
  NODE_ENV: "production"
timeout: 30000
trust: true
include_tools: ["read_file", "write_file"]
exclude_tools: ["delete_file"]
enabled_globally: true
clients:
  claude_code: true
`,
			expected: MCPServer{
				Name:            "filesystem",
				Command:         "npx",
				Args:            []string{"@modelcontextprotocol/server-filesystem", "/path"},
				Env:             map[string]string{"NODE_ENV": "production"},
				Timeout:         30000,
				Trust:           true,
				IncludeTools:    []string{"read_file", "write_file"},
				ExcludeTools:    []string{"delete_file"},
				EnabledGlobally: true,
				Clients:         map[string]bool{"claude_code": true},
			},
		},
		{
			name: "HTTP Transport",
			yaml: `
name: "context7"
http_url: "https://mcp.context7.com/mcp"
headers:
  Authorization: "Bearer token123"
  Content-Type: "application/json"
timeout: 5000
enabled_globally: false
clients:
  claude_code: false
  gemini_cli: true
`,
			expected: MCPServer{
				Name:    "context7",
				HttpURL: "https://mcp.context7.com/mcp",
				Headers: map[string]string{
					"Authorization": "Bearer token123",
					"Content-Type":  "application/json",
				},
				Timeout:         5000,
				EnabledGlobally: false,
				Clients:         map[string]bool{"claude_code": false, "gemini_cli": true},
			},
		},
		{
			name: "SSE Transport",
			yaml: `
name: "sse_server"
url: "http://localhost:8080/sse"
headers:
  X-API-Key: "secret123"
timeout: 10000
enabled_globally: true
clients:
  claude_code: true
`,
			expected: MCPServer{
				Name:            "sse_server",
				URL:             "http://localhost:8080/sse",
				Headers:         map[string]string{"X-API-Key": "secret123"},
				Timeout:         10000,
				EnabledGlobally: true,
				Clients:         map[string]bool{"claude_code": true},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var server MCPServer
			err := yaml.Unmarshal([]byte(tt.yaml), &server)
			if err != nil {
				t.Fatalf("Failed to parse YAML: %v", err)
			}

			// Compare fields
			if server.Name != tt.expected.Name {
				t.Errorf("Name: got %q, want %q", server.Name, tt.expected.Name)
			}
			if server.Command != tt.expected.Command {
				t.Errorf("Command: got %q, want %q", server.Command, tt.expected.Command)
			}
			if server.URL != tt.expected.URL {
				t.Errorf("URL: got %q, want %q", server.URL, tt.expected.URL)
			}
			if server.HttpURL != tt.expected.HttpURL {
				t.Errorf("HttpURL: got %q, want %q", server.HttpURL, tt.expected.HttpURL)
			}
			if server.Timeout != tt.expected.Timeout {
				t.Errorf("Timeout: got %d, want %d", server.Timeout, tt.expected.Timeout)
			}
			if server.Trust != tt.expected.Trust {
				t.Errorf("Trust: got %t, want %t", server.Trust, tt.expected.Trust)
			}
			if server.EnabledGlobally != tt.expected.EnabledGlobally {
				t.Errorf("EnabledGlobally: got %t, want %t", server.EnabledGlobally, tt.expected.EnabledGlobally)
			}

			// Compare slices and maps
			if !equalStringSlices(server.Args, tt.expected.Args) {
				t.Errorf("Args: got %v, want %v", server.Args, tt.expected.Args)
			}
			if !equalStringSlices(server.IncludeTools, tt.expected.IncludeTools) {
				t.Errorf("IncludeTools: got %v, want %v", server.IncludeTools, tt.expected.IncludeTools)
			}
			if !equalStringSlices(server.ExcludeTools, tt.expected.ExcludeTools) {
				t.Errorf("ExcludeTools: got %v, want %v", server.ExcludeTools, tt.expected.ExcludeTools)
			}
			if !equalStringMaps(server.Env, tt.expected.Env) {
				t.Errorf("Env: got %v, want %v", server.Env, tt.expected.Env)
			}
			if !equalStringMaps(server.Headers, tt.expected.Headers) {
				t.Errorf("Headers: got %v, want %v", server.Headers, tt.expected.Headers)
			}
			if !equalBoolMaps(server.Clients, tt.expected.Clients) {
				t.Errorf("Clients: got %v, want %v", server.Clients, tt.expected.Clients)
			}
		})
	}
}

func TestMCPServerJSONSerialization(t *testing.T) {
	server := MCPServer{
		Name:            "test_server",
		Command:         "python",
		Args:            []string{"-m", "server"},
		HttpURL:         "https://example.com/mcp", // This should be ignored due to having Command
		Env:             map[string]string{"API_KEY": "secret"},
		Headers:         map[string]string{"Auth": "Bearer token"},
		Timeout:         5000,
		Trust:           true,
		IncludeTools:    []string{"tool1", "tool2"},
		ExcludeTools:    []string{"danger"},
		EnabledGlobally: true,
		Clients:         map[string]bool{"claude": true, "gemini": false},
	}

	// Test JSON serialization
	jsonData, err := json.Marshal(server)
	if err != nil {
		t.Fatalf("Failed to marshal JSON: %v", err)
	}

	// Test JSON deserialization
	var parsedServer MCPServer
	err = json.Unmarshal(jsonData, &parsedServer)
	if err != nil {
		t.Fatalf("Failed to unmarshal JSON: %v", err)
	}

	// Verify key fields
	if parsedServer.Name != server.Name {
		t.Errorf("Name: got %q, want %q", parsedServer.Name, server.Name)
	}
	if parsedServer.Command != server.Command {
		t.Errorf("Command: got %q, want %q", parsedServer.Command, server.Command)
	}
	if parsedServer.HttpURL != server.HttpURL {
		t.Errorf("HttpURL: got %q, want %q", parsedServer.HttpURL, server.HttpURL)
	}
}

func TestConfigStructure(t *testing.T) {
	configYAML := `
server_port: 6543
mcp_servers:
  - name: "filesystem"
    command: "npx"
    args: ["@modelcontextprotocol/server-filesystem"]
    enabled_globally: true
    clients:
      claude_code: true
  - name: "context7"
    http_url: "https://mcp.context7.com/mcp"
    headers:
      CONTEXT7_API_KEY: "key123"
    enabled_globally: false
    clients:
      gemini_cli: true
clients:
  - name: "claude_code"
    config_path: "~/.claude/settings.json"
  - name: "gemini_cli"
    config_path: "~/.gemini/settings.json"
`

	var config Config
	err := yaml.Unmarshal([]byte(configYAML), &config)
	if err != nil {
		t.Fatalf("Failed to parse config YAML: %v", err)
	}

	if config.ServerPort != 6543 {
		t.Errorf("ServerPort: got %d, want %d", config.ServerPort, 6543)
	}

	if len(config.MCPServers) != 2 {
		t.Fatalf("Expected 2 MCP servers, got %d", len(config.MCPServers))
	}

	if len(config.Clients) != 2 {
		t.Fatalf("Expected 2 clients, got %d", len(config.Clients))
	}

	// Check first server (STDIO)
	fs := config.MCPServers[0]
	if fs.Name != "filesystem" || fs.Command != "npx" {
		t.Errorf("First server: got name=%q command=%q, want name=%q command=%q",
			fs.Name, fs.Command, "filesystem", "npx")
	}

	// Check second server (HTTP)
	ctx7 := config.MCPServers[1]
	if ctx7.Name != "context7" || ctx7.HttpURL != "https://mcp.context7.com/mcp" {
		t.Errorf("Second server: got name=%q http_url=%q, want name=%q http_url=%q",
			ctx7.Name, ctx7.HttpURL, "context7", "https://mcp.context7.com/mcp")
	}
}

// Helper functions
func equalStringSlices(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i, v := range a {
		if v != b[i] {
			return false
		}
	}
	return true
}

func equalStringMaps(a, b map[string]string) bool {
	if len(a) != len(b) {
		return false
	}
	for k, v := range a {
		if b[k] != v {
			return false
		}
	}
	return true
}

func equalBoolMaps(a, b map[string]bool) bool {
	if len(a) != len(b) {
		return false
	}
	for k, v := range a {
		if b[k] != v {
			return false
		}
	}
	return true
}