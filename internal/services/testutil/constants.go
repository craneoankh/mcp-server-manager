package testutil

// Common test file names
const (
	TestConfigYAML = "config.yaml"
	TestClientJSON = "client.json"
)

// Common test paths
const (
	TestClientPath   = "~/.test.json"
	TestConfigPath   = "./config.yaml"
	TestExplicitYAML = "explicit.yaml"
)

// Common test server names
const (
	TestServerName = "test-server"
	HTTPServerName = "http-server"
	MultiTransport = "multi-transport"
)

// Common test client names
const (
	TestClientName = "test_client"
)

// Common test URLs
const (
	TestExampleURL  = "https://example.com"
	TestContext7URL = "https://mcp.context7.com/mcp"
)

// Common error message fragments
const (
	ErrNameEmpty           = "name cannot be empty"
	ErrExactlyOneTransport = "exactly one transport type"
)

// Common error format strings
const (
	ErrConfigFailedFmt             = "config failed: %v"
	ErrLoadConfigFailedFmt         = "LoadConfig failed: %v"
	ErrWriteConfigFailedFmt        = "Failed to write test config: %v"
	ErrWriteInitialConfigFailedFmt = "Failed to write initial config: %v"
	ErrReadClientConfigFailedFmt   = "ReadClientConfig failed: %v"
	ErrWriteClientConfigFailedFmt  = "WriteClientConfig failed: %v"
	ErrUpdateMCPStatusFailedFmt    = "UpdateMCPServerStatus failed: %v"
	ErrGetMCPStatusFailedFmt       = "GetMCPServerStatus failed: %v"
	ErrAddServerFailedFmt          = "AddServer failed: %v"
)