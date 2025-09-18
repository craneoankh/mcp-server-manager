package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/vlazic/mcp-server-manager/internal/services"
)

type WebHandler struct {
	mcpManager *services.MCPManagerService
}

func NewWebHandler(mcpManager *services.MCPManagerService) *WebHandler {
	return &WebHandler{
		mcpManager: mcpManager,
	}
}

func (h *WebHandler) Index(c *gin.Context) {
	servers := h.mcpManager.GetMCPServers()
	clients := h.mcpManager.GetClients()

	c.HTML(http.StatusOK, "index.html", gin.H{
		"servers": servers,
		"clients": clients,
	})
}

func (h *WebHandler) ToggleGlobalServerHTMX(c *gin.Context) {
	serverName := c.Param("server")
	enabledStr := c.PostForm("enabled")

	enabled, err := strconv.ParseBool(enabledStr)
	if err != nil {
		c.String(http.StatusBadRequest, "Invalid enabled value")
		return
	}

	if err := h.mcpManager.ToggleGlobalMCPServer(serverName, enabled); err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	server, _ := h.mcpManager.GetServerStatus(serverName)
	c.HTML(http.StatusOK, "server_row.html", gin.H{
		"server":  server,
		"clients": h.mcpManager.GetClients(),
	})
}

func (h *WebHandler) ToggleClientServerHTMX(c *gin.Context) {
	clientName := c.Param("client")
	serverName := c.Param("server")
	enabledStr := c.PostForm("enabled")

	enabled, err := strconv.ParseBool(enabledStr)
	if err != nil {
		c.String(http.StatusBadRequest, "Invalid enabled value: %s", enabledStr)
		return
	}

	if err := h.mcpManager.ToggleClientMCPServer(clientName, serverName, enabled); err != nil {
		c.String(http.StatusInternalServerError, "Error toggling server: %s", err.Error())
		return
	}

	server, err := h.mcpManager.GetServerStatus(serverName)
	if err != nil {
		c.String(http.StatusInternalServerError, "Error getting server status: %s", err.Error())
		return
	}

	c.HTML(http.StatusOK, "client_toggle.html", gin.H{
		"server": server,
		"client": clientName,
	})
}