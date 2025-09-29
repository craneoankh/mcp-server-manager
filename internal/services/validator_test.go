package services

import (
	"testing"
)

func TestValidateMCPServerConfig(t *testing.T) {
	validator := NewValidatorService()

	tests := []struct {
		name       string
		serverName string
		config     map[string]interface{}
		wantErr    bool
	}{
		{
			name:       "Valid STDIO server",
			serverName: "filesystem",
			config: map[string]interface{}{
				"command": "echo",
				"args":    []interface{}{"test"},
			},
			wantErr: false,
		},
		{
			name:       "Valid HTTP server with type",
			serverName: "context7",
			config: map[string]interface{}{
				"type": "http",
				"url":  "https://mcp.context7.com/mcp",
			},
			wantErr: false,
		},
		{
			name:       "Valid HTTP server with httpUrl",
			serverName: "context7-gemini",
			config: map[string]interface{}{
				"httpUrl": "https://mcp.context7.com/mcp",
			},
			wantErr: false,
		},
		{
			name:       "Empty server name",
			serverName: "",
			config: map[string]interface{}{
				"command": "echo",
			},
			wantErr: true,
		},
		{
			name:       "No transport type",
			serverName: "invalid",
			config:     map[string]interface{}{},
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateMCPServerConfig(tt.serverName, tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateMCPServerConfig() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}