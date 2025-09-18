#!/bin/bash

echo "MCP Server Manager installed successfully!"
echo ""
echo "To start the service:"
echo "  systemctl --user enable --now mcp-server-manager"
echo ""
echo "Access the web interface at: http://localhost:6543"
echo "Configuration will be auto-created on first run at ~/.config/mcp-server-manager/config.yaml"
echo ""
echo "For more information, see: https://github.com/vlazic/mcp-server-manager"