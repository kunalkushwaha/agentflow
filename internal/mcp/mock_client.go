package mcp

import (
	"context"
	"fmt"
	"time"
)

// MockClient provides a mock implementation for testing
type MockClient struct {
	connected     bool
	config        ServerConfig
	tools         []ToolMetadata
	resources     []ResourceMetadata
	prompts       []PromptMetadata
	notifications NotificationHandler
	errors        ErrorHandler
}

// NewMockClient creates a new mock MCP client
func NewMockClient(config map[string]interface{}) (MCPClient, error) {
	// Convert generic config to ServerConfig
	serverConfig, err := convertMapToServerConfig(config)
	if err != nil {
		return nil, fmt.Errorf("invalid mock client configuration: %w", err)
	}

	return &MockClient{
		config:    serverConfig,
		connected: false,
		tools: []ToolMetadata{
			{
				Name:        "mock_tool",
				Description: "A mock tool for testing",
				Schema: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"input": map[string]interface{}{
							"type":        "string",
							"description": "Test input",
						},
					},
				},
				ServerID: serverConfig.ID,
				Tags:     []string{"mock", "test"},
			},
		},
		resources: []ResourceMetadata{
			{
				URI:         "mock://test-resource",
				Name:        "test-resource",
				Description: "A mock resource for testing",
				MimeType:    "text/plain",
				ServerID:    serverConfig.ID,
			},
		}, prompts: []PromptMetadata{
			{
				Name:        "mock_prompt",
				Description: "A mock prompt for testing",
				Arguments: map[string]interface{}{
					"input": map[string]interface{}{
						"type":        "string",
						"description": "Test input",
						"required":    true,
					},
				},
				ServerID: serverConfig.ID,
			},
		},
	}, nil
}

// Connection Management

func (m *MockClient) Connect(ctx context.Context, config ServerConfig) error {
	if m.connected {
		return nil
	}
	m.config = config
	m.connected = true
	return nil
}

func (m *MockClient) Disconnect(ctx context.Context) error {
	m.connected = false
	return nil
}

func (m *MockClient) IsConnected() bool {
	return m.connected
}

func (m *MockClient) Ping(ctx context.Context) error {
	if !m.connected {
		return fmt.Errorf("client not connected")
	}
	return nil
}

// Server Information

func (m *MockClient) GetServerInfo(ctx context.Context) (*ServerInfo, error) {
	if !m.connected {
		return nil, fmt.Errorf("client not connected")
	}

	return &ServerInfo{
		ID:      m.config.ID,
		Name:    "Mock MCP Server",
		Version: "1.0.0",
		Capabilities: ServerCapabilities{
			Tools:     true,
			Resources: true,
			Prompts:   true,
			Logging:   true,
			Features:  []string{"mock", "test"},
		},
		Status:   StatusConnected,
		LastSeen: time.Now(),
		Metadata: map[string]string{
			"description": "A mock MCP server for testing",
			"author":      "AgentFlow",
			"homepage":    "https://github.com/kunalkushwaha/agentflow",
			"license":     "MIT",
		},
	}, nil
}

func (m *MockClient) GetCapabilities(ctx context.Context) (*ServerCapabilities, error) {
	if !m.connected {
		return nil, fmt.Errorf("client not connected")
	}

	return &ServerCapabilities{
		Tools:     true,
		Resources: true,
		Prompts:   true,
		Logging:   true,
		Features:  []string{"mock", "test", "experimental"},
	}, nil
}

// Tool Operations

func (m *MockClient) ListTools(ctx context.Context) ([]ToolMetadata, error) {
	if !m.connected {
		return nil, fmt.Errorf("client not connected")
	}
	return m.tools, nil
}

func (m *MockClient) CallTool(ctx context.Context, request ToolCallRequest) (*ToolCallResult, error) {
	if !m.connected {
		return nil, fmt.Errorf("client not connected")
	}

	// Simulate tool execution
	if request.Name == "mock_tool" {
		input, exists := request.Arguments["input"]
		if !exists {
			return &ToolCallResult{
				Content: []ContentBlock{{Type: "text", Content: "No input provided"}},
				IsError: true,
			}, nil
		}

		response := fmt.Sprintf("Mock tool executed with input: %v", input)
		return &ToolCallResult{
			Content: []ContentBlock{{Type: "text", Content: response}},
			IsError: false,
		}, nil
	}

	return &ToolCallResult{
		Content: []ContentBlock{{Type: "text", Content: "Unknown tool"}},
		IsError: true,
	}, nil
}

