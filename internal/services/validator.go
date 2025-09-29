package services

import (
	"fmt"

	"github.com/vlazic/mcp-server-manager/internal/models"
)

type ValidatorService struct{}

func NewValidatorService() *ValidatorService {
	return &ValidatorService{}
}

// ValidateConfig validates the entire configuration
func (v *ValidatorService) ValidateConfig(config *models.Config) error {
	if config.ServerPort < 1 || config.ServerPort > 65535 {
		return fmt.Errorf("invalid server port: %d", config.ServerPort)
	}

	if len(config.MCPServers) == 0 {
		return fmt.Errorf("no MCP servers configured")
	}

	if len(config.Clients) == 0 {
		return fmt.Errorf("no clients configured")
	}

	// Validate each MCP server
	for _, server := range config.MCPServers {
		if err := v.ValidateMCPServerConfig(server.Name, server.Config); err != nil {
			return fmt.Errorf("invalid MCP server '%s': %w", server.Name, err)
		}
	}

	// Build server name set for validation
	serverNames := make(map[string]bool)
	for _, server := range config.MCPServers {
		serverNames[server.Name] = true
	}

	// Validate each client
	for clientName, client := range config.Clients {
		if err := v.ValidateClient(clientName, client); err != nil {
			return fmt.Errorf("invalid client '%s': %w", clientName, err)
		}

		// Validate that enabled servers exist
		for _, serverName := range client.Enabled {
			if !serverNames[serverName] {
				return fmt.Errorf("client '%s' references non-existent server '%s'", clientName, serverName)
			}
		}
	}

	return nil
}