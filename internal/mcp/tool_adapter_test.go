package mcp

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMCPToolWrapper_Implementation(t *testing.T) {
	// Create a mock client with proper configuration
	config := map[string]interface{}{
		"id":      "test-server",
		"timeout": 5000,
	}
	client, err := NewMockClient(config)
	if err != nil {
		t.Fatalf("Failed to create mock client: %v", err)
	}

	// Connect the client
	ctx := context.Background()
	serverConfig := ServerConfig{
		ID:      "test-server",
		Timeout: 5000,
		Connection: ConnectionConfig{
			Transport: "stdio",
			Command:   []string{"mock"},
		},
	}
	err = client.Connect(ctx, serverConfig)
	if err != nil {
		t.Fatalf("Failed to connect client: %v", err)
	}

	// Get the actual tool metadata from the mock client
	tools, err := client.ListTools(ctx)
	if err != nil {
		t.Fatalf("Failed to list tools: %v", err)
	}
	if len(tools) == 0 {
		t.Fatal("Mock client has no tools")
	}

	// Use the first tool from the mock client
	metadata := tools[0]

	// Create tool wrapper
	wrapper := NewMCPToolWrapper(client, "test-server", metadata)
	if wrapper == nil {
		t.Fatal("Wrapper is nil")
	}

	// Test Name method - wrapper returns the raw tool name
	if wrapper.Name() != metadata.Name {
		t.Errorf("Expected name '%s', got '%s'", metadata.Name, wrapper.Name())
	}

	// Test Call method
	result, err := wrapper.Call(ctx, map[string]any{
		"input": "test",
	})
	if err != nil {
		t.Fatalf("Failed to call tool: %v", err)
	}
	if result == nil {
		t.Error("Result is nil")
	}
}

func TestDefaultMCPToolAdapter_Basic(t *testing.T) {
	adapter := NewDefaultMCPToolAdapter()
	assert.NotNil(t, adapter)

	// Create mock client with proper configuration and connect it
	config := map[string]interface{}{
		"id":      "test-server",
		"timeout": 5000,
	}
	client, err := NewMockClient(config)
	assert.NoError(t, err)

	// Connect the client
	ctx := context.Background()
	serverConfig := ServerConfig{
		ID:      "test-server",
		Timeout: 5000,
		Connection: ConnectionConfig{
			Transport: "stdio",
			Command:   []string{"mock"},
		},
	}
	err = client.Connect(ctx, serverConfig)
	assert.NoError(t, err)

	adapter.AddServer("test-server", client)

	// Test getting MCP tools from a specific server
	err = adapter.RefreshTools(ctx, "test-server")
	assert.NoError(t, err)

	tools := adapter.GetMCPTools("test-server")
	assert.NotEmpty(t, tools)

	// Test executing a tool (need server ID, tool name, and args)
	result, err := adapter.ExecuteTool(ctx, "test-server", "mock_tool", map[string]interface{}{
		"input": "test",
	})
	assert.NoError(t, err)
	assert.NotNil(t, result)

	// Test removing server
	adapter.RemoveServer("test-server")
	tools = adapter.GetMCPTools("test-server")
	assert.Empty(t, tools)
}
