package bridge

import (
	"context"
	"fmt"
	"time"

	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/client/transport"
	"github.com/mark3labs/mcp-go/mcp"
)

// Mark3LabsBridge implements the MCP client interface by bridging to the mark3labs client
// This is in a separate package to avoid import cycles
type Mark3LabsBridge struct {
	client     *client.Client
	config     BridgeServerConfig
	connected  bool
	serverInfo *BridgeServerInfo
}

// BridgeServerConfig represents the server configuration needed by the bridge
// We copy the structure to avoid importing the main mcp package
type BridgeServerConfig struct {
	ID         string                 `json:"id"`
	Name       string                 `json:"name"`
	Type       string                 `json:"type"`
	ClientType string                 `json:"client_type"`
	Connection BridgeConnectionConfig `json:"connection"`
	Enabled    bool                   `json:"enabled"`
}

type BridgeConnectionConfig struct {
	Transport   string            `json:"transport"`
	Command     []string          `json:"command"`
	Environment map[string]string `json:"environment"`
	Headers     map[string]string `json:"headers"`
	Address     string            `json:"address"`
}

// BridgeServerInfo represents server information in bridge format
type BridgeServerInfo struct {
	ID           string                   `json:"id"`
	Name         string                   `json:"name"`
	Version      string                   `json:"version"`
	Status       string                   `json:"status"`
	LastSeen     *time.Time               `json:"last_seen"`
	Capabilities BridgeServerCapabilities `json:"capabilities"`
	Metadata     map[string]string        `json:"metadata"`
}

type BridgeServerCapabilities struct {
	Tools     bool     `json:"tools"`
	Resources bool     `json:"resources"`
	Prompts   bool     `json:"prompts"`
	Logging   bool     `json:"logging"`
	Features  []string `json:"features"`
}

// NewMark3LabsBridge creates a new bridge instance
func NewMark3LabsBridge() *Mark3LabsBridge {
	return &Mark3LabsBridge{
		connected: false,
	}
}

// Connect establishes connection to the MCP server
func (b *Mark3LabsBridge) Connect(ctx context.Context, config BridgeServerConfig) error {
	b.config = config

	// Create appropriate transport based on configuration
	var trans transport.Interface
	var err error

	switch config.Connection.Transport {
	case "stdio":
		if len(config.Connection.Command) == 0 {
			return fmt.Errorf("stdio transport requires command")
		}

		command := config.Connection.Command[0]
		args := config.Connection.Command[1:]

		fmt.Printf("DEBUG: Creating stdio transport with command='%s', args=%v\n", command, args)

		env := make([]string, 0)
		for k, v := range config.Connection.Environment {
			env = append(env, fmt.Sprintf("%s=%s", k, v))
		}

		trans = transport.NewStdio(command, env, args...)

	case "http", "https":
		if config.Connection.Address == "" {
			return fmt.Errorf("HTTP transport requires address")
		}

		options := []transport.StreamableHTTPCOption{}
		if len(config.Connection.Headers) > 0 {
			options = append(options, transport.WithHTTPHeaders(config.Connection.Headers))
		}

		trans, err = transport.NewStreamableHTTP(config.Connection.Address, options...)
		if err != nil {
			return fmt.Errorf("failed to create HTTP transport: %w", err)
		}

	case "sse":
		if config.Connection.Address == "" {
			return fmt.Errorf("SSE transport requires address")
		}

		options := []transport.ClientOption{}
		if len(config.Connection.Headers) > 0 {
			options = append(options, transport.WithHeaders(config.Connection.Headers))
		}

		trans, err = transport.NewSSE(config.Connection.Address, options...)
		if err != nil {
			return fmt.Errorf("failed to create SSE transport: %w", err)
		}

	default:
		return fmt.Errorf("unsupported transport type: %s", config.Connection.Transport)
	}

	// Create the client
	b.client = client.NewClient(trans)
	// Start the client with a non-cancellable context
	if err := b.client.Start(context.Background()); err != nil {
		return fmt.Errorf("failed to start client: %w", err)
	}

	// Initialize the client with a non-cancellable context
	initRequest := mcp.InitializeRequest{}
	initRequest.Params.ProtocolVersion = mcp.LATEST_PROTOCOL_VERSION
	initRequest.Params.ClientInfo = mcp.Implementation{
		Name:    "agentflow-mcp-client",
		Version: "1.0.0",
	}
	initRequest.Params.Capabilities = mcp.ClientCapabilities{}

	initResult, err := b.client.Initialize(context.Background(), initRequest)
	if err != nil {
		b.client.Close()
		return fmt.Errorf("failed to initialize client: %w", err)
	}

	// Convert server info to our format
	b.serverInfo = &BridgeServerInfo{
		ID:       config.ID,
		Name:     initResult.ServerInfo.Name,
		Version:  initResult.ServerInfo.Version,
		Status:   "connected",
		LastSeen: nil, // Will be set by health checks
		Capabilities: BridgeServerCapabilities{
			Tools:     initResult.Capabilities.Tools != nil,
			Resources: initResult.Capabilities.Resources != nil,
			Prompts:   initResult.Capabilities.Prompts != nil,
			Logging:   initResult.Capabilities.Logging != nil,
			Features:  []string{"agentflow"},
		},
		Metadata: map[string]string{
			"transport": config.Connection.Transport,
		},
	}

	b.connected = true
	return nil
}