// Resource Operations

func (m *MockClient) ListResources(ctx context.Context) ([]ResourceMetadata, error) {
	if !m.connected {
		return nil, fmt.Errorf("client not connected")
	}
	return m.resources, nil
}

func (m *MockClient) ReadResource(ctx context.Context, uri string) (*ResourceContent, error) {
	if !m.connected {
		return nil, fmt.Errorf("client not connected")
	}

	if uri == "mock://test-resource" {
		return &ResourceContent{
			URI: uri,
			Content: []ContentBlock{
				{
					Type:     "text",
					Content:  "This is mock resource content",
					MimeType: "text/plain",
				},
			},
			Metadata: map[string]interface{}{
				"source": "mock",
				"type":   "test-resource",
			},
		}, nil
	}

	return nil, fmt.Errorf("resource not found: %s", uri)
}

// Prompt Operations

func (m *MockClient) ListPrompts(ctx context.Context) ([]PromptMetadata, error) {
	if !m.connected {
		return nil, fmt.Errorf("client not connected")
	}
	return m.prompts, nil
}

func (m *MockClient) GetPrompt(ctx context.Context, name string, args map[string]interface{}) (*PromptResult, error) {
	if !m.connected {
		return nil, fmt.Errorf("client not connected")
	}

	if name == "mock_prompt" {
		input, exists := args["input"]
		if !exists {
			input = "default"
		}

		return &PromptResult{
			Messages: []PromptMessage{
				{
					Role: "user",
					Content: []ContentBlock{
						{Type: "text", Content: fmt.Sprintf("Mock prompt with input: %v", input)},
					},
				},
			},
			Metadata: map[string]interface{}{
				"prompt_name": name,
				"generated":   time.Now().Format(time.RFC3339),
			},
		}, nil
	}

	return nil, fmt.Errorf("prompt not found: %s", name)
}

// Event Handling

func (m *MockClient) SetNotificationHandler(handler NotificationHandler) {
	m.notifications = handler
}

func (m *MockClient) SetErrorHandler(handler ErrorHandler) {
	m.errors = handler
}

// CustomClient provides a placeholder for custom MCP client implementations
type CustomClient struct {
	config ServerConfig
}

// NewCustomClient creates a new custom MCP client
func NewCustomClient(config map[string]interface{}) (MCPClient, error) {
	return nil, fmt.Errorf("custom MCP client not implemented - use this as a template for custom implementations")
}

// Helper function to convert generic config map to ServerConfig
func convertMapToServerConfig(config map[string]interface{}) (ServerConfig, error) {
	var serverConfig ServerConfig

	if id, ok := config["id"].(string); ok {
		serverConfig.ID = id
	} else {
		serverConfig.ID = "mock-server"
	}

	if name, ok := config["name"].(string); ok {
		serverConfig.Name = name
	} else {
		serverConfig.Name = "Mock Server"
	}

	if serverType, ok := config["type"].(string); ok {
		serverConfig.Type = serverType
	} else {
		serverConfig.Type = "mock"
	}

	if clientType, ok := config["client_type"].(string); ok {
		serverConfig.ClientType = clientType
	} else {
		serverConfig.ClientType = "mock"
	}

	// Parse connection configuration
	if connConfig, ok := config["connection"].(map[string]interface{}); ok {
		var connectionConfig ConnectionConfig

		if transport, ok := connConfig["transport"].(string); ok {
			connectionConfig.Transport = transport
		}

		if command, ok := connConfig["command"].([]interface{}); ok {
			stringCommand := make([]string, len(command))
			for i, v := range command {
				stringCommand[i] = fmt.Sprintf("%v", v)
			}
			connectionConfig.Command = stringCommand
		}

		if env, ok := connConfig["environment"].(map[string]interface{}); ok {
			envMap := make(map[string]string)
			for k, v := range env {
				envMap[k] = fmt.Sprintf("%v", v)
			}
			connectionConfig.Environment = envMap
		}

		if headers, ok := connConfig["headers"].(map[string]interface{}); ok {
			headerMap := make(map[string]string)
			for k, v := range headers {
				headerMap[k] = fmt.Sprintf("%v", v)
			}
			connectionConfig.Headers = headerMap
		}

		if address, ok := connConfig["address"].(string); ok {
			connectionConfig.Address = address
		}

		serverConfig.Connection = connectionConfig
	}

	serverConfig.Enabled = true
	serverConfig.Timeout = 30 * time.Second

	return serverConfig, nil
}
