package mcp

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestConfigLoader_LoadConfiguration(t *testing.T) { // Create a temporary config file
	configContent := `
[mcp]
enabled = true
default_client = "mock"
health_check = true
auto_discovery = false

[mcp.servers.test-server]
name = "Test Server"
type = "mock"
client_type = "mock"
enabled = true
timeout = 5000

[mcp.servers.test-server.transport]
type = "stdio"
command = "test-command"
args = ["arg1", "arg2"]

[mcp.servers.test-server.retry_policy]
max_retries = 5
initial_delay_ms = 500
max_delay_ms = 10000
`

	tmpFile := "test_config.toml"
	err := os.WriteFile(tmpFile, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create temp config file: %v", err)
	}
	defer os.Remove(tmpFile)

	// Test loading configuration
	loader := NewConfigLoader(tmpFile)
	config, err := loader.LoadConfiguration()
	if err != nil {
		t.Fatalf("Failed to load configuration: %v", err)
	}

	// Verify configuration values
	assert.True(t, config.MCP.Enabled)
	assert.Equal(t, "mock", config.MCP.DefaultClient)
	assert.True(t, config.MCP.HealthCheck)
	assert.False(t, config.MCP.AutoDiscovery)

	// Verify server configuration
	assert.Len(t, config.MCP.Servers, 1)

	server, exists := config.MCP.Servers["test-server"]
	assert.True(t, exists)
	assert.Equal(t, "Test Server", server.Name)
	assert.Equal(t, "mock", server.Type)
	assert.True(t, server.Enabled)
	assert.Equal(t, 5000, server.Timeout)

	// Verify transport configuration
	assert.Equal(t, "stdio", server.Transport.Type)
	assert.Equal(t, "test-command", server.Transport.Command)
	assert.Len(t, server.Transport.Args, 2)

	// Verify retry policy
	assert.Equal(t, 5, server.RetryPolicy.MaxRetries)
	assert.Equal(t, 500, server.RetryPolicy.InitialDelay)
	assert.Equal(t, 10000, server.RetryPolicy.MaxDelay)
}

func TestConfigLoader_ConvertToServerConfigs(t *testing.T) {
	config := &MCPConfiguration{
		MCP: MCPConfig{
			DefaultClient: "mock",
			Servers: map[string]MCPServer{
				"server1": {
					Name:       "Server 1",
					Type:       "stdio",
					ClientType: "mock",
					Enabled:    true,
					Timeout:    10000,
					Transport: MCPTransport{
						Type:    "stdio",
						Command: "python",
						Args:    []string{"-m", "server"},
					},
					RetryPolicy: MCPRetryPolicy{
						MaxRetries:   3,
						InitialDelay: 1000,
						MaxDelay:     5000,
					},
				},
			},
		},
	}

	loader := NewConfigLoader()
	serverConfigs := loader.ConvertToServerConfigs(config)

	assert.Len(t, serverConfigs, 1)

	// Find server1 config
	var server1Config *ServerConfig
	for _, sc := range serverConfigs {
		if sc.ID == "server1" {
			server1Config = &sc
			break
		}
	}

	assert.NotNil(t, server1Config)
	assert.Equal(t, "Server 1", server1Config.Name)
	assert.Equal(t, "stdio", server1Config.Type)
	assert.True(t, server1Config.Enabled)
	assert.Equal(t, time.Duration(10000), server1Config.Timeout)
	assert.Equal(t, "stdio", server1Config.Connection.Transport)
	assert.Len(t, server1Config.Connection.Command, 2)
	assert.Equal(t, "python", server1Config.Connection.Command[0])
	assert.Equal(t, 3, server1Config.Retry.MaxAttempts)
}

func TestConfigLoader_GetIntegrationConfig(t *testing.T) {
	config := &MCPConfiguration{
		MCP: MCPConfig{
			DefaultClient: "mark3labs",
			HealthCheck:   true,
			AutoDiscovery: true,
		},
	}

	loader := NewConfigLoader()
	integrationConfig := loader.GetIntegrationConfig(config)

	assert.Equal(t, "mark3labs", integrationConfig.ClientType)
	assert.True(t, integrationConfig.HealthCheck)
	assert.True(t, integrationConfig.AutoDiscovery)
}

func TestCreateSampleConfig(t *testing.T) {
	filename := "test_sample_config.toml"
	defer os.Remove(filename)

	err := CreateSampleConfig(filename)
	assert.NoError(t, err)

	// Verify file was created
	_, err = os.Stat(filename)
	assert.NoError(t, err)

	// Try to load the sample config
	loader := NewConfigLoader(filename)
	config, err := loader.LoadConfiguration()
	assert.NoError(t, err)

	// Verify sample config has expected structure
	assert.True(t, config.MCP.Enabled)
	assert.NotEmpty(t, config.MCP.Servers)
}
