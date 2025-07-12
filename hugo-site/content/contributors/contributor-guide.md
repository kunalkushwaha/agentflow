---
title: "Contributor Guide"
description: "Everything you need to know to start contributing to AgentFlow"
weight: 10
---

Welcome to AgentFlow! This guide will help you get started with contributing to the project, understanding the codebase, and following our development practices.

## Quick Start for Contributors

### 1. Development Setup

```bash
# Clone the repository
git clone https://github.com/kunalkushwaha/agentflow.git
cd agentflow

# Install dependencies
go mod tidy

# Run tests to ensure everything works
go test ./...

# Install development tools
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
```

### 2. Project Structure

```
agentflow/
├── cmd/                     # CLI tools
│   └── agentcli/           # AgentFlow CLI
├── core/                   # Public API (this is what users import)
│   ├── agent.go            # Agent interfaces
│   ├── mcp.go              # MCP integration public API
│   ├── factory.go          # Factory functions
│   └── ...
├── internal/               # Private implementation (not importable)
│   ├── agents/             # Agent implementations
│   ├── mcp/                # MCP client management
│   ├── llm/                # LLM provider implementations
│   ├── orchestrator/       # Workflow orchestration
│   └── scaffold/           # CLI project generation
├── examples/               # Example projects and demos
├── docs/                   # Documentation
└── benchmarks/             # Performance benchmarks
```

### 3. Core vs Internal Architecture

**`core/` Package (Public API):**
- This is what users import: `import "github.com/kunalkushwaha/agentflow/core"`
- Contains interfaces, types, and factory functions
- Must maintain backward compatibility
- All functions here should be well-documented and tested

**`internal/` Package (Private Implementation):**
- Implementation details that users don't need to know about
- Can change without breaking user code
- Contains concrete implementations of core interfaces
- Business logic and complex algorithms

**Example:**
```go
// core/agent.go - Public interface
type AgentHandler interface {
    Run(ctx context.Context, event Event, state State) (AgentResult, error)
}

// internal/agents/basic_agent.go - Private implementation  
type basicAgent struct {
    llm        llm.Provider
    mcpManager mcp.Manager
}

// core/factory.go - Public factory
func NewMCPAgent(name string, llm ModelProvider) AgentHandler {
    // Creates internal implementation but returns public interface
    return internal.NewBasicAgent(name, llm)
}
```

## Development Workflow

### 1. Feature Development Process

1. **Create Issue**: Discuss the feature/bug on GitHub Issues
2. **Fork & Branch**: Create a feature branch from `main`
3. **Develop**: Write code following our standards (see below)
4. **Test**: Add comprehensive tests for your changes
5. **Document**: Update documentation and examples
6. **PR**: Submit a pull request with clear description

### 2. Branch Naming

```bash
feature/mcp-server-discovery     # New features
fix/agent-error-handling         # Bug fixes
docs/api-reference-update        # Documentation
refactor/internal-mcp-client     # Code improvements
```

### 3. Commit Convention

```bash
git commit -m "feat(mcp): add server discovery caching"
git commit -m "fix(agent): handle nil state gracefully"  
git commit -m "docs(api): update MCP integration examples"
git commit -m "test(core): add agent factory unit tests"
```

**Types:** `feat`, `fix`, `docs`, `test`, `refactor`, `perf`, `ci`

## Code Standards

### 1. Go Code Style

Follow standard Go conventions plus AgentFlow-specific patterns:

```go
// Good: Clear interface with documentation
// AgentHandler processes events and manages agent state.
type AgentHandler interface {
    // Run executes the agent logic for the given event and state.
    // It returns the result and potentially modified state.
    Run(ctx context.Context, event Event, state State) (AgentResult, error)
}

// Good: Factory function with clear naming
func NewMCPAgent(name string, llm ModelProvider, mcp MCPManager) AgentHandler {
    return &mcpAgent{
        name:       name,
        llm:        llm,
        mcpManager: mcp,
    }
}

// Good: Error handling with context
func (a *mcpAgent) Run(ctx context.Context, event Event, state State) (AgentResult, error) {
    if event == nil {
        return AgentResult{}, fmt.Errorf("event cannot be nil")
    }
    
    message, ok := event.GetData()["message"]
    if !ok {
        return AgentResult{}, fmt.Errorf("missing required field: message")
    }
    
    // ... implementation
    
    return AgentResult{Result: response, State: state}, nil
}
```

### 2. Public API Guidelines

**Do:**
- Use clear, descriptive names
- Provide comprehensive documentation
- Include usage examples in godoc
- Return errors instead of panicking
- Use context.Context for cancellation
- Design for testability

