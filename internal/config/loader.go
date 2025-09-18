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

	var config models.Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, "", fmt.Errorf("failed to parse config file: %w", err)
	}

	if config.ServerPort == 0 {
		config.ServerPort = 6543
	}

	return &config, actualPath, nil
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

	defaultConfig := `# MCP Server Manager Configuration
# Edit this file to configure your MCP servers and clients

server_port: 6543

mcp_servers:
  - name: "filesystem"
    command: "npx"
    args: ["@modelcontextprotocol/server-filesystem", "/path/to/your/directory"]
    env: {}
    enabled_globally: false
    clients:
      claude_code: false
      gemini_cli: false

clients:
  - name: "claude_code"
    config_path: "~/.claude/settings.json"
  - name: "gemini_cli"
    config_path: "~/.gemini/settings.json"

# Add more MCP servers and clients as needed
# Restart the service after making changes: systemctl --user restart mcp-server-manager
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

	data, err := yaml.Marshal(config)
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
