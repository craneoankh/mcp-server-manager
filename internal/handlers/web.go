package handlers

import (
	"fmt"
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
	clientsMap := h.mcpManager.GetClients()

	// Convert to view structures
	type ServerView struct {
		Name   string
		Config map[string]interface{}
	}

	type ClientView struct {
		Name       string
		ConfigPath string
		Enabled    []string
	}

	// Servers already ordered from config
	serverViews := make([]ServerView, 0, len(servers))
	for _, server := range servers {
		serverViews = append(serverViews, ServerView{
			Name:   server.Name,
			Config: server.Config,
		})
	}

	clients := make([]ClientView, 0, len(clientsMap))
	for name, client := range clientsMap {
		clients = append(clients, ClientView{
			Name:       name,
			ConfigPath: client.ConfigPath,
			Enabled:    client.Enabled,
		})
	}

	c.HTML(http.StatusOK, "index.html", gin.H{
		"servers": serverViews,
		"clients": clients,
	})
}


func (h *WebHandler) ToggleClientServerHTMX(c *gin.Context) {
	clientName := c.Param("client")
	serverName := c.Param("server")
	enabledStr := c.PostForm("enabled")

	enabled, err := strconv.ParseBool(enabledStr)
	if err != nil {
		errorHTML := renderClientToggleWithError(clientName, serverName, "Invalid enabled value: "+enabledStr)
		c.Data(http.StatusBadRequest, "text/html", []byte(errorHTML))
		return
	}

	if err := h.mcpManager.ToggleClientMCPServer(clientName, serverName, enabled); err != nil {
		errorHTML := renderClientToggleWithError(clientName, serverName, "Error: "+err.Error())
		c.Data(http.StatusBadRequest, "text/html", []byte(errorHTML))
		return
	}

	serverConfig, err := h.mcpManager.GetServerStatus(serverName)
	if err != nil {
		errorHTML := renderClientToggleWithError(clientName, serverName, "Error getting server status: "+err.Error())
		c.Data(http.StatusInternalServerError, "text/html", []byte(errorHTML))
		return
	}

	// Get client to check enabled status
	clients := h.mcpManager.GetClients()
	client, exists := clients[clientName]
	if !exists {
		errorHTML := renderClientToggleWithError(clientName, serverName, "Client not found")
		c.Data(http.StatusInternalServerError, "text/html", []byte(errorHTML))
		return
	}

	// Success - return normal toggle with hidden error container
	c.HTML(http.StatusOK, "client_toggle.html", gin.H{
		"serverName":    serverName,
		"serverConfig":  serverConfig,
		"client":        clientName,
		"clientEnabled": client.Enabled,
	})
}

// Helper functions

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func renderErrorBox(message string) string {
	return fmt.Sprintf(`
		<div class="text-red-600 text-sm font-medium p-2 bg-red-50 rounded border border-red-200">
			%s
		</div>
	`, message)
}

func renderClientToggleWithError(clientName, serverName, errorMessage string) string {
	errorContainer := fmt.Sprintf(`
		<div id="error-%s-%s" class="mb-2">
			%s
		</div>
	`, clientName, serverName, renderErrorBox(errorMessage))

	// Simple checkbox (we can't easily render the full template here)
	toggleHTML := `
		<label class="inline-flex items-center">
			<input type="checkbox"
				   class="form-checkbox h-5 w-5 text-green-600"
				   disabled
				   title="Fix the error above to enable toggling">
		</label>
	`

	return errorContainer + toggleHTML
}

