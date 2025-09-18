package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/vlazic/mcp-server-manager/internal/models"
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

func (h *APIHandler) ToggleGlobalServer(c *gin.Context) {
	serverName := c.Param("server")
	enabledStr := c.PostForm("enabled")

	enabled, err := strconv.ParseBool(enabledStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid enabled value"})
		return
	}

	if err := h.mcpManager.ToggleGlobalMCPServer(serverName, enabled); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
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
	var server models.MCPServer
	if err := c.ShouldBindJSON(&server); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON: " + err.Error()})
		return
	}

	if err := h.mcpManager.AddServer(&server); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "server": server})
}