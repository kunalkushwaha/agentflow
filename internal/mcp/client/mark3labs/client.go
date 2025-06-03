package mark3labs

import (
	"context"
	"fmt"
	"time"

	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/client/transport"
	"github.com/mark3labs/mcp-go/mcp"
)

// ServerConfig represents the server configuration needed by the client
// This is a copy to avoid import cycles with the main mcp package
type ServerConfig struct {
	ID         string           `json:"id"`
	Name       string           `json:"name"`
	Type       string           `json:"type"`
	ClientType string           `json:"client_type"`
	Connection ConnectionConfig `json:"connection"`
	Enabled    bool             `json:"enabled"`
}

type ConnectionConfig struct {
	Transport string            `json:"transport"`
	Command   []string          `json:"command"`
	Args      []string          `json:"args"`
	Env       map[string]string `json:"env"`
	Cwd       string            `json:"cwd"`
	Endpoint  string            `json:"endpoint"`
	Headers   map[string]string `json:"headers"`
}

// Data structures to avoid import cycles
type ServerInfo struct {
	ID           string             `json:"id"`
	Name         string             `json:"name"`
	Version      string             `json:"version"`
	Status       string             `json:"status"`
	LastSeen     *time.Time         `json:"last_seen"`
	Capabilities ServerCapabilities `json:"capabilities"`
	Metadata     map[string]string  `json:"metadata"`
}

type ServerCapabilities struct {
	Tools     bool     `json:"tools"`
	Resources bool     `json:"resources"`
	Prompts   bool     `json:"prompts"`
	Logging   bool     `json:"logging"`
	Features  []string `json:"features"`
}

type ToolMetadata struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	InputSchema map[string]interface{} `json:"input_schema"`
}

type ToolCallRequest struct {
	Name      string                 `json:"name"`
	Arguments map[string]interface{} `json:"arguments"`
	ServerID  string                 `json:"server_id"`
	Context   map[string]interface{} `json:"context,omitempty"`
}

