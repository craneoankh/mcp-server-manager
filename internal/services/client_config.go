package services

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/vlazic/mcp-server-manager/internal/config"
	"github.com/vlazic/mcp-server-manager/internal/models"
)

type ClientConfigService struct {
	config    *models.Config
	validator *ValidatorService
}

func NewClientConfigService(cfg *models.Config) *ClientConfigService {
	return &ClientConfigService{
		config:    cfg,
		validator: NewValidatorService(),
	}
}

func (s *ClientConfigService) ReadClientConfig(clientName string) (map[string]interface{}, error) {
	client := s.findClient(clientName)
	if client == nil {
		return nil, fmt.Errorf("client '%s' not found", clientName)
	}

	configPath := config.ExpandPath(client.ConfigPath)
	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			// Create empty config if file doesn't exist
			return map[string]interface{}{
				"mcpServers": make(map[string]interface{}),
			}, nil
		}
		return nil, fmt.Errorf("failed to read client config '%s': %w", configPath, err)
	}

	var rawConfig map[string]interface{}
	if err := json.Unmarshal(data, &rawConfig); err != nil {
		return nil, fmt.Errorf("failed to parse client config '%s': %w", configPath, err)
	}

	// Initialize mcpServers if it doesn't exist
	if rawConfig["mcpServers"] == nil {
		rawConfig["mcpServers"] = make(map[string]interface{})
	}

	return rawConfig, nil
}

func (s *ClientConfigService) WriteClientConfig(clientName string, rawConfig map[string]interface{}) error {
	client := s.findClient(clientName)
	if client == nil {
		return fmt.Errorf("client '%s' not found", clientName)
	}

	configPath := config.ExpandPath(client.ConfigPath)

	if err := s.backupConfig(configPath); err != nil {
		return fmt.Errorf("failed to backup config: %w", err)
	}

	if err := os.MkdirAll(filepath.Dir(configPath), 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	data, err := json.MarshalIndent(rawConfig, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal client config: %w", err)
	}

	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write client config '%s': %w", configPath, err)
	}

	return nil
}

func (s *ClientConfigService) UpdateMCPServerStatus(clientName, serverName string, enabled bool) error {
	rawConfig, err := s.ReadClientConfig(clientName)
	if err != nil {
		return err
	}

	// Get or create mcpServers section
	mcpServers, ok := rawConfig["mcpServers"].(map[string]interface{})
	if !ok {
		mcpServers = make(map[string]interface{})
		rawConfig["mcpServers"] = mcpServers
	}

	if enabled {
		// Get server config from app config
		serverConfig, exists := s.config.MCPServers[serverName]
		if !exists {
			return fmt.Errorf("MCP server '%s' not found in app config", serverName)
		}

		// CRITICAL FIX: Copy the ENTIRE server config map without filtering
		// This preserves ALL fields: type, url, httpUrl, command, args, env, headers, etc.
		// Deep copy to avoid mutations
		copiedConfig := make(map[string]interface{})
		for key, value := range serverConfig {
			copiedConfig[key] = value
		}

		mcpServers[serverName] = copiedConfig
	} else {
		// Remove server from client config
		delete(mcpServers, serverName)
	}

	return s.WriteClientConfig(clientName, rawConfig)
}

func (s *ClientConfigService) GetMCPServerStatus(clientName, serverName string) (bool, error) {
	rawConfig, err := s.ReadClientConfig(clientName)
	if err != nil {
		return false, err
	}

	mcpServers, ok := rawConfig["mcpServers"].(map[string]interface{})
	if !ok {
		return false, nil
	}

	_, exists := mcpServers[serverName]
	return exists, nil
}

func (s *ClientConfigService) findClient(name string) *models.Client {
	if client, exists := s.config.Clients[name]; exists {
		return client
	}
	return nil
}

func (s *ClientConfigService) backupConfig(configPath string) error {
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return nil
	}

	backupPath := configPath + ".backup." + time.Now().Format("20060102-150405")

	data, err := os.ReadFile(configPath)
	if err != nil {
		return err
	}

	return os.WriteFile(backupPath, data, 0644)
}