// Disconnect closes the connection to the MCP server
func (b *Mark3LabsBridge) Disconnect(ctx context.Context) error {
	if b.client != nil {
		err := b.client.Close()
		b.client = nil
		b.connected = false
		return err
	}
	return nil
}

// IsConnected returns whether the client is connected
func (b *Mark3LabsBridge) IsConnected() bool {
	return b.connected && b.client != nil
}

// Ping checks if the server is alive
func (b *Mark3LabsBridge) Ping(ctx context.Context) error {
	if !b.IsConnected() {
		return fmt.Errorf("client not connected")
	}
	return b.client.Ping(ctx)
}

// GetServerInfo returns server information
func (b *Mark3LabsBridge) GetServerInfo(ctx context.Context) (*BridgeServerInfo, error) {
	if !b.IsConnected() {
		return nil, fmt.Errorf("client not connected")
	}
	return b.serverInfo, nil
}

// GetCapabilities returns server capabilities
func (b *Mark3LabsBridge) GetCapabilities(ctx context.Context) (*BridgeServerCapabilities, error) {
	if !b.IsConnected() {
		return nil, fmt.Errorf("client not connected")
	}
	return &b.serverInfo.Capabilities, nil
}

// BridgeToolMetadata represents tool metadata in bridge format
type BridgeToolMetadata struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	InputSchema map[string]interface{} `json:"input_schema"`
}

// ListTools returns available tools
func (b *Mark3LabsBridge) ListTools(ctx context.Context) ([]BridgeToolMetadata, error) {
	if !b.IsConnected() {
		return nil, fmt.Errorf("client not connected")
	}

	request := mcp.ListToolsRequest{}
	result, err := b.client.ListTools(ctx, request)
	if err != nil {
		return nil, fmt.Errorf("failed to list tools: %w", err)
	}

	tools := make([]BridgeToolMetadata, len(result.Tools))
	for i, tool := range result.Tools {
		tools[i] = BridgeToolMetadata{
			Name:        tool.Name,
			Description: tool.Description,
			InputSchema: convertToolInputSchema(tool.InputSchema),
		}
	}

	return tools, nil
}

// BridgeToolCallRequest represents a tool call request
type BridgeToolCallRequest struct {
	Name      string                 `json:"name"`
	Arguments map[string]interface{} `json:"arguments"`
	ServerID  string                 `json:"server_id"`
	Context   map[string]interface{} `json:"context,omitempty"`
}

// BridgeToolCallResult represents a tool call result
type BridgeToolCallResult struct {
	Content   []BridgeContentBlock   `json:"content"`
	IsError   bool                   `json:"is_error"`
	ErrorCode string                 `json:"error_code,omitempty"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// BridgeContentBlock represents content in bridge format
type BridgeContentBlock struct {
	Type     string `json:"type"`
	Content  string `json:"content"`
	MimeType string `json:"mime_type,omitempty"`
}

// CallTool executes a tool
func (b *Mark3LabsBridge) CallTool(ctx context.Context, request BridgeToolCallRequest) (*BridgeToolCallResult, error) {
	if !b.IsConnected() {
		return nil, fmt.Errorf("client not connected")
	}

	mcpRequest := mcp.CallToolRequest{}
	mcpRequest.Params.Name = request.Name
	mcpRequest.Params.Arguments = request.Arguments

	result, err := b.client.CallTool(ctx, mcpRequest)
	if err != nil {
		return nil, fmt.Errorf("failed to call tool %s: %w", request.Name, err)
	}

	// Convert result content to our format
	content := make([]BridgeContentBlock, len(result.Content))
	for i, item := range result.Content {
		switch v := item.(type) {
		case mcp.TextContent:
			content[i] = BridgeContentBlock{
				Type:    "text",
				Content: v.Text,
			}
		case mcp.ImageContent:
			content[i] = BridgeContentBlock{
				Type:     "image",
				Content:  v.Data,
				MimeType: v.MIMEType,
			}
		default:
			// Handle other content types as text for now
			content[i] = BridgeContentBlock{
				Type:    "text",
				Content: fmt.Sprintf("%v", v),
			}
		}
	}

	return &BridgeToolCallResult{
		Content:  content,
		IsError:  result.IsError,
		Metadata: nil, // mark3labs doesn't provide metadata in tool results
	}, nil
}

// Helper function to convert ToolInputSchema to map[string]interface{}
func convertToolInputSchema(schema mcp.ToolInputSchema) map[string]interface{} {
	result := make(map[string]interface{})
	result["type"] = schema.Type
	if schema.Properties != nil {
		result["properties"] = schema.Properties
	}
	if len(schema.Required) > 0 {
		result["required"] = schema.Required
	}
	return result
}
