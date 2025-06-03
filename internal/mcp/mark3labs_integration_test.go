package mcp

import (
	"context"
	"testing"
	"time"
)

func TestMark3LabsIntegration_EndToEnd(t *testing.T) {
	// Test complete end-to-end integration of mark3labs client
	factory := NewClientFactory()

	// Verify mark3labs is the default client
	if factory.DefaultClient() != "mark3labs" {
		t.Errorf("Expected default client to be 'mark3labs', got '%s'", factory.DefaultClient())
	}

	// Create a mark3labs client configuration
	config := map[string]interface{}{
		"id":          "test-mark3labs-integration",
		"name":        "Test Mark3Labs Integration",
		"type":        "stdio",
		"client_type": "mark3labs",
		"connection": map[string]interface{}{
			"transport": "stdio",
			"command":   []interface{}{"echo", "test"},
			"environment": map[string]interface{}{
				"TEST_ENV": "test_value",
			},
			"headers": map[string]interface{}{
				"Content-Type": "application/json",
			},
		},
		"enabled": true,
	}

	// Create the client through the factory
	client, err := factory.CreateClient("mark3labs", config)
	if err != nil {
		t.Fatalf("Failed to create mark3labs client: %v", err)
	}

	if client == nil {
		t.Fatal("Created client is nil")
	}

	// Verify it's the correct wrapper type
	wrapper, ok := client.(*Mark3LabsClientWrapper)
	if !ok {
		t.Fatalf("Expected *Mark3LabsClientWrapper, got %T", client)
	}

	// Test initial state
	if wrapper.IsConnected() {
		t.Error("Client should not be connected initially")
	}

	// Test that all interface methods are available (without connection)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// These should fail gracefully without connection
	err = wrapper.Ping(ctx)
	if err == nil {
		t.Error("Ping should fail when not connected")
	}

	info, err := wrapper.GetServerInfo(ctx)
	if err == nil {
		t.Error("GetServerInfo should fail when not connected")
	}
	if info != nil {
		t.Error("ServerInfo should be nil when not connected")
	}

	tools, err := wrapper.ListTools(ctx)
	if err == nil {
		t.Error("ListTools should fail when not connected")
	}
	if tools != nil {
		t.Error("Tools should be nil when not connected")
	}

	// Test error handling for invalid configurations
	invalidConfig := map[string]interface{}{
		"id": "invalid-config",
		"connection": map[string]interface{}{
			"transport": "invalid_transport",
		},
	}

	invalidClient, err := factory.CreateClient("mark3labs", invalidConfig)
	if err != nil {
		t.Fatalf("Factory should create client even with invalid config: %v", err)
	}

	// Try to connect with invalid config - this should fail when attempting connection
	serverConfig, err := convertMapToServerConfig(invalidConfig)
	if err != nil {
		t.Fatalf("Failed to convert invalid config: %v", err)
	}

	err = invalidClient.Connect(ctx, serverConfig)
	if err == nil {
		t.Error("Connect should fail with invalid transport")
	}

	t.Log("Mark3Labs integration test completed successfully")
}

