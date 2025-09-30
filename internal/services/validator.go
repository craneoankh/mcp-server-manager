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
	if err := v.validateBasicConfig(config); err != nil {
		return err
	}

	if err := v.validateMCPServers(config.MCPServers); err != nil {
		return err
	}

	serverNames := buildServerNameSet(config.MCPServers)

	if err := v.validateClients(config.Clients, serverNames); err != nil {
		return err
	}

	return nil
}

// validateBasicConfig checks port and existence of servers/clients
func (v *ValidatorService) validateBasicConfig(config *models.Config) error {
	if config.ServerPort < 1 || config.ServerPort > 65535 {
		return fmt.Errorf("invalid server port: %d", config.ServerPort)
	}

	if len(config.MCPServers) == 0 {
		return fmt.Errorf("no MCP servers configured")
	}

	if len(config.Clients) == 0 {
		return fmt.Errorf("no clients configured")
	}

	return nil
}

// validateMCPServers validates all MCP server configurations
func (v *ValidatorService) validateMCPServers(servers []models.MCPServer) error {
	for _, server := range servers {
		if err := v.ValidateMCPServerConfig(server.Name, server.Config); err != nil {
			return fmt.Errorf("invalid MCP server '%s': %w", server.Name, err)
		}
	}
	return nil
}

// buildServerNameSet creates a map of server names for lookup
func buildServerNameSet(servers []models.MCPServer) map[string]bool {
	serverNames := make(map[string]bool)
	for _, server := range servers {
		serverNames[server.Name] = true
	}
	return serverNames
}

// validateClients validates all client configurations and their server references
func (v *ValidatorService) validateClients(clients map[string]*models.Client, serverNames map[string]bool) error {
	for clientName, client := range clients {
		if err := v.ValidateClient(clientName, client); err != nil {
			return fmt.Errorf("invalid client '%s': %w", clientName, err)
		}

		if err := validateClientServerReferences(clientName, client, serverNames); err != nil {
			return err
		}
	}
	return nil
}

// validateClientServerReferences checks that all enabled servers exist
func validateClientServerReferences(clientName string, client *models.Client, serverNames map[string]bool) error {
	for _, serverName := range client.Enabled {
		if !serverNames[serverName] {
			return fmt.Errorf("client '%s' references non-existent server '%s'", clientName, serverName)
		}
	}
	return nil
}