package services

import (
	"fmt"

	"github.com/vlazic/mcp-server-manager/internal/config"
	"github.com/vlazic/mcp-server-manager/internal/models"
)

type MCPManagerService struct {
	config              *models.Config
	clientConfigService *ClientConfigService
	validator           *ValidatorService
	configPath          string
}

func NewMCPManagerService(cfg *models.Config, configPath string) *MCPManagerService {
	return &MCPManagerService{
		config:              cfg,
		clientConfigService: NewClientConfigService(cfg),
		validator:           NewValidatorService(),
		configPath:          configPath,
	}
}

// GetMCPServers returns the server map
func (s *MCPManagerService) GetMCPServers() map[string]map[string]interface{} {
	return s.config.MCPServers
}

// GetClients returns the client map
func (s *MCPManagerService) GetClients() map[string]*models.Client {
	return s.config.Clients
}

// ToggleClientMCPServer enables or disables a server for a specific client
func (s *MCPManagerService) ToggleClientMCPServer(clientName, serverName string, enabled bool) error {
	// Check if client exists
	client, exists := s.config.Clients[clientName]
	if !exists {
		return fmt.Errorf("client '%s' not found", clientName)
	}

	// Check if server exists
	if _, exists := s.config.MCPServers[serverName]; !exists {
		return fmt.Errorf("MCP server '%s' not found", serverName)
	}

	// Initialize enabled list if nil
	if client.Enabled == nil {
		client.Enabled = []string{}
	}

	// Update enabled list
	if enabled {
		// Add server to enabled list if not already present
		found := false
		for _, name := range client.Enabled {
			if name == serverName {
				found = true
				break
			}
		}
		if !found {
			client.Enabled = append(client.Enabled, serverName)
		}
	} else {
		// Remove server from enabled list
		newEnabled := []string{}
		for _, name := range client.Enabled {
			if name != serverName {
				newEnabled = append(newEnabled, name)
			}
		}
		client.Enabled = newEnabled
	}

	// Save config
	if err := s.saveConfig(); err != nil {
		return err
	}

	// Update client config file
	return s.clientConfigService.UpdateMCPServerStatus(clientName, serverName, enabled)
}

// GetServerStatus returns server configuration by name
func (s *MCPManagerService) GetServerStatus(serverName string) (map[string]interface{}, error) {
	serverConfig, exists := s.config.MCPServers[serverName]
	if !exists {
		return nil, fmt.Errorf("MCP server '%s' not found", serverName)
	}
	return serverConfig, nil
}

// SyncAllClients synchronizes all client configurations based on enabled lists
func (s *MCPManagerService) SyncAllClients() error {
	for clientName, client := range s.config.Clients {
		// Build set of enabled servers for quick lookup
		enabledSet := make(map[string]bool)
		for _, serverName := range client.Enabled {
			enabledSet[serverName] = true
		}

		// Sync each server in the config
		for serverName := range s.config.MCPServers {
			enabled := enabledSet[serverName]
			if err := s.clientConfigService.UpdateMCPServerStatus(clientName, serverName, enabled); err != nil {
				return fmt.Errorf("failed to sync client '%s': %w", clientName, err)
			}
		}
	}
	return nil
}

func (s *MCPManagerService) GetConfig() *models.Config {
	return s.config
}

func (s *MCPManagerService) ValidateConfig() error {
	return s.validator.ValidateConfig(s.config)
}

// AddServer adds a new MCP server to the configuration
func (s *MCPManagerService) AddServer(serverName string, serverConfig map[string]interface{}) error {
	// Validate the server config
	if err := s.validator.ValidateMCPServerConfig(serverName, serverConfig); err != nil {
		return fmt.Errorf("server validation failed: %w", err)
	}

	// Check if server with this name already exists
	if _, exists := s.config.MCPServers[serverName]; exists {
		return fmt.Errorf("server with name '%s' already exists", serverName)
	}

	// Initialize servers map if nil
	if s.config.MCPServers == nil {
		s.config.MCPServers = make(map[string]map[string]interface{})
	}

	// Add the server to the config
	s.config.MCPServers[serverName] = serverConfig

	// Save the config
	return s.saveConfig()
}

func (s *MCPManagerService) saveConfig() error {
	if err := s.ValidateConfig(); err != nil {
		return fmt.Errorf("config validation failed: %w", err)
	}
	return config.SaveConfig(s.config, s.configPath)
}