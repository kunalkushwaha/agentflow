package core

import (
	"fmt"
	"os"

	"github.com/kunalkushwaha/agentflow/internal/mcp"
)

// MCPConfigLoader provides utilities for loading MCP configurations from files
type MCPConfigLoader struct {
	loader *mcp.ConfigLoader
}

// NewMCPConfigLoader creates a new MCP configuration loader
func NewMCPConfigLoader(configPath ...string) *MCPConfigLoader {
	var loader *mcp.ConfigLoader
	if len(configPath) > 0 && configPath[0] != "" {
		loader = mcp.NewConfigLoader(configPath[0])
	} else {
		loader = mcp.NewConfigLoader()
	}
	
	return &MCPConfigLoader{
		loader: loader,
	}
}

// LoadConfiguration loads MCP configuration from a TOML file
func (c *MCPConfigLoader) LoadConfiguration() (*MCPFileConfig, error) {
	config, err := c.loader.LoadConfiguration()
	if err != nil {
		return nil, err
	}
	
	return convertInternalToPublicConfig(config), nil
}

// LoadConfigurationFromString loads MCP configuration from a TOML string
func (c *MCPConfigLoader) LoadConfigurationFromString(configContent string) (*MCPFileConfig, error) {
	config, err := c.loader.LoadConfigurationFromString(configContent)
	if err != nil {
		return nil, err
	}
	
	return convertInternalToPublicConfig(config), nil
}

// GetIntegrationConfig extracts MCPConfig from file configuration
func (c *MCPConfigLoader) GetIntegrationConfig(fileConfig *MCPFileConfig) MCPConfig {
	return MCPConfig{
		ClientType:    fileConfig.MCP.DefaultClient,
		HealthCheck:   fileConfig.MCP.HealthCheck,
		AutoDiscovery: fileConfig.MCP.AutoDiscovery,
	}
}

// GetServerConfigs extracts server configurations from file configuration
func (c *MCPConfigLoader) GetServerConfigs(fileConfig *MCPFileConfig) []MCPServerConfig {
	var serverConfigs []MCPServerConfig
	
	for id, server := range fileConfig.MCP.Servers {
		config := MCPServerConfig{
			ID:         id,
			Name:       server.Name,
			Type:       server.Type,
			ClientType: server.ClientType,
			Enabled:    server.Enabled,
			Connection: MCPConnectionConfig{
				Transport:        server.Transport.Type,
				Command:          []string{server.Transport.Command},
				Args:             server.Transport.Args,
				Env:              server.Transport.Env,
				Endpoint:         server.Transport.Address,
				Headers:          make(map[string]string),
				WorkingDirectory: "",
			},
			Timeout: server.Timeout,
		}
		
		// Set defaults if not specified
		if config.ClientType == "" {
			config.ClientType = fileConfig.MCP.DefaultClient
		}
		if config.Timeout == 0 {
			config.Timeout = 30000 // 30 seconds
		}
		
		serverConfigs = append(serverConfigs, config)
	}
	
	return serverConfigs
}

// CreateSampleConfigFile creates a sample MCP configuration file
func CreateSampleConfigFile(filename string) error {
	return mcp.CreateSampleConfig(filename)
}

// MCPFileConfig represents the structure of an MCP configuration file
type MCPFileConfig struct {
	MCP MCPFileConfigSection `toml:"mcp"`
}

// MCPFileConfigSection represents the main MCP configuration section
type MCPFileConfigSection struct {
	Enabled       bool                        `toml:"enabled"`
	DefaultClient string                      `toml:"default_client"`
	HealthCheck   bool                        `toml:"health_check"`
	AutoDiscovery bool                        `toml:"auto_discovery"`
	Servers       map[string]MCPFileServer    `toml:"servers"`
}

// MCPFileServer represents an individual MCP server configuration in file format
type MCPFileServer struct {
	Name        string                      `toml:"name"`
	Type        string                      `toml:"type"`
	ClientType  string                      `toml:"client_type"`
	Enabled     bool                        `toml:"enabled"`
	Transport   MCPFileTransport            `toml:"transport"`
	Timeout     int                         `toml:"timeout"`
	Metadata    map[string]interface{}      `toml:"metadata"`
}

// MCPFileTransport represents transport configuration in file format
type MCPFileTransport struct {
	Type    string                 `toml:"type"`
	Address string                 `toml:"address"`
	Command string                 `toml:"command"`
	Args    []string               `toml:"args"`
	Env     map[string]string      `toml:"env"`
	Options map[string]interface{} `toml:"options"`
}

// Helper function to convert internal config to public config
func convertInternalToPublicConfig(internal *mcp.MCPConfiguration) *MCPFileConfig {
	public := &MCPFileConfig{
		MCP: MCPFileConfigSection{
			Enabled:       internal.MCP.Enabled,
			DefaultClient: internal.MCP.DefaultClient,
			HealthCheck:   internal.MCP.HealthCheck,
			AutoDiscovery: internal.MCP.AutoDiscovery,
			Servers:       make(map[string]MCPFileServer),
		},
	}
	
	for id, server := range internal.MCP.Servers {
		public.MCP.Servers[id] = MCPFileServer{
			Name:       server.Name,
			Type:       server.Type,
			ClientType: server.ClientType,
			Enabled:    server.Enabled,
			Transport: MCPFileTransport{
				Type:    server.Transport.Type,
				Address: server.Transport.Address,
				Command: server.Transport.Command,
				Args:    server.Transport.Args,
				Env:     server.Transport.Env,
				Options: server.Transport.Options,
			},
			Timeout:  server.Timeout,
			Metadata: server.Metadata,
		}
	}
	
	return public
}

// CreateBasicMCPConfig creates a basic MCP configuration for common use cases
func CreateBasicMCPConfig(clientType string) MCPConfig {
	if clientType == "" {
		clientType = "mock" // Default to mock for testing
	}
	
	return MCPConfig{
		ClientType:    clientType,
		HealthCheck:   true,
		AutoDiscovery: false,
	}
}

// CreateMCPManagerFromEnvironment creates an MCP manager using environment variables
// This is a convenience function for common deployment scenarios
func CreateMCPManagerFromEnvironment() (*MCPManager, error) {
	// Check for configuration file path in environment
	configPath := os.Getenv("AGENTFLOW_MCP_CONFIG")
	if configPath != "" {
		return NewMCPManagerFromConfig(configPath)
	}
	
	// Create basic configuration from environment variables
	clientType := os.Getenv("AGENTFLOW_MCP_CLIENT_TYPE")
	if clientType == "" {
		clientType = "mock" // Default to mock
	}
	
	config := CreateBasicMCPConfig(clientType)
	manager, err := NewMCPManager(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create MCP manager from environment: %w", err)
	}
	
	// Add servers from environment if specified
	serverConfig := os.Getenv("AGENTFLOW_MCP_SERVER_CONFIG")
	if serverConfig != "" {
		// This could be extended to parse server configurations from environment
		// For now, we just create a basic mock server
		mockConfig := CreateMockServerConfig("env-mock-server", "Environment Mock Server")
		if err := manager.AddServer(mockConfig); err != nil {
			return nil, fmt.Errorf("failed to add environment server: %w", err)
		}
	}
	
	return manager, nil
}
