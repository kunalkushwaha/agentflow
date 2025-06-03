package mcp

import (
	"context"
	"time"
)

// MCPClient defines the core interface for MCP client implementations.
// This abstraction allows AgentFlow to work with any MCP client library.
type MCPClient interface {
	// Connection Management
	Connect(ctx context.Context, config ServerConfig) error
	Disconnect(ctx context.Context) error
	IsConnected() bool
	Ping(ctx context.Context) error

	// Server Information
	GetServerInfo(ctx context.Context) (*ServerInfo, error)
	GetCapabilities(ctx context.Context) (*ServerCapabilities, error)

	// Tool Operations
	ListTools(ctx context.Context) ([]ToolMetadata, error)
	CallTool(ctx context.Context, request ToolCallRequest) (*ToolCallResult, error)

	// Resource Operations (if supported)
	ListResources(ctx context.Context) ([]ResourceMetadata, error)
	ReadResource(ctx context.Context, uri string) (*ResourceContent, error)

	// Prompt Operations (if supported)
	ListPrompts(ctx context.Context) ([]PromptMetadata, error)
	GetPrompt(ctx context.Context, name string, args map[string]interface{}) (*PromptResult, error)

	// Event Handling
	SetNotificationHandler(handler NotificationHandler)
	SetErrorHandler(handler ErrorHandler)
}

// MCPServerManager manages multiple MCP server connections
type MCPServerManager interface {
	// Server Management
	AddServer(config ServerConfig) error
	RemoveServer(serverID string) error
	GetServer(serverID string) (MCPClient, error)
	ListServers() []ServerInfo

	// Discovery
	DiscoverServers(ctx context.Context) ([]ServerConfig, error)
	StartAutoDiscovery(ctx context.Context, interval time.Duration) error
	StopAutoDiscovery() error

	// Health Monitoring
	HealthCheck(ctx context.Context) map[string]HealthStatus
	SetHealthCheckInterval(interval time.Duration)

	// Event Handling
	SetServerEventHandler(handler ServerEventHandler)
}

// MCPToolAdapter adapts MCP tools to AgentFlow's tool interface
type MCPToolAdapter interface {
	// Tool Registration
	RegisterMCPTool(serverID string, tool ToolMetadata) error
	UnregisterMCPTool(serverID, toolName string) error

	// Tool Discovery
	RefreshTools(ctx context.Context, serverID string) error
	GetMCPTools(serverID string) []ToolMetadata

	// Tool Execution
	ExecuteTool(ctx context.Context, serverID, toolName string, args map[string]interface{}) (*ToolCallResult, error)
	// AgentFlow Integration
	ConvertToMCPFunctionTool(serverID string, tool ToolMetadata) (MCPFunctionTool, error)
}

// MCPClientFactory creates MCP client instances based on configuration
type MCPClientFactory interface {
	CreateClient(clientType string, config map[string]interface{}) (MCPClient, error)
	SupportedClients() []string
	DefaultClient() string
}

// Data Structures

type ServerConfig struct {
	ID           string                 `toml:"id" json:"id"`
	Name         string                 `toml:"name" json:"name"`
	Type         string                 `toml:"type" json:"type"`               // "stdio", "http", "websocket"
	ClientType   string                 `toml:"client_type" json:"client_type"` // "mark3labs", "custom", etc.
	Connection   ConnectionConfig       `toml:"connection" json:"connection"`
	Capabilities []string               `toml:"capabilities" json:"capabilities"`
	Config       map[string]interface{} `toml:"config" json:"config"`
	Timeout      time.Duration          `toml:"timeout" json:"timeout"`
	Retry        RetryConfig            `toml:"retry" json:"retry"`
	Tags         []string               `toml:"tags" json:"tags"`
	Enabled      bool                   `toml:"enabled" json:"enabled"`
}

type ConnectionConfig struct {
	Transport   string            `toml:"transport" json:"transport"` // "stdio", "http", "ws"
	Address     string            `toml:"address" json:"address"`
	Command     []string          `toml:"command" json:"command"`
	Environment map[string]string `toml:"environment" json:"environment"`
	Headers     map[string]string `toml:"headers" json:"headers"`
	TLS         TLSConfig         `toml:"tls" json:"tls"`
}

type TLSConfig struct {
	Enabled            bool   `toml:"enabled" json:"enabled"`
	CertFile           string `toml:"cert_file" json:"cert_file"`
	KeyFile            string `toml:"key_file" json:"key_file"`
	CAFile             string `toml:"ca_file" json:"ca_file"`
	InsecureSkipVerify bool   `toml:"insecure_skip_verify" json:"insecure_skip_verify"`
}

