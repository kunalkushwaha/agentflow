package core

import (
	"context"
	"fmt"
	"time"

	"github.com/kunalkushwaha/agentflow/internal/mcp"
	"github.com/kunalkushwaha/agentflow/internal/tools"
)

// MCPManager provides a public API for MCP (Model Context Protocol) integration
// This allows external applications to use MCP functionality without importing internal packages
type MCPManager struct {
	integration *mcp.MCPIntegration
	registry    *tools.ToolRegistry
}

// MCPConfig represents the public configuration for MCP integration
type MCPConfig struct {
	// ClientType specifies the MCP client implementation to use ("mock", "mark3labs", "custom")
	ClientType string `json:"client_type"`
	// HealthCheck enables health monitoring of MCP servers
	HealthCheck bool `json:"health_check"`
	// AutoDiscovery enables automatic discovery of MCP servers
	AutoDiscovery bool `json:"auto_discovery"`
}

// MCPServerConfig represents configuration for a single MCP server
type MCPServerConfig struct {
	// ID is the unique identifier for the server
	ID string `json:"id"`
	// Name is a human-readable name for the server
	Name string `json:"name"`
	// Type specifies the server type ("mock", "stdio", "http", "websocket")
	Type string `json:"type"`
	// ClientType specifies the client implementation to use
	ClientType string `json:"client_type"`
	// Enabled indicates whether the server should be active
	Enabled bool `json:"enabled"`
	// Connection contains connection-specific configuration
	Connection MCPConnectionConfig `json:"connection"`
	// Timeout in milliseconds for server operations
	Timeout int `json:"timeout"`
}

// MCPConnectionConfig represents connection configuration for MCP servers
type MCPConnectionConfig struct {
	// Transport type ("stdio", "http", "websocket", "sse")
	Transport string `json:"transport"`
	// Command for stdio transport (e.g., ["python", "server.py"])
	Command []string `json:"command"`
	// Args for additional command arguments
	Args []string `json:"args"`
	// Env for environment variables
	Env map[string]string `json:"env"`
	// Endpoint for network-based transports
	Endpoint string `json:"endpoint"`
	// Headers for HTTP-based transports
	Headers map[string]string `json:"headers"`
	// WorkingDirectory for stdio processes
	WorkingDirectory string `json:"working_directory"`
}

// MCPToolInfo represents information about an available MCP tool
type MCPToolInfo struct {
	// Name is the tool identifier
	Name string `json:"name"`
	// Description explains what the tool does
	Description string `json:"description"`
	// ServerID identifies which MCP server provides this tool
	ServerID string `json:"server_id"`
	// Schema describes the tool's input parameters
	Schema map[string]interface{} `json:"schema"`
	// UniqueName is the name registered in the tool registry (includes server prefix)
	UniqueName string `json:"unique_name"`
}

// MCPServerInfo represents information about an MCP server
type MCPServerInfo struct {
	// ID is the server identifier
	ID string `json:"id"`
	// Name is the server name
	Name string `json:"name"`
	// Status indicates the server's current state
	Status string `json:"status"`
	// ToolCount is the number of tools available from this server
	ToolCount int `json:"tool_count"`
	// LastSeen is when the server was last contacted
	LastSeen *time.Time `json:"last_seen"`
}

// MCPToolResult represents the result of executing an MCP tool
type MCPToolResult struct {
	// Success indicates whether the tool call succeeded
	Success bool `json:"success"`
	// Result contains the tool's output data
	Result map[string]interface{} `json:"result"`
	// Error contains error information if the call failed
	Error string `json:"error,omitempty"`
	// ServerID identifies which server executed the tool
	ServerID string `json:"server_id"`
	// ToolName is the name of the executed tool
	ToolName string `json:"tool_name"`
}

// NewMCPManager creates a new MCP manager with the given configuration
func NewMCPManager(config MCPConfig) (*MCPManager, error) {
	// Create a tool registry for MCP tools
	registry := tools.NewToolRegistry()

	// Convert public config to internal config
	internalConfig := mcp.MCPIntegrationConfig{
		ClientType:    config.ClientType,
		HealthCheck:   config.HealthCheck,
		AutoDiscovery: config.AutoDiscovery,
	}

	// Create MCP integration with the registry
	integration, err := mcp.NewMCPIntegrationWithRegistry(registry, internalConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create MCP integration: %w", err)
	}

	return &MCPManager{
		integration: integration,
		registry:    registry,
	}, nil
}

// NewMCPManagerFromConfig creates a new MCP manager from a TOML configuration file
func NewMCPManagerFromConfig(configPath string) (*MCPManager, error) {
	// Create a tool registry for MCP tools
	registry := tools.NewToolRegistry()

	// Create MCP integration from config file
	integration, err := mcp.NewMCPIntegrationFromConfig(registry, configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create MCP integration from config: %w", err)
	}

	return &MCPManager{
		integration: integration,
		registry:    registry,
	}, nil
}

// AddServer adds an MCP server to the manager
func (m *MCPManager) AddServer(config MCPServerConfig) error {
	// Convert public config to internal config
	internalConfig := mcp.ServerConfig{
		ID:         config.ID,
		Name:       config.Name,
		Type:       config.Type,
		ClientType: config.ClientType,
		Enabled:    config.Enabled,
		Connection: mcp.ConnectionConfig{
			Transport:   config.Connection.Transport,
			Command:     config.Connection.Command,
			Address:     config.Connection.Endpoint,
			Environment: config.Connection.Env,
			Headers:     config.Connection.Headers,
		},
		Timeout: time.Duration(config.Timeout) * time.Millisecond,
	}

	return m.integration.AddServer(internalConfig)
}

