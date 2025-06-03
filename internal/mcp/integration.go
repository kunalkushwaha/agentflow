package mcp

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/kunalkushwaha/agentflow/internal/tools"
)

// MCPIntegration provides the main entry point for MCP integration with AgentFlow
type MCPIntegration struct {
	serverManager MCPServerManager
	toolAdapter   MCPToolAdapter
	toolRegistry  *tools.ToolRegistry
	clientFactory MCPClientFactory
	mu            sync.RWMutex
}

// MCPIntegrationConfig represents configuration for creating an MCP integration
type MCPIntegrationConfig struct {
	ClientType    string `json:"client_type"`
	HealthCheck   bool   `json:"health_check"`
	AutoDiscovery bool   `json:"auto_discovery"`
}

// MCPConfig represents the configuration for MCP integration
// This type alias points to the proper configuration structure from config.go
type MCPConfig = MCPConfigSection

// NewMCPIntegration creates a new MCP integration instance
func NewMCPIntegration(config MCPIntegrationConfig) (*MCPIntegration, error) {
	// Create a default tool registry if not provided
	toolRegistry := tools.NewToolRegistry()

	// Create client factory
	factory := NewClientFactory()

	// Use specified client as default
	if config.ClientType != "" {
		// This would be implemented based on the specific factory logic
	}

	// Create server manager
	serverManager := NewDefaultMCPServerManager(factory)

	// Create tool adapter
	toolAdapter := NewDefaultMCPToolAdapter()

	integration := &MCPIntegration{
		serverManager: serverManager,
		toolAdapter:   toolAdapter,
		toolRegistry:  toolRegistry,
		clientFactory: factory,
	}

	// Set up event handling
	serverManager.SetServerEventHandler(integration.handleServerEvent)

	// Start health checking if enabled
	if config.HealthCheck {
		serverManager.StartHealthChecking()
	}

	// Start auto-discovery if enabled
	if config.AutoDiscovery {
		ctx := context.Background()
		// Use a default interval for auto-discovery
		if err := serverManager.StartAutoDiscovery(ctx, 30*time.Second); err != nil {
			return nil, fmt.Errorf("failed to start auto-discovery: %w", err)
		}
	}

	return integration, nil
}

// NewMCPIntegrationWithRegistry creates a new MCP integration instance with a provided tool registry
func NewMCPIntegrationWithRegistry(toolRegistry *tools.ToolRegistry, config MCPIntegrationConfig) (*MCPIntegration, error) {
	if toolRegistry == nil {
		return nil, fmt.Errorf("tool registry cannot be nil")
	}

	// Create client factory
	factory := NewClientFactory()

	// Use specified client as default
	if config.ClientType != "" {
		// This would be implemented based on the specific factory logic
	}

	// Create server manager
	serverManager := NewDefaultMCPServerManager(factory)

	// Create tool adapter
	toolAdapter := NewDefaultMCPToolAdapter()

	integration := &MCPIntegration{
		serverManager: serverManager,
		toolAdapter:   toolAdapter,
		toolRegistry:  toolRegistry,
		clientFactory: factory,
	}

	// Set up event handling
	serverManager.SetServerEventHandler(integration.handleServerEvent)

	// Start health checking if enabled
	if config.HealthCheck {
		serverManager.StartHealthChecking()
	}

	// Start auto-discovery if enabled
	if config.AutoDiscovery {
		ctx := context.Background()
		// Use a default interval for auto-discovery
		if err := serverManager.StartAutoDiscovery(ctx, 30*time.Second); err != nil {
			return nil, fmt.Errorf("failed to start auto-discovery: %w", err)
		}
	}

	return integration, nil
}

// NewMCPIntegrationFromConfig creates a new MCP integration instance from TOML configuration
func NewMCPIntegrationFromConfig(toolRegistry *tools.ToolRegistry, configPath string) (*MCPIntegration, error) {
	// Load configuration
	loader := NewConfigLoader(configPath)
	config, err := loader.LoadConfiguration()
	if err != nil {
		return nil, fmt.Errorf("failed to load configuration: %w", err)
	}

	// Check if MCP is enabled
	if !config.MCP.Enabled {
		return nil, fmt.Errorf("MCP integration is disabled in configuration")
	}

	// Convert to integration config
	integrationConfig := loader.GetIntegrationConfig(config)

	// Create integration with registry
	integration, err := NewMCPIntegrationWithRegistry(toolRegistry, integrationConfig)
	if err != nil {
		return nil, err
	}

	// Convert server configurations and add them
	serverConfigs := loader.ConvertToServerConfigs(config)
	for _, serverConfig := range serverConfigs {
		if err := integration.AddServer(serverConfig); err != nil {
			return nil, fmt.Errorf("failed to add server %s: %w", serverConfig.ID, err)
		}
	}

	return integration, nil
}

