package mcp

import (
	"context"
	"testing"
)

func TestMark3LabsClientWrapper_Creation(t *testing.T) {
	factory := NewClientFactory()

	// Test that mark3labs is in supported clients
	clients := factory.SupportedClients()
	found := false
	for _, client := range clients {
		if client == "mark3labs" {
			found = true
			break
		}
	}
	if !found {
		t.Error("mark3labs client not found in supported clients")
	}

	// Test creating mark3labs client
	config := map[string]interface{}{
		"id":   "test-mark3labs",
		"name": "Test Mark3Labs Server",
		"connection": map[string]interface{}{
			"transport": "stdio",
			"command":   []interface{}{"echo", "test"},
		},
	}

	client, err := factory.CreateClient("mark3labs", config)
	if err != nil {
		t.Fatalf("Failed to create mark3labs client: %v", err)
	}

	if client == nil {
		t.Fatal("Created client is nil")
	}

	// Verify it's the correct type
	wrapper, ok := client.(*Mark3LabsClientWrapper)
	if !ok {
		t.Fatalf("Expected *Mark3LabsClientWrapper, got %T", client)
	}

	// Test basic interface methods (without actually connecting)
	if wrapper.IsConnected() {
		t.Error("Client should not be connected initially")
	}
}

func TestMark3LabsClientWrapper_ConfigConversion(t *testing.T) {
	// Test config conversion
	config := map[string]interface{}{
		"id":          "test-server",
		"name":        "Test Server",
		"type":        "stdio",
		"client_type": "mark3labs",
		"enabled":     true,
		"connection": map[string]interface{}{
			"transport":   "stdio",
			"command":     []interface{}{"python", "server.py"},
			"environment": map[string]interface{}{"PATH": "/usr/bin"},
			"address":     "localhost:8080",
			"headers":     map[string]interface{}{"Authorization": "Bearer token"},
		},
	}

	serverConfig, err := convertMapToServerConfig(config)
	if err != nil {
		t.Fatalf("Failed to convert config: %v", err)
	}

	// Verify conversion
	if serverConfig.ID != "test-server" {
		t.Errorf("Expected ID 'test-server', got '%s'", serverConfig.ID)
	}
	if serverConfig.Name != "Test Server" {
		t.Errorf("Expected Name 'Test Server', got '%s'", serverConfig.Name)
	}
	if serverConfig.Connection.Transport != "stdio" {
		t.Errorf("Expected Transport 'stdio', got '%s'", serverConfig.Connection.Transport)
	}
	if len(serverConfig.Connection.Command) != 2 || serverConfig.Connection.Command[0] != "python" {
		t.Errorf("Expected Command ['python', 'server.py'], got %v", serverConfig.Connection.Command)
	}
}

func TestMark3LabsClientWrapper_InterfaceMethods(t *testing.T) {
	// Create a mark3labs client wrapper
	config := map[string]interface{}{
		"id":   "test-interface",
		"name": "Test Interface",
		"connection": map[string]interface{}{
			"transport": "stdio",
			"command":   []interface{}{"echo", "test"},
		},
	}

	client, err := NewMark3LabsBridge(config)
	if err != nil {
		t.Fatalf("Failed to create mark3labs bridge: %v", err)
	}

	wrapper, ok := client.(*Mark3LabsClientWrapper)
	if !ok {
		t.Fatalf("Expected *Mark3LabsClientWrapper, got %T", client)
	}

	ctx := context.Background()

	// Test methods that should work without connection
	if wrapper.IsConnected() {
		t.Error("Should not be connected initially")
	}

	// Test ping on disconnected client (should return error)
	err = wrapper.Ping(ctx)
	if err == nil {
		t.Error("Ping should fail when not connected")
	}

	// Test GetServerInfo on disconnected client (should return error)
	_, err = wrapper.GetServerInfo(ctx)
	if err == nil {
		t.Error("GetServerInfo should fail when not connected")
	}

	// Test methods that return empty results for unimplemented features
	resources, err := wrapper.ListResources(ctx)
	if err != nil {
		t.Errorf("ListResources should not error: %v", err)
	}
	if len(resources) != 0 {
		t.Error("ListResources should return empty list")
	}

	prompts, err := wrapper.ListPrompts(ctx)
	if err != nil {
		t.Errorf("ListPrompts should not error: %v", err)
	}
	if len(prompts) != 0 {
		t.Error("ListPrompts should return empty list")
	}

	// Test error methods for unimplemented features
	_, err = wrapper.ReadResource(ctx, "test://resource")
	if err == nil {
		t.Error("ReadResource should return error for unimplemented feature")
	}

	_, err = wrapper.GetPrompt(ctx, "test", nil)
	if err == nil {
		t.Error("GetPrompt should return error for unimplemented feature")
	}

	// Test notification and error handlers (should not panic)
	wrapper.SetNotificationHandler(func(notification MCPNotification) {})
	wrapper.SetErrorHandler(func(err MCPError) {})
}

func TestMark3LabsClientWrapper_DefaultClient(t *testing.T) {
	factory := NewClientFactory()
	defaultClient := factory.DefaultClient()

	if defaultClient != "mark3labs" {
		t.Errorf("Expected default client to be 'mark3labs', got '%s'", defaultClient)
	}
}
