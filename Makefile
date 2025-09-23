BINARY_NAME=mcp-server-manager
BUILD_DIR=bin
SERVICE_NAME=mcp-server-manager.service

.PHONY: build run install-service enable-service disable-service start-service stop-service status-service clean sync-assets test-release release

build: sync-assets
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

sync-assets:
	@echo "Syncing web assets to internal/assets/web/..."
	@mkdir -p internal/assets/web/templates internal/assets/web/static
	@cp -r web/templates/* internal/assets/web/templates/
	@cp -r web/static/* internal/assets/web/static/
	@echo "Assets synced successfully"

test-release: sync-assets
	@echo "Building local test release..."
	@goreleaser release --snapshot --clean --skip=publish
	@echo "Installing .deb package..."
	@sudo dpkg -i dist/$(BINARY_NAME)_*_linux_amd64.deb
	@echo "Restarting service..."
	@systemctl --user restart --now $(BINARY_NAME)
	@echo ""
	@echo "✅ Test release complete! Go to: http://localhost:6543"

release: sync-assets
	@echo "Creating release..."
	@if [ -n "$$(git status --porcelain)" ]; then \
		echo "❌ Error: Working directory is not clean. Please commit or stash changes."; \
		exit 1; \
	fi
	@if [ -z "$(VERSION)" ]; then \
		echo "❌ Please specify VERSION: make release VERSION=v1.1.0"; \
		exit 1; \
	fi
	@echo "Current branch: $$(git branch --show-current)"
	@echo "Last commit: $$(git log -1 --oneline)"
	@echo ""
	@if [ -n "$(MESSAGE)" ]; then \
		echo "Creating annotated tag: $(VERSION) with custom message"; \
		git tag -a "$(VERSION)" -m "$(MESSAGE)"; \
	else \
		echo "Creating lightweight tag: $(VERSION)"; \
		git tag "$(VERSION)"; \
	fi && \
	echo "Pushing tag: $(VERSION)" && \
	git push origin "$(VERSION)" && \
	echo "" && \
	echo "✅ Release $(VERSION) created!" && \
	echo "🚀 GitHub Actions will build cross-platform binaries" && \
	echo "📦 Check: https://github.com/vlazic/mcp-server-manager/actions" && \
	echo "📋 Releases: https://github.com/vlazic/mcp-server-manager/releases"

setup: install-deps build install-service enable-service start-service
	@echo "Setup complete! MCP Manager is running on http://localhost:6543"
