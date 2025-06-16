package mcp

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/BurntSushi/toml"
)

// MCPConfiguration represents the MCP configuration from TOML files
type MCPConfiguration struct {
	MCP MCPConfigSection `toml:"mcp"`
}

// MCPConfigSection represents the main MCP configuration section
type MCPConfigSection struct {
	Enabled       bool                 `toml:"enabled"`
	DefaultClient string               `toml:"default_client"`
	HealthCheck   bool                 `toml:"health_check"`
	AutoDiscovery bool                 `toml:"auto_discovery"`
	Servers       map[string]MCPServer `toml:"servers"`
}

// MCPServer represents an individual MCP server configuration
type MCPServer struct {
	Name        string                 `toml:"name"`
	Type        string                 `toml:"type"`
	ClientType  string                 `toml:"client_type"`
	Enabled     bool                   `toml:"enabled"`
	Transport   MCPTransport           `toml:"transport"`
	Auth        MCPAuth                `toml:"auth"`
	Timeout     int                    `toml:"timeout"`
	RetryPolicy MCPRetryPolicy         `toml:"retry_policy"`
	Metadata    map[string]interface{} `toml:"metadata"`
}

// MCPTransport represents transport configuration
type MCPTransport struct {
	Type    string                 `toml:"type"`
	Address string                 `toml:"address"`
	Port    int                    `toml:"port"`
	Path    string                 `toml:"path"`
	Command string                 `toml:"command"`
	Args    []string               `toml:"args"`
	Env     map[string]string      `toml:"env"`
	Options map[string]interface{} `toml:"options"`
}

// MCPAuth represents authentication configuration
type MCPAuth struct {
	Type     string            `toml:"type"`
	Username string            `toml:"username"`
	Password string            `toml:"password"`
	Token    string            `toml:"token"`
	Headers  map[string]string `toml:"headers"`
}

// MCPRetryPolicy represents retry policy configuration
type MCPRetryPolicy struct {
	MaxRetries    int `toml:"max_retries"`
	InitialDelay  int `toml:"initial_delay_ms"`
	MaxDelay      int `toml:"max_delay_ms"`
	BackoffFactor int `toml:"backoff_factor"`
}

// ConfigLoader handles loading MCP configuration from TOML files
type ConfigLoader struct {
	configPaths []string
}

// NewConfigLoader creates a new configuration loader
func NewConfigLoader(configPaths ...string) *ConfigLoader {
	if len(configPaths) == 0 {
		// Default configuration paths
		configPaths = []string{
			"agentflow.toml",
			"config/agentflow.toml",
			"./agentflow.toml",
		}
	}

	return &ConfigLoader{
		configPaths: configPaths,
	}
}

// LoadConfiguration loads MCP configuration from TOML files
func (c *ConfigLoader) LoadConfiguration() (*MCPConfiguration, error) {
	var config MCPConfiguration

	// Try each configuration path
	for _, path := range c.configPaths {
		if _, err := os.Stat(path); err == nil {
			// File exists, try to load it
			if _, err := toml.DecodeFile(path, &config); err != nil {
				return nil, fmt.Errorf("failed to parse TOML config file %s: %w", path, err)
			}

			// Validate and set defaults
			c.setDefaults(&config)

			if err := c.validateConfiguration(&config); err != nil {
				return nil, fmt.Errorf("invalid configuration in %s: %w", path, err)
			}

			return &config, nil
		}
	}

	// No configuration file found, return default configuration
	defaultConfig := c.getDefaultConfiguration()
	return &defaultConfig, nil
}

// LoadConfigurationFromString loads configuration from a TOML string
func (c *ConfigLoader) LoadConfigurationFromString(tomlContent string) (*MCPConfiguration, error) {
	var config MCPConfiguration

	if _, err := toml.Decode(tomlContent, &config); err != nil {
		return nil, fmt.Errorf("failed to parse TOML content: %w", err)
	}

	c.setDefaults(&config)

	if err := c.validateConfiguration(&config); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return &config, nil
}

// ConvertToServerConfigs converts TOML configuration to ServerConfig structs
func (c *ConfigLoader) ConvertToServerConfigs(config *MCPConfiguration) []ServerConfig {
	var serverConfigs []ServerConfig

	for id, server := range config.MCP.Servers { // Convert command and args to command slice
		var command []string
		if server.Transport.Command != "" {
			if len(server.Transport.Args) > 0 {
				// Combine args into a single argument string
				argsStr := strings.Join(server.Transport.Args, " ")
				command = []string{server.Transport.Command, argsStr}
			} else {
				command = []string{server.Transport.Command}
			}
		}

		serverConfig := ServerConfig{
			ID:         id,
			Name:       server.Name,
			Type:       server.Type,
			ClientType: server.ClientType,
			Connection: ConnectionConfig{
				Transport:   server.Transport.Type,
				Address:     server.Transport.Address,
				Command:     command,
				Environment: server.Transport.Env,
				Headers:     server.Auth.Headers,
			},
			Capabilities: []string{}, // Default empty capabilities
			Config:       server.Metadata,
			Timeout:      time.Duration(server.Timeout),
			Retry: RetryConfig{
				MaxAttempts: server.RetryPolicy.MaxRetries,
				Backoff:     time.Duration(server.RetryPolicy.InitialDelay) * time.Millisecond,
				MaxBackoff:  time.Duration(server.RetryPolicy.MaxDelay) * time.Millisecond,
			},
			Tags:    []string{}, // Default empty tags
			Enabled: server.Enabled,
		}

		serverConfigs = append(serverConfigs, serverConfig)
	}

	return serverConfigs
}

