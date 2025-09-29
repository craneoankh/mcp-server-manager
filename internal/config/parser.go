package config

import (
	"fmt"

	"gopkg.in/yaml.v3"

	"github.com/vlazic/mcp-server-manager/internal/models"
)

// rawConfigData is the intermediate structure for YAML parsing
type rawConfigData struct {
	MCPServers map[string]map[string]interface{} `yaml:"mcpServers"`
	Clients    map[string]*models.Client         `yaml:"clients"`
	ServerPort int                               `yaml:"server_port"`
}

// extractServerOrder extracts the server order from YAML node structure
func extractServerOrder(node *yaml.Node) []string {
	var serverOrder []string

	if len(node.Content) == 0 || len(node.Content[0].Content) == 0 {
		return serverOrder
	}

	// Iterate through top-level keys to find mcpServers
	for i := 0; i < len(node.Content[0].Content); i += 2 {
		keyNode := node.Content[0].Content[i]
		if keyNode.Value == "mcpServers" && i+1 < len(node.Content[0].Content) {
			serversNode := node.Content[0].Content[i+1]
			// Extract keys in order
			for j := 0; j < len(serversNode.Content); j += 2 {
				serverName := serversNode.Content[j].Value
				serverOrder = append(serverOrder, serverName)
			}
			break
		}
	}

	return serverOrder
}

// buildOrderedServers converts the map to ordered slice based on server order
func buildOrderedServers(serverOrder []string, serversMap map[string]map[string]interface{}) []models.MCPServer {
	servers := make([]models.MCPServer, 0, len(serversMap))

	if len(serverOrder) > 0 {
		// Use explicit order
		for _, name := range serverOrder {
			if serverConfig, exists := serversMap[name]; exists {
				servers = append(servers, models.MCPServer{
					Name:   name,
					Config: serverConfig,
				})
			}
		}
	} else {
		// Fallback: map iteration (order not guaranteed)
		for name, serverConfig := range serversMap {
			servers = append(servers, models.MCPServer{
				Name:   name,
				Config: serverConfig,
			})
		}
	}

	return servers
}

// parseYAMLConfig parses YAML data and returns the config and server order
func parseYAMLConfig(data []byte) (*rawConfigData, []string, error) {
	var rawConfig rawConfigData

	// Use yaml.v3 Node to preserve order
	var node yaml.Node
	if err := yaml.Unmarshal(data, &node); err != nil {
		return nil, nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	if err := node.Decode(&rawConfig); err != nil {
		return nil, nil, fmt.Errorf("failed to decode config: %w", err)
	}

	serverOrder := extractServerOrder(&node)
	return &rawConfig, serverOrder, nil
}