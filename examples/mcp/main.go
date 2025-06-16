package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/kunalkushwaha/agentflow/internal/mcp"
	"github.com/kunalkushwaha/agentflow/internal/tools"
)

func main() {
	fmt.Println("AgentFlow MCP Integration Example")
	fmt.Println("=================================")

	// Example 1: Basic MCP Integration with Mock Client
	if err := basicMCPIntegrationExample(); err != nil {
		log.Printf("Basic integration example failed: %v", err)
	}

	// Example 2: Configuration-based MCP Integration
	if err := configurationBasedExample(); err != nil {
		log.Printf("Configuration-based example failed: %v", err)
	}

	// Example 3: AgentFlow Tool Registry Integration
	if err := agentFlowIntegrationExample(); err != nil {
		log.Printf("AgentFlow integration example failed: %v", err)
	}

	// Example 4: Advanced Usage with Multiple Servers
	if err := multipleServersExample(); err != nil {
		log.Printf("Multiple servers example failed: %v", err)
	}

	// Example 5: Agent using MCP Tool
	if err := agentWithMCPDemo(); err != nil {
		log.Printf("Agent MCP demo failed: %v", err)
	}

	fmt.Println("\nAll examples completed. Press Ctrl+C to exit.")

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	fmt.Println("\nShutting down gracefully...")
}

// basicMCPIntegrationExample demonstrates basic MCP integration
func basicMCPIntegrationExample() error {
	fmt.Println("\n--- Example 1: Basic MCP Integration ---")

	// Create integration configuration
	config := mcp.MCPIntegrationConfig{
		ClientType:    "mock",
		HealthCheck:   true,
		AutoDiscovery: false,
	}

	// Create MCP integration
	integration, err := mcp.NewMCPIntegration(config)
	if err != nil {
		return fmt.Errorf("failed to create MCP integration: %w", err)
	}
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		integration.Shutdown(ctx)
	}()
	// Add a mock MCP server
	serverConfig := mcp.ServerConfig{
		ID:         "example-mock-server",
		Name:       "Example Mock Server",
		Type:       "mock",
		ClientType: "mock",
		Enabled:    true,
		Connection: mcp.ConnectionConfig{
			Transport: "stdio",
		},
	}

	if err := integration.AddServer(serverConfig); err != nil {
		return fmt.Errorf("failed to add server: %w", err)
	}

	fmt.Printf("✓ Added server: %s\n", serverConfig.Name) // List available tools
	ctx := context.Background()
	if err := integration.RefreshTools(ctx); err != nil {
		return fmt.Errorf("failed to refresh tools: %w", err)
	}
	// Get tool information by checking each server
	servers := integration.ListServers()
	fmt.Printf("✓ Found %d servers\n", len(servers))

	// For demonstration, just show we connected successfully
	for _, server := range servers {
		fmt.Printf("  - %s: %s\n", server.Name, server.Status)
	}

	return nil
}

// configurationBasedExample demonstrates loading configuration from TOML
func configurationBasedExample() error {
	fmt.Println("\n--- Example 2: Configuration-based Integration ---")

	// Create a sample configuration
	configContent := `
[mcp]
enabled = true
default_client = "mock"
health_check = true
auto_discovery = false

[mcp.servers.config-example]
name = "Configuration Example Server"
type = "mock"
client_type = "mock"
enabled = true
timeout = 5000

[mcp.servers.config-example.transport]
type = "stdio"
command = "mock-server"

[mcp.servers.config-example.metadata]
description = "A server loaded from configuration"
example = true
`

	// Load configuration from string
	loader := mcp.NewConfigLoader()
	config, err := loader.LoadConfigurationFromString(configContent)
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	fmt.Printf("✓ Loaded configuration with %d servers\n", len(config.MCP.Servers))

	// Create integration from configuration
	integrationConfig := loader.GetIntegrationConfig(config)
	integration, err := mcp.NewMCPIntegration(integrationConfig)
	if err != nil {
		return fmt.Errorf("failed to create integration: %w", err)
	}
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		integration.Shutdown(ctx)
	}()

	// Add servers from configuration
	serverConfigs := loader.ConvertToServerConfigs(config)
	for _, serverConfig := range serverConfigs {
		if err := integration.AddServer(serverConfig); err != nil {
			return fmt.Errorf("failed to add server %s: %w", serverConfig.ID, err)
		}
		fmt.Printf("✓ Added server from config: %s\n", serverConfig.Name)
	}

	// List servers from configured servers
	servers := integration.ListServers()
	fmt.Printf("✓ Configuration-based servers: %d\n", len(servers))

	return nil
}

