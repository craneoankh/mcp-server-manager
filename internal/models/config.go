package models

// Client represents an MCP client configuration
type Client struct {
	ConfigPath string   `yaml:"config_path" json:"config_path"`
	Enabled    []string `yaml:"enabled,omitempty" json:"enabled,omitempty"` // List of enabled server names
}

// Config is the main application configuration
type Config struct {
	MCPServers map[string]map[string]interface{} `yaml:"mcpServers" json:"mcpServers"` // Server name -> config map
	Clients    map[string]*Client                `yaml:"clients" json:"clients"`       // Client name -> client config
	ServerPort int                               `yaml:"server_port" json:"server_port"`
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