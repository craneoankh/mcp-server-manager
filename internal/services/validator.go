package services

import (
	"fmt"
	"net/url"
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

	// Validate each MCP server
	for serverName, serverConfig := range config.MCPServers {
		if err := v.ValidateMCPServerConfig(serverName, serverConfig); err != nil {
			return fmt.Errorf("invalid MCP server '%s': %w", serverName, err)
		}
	}

	// Validate each client
	for clientName, client := range config.Clients {
		if err := v.ValidateClient(clientName, client); err != nil {
			return fmt.Errorf("invalid client '%s': %w", clientName, err)
		}

		// Validate that enabled servers exist
		for _, serverName := range client.Enabled {
			if _, exists := config.MCPServers[serverName]; !exists {
				return fmt.Errorf("client '%s' references non-existent server '%s'", clientName, serverName)
			}
		}
	}

	return nil
}

// ValidateMCPServerConfig validates a server configuration map
func (v *ValidatorService) ValidateMCPServerConfig(serverName string, serverConfig map[string]interface{}) error {
	if strings.TrimSpace(serverName) == "" {
		return fmt.Errorf("server name cannot be empty")
	}

	// Validate transport type (exactly one required)
	hasCommand := false
	hasURL := false
	hasHttpURL := false

	if command, exists := serverConfig["command"]; exists && command != nil {
		if cmdStr, ok := command.(string); ok && strings.TrimSpace(cmdStr) != "" {
			hasCommand = true
			// Validate command is in PATH
			if !v.IsCommandAvailable(cmdStr) {
				return fmt.Errorf("command '%s' not found in PATH", cmdStr)
			}
		}
	}

	if url, exists := serverConfig["url"]; exists && url != nil {
		if urlStr, ok := url.(string); ok && strings.TrimSpace(urlStr) != "" {
			hasURL = true
			if err := v.validateURL(urlStr); err != nil {
				return fmt.Errorf("invalid URL '%s': %w", urlStr, err)
			}
		}
	}

	if httpUrl, exists := serverConfig["httpUrl"]; exists && httpUrl != nil {
		if httpUrlStr, ok := httpUrl.(string); ok && strings.TrimSpace(httpUrlStr) != "" {
			hasHttpURL = true
			if err := v.validateURL(httpUrlStr); err != nil {
				return fmt.Errorf("invalid httpUrl '%s': %w", httpUrlStr, err)
			}
		}
	}

	transportCount := 0
	if hasCommand {
		transportCount++
	}
	if hasURL {
		transportCount++
	}
	if hasHttpURL {
		transportCount++
	}

	if transportCount == 0 {
		return fmt.Errorf("server must have exactly one transport type: command, url, or httpUrl")
	}
	if transportCount > 1 {
		return fmt.Errorf("server must have exactly one transport type, found %d", transportCount)
	}

	// Validate timeout if present
	if timeout, exists := serverConfig["timeout"]; exists && timeout != nil {
		if timeoutNum, ok := timeout.(int); ok && timeoutNum < 0 {
			return fmt.Errorf("timeout cannot be negative")
		}
	}

	// Validate environment variables if present
	if env, exists := serverConfig["env"]; exists && env != nil {
		if envMap, ok := env.(map[string]interface{}); ok {
			for key, value := range envMap {
				if strings.TrimSpace(key) == "" {
					return fmt.Errorf("environment variable key cannot be empty")
				}
				if strings.Contains(key, "=") {
					return fmt.Errorf("environment variable key cannot contain '='")
				}
				if valStr, ok := value.(string); !ok || strings.TrimSpace(valStr) == "" {
					return fmt.Errorf("environment variable value for '%s' cannot be empty", key)
				}
			}
		}
	}

	return nil
}

func (v *ValidatorService) ValidateClient(clientName string, client *models.Client) error {
	if strings.TrimSpace(clientName) == "" {
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

func (v *ValidatorService) validateURL(urlStr string) error {
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return err
	}

	// Require scheme and host
	if parsedURL.Scheme == "" {
		return fmt.Errorf("URL missing scheme")
	}
	if parsedURL.Host == "" {
		return fmt.Errorf("URL missing host")
	}

	// Only allow http and https schemes
	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return fmt.Errorf("URL scheme must be http or https, got %s", parsedURL.Scheme)
	}

	return nil
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
			// Check if command exists (for STDIO servers)
			if command, exists := serverMap["command"]; exists {
				if commandStr, ok := command.(string); !ok || commandStr == "" {
					return fmt.Errorf("server '%s' command must be a non-empty string", name)
				}
			}

			// Check if httpUrl exists (for HTTP servers)
			if httpUrl, exists := serverMap["httpUrl"]; exists {
				if httpUrlStr, ok := httpUrl.(string); !ok || httpUrlStr == "" {
					return fmt.Errorf("server '%s' httpUrl must be a non-empty string", name)
				}
			}

			// Check if url exists (for SSE servers)
			if url, exists := serverMap["url"]; exists {
				if urlStr, ok := url.(string); !ok || urlStr == "" {
					return fmt.Errorf("server '%s' url must be a non-empty string", name)
				}
			}

			// Must have exactly one transport type
			_, hasCommand := serverMap["command"]
			_, hasHttpUrl := serverMap["httpUrl"]
			_, hasUrl := serverMap["url"]

			transportCount := 0
			if hasCommand {
				transportCount++
			}
			if hasHttpUrl {
				transportCount++
			}
			if hasUrl {
				transportCount++
			}

			if transportCount == 0 {
				return fmt.Errorf("server '%s' must have exactly one transport type: command, httpUrl, or url", name)
			}
			if transportCount > 1 {
				return fmt.Errorf("server '%s' must have exactly one transport type, found %d", name, transportCount)
			}
		}
	}

	return nil
}