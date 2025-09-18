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

func (s *ClientConfigService) ReadClientConfig(clientName string) (*models.ClientConfig, error) {
	client := s.findClient(clientName)
	if client == nil {
		return nil, fmt.Errorf("client '%s' not found", clientName)
	}

	configPath := config.ExpandPath(client.ConfigPath)
	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			// Create empty config if file doesn't exist
			return &models.ClientConfig{
				MCPServers: make(map[string]interface{}),
			}, nil
		}
		return nil, fmt.Errorf("failed to read client config '%s': %w", configPath, err)
	}

	var clientConfig models.ClientConfig
	if err := json.Unmarshal(data, &clientConfig); err != nil {
		return nil, fmt.Errorf("failed to parse client config '%s': %w", configPath, err)
	}

	// Initialize MCPServers map if nil
	if clientConfig.MCPServers == nil {
		clientConfig.MCPServers = make(map[string]interface{})
	}

	return &clientConfig, nil
}

func (s *ClientConfigService) WriteClientConfig(clientName string, clientConfig *models.ClientConfig) error {
	client := s.findClient(clientName)
	if client == nil {
		return fmt.Errorf("client '%s' not found", clientName)
	}

	if err := s.validator.ValidateClientConfig(clientConfig); err != nil {
		return fmt.Errorf("client config validation failed: %w", err)
	}

	configPath := config.ExpandPath(client.ConfigPath)

	if err := s.backupConfig(configPath); err != nil {
		return fmt.Errorf("failed to backup config: %w", err)
	}

	if err := os.MkdirAll(filepath.Dir(configPath), 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	data, err := json.MarshalIndent(clientConfig, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal client config: %w", err)
	}

	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write client config: %w", err)
	}

	return nil
}

func (s *ClientConfigService) UpdateMCPServerStatus(clientName, serverName string, enabled bool) error {
	clientConfig, err := s.ReadClientConfig(clientName)
	if err != nil {
		return err
	}

	if clientConfig.MCPServers == nil {
		clientConfig.MCPServers = make(map[string]interface{})
	}

	server := s.findMCPServer(serverName)
	if server == nil {
		return fmt.Errorf("MCP server '%s' not found", serverName)
	}

	if enabled {
		// Create server config based on the original server definition
		serverConfig := map[string]interface{}{}

		// Add transport type based on server configuration
		if server.Command != "" {
			// STDIO transport
			serverConfig["command"] = server.Command
			if len(server.Args) > 0 {
				serverConfig["args"] = server.Args
			}
		} else if server.HttpURL != "" {
			// HTTP transport
			serverConfig["httpUrl"] = server.HttpURL
		} else if server.URL != "" {
			// SSE transport
			serverConfig["url"] = server.URL
		}

		// Add common fields
		if len(server.Env) > 0 {
			serverConfig["env"] = server.Env
		}

		if len(server.Headers) > 0 {
			serverConfig["headers"] = server.Headers
		}

		clientConfig.MCPServers[serverName] = serverConfig
	} else {
		delete(clientConfig.MCPServers, serverName)
	}

	return s.WriteClientConfig(clientName, clientConfig)
}

func (s *ClientConfigService) GetMCPServerStatus(clientName, serverName string) (bool, error) {
	clientConfig, err := s.ReadClientConfig(clientName)
	if err != nil {
		return false, err
	}

	_, exists := clientConfig.MCPServers[serverName]
	return exists, nil
}

func (s *ClientConfigService) findClient(name string) *models.Client {
	for _, client := range s.config.Clients {
		if client.Name == name {
			return &client
		}
	}
	return nil
}

func (s *ClientConfigService) findMCPServer(name string) *models.MCPServer {
	for _, server := range s.config.MCPServers {
		if server.Name == name {
			return &server
		}
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