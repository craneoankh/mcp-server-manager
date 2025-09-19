// MCP Server Manager - Client-side JavaScript

/**
 * Configuration and state management
 */
const MCPManager = {
    // Configuration
    config: {
        fadeOutDelay: 3000,
        fadeOutDuration: 500,
        themeStorageKey: 'mcp-theme-preference'
    },

    // State
    state: {
        isSubmitting: false,
        currentTheme: 'system'
    },

    // DOM element getters
    elements: {
        get serverJsonTextarea() { return document.getElementById('server-json'); },
        get successMessage() { return document.getElementById('success-message'); },
        get errorMessage() { return document.getElementById('error-message'); },
        get errorText() { return document.getElementById('error-text'); },
        get newServerForm() { return document.getElementById('new-server-form'); },
        get addServerForm() { return document.getElementById('add-server-form'); },
        get themeOptions() { return document.querySelectorAll('.theme-option'); }
    }
};

/**
 * Form validation functions
 */
const FormValidator = {
    /**
     * Validates JSON syntax
     * @param {string} jsonText - JSON string to validate
     * @returns {Object} - {valid: boolean, data?: object, error?: string}
     */
    validateJSON(jsonText) {
        try {
            const data = JSON.parse(jsonText);
            return { valid: true, data };
        } catch (err) {
            return { valid: false, error: 'Invalid JSON syntax: ' + err.message };
        }
    },

    /**
     * Validates MCP client configuration format
     * @param {Object} parsedJSON - Parsed JSON object
     * @returns {Object} - {valid: boolean, serverName?: string, serverConfig?: object, error?: string}
     */
    validateMCPClientConfig(parsedJSON) {
        // Check for mcpServers wrapper
        if (!parsedJSON.mcpServers || typeof parsedJSON.mcpServers !== 'object') {
            return {
                valid: false,
                error: 'Invalid format. Please provide MCP client configuration: {"mcpServers": {"server-name": {...}}}'
            };
        }

        const serverNames = Object.keys(parsedJSON.mcpServers);

        if (serverNames.length === 0) {
            return { valid: false, error: 'No servers found in mcpServers object' };
        }

        if (serverNames.length > 1) {
            return { valid: false, error: 'Multiple servers found in mcpServers. Please add one server at a time.' };
        }

        const serverName = serverNames[0];
        const serverConfig = { ...parsedJSON.mcpServers[serverName] };
        serverConfig.name = serverName;

        return { valid: true, serverName, serverConfig };
    },

    /**
     * Validates server transport configuration
     * @param {Object} serverConfig - Server configuration object
     * @returns {Object} - {valid: boolean, normalizedConfig?: object, error?: string}
     */
    validateTransportConfig(serverConfig) {
        // Basic server name validation
        if (!serverConfig.name || typeof serverConfig.name !== 'string' || serverConfig.name.trim() === '') {
            return { valid: false, error: 'Server name is required and must be a non-empty string' };
        }

        // Transport validation
        const hasCommand = serverConfig.command && typeof serverConfig.command === 'string';
        const hasHttpUrl = (serverConfig.http_url && typeof serverConfig.http_url === 'string') ||
                          (serverConfig.httpUrl && typeof serverConfig.httpUrl === 'string');
        const hasUrl = serverConfig.url && typeof serverConfig.url === 'string';

        // Normalize httpUrl to http_url for internal consistency
        const normalizedConfig = { ...serverConfig };
        if (normalizedConfig.httpUrl && !normalizedConfig.http_url) {
            normalizedConfig.http_url = normalizedConfig.httpUrl;
            delete normalizedConfig.httpUrl;
        }

        const transportCount = (hasCommand ? 1 : 0) + (hasHttpUrl ? 1 : 0) + (hasUrl ? 1 : 0);

        if (transportCount === 0) {
            return { valid: false, error: 'Server must have exactly one transport type: command, http_url, or url' };
        }

        if (transportCount > 1) {
            return { valid: false, error: 'Server must have exactly one transport type, found ' + transportCount };
        }

        // Set default values
        if (normalizedConfig.enabled_globally === undefined) {
            normalizedConfig.enabled_globally = false;
        }
        if (!normalizedConfig.clients) {
            normalizedConfig.clients = {};
        }

        return { valid: true, normalizedConfig };
    },

    /**
     * Validates complete server configuration
     * @param {string} jsonText - JSON string from form
     * @returns {Object} - {valid: boolean, serverConfig?: object, error?: string}
     */
    validateServerConfiguration(jsonText) {
        // Step 1: JSON syntax validation
        const jsonResult = this.validateJSON(jsonText);
        if (!jsonResult.valid) {
            return { valid: false, error: jsonResult.error };
        }

        // Step 2: MCP client config format validation
        const configResult = this.validateMCPClientConfig(jsonResult.data);
        if (!configResult.valid) {
            return { valid: false, error: configResult.error };
        }

        // Step 3: Transport configuration validation
        const transportResult = this.validateTransportConfig(configResult.serverConfig);
        if (!transportResult.valid) {
            return { valid: false, error: transportResult.error };
        }

        return { valid: true, serverConfig: transportResult.normalizedConfig };
    }
};