// GetIntegrationConfig creates an MCPIntegrationConfig from TOML configuration
func (c *ConfigLoader) GetIntegrationConfig(config *MCPConfiguration) MCPIntegrationConfig {
	clientType := config.MCP.DefaultClient
	if clientType == "" {
		clientType = "mock" // Default to mock for testing
	}

	return MCPIntegrationConfig{
		ClientType:    clientType,
		HealthCheck:   config.MCP.HealthCheck,
		AutoDiscovery: config.MCP.AutoDiscovery,
	}
}

// setDefaults sets default values for configuration
func (c *ConfigLoader) setDefaults(config *MCPConfiguration) {
	if config.MCP.DefaultClient == "" {
		config.MCP.DefaultClient = "mock"
	}

	// Set defaults for each server
	for id, server := range config.MCP.Servers {
		if server.Name == "" {
			server.Name = id
		}
		if server.ClientType == "" {
			server.ClientType = config.MCP.DefaultClient
		}
		if server.Timeout == 0 {
			server.Timeout = 30000 // 30 seconds
		}
		if server.Transport.Type == "" {
			server.Transport.Type = "stdio"
		}

		// Set retry policy defaults
		if server.RetryPolicy.MaxRetries == 0 {
			server.RetryPolicy.MaxRetries = 3
		}
		if server.RetryPolicy.InitialDelay == 0 {
			server.RetryPolicy.InitialDelay = 1000 // 1 second
		}
		if server.RetryPolicy.MaxDelay == 0 {
			server.RetryPolicy.MaxDelay = 30000 // 30 seconds
		}
		if server.RetryPolicy.BackoffFactor == 0 {
			server.RetryPolicy.BackoffFactor = 2
		}

		// Update the server in the map
		config.MCP.Servers[id] = server
	}
}

// validateConfiguration validates the configuration
func (c *ConfigLoader) validateConfiguration(config *MCPConfiguration) error {
	// Validate default client type
	supportedClients := []string{"mock", "mark3labs", "custom"}
	validClient := false
	for _, client := range supportedClients {
		if config.MCP.DefaultClient == client {
			validClient = true
			break
		}
	}
	if !validClient {
		return fmt.Errorf("unsupported default client type: %s", config.MCP.DefaultClient)
	}

	// Validate each server configuration
	for id, server := range config.MCP.Servers {
		if server.Name == "" {
			return fmt.Errorf("server %s: name is required", id)
		}

		// Validate transport type
		validTransport := false
		supportedTransports := []string{"stdio", "websocket", "http"}
		for _, transport := range supportedTransports {
			if server.Transport.Type == transport {
				validTransport = true
				break
			}
		}
		if !validTransport {
			return fmt.Errorf("server %s: unsupported transport type: %s", id, server.Transport.Type)
		}

		// Validate transport configuration based on type
		switch server.Transport.Type {
		case "stdio":
			if server.Transport.Command == "" {
				return fmt.Errorf("server %s: command is required for stdio transport", id)
			}
		case "websocket", "http":
			if server.Transport.Address == "" {
				return fmt.Errorf("server %s: address is required for %s transport", id, server.Transport.Type)
			}
		}

		// Validate timeout
		if server.Timeout < 0 {
			return fmt.Errorf("server %s: timeout cannot be negative", id)
		}

		// Validate retry policy
		if server.RetryPolicy.MaxRetries < 0 {
			return fmt.Errorf("server %s: max_retries cannot be negative", id)
		}
		if server.RetryPolicy.InitialDelay < 0 {
			return fmt.Errorf("server %s: initial_delay_ms cannot be negative", id)
		}
		if server.RetryPolicy.MaxDelay < server.RetryPolicy.InitialDelay {
			return fmt.Errorf("server %s: max_delay_ms cannot be less than initial_delay_ms", id)
		}
	}

	return nil
}

// getDefaultConfiguration returns a default configuration
func (c *ConfigLoader) getDefaultConfiguration() MCPConfiguration {
	return MCPConfiguration{
		MCP: MCPConfigSection{
			Enabled:       true,
			DefaultClient: "mock",
			HealthCheck:   true,
			AutoDiscovery: false,
			Servers:       make(map[string]MCPServer),
		},
	}
}

// CreateSampleConfig creates a sample TOML configuration file
func CreateSampleConfig(filename string) error {
	sampleConfig := `
# AgentFlow MCP Configuration

[mcp]
enabled = true
default_client = "mock"
health_check = true
auto_discovery = false

# Example MCP server configurations
[mcp.servers.example-stdio]
name = "Example STDIO Server"
type = "stdio"
client_type = "mark3labs"
enabled = false
timeout = 30000

[mcp.servers.example-stdio.transport]
type = "stdio"
command = "python"
args = ["-m", "example_mcp_server"]
env = { "PYTHONPATH" = "/path/to/server" }

[mcp.servers.example-stdio.retry_policy]
max_retries = 3
initial_delay_ms = 1000
max_delay_ms = 30000
backoff_factor = 2

[mcp.servers.example-websocket]
name = "Example WebSocket Server"
type = "websocket"
client_type = "mark3labs"
enabled = false
timeout = 30000

[mcp.servers.example-websocket.transport]
type = "websocket"
address = "ws://localhost:8080"
path = "/mcp"

[mcp.servers.example-websocket.auth]
type = "bearer"
token = "your-auth-token"

[mcp.servers.mock-server]
name = "Mock Server for Testing"
type = "mock"
client_type = "mock"
enabled = true
timeout = 5000

[mcp.servers.mock-server.transport]
type = "stdio"
command = "mock"

[mcp.servers.mock-server.metadata]
description = "A mock server for development and testing"
environment = "development"
`

	return os.WriteFile(filename, []byte(sampleConfig), 0644)
}