**Don't:**
- Expose internal implementation details
- Break backward compatibility
- Use package-level global state
- Ignore errors
- Use `interface{}` unnecessarily

### 3. Testing Standards

Every public function needs tests:

```go
func TestNewMCPAgent(t *testing.T) {
    // Test successful creation
    llm := &MockModelProvider{}
    mcp := &MockMCPManager{}
    
    agent := core.NewMCPAgent("test-agent", llm, mcp)
    assert.NotNil(t, agent)
    
    // Test with nil parameters
    nilAgent := core.NewMCPAgent("", nil, nil)
    assert.NotNil(t, nilAgent) // Should handle gracefully
}

func TestMCPAgent_Run(t *testing.T) {
    tests := []struct {
        name        string
        event       core.Event
        expectedErr bool
        setup       func() (core.ModelProvider, core.MCPManager)
    }{
        {
            name:        "successful execution",
            event:       core.NewEvent("test", core.EventData{"message": "hello"}, nil),
            expectedErr: false,
            setup: func() (core.ModelProvider, core.MCPManager) {
                llm := &MockModelProvider{response: "Hello response"}
                mcp := &MockMCPManager{}
                return llm, mcp
            },
        },
        {
            name:        "missing message",
            event:       core.NewEvent("test", core.EventData{}, nil),
            expectedErr: true,
            setup: func() (core.ModelProvider, core.MCPManager) {
                return &MockModelProvider{}, &MockMCPManager{}
            },
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            llm, mcp := tt.setup()
            agent := core.NewMCPAgent("test", llm, mcp)
            state := core.NewState()
            
            result, err := agent.Run(context.Background(), tt.event, state)
            
            if tt.expectedErr {
                assert.Error(t, err)
            } else {
                assert.NoError(t, err)
                assert.NotEmpty(t, result.Result)
            }
        })
    }
}
```

## Architecture Decisions

### 1. Why core/ and internal/ ?

This separation provides:
- **Clear Public API**: Users know exactly what they can import
- **Implementation Freedom**: We can refactor internal code without breaking users
- **Better Testing**: Public API has different testing requirements than internal code
- **Documentation Focus**: We document the public API extensively

### 2. Interface-First Design

We define interfaces in `core/` and implementations in `internal/`:

```go
// core/mcp.go
type MCPManager interface {
    ListTools(ctx context.Context) ([]ToolSchema, error)
    CallTool(ctx context.Context, name string, args map[string]interface{}) (interface{}, error)
    // ... more methods
}

// internal/mcp/manager.go
type manager struct {
    clients map[string]*client.MCPClient
    cache   *toolCache
}

func (m *manager) ListTools(ctx context.Context) ([]ToolSchema, error) {
    // Implementation details here
}

// core/factory.go
func InitializeProductionMCP(ctx context.Context, config MCPConfig) (MCPManager, error) {
    return mcp.NewManager(config), nil // Returns internal implementation as interface
}
```

### 3. Factory Pattern Usage

We use factories to:
- Hide complex initialization
- Provide sensible defaults
- Allow easy testing with mocks
- Support configuration-driven setup

```go
// Simple factory
func NewMCPAgent(name string, llm ModelProvider) AgentHandler

// Configuration-driven factory  
func NewMCPAgentWithConfig(config AgentConfig) AgentHandler

// Auto-configuration factory
func NewMCPAgentFromWorkingDir() (AgentHandler, error)
```

## Adding New Features

### 1. Adding a New Agent Type

1. **Define Interface** (if needed) in `core/agent.go`
2. **Implement** in `internal/agents/your_agent.go`
3. **Add Factory** in `core/factory.go`
4. **Write Tests** for both interface and implementation
5. **Update Documentation** with examples

Example:
```go
// core/agent.go - Add to existing interfaces or create new one
type SpecializedAgent interface {
    AgentHandler
    GetSpecialization() string
}

// internal/agents/research_agent.go
type researchAgent struct {
    name           string
    llm            llm.Provider
    mcpManager     mcp.Manager
    specialization string
}

func (r *researchAgent) GetSpecialization() string {
    return r.specialization
}

func (r *researchAgent) Run(ctx context.Context, event core.Event, state core.State) (core.AgentResult, error) {
    // Implementation
}

// core/factory.go
func NewResearchAgent(name string, llm ModelProvider, mcp MCPManager) SpecializedAgent {
    return &agents.researchAgent{
        name:           name,
        llm:            llm,
        mcpManager:     mcp,
        specialization: "research",
    }
}
```

### 2. Adding a New LLM Provider

1. **Implement Interface** in `internal/llm/your_provider.go`
2. **Add Configuration** types in `core/llm.go`
3. **Add Factory** in `core/factory.go`
4. **Update CLI** scaffolding to support new provider
5. **Add Tests** and documentation

