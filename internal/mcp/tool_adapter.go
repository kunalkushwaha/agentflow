package mcp

import (
	"context"
	"fmt"
	"sync"

	"github.com/kunalkushwaha/agentflow/internal/tools"
)

// MCPToolWrapper implements both MCPFunctionTool and AgentFlow's FunctionTool interface
// This allows MCP tools to be seamlessly integrated with AgentFlow's ToolRegistry
type MCPToolWrapper struct {
	name        string
	description string
	schema      map[string]interface{}
	serverID    string
	metadata    map[string]interface{}
	client      MCPClient
	toolMeta    ToolMetadata
}

// Ensure MCPToolWrapper implements both interfaces
var _ MCPFunctionTool = (*MCPToolWrapper)(nil)
var _ tools.FunctionTool = (*MCPToolWrapper)(nil)

// NewMCPToolWrapper creates a new MCP tool wrapper
func NewMCPToolWrapper(client MCPClient, serverID string, toolMeta ToolMetadata) *MCPToolWrapper {
	// Convert tags to metadata map
	metadata := make(map[string]interface{})
	if len(toolMeta.Tags) > 0 {
		metadata["tags"] = toolMeta.Tags
	}

	return &MCPToolWrapper{
		name:        toolMeta.Name,
		description: toolMeta.Description,
		schema:      toolMeta.Schema,
		serverID:    serverID,
		metadata:    metadata,
		client:      client,
		toolMeta:    toolMeta,
	}
}

// FunctionTool interface implementation

// Name returns the tool name (implements AgentFlow's FunctionTool interface)
func (w *MCPToolWrapper) Name() string {
	return w.name
}

// Call executes the tool via MCP client (implements AgentFlow's FunctionTool interface)
func (w *MCPToolWrapper) Call(ctx context.Context, args map[string]any) (map[string]any, error) {
	// Convert args to the format expected by MCP
	mcpArgs := make(map[string]interface{})
	for k, v := range args {
		mcpArgs[k] = v
	}

	// Create MCP tool call request
	request := ToolCallRequest{
		Name:      w.name,
		Arguments: mcpArgs,
	}

	// Execute the tool via MCP client
	result, err := w.client.CallTool(ctx, request)
	if err != nil {
		return nil, fmt.Errorf("MCP tool call failed: %w", err)
	}

	if result.IsError {
		return nil, fmt.Errorf("MCP tool returned error: %v", result.Content)
	}

	// Convert result to AgentFlow format
	response := make(map[string]any)

	// Handle different content types
	if len(result.Content) > 0 {
		// If there's only one content block and it's text, return it directly
		if len(result.Content) == 1 && result.Content[0].Type == "text" {
			response["result"] = result.Content[0].Content
		} else {
			// Return all content blocks for complex responses
			response["content"] = result.Content
		}
	}

	// Add metadata
	response["tool_name"] = w.name
	response["server_id"] = w.serverID

	return response, nil
}

// MCPFunctionTool interface implementation (additional methods)

// Description returns the tool description
func (w *MCPToolWrapper) Description() string {
	return w.description
}

// Schema returns the tool's parameter schema
func (w *MCPToolWrapper) Schema() map[string]interface{} {
	return w.schema
}

// ServerID returns the MCP server ID
func (w *MCPToolWrapper) ServerID() string {
	return w.serverID
}

// Metadata returns additional tool metadata
func (w *MCPToolWrapper) Metadata() map[string]interface{} {
	meta := make(map[string]interface{})
	meta["description"] = w.description
	meta["server_id"] = w.serverID
	meta["schema"] = w.schema

	// Add any tags or additional metadata
	for k, v := range w.metadata {
		meta[k] = v
	}

	return meta
}

// DefaultMCPToolAdapter provides a default implementation of MCPToolAdapter
type DefaultMCPToolAdapter struct {
	servers map[string]MCPClient
	tools   map[string]map[string]*MCPToolWrapper // serverID -> toolName -> wrapper
	mu      sync.RWMutex
}

// NewDefaultMCPToolAdapter creates a new default MCP tool adapter
func NewDefaultMCPToolAdapter() *DefaultMCPToolAdapter {
	return &DefaultMCPToolAdapter{
		servers: make(map[string]MCPClient),
		tools:   make(map[string]map[string]*MCPToolWrapper),
	}
}

