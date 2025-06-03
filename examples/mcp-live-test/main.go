package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/kunalkushwaha/agentflow/internal/mcp"
	"github.com/kunalkushwaha/agentflow/internal/tools"
)

func main() {
	fmt.Println("AgentFlow MCP Live Testing")
	fmt.Println("==========================")

	// Test 1: Mock Client (always works)
	fmt.Println("\n1. Testing Mock Client...")
	if err := testMockClient(); err != nil {
		log.Printf("Mock client test failed: %v", err)
	} else {
		fmt.Println("✓ Mock client test passed")
	}

	// Test 2: Mark3Labs Client with Python MCP Server
	fmt.Println("\n2. Testing Mark3Labs Client with Python MCP Server...")
	if err := testMark3LabsWithPythonServer(); err != nil {
		log.Printf("Mark3Labs with Python server test failed: %v", err)
	} else {
		fmt.Println("✓ Mark3Labs with Python server test passed")
	}

	// Test 3: Configuration-based Integration
	fmt.Println("\n3. Testing Configuration-based Integration...")
	if err := testConfigurationIntegration(); err != nil {
		log.Printf("Configuration integration test failed: %v", err)
	} else {
		fmt.Println("✓ Configuration integration test passed")
	}

	// Test 4: Performance Testing
	fmt.Println("\n4. Running Performance Tests...")
	if err := runPerformanceTests(); err != nil {
		log.Printf("Performance tests failed: %v", err)
	} else {
		fmt.Println("✓ Performance tests passed")
	}

	fmt.Println("\nLive testing completed!")
}

func testMockClient() error {
	// Create MCP integration with mock client
	config := mcp.MCPIntegrationConfig{
		ClientType:    "mock",
		HealthCheck:   true,
		AutoDiscovery: false,
	}

	integration, err := mcp.NewMCPIntegration(config)
	if err != nil {
		return fmt.Errorf("failed to create integration: %w", err)
	}
	defer integration.Shutdown(context.Background())

	// Add mock server
	serverConfig := mcp.ServerConfig{
		ID:         "mock-test-server",
		Name:       "Mock Test Server",
		Type:       "mock",
		ClientType: "mock",
		Enabled:    true,
		Connection: mcp.ConnectionConfig{
			Transport: "stdio",
		},
		Timeout: 5000,
	}

	if err := integration.AddServer(serverConfig); err != nil {
		return fmt.Errorf("failed to add server: %w", err)
	}

	// Wait for connection and refresh tools
	time.Sleep(100 * time.Millisecond)
	ctx := context.Background()

	if err := integration.RefreshTools(ctx); err != nil {
		return fmt.Errorf("failed to refresh tools: %w", err)
	}

	// Test tool execution
	result, err := integration.ExecuteTool(ctx, "mock-test-server", "mock_tool", map[string]interface{}{
		"input": "test input for mock",
	})
	if err != nil {
		return fmt.Errorf("failed to execute tool: %w", err)
	}

	fmt.Printf("  Mock tool result: %+v\n", result)
	return nil
}

