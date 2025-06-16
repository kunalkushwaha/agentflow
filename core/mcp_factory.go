package core

import (
	"context"
	"fmt"
	"time"
)

// MCPFactory provides convenience methods for creating common MCP configurations
type MCPFactory struct{}

// NewMCPFactory creates a new MCP factory instance
func NewMCPFactory() *MCPFactory {
	return &MCPFactory{}
}

// CreateMockMCPManager creates an MCP manager with a mock server for testing
func (f *MCPFactory) CreateMockMCPManager() (*MCPManager, error) {
	config := MCPConfig{
		ClientType:    "mock",
		HealthCheck:   true,
		AutoDiscovery: false,
	}

	manager, err := NewMCPManager(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create mock MCP manager: %w", err)
	}

	// Add a mock server
	serverConfig := CreateMockServerConfig("mock-server", "Mock Test Server")
	if err := manager.AddServer(serverConfig); err != nil {
		return nil, fmt.Errorf("failed to add mock server: %w", err)
	}

	return manager, nil
}

// CreateProductionMCPManager creates an MCP manager suitable for production use
func (f *MCPFactory) CreateProductionMCPManager(clientType string) (*MCPManager, error) {
	if clientType == "" {
		clientType = "mark3labs"
	}

	config := MCPConfig{
		ClientType:    clientType,
		HealthCheck:   true,
		AutoDiscovery: true,
	}

	return NewMCPManager(config)
}

// CreateStdioMCPServer creates a server config for stdio-based MCP servers
func (f *MCPFactory) CreateStdioMCPServer(serverID, serverName string, command []string) MCPServerConfig {
	return MCPServerConfig{
		ID:         serverID,
		Name:       serverName,
		Type:       "stdio",
		ClientType: "mark3labs",
		Enabled:    true,
		Connection: MCPConnectionConfig{
			Transport: "stdio",
			Command:   command,
		},
		Timeout: 30000, // 30 seconds
	}
}

// CreateHTTPMCPServer creates a server config for HTTP-based MCP servers
func (f *MCPFactory) CreateHTTPMCPServer(serverID, serverName, endpoint string) MCPServerConfig {
	return MCPServerConfig{
		ID:         serverID,
		Name:       serverName,
		Type:       "http",
		ClientType: "mark3labs",
		Enabled:    true,
		Connection: MCPConnectionConfig{
			Transport: "http",
			Endpoint:  endpoint,
		},
		Timeout: 30000, // 30 seconds
	}
}

// CreateWebSocketMCPServer creates a server config for WebSocket-based MCP servers
func (f *MCPFactory) CreateWebSocketMCPServer(serverID, serverName, endpoint string) MCPServerConfig {
	return MCPServerConfig{
		ID:         serverID,
		Name:       serverName,
		Type:       "websocket",
		ClientType: "mark3labs",
		Enabled:    true,
		Connection: MCPConnectionConfig{
			Transport: "websocket",
			Endpoint:  endpoint,
		},
		Timeout: 30000, // 30 seconds
	}
}

// QuickStart provides a one-line setup for common MCP scenarios
type QuickStart struct {
	manager *MCPManager
}

// NewQuickStart creates a quick start instance for rapid MCP setup
func NewQuickStart() *QuickStart {
	return &QuickStart{}
}

// WithMockServer sets up MCP with a mock server for testing
func (q *QuickStart) WithMockServer() (*MCPManager, error) {
	factory := NewMCPFactory()
	manager, err := factory.CreateMockMCPManager()
	if err != nil {
		return nil, err
	}

	q.manager = manager
	return manager, nil
}

// WithStdioServer sets up MCP with a stdio-based server
func (q *QuickStart) WithStdioServer(serverName string, command []string) (*MCPManager, error) {
	factory := NewMCPFactory()
	manager, err := factory.CreateProductionMCPManager("mark3labs")
	if err != nil {
		return nil, err
	}

	serverConfig := factory.CreateStdioMCPServer("stdio-server", serverName, command)
	if err := manager.AddServer(serverConfig); err != nil {
		return nil, fmt.Errorf("failed to add stdio server: %w", err)
	}

	q.manager = manager
	return manager, nil
}

// WithHTTPServer sets up MCP with an HTTP-based server
func (q *QuickStart) WithHTTPServer(serverName, endpoint string) (*MCPManager, error) {
	factory := NewMCPFactory()
	manager, err := factory.CreateProductionMCPManager("mark3labs")
	if err != nil {
		return nil, err
	}

	serverConfig := factory.CreateHTTPMCPServer("http-server", serverName, endpoint)
	if err := manager.AddServer(serverConfig); err != nil {
		return nil, fmt.Errorf("failed to add HTTP server: %w", err)
	}

	q.manager = manager
	return manager, nil
}