// RegisterMCPTool registers an MCP tool with the adapter
func (a *DefaultMCPToolAdapter) RegisterMCPTool(serverID string, tool ToolMetadata) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	client, exists := a.servers[serverID]
	if !exists {
		return fmt.Errorf("server %s not found", serverID)
	}

	if a.tools[serverID] == nil {
		a.tools[serverID] = make(map[string]*MCPToolWrapper)
	}

	wrapper := NewMCPToolWrapper(client, serverID, tool)
	a.tools[serverID][tool.Name] = wrapper

	return nil
}

// UnregisterMCPTool removes an MCP tool from the adapter
func (a *DefaultMCPToolAdapter) UnregisterMCPTool(serverID, toolName string) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	if serverTools, exists := a.tools[serverID]; exists {
		delete(serverTools, toolName)
		return nil
	}

	return fmt.Errorf("tool %s not found on server %s", toolName, serverID)
}

// RefreshTools refreshes tools from the specified server
func (a *DefaultMCPToolAdapter) RefreshTools(ctx context.Context, serverID string) error {
	a.mu.Lock()
	client, exists := a.servers[serverID]
	a.mu.Unlock()

	if !exists {
		return fmt.Errorf("server %s not found", serverID)
	}

	// Get fresh tool list from server
	tools, err := client.ListTools(ctx)
	if err != nil {
		return fmt.Errorf("failed to list tools from server %s: %w", serverID, err)
	}

	a.mu.Lock()
	defer a.mu.Unlock()

	// Clear existing tools for this server
	a.tools[serverID] = make(map[string]*MCPToolWrapper)

	// Register all tools
	for _, tool := range tools {
		wrapper := NewMCPToolWrapper(client, serverID, tool)
		a.tools[serverID][tool.Name] = wrapper
	}

	return nil
}

// GetMCPTools returns all tools for a server
func (a *DefaultMCPToolAdapter) GetMCPTools(serverID string) []ToolMetadata {
	a.mu.RLock()
	defer a.mu.RUnlock()

	serverTools, exists := a.tools[serverID]
	if !exists {
		return nil
	}

	tools := make([]ToolMetadata, 0, len(serverTools))
	for _, wrapper := range serverTools {
		tools = append(tools, wrapper.toolMeta)
	}

	return tools
}

// ExecuteTool executes an MCP tool
func (a *DefaultMCPToolAdapter) ExecuteTool(ctx context.Context, serverID, toolName string, args map[string]interface{}) (*ToolCallResult, error) {
	a.mu.RLock()
	serverTools, exists := a.tools[serverID]
	if !exists {
		a.mu.RUnlock()
		return nil, fmt.Errorf("server %s not found", serverID)
	}

	wrapper, exists := serverTools[toolName]
	a.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("tool %s not found on server %s", toolName, serverID)
	}

	// Convert args and call the wrapper
	convertedArgs := make(map[string]any)
	for k, v := range args {
		convertedArgs[k] = v
	}

	result, err := wrapper.Call(ctx, convertedArgs)
	if err != nil {
		return &ToolCallResult{
			Content: []ContentBlock{{Type: "text", Content: err.Error()}},
			IsError: true,
		}, nil
	}

	// Convert result back to ToolCallResult format
	content := []ContentBlock{{Type: "text", Content: fmt.Sprintf("%v", result)}}

	return &ToolCallResult{
		Content: content,
		IsError: false,
	}, nil
}

// ConvertToMCPFunctionTool converts tool metadata to an MCPFunctionTool
func (a *DefaultMCPToolAdapter) ConvertToMCPFunctionTool(serverID string, tool ToolMetadata) (MCPFunctionTool, error) {
	a.mu.RLock()
	client, exists := a.servers[serverID]
	a.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("server %s not found", serverID)
	}

	return NewMCPToolWrapper(client, serverID, tool), nil
}

// AddServer adds an MCP server client to the adapter
func (a *DefaultMCPToolAdapter) AddServer(serverID string, client MCPClient) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.servers[serverID] = client
}

// RemoveServer removes an MCP server from the adapter
func (a *DefaultMCPToolAdapter) RemoveServer(serverID string) {
	a.mu.Lock()
	defer a.mu.Unlock()
	delete(a.servers, serverID)
	delete(a.tools, serverID)
}