/**
 * UI management functions
 */
const UIManager = {
    /**
     * Shows error message in the form
     * @param {string} message - Error message to display
     */
    showErrorMessage(message) {
        const errorText = MCPManager.elements.errorText;
        const errorMessage = MCPManager.elements.errorMessage;

        if (errorText && errorMessage) {
            errorText.textContent = message;
            errorMessage.classList.remove('hidden');
        }
    },

    /**
     * Hides error message
     */
    hideErrorMessage() {
        const errorMessage = MCPManager.elements.errorMessage;
        if (errorMessage) {
            errorMessage.classList.add('hidden');
        }
    },

    /**
     * Shows success message with auto fade-out
     */
    showSuccessMessage() {
        const successMsg = MCPManager.elements.successMessage;
        if (!successMsg) return;

        successMsg.classList.remove('hidden');

        // Fade out after delay
        setTimeout(() => {
            successMsg.style.transition = `opacity ${MCPManager.config.fadeOutDuration}ms ease-out`;
            successMsg.style.opacity = '0';

            setTimeout(() => {
                successMsg.classList.add('hidden');
                successMsg.style.opacity = '1';
                successMsg.style.transition = '';
            }, MCPManager.config.fadeOutDuration);
        }, MCPManager.config.fadeOutDelay);
    },

    /**
     * Hides success message
     */
    hideSuccessMessage() {
        const successMessage = MCPManager.elements.successMessage;
        if (successMessage) {
            successMessage.classList.add('hidden');
        }
    },

    /**
     * Resets form to default state
     */
    resetForm() {
        const textarea = MCPManager.elements.serverJsonTextarea;
        if (textarea) {
            textarea.value = `{
  "mcpServers": {
    "my-server": {
      "command": "npx",
      "args": ["@my/mcp-server"],
      "env": {
        "API_KEY": "your-key"
      }
    }
  }
}`;
        }
    },

    /**
     * Hides the add server form
     */
    hideForm() {
        const form = MCPManager.elements.addServerForm;
        if (form) {
            form.classList.add('hidden');
        }
    }
};

/**
 * Server API communication
 */
const ServerAPI = {
    /**
     * Submits new server configuration to backend
     * @param {Object} serverConfig - Server configuration object
     * @returns {Promise<Object>} - API response
     */
    async submitNewServer(serverConfig) {
        const response = await fetch('/api/servers', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify(serverConfig)
        });

        if (!response.ok) {
            const text = await response.text();
            try {
                const errorData = JSON.parse(text);
                const errorMessage = errorData.error || 'Server error';
                throw new Error(errorMessage);
            } catch (parseError) {
                throw new Error(text || 'Server error');
            }
        }

        return await response.json();
    }
};

/**
 * Theme management
 */
const ThemeManager = {
    /**
     * Gets the stored theme preference or system default
     * @returns {string} - Theme preference ('light', 'dark', 'system')
     */
    getStoredTheme() {
        try {
            return localStorage.getItem(MCPManager.config.themeStorageKey) || 'system';
        } catch (error) {
            console.warn('Theme: localStorage not available, using system theme');
            return 'system';
        }
    },

    /**
     * Stores theme preference
     * @param {string} theme - Theme to store ('light', 'dark', 'system')
     */
    storeTheme(theme) {
        try {
            localStorage.setItem(MCPManager.config.themeStorageKey, theme);
        } catch (error) {
            console.warn('Theme: localStorage not available, theme not persisted');
        }
    },

    /**
     * Gets system theme preference
     * @returns {string} - 'light' or 'dark'
     */
    getSystemTheme() {
        return window.matchMedia && window.matchMedia('(prefers-color-scheme: dark)').matches
            ? 'dark'
            : 'light';
    },

    /**
     * Applies theme to document
     * @param {string} theme - Theme to apply ('light', 'dark', 'system')
     */
    applyTheme(theme) {
        const root = document.documentElement;

        if (theme === 'system') {
            // Remove data-theme attribute to use CSS media query
            root.removeAttribute('data-theme');
        } else {
            // Set explicit theme
            root.setAttribute('data-theme', theme);
        }

        MCPManager.state.currentTheme = theme;
        this.updateThemeUI();
    },

    /**
     * Updates theme toggle UI to reflect current theme
     */
    updateThemeUI() {
        const options = MCPManager.elements.themeOptions;
        options.forEach(option => {
            const isActive = option.dataset.theme === MCPManager.state.currentTheme;
            option.classList.toggle('active', isActive);
        });
    },

    /**
     * Handles theme option click
     * @param {Event} event - Click event
     */
    handleThemeChange(event) {
        const theme = event.target.closest('.theme-option')?.dataset.theme;
        if (!theme) return;

        this.applyTheme(theme);
        this.storeTheme(theme);
    },

    /**
     * Initializes theme system
     */
    init() {
        // Load stored theme immediately to prevent flash
        const storedTheme = this.getStoredTheme();
        this.applyTheme(storedTheme);

        // Wait for DOM to be ready before setting up UI
        const setupUI = () => {
            // Add click handlers to theme options
            const options = MCPManager.elements.themeOptions;
            options.forEach(option => {
                option.addEventListener('click', (event) => this.handleThemeChange(event));
            });

            // Update UI to reflect current theme
            this.updateThemeUI();
        };

        // Setup UI after a short delay to ensure DOM is ready
        setTimeout(setupUI, 100);

        // Listen for system theme changes
        if (window.matchMedia) {
            const mediaQuery = window.matchMedia('(prefers-color-scheme: dark)');
            mediaQuery.addEventListener('change', () => {
                // Only update if currently using system theme
                if (MCPManager.state.currentTheme === 'system') {
                    this.applyTheme('system');
                }
            });
        }
    }
};

