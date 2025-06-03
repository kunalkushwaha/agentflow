package mcp

import (
	"context"
	"testing"
	"time"

	"github.com/kunalkushwaha/agentflow/internal/tools"
	"github.com/stretchr/testify/assert"
)

func TestMCPIntegration_NewIntegration(t *testing.T) {
	config := MCPIntegrationConfig{
		ClientType:    "mock",
		HealthCheck:   true,
		AutoDiscovery: false,
	}

	integration, err := NewMCPIntegration(config)
	assert.NoError(t, err)
	assert.NotNil(t, integration)

	// Test that components are properly initialized
	assert.NotNil(t, integration.serverManager)
	assert.NotNil(t, integration.toolAdapter)
	assert.NotNil(t, integration.clientFactory)
}

func TestMCPIntegration_AddAndRemoveServer(t *testing.T) {
	config := MCPIntegrationConfig{
		ClientType:    "mock",
		HealthCheck:   false,
		AutoDiscovery: false,
	}

	integration, err := NewMCPIntegration(config)
	assert.NoError(t, err)

	// Test adding a server
	serverConfig := ServerConfig{
		ID:         "test-server",
		Name:       "Test Server",
		Type:       "mock",
		ClientType: "mock",
		Enabled:    true,
		Connection: ConnectionConfig{
			Transport: "stdio",
		},
	}

	err = integration.AddServer(serverConfig)
	assert.NoError(t, err)

	// Verify server was added
	servers := integration.ListServers()
	assert.Len(t, servers, 1)
	assert.Equal(t, "test-server", servers[0].ID)

	// Test removing the server
	err = integration.RemoveServer("test-server")
	assert.NoError(t, err)

	// Verify server was removed
	servers = integration.ListServers()
	assert.Len(t, servers, 0)
}

func TestMCPIntegration_GetTools(t *testing.T) {
	config := MCPIntegrationConfig{
		ClientType:    "mock",
		HealthCheck:   false,
		AutoDiscovery: false,
	}

	integration, err := NewMCPIntegration(config)
	assert.NoError(t, err)

	// Add a mock server
	serverConfig := ServerConfig{
		ID:         "test-server",
		Name:       "Test Server",
		Type:       "mock",
		ClientType: "mock",
		Enabled:    true,
		Connection: ConnectionConfig{
			Transport: "stdio",
		},
	}
	err = integration.AddServer(serverConfig)
	assert.NoError(t, err)

	// Give the async connection time to complete
	time.Sleep(100 * time.Millisecond)

	// Get available tools from a specific server
	ctx := context.Background()
	err = integration.RefreshTools(ctx)
	assert.NoError(t, err)

	// Mock client should provide at least one tool from test-server
	tools := integration.toolAdapter.GetMCPTools("test-server")
	assert.NotEmpty(t, tools)

	// Verify tool has required fields
	tool := tools[0]
	assert.NotEmpty(t, tool.Name)
	assert.NotEmpty(t, tool.Description)
}

func TestMCPIntegration_ExecuteTool(t *testing.T) {
	config := MCPIntegrationConfig{
		ClientType:    "mock",
		HealthCheck:   false,
		AutoDiscovery: false,
	}

	integration, err := NewMCPIntegration(config)
	assert.NoError(t, err)

	// Add a mock server
	serverConfig := ServerConfig{
		ID:         "test-server",
		Name:       "Test Server",
		Type:       "mock",
		ClientType: "mock",
		Enabled:    true,
		Connection: ConnectionConfig{
			Transport: "stdio",
		},
	}
	err = integration.AddServer(serverConfig)
	assert.NoError(t, err)

	// Give the async connection time to complete
	time.Sleep(100 * time.Millisecond)

	// Execute mock tool (need server ID, tool name, and args)
	ctx := context.Background()
	result, err := integration.toolAdapter.ExecuteTool(ctx, "test-server", "mock_tool", map[string]interface{}{
		"input": "test input",
	})

	assert.NoError(t, err)
	assert.NotNil(t, result)
}

func TestMCPIntegration_RegisterWithAgentFlow(t *testing.T) {
	config := MCPIntegrationConfig{
		ClientType:    "mock",
		HealthCheck:   false,
		AutoDiscovery: false,
	}

	integration, err := NewMCPIntegration(config)
	assert.NoError(t, err)

	// Create a tool registry
	toolRegistry := tools.NewToolRegistry()

	// Add a mock server
	serverConfig := ServerConfig{
		ID:         "test-server",
		Name:       "Test Server",
		Type:       "mock",
		ClientType: "mock",
		Enabled:    true,
		Connection: ConnectionConfig{
			Transport: "stdio",
		},
	}

	err = integration.AddServer(serverConfig)
	assert.NoError(t, err)
	// Register MCP tools with AgentFlow
	err = integration.RegisterWithAgentFlow(toolRegistry)
	assert.NoError(t, err)

	// Verify tools were registered
	registeredToolNames := toolRegistry.List()
	assert.NotEmpty(t, registeredToolNames)

	// Verify at least one MCP tool is present
	mcpToolFound := false
	for _, toolName := range registeredToolNames {
		if toolName == "mcp_test-server_mock_tool" {
			mcpToolFound = true
			break
		}
	}

	assert.True(t, mcpToolFound)
}

func TestMCPIntegration_Shutdown(t *testing.T) {
	config := MCPIntegrationConfig{
		ClientType:    "mock",
		HealthCheck:   true,
		AutoDiscovery: false,
	}

	integration, err := NewMCPIntegration(config)
	assert.NoError(t, err)

	// Add a mock server
	serverConfig := ServerConfig{
		ID:         "test-server",
		Name:       "Test Server",
		Type:       "mock",
		ClientType: "mock",
		Enabled:    true,
		Connection: ConnectionConfig{
			Transport: "stdio",
		},
	}

	err = integration.AddServer(serverConfig)
	assert.NoError(t, err)

	// Test shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = integration.Shutdown(ctx)
	assert.NoError(t, err)
}
