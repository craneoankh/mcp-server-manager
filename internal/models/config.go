package models

type MCPServer struct {
	Name            string            `yaml:"name" json:"name"`

	// Transport Types (exactly one required)
	Command         string            `yaml:"command,omitempty" json:"command,omitempty"`           // STDIO transport
	URL             string            `yaml:"url,omitempty" json:"url,omitempty"`                   // SSE transport
	HttpURL         string            `yaml:"http_url,omitempty" json:"httpUrl,omitempty"`         // HTTP transport

	// Transport-specific properties
	Args            []string          `yaml:"args,omitempty" json:"args,omitempty"`                // STDIO only
	Cwd             string            `yaml:"cwd,omitempty" json:"cwd,omitempty"`                  // STDIO only
	Headers         map[string]string `yaml:"headers,omitempty" json:"headers,omitempty"`          // HTTP/SSE only

	// Common properties
	Env             map[string]string `yaml:"env,omitempty" json:"env,omitempty"`                  // Environment variables
	Timeout         int               `yaml:"timeout,omitempty" json:"timeout,omitempty"`          // Request timeout in ms
	Trust           bool              `yaml:"trust,omitempty" json:"trust,omitempty"`              // Bypass confirmations
	IncludeTools    []string          `yaml:"include_tools,omitempty" json:"includeTools,omitempty"` // Tool whitelist
	ExcludeTools    []string          `yaml:"exclude_tools,omitempty" json:"excludeTools,omitempty"` // Tool blacklist

	// Manager-specific properties
	Clients         map[string]bool   `yaml:"clients" json:"clients"`
}

type Client struct {
	Name       string `yaml:"name" json:"name"`
	ConfigPath string `yaml:"config_path" json:"config_path"`
}

type Config struct {
	MCPServers []MCPServer `yaml:"mcp_servers" json:"mcp_servers"`
	Clients    []Client    `yaml:"clients" json:"clients"`
	ServerPort int         `yaml:"server_port" json:"server_port"`
}

type ClientConfig struct {
	MCPServers map[string]interface{} `json:"mcpServers,omitempty"`
	// Keep other fields that might exist
	FeedbackSurveyState *map[string]interface{} `json:"feedbackSurveyState,omitempty"`
	SelectedAuthType    *string                 `json:"selectedAuthType,omitempty"`
	Theme               *string                 `json:"theme,omitempty"`
	PreferredEditor     *string                 `json:"preferredEditor,omitempty"`
	// Preserve any other unknown fields
	Other map[string]interface{} `json:"-"`
}

type MCPServerConfig struct {
	Command   string                 `json:"command,omitempty"`
	Args      []string               `json:"args,omitempty"`
	Env       map[string]string      `json:"env,omitempty"`
	HttpUrl   string                 `json:"httpUrl,omitempty"`
	Headers   map[string]interface{} `json:"headers,omitempty"`
}