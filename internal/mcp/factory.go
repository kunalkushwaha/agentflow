package mcp

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/kunalkushwaha/agentflow/internal/mcp/bridge"
)

// ClientFactory implements MCPClientFactory interface
type ClientFactory struct {
	creators map[string]ClientCreator
	mu       sync.RWMutex
}

// ClientCreator is a function that creates an MCP client instance
type ClientCreator func(config map[string]interface{}) (MCPClient, error)

// NewClientFactory creates a new client factory
func NewClientFactory() *ClientFactory {
	factory := &ClientFactory{
		creators: make(map[string]ClientCreator),
	}

	// Register default implementations
	factory.registerDefaultClients()

	return factory
}

// RegisterClient registers a new client implementation
func (f *ClientFactory) RegisterClient(clientType string, creator ClientCreator) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.creators[clientType] = creator
}

// CreateClient creates a client of the specified type
func (f *ClientFactory) CreateClient(clientType string, config map[string]interface{}) (MCPClient, error) {
	f.mu.RLock()
	creator, exists := f.creators[clientType]
	f.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("unsupported MCP client type: %s", clientType)
	}

	return creator(config)
}

// SupportedClients returns a list of supported client types
func (f *ClientFactory) SupportedClients() []string {
	f.mu.RLock()
	defer f.mu.RUnlock()

	clients := make([]string, 0, len(f.creators))
	for clientType := range f.creators {
		clients = append(clients, clientType)
	}
	return clients
}

// DefaultClient returns the default client type
func (f *ClientFactory) DefaultClient() string {
	return "mark3labs" // Current default implementation
}

// registerDefaultClients registers the built-in client implementations
func (f *ClientFactory) registerDefaultClients() {
	// Register mark3labs implementation
	f.RegisterClient("mark3labs", func(config map[string]interface{}) (MCPClient, error) {
		return NewMark3LabsBridge(config)
	})

	// Register custom implementation (placeholder for future)
	f.RegisterClient("custom", func(config map[string]interface{}) (MCPClient, error) {
		return NewCustomClient(config)
	})

	// Register mock implementation for testing
	f.RegisterClient("mock", func(config map[string]interface{}) (MCPClient, error) {
		return NewMockClient(config)
	})
}

// Global factory instance
var globalFactory = NewClientFactory()

// RegisterGlobalClient registers a client implementation globally
func RegisterGlobalClient(clientType string, creator ClientCreator) {
	globalFactory.RegisterClient(clientType, creator)
}

// CreateGlobalClient creates a client using the global factory
func CreateGlobalClient(clientType string, config map[string]interface{}) (MCPClient, error) {
	return globalFactory.CreateClient(clientType, config)
}

// GetSupportedClients returns supported clients from global factory
func GetSupportedClients() []string {
	return globalFactory.SupportedClients()
}

// GetDefaultClient returns the default client type
func GetDefaultClient() string {
	return globalFactory.DefaultClient()
}

// NewMark3LabsBridge creates a new bridge to the mark3labs client
// This avoids import cycles by using a separate bridge package
func NewMark3LabsBridge(config map[string]interface{}) (MCPClient, error) {
	return NewMark3LabsClientWrapper(config)
}

// Mark3LabsClientWrapper implements MCPClient by wrapping the mark3labs bridge
type Mark3LabsClientWrapper struct {
	bridge *bridge.Mark3LabsBridge
	config ServerConfig
}

// NewMark3LabsClientWrapper creates a new wrapper around the mark3labs bridge
func NewMark3LabsClientWrapper(config map[string]interface{}) (MCPClient, error) {
	bridge := bridge.NewMark3LabsBridge()
	wrapper := &Mark3LabsClientWrapper{
		bridge: bridge,
	}

	// Convert config map to ServerConfig
	serverConfig, err := convertMapToServerConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to convert config: %w", err)
	}
	wrapper.config = serverConfig

	return wrapper, nil
}

// Implement MCPClient interface methods

func (w *Mark3LabsClientWrapper) Connect(ctx context.Context, config ServerConfig) error {
	// Convert MCPClient ServerConfig to bridge ServerConfig
	bridgeConfig := bridge.BridgeServerConfig{
		ID:         config.ID,
		Name:       config.Name,
		Type:       config.Type,
		ClientType: config.ClientType,
		Connection: bridge.BridgeConnectionConfig{
			Transport:   config.Connection.Transport,
			Command:     config.Connection.Command,
			Environment: config.Connection.Environment,
			Headers:     config.Connection.Headers,
			Address:     config.Connection.Address,
		},
		Enabled: config.Enabled,
	}

	w.config = config
	return w.bridge.Connect(ctx, bridgeConfig)
}