// AddServer adds an MCP server to the integration
func (m *MCPIntegration) AddServer(config ServerConfig) error {
	if err := m.serverManager.AddServer(config); err != nil {
		return err
	}

	// Get the client for tool registration
	client, err := m.serverManager.GetServer(config.ID)
	if err != nil {
		return fmt.Errorf("failed to get client for server %s: %w", config.ID, err)
	}

	// Add client to tool adapter
	if adapter, ok := m.toolAdapter.(*DefaultMCPToolAdapter); ok {
		adapter.AddServer(config.ID, client)
	}

	return nil
}

// RemoveServer removes an MCP server from the integration
func (m *MCPIntegration) RemoveServer(serverID string) error {
	// Remove tools from AgentFlow registry first
	if err := m.unregisterServerTools(serverID); err != nil {
		return fmt.Errorf("failed to unregister tools for server %s: %w", serverID, err)
	}

	// Remove from tool adapter
	if adapter, ok := m.toolAdapter.(*DefaultMCPToolAdapter); ok {
		adapter.RemoveServer(serverID)
	}

	// Remove from server manager
	return m.serverManager.RemoveServer(serverID)
}

// RefreshTools refreshes tools from all servers and updates the AgentFlow tool registry
func (m *MCPIntegration) RefreshTools(ctx context.Context) error {
	servers := m.serverManager.ListServers()

	for _, server := range servers {
		if err := m.RefreshServerTools(ctx, server.ID); err != nil {
			// Log error but continue with other servers
			fmt.Printf("Failed to refresh tools for server %s: %v\n", server.ID, err)
		}
	}

	return nil
}

// RefreshServerTools refreshes tools from a specific server
func (m *MCPIntegration) RefreshServerTools(ctx context.Context, serverID string) error {
	// Refresh tools in adapter
	if err := m.toolAdapter.RefreshTools(ctx, serverID); err != nil {
		return err
	}

	// Get refreshed tools
	tools := m.toolAdapter.GetMCPTools(serverID)

	// Register tools with AgentFlow registry
	for _, toolMeta := range tools {
		if err := m.registerToolWithAgentFlow(serverID, toolMeta); err != nil {
			fmt.Printf("Failed to register tool %s from server %s: %v\n", toolMeta.Name, serverID, err)
		}
	}

	return nil
}

// GetServerTools returns all tools for a specific server
func (m *MCPIntegration) GetServerTools(serverID string) []ToolMetadata {
	return m.toolAdapter.GetMCPTools(serverID)
}

// GetAllTools returns all MCP tools from all servers
func (m *MCPIntegration) GetAllTools() map[string][]ToolMetadata {
	servers := m.serverManager.ListServers()
	allTools := make(map[string][]ToolMetadata)

	for _, server := range servers {
		tools := m.toolAdapter.GetMCPTools(server.ID)
		if len(tools) > 0 {
			allTools[server.ID] = tools
		}
	}

	return allTools
}

// ExecuteTool executes an MCP tool directly
func (m *MCPIntegration) ExecuteTool(ctx context.Context, serverID, toolName string, args map[string]interface{}) (*ToolCallResult, error) {
	return m.toolAdapter.ExecuteTool(ctx, serverID, toolName, args)
}

// GetHealthStatus returns health status for all servers
func (m *MCPIntegration) GetHealthStatus(ctx context.Context) map[string]HealthStatus {
	return m.serverManager.HealthCheck(ctx)
}

// ListServers returns information about all managed servers
func (m *MCPIntegration) ListServers() []ServerInfo {
	return m.serverManager.ListServers()
}

// Shutdown gracefully shuts down the MCP integration
func (m *MCPIntegration) Shutdown(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Stop auto-discovery
	if err := m.serverManager.StopAutoDiscovery(); err != nil {
		fmt.Printf("Error stopping auto-discovery: %v\n", err)
	}

	// Stop health checking
	if manager, ok := m.serverManager.(*DefaultMCPServerManager); ok {
		manager.StopHealthChecking()
	}

	// Disconnect all servers
	servers := m.serverManager.ListServers()
	for _, server := range servers {
		if err := m.serverManager.RemoveServer(server.ID); err != nil {
			fmt.Printf("Error removing server %s: %v\n", server.ID, err)
		}
	}

	return nil
}

