---
title: "Architecture Overview"
description: "Core vs Internal package design principles and AgentFlow's system architecture"
weight: 40
---

AgentFlow uses a clear separation between public API (`core/`) and private implementation (`internal/`) to provide a stable, developer-friendly interface while maintaining implementation flexibility.

## Package Structure Overview

```
agentflow/
├── core/           # Public API - what users import
│   ├── agent.go    # Agent interfaces and types
│   ├── mcp.go      # MCP integration public API  
│   ├── factory.go  # Factory functions for creating components
│   ├── llm.go      # LLM provider interfaces
│   └── ...         # Other public interfaces
└── internal/       # Private implementation - not importable
    ├── agents/     # Concrete agent implementations
    ├── mcp/        # MCP client and server management
    ├── llm/        # LLM provider implementations
    ├── orchestrator/ # Workflow orchestration logic
    └── ...         # Other implementation packages
```

## Design Principles

### 1. Interface Segregation

**Public interfaces are defined in `core/`:**

```go
// core/agent.go
type AgentHandler interface {
    Run(ctx context.Context, event Event, state State) (AgentResult, error)
}

type ModelProvider interface {
    Generate(ctx context.Context, prompt string) (string, error)
    GenerateWithHistory(ctx context.Context, messages []Message) (string, error)
    Name() string
}

type MCPManager interface {
    ListTools(ctx context.Context) ([]ToolSchema, error)
    CallTool(ctx context.Context, name string, args map[string]interface{}) (interface{}, error)
    // ... more methods
}
```

**Implementations are in `internal/`:**

```go
// internal/agents/mcp_agent.go
type mcpAgent struct {
    name       string
    llm        llm.Provider          // internal interface
    mcpManager mcp.Manager           // internal interface
}

func (a *mcpAgent) Run(ctx context.Context, event core.Event, state core.State) (core.AgentResult, error) {
    // Implementation details hidden from users
}
```

### 2. Factory Pattern

**Factories in `core/` create internal implementations:**

```go
// core/factory.go
func NewMCPAgent(name string, llm ModelProvider, mcp MCPManager) AgentHandler {
    // Create internal implementation
    return agents.NewMCPAgent(name, llm, mcp)
}

func InitializeProductionMCP(ctx context.Context, config MCPConfig) (MCPManager, error) {
    // Create internal MCP manager
    return mcp.NewProductionManager(ctx, config)
}
```

This pattern allows users to work with interfaces while we manage complex implementations internally.

## Core Package Structure

### agent.go - Agent System

```go
// Primary interfaces for agent development
type AgentHandler interface {
    Run(ctx context.Context, event Event, state State) (AgentResult, error)
}

type Agent interface {
    Run(ctx context.Context, inputState State) (State, error)
    Name() string
}

// Supporting types
type Event interface {
    GetData() EventData
    GetSessionID() string
    GetMeta() map[string]interface{}
}

type State interface {
    Get(key string) (interface{}, bool)
    Set(key string, value interface{})
    GetMeta(key string) (interface{}, bool)
    SetMeta(key string, value interface{})
}

type AgentResult struct {
    Result string
    State  State
    Error  error
}
```

### mcp.go - MCP Integration

```go
// MCP interfaces for tool integration
type MCPManager interface {
    ListTools(ctx context.Context) ([]ToolSchema, error)
    CallTool(ctx context.Context, name string, args map[string]interface{}) (interface{}, error)
    Connect(ctx context.Context) error
    Disconnect() error
    IsConnected() bool
}

type ToolSchema struct {
    Name        string                 `json:"name"`
    Description string                 `json:"description"`
    Parameters  map[string]interface{} `json:"parameters"`
}

// High-level MCP functions
func FormatToolsForPrompt(ctx context.Context, manager MCPManager) string
func ParseAndExecuteToolCalls(ctx context.Context, manager MCPManager, response string) []interface{}
```

### llm.go - LLM Providers

```go
// LLM provider abstraction
type ModelProvider interface {
    Generate(ctx context.Context, prompt string) (string, error)
    GenerateWithHistory(ctx context.Context, messages []Message) (string, error)
    Name() string
}

type Message struct {
    Role    string `json:"role"`    // system, user, assistant
    Content string `json:"content"`
}

// Provider configuration
type ProviderConfig struct {
    Type       string            `toml:"type"`
    APIKey     string            `toml:"api_key"`
    Endpoint   string            `toml:"endpoint"`
    Model      string            `toml:"model"`
    MaxTokens  int               `toml:"max_tokens"`
    Temperature float64          `toml:"temperature"`
    Timeout    string            `toml:"timeout"`
    Extra      map[string]interface{} `toml:"extra"`
}
```

### factory.go - Creation Functions

```go
// Agent creation
func NewMCPAgent(name string, llm ModelProvider, mcp MCPManager) AgentHandler
func NewBasicAgent(name string, llm ModelProvider) AgentHandler

// Service initialization
func InitializeLLMProvider(config ProviderConfig) (ModelProvider, error)
func InitializeProductionMCP(ctx context.Context, config MCPConfig) (MCPManager, error)
func InitializeRunner(config RunnerConfig) (*Runner, error)

// Configuration loading
func LoadConfig(configPath string) (*Config, error)
func LoadConfigFromEnv() (*Config, error)
```