func (w *Mark3LabsClientWrapper) Disconnect(ctx context.Context) error {
	return w.bridge.Disconnect(ctx)
}

func (w *Mark3LabsClientWrapper) IsConnected() bool {
	return w.bridge.IsConnected()
}

func (w *Mark3LabsClientWrapper) Ping(ctx context.Context) error {
	return w.bridge.Ping(ctx)
}

func (w *Mark3LabsClientWrapper) GetServerInfo(ctx context.Context) (*ServerInfo, error) {
	info, err := w.bridge.GetServerInfo(ctx)
	if err != nil {
		return nil, err
	}

	var lastSeen time.Time
	if info.LastSeen != nil {
		lastSeen = *info.LastSeen
	}

	return &ServerInfo{
		ID:       info.ID,
		Name:     info.Name,
		Version:  info.Version,
		Status:   ConnectionStatus(info.Status),
		LastSeen: lastSeen,
		Capabilities: ServerCapabilities{
			Tools:     info.Capabilities.Tools,
			Resources: info.Capabilities.Resources,
			Prompts:   info.Capabilities.Prompts,
			Logging:   info.Capabilities.Logging,
			Features:  info.Capabilities.Features,
		},
		Metadata: info.Metadata,
	}, nil
}

func (w *Mark3LabsClientWrapper) GetCapabilities(ctx context.Context) (*ServerCapabilities, error) {
	caps, err := w.bridge.GetCapabilities(ctx)
	if err != nil {
		return nil, err
	}

	return &ServerCapabilities{
		Tools:     caps.Tools,
		Resources: caps.Resources,
		Prompts:   caps.Prompts,
		Logging:   caps.Logging,
		Features:  caps.Features,
	}, nil
}

func (w *Mark3LabsClientWrapper) ListTools(ctx context.Context) ([]ToolMetadata, error) {
	tools, err := w.bridge.ListTools(ctx)
	if err != nil {
		return nil, err
	}

	mcpTools := make([]ToolMetadata, len(tools))
	for i, tool := range tools {
		mcpTools[i] = ToolMetadata{
			Name:        tool.Name,
			Description: tool.Description,
			Schema:      tool.InputSchema,
			ServerID:    w.config.ID,
			Tags:        []string{},
			Annotations: make(map[string]interface{}),
		}
	}

	return mcpTools, nil
}

func (w *Mark3LabsClientWrapper) CallTool(ctx context.Context, request ToolCallRequest) (*ToolCallResult, error) {
	bridgeRequest := bridge.BridgeToolCallRequest{
		Name:      request.Name,
		Arguments: request.Arguments,
		ServerID:  request.ServerID,
		Context:   request.Context,
	}

	result, err := w.bridge.CallTool(ctx, bridgeRequest)
	if err != nil {
		return nil, err
	}

	mcpContent := make([]ContentBlock, len(result.Content))
	for i, content := range result.Content {
		mcpContent[i] = ContentBlock{
			Type:     content.Type,
			Content:  content.Content,
			MimeType: content.MimeType,
			Metadata: make(map[string]interface{}),
		}
	}

	return &ToolCallResult{
		Content:   mcpContent,
		IsError:   result.IsError,
		ErrorCode: result.ErrorCode,
		Metadata:  result.Metadata,
	}, nil
}

func (w *Mark3LabsClientWrapper) ListResources(ctx context.Context) ([]ResourceMetadata, error) {
	// mark3labs bridge doesn't implement resources yet
	return []ResourceMetadata{}, nil
}

func (w *Mark3LabsClientWrapper) ReadResource(ctx context.Context, uri string) (*ResourceContent, error) {
	// mark3labs bridge doesn't implement resources yet
	return nil, fmt.Errorf("resources not implemented in mark3labs bridge")
}

func (w *Mark3LabsClientWrapper) ListPrompts(ctx context.Context) ([]PromptMetadata, error) {
	// mark3labs bridge doesn't implement prompts yet
	return []PromptMetadata{}, nil
}

func (w *Mark3LabsClientWrapper) GetPrompt(ctx context.Context, name string, args map[string]interface{}) (*PromptResult, error) {
	// mark3labs bridge doesn't implement prompts yet
	return nil, fmt.Errorf("prompts not implemented in mark3labs bridge")
}

func (w *Mark3LabsClientWrapper) SetNotificationHandler(handler NotificationHandler) {
	// mark3labs bridge doesn't implement notifications yet
}

func (w *Mark3LabsClientWrapper) SetErrorHandler(handler ErrorHandler) {
	// mark3labs bridge doesn't implement error handlers yet
}
