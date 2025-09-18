package services

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/vlazic/mcp-server-manager/internal/models"
)

type ValidatorService struct{}

func NewValidatorService() *ValidatorService {
	return &ValidatorService{}
}

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

	for _, server := range config.MCPServers {
		if err := v.ValidateMCPServer(&server); err != nil {
			return fmt.Errorf("invalid MCP server '%s': %w", server.Name, err)
		}
	}

	for _, client := range config.Clients {
		if err := v.ValidateClient(&client); err != nil {
			return fmt.Errorf("invalid client '%s': %w", client.Name, err)
		}
	}

	return nil
}

func (v *ValidatorService) ValidateMCPServer(server *models.MCPServer) error {
	if server.Name == "" {
		return fmt.Errorf("server name cannot be empty")
	}

	if server.Command == "" {
		return fmt.Errorf("server command cannot be empty")
	}

	if !v.IsCommandAvailable(server.Command) {
		return fmt.Errorf("command '%s' not found in PATH", server.Command)
	}

	for key, value := range server.Env {
		if strings.TrimSpace(key) == "" {
			return fmt.Errorf("environment variable key cannot be empty")
		}
		if strings.Contains(key, "=") {
			return fmt.Errorf("environment variable key cannot contain '='")
		}
		if strings.TrimSpace(value) == "" {
			return fmt.Errorf("environment variable value for '%s' cannot be empty", key)
		}
	}

	return nil
}

func (v *ValidatorService) ValidateClient(client *models.Client) error {
	if client.Name == "" {
		return fmt.Errorf("client name cannot be empty")
	}

	if client.ConfigPath == "" {
		return fmt.Errorf("client config path cannot be empty")
	}

	// Don't require the directory to exist - we'll create it if needed
	return nil
}

func (v *ValidatorService) IsCommandAvailable(command string) bool {
	_, err := exec.LookPath(command)
	return err == nil
}

func (v *ValidatorService) ValidateClientConfig(clientConfig *models.ClientConfig) error {
	if clientConfig.MCPServers == nil {
		return nil
	}

	for name, serverInterface := range clientConfig.MCPServers {
		if strings.TrimSpace(name) == "" {
			return fmt.Errorf("server name cannot be empty")
		}

		// Basic validation - just check if it's a map
		if serverMap, ok := serverInterface.(map[string]interface{}); ok {
			// Check if command exists (for command-based servers)
			if command, exists := serverMap["command"]; exists {
				if commandStr, ok := command.(string); !ok || commandStr == "" {
					return fmt.Errorf("server '%s' command must be a non-empty string", name)
				}
			}

			// Check if httpUrl exists (for HTTP-based servers)
			if httpUrl, exists := serverMap["httpUrl"]; exists {
				if httpUrlStr, ok := httpUrl.(string); !ok || httpUrlStr == "" {
					return fmt.Errorf("server '%s' httpUrl must be a non-empty string", name)
				}
			}

			// If neither command nor httpUrl exists, that's an error
			if _, hasCommand := serverMap["command"]; !hasCommand {
				if _, hasHttpUrl := serverMap["httpUrl"]; !hasHttpUrl {
					return fmt.Errorf("server '%s' must have either 'command' or 'httpUrl'", name)
				}
			}
		}
	}

	return nil
}