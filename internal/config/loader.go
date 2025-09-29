package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"

	"github.com/vlazic/mcp-server-manager/internal/models"
)

const DefaultConfigPath = "configs/config.yaml"

func LoadConfig(configPath string) (*models.Config, string, error) {
	actualPath, err := resolveConfigPath(configPath)
	if err != nil {
		return nil, "", err
	}

	data, err := os.ReadFile(actualPath)
	if err != nil {
		return nil, "", fmt.Errorf("failed to read config file: %w", err)
	}

	// Parse YAML manually to preserve server order
	var rawConfig struct {
		MCPServers map[string]map[string]interface{} `yaml:"mcpServers"`
		Clients    map[string]*models.Client         `yaml:"clients"`
		ServerPort int                               `yaml:"server_port"`
	}

	// Use yaml.v3 Node to preserve order
	var node yaml.Node
	if err := yaml.Unmarshal(data, &node); err != nil {
		return nil, "", fmt.Errorf("failed to parse config file: %w", err)
	}

	if err := node.Decode(&rawConfig); err != nil {
		return nil, "", fmt.Errorf("failed to decode config: %w", err)
	}

	// Extract server order from YAML node
	var serverOrder []string
	if len(node.Content) > 0 && len(node.Content[0].Content) > 0 {
		for i := 0; i < len(node.Content[0].Content); i += 2 {
			keyNode := node.Content[0].Content[i]
			if keyNode.Value == "mcpServers" && i+1 < len(node.Content[0].Content) {
				serversNode := node.Content[0].Content[i+1]
				// Extract keys in order
				for j := 0; j < len(serversNode.Content); j += 2 {
					serverName := serversNode.Content[j].Value
					serverOrder = append(serverOrder, serverName)
				}
				break
			}
		}
	}

	// Convert map to ordered slice
	config := &models.Config{
		MCPServers: make([]models.MCPServer, 0, len(rawConfig.MCPServers)),
		Clients:    rawConfig.Clients,
		ServerPort: rawConfig.ServerPort,
	}

	// Use extracted order, or fallback to map iteration
	if len(serverOrder) > 0 {
		for _, name := range serverOrder {
			if serverConfig, exists := rawConfig.MCPServers[name]; exists {
				config.MCPServers = append(config.MCPServers, models.MCPServer{
					Name:   name,
					Config: serverConfig,
				})
			}
		}
	} else {
		// Fallback: map iteration (order not guaranteed)
		for name, serverConfig := range rawConfig.MCPServers {
			config.MCPServers = append(config.MCPServers, models.MCPServer{
				Name:   name,
				Config: serverConfig,
			})
		}
	}

	if config.ServerPort == 0 {
		config.ServerPort = 6543
	}

	return config, actualPath, nil
}

// resolveConfigPath implements smart config path resolution with fallback
func resolveConfigPath(configPath string) (string, error) {
	// If explicit path provided, try to use it - create if it doesn't exist
	if configPath != "" {
		expanded := ExpandPath(configPath)
		if _, err := os.Stat(expanded); err != nil {
			// If explicit path doesn't exist, try to create it
			if err := createDefaultConfig(expanded); err != nil {
				return "", fmt.Errorf("specified config file not found and could not create: %s", expanded)
			}
			fmt.Printf("Created config file at: %s\n", expanded)
		}
		return expanded, nil
	}

	// Priority order for auto-resolution:
	// 1. ~/.config/mcp-server-manager/config.yaml (user config)
	// 2. ./config.yaml (current directory)
	// 3. configs/config.yaml (relative to binary)
	// 4. Auto-create user config if none found

	candidates := []string{
		ExpandPath("~/.config/mcp-server-manager/config.yaml"),
		"./config.yaml",
		DefaultConfigPath,
	}

	for _, path := range candidates {
		if _, err := os.Stat(path); err == nil {
			return path, nil
		}
	}

	// No config found, auto-create user config
	userConfigPath := ExpandPath("~/.config/mcp-server-manager/config.yaml")
	if err := createDefaultConfig(userConfigPath); err != nil {
		return "", fmt.Errorf("failed to create default config: %w", err)
	}

	fmt.Printf("Created default config file at: %s\n", userConfigPath)
	fmt.Println("Please edit this file to configure your MCP servers and clients.")

	return userConfigPath, nil
}