### 3. Adding MCP Features

1. **Update Interface** in `core/mcp.go`
2. **Implement** in `internal/mcp/`
3. **Update Helper Functions** like `FormatToolsForPrompt`
4. **Test Integration** with various MCP servers

## Testing Strategy

### 1. Test Categories

**Unit Tests**: Test individual functions and methods
```bash
go test ./core/...          # Test public API
go test ./internal/...      # Test implementations
```

**Integration Tests**: Test component interactions
```bash
go test -tags=integration ./...
```

**End-to-End Tests**: Test complete workflows
```bash
go test -tags=e2e ./...
```

### 2. Mock Strategy

We provide mocks for testing:

```go
// tests/mocks/mcp.go
type MockMCPManager struct {
    tools       []core.ToolSchema
    toolResults map[string]interface{}
}

func (m *MockMCPManager) ListTools(ctx context.Context) ([]core.ToolSchema, error) {
    return m.tools, nil
}

// Usage in tests
func TestAgentWithMCP(t *testing.T) {
    mockMCP := &mocks.MockMCPManager{
        tools: []core.ToolSchema{{Name: "search", Description: "Search tool"}},
        toolResults: map[string]interface{}{"search": "Mock results"},
    }
    
    agent := core.NewMCPAgent("test", mockLLM, mockMCP)
    // ... test logic
}
```

### 3. Benchmark Tests

For performance-critical code:

```go
func BenchmarkAgentRun(b *testing.B) {
    agent := core.NewMCPAgent("bench", mockLLM, mockMCP)
    event := core.NewEvent("test", core.EventData{"message": "benchmark"}, nil)
    state := core.NewState()
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        _, err := agent.Run(context.Background(), event, state)
        if err != nil {
            b.Fatal(err)
        }
    }
}
```

## Documentation Standards

### 1. Code Documentation

Every public function needs comprehensive godoc:

```go
// NewMCPAgent creates a new agent with MCP tool integration capabilities.
//
// The agent will automatically discover available tools from the MCP manager
// and include them in LLM prompts. Tool calls in LLM responses will be
// parsed and executed automatically.
//
// Parameters:
//   - name: Unique identifier for the agent
//   - llm: Model provider for generating responses
//   - mcp: MCP manager for tool discovery and execution
//
// Returns an AgentHandler that can process events with tool integration.
//
// Example:
//   provider, _ := core.NewAzureProvider(config)
//   mcpManager, _ := core.InitializeProductionMCP(ctx, mcpConfig)
//   agent := core.NewMCPAgent("research-agent", provider, mcpManager)
//
//   result, err := agent.Run(ctx, event, state)
func NewMCPAgent(name string, llm ModelProvider, mcp MCPManager) AgentHandler {
    // Implementation
}
```

### 2. User Documentation

When adding features, also update:
- Guides in `docs/guides/`
- API reference in `docs/api/`
- Examples in `examples/`
- Main README if it's a major feature

## Release Process

### 1. Version Management

We use semantic versioning:
- `v1.0.0` - Major release (breaking changes)
- `v1.1.0` - Minor release (new features)
- `v1.0.1` - Patch release (bug fixes)

### 2. Release Checklist

Before releasing:
- [ ] All tests pass
- [ ] Documentation is updated  
- [ ] Examples work with new version
- [ ] Breaking changes are documented
- [ ] Migration guide is provided (if needed)

### 3. Backward Compatibility

We maintain backward compatibility in the `core/` package. When making breaking changes:
1. Deprecate old functions (add `// Deprecated:` comment)
2. Provide new functions alongside old ones
3. Update documentation and examples
4. Remove deprecated functions only in major releases

## Getting Help

### 1. Development Questions

- **GitHub Discussions**: General questions about contributing
- **GitHub Issues**: Bug reports and feature requests
- **Code Reviews**: Ask questions in PR comments

### 2. Architecture Decisions

For major architectural changes:
1. Create GitHub Issue with `[RFC]` prefix
2. Discuss with maintainers
3. Create design document if needed
4. Implement after consensus

### 3. Code Style Questions

Follow existing patterns in the codebase. When in doubt:
- Look at similar functions in the same package
- Check `go fmt` and `golangci-lint` outputs
- Ask in PR comments

## Next Steps

Ready to contribute? Here are some good ways to get started:

1. **Look for "good first issue" labels** in the GitHub issues
2. **Improve documentation** - always appreciated!
3. **Add examples** - help other developers learn AgentFlow
4. **Fix bugs** - check the issue tracker for bugs to fix
5. **Add tests** - improve test coverage for existing code

We're excited to have you contribute to AgentFlow! 🚀