type ToolCallResult struct {
	Content   []ContentBlock         `json:"content"`
	IsError   bool                   `json:"is_error"`
	ErrorCode string                 `json:"error_code,omitempty"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

type ContentBlock struct {
	Type     string `json:"type"`
	Text     string `json:"text,omitempty"`
	Data     string `json:"data,omitempty"`
	MimeType string `json:"mime_type,omitempty"`
}

type ResourceMetadata struct {
	URI         string `json:"uri"`
	Name        string `json:"name"`
	Description string `json:"description"`
	MimeType    string `json:"mime_type"`
}

type ResourceContent struct {
	URI      string `json:"uri"`
	Text     string `json:"text,omitempty"`
	Data     string `json:"data,omitempty"`
	MimeType string `json:"mime_type"`
}

type PromptMetadata struct {
	Name        string           `json:"name"`
	Description string           `json:"description"`
	Arguments   []PromptArgument `json:"arguments"`
}

type PromptArgument struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Required    bool   `json:"required"`
}

type PromptResult struct {
	Description string          `json:"description"`
	Messages    []PromptMessage `json:"messages"`
}

type PromptMessage struct {
	Role    string       `json:"role"`
	Content ContentBlock `json:"content"`
}

type Notification struct {
	Method string      `json:"method"`
	Params interface{} `json:"params"`
}

type NotificationHandler func(Notification)
type ErrorHandler func(error)

// Mark3LabsClient wraps the mark3labs/mcp-go client to implement our MCPClient interface
type Mark3LabsClient struct {
	client     *client.Client
	config     ServerConfig
	connected  bool
	serverInfo *ServerInfo
}

// NewMark3LabsClient creates a new Mark3Labs MCP client
func NewMark3LabsClient() *Mark3LabsClient {
	return &Mark3LabsClient{
		connected: false,
	}
}

// Connect establishes connection to the MCP server
func (c *Mark3LabsClient) Connect(ctx context.Context, config ServerConfig) error {
	c.config = config

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
		if len(config.Connection.Args) > 0 {
			args = append(args, config.Connection.Args...)
		}

		env := make([]string, 0)
		for k, v := range config.Connection.Env {
			env = append(env, fmt.Sprintf("%s=%s", k, v))
		}

		trans = transport.NewStdio(command, env, args...)
	case "http", "https":
		if config.Connection.Endpoint == "" {
			return fmt.Errorf("HTTP transport requires endpoint")
		}

		options := []transport.StreamableHTTPCOption{}
		if len(config.Connection.Headers) > 0 {
			options = append(options, transport.WithHTTPHeaders(config.Connection.Headers))
		}

		trans, err = transport.NewStreamableHTTP(config.Connection.Endpoint, options...)
		if err != nil {
			return fmt.Errorf("failed to create HTTP transport: %w", err)
		}

	case "sse":
		if config.Connection.Endpoint == "" {
			return fmt.Errorf("SSE transport requires endpoint")
		}

		options := []transport.ClientOption{}
		if len(config.Connection.Headers) > 0 {
			options = append(options, transport.WithHeaders(config.Connection.Headers))
		}

		trans, err = transport.NewSSE(config.Connection.Endpoint, options...)
		if err != nil {
			return fmt.Errorf("failed to create SSE transport: %w", err)
		}

	default:
		return fmt.Errorf("unsupported transport type: %s", config.Connection.Transport)
	}

	// Create the client
	c.client = client.NewClient(trans)

	// Start the client
	if err := c.client.Start(ctx); err != nil {
		return fmt.Errorf("failed to start client: %w", err)
	}

	// Initialize the client
	initRequest := mcp.InitializeRequest{}
	initRequest.Params.ProtocolVersion = mcp.LATEST_PROTOCOL_VERSION
	initRequest.Params.ClientInfo = mcp.Implementation{
		Name:    "agentflow-mcp-client",
		Version: "1.0.0",
	}
	initRequest.Params.Capabilities = mcp.ClientCapabilities{}

	initResult, err := c.client.Initialize(ctx, initRequest)
	if err != nil {
		c.client.Close()
		return fmt.Errorf("failed to initialize client: %w", err)
	}

	// Convert server info to our format
	c.serverInfo = &ServerInfo{
		ID:       config.ID,
		Name:     initResult.ServerInfo.Name,
		Version:  initResult.ServerInfo.Version,
		Status:   "connected",
		LastSeen: nil, // Will be set by health checks
		Capabilities: ServerCapabilities{
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

	c.connected = true
	return nil
}

// Disconnect closes the connection to the MCP server
func (c *Mark3LabsClient) Disconnect(ctx context.Context) error {
	if c.client != nil {
		err := c.client.Close()
		c.client = nil
		c.connected = false
		return err
	}
	return nil
}

// IsConnected returns whether the client is connected
func (c *Mark3LabsClient) IsConnected() bool {
	return c.connected && c.client != nil
}

// Ping checks if the server is alive
func (c *Mark3LabsClient) Ping(ctx context.Context) error {
	if !c.IsConnected() {
		return fmt.Errorf("client not connected")
	}
	return c.client.Ping(ctx)
}

// GetServerInfo returns server information
func (c *Mark3LabsClient) GetServerInfo(ctx context.Context) (*ServerInfo, error) {
	if !c.IsConnected() {
		return nil, fmt.Errorf("client not connected")
	}
	return c.serverInfo, nil
}

// GetCapabilities returns server capabilities
func (c *Mark3LabsClient) GetCapabilities(ctx context.Context) (*ServerCapabilities, error) {
	if !c.IsConnected() {
		return nil, fmt.Errorf("client not connected")
	}
	return &c.serverInfo.Capabilities, nil
}

// ListTools returns available tools
func (c *Mark3LabsClient) ListTools(ctx context.Context) ([]ToolMetadata, error) {
	if !c.IsConnected() {
		return nil, fmt.Errorf("client not connected")
	}

	request := mcp.ListToolsRequest{}
	result, err := c.client.ListTools(ctx, request)
	if err != nil {
		return nil, fmt.Errorf("failed to list tools: %w", err)
	}
	tools := make([]ToolMetadata, len(result.Tools))
	for i, tool := range result.Tools {
		tools[i] = ToolMetadata{
			Name:        tool.Name,
			Description: tool.Description,
			InputSchema: convertToolInputSchema(tool.InputSchema),
		}
	}

	return tools, nil
}

// CallTool executes a tool
func (c *Mark3LabsClient) CallTool(ctx context.Context, request ToolCallRequest) (*ToolCallResult, error) {
	if !c.IsConnected() {
		return nil, fmt.Errorf("client not connected")
	}

	mcpRequest := mcp.CallToolRequest{}
	mcpRequest.Params.Name = request.Name
	mcpRequest.Params.Arguments = request.Arguments

	result, err := c.client.CallTool(ctx, mcpRequest)
	if err != nil {
		return nil, fmt.Errorf("failed to call tool %s: %w", request.Name, err)
	}

	// Convert result content to our format
	content := make([]ContentBlock, len(result.Content))
	for i, item := range result.Content {
		switch v := item.(type) {
		case mcp.TextContent:
			content[i] = ContentBlock{
				Type: "text",
				Text: v.Text,
			}
		case mcp.ImageContent:
			content[i] = ContentBlock{
				Type:     "image",
				Data:     v.Data,
				MimeType: v.MIMEType,
			}
		default:
			// Handle other content types as text for now
			content[i] = ContentBlock{
				Type: "text",
				Text: fmt.Sprintf("%v", v),
			}
		}
	}

	return &ToolCallResult{
		Content:  content,
		IsError:  result.IsError,
		Metadata: nil, // mark3labs doesn't provide metadata in tool results
	}, nil
}

// ListResources returns available resources
func (c *Mark3LabsClient) ListResources(ctx context.Context) ([]ResourceMetadata, error) {
	if !c.IsConnected() {
		return nil, fmt.Errorf("client not connected")
	}

	request := mcp.ListResourcesRequest{}
	result, err := c.client.ListResources(ctx, request)
	if err != nil {
		return nil, fmt.Errorf("failed to list resources: %w", err)
	}

	resources := make([]ResourceMetadata, len(result.Resources))
	for i, resource := range result.Resources {
		resources[i] = ResourceMetadata{
			URI:         resource.URI,
			Name:        resource.Name,
			Description: resource.Description,
			MimeType:    resource.MIMEType,
		}
	}

	return resources, nil
}

// ReadResource reads a specific resource
func (c *Mark3LabsClient) ReadResource(ctx context.Context, uri string) (*ResourceContent, error) {
	if !c.IsConnected() {
		return nil, fmt.Errorf("client not connected")
	}

	request := mcp.ReadResourceRequest{}
	request.Params.URI = uri

	result, err := c.client.ReadResource(ctx, request)
	if err != nil {
		return nil, fmt.Errorf("failed to read resource %s: %w", uri, err)
	}

	// Convert the first content item (mark3labs may return multiple)
	if len(result.Contents) == 0 {
		return nil, fmt.Errorf("no content returned for resource %s", uri)
	}

	firstContent := result.Contents[0]

	// Try to assert to different content types
	var content ResourceContent
	content.URI = uri

	switch v := firstContent.(type) {
	case mcp.TextResourceContents:
		content.Text = v.Text
		content.MimeType = v.MIMEType
	case mcp.BlobResourceContents:
		content.Data = v.Blob
		content.MimeType = v.MIMEType
	default:
		// Fallback to text representation
		content.Text = fmt.Sprintf("%v", v)
		content.MimeType = "text/plain"
	}

	return &content, nil
}

// ListPrompts returns available prompts
func (c *Mark3LabsClient) ListPrompts(ctx context.Context) ([]PromptMetadata, error) {
	if !c.IsConnected() {
		return nil, fmt.Errorf("client not connected")
	}

	request := mcp.ListPromptsRequest{}
	result, err := c.client.ListPrompts(ctx, request)
	if err != nil {
		return nil, fmt.Errorf("failed to list prompts: %w", err)
	}

	prompts := make([]PromptMetadata, len(result.Prompts))
	for i, prompt := range result.Prompts {
		// Convert arguments
		args := make([]PromptArgument, len(prompt.Arguments))
		for j, arg := range prompt.Arguments {
			args[j] = PromptArgument{
				Name:        arg.Name,
				Description: arg.Description,
				Required:    arg.Required,
			}
		}

		prompts[i] = PromptMetadata{
			Name:        prompt.Name,
			Description: prompt.Description,
			Arguments:   args,
		}
	}

	return prompts, nil
}

// GetPrompt executes a prompt
func (c *Mark3LabsClient) GetPrompt(ctx context.Context, name string, args map[string]interface{}) (*PromptResult, error) {
	if !c.IsConnected() {
		return nil, fmt.Errorf("client not connected")
	}

	request := mcp.GetPromptRequest{}
	request.Params.Name = name
	// Convert args to the expected format
	if args != nil {
		request.Params.Arguments = make(map[string]string)
		for k, v := range args {
			request.Params.Arguments[k] = fmt.Sprintf("%v", v)
		}
	}

	result, err := c.client.GetPrompt(ctx, request)
	if err != nil {
		return nil, fmt.Errorf("failed to get prompt %s: %w", name, err)
	}

	// Convert messages
	messages := make([]PromptMessage, len(result.Messages))
	for i, msg := range result.Messages {
		var content ContentBlock

		// Convert content based on type
		switch v := msg.Content.(type) {
		case mcp.TextContent:
			content = ContentBlock{
				Type: "text",
				Text: v.Text,
			}
		case mcp.ImageContent:
			content = ContentBlock{
				Type:     "image",
				Data:     v.Data,
				MimeType: v.MIMEType,
			}
		default:
			content = ContentBlock{
				Type: "text",
				Text: fmt.Sprintf("%v", v),
			}
		}

		messages[i] = PromptMessage{
			Role:    string(msg.Role),
			Content: content,
		}
	}

	return &PromptResult{
		Description: result.Description,
		Messages:    messages,
	}, nil
}

// SetNotificationHandler sets the notification handler
func (c *Mark3LabsClient) SetNotificationHandler(handler NotificationHandler) {
	if c.client != nil {
		c.client.OnNotification(func(notification mcp.JSONRPCNotification) {
			handler(Notification{
				Method: notification.Method,
				Params: notification.Params,
			})
		})
	}
}

// SetErrorHandler sets the error handler
func (c *Mark3LabsClient) SetErrorHandler(handler ErrorHandler) {
	// mark3labs client doesn't have a separate error handler mechanism
	// Errors are returned through normal function calls
}

// Helper function to convert JSON schema
func convertJSONSchema(schema map[string]interface{}) map[string]interface{} {
	if schema == nil {
		return make(map[string]interface{})
	}
	return schema
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