type RetryConfig struct {
	MaxAttempts int           `toml:"max_attempts" json:"max_attempts"`
	Backoff     time.Duration `toml:"backoff" json:"backoff"`
	MaxBackoff  time.Duration `toml:"max_backoff" json:"max_backoff"`
}

type ServerInfo struct {
	ID           string             `json:"id"`
	Name         string             `json:"name"`
	Version      string             `json:"version"`
	Capabilities ServerCapabilities `json:"capabilities"`
	Status       ConnectionStatus   `json:"status"`
	LastSeen     time.Time          `json:"last_seen"`
	Metadata     map[string]string  `json:"metadata"`
}

type ServerCapabilities struct {
	Tools     bool     `json:"tools"`
	Resources bool     `json:"resources"`
	Prompts   bool     `json:"prompts"`
	Logging   bool     `json:"logging"`
	Features  []string `json:"features"`
}

type ConnectionStatus string

const (
	StatusConnected    ConnectionStatus = "connected"
	StatusDisconnected ConnectionStatus = "disconnected"
	StatusConnecting   ConnectionStatus = "connecting"
	StatusError        ConnectionStatus = "error"
)

type ToolMetadata struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Schema      map[string]interface{} `json:"schema"`
	ServerID    string                 `json:"server_id"`
	Tags        []string               `json:"tags"`
	Annotations map[string]interface{} `json:"annotations"`
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
	Type     string                 `json:"type"` // "text", "image", "audio", etc.
	Content  string                 `json:"content"`
	MimeType string                 `json:"mime_type,omitempty"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

type ResourceMetadata struct {
	URI         string            `json:"uri"`
	Name        string            `json:"name"`
	Description string            `json:"description"`
	MimeType    string            `json:"mime_type"`
	ServerID    string            `json:"server_id"`
	Tags        []string          `json:"tags"`
	Metadata    map[string]string `json:"metadata"`
}

type ResourceContent struct {
	URI      string                 `json:"uri"`
	Content  []ContentBlock         `json:"content"`
	Metadata map[string]interface{} `json:"metadata"`
}

type PromptMetadata struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Arguments   map[string]interface{} `json:"arguments"`
	ServerID    string                 `json:"server_id"`
	Tags        []string               `json:"tags"`
}

type PromptResult struct {
	Messages []PromptMessage        `json:"messages"`
	Metadata map[string]interface{} `json:"metadata"`
}

type PromptMessage struct {
	Role    string         `json:"role"`
	Content []ContentBlock `json:"content"`
}

type HealthStatus struct {
	Status       ConnectionStatus  `json:"status"`
	LastCheck    time.Time         `json:"last_check"`
	ResponseTime time.Duration     `json:"response_time"`
	Error        string            `json:"error,omitempty"`
	Metadata     map[string]string `json:"metadata"`
}

// Event Handlers

type NotificationHandler func(notification MCPNotification)

type MCPNotification struct {
	Type      string                 `json:"type"`
	ServerID  string                 `json:"server_id"`
	Data      map[string]interface{} `json:"data"`
	Timestamp time.Time              `json:"timestamp"`
}

type ErrorHandler func(err MCPError)

type MCPError struct {
	Type      string                 `json:"type"`
	ServerID  string                 `json:"server_id"`
	Operation string                 `json:"operation"`
	Error     error                  `json:"error"`
	Timestamp time.Time              `json:"timestamp"`
	Context   map[string]interface{} `json:"context"`
}

type ServerEventHandler func(event ServerEvent)

type ServerEvent struct {
	Type      string                 `json:"type"` // "connected", "disconnected", "tools_updated", etc.
	ServerID  string                 `json:"server_id"`
	Data      map[string]interface{} `json:"data"`
	Timestamp time.Time              `json:"timestamp"`
}

// AgentFlow Integration

// MCPFunctionTool adapts an MCP tool to AgentFlow's FunctionTool interface
// This allows MCP tools to be registered and used seamlessly with AgentFlow's ToolRegistry
type MCPFunctionTool interface {
	// FunctionTool interface from AgentFlow
	Name() string
	Call(ctx context.Context, args map[string]any) (map[string]any, error)

	// Additional MCP-specific metadata
	Description() string
	Schema() map[string]interface{}
	ServerID() string
	Metadata() map[string]interface{}
}
