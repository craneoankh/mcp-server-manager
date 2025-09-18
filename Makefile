BINARY_NAME=mcp-server-manager
BUILD_DIR=bin
SERVICE_NAME=mcp-server-manager.service

.PHONY: build run install-service enable-service disable-service start-service stop-service status-service clean

build:
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	@go build -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/server

run: build
	@echo "Running $(BINARY_NAME)..."
	@./$(BUILD_DIR)/$(BINARY_NAME)

install-deps:
	@echo "Installing Go dependencies..."
	@go mod tidy
	@go mod download

install-service: build
	@echo "Installing binary to /usr/local/bin/..."
	@sudo cp $(BUILD_DIR)/$(BINARY_NAME) /usr/local/bin/
	@echo "Installing systemd user service..."
	@mkdir -p ~/.config/systemd/user
	@cp systemd/$(SERVICE_NAME) ~/.config/systemd/user/
	@systemctl --user daemon-reload
	@echo "Service installed for current user. Use 'make enable-service' to enable it."

enable-service:
	@echo "Enabling $(SERVICE_NAME)..."
	@systemctl --user enable $(SERVICE_NAME)

disable-service:
	@echo "Disabling $(SERVICE_NAME)..."
	@systemctl --user disable $(SERVICE_NAME)

start-service:
	@echo "Starting $(SERVICE_NAME)..."
	@systemctl --user start $(SERVICE_NAME)

stop-service:
	@echo "Stopping $(SERVICE_NAME)..."
	@systemctl --user stop $(SERVICE_NAME)

restart-service:
	@echo "Restarting $(SERVICE_NAME)..."
	@systemctl --user restart $(SERVICE_NAME)

status-service:
	@systemctl --user status $(SERVICE_NAME)

logs-service:
	@journalctl --user -u $(SERVICE_NAME) -f

clean:
	@echo "Cleaning build artifacts..."
	@rm -rf $(BUILD_DIR)

setup: install-deps build install-service enable-service start-service
	@echo "Setup complete! MCP Manager is running on http://localhost:6543"