// RegisterWithAgentFlow registers all MCP tools with an external AgentFlow tool registry
func (m *MCPIntegration) RegisterWithAgentFlow(registry *tools.ToolRegistry) error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	ctx := context.Background()

	// Wait for all servers to be connected before proceeding
	m.mu.RUnlock() // Unlock to allow status check
	if err := m.waitForServersConnected(ctx); err != nil {
		m.mu.RLock() // Re-acquire read lock before returning
		return fmt.Errorf("failed to wait for servers to connect: %w", err)
	}
	m.mu.RLock() // Re-acquire read lock

	// Get all tools from all servers
	allTools := m.GetAllTools()

	// If no tools are found, try refreshing them first
	if len(allTools) == 0 {
		m.mu.RUnlock() // Unlock to allow RefreshTools to acquire the lock
		if err := m.RefreshTools(ctx); err != nil {
			m.mu.RLock() // Re-acquire read lock before returning
			return fmt.Errorf("failed to refresh tools: %w", err)
		}
		m.mu.RLock() // Re-acquire read lock
		allTools = m.GetAllTools()
	}

	for serverID, tools := range allTools {
		for _, toolMeta := range tools {
			// Convert MCP tool to AgentFlow FunctionTool
			mcpTool, err := m.toolAdapter.ConvertToMCPFunctionTool(serverID, toolMeta)
			if err != nil {
				return fmt.Errorf("failed to convert MCP tool %s from server %s: %w", toolMeta.Name, serverID, err)
			}

			// Create a unique name for the tool to avoid conflicts
			toolName := fmt.Sprintf("mcp_%s_%s", serverID, toolMeta.Name)

			// Create a wrapper that implements FunctionTool
			wrapper := &agentFlowToolWrapper{
				mcpTool:      mcpTool,
				originalName: toolMeta.Name,
				uniqueName:   toolName,
			}

			// Register with the external AgentFlow tool registry
			if err := registry.Register(wrapper); err != nil {
				return fmt.Errorf("failed to register MCP tool %s: %w", toolName, err)
			}
		}
	}

	return nil
}

// Private methods

func (m *MCPIntegration) handleServerEvent(event ServerEvent) {
	switch event.Type {
	case "server_connected":
		// Refresh tools when a server connects
		ctx := context.Background()
		if err := m.RefreshServerTools(ctx, event.ServerID); err != nil {
			fmt.Printf("Failed to refresh tools after server connection: %v\n", err)
		}
	case "server_disconnected":
		// Unregister tools when a server disconnects
		if err := m.unregisterServerTools(event.ServerID); err != nil {
			fmt.Printf("Failed to unregister tools after server disconnection: %v\n", err)
		}
	case "tools_updated":
		// Refresh tools when server reports tool updates
		ctx := context.Background()
		if err := m.RefreshServerTools(ctx, event.ServerID); err != nil {
			fmt.Printf("Failed to refresh tools after tool update: %v\n", err)
		}
	}
}

func (m *MCPIntegration) registerToolWithAgentFlow(serverID string, toolMeta ToolMetadata) error {
	// Convert MCP tool to AgentFlow FunctionTool
	mcpTool, err := m.toolAdapter.ConvertToMCPFunctionTool(serverID, toolMeta)
	if err != nil {
		return err
	}

	// Create a unique name for the tool to avoid conflicts
	toolName := fmt.Sprintf("mcp_%s_%s", serverID, toolMeta.Name)

	// Create a wrapper that implements FunctionTool
	wrapper := &agentFlowToolWrapper{
		mcpTool:      mcpTool,
		originalName: toolMeta.Name,
		uniqueName:   toolName,
	}

	// Register with AgentFlow tool registry
	return m.toolRegistry.Register(wrapper)
}

func (m *MCPIntegration) unregisterServerTools(serverID string) error {
	// This would require extending AgentFlow's ToolRegistry to support unregistration
	// For now, we just log that tools should be unregistered
	tools := m.toolAdapter.GetMCPTools(serverID)
	for _, tool := range tools {
		fmt.Printf("Should unregister tool: mcp_%s_%s\n", serverID, tool.Name)
	}
	return nil
}

// agentFlowToolWrapper wraps MCPFunctionTool to implement AgentFlow's FunctionTool interface
type agentFlowToolWrapper struct {
	mcpTool      MCPFunctionTool
	originalName string
	uniqueName   string
}

// Name implements tools.FunctionTool
func (w *agentFlowToolWrapper) Name() string {
	return w.uniqueName
}

// Call implements tools.FunctionTool
func (w *agentFlowToolWrapper) Call(ctx context.Context, args map[string]any) (map[string]any, error) {
	return w.mcpTool.Call(ctx, args)
}

// waitForServersConnected waits for all servers to reach the connected status
func (m *MCPIntegration) waitForServersConnected(ctx context.Context) error {
	// Set a reasonable timeout for server connections
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	ticker := time.NewTicker(50 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("timeout waiting for servers to connect: %w", ctx.Err())
		case <-ticker.C:
			allConnected := true
			healthStatus := m.serverManager.HealthCheck(ctx) // Check if all servers are connected
			for _, status := range healthStatus {
				if status.Status != StatusConnected {
					allConnected = false
					break
				}
			}

			// If we have no servers, consider it "all connected"
			if len(healthStatus) == 0 {
				return nil
			}

			if allConnected {
				return nil
			}
		}
	}
}

// Helper function to register custom MCP client implementations
func RegisterCustomMCPClient(clientType string, creator ClientCreator) {
	RegisterGlobalClient(clientType, creator)
}

// Helper function to get supported MCP client types
func GetSupportedMCPClients() []string {
	return GetSupportedClients()
}
