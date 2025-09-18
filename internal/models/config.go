package models

type MCPServer struct {
	Name            string            `yaml:"name" json:"name"`
	Command         string            `yaml:"command" json:"command"`
	Args            []string          `yaml:"args" json:"args"`
	Env             map[string]string `yaml:"env,omitempty" json:"env,omitempty"`
	EnabledGlobally bool              `yaml:"enabled_globally" json:"enabled_globally"`
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
}

type MCPServerConfig struct {
	Command   string                 `json:"command,omitempty"`
	Args      []string               `json:"args,omitempty"`
	Env       map[string]string      `json:"env,omitempty"`
	HttpUrl   string                 `json:"httpUrl,omitempty"`
	Headers   map[string]interface{} `json:"headers,omitempty"`
}