func testMark3LabsWithPythonServer() error {
	// Check if Python is available
	pythonPath, err := findPython()
	if err != nil {
		return fmt.Errorf("Python not found, skipping: %w", err)
	}

	// Get the script path
	scriptPath, err := getTestServerPath()
	if err != nil {
		return fmt.Errorf("test server script not found: %w", err)
	}

	fmt.Printf("  Using Python: %s\n", pythonPath)
	fmt.Printf("  Using script: %s\n", scriptPath)

	// Create MCP integration with mark3labs client
	config := mcp.MCPIntegrationConfig{
		ClientType:    "mark3labs",
		HealthCheck:   true,
		AutoDiscovery: false,
	}

	integration, err := mcp.NewMCPIntegration(config)
	if err != nil {
		return fmt.Errorf("failed to create integration: %w", err)
	}
	defer integration.Shutdown(context.Background())

	// Add Python MCP server
	serverConfig := mcp.ServerConfig{
		ID:         "python-test-server",
		Name:       "Python Test Server",
		Type:       "stdio",
		ClientType: "mark3labs",
		Enabled:    true,
		Connection: mcp.ConnectionConfig{
			Transport: "stdio",
			Command:   []string{pythonPath, scriptPath},
		},
		Timeout: 30000, // 30 seconds
	}

	if err := integration.AddServer(serverConfig); err != nil {
		return fmt.Errorf("failed to add server: %w", err)
	}

	// Wait for connection
	fmt.Println("  Waiting for server connection...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel() // Wait for server to be ready
	for i := 0; i < 20; i++ {
		servers := integration.ListServers()
		if len(servers) > 0 {
			fmt.Printf("  Server status: %s\n", servers[0].Status)

			// Check health status to get error details
			healthStatus := integration.GetHealthStatus(context.Background())
			if status, exists := healthStatus["python-test-server"]; exists && status.Error != "" {
				fmt.Printf("  Connection error: %s\n", status.Error)
			}

			if servers[0].Status == "connected" {
				break
			}
		} else {
			fmt.Printf("  No servers found (attempt %d)\n", i+1)
		}
		time.Sleep(500 * time.Millisecond)
	}

	// Refresh tools
	if err := integration.RefreshTools(ctx); err != nil {
		return fmt.Errorf("failed to refresh tools: %w", err)
	}
	// Test available tools
	tools := integration.GetServerTools("python-test-server")
	if len(tools) == 0 {
		return fmt.Errorf("no tools found from Python server")
	}

	fmt.Printf("  Found %d tools from Python server\n", len(tools))
	for _, tool := range tools {
		fmt.Printf("    - %s: %s\n", tool.Name, tool.Description)
	}

	// Test timestamp tool (no arguments)
	result, err := integration.ExecuteTool(ctx, "python-test-server", "get_timestamp", map[string]interface{}{})
	if err != nil {
		return fmt.Errorf("failed to execute get_timestamp tool: %w", err)
	}
	fmt.Printf("  Timestamp result: %+v\n", result)

	// Test directory listing
	tempDir := os.TempDir()
	result, err = integration.ExecuteTool(ctx, "python-test-server", "list_directory", map[string]interface{}{
		"path": tempDir,
	})
	if err != nil {
		return fmt.Errorf("failed to execute list_directory tool: %w", err)
	}
	fmt.Printf("  Directory listing result: success\n")

	// Test file operations
	testFile := filepath.Join(tempDir, "agentflow_mcp_test.txt")
	testContent := "Hello from AgentFlow MCP integration!"

	// Write file
	result, err = integration.ExecuteTool(ctx, "python-test-server", "write_file", map[string]interface{}{
		"path":    testFile,
		"content": testContent,
	})
	if err != nil {
		return fmt.Errorf("failed to execute write_file tool: %w", err)
	}
	fmt.Printf("  File write result: success\n")

	// Read file back
	result, err = integration.ExecuteTool(ctx, "python-test-server", "read_file", map[string]interface{}{
		"path": testFile,
	})
	if err != nil {
		return fmt.Errorf("failed to execute read_file tool: %w", err)
	}
	fmt.Printf("  File read result: success\n")

	// Clean up test file
	os.Remove(testFile)

	return nil
}

func testConfigurationIntegration() error {
	// Create temporary config file
	configContent := `
[mcp]
enabled = true
default_client = "mock"
health_check = true
auto_discovery = false

[mcp.servers.config-test]
name = "Configuration Test Server"
type = "mock"
client_type = "mock"
enabled = true
timeout = 5000

[mcp.servers.config-test.transport]
type = "stdio"
command = "mock"

[mcp.servers.config-test.metadata]
description = "A server loaded from configuration"
test_mode = true
`

	// Create configuration loader
	loader := mcp.NewConfigLoader()
	config, err := loader.LoadConfigurationFromString(configContent)
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// Create integration from configuration
	integrationConfig := loader.GetIntegrationConfig(config)
	integration, err := mcp.NewMCPIntegration(integrationConfig)
	if err != nil {
		return fmt.Errorf("failed to create integration: %w", err)
	}
	defer integration.Shutdown(context.Background())

	// Add servers from configuration
	serverConfigs := loader.ConvertToServerConfigs(config)
	for _, serverConfig := range serverConfigs {
		if err := integration.AddServer(serverConfig); err != nil {
			return fmt.Errorf("failed to add server %s: %w", serverConfig.ID, err)
		}
	}

	// Test that servers were loaded
	servers := integration.ListServers()
	if len(servers) == 0 {
		return fmt.Errorf("no servers loaded from configuration")
	}

	fmt.Printf("  Loaded %d servers from configuration\n", len(servers))
	for _, server := range servers {
		fmt.Printf("    - %s: %s\n", server.Name, server.Status)
	}

	return nil
}

func runPerformanceTests() error {
	// Create mock integration for performance testing
	config := mcp.MCPIntegrationConfig{
		ClientType:    "mock",
		HealthCheck:   false, // Disable for performance testing
		AutoDiscovery: false,
	}

	integration, err := mcp.NewMCPIntegration(config)
	if err != nil {
		return fmt.Errorf("failed to create integration: %w", err)
	}
	defer integration.Shutdown(context.Background())

	// Add mock server
	serverConfig := mcp.ServerConfig{
		ID:         "perf-test-server",
		Name:       "Performance Test Server",
		Type:       "mock",
		ClientType: "mock",
		Enabled:    true,
		Connection: mcp.ConnectionConfig{
			Transport: "stdio",
		},
		Timeout: 5000,
	}

	if err := integration.AddServer(serverConfig); err != nil {
		return fmt.Errorf("failed to add server: %w", err)
	}

	time.Sleep(100 * time.Millisecond)
	ctx := context.Background()

	// Performance test: Multiple tool executions
	const numExecutions = 100
	start := time.Now()

	for i := 0; i < numExecutions; i++ {
		_, err := integration.ExecuteTool(ctx, "perf-test-server", "mock_tool", map[string]interface{}{
			"input": fmt.Sprintf("performance test %d", i),
		})
		if err != nil {
			return fmt.Errorf("tool execution %d failed: %w", i, err)
		}
	}

	duration := time.Since(start)
	avgDuration := duration / numExecutions

	fmt.Printf("  Executed %d tools in %v\n", numExecutions, duration)
	fmt.Printf("  Average execution time: %v\n", avgDuration)

	if avgDuration > 10*time.Millisecond {
		fmt.Printf("  Warning: Average execution time is high (>10ms)\n")
	}

	// Performance test: Multiple servers
	start = time.Now()
	const numServers = 10

	servers := make([]string, numServers)
	for i := 0; i < numServers; i++ {
		serverID := fmt.Sprintf("perf-server-%d", i)
		servers[i] = serverID

		serverConfig := mcp.ServerConfig{
			ID:         serverID,
			Name:       fmt.Sprintf("Performance Server %d", i),
			Type:       "mock",
			ClientType: "mock",
			Enabled:    true,
			Connection: mcp.ConnectionConfig{
				Transport: "stdio",
			},
			Timeout: 5000,
		}

		if err := integration.AddServer(serverConfig); err != nil {
			return fmt.Errorf("failed to add server %d: %w", i, err)
		}
	}

	serverCreationTime := time.Since(start)
	fmt.Printf("  Created %d servers in %v\n", numServers, serverCreationTime)

	// Test tool registry integration performance
	toolRegistry := tools.NewToolRegistry()
	start = time.Now()

	if err := integration.RegisterWithAgentFlow(toolRegistry); err != nil {
		return fmt.Errorf("failed to register with AgentFlow: %w", err)
	}

	registrationTime := time.Since(start)
	registeredTools := toolRegistry.List()

	fmt.Printf("  Registered %d tools in %v\n", len(registeredTools), registrationTime)

	return nil
}

func findPython() (string, error) {
	// Try common Python executables
	pythonCommands := []string{"python3", "python", "py"}

	for _, cmd := range pythonCommands {
		if path, err := exec.LookPath(cmd); err == nil {
			return path, nil
		}
	}

	return "", fmt.Errorf("Python not found in PATH")
}

func getTestServerPath() (string, error) {
	// Get the current working directory
	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}

	// Look for the test server script
	scriptPath := filepath.Join(cwd, "scripts", "test_mcp_server.py")
	if _, err := os.Stat(scriptPath); err == nil {
		return scriptPath, nil
	}

	// Try relative to this file's location
	scriptPath = filepath.Join(cwd, "..", "scripts", "test_mcp_server.py")
	if _, err := os.Stat(scriptPath); err == nil {
		return scriptPath, nil
	}

	return "", fmt.Errorf("test server script not found")
}
