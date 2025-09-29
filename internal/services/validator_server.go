package services

import (
	"fmt"
	"net/url"
	"os/exec"
	"strings"
)

// TransportType represents the type of MCP server transport
type TransportType int

const (
	TransportNone TransportType = iota
	TransportCommand
	TransportURL
	TransportHTTP
)

// detectTransportType identifies the transport type from server config
func detectTransportType(serverConfig map[string]interface{}) (TransportType, string, error) {
	var transportType TransportType
	var transportValue string
	count := 0

	if command, exists := serverConfig["command"]; exists && command != nil {
		if cmdStr, ok := command.(string); ok && strings.TrimSpace(cmdStr) != "" {
			transportType = TransportCommand
			transportValue = cmdStr
			count++
		}
	}

	if urlVal, exists := serverConfig["url"]; exists && urlVal != nil {
		if urlStr, ok := urlVal.(string); ok && strings.TrimSpace(urlStr) != "" {
			transportType = TransportURL
			transportValue = urlStr
			count++
		}
	}

	if httpUrl, exists := serverConfig["httpUrl"]; exists && httpUrl != nil {
		if httpUrlStr, ok := httpUrl.(string); ok && strings.TrimSpace(httpUrlStr) != "" {
			transportType = TransportHTTP
			transportValue = httpUrlStr
			count++
		}
	}

	if count == 0 {
		return TransportNone, "", fmt.Errorf("server must have exactly one transport type: command, url, or httpUrl")
	}
	if count > 1 {
		return TransportNone, "", fmt.Errorf("server must have exactly one transport type, found %d", count)
	}

	return transportType, transportValue, nil
}

// validateTransportValue validates the specific transport value based on type
func (v *ValidatorService) validateTransportValue(transportType TransportType, value string) error {
	switch transportType {
	case TransportCommand:
		if !v.IsCommandAvailable(value) {
			return fmt.Errorf("command '%s' not found in PATH", value)
		}
	case TransportURL, TransportHTTP:
		if err := v.validateURL(value); err != nil {
			return fmt.Errorf("invalid URL '%s': %w", value, err)
		}
	}
	return nil
}

// validateTimeout validates timeout configuration
func validateTimeout(serverConfig map[string]interface{}) error {
	if timeout, exists := serverConfig["timeout"]; exists && timeout != nil {
		if timeoutNum, ok := timeout.(int); ok && timeoutNum < 0 {
			return fmt.Errorf("timeout cannot be negative")
		}
	}
	return nil
}

// validateEnvironmentVariables validates environment variable configuration
func validateEnvironmentVariables(serverConfig map[string]interface{}) error {
	env, exists := serverConfig["env"]
	if !exists || env == nil {
		return nil
	}

	envMap, ok := env.(map[string]interface{})
	if !ok {
		return nil
	}

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
	return nil
}

// ValidateMCPServerConfig validates a server configuration map
func (v *ValidatorService) ValidateMCPServerConfig(serverName string, serverConfig map[string]interface{}) error {
	if strings.TrimSpace(serverName) == "" {
		return fmt.Errorf("server name cannot be empty")
	}

	// Detect and validate transport type
	transportType, transportValue, err := detectTransportType(serverConfig)
	if err != nil {
		return err
	}

	if err := v.validateTransportValue(transportType, transportValue); err != nil {
		return err
	}

	// Validate optional fields
	if err := validateTimeout(serverConfig); err != nil {
		return err
	}

	if err := validateEnvironmentVariables(serverConfig); err != nil {
		return err
	}

	return nil
}

// IsCommandAvailable checks if a command is available in PATH
func (v *ValidatorService) IsCommandAvailable(command string) bool {
	_, err := exec.LookPath(command)
	return err == nil
}

// validateURL validates a URL string
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