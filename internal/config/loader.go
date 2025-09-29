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

	// Parse YAML with order preservation
	rawConfig, serverOrder, err := parseYAMLConfig(data)
	if err != nil {
		return nil, "", err
	}

	// Build final config structure
	config := &models.Config{
		MCPServers: buildOrderedServers(serverOrder, rawConfig.MCPServers),
		Clients:    rawConfig.Clients,
		ServerPort: rawConfig.ServerPort,
	}

	if config.ServerPort == 0 {
		config.ServerPort = 6543
	}

	return config, actualPath, nil
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