/**
 * Example configurations for different transport types
 */
const ExampleConfigs = {
    stdio: {
        mcpServers: {
            filesystem: {
                command: "npx",
                args: ["@modelcontextprotocol/server-filesystem", "/path/to/directory"],
                env: {
                    NODE_ENV: "production"
                }
            }
        }
    },

    http: {
        mcpServers: {
            "my-http-server": {
                httpUrl: "https://api.example.com/mcp",
                headers: {
                    Authorization: "Bearer YOUR_TOKEN",
                    "Content-Type": "application/json"
                }
            }
        }
    },

    sse: {
        mcpServers: {
            "sse-server": {
                url: "http://localhost:8080/sse",
                headers: {
                    "X-API-Key": "your-api-key",
                    Authorization: "Bearer token"
                }
            }
        }
    },

    context7: {
        mcpServers: {
            context7: {
                url: "https://mcp.context7.com/mcp",
                headers: {
                    CONTEXT7_API_KEY: "YOUR_API_KEY"
                }
            }
        }
    }
};

/**
 * Main form handling functions
 */
const FormHandler = {
    /**
     * Handles form submission
     * @param {Event} event - Form submit event
     */
    async handleFormSubmit(event) {
        event.preventDefault();

        if (MCPManager.state.isSubmitting) return;
        MCPManager.state.isSubmitting = true;

        try {
            const textarea = MCPManager.elements.serverJsonTextarea;
            const jsonText = textarea.value.trim();

            // Hide previous messages
            UIManager.hideSuccessMessage();
            UIManager.hideErrorMessage();

            // Validate configuration
            const validationResult = FormValidator.validateServerConfiguration(jsonText);
            if (!validationResult.valid) {
                UIManager.showErrorMessage(validationResult.error);
                return;
            }

            // Submit to backend
            await ServerAPI.submitNewServer(validationResult.serverConfig);

            // Handle success
            UIManager.showSuccessMessage();
            UIManager.resetForm();
            UIManager.hideForm();

            // Trigger config refresh and reload
            document.body.dispatchEvent(new CustomEvent('configChanged'));
            window.location.reload();

        } catch (error) {
            UIManager.showErrorMessage(error.message);
        } finally {
            MCPManager.state.isSubmitting = false;
        }
    },

    /**
     * Loads example configuration into textarea
     * @param {string} type - Example type (stdio, http, sse, context7)
     */
    loadExample(type) {
        const textarea = MCPManager.elements.serverJsonTextarea;
        const example = ExampleConfigs[type];

        if (example && textarea) {
            textarea.value = JSON.stringify(example, null, 2);
        }
    }
};

/**
 * Event listeners and initialization
 */
const EventHandlers = {
    /**
     * Sets up all event listeners
     */
    init() {
        // Initialize theme system
        ThemeManager.init();

        // HTMX configuration
        document.body.addEventListener('htmx:configRequest', function(evt) {
            evt.detail.headers['Content-Type'] = 'application/x-www-form-urlencoded';
        });

        // Re-highlight code after HTMX updates
        document.body.addEventListener('htmx:afterSwap', function(evt) {
            if (typeof Prism !== 'undefined') {
                Prism.highlightAllUnder(evt.detail.target);
            }
        });

        // Add server form submission
        const form = MCPManager.elements.newServerForm;
        if (form) {
            form.addEventListener('submit', FormHandler.handleFormSubmit);
        }
    }
};

// Global functions for onclick handlers (needed for template compatibility)
function loadExample(type) {
    FormHandler.loadExample(type);
}

// Initialize when DOM is loaded
if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', EventHandlers.init);
} else {
    EventHandlers.init();
}