// createDefaultConfig creates a default config file with example configuration
func createDefaultConfig(configPath string) error {
	if err := os.MkdirAll(filepath.Dir(configPath), 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	defaultConfig := `# MCP Server Manager Configuration v2.0
# This matches standard MCP client config format for maximum compatibility
# Edit this file to configure your MCP servers and clients

server_port: 6543

# MCP Servers - Standard format matching MCP clients
# Server names are keys; configurations are values (pass through to clients)
mcpServers:
  # STDIO Transport Example (command-based)
  filesystem:
    command: "npx"
    args: ["@modelcontextprotocol/server-filesystem", "/path/to/your/directory"]
    env:
      NODE_ENV: "production"
    timeout: 30000  # Optional: request timeout in ms
    trust: false    # Optional: bypass tool confirmations

  # HTTP Transport Example (with type field for VS Code compatibility)
  context7-vscode:
    type: "http"
    url: "https://mcp.context7.com/mcp"
    headers:
      CONTEXT7_API_KEY: "ADD_YOUR_API_KEY"
      Accept: "application/json, text/event-stream"
    timeout: 10000

  # HTTP Transport Example (httpUrl variant for Gemini CLI)
  context7-gemini:
    httpUrl: "https://mcp.context7.com/mcp"
    headers:
      CONTEXT7_API_KEY: "ADD_YOUR_API_KEY"
      Accept: "application/json, text/event-stream"

  # SSE Transport Example (uncomment to use)
  # sse_server:
  #   url: "http://localhost:8080/sse"
  #   headers:
  #     Authorization: "Bearer YOUR_TOKEN"
  #   timeout: 15000

  # Advanced STDIO Example with tool filtering
  # git_server:
  #   command: "npx"
  #   args: ["@modelcontextprotocol/server-git", "--repository", "/path/to/repo"]
  #   cwd: "/path/to/working/directory"
  #   env:
  #     GIT_AUTHOR_NAME: "MCP User"
  #     GIT_AUTHOR_EMAIL: "user@example.com"
  #   timeout: 45000
  #   trust: false
  #   includeTools: ["git_log", "git_diff", "git_show"]  # Only allow these tools
  #   excludeTools: ["git_push", "git_reset"]            # Block dangerous tools

# MCP Clients - Configure which servers each client uses
clients:
  claude_code:
    config_path: "~/.claude.json"
    enabled:
      - filesystem
      # - context7-vscode

  gemini_cli:
    config_path: "~/.gemini/settings.json"
    enabled:
      # - context7-gemini
      # - filesystem

# Notes:
# - ALL fields in mcpServers are passed through to client configs (no filtering)
# - Supports any MCP spec fields: type, url, httpUrl, command, args, env, headers, etc.
# - Use 'enabled' array per client to control which servers each client uses
# - Transport Types:
#   * STDIO: command + args (local processes)
#   * HTTP: url/httpUrl + headers (remote HTTP endpoints)
#   * SSE: url + headers (Server-Sent Events)
# - Restart service after changes: systemctl --user restart mcp-server-manager
`

	if err := os.WriteFile(configPath, []byte(defaultConfig), 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

func SaveConfig(config *models.Config, configPath string) error {
	if configPath == "" {
		configPath = DefaultConfigPath
	}

	if err := os.MkdirAll(filepath.Dir(configPath), 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Convert ordered slice back to map for standard YAML marshaling
	// This preserves order through yaml.v3's MapSlice or custom marshaling
	serversMap := make(map[string]interface{})
	for _, server := range config.MCPServers {
		serversMap[server.Name] = server.Config
	}

	// Create temporary struct for marshaling with proper order
	type ConfigForSave struct {
		ServerPort int                    `yaml:"server_port"`
		MCPServers map[string]interface{} `yaml:"mcpServers"`
		Clients    map[string]*models.Client `yaml:"clients"`
	}

	saveConfig := ConfigForSave{
		ServerPort: config.ServerPort,
		MCPServers: serversMap,
		Clients:    config.Clients,
	}

	data, err := yaml.Marshal(saveConfig)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

func ExpandPath(path string) string {
	if len(path) > 0 && path[0] == '~' {
		home, err := os.UserHomeDir()
		if err != nil {
			return path
		}
		return filepath.Join(home, path[1:])
	}
	return path
}
