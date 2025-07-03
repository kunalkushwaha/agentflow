---
title: "Architecture Overview"
weight: 5
description: >
  High-level overview of AgentFlow's architecture and design principles.
---

## Core Architecture

AgentFlow is built with a clear separation between public APIs and internal implementation, following Go best practices for package organization and interface design.

```
┌─────────────────────────────────────────────────┐
│                User Applications                │
├─────────────────────────────────────────────────┤
│                core/ (Public API)              │
│  ┌─────────────┐ ┌─────────────┐ ┌─────────────┐│
│  │   Agents    │ │     LLM     │ │     MCP     ││
│  │ Interfaces  │ │ Providers   │ │Integration  ││
│  └─────────────┘ └─────────────┘ └─────────────┘│
├─────────────────────────────────────────────────┤
│            internal/ (Implementation)           │
│  ┌─────────────┐ ┌─────────────┐ ┌─────────────┐│
│  │   Agent     │ │    LLM      │ │     MCP     ││
│  │Implementations│ │Implementations│ │Implementation││
│  └─────────────┘ └─────────────┘ └─────────────┘│
└─────────────────────────────────────────────────┘
```

## Design Principles

### 1. Interface-First Design

All components are defined as interfaces in the `core/` package:

```go
// core/agent.go
type AgentHandler interface {
    Process(ctx context.Context, message string) (string, error)
    GetName() string
    GetCapabilities() []string
}

// core/llm.go  
type ModelProvider interface {
    Generate(ctx context.Context, messages []Message) (*Response, error)
    GetModel() string
    GetConfig() ProviderConfig
}
```

### 2. Dependency Injection

Components are wired together through constructor functions:

```go
// Factory pattern for clean dependencies
func NewMCPAgent(name string, provider ModelProvider, mcpConfig *MCPConfig) AgentHandler {
    return &internal.MCPAgent{
        name:     name,
        provider: provider,
        mcp:      internal.NewMCPClient(mcpConfig),
    }
}
```

### 3. Error Handling

Comprehensive error types provide context for debugging:

```go
type AgentError struct {
    Agent     string
    Operation string
    Cause     error
}

type MCPError struct {
    Tool    string
    Message string
    Cause   error
}
```

## Package Structure

### Core Package (`core/`)

**Public API contracts that users depend on:**

- `agent.go` - Agent interfaces and types
- `llm.go` - LLM provider interfaces  
- `mcp.go` - MCP integration interfaces
- `factory.go` - Constructor functions
- `errors.go` - Public error types

### Internal Package (`internal/`)

**Implementation details that can change:**

- `agents/` - Concrete agent implementations
- `llm/` - LLM provider implementations  
- `mcp/` - MCP protocol handling
- `config/` - Configuration management
- `utils/` - Shared utilities

### Command Package (`cmd/`)

**CLI applications and tooling:**

- `agentcli/` - Main CLI application
- `tools/` - Development and build tools

## Key Components

### Agent System

```go
// Basic agent for simple interactions
type BasicAgent struct {
    name     string
    provider ModelProvider
}

// MCP-enabled agent with tool access
type MCPAgent struct {
    name     string
    provider ModelProvider
    mcp      MCPClient
    cache    ToolCache
}
```

### LLM Providers

```go
// Azure OpenAI implementation
type AzureProvider struct {
    client     *azureopenai.Client
    deployment string
    config     AzureConfig
}

// OpenAI implementation  
type OpenAIProvider struct {
    client *openai.Client
    config OpenAIConfig
}
```

### MCP Integration

```go
// MCP client for tool discovery and execution
type MCPClient struct {
    servers    map[string]MCPServer
    cache      ToolCache
    config     MCPConfig
}
```

## Data Flow

### Simple Agent Interaction

```
User Request → Agent → LLM Provider → Response
```

### MCP-Enhanced Interaction

```
User Request → Agent → Tool Discovery → Tool Execution → LLM Provider → Response
     ↑                      ↓              ↓
     └── Response ←── Result Processing ←── Tool Results
```

## Configuration Management

### Hierarchical Configuration

```toml
# Base configuration
[provider]
type = "azure"
model = "gpt-4"

# Environment-specific overrides
[provider.dev]
type = "mock"

[provider.prod]
max_tokens = 4000
timeout = "60s"
```

### Environment Integration

```go
// Configuration loading with environment variable support
config := agentflow.LoadConfig("agentflow.toml")
config.ExpandEnvironmentVariables()
```

## Concurrency and Performance

### Context Propagation

All operations accept `context.Context` for:
- Request cancellation
- Deadline enforcement  
- Request tracing
- Value propagation

### Resource Management

```go
// Automatic cleanup of MCP servers
defer agent.Close()

// Connection pooling for LLM providers
provider.SetMaxConnections(10)

// Caching for tool discovery
cache := agentflow.NewToolCache(time.Minute * 5)
```

## Extension Points

### Custom Agents

```go
type CustomAgent struct {
    // Embed standard functionality
    *internal.BasicAgent
    
    // Add custom behavior
    customTool Tool
}

func (a *CustomAgent) Process(ctx context.Context, message string) (string, error) {
    // Custom pre-processing
    processed := a.preProcess(message)
    
    // Delegate to base implementation
    response, err := a.BasicAgent.Process(ctx, processed)
    
    // Custom post-processing
    return a.postProcess(response), err
}
```

### Custom Providers

```go
type CustomProvider struct {
    endpoint string
    apiKey   string
}

func (p *CustomProvider) Generate(ctx context.Context, messages []Message) (*Response, error) {
    // Custom LLM integration
    return p.callCustomAPI(ctx, messages)
}
```

## Testing Strategy

### Interface Mocking

```go
type MockProvider struct {
    responses []string
    index     int
}

func (m *MockProvider) Generate(ctx context.Context, messages []Message) (*Response, error) {
    response := m.responses[m.index]
    m.index++
    return &Response{Content: response}, nil
}
```

### Integration Testing

```go
func TestAgentIntegration(t *testing.T) {
    provider := &MockProvider{responses: []string{"Hello!"}}
    agent := agentflow.NewBasicAgent("test", provider)
    
    response, err := agent.Process(context.Background(), "Hi")
    assert.NoError(t, err)
    assert.Equal(t, "Hello!", response)
}
```

## Next Steps

- Learn about [testing strategies](../contributors/testing/) for AgentFlow components
- Explore [adding new features](../contributors/adding-features/) to the framework
- Understand [performance considerations](performance/) for production deployments