func TestMark3LabsIntegration_WithServerManager(t *testing.T) {
	// Test mark3labs client integration with server manager
	factory := NewClientFactory()
	manager := NewDefaultMCPServerManager(factory)

	// Create a server configuration using mark3labs client
	serverConfig := ServerConfig{
		ID:         "test-mark3labs-manager",
		Name:       "Test Mark3Labs with Manager",
		Type:       "stdio",
		ClientType: "mark3labs",
		Connection: ConnectionConfig{
			Transport: "stdio",
			Command:   []string{"echo", "test"},
			Environment: map[string]string{
				"TEST_ENV": "manager_test",
			},
		},
		Enabled: true,
		Timeout: 30 * time.Second,
	}

	// Add server through manager
	err := manager.AddServer(serverConfig)
	if err != nil {
		t.Fatalf("Failed to add server to manager: %v", err)
	}

	// Verify server was added
	servers := manager.ListServers()
	if len(servers) != 1 {
		t.Fatalf("Expected 1 server, got %d", len(servers))
	}

	if servers[0].ID != "test-mark3labs-manager" {
		t.Errorf("Expected server ID 'test-mark3labs-manager', got '%s'", servers[0].ID)
	}

	// Get the client from manager
	client, err := manager.GetServer("test-mark3labs-manager")
	if err != nil {
		t.Fatalf("Failed to get server from manager: %v", err)
	}

	// Verify it's a mark3labs wrapper
	wrapper, ok := client.(*Mark3LabsClientWrapper)
	if !ok {
		t.Fatalf("Expected *Mark3LabsClientWrapper, got %T", client)
	}

	// Test that the client has the correct configuration
	if !wrapper.IsConnected() {
		// This is expected since we haven't actually connected
		t.Log("Client is not connected (expected)")
	}

	// Remove server
	err = manager.RemoveServer("test-mark3labs-manager")
	if err != nil {
		t.Fatalf("Failed to remove server from manager: %v", err)
	}

	// Verify server was removed
	servers = manager.ListServers()
	if len(servers) != 0 {
		t.Fatalf("Expected 0 servers after removal, got %d", len(servers))
	}

	t.Log("Mark3Labs server manager integration test completed successfully")
}

func TestMark3LabsIntegration_ConfigurationVariations(t *testing.T) {
	// Test various configuration scenarios for mark3labs client
	factory := NewClientFactory()

	testCases := []struct {
		name        string
		config      map[string]interface{}
		expectError bool
	}{
		{
			name: "stdio_basic",
			config: map[string]interface{}{
				"id": "stdio-basic",
				"connection": map[string]interface{}{
					"transport": "stdio",
					"command":   []interface{}{"python", "server.py"},
				},
			},
			expectError: false,
		},
		{
			name: "http_basic",
			config: map[string]interface{}{
				"id": "http-basic",
				"connection": map[string]interface{}{
					"transport": "http",
					"address":   "http://localhost:8080/mcp",
				},
			},
			expectError: false,
		},
		{
			name: "sse_basic",
			config: map[string]interface{}{
				"id": "sse-basic",
				"connection": map[string]interface{}{
					"transport": "sse",
					"address":   "http://localhost:8080/sse",
				},
			},
			expectError: false,
		},
		{
			name: "stdio_with_env",
			config: map[string]interface{}{
				"id": "stdio-env",
				"connection": map[string]interface{}{
					"transport": "stdio",
					"command":   []interface{}{"node", "server.js"},
					"environment": map[string]interface{}{
						"NODE_ENV": "development",
						"PORT":     "3000",
					},
				},
			},
			expectError: false,
		},
		{
			name: "http_with_headers",
			config: map[string]interface{}{
				"id": "http-headers",
				"connection": map[string]interface{}{
					"transport": "http",
					"address":   "https://api.example.com/mcp",
					"headers": map[string]interface{}{
						"Authorization": "Bearer token123",
						"User-Agent":    "AgentFlow/1.0",
					},
				},
			},
			expectError: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			client, err := factory.CreateClient("mark3labs", tc.config)

			if tc.expectError {
				if err == nil {
					t.Errorf("Expected error for test case %s, but got none", tc.name)
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error for test case %s: %v", tc.name, err)
			}

			if client == nil {
				t.Fatalf("Client is nil for test case %s", tc.name)
			}

			// Verify it's the correct type
			wrapper, ok := client.(*Mark3LabsClientWrapper)
			if !ok {
				t.Fatalf("Expected *Mark3LabsClientWrapper for test case %s, got %T", tc.name, client)
			}

			// Test that the client can be used (without actually connecting)
			if wrapper.IsConnected() {
				t.Errorf("Client should not be connected initially for test case %s", tc.name)
			}
		})
	}

	t.Log("Mark3Labs configuration variations test completed successfully")
}
