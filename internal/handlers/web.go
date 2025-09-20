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
	clients := h.mcpManager.GetClients()

	c.HTML(http.StatusOK, "index.html", gin.H{
		"servers": servers,
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

	server, err := h.mcpManager.GetServerStatus(serverName)
	if err != nil {
		errorHTML := renderClientToggleWithError(clientName, serverName, "Error getting server status: "+err.Error())
		c.Data(http.StatusInternalServerError, "text/html", []byte(errorHTML))
		return
	}

	// Success - return normal toggle with hidden error container
	c.HTML(http.StatusOK, "client_toggle.html", gin.H{
		"server": server,
		"client": clientName,
	})
}

// Helper functions for error handling

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