## Internal Package Structure

### internal/agents/ - Agent Implementations

```go
// internal/agents/mcp_agent.go
type mcpAgent struct {
    name       string
    llm        llm.Provider
    mcpManager mcp.Manager
    logger     *logger.Logger
}

// internal/agents/basic_agent.go  
type basicAgent struct {
    name   string
    llm    llm.Provider
    logger *logger.Logger
}

// internal/agents/workflow_agent.go
type workflowAgent struct {
    name     string
    steps    []WorkflowStep
    executor *executor.Engine
}
```

### internal/mcp/ - MCP Implementation

```go
// internal/mcp/manager.go
type manager struct {
    clients map[string]*client.MCPClient
    cache   *toolCache
    config  MCPConfig
    logger  *logger.Logger
}

// internal/mcp/client.go
type mcpClient struct {
    cmd     *exec.Cmd
    stdin   io.WriteCloser
    stdout  io.ReadCloser
    config  ServerConfig
}

// internal/mcp/cache.go
type toolCache struct {
    schemas map[string][]ToolSchema
    ttl     time.Duration
    mutex   sync.RWMutex
}
```

### internal/llm/ - LLM Implementations

```go
// internal/llm/azure_openai.go
type azureOpenAIProvider struct {
    client   *openai.Client
    config   AzureConfig
    logger   *logger.Logger
}

// internal/llm/openai.go
type openAIProvider struct {
    client *openai.Client
    config OpenAIConfig
    logger *logger.Logger
}

// internal/llm/ollama.go
type ollamaProvider struct {
    client *http.Client
    config OllamaConfig
    logger *logger.Logger
}
```

## Benefits of This Architecture

### 1. Stable Public API

- Users only import from `core/`
- Internal changes don't break user code
- Clear separation of concerns
- Easy to document and maintain

### 2. Implementation Flexibility

- Can refactor internal code freely
- Add new features without API changes
- Optimize performance without user impact
- Support multiple implementations

### 3. Testing Boundaries

- Mock interfaces for unit testing
- Test internal and public APIs separately
- Clear testing responsibilities
- Easy to write focused tests

### 4. Clear Documentation Focus

- Document public API thoroughly
- Internal documentation for contributors
- Clear usage patterns
- Reduced cognitive load for users

## Guidelines for Development

### When to Add to `core/`

Add to `core/` when:
- Users need to import the interface/type
- It's part of the public API contract
- It defines behavior users depend on
- It's a factory function users call

**Example:**
```go
// core/agent.go - Users need this interface
type AgentHandler interface {
    Run(ctx context.Context, event Event, state State) (AgentResult, error)
}

// core/factory.go - Users call this function
func NewMCPAgent(name string, llm ModelProvider) AgentHandler
```

### When to Keep in `internal/`

Keep in `internal/` when:
- It's an implementation detail
- Users don't need to know about it
- It might change frequently
- It's complex business logic

**Example:**
```go
// internal/mcp/client.go - Implementation detail
type mcpClient struct {
    cmd    *exec.Cmd
    stdin  io.WriteCloser
    stdout io.ReadCloser
}

// internal/agents/mcp_agent.go - Concrete implementation
type mcpAgent struct {
    name       string
    llm        llm.Provider
    mcpManager mcp.Manager
}
```

### Interface Design Patterns

**Do:**
```go
// Small, focused interfaces
type ToolExecutor interface {
    ExecuteTool(ctx context.Context, name string, args map[string]interface{}) (interface{}, error)
}

// Composition of interfaces
type MCPManager interface {
    ToolDiscoverer
    ToolExecutor
    ConnectionManager
}
```

**Don't:**
```go
// Large, monolithic interfaces
type EverythingManager interface {
    // 20+ methods that do different things
}
```

## Migration and Versioning

### Adding New Features

1. **Add to internal first**: Implement and test internally
2. **Design public interface**: Create minimal, focused interface
3. **Add factory function**: Provide creation function in `core/`
4. **Update documentation**: Document new capabilities

### Deprecating Features

1. **Mark as deprecated**: Add deprecation comments
2. **Provide migration path**: Show how to use new API
3. **Keep old implementation**: Don't break existing code
4. **Remove in major version**: Clean up in next major release

## Performance Considerations

### Interface Overhead

Go interfaces have minimal overhead:
- Method calls through interfaces are fast
- Interface conversions are optimized
- The separation doesn't impact performance

### Memory Management

- Interfaces don't increase memory usage significantly
- Internal implementations can be optimized independently
- Factory functions don't add overhead

### Compilation Benefits

- Users only compile against `core/` interfaces
- `internal/` changes don't trigger user recompilation
- Faster development iteration

## Next Steps

- **[Agent Fundamentals](../agent-basics/)** - Learn how to build agents with this architecture
- **[Tool Integration](../tool-integration/)** - Understand MCP integration patterns
- **[Configuration Management](../configuration/)** - Configure your AgentFlow applications