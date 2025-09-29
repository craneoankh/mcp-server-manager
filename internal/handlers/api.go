package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/vlazic/mcp-server-manager/internal/services"
)

type APIHandler struct {
	mcpManager *services.MCPManagerService
}

func NewAPIHandler(mcpManager *services.MCPManagerService) *APIHandler {
	return &APIHandler{
		mcpManager: mcpManager,
	}
}

func (h *APIHandler) GetMCPServers(c *gin.Context) {
	servers := h.mcpManager.GetMCPServers()
	c.JSON(http.StatusOK, gin.H{"servers": servers})
}

func (h *APIHandler) GetClients(c *gin.Context) {
	clients := h.mcpManager.GetClients()
	c.JSON(http.StatusOK, gin.H{"clients": clients})
}


func (h *APIHandler) ToggleClientServer(c *gin.Context) {
	clientName := c.Param("client")
	serverName := c.Param("server")
	enabledStr := c.PostForm("enabled")

	enabled, err := strconv.ParseBool(enabledStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid enabled value"})
		return
	}

	if err := h.mcpManager.ToggleClientMCPServer(clientName, serverName, enabled); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}

func (h *APIHandler) GetServerStatus(c *gin.Context) {
	serverName := c.Param("server")

	server, err := h.mcpManager.GetServerStatus(serverName)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, server)
}

func (h *APIHandler) SyncAllClients(c *gin.Context) {
	if err := h.mcpManager.SyncAllClients(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}

func (h *APIHandler) AddServer(c *gin.Context) {
	// Expect JSON in format: {"mcpServers": {"server-name": {config...}}}
	var requestBody struct {
		MCPServers map[string]map[string]interface{} `json:"mcpServers"`
	}

	if err := c.ShouldBindJSON(&requestBody); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON: " + err.Error()})
		return
	}

	if len(requestBody.MCPServers) != 1 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Must provide exactly one server in mcpServers"})
		return
	}

	// Extract the single server name and config
	var serverName string
	var serverConfig map[string]interface{}
	for name, config := range requestBody.MCPServers {
		serverName = name
		serverConfig = config
		break
	}

	if err := h.mcpManager.AddServer(serverName, serverConfig); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"server": map[string]interface{}{
			"name":   serverName,
			"config": serverConfig,
		},
	})
}