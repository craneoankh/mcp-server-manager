package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/vlazic/mcp-server-manager/internal/config"
	"github.com/vlazic/mcp-server-manager/internal/services"
)

type ConfigViewerHandler struct {
	mcpManager *services.MCPManagerService
	configPath string
}

func NewConfigViewerHandler(mcpManager *services.MCPManagerService, configPath string) *ConfigViewerHandler {
	return &ConfigViewerHandler{
		mcpManager: mcpManager,
		configPath: configPath,
	}
}

func (h *ConfigViewerHandler) GetAppConfig(c *gin.Context) {
	cfg, _, err := config.LoadConfig(h.configPath)
	if err != nil {
		c.String(http.StatusInternalServerError, "Error loading config: %s", err.Error())
		return
	}

	configJson, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		c.String(http.StatusInternalServerError, "Error marshaling config: %s", err.Error())
		return
	}

	c.HTML(http.StatusOK, "config_content.html", gin.H{
		"title":    "Application Config",
		"content":  string(configJson),
		"language": "json",
	})
}

func (h *ConfigViewerHandler) GetClientConfig(c *gin.Context) {
	clientName := c.Param("client")

	clientConfigService := services.NewClientConfigService(h.mcpManager.GetConfig())
	clientConfig, err := clientConfigService.ReadClientConfig(clientName)
	if err != nil {
		c.String(http.StatusInternalServerError, "Error loading client config: %s", err.Error())
		return
	}

	configJson, err := json.MarshalIndent(clientConfig, "", "  ")
	if err != nil {
		c.String(http.StatusInternalServerError, "Error marshaling config: %s", err.Error())
		return
	}

	c.HTML(http.StatusOK, "config_content.html", gin.H{
		"title":    clientName + " Config",
		"content":  string(configJson),
		"language": "json",
	})
}