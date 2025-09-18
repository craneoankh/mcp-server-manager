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

	// Validate transport type (exactly one required)
	transportCount := 0
	if server.Command != "" {
		transportCount++
	}
	if server.URL != "" {
		transportCount++
	}
	if server.HttpURL != "" {
		transportCount++
	}

	if transportCount == 0 {
		return fmt.Errorf("server must have exactly one transport type: command, url, or http_url")
	}
	if transportCount > 1 {
		return fmt.Errorf("server must have exactly one transport type, found %d", transportCount)
	}

	// Validate STDIO transport
	if server.Command != "" {
		if !v.IsCommandAvailable(server.Command) {
			return fmt.Errorf("command '%s' not found in PATH", server.Command)
		}
		// Args and Cwd are only valid for STDIO transport
	}

	// Validate SSE transport
	if server.URL != "" {
		if err := v.validateURL(server.URL); err != nil {
			return fmt.Errorf("invalid SSE URL '%s': %w", server.URL, err)
		}
		// Headers are valid for SSE transport
	}

	// Validate HTTP transport
	if server.HttpURL != "" {
		if err := v.validateURL(server.HttpURL); err != nil {
			return fmt.Errorf("invalid HTTP URL '%s': %w", server.HttpURL, err)
		}
		// Headers are valid for HTTP transport
	}

	// Validate common properties
	if server.Timeout < 0 {
		return fmt.Errorf("timeout cannot be negative")
	}

	// Validate environment variables
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

	// Validate tool lists
	for _, tool := range server.IncludeTools {
		if strings.TrimSpace(tool) == "" {
			return fmt.Errorf("include_tools cannot contain empty tool names")
		}
	}
	for _, tool := range server.ExcludeTools {
		if strings.TrimSpace(tool) == "" {
			return fmt.Errorf("exclude_tools cannot contain empty tool names")
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