// RemoveServer removes an MCP server from the manager
func (m *MCPManager) RemoveServer(serverID string) error {
	return m.integration.RemoveServer(serverID)
}

// ListServers returns information about all managed MCP servers
func (m *MCPManager) ListServers() []MCPServerInfo {
	servers := m.integration.ListServers()
	result := make([]MCPServerInfo, len(servers))

	for i, server := range servers {
		// Get tool count for this server
		tools := m.integration.GetServerTools(server.ID)
		result[i] = MCPServerInfo{
			ID:        server.ID,
			Name:      server.Name,
			Status:    string(server.Status),
			ToolCount: len(tools),
			LastSeen:  &server.LastSeen,
		}
	}

	return result
}

// ListTools returns information about all available MCP tools
func (m *MCPManager) ListTools() []MCPToolInfo {
	allTools := m.integration.GetAllTools()
	var result []MCPToolInfo

	for serverID, tools := range allTools {
		for _, tool := range tools {
			uniqueName := fmt.Sprintf("mcp_%s_%s", serverID, tool.Name)
			result = append(result, MCPToolInfo{
				Name:        tool.Name,
				Description: tool.Description,
				ServerID:    serverID,
				Schema:      tool.Schema,
				UniqueName:  uniqueName,
			})
		}
	}

	return result
}

// RefreshTools refreshes the tool list from all MCP servers
func (m *MCPManager) RefreshTools(ctx context.Context) error {
	return m.integration.RefreshTools(ctx)
}

// ExecuteTool executes an MCP tool directly by server ID and tool name
func (m *MCPManager) ExecuteTool(ctx context.Context, serverID, toolName string, args map[string]interface{}) (*MCPToolResult, error) {
	result, err := m.integration.ExecuteTool(ctx, serverID, toolName, args)
	if err != nil {
		return &MCPToolResult{
			Success:  false,
			Error:    err.Error(),
			ServerID: serverID,
			ToolName: toolName,
		}, nil
	}

	// Convert internal result to public result
	resultData := make(map[string]interface{})
	if result != nil && len(result.Content) > 0 {
		// Extract content from the result
		if len(result.Content) == 1 && result.Content[0].Type == "text" {
			resultData["result"] = result.Content[0].Content
		} else {
			resultData["content"] = result.Content
		}
		resultData["server_id"] = serverID
		resultData["tool_name"] = toolName
	}

	return &MCPToolResult{
		Success:  !result.IsError,
		Result:   resultData,
		ServerID: serverID,
		ToolName: toolName,
	}, nil
}

// ExecuteToolByUniqueName executes an MCP tool using its unique registered name
func (m *MCPManager) ExecuteToolByUniqueName(ctx context.Context, uniqueName string, args map[string]interface{}) (*MCPToolResult, error) {
	// Convert args to the expected format
	toolArgs := make(map[string]any)
	for k, v := range args {
		toolArgs[k] = v
	}

	// Execute through the tool registry
	result, err := m.registry.CallTool(ctx, uniqueName, toolArgs)
	if err != nil {
		return &MCPToolResult{
			Success: false,
			Error:   err.Error(),
		}, nil
	}

	return &MCPToolResult{
		Success: true,
		Result:  result,
	}, nil
}

// GetToolRegistry returns the internal tool registry for advanced usage
// This allows external applications to access the registry directly if needed
func (m *MCPManager) GetToolRegistry() *tools.ToolRegistry {
	return m.registry
}

// GetTools returns a public ToolRegistry interface for accessing MCP tools
// This is the recommended way for agents to access MCP tools
func (m *MCPManager) GetTools() ToolRegistry {
	return NewToolRegistryAdapter(m.registry)
}

// RegisterWithToolRegistry registers all MCP tools with an external tool registry
// This is useful when you want to integrate MCP tools with an existing tool system
func (m *MCPManager) RegisterWithToolRegistry(registry *tools.ToolRegistry) error {
	return m.integration.RegisterWithAgentFlow(registry)
}

// HealthCheck returns the health status of all MCP servers
func (m *MCPManager) HealthCheck(ctx context.Context) map[string]string {
	healthStatus := m.integration.GetHealthStatus(ctx)
	result := make(map[string]string)

	for serverID, status := range healthStatus {
		result[serverID] = string(status.Status)
	}

	return result
}

// Shutdown gracefully shuts down the MCP manager and all connections
func (m *MCPManager) Shutdown(ctx context.Context) error {
	return m.integration.Shutdown(ctx)
}

// CreateMockServerConfig creates a sample configuration for a mock MCP server
// This is useful for testing and development
func CreateMockServerConfig(serverID, serverName string) MCPServerConfig {
	return MCPServerConfig{
		ID:         serverID,
		Name:       serverName,
		Type:       "mock",
		ClientType: "mock",
		Enabled:    true,
		Connection: MCPConnectionConfig{
			Transport: "stdio",
		},
		Timeout: 5000,
	}
}

// GetSupportedMCPClients returns a list of supported MCP client types
func GetSupportedMCPClients() []string {
	return mcp.GetSupportedMCPClients()
}

// RegisterCustomMCPClient registers a custom MCP client implementation
// This allows extending the MCP manager with custom client types
func RegisterCustomMCPClient(clientType string, creator func(config map[string]interface{}) (interface{}, error)) {
	// This would need to be implemented based on the internal factory pattern
	// For now, we provide a placeholder that can be extended
	fmt.Printf("Custom MCP client registration for '%s' - implementation needed\n", clientType)
}