// agentFlowIntegrationExample demonstrates integration with AgentFlow's tool system
func agentFlowIntegrationExample() error {
	fmt.Println("\n--- Example 3: AgentFlow Tool Registry Integration ---")

	// Create AgentFlow tool registry
	toolRegistry := tools.NewToolRegistry()

	// Create MCP integration
	config := mcp.MCPIntegrationConfig{
		ClientType:    "mock",
		HealthCheck:   false,
		AutoDiscovery: false,
	}

	integration, err := mcp.NewMCPIntegrationWithRegistry(toolRegistry, config)
	if err != nil {
		return fmt.Errorf("failed to create integration: %w", err)
	}
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		integration.Shutdown(ctx)
	}()

	// Add MCP server
	serverConfig := mcp.ServerConfig{
		ID:         "agentflow-integration",
		Name:       "AgentFlow Integration Server",
		Type:       "mock",
		ClientType: "mock",
		Enabled:    true,
		Connection: mcp.ConnectionConfig{
			Transport: "stdio",
		},
	}

	if err := integration.AddServer(serverConfig); err != nil {
		return fmt.Errorf("failed to add server: %w", err)
	}

	fmt.Printf("✓ Added server and registered with AgentFlow\n")

	// Check if tools are available through the integration
	servers := integration.ListServers()
	fmt.Printf("✓ AgentFlow integration has %d servers\n", len(servers))

	return nil
}

// multipleServersExample demonstrates working with multiple MCP servers
func multipleServersExample() error {
	fmt.Println("\n--- Example 4: Multiple Servers ---")

	// Create integration
	config := mcp.MCPIntegrationConfig{
		ClientType:    "mock",
		HealthCheck:   true,
		AutoDiscovery: false,
	}

	integration, err := mcp.NewMCPIntegration(config)
	if err != nil {
		return fmt.Errorf("failed to create integration: %w", err)
	}
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		integration.Shutdown(ctx)
	}()

	// Add multiple servers
	servers := []mcp.ServerConfig{
		{
			ID:         "server-1",
			Name:       "Mock Server 1",
			Type:       "mock",
			ClientType: "mock",
			Enabled:    true,
			Connection: mcp.ConnectionConfig{
				Transport: "stdio",
			},
		},
		{
			ID:         "server-2",
			Name:       "Mock Server 2",
			Type:       "mock",
			ClientType: "mock",
			Enabled:    true,
			Connection: mcp.ConnectionConfig{
				Transport: "stdio",
			},
		},
		{
			ID:         "server-3",
			Name:       "Mock Server 3",
			Type:       "mock",
			ClientType: "mock",
			Enabled:    true,
			Connection: mcp.ConnectionConfig{
				Transport: "stdio",
			},
		},
	}

	for _, serverConfig := range servers {
		if err := integration.AddServer(serverConfig); err != nil {
			return fmt.Errorf("failed to add server %s: %w", serverConfig.ID, err)
		}
		fmt.Printf("✓ Added server: %s\n", serverConfig.Name)
	}

	// Check server status
	fmt.Println("✓ Server health status:")
	serverList := integration.ListServers()
	for _, server := range serverList {
		fmt.Printf("  - %s: %s\n", server.Name, server.Status)
	}

	// Demonstrate removing a server
	if err := integration.RemoveServer("server-2"); err != nil {
		return fmt.Errorf("failed to remove server: %w", err)
	}
	fmt.Printf("✓ Removed server-2\n")

	// Check servers after removal
	remainingServers := integration.ListServers()
	fmt.Printf("✓ Servers after removal: %d\n", len(remainingServers))

	return nil
}

// agentWithMCPDemo demonstrates an agent using an MCP tool via the ToolRegistry
func agentWithMCPDemo() error {
	fmt.Println("\n--- Example 5: Agent using MCP Tool ---")

	// Create a ToolRegistry and MCP integration
	toolRegistry := tools.NewToolRegistry()
	config := mcp.MCPIntegrationConfig{
		ClientType:    "mock",
		HealthCheck:   false,
		AutoDiscovery: false,
	}
	integration, err := mcp.NewMCPIntegrationWithRegistry(toolRegistry, config)
	if err != nil {
		return fmt.Errorf("failed to create integration: %w", err)
	}
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		integration.Shutdown(ctx)
	}()

	// Add a mock MCP server
	serverConfig := mcp.ServerConfig{
		ID:         "agent-mcp-server",
		Name:       "Agent MCP Server",
		Type:       "mock",
		ClientType: "mock",
		Enabled:    true,
		Connection: mcp.ConnectionConfig{
			Transport: "stdio",
		},
	}
	if err := integration.AddServer(serverConfig); err != nil {
		return fmt.Errorf("failed to add server: %w", err)
	}

	// Refresh tools to ensure MCP tools are registered
	ctx := context.Background()
	if err := integration.RefreshTools(ctx); err != nil {
		return fmt.Errorf("failed to refresh tools: %w", err)
	}

	// List all available tools
	allTools := toolRegistry.List()
	fmt.Printf("Available tools (%d):\n", len(allTools))
	for _, tool := range allTools {
		fmt.Printf("  - %s\n", tool.Name())
	}

	if len(allTools) == 0 {
		fmt.Println("No tools available.")
		return nil
	}

	// Call the first available tool with dummy args
	toolToCall := allTools[0]
	fmt.Printf("\nCalling tool: %s\n", toolToCall.Name())
	result, err := toolToCall.Call(ctx, map[string]any{"example": "test"})
	if err != nil {
		fmt.Printf("Tool call failed: %v\n", err)
	} else {
		fmt.Printf("Tool call result: %+v\n", result)
	}

	return nil
}