// WithConfigFile sets up MCP from a configuration file
func (q *QuickStart) WithConfigFile(configPath string) (*MCPManager, error) {
	manager, err := NewMCPManagerFromConfig(configPath)
	if err != nil {
		return nil, err
	}

	q.manager = manager
	return manager, nil
}

// RefreshAndWait refreshes tools and waits for them to be available
func (q *QuickStart) RefreshAndWait(ctx context.Context, timeout time.Duration) error {
	if q.manager == nil {
		return fmt.Errorf("no MCP manager configured")
	}

	// Create a timeout context
	timeoutCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// Refresh tools
	if err := q.manager.RefreshTools(timeoutCtx); err != nil {
		return fmt.Errorf("failed to refresh tools: %w", err)
	}

	// Wait a bit for async operations to complete
	time.Sleep(100 * time.Millisecond)

	return nil
}

// MCPBuilder provides a fluent interface for building MCP configurations
type MCPBuilder struct {
	config      MCPConfig
	servers     []MCPServerConfig
	initialized bool
}

// NewMCPBuilder creates a new MCP builder
func NewMCPBuilder() *MCPBuilder {
	return &MCPBuilder{
		config: MCPConfig{
			ClientType:    "mock",
			HealthCheck:   true,
			AutoDiscovery: false,
		},
		servers: make([]MCPServerConfig, 0),
	}
}

// WithClientType sets the MCP client type
func (b *MCPBuilder) WithClientType(clientType string) *MCPBuilder {
	b.config.ClientType = clientType
	return b
}

// WithHealthCheck enables or disables health checking
func (b *MCPBuilder) WithHealthCheck(enabled bool) *MCPBuilder {
	b.config.HealthCheck = enabled
	return b
}

// WithAutoDiscovery enables or disables auto-discovery
func (b *MCPBuilder) WithAutoDiscovery(enabled bool) *MCPBuilder {
	b.config.AutoDiscovery = enabled
	return b
}

// AddMockServer adds a mock server to the configuration
func (b *MCPBuilder) AddMockServer(serverID, serverName string) *MCPBuilder {
	config := CreateMockServerConfig(serverID, serverName)
	b.servers = append(b.servers, config)
	return b
}

// AddStdioServer adds a stdio server to the configuration
func (b *MCPBuilder) AddStdioServer(serverID, serverName string, command []string) *MCPBuilder {
	factory := NewMCPFactory()
	config := factory.CreateStdioMCPServer(serverID, serverName, command)
	config.ClientType = b.config.ClientType
	b.servers = append(b.servers, config)
	return b
}

// AddHTTPServer adds an HTTP server to the configuration
func (b *MCPBuilder) AddHTTPServer(serverID, serverName, endpoint string) *MCPBuilder {
	factory := NewMCPFactory()
	config := factory.CreateHTTPMCPServer(serverID, serverName, endpoint)
	config.ClientType = b.config.ClientType
	b.servers = append(b.servers, config)
	return b
}

// AddCustomServer adds a custom server configuration
func (b *MCPBuilder) AddCustomServer(config MCPServerConfig) *MCPBuilder {
	b.servers = append(b.servers, config)
	return b
}

// Build creates the MCP manager with the configured settings
func (b *MCPBuilder) Build() (*MCPManager, error) {
	// Create the manager
	manager, err := NewMCPManager(b.config)
	if err != nil {
		return nil, fmt.Errorf("failed to create MCP manager: %w", err)
	}

	// Add all configured servers
	for _, serverConfig := range b.servers {
		if err := manager.AddServer(serverConfig); err != nil {
			return nil, fmt.Errorf("failed to add server %s: %w", serverConfig.ID, err)
		}
	}

	b.initialized = true
	return manager, nil
}

// BuildAndRefresh creates the MCP manager and refreshes tools
func (b *MCPBuilder) BuildAndRefresh(ctx context.Context) (*MCPManager, error) {
	manager, err := b.Build()
	if err != nil {
		return nil, err
	}

	// Refresh tools
	if err := manager.RefreshTools(ctx); err != nil {
		return nil, fmt.Errorf("failed to refresh tools: %w", err)
	}

	// Wait a bit for async operations
	time.Sleep(100 * time.Millisecond)

	return manager, nil
}
