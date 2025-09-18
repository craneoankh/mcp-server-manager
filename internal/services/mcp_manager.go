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

func (s *MCPManagerService) GetMCPServers() []models.MCPServer {
	return s.config.MCPServers
}

func (s *MCPManagerService) GetClients() []models.Client {
	return s.config.Clients
}

func (s *MCPManagerService) ToggleGlobalMCPServer(serverName string, enabled bool) error {
	for i, server := range s.config.MCPServers {
		if server.Name == serverName {
			s.config.MCPServers[i].EnabledGlobally = enabled
			return s.saveConfig()
		}
	}
	return fmt.Errorf("MCP server '%s' not found", serverName)
}

func (s *MCPManagerService) ToggleClientMCPServer(clientName, serverName string, enabled bool) error {
	for i, server := range s.config.MCPServers {
		if server.Name == serverName {
			if s.config.MCPServers[i].Clients == nil {
				s.config.MCPServers[i].Clients = make(map[string]bool)
			}
			s.config.MCPServers[i].Clients[clientName] = enabled

			if err := s.saveConfig(); err != nil {
				return err
			}

			return s.clientConfigService.UpdateMCPServerStatus(clientName, serverName, enabled)
		}
	}
	return fmt.Errorf("MCP server '%s' not found", serverName)
}

func (s *MCPManagerService) GetServerStatus(serverName string) (*models.MCPServer, error) {
	for _, server := range s.config.MCPServers {
		if server.Name == serverName {
			return &server, nil
		}
	}
	return nil, fmt.Errorf("MCP server '%s' not found", serverName)
}

func (s *MCPManagerService) SyncAllClients() error {
	for _, client := range s.config.Clients {
		for _, server := range s.config.MCPServers {
			enabled := server.EnabledGlobally
			if clientEnabled, exists := server.Clients[client.Name]; exists {
				enabled = clientEnabled
			}

			if err := s.clientConfigService.UpdateMCPServerStatus(client.Name, server.Name, enabled); err != nil {
				return fmt.Errorf("failed to sync client '%s': %w", client.Name, err)
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

func (s *MCPManagerService) saveConfig() error {
	if err := s.ValidateConfig(); err != nil {
		return fmt.Errorf("config validation failed: %w", err)
	}
	return config.SaveConfig(s.config, s.configPath)
}