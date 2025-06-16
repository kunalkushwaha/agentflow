# AgentFlow Library Usage Guide

## Overview

AgentFlow is now available as a fully-featured Go library that you can import and use in your own projects. This guide provides comprehensive examples and best practices for using AgentFlow as an external dependency, including detailed coverage of MCP (Model Context Protocol) integration for enhanced tool capabilities.

## Installation

Add AgentFlow to your Go project:

```bash
go get github.com/kunalkushwaha/agentflow@latest
```

## Quick Start

### Basic Agent Implementation

```go
package main

import (
    "context"
    "fmt"
    "log"
    "time"

    agentflow "github.com/kunalkushwaha/agentflow/core"
)

// SimpleAgent demonstrates basic agent implementation
type SimpleAgent struct {
    name string
}

func NewSimpleAgent(name string) *SimpleAgent {
    return &SimpleAgent{name: name}
}

func (a *SimpleAgent) Run(ctx context.Context, event agentflow.Event, state agentflow.State) (agentflow.AgentResult, error) {
    startTime := time.Now()
    
    // Log agent execution
    agentflow.Logger().Info().
        Str("agent", a.name).
        Str("event_id", event.GetID()).
        Msg("Processing event")

    // Get data from event
    eventData := event.GetData()
    message, ok := eventData["message"]
    if !ok {
        message = "No message provided"
    }

    // Process the message
    response := fmt.Sprintf("%s processed: %v", a.name, message)

    // Create output state by cloning input state
    outputState := state.Clone()
    outputState.Set("response", response)
    outputState.Set("processed_by", a.name)
    outputState.Set("timestamp", time.Now().Format(time.RFC3339))

    endTime := time.Now()
    
    return agentflow.AgentResult{
        OutputState: outputState,
        StartTime:   startTime,
        EndTime:     endTime,
        Duration:    endTime.Sub(startTime),
    }, nil
}

func main() {
    // Set log level
    agentflow.SetLogLevel(agentflow.INFO)

    // Create agents
    agents := map[string]agentflow.AgentHandler{
        "processor": NewSimpleAgent("ProcessorAgent"),
    }

    // Create runner with configuration and tracing
    traceLogger := agentflow.NewInMemoryTraceLogger()
    runner := agentflow.NewRunnerWithConfig(agentflow.RunnerConfig{
        Agents:      agents,
        QueueSize:   10,
        TraceLogger: traceLogger,
    })

    // Start the runner
    ctx := context.Background()
    if err := runner.Start(ctx); err != nil {
        log.Fatalf("Failed to start runner: %v", err)
    }
    defer runner.Stop()

    fmt.Println("AgentFlow runner started successfully!")

    // Create and emit event with proper session handling
    sessionID := "library-test-session"
    eventData := agentflow.EventData{
        "message": "Hello from AgentFlow library!",
        "type":    "test_message",
    }

    metadata := map[string]string{
        agentflow.RouteMetadataKey: "processor",
        agentflow.SessionIDKey:     sessionID,
    }

    event := agentflow.NewEvent("processor", eventData, metadata)

    fmt.Printf("Emitting event: %s\n", event.GetID())
    if err := runner.Emit(event); err != nil {
        log.Fatalf("Failed to emit event: %v", err)
    }

    // Wait for processing
    time.Sleep(time.Second * 2)
    
    // Retrieve and display traces
    traces, err := runner.DumpTrace(sessionID)
    if err != nil {
        log.Printf("Error getting traces: %v", err)
    } else {
        fmt.Printf("Found %d trace entries for session %s\n", len(traces), sessionID)
    }
    
    fmt.Println("Event processing completed!")
}
```

## Advanced Features

### Circuit Breaker and Retry Logic

AgentFlow includes built-in circuit breaker and retry capabilities for resilient agent execution:

```go
// Agent with circuit breaker and retry configuration
type ResilientAgent struct {
    name string
}

func (a *ResilientAgent) Run(ctx context.Context, event agentflow.Event, state agentflow.State) (agentflow.AgentResult, error) {
    startTime := time.Now()
    
    // Simulate potentially failing operation
    if shouldSimulateFailure() {
        return agentflow.AgentResult{}, fmt.Errorf("simulated failure for circuit breaker testing")
    }
    
    outputState := state.Clone()
    outputState.Set("processed_by", a.name)
    outputState.Set("success", true)
    
    return agentflow.AgentResult{
        OutputState: outputState,
        StartTime:   startTime,
        EndTime:     time.Now(),
        Duration:    time.Since(startTime),
    }, nil
}

// Create runner with circuit breaker configuration
func createResilientRunner() *agentflow.Runner {
    agents := map[string]agentflow.AgentHandler{
        "resilient": &ResilientAgent{name: "ResilientAgent"},
    }
    
    return agentflow.NewRunnerWithConfig(agentflow.RunnerConfig{
        Agents:        agents,
        QueueSize:     10,
        TraceLogger:   agentflow.NewInMemoryTraceLogger(),
    })
}
```

### Responsible AI Integration

Implement responsible AI checks in your workflows:

```go
// ResponsibleAIAgent performs content safety and ethical checks
type ResponsibleAIAgent struct{}

func (a *ResponsibleAIAgent) Run(ctx context.Context, event agentflow.Event, state agentflow.State) (agentflow.AgentResult, error) {
    startTime := time.Now()
    
    // Get content to check
    content, exists := event.GetData()["content"]
    if !exists {
        return agentflow.AgentResult{}, fmt.Errorf("no content provided for responsible AI check")
    }
    
    // Perform safety checks (placeholder implementation)
    if isContentUnsafe(content) {
        outputState := state.Clone()
        outputState.Set("responsible_ai_check", "failed")
        outputState.Set("reason", "unsafe content detected")
        outputState.SetMeta(agentflow.RouteMetadataKey, "error-handler")
        
        return agentflow.AgentResult{
            OutputState: outputState,
            StartTime:   startTime,
            EndTime:     time.Now(),
            Duration:    time.Since(startTime),
        }, nil
    }
    
    // Content is safe, continue processing
    outputState := state.Clone()
    outputState.Set("responsible_ai_check", "passed")
    outputState.SetMeta(agentflow.RouteMetadataKey, "content-processor")
    
    return agentflow.AgentResult{
        OutputState: outputState,
        StartTime:   startTime,
        EndTime:     time.Now(),
        Duration:    time.Since(startTime),
    }, nil
}

func isContentUnsafe(content interface{}) bool {
    // Implement your content safety logic here
    contentStr, ok := content.(string)
    if !ok {
        return true
    }
    // Simple example: check for harmful keywords
    harmfulKeywords := []string{"violence", "hate", "harmful"}
    for _, keyword := range harmfulKeywords {
        if strings.Contains(strings.ToLower(contentStr), keyword) {
            return true
        }
    }
    return false
}
```

### Enhanced Error Routing

Implement sophisticated error handling with routing:

```go
// ErrorRoutingAgent demonstrates enhanced error handling
type ErrorRoutingAgent struct{}

func (a *ErrorRoutingAgent) Run(ctx context.Context, event agentflow.Event, state agentflow.State) (agentflow.AgentResult, error) {
    startTime := time.Now()
    
    // Get error information from state
    errorType, _ := state.Get("error_type")
    errorMessage, _ := state.Get("error_message")
    
    outputState := state.Clone()
    
    // Route based on error type
    switch errorType {
    case "validation_error":
        outputState.Set("recovery_action", "request_user_input")
        outputState.SetMeta(agentflow.RouteMetadataKey, "validation-handler")
    case "timeout_error":
        outputState.Set("recovery_action", "retry_with_backoff")
        outputState.SetMeta(agentflow.RouteMetadataKey, "retry-handler")
    case "critical_error":
        outputState.Set("recovery_action", "escalate_to_admin")
        outputState.SetMeta(agentflow.RouteMetadataKey, "admin-notifier")
    default:
        outputState.Set("recovery_action", "generic_error_handling")
    }
    
    outputState.Set("error_handled", true)
    outputState.Set("handled_by", "ErrorRoutingAgent")
    outputState.Set("original_error", errorMessage)
    
    return agentflow.AgentResult{
        OutputState: outputState,
        StartTime:   startTime,
        EndTime:     time.Now(),
        Duration:    time.Since(startTime),
    }, nil
}
```

### Workflow Validation

Implement workflow validation for complex agent chains:

```go
// WorkflowValidationAgent validates workflow state and transitions
type WorkflowValidationAgent struct{}

func (a *WorkflowValidationAgent) Run(ctx context.Context, event agentflow.Event, state agentflow.State) (agentflow.AgentResult, error) {
    startTime := time.Now()
    
    // Validate required workflow fields
    requiredFields := []string{"workflow_id", "step", "user_id"}
    for _, field := range requiredFields {
        if _, exists := state.Get(field); !exists {
            return agentflow.AgentResult{}, fmt.Errorf("missing required field: %s", field)
        }
    }
    
    // Validate workflow step progression
    currentStep, _ := state.Get("step")
    if !isValidStepTransition(currentStep) {
        outputState := state.Clone()
        outputState.Set("validation_error", "invalid step transition")
        outputState.SetMeta(agentflow.RouteMetadataKey, "error-handler")
        
        return agentflow.AgentResult{
            OutputState: outputState,
            StartTime:   startTime,
            EndTime:     time.Now(),
            Duration:    time.Since(startTime),
        }, nil
    }
    
    // Validation passed
    outputState := state.Clone()
    outputState.Set("validation_status", "passed")
    outputState.Set("validated_at", time.Now().Format(time.RFC3339))
    
    return agentflow.AgentResult{
        OutputState: outputState,
        StartTime:   startTime,
        EndTime:     time.Now(),
        Duration:    time.Since(startTime),
    }, nil
}

func isValidStepTransition(step interface{}) bool {
    // Implement your workflow validation logic
    stepStr, ok := step.(string)
    if !ok {
        return false
    }
    validSteps := map[string]bool{
        "init": true, "process": true, "validate": true, "complete": true,
    }
    return validSteps[stepStr]
}
```

### Agent Chaining

Chain multiple agents together to create complex workflows:

```go
// ChainedAgent demonstrates agent chaining
type ChainedAgent struct {
    name     string
    nextAgent string
    step     int
}

func (a *ChainedAgent) Run(ctx context.Context, event agentflow.Event, state agentflow.State) (agentflow.AgentResult, error) {
    agentflow.Logger().Info().
        Str("agent", a.name).
        Str("event_id", event.GetID()).
        Int("step", a.step).
        Msg("ChainedAgent processing")

    // Process current step
    outputState := state.Clone()
    outputState.Set(fmt.Sprintf("step_%d_completed", a.step), true)
    outputState.Set("current_step", a.step)

    result := agentflow.AgentResult{
        OutputState: outputState,
        StartTime:   time.Now(),
        EndTime:     time.Now(),
        Duration:    time.Millisecond * 100,
    }

    // Chain to next agent if specified
    if a.nextAgent != "" && a.step < 3 {
        agentflow.Logger().Info().
            Str("agent", a.name).
            Str("next_agent", a.nextAgent).
            Int("step", a.step).
            Msg("Routing to next agent")
        
        result.OutputState.SetMeta(agentflow.RouteMetadataKey, a.nextAgent)
    } else {
        agentflow.Logger().Info().
            Str("agent", a.name).
            Int("final_step", a.step).
            Msg("Chain completed")
    }

    return result, nil
}

// Usage
func createChainedAgents() map[string]agentflow.AgentHandler {
    return map[string]agentflow.AgentHandler{
        "step1": &ChainedAgent{name: "Step1Agent", nextAgent: "step2", step: 1},
        "step2": &ChainedAgent{name: "Step2Agent", nextAgent: "step3", step: 2},
        "step3": &ChainedAgent{name: "Step3Agent", nextAgent: "", step: 3},
    }
}
```

### Error Handling

Implement robust error handling and recovery:

```go
type ErrorHandlingAgent struct {
    shouldError bool
}

func (a *ErrorHandlingAgent) Run(ctx context.Context, event agentflow.Event, state agentflow.State) (agentflow.AgentResult, error) {
    if a.shouldError {
        return agentflow.AgentResult{}, fmt.Errorf("simulated agent error")
    }

    // Normal processing
    outputState := agentflow.NewState()
    outputState.Set("success", true)
    
    return agentflow.AgentResult{
        OutputState: outputState,
        StartTime:   time.Now(),
        EndTime:     time.Now(),
        Duration:    time.Millisecond * 50,
    }, nil
}

// Error handler agent
type ErrorHandlerAgent struct{}

func (a *ErrorHandlerAgent) Run(ctx context.Context, event agentflow.Event, state agentflow.State) (agentflow.AgentResult, error) {
    agentflow.Logger().Error().
        Str("event_id", event.GetID()).
        Msg("ErrorHandlerAgent handling error")

    outputState := agentflow.NewState()
    outputState.Set("error_handled", true)
    outputState.Set("recovery_time", time.Now().Format(time.RFC3339))

    return agentflow.AgentResult{
        OutputState: outputState,
        StartTime:   time.Now(),
        EndTime:     time.Now(),
        Duration:    time.Millisecond * 5,
    }, nil
}
```

### Tracing and Observability

Enable comprehensive tracing for debugging and monitoring:

```go
func createRunnerWithTracing() *agentflow.Runner {
    // Create in-memory trace logger
    traceLogger := agentflow.NewInMemoryTraceLogger()

    // Create runner with tracing enabled
    runner := agentflow.NewRunnerWithConfig(agentflow.RunnerConfig{
        Agents:      agents,
        QueueSize:   10,
        TraceLogger: traceLogger,
    })

    return runner
}

func analyzeTraces(runner *agentflow.Runner, sessionID string) {
    traces, err := runner.DumpTrace(sessionID)
    if err != nil {
        fmt.Printf("Error getting traces: %v\n", err)
        return
    }

    fmt.Printf("Found %d trace entries:\n", len(traces))
    for i, trace := range traces {
        fmt.Printf("  %d. %s: Type=%s, EventID=%s, AgentID=%s\n",
            i+1,
            trace.Timestamp.Format("15:04:05.000"),
            trace.Type,
            trace.EventID,
            trace.AgentID)
    }
}
```

### Concurrent Processing

Handle multiple events concurrently:

```go
func concurrentProcessing(runner *agentflow.Runner) {
    var wg sync.WaitGroup
    
    // Emit multiple events concurrently
    for i := 0; i < 5; i++ {
        wg.Add(1)
        go func(id int) {
            defer wg.Done()
            
            eventData := agentflow.EventData{
                "message": fmt.Sprintf("Concurrent event %d", id),
                "id":      id,
            }
            
            metadata := map[string]string{
                agentflow.RouteMetadataKey: "processor",
                "session_id":               fmt.Sprintf("concurrent-session-%d", id),
            }
            
            event := agentflow.NewEvent("processor", eventData, metadata)
            
            if err := runner.Emit(event); err != nil {
                fmt.Printf("Failed to emit event %d: %v\n", id, err)
            }
        }(i)
    }
      wg.Wait()
    fmt.Println("All concurrent events emitted")
}
```

## MCP Integration (Model Context Protocol)

AgentFlow's MCP integration allows your library-based applications to leverage external tools and services through the Model Context Protocol. This section covers practical patterns for integrating MCP into your projects.

### Quick MCP Setup for Libraries

The simplest way to add MCP support to your library-based AgentFlow project:

```go
package main

import (
    "context"
    "fmt"
    "log"
    "os"
    
    agentflow "github.com/kunalkushwaha/agentflow/core"
)

func main() {
    // Quick MCP-enabled setup
    ctx := context.Background()
    
    // Initialize with MCP support using environment-based configuration
    mcpConfig := agentflow.MCPConfig{
        Enabled: true,
        Servers: []agentflow.MCPServerConfig{
            {
                Name:    "filesystem",
                Command: "node",
                Args:    []string{os.Getenv("MCP_FILESYSTEM_SERVER_PATH"), "./data"},
                Env:     map[string]string{"NODE_ENV": "production"},
            },
        },
    }
    
    // Create agents with MCP awareness
    agents := map[string]agentflow.AgentHandler{
        "file-processor": &FileProcessorAgent{},
        "data-analyzer":  &DataAnalyzerAgent{},
    }
    
    // Build runner with MCP support
    runner := agentflow.NewRunnerWithConfig(agentflow.RunnerConfig{
        Agents:     agents,
        QueueSize:  100,
        MCPConfig:  &mcpConfig,
        TraceLogger: agentflow.NewInMemoryTraceLogger(),
    })
    
    // Start the runner
    if err := runner.Start(ctx); err != nil {
        log.Fatal("Failed to start runner:", err)
    }
    defer runner.Stop()
    
    // MCP tools are now available to agents
    fmt.Println("AgentFlow with MCP support is running...")
}
```

### MCP-Aware Agent Implementation

Create agents that can leverage MCP tools while maintaining fallback capabilities:

```go
type FileProcessorAgent struct {
    toolRegistry agentflow.ToolRegistry
}

func (a *FileProcessorAgent) Run(ctx context.Context, event agentflow.Event, state agentflow.State) (agentflow.AgentResult, error) {
    startTime := time.Now()
    
    // Extract file path from event
    eventData := event.GetData()
    filePath, ok := eventData["file_path"].(string)
    if !ok {
        return agentflow.AgentResult{}, fmt.Errorf("file_path not provided")
    }
      // Try to use MCP filesystem tool first
    var fileContent string
    var dataSource string
    
    if a.toolRegistry.HasTool("filesystem_read_file") {
        result, err := a.toolRegistry.CallTool(ctx, "filesystem_read_file", map[string]any{
            "path": filePath,
        })
        
        if err != nil {
            // MCP tool failed - log and use fallback
            log.Printf("MCP filesystem tool failed: %v, using fallback", err)
            fileContent, err = a.readFileDirectly(filePath)
            if err != nil {
                return agentflow.AgentResult{}, fmt.Errorf("both MCP and fallback failed: %w", err)
            }
            dataSource = "fallback_direct_read"
        } else {
            fileContent = result["content"].(string)
            dataSource = "mcp_filesystem"
        }
    } else {
        // MCP tool not available - use direct fallback
        var err error
        fileContent, err = a.readFileDirectly(filePath)
        if err != nil {
            return agentflow.AgentResult{}, fmt.Errorf("file read failed: %w", err)
        }
        dataSource = "direct_read"
    }
    
    // Process file content
    processedData := a.processContent(fileContent)
    
    // Update state with results
    outputState := state.Clone()
    outputState.Set("file_content", fileContent)
    outputState.Set("processed_data", processedData)
    outputState.Set("data_source", dataSource)
    outputState.Set("file_path", filePath)
    
    return agentflow.AgentResult{
        OutputState: outputState,
        StartTime:   startTime,
        EndTime:     time.Now(),
        Duration:    time.Since(startTime),
    }, nil
}

func (a *FileProcessorAgent) readFileDirectly(path string) (string, error) {
    // Fallback implementation
    data, err := os.ReadFile(path)
    if err != nil {
        return "", err
    }
    return string(data), nil
}

func (a *FileProcessorAgent) processContent(content string) map[string]interface{} {
    // Your content processing logic
    return map[string]interface{}{
        "word_count": len(strings.Fields(content)),
        "char_count": len(content),
        "lines":      len(strings.Split(content, "\n")),
    }
}
```

### Library Integration Patterns

#### Pattern 1: Configuration-Driven MCP Setup

Create a configuration system for your library users:

```go
// config.go - Your library's configuration
type AppConfig struct {
    AgentFlow *AgentFlowConfig `toml:"agentflow"`
    MCP       *MCPConfig       `toml:"mcp"`
}

type AgentFlowConfig struct {
    QueueSize   int    `toml:"queue_size"`
    LogLevel    string `toml:"log_level"`
    TraceToFile bool   `toml:"trace_to_file"`
}

type MCPConfig struct {
    Enabled bool                `toml:"enabled"`
    Servers []MCPServerConfig   `toml:"servers"`
}

type MCPServerConfig struct {
    Name    string            `toml:"name"`
    Command string            `toml:"command"`
    Args    []string          `toml:"args"`
    Env     map[string]string `toml:"env"`
}

// Initialize AgentFlow with user configuration
func NewAgentFlowFromConfig(configPath string) (*agentflow.Runner, error) {
    var config AppConfig
    if _, err := toml.DecodeFile(configPath, &config); err != nil {
        return nil, fmt.Errorf("failed to load config: %w", err)
    }
    
    // Convert to AgentFlow configuration
    mcpConfig := &agentflow.MCPConfig{
        Enabled: config.MCP.Enabled,
        Servers: make([]agentflow.MCPServerConfig, len(config.MCP.Servers)),
    }
    
    for i, server := range config.MCP.Servers {
        mcpConfig.Servers[i] = agentflow.MCPServerConfig{
            Name:    server.Name,
            Command: server.Command,
            Args:    server.Args,
            Env:     server.Env,
        }
    }
    
    // Create agents for your application
    agents := createApplicationAgents()
    
    runnerConfig := agentflow.RunnerConfig{
        Agents:    agents,
        QueueSize: config.AgentFlow.QueueSize,
        MCPConfig: mcpConfig,
    }
    
    if config.AgentFlow.TraceToFile {
        runnerConfig.TraceLogger = agentflow.NewFileTraceLogger("./traces")
    } else {
        runnerConfig.TraceLogger = agentflow.NewInMemoryTraceLogger()
    }
    
    return agentflow.NewRunnerWithConfig(runnerConfig), nil
}
```

Users can then configure your library with:

```toml
# app.toml - User configuration
[agentflow]
queue_size = 100
log_level = "info"
trace_to_file = true

[mcp]
enabled = true

[[mcp.servers]]
name = "filesystem"
command = "node"
args = ["filesystem-server.js", "/home/user/documents"]
env = { NODE_ENV = "production" }

[[mcp.servers]]
name = "web-search"
command = "python"
args = ["-m", "web_search_server"]
env = { SEARCH_API_KEY = "${SEARCH_API_KEY}" }
```

#### Pattern 2: Tool Registry Abstraction

Provide a clean interface for your library users to access MCP tools:

```go
// Your library's tool interface
type ToolProvider interface {
    GetTool(name string) Tool
    ListAvailableTools() []string
    IsToolAvailable(name string) bool
}

type Tool interface {
    Execute(ctx context.Context, params map[string]interface{}) (map[string]interface{}, error)
    Description() string
    Schema() map[string]interface{}
}

// Implementation that wraps AgentFlow's MCP tools
type AgentFlowToolProvider struct {
    registry agentflow.ToolRegistry
}

func (p *AgentFlowToolProvider) GetTool(name string) Tool {
    return &ToolWrapper{tool: p.registry.GetTool(name)}
}

func (p *AgentFlowToolProvider) ListAvailableTools() []string {
    return p.registry.ListTools()
}

func (p *AgentFlowToolProvider) IsToolAvailable(name string) bool {
    return p.registry.GetTool(name) != nil
}

// Your library's public API
func (app *YourApp) GetToolProvider() ToolProvider {
    return &AgentFlowToolProvider{
        registry: app.agentFlowRunner.ToolRegistry(),
    }
}
```

#### Pattern 3: Health Monitoring Integration

Integrate MCP health monitoring into your application's health checks:

```go
type HealthChecker struct {
    runner agentflow.Runner
}

func (h *HealthChecker) CheckHealth() map[string]interface{} {
    health := map[string]interface{}{
        "status":    "healthy",
        "timestamp": time.Now().Unix(),
        "checks":    make(map[string]interface{}),
    }
    
    // Check AgentFlow runner health
    if h.runner == nil {
        health["status"] = "unhealthy"
        health["checks"]["agentflow"] = map[string]interface{}{
            "status": "down",
            "error":  "runner not initialized",
        }
        return health
    }
    
    health["checks"]["agentflow"] = map[string]interface{}{
        "status": "up",
    }
    
    // Check MCP server health if available
    if mcpManager := h.runner.MCPManager(); mcpManager != nil {
        mcpHealth := map[string]interface{}{}
        serverStatus := mcpManager.GetServerStatus()
        
        allHealthy := true
        for serverName, status := range serverStatus {
            serverHealth := map[string]interface{}{
                "healthy":    status.Healthy,
                "tools":      len(status.AvailableTools),
                "last_check": status.LastHealthCheck.Unix(),
            }
            
            if !status.Healthy {
                serverHealth["error"] = status.Error
                allHealthy = false
            }
            
            mcpHealth[serverName] = serverHealth
        }
        
        health["checks"]["mcp"] = map[string]interface{}{
            "status":  map[bool]string{true: "up", false: "degraded"}[allHealthy],
            "servers": mcpHealth,
        }
        
        if !allHealthy {
            health["status"] = "degraded"
        }
    }
    
    return health
}
```

### Performance Considerations for Library Usage

When using MCP in library contexts, consider these performance optimizations:

#### 1. Tool Caching
```go
type CachedToolProvider struct {
    provider    ToolProvider
    cache       map[string]CachedResult
    cacheTTL    time.Duration
    mu          sync.RWMutex
}

func (c *CachedToolProvider) Execute(toolName string, params map[string]interface{}) (map[string]interface{}, error) {
    // Check cache first for idempotent operations
    if result := c.getCachedResult(toolName, params); result != nil {
        return result.Data, nil
    }
    
    // Execute tool and cache result
    result, err := c.provider.GetTool(toolName).Execute(ctx, params)
    if err != nil {
        return nil, err
    }
    
    c.cacheResult(toolName, params, result)
    return result, nil
}
```

#### 2. Connection Pooling
```go
// Pre-warm MCP connections at startup
func (app *YourApp) warmupMCPConnections(ctx context.Context) error {
    if mcpManager := app.runner.MCPManager(); mcpManager != nil {
        // Test all configured servers
        servers := mcpManager.ListServers()
        for _, serverName := range servers {
            if err := mcpManager.PingServer(ctx, serverName); err != nil {
                log.Printf("Warning: MCP server %s not responding: %v", serverName, err)
            }
        }
    }
    return nil
}
```

#### 3. Graceful Degradation
```go
type ResilientService struct {
    toolProvider ToolProvider
    fallbacks    map[string]func(map[string]interface{}) (map[string]interface{}, error)
}

func (s *ResilientService) ProcessData(data interface{}) (interface{}, error) {
    // Try MCP tool first
    if tool := s.toolProvider.GetTool("data_processor"); tool != nil {
        result, err := tool.Execute(ctx, map[string]interface{}{"data": data})
        if err == nil {
            return result, nil
        }
        log.Printf("MCP tool failed: %v, using fallback", err)
    }
    
    // Use fallback implementation
    if fallback, exists := s.fallbacks["data_processor"]; exists {
        return fallback(map[string]interface{}{"data": data})
    }
    
    return nil, fmt.Errorf("no processing method available")
}
```

## Configuration Options

### Runner Configuration

```go
// Comprehensive RunnerConfig options
runner := agentflow.NewRunnerWithConfig(agentflow.RunnerConfig{
    Agents:      agents,                                    // Map of agent handlers
    QueueSize:   100,                                      // Event queue size
    TraceLogger: agentflow.NewInMemoryTraceLogger(),       // In-memory tracing
    // Or use file-based tracing for production:
    // TraceLogger: agentflow.NewFileTraceLogger("./traces"),
})
```

### Logging Configuration

```go
// Set log level
agentflow.SetLogLevel(agentflow.DEBUG) // DEBUG, INFO, WARN, ERROR

// Get logger for custom logging
logger := agentflow.Logger()
logger.Info().
    Str("session_id", sessionID).
    Str("agent", "my-agent").
    Msg("Custom log message")
```

### Key Constants and Metadata Keys

AgentFlow provides several important constants for metadata handling:

```go
// Standard metadata keys
const (
    RouteMetadataKey = "route_to"     // Agent routing
    SessionIDKey     = "session_id"   // Session tracking
)

// Usage in event creation
metadata := map[string]string{
    agentflow.RouteMetadataKey: "target-agent",
    agentflow.SessionIDKey:     "session-123",
    "custom_key":               "custom_value",
}

// Access in agents
func (a *MyAgent) Run(ctx context.Context, event agentflow.Event, state agentflow.State) (agentflow.AgentResult, error) {
    // Get routing information
    targetAgent, exists := event.GetMetadataValue(agentflow.RouteMetadataKey)
    if exists {
        fmt.Printf("Event routed to: %s\n", targetAgent)
    }
    
    // Get session ID for tracing
    sessionID, exists := event.GetMetadataValue(agentflow.SessionIDKey)
    if exists {
        fmt.Printf("Session ID: %s\n", sessionID)
    }
    
    // Set routing for next agent
    outputState := state.Clone()
    outputState.SetMeta(agentflow.RouteMetadataKey, "next-agent")
    
    return agentflow.AgentResult{OutputState: outputState}, nil
}
```

### Trace Logger Options

```go
// In-memory trace logger (for development/testing)
memoryLogger := agentflow.NewInMemoryTraceLogger()

// File-based trace logger (for production)
fileLogger := agentflow.NewFileTraceLogger("./application-traces")

// Using with runner
runner := agentflow.NewRunnerWithConfig(agentflow.RunnerConfig{
    Agents:      agents,
    TraceLogger: fileLogger, // or memoryLogger
})

// Retrieve traces
traces, err := runner.DumpTrace(sessionID)
if err != nil {
    log.Printf("Error retrieving traces: %v", err)
    return
}

// Analyze trace entries
for _, trace := range traces {
    fmt.Printf("Trace: %s - Agent: %s - Type: %s - Duration: %v\n",
        trace.Timestamp.Format("15:04:05.000"),
        trace.AgentID,
        trace.Type,
        trace.Duration)
}
```

## Best Practices

### 1. Agent Design

- **Single Responsibility**: Keep agents focused on one specific task
- **Use Meaningful Names**: Choose descriptive names for agents and state keys
- **Proper Error Handling**: Implement comprehensive error handling with routing
- **Include Comprehensive Logging**: Use structured logging with context
- **State Cloning**: Always clone state to avoid mutations affecting other agents

```go
func (a *MyAgent) Run(ctx context.Context, event agentflow.Event, state agentflow.State) (agentflow.AgentResult, error) {
    startTime := time.Now()
    
    // Always log execution start
    agentflow.Logger().Info().
        Str("agent", a.name).
        Str("event_id", event.GetID()).
        Msg("Agent execution started")
    
    // Always clone state
    outputState := state.Clone()
    
    // Your processing logic...
    
    // Always log completion with timing
    agentflow.Logger().Info().
        Str("agent", a.name).
        Dur("duration", time.Since(startTime)).
        Msg("Agent execution completed")
    
    return agentflow.AgentResult{
        OutputState: outputState,
        StartTime:   startTime,
        EndTime:     time.Now(),
        Duration:    time.Since(startTime),
    }, nil
}
```

### 2. State Management

- **Clone Before Modify**: Use `state.Clone()` to avoid side effects
- **Consistent Key Naming**: Use descriptive, consistent naming conventions
- **Include Metadata**: Use metadata for routing and debugging information
- **Validate Required Fields**: Check for required data before processing

```go
// Good state management patterns
outputState := state.Clone()

// Check for required fields
userID, exists := state.Get("user_id")
if !exists {
    return agentflow.AgentResult{}, fmt.Errorf("missing required field: user_id")
}

// Use descriptive keys
outputState.Set("processing_result", result)
outputState.Set("processed_by_agent", a.name)
outputState.Set("processing_timestamp", time.Now().Format(time.RFC3339))

// Set routing metadata
outputState.SetMeta(agentflow.RouteMetadataKey, "next-agent")
outputState.SetMeta("processing_stage", "validation")
```

### 3. Event Handling

- **Use Unique Session IDs**: Essential for tracing and debugging
- **Include Relevant Metadata**: Add routing and context information
- **Handle Context Cancellation**: Respect context cancellation signals
- **Validate Event Data**: Verify event data before processing

```go
// Create events with proper session handling
sessionID := fmt.Sprintf("workflow-%s-%d", workflowType, time.Now().UnixNano())

event := agentflow.NewEvent("source-system", agentflow.EventData{
    "task_type":    "analysis",
    "user_prompt":  userInput,
    "priority":     "high",
}, map[string]string{
    agentflow.RouteMetadataKey: "analyzer-agent",
    agentflow.SessionIDKey:     sessionID,
    "workflow_type":            workflowType,
    "user_id":                  userID,
})

// In agent, handle context cancellation
func (a *MyAgent) Run(ctx context.Context, event agentflow.Event, state agentflow.State) (agentflow.AgentResult, error) {
    select {
    case <-ctx.Done():
        return agentflow.AgentResult{}, ctx.Err()
    default:
        // Continue processing
    }
    
    // Validate event data
    taskType, exists := event.GetData()["task_type"]
    if !exists {
        return agentflow.AgentResult{}, fmt.Errorf("missing task_type in event data")
    }
    
    // Your processing logic...
}
```

### 4. Error Handling and Resilience

- **Implement Retry Logic**: For transient failures
- **Use Circuit Breaker Patterns**: Prevent cascade failures
- **Route Errors Appropriately**: Use error routing for recovery
- **Log Errors with Context**: Include sufficient debugging information

```go
type ResilientAgent struct {
    name       string
    maxRetries int
    retryDelay time.Duration
}

func (a *ResilientAgent) Run(ctx context.Context, event agentflow.Event, state agentflow.State) (agentflow.AgentResult, error) {
    var lastErr error
    
    for attempt := 0; attempt <= a.maxRetries; attempt++ {
        if attempt > 0 {
            agentflow.Logger().Warn().
                Str("agent", a.name).
                Int("attempt", attempt).
                Err(lastErr).
                Msg("Retrying agent execution")
                
            select {
            case <-ctx.Done():
                return agentflow.AgentResult{}, ctx.Err()
            case <-time.After(a.retryDelay):
            }
        }
        
        result, err := a.processEvent(ctx, event, state)
        if err == nil {
            return result, nil
        }
        
        lastErr = err
        if !a.isRetryableError(err) {
            break
        }
    }
    
    // Route to error handler
    outputState := state.Clone()
    outputState.Set("error_message", lastErr.Error())
    outputState.Set("failed_agent", a.name)
    outputState.SetMeta(agentflow.RouteMetadataKey, "error-handler")
    
    return agentflow.AgentResult{OutputState: outputState}, nil
}
```

### 5. Performance and Monitoring

- **Configure Appropriate Queue Sizes**: Balance memory usage and throughput
- **Monitor Trace Logs**: Use traces for bottleneck identification
- **Implement Custom Metrics**: Track business-relevant metrics
- **Use File-Based Tracing in Production**: For persistent trace storage

```go
// Production configuration
runner := agentflow.NewRunnerWithConfig(agentflow.RunnerConfig{
    Agents:      agents,
    QueueSize:   1000,  // Larger queue for production
    TraceLogger: agentflow.NewFileTraceLogger("./production-traces"),
})

// Custom metrics collection
registry := runner.CallbackRegistry()
registry.Register(agentflow.HookAfterAgentRun, "metrics-collector", func(ctx context.Context, event agentflow.Event, state agentflow.State) error {
    // Collect custom metrics
    agentName := event.GetTargetAgentID()
    duration := state.Get("execution_duration")
    
    // Send to your metrics system
    metricsCollector.RecordAgentExecution(agentName, duration)
    return nil
})

// Performance analysis
func analyzePerformance(runner *agentflow.Runner, sessionID string) {
    traces, _ := runner.DumpTrace(sessionID)
    
    var totalDuration time.Duration
    agentCounts := make(map[string]int)
    
    for _, trace := range traces {
        if trace.Type == "agent_end" {
            totalDuration += trace.Duration
            agentCounts[trace.AgentID]++
        }
    }
    
    fmt.Printf("Performance Summary:\n")
    fmt.Printf("Total Duration: %v\n", totalDuration)
    for agent, count := range agentCounts {
        fmt.Printf("  %s: %d executions\n", agent, count)
    }
}
```

### 6. Testing

- **Use In-Memory Trace Loggers**: Perfect for unit tests
- **Create Test Agents**: Implement simple agents for testing
- **Test Error Scenarios**: Verify error handling and recovery
- **Validate State Transitions**: Ensure proper state management

```go
func TestAgentWorkflow(t *testing.T) {
    // Create test agents
    agents := map[string]agentflow.AgentHandler{
        "test-agent": &TestAgent{shouldSucceed: true},
        "error-handler": &TestErrorHandler{},
    }
    
    // Use in-memory tracing for tests
    runner := agentflow.NewRunnerWithConfig(agentflow.RunnerConfig{
        Agents:      agents,
        QueueSize:   10,
        TraceLogger: agentflow.NewInMemoryTraceLogger(),
    })
    
    ctx := context.Background()
    err := runner.Start(ctx)
    require.NoError(t, err)
    defer runner.Stop()
    
    // Test successful execution
    sessionID := "test-session-123"
    event := agentflow.NewEvent("test", agentflow.EventData{
        "test_data": "example",
    }, map[string]string{
        agentflow.RouteMetadataKey: "test-agent",
        agentflow.SessionIDKey:     sessionID,
    })
    
    err = runner.Emit(event)
    require.NoError(t, err)
    
    // Wait and verify traces
    time.Sleep(100 * time.Millisecond)
    traces, err := runner.DumpTrace(sessionID)
    require.NoError(t, err)    require.Greater(t, len(traces), 0)
}
```

### 7. MCP Integration Best Practices

When using MCP (Model Context Protocol) in your library-based AgentFlow applications, follow these best practices:

#### Tool Selection and Fallbacks
```go
// Always implement fallback strategies for MCP tools
type MCPAwareAgent struct {
    toolRegistry agentflow.ToolRegistry
    fallbackImpl map[string]func(context.Context, map[string]interface{}) (map[string]interface{}, error)
}

func (a *MCPAwareAgent) executeWithFallback(ctx context.Context, toolName string, params map[string]interface{}) (map[string]interface{}, error) {
    // Try MCP tool first
    if tool := a.toolRegistry.GetTool(toolName); tool != nil {
        result, err := tool.Execute(ctx, params)
        if err == nil {
            return result, nil
        }
        
        // Log MCP failure but continue
        agentflow.Logger().Warn().
            Err(err).
            Str("tool", toolName).
            Msg("MCP tool failed, using fallback")
    }
    
    // Use fallback implementation
    if fallback, exists := a.fallbackImpl[toolName]; exists {
        return fallback(ctx, params)
    }
    
    return nil, fmt.Errorf("tool %s not available and no fallback implemented", toolName)
}
```

#### Configuration Management
```go
// Separate MCP configuration from core application config
type MCPConfiguration struct {
    Enabled              bool                      `json:"enabled"`
    Servers              []MCPServerConfig         `json:"servers"`
    HealthCheckInterval  time.Duration             `json:"health_check_interval"`
    ConnectionTimeout    time.Duration             `json:"connection_timeout"`
    FallbackBehavior     string                    `json:"fallback_behavior"` // "fail", "warn", "silent"
}

// Load configuration with validation
func LoadMCPConfig(path string) (*MCPConfiguration, error) {
    config := &MCPConfiguration{
        HealthCheckInterval: 60 * time.Second,
        ConnectionTimeout:   10 * time.Second,
        FallbackBehavior:    "warn",
    }
    
    if data, err := os.ReadFile(path); err == nil {
        if err := json.Unmarshal(data, config); err != nil {
            return nil, fmt.Errorf("invalid MCP configuration: %w", err)
        }
    }
    
    return config, nil
}
```

#### Health Monitoring Integration
```go
// Monitor MCP server health and adapt behavior
type MCPHealthMonitor struct {
    mcpManager   agentflow.MCPManager
    healthStatus map[string]bool
    mu          sync.RWMutex
}

func (m *MCPHealthMonitor) StartHealthChecks(ctx context.Context) {
    ticker := time.NewTicker(30 * time.Second)
    defer ticker.Stop()
    
    for {
        select {
        case <-ctx.Done():
            return
        case <-ticker.C:
            m.checkServerHealth(ctx)
        }
    }
}

func (m *MCPHealthMonitor) checkServerHealth(ctx context.Context) {
    servers := m.mcpManager.ListServers()
    
    m.mu.Lock()
    defer m.mu.Unlock()
    
    for _, serverName := range servers {
        healthy := m.mcpManager.PingServer(ctx, serverName) == nil
        
        if prevHealthy, exists := m.healthStatus[serverName]; exists && prevHealthy != healthy {
            // Health status changed
            agentflow.Logger().Info().
                Str("server", serverName).
                Bool("healthy", healthy).
                Msg("MCP server health status changed")
        }
        
        m.healthStatus[serverName] = healthy
    }
}

func (m *MCPHealthMonitor) IsServerHealthy(serverName string) bool {
    m.mu.RLock()
    defer m.mu.RUnlock()
    return m.healthStatus[serverName]
}
```

#### Error Handling Patterns
```go
// Comprehensive error handling for MCP operations
func (a *MCPAwareAgent) processWithMCP(ctx context.Context, toolName string, params map[string]interface{}) (map[string]interface{}, error) {
    // Set timeout for MCP operations
    ctx, cancel := context.WithTimeout(ctx, 15*time.Second)
    defer cancel()
    
    tool := a.toolRegistry.GetTool(toolName)
    if tool == nil {
        return a.handleMissingTool(toolName, params)
    }
    
    result, err := tool.Execute(ctx, params)
    
    switch {
    case err == nil:
        // Success - log for monitoring
        agentflow.Logger().Debug().
            Str("tool", toolName).
            Msg("MCP tool executed successfully")
        return result, nil
        
    case errors.Is(err, context.DeadlineExceeded):
        // Timeout - try fallback
        agentflow.Logger().Warn().
            Str("tool", toolName).
            Msg("MCP tool timed out, using fallback")
        return a.executeFallback(toolName, params)
        
    case isMCPConnectionError(err):
        // Connection issue - may be temporary
        agentflow.Logger().Error().
            Err(err).
            Str("tool", toolName).
            Msg("MCP connection error, using fallback")
        return a.executeFallback(toolName, params)
        
    default:
        // Other error - depends on configuration
        switch a.config.FallbackBehavior {
        case "fail":
            return nil, fmt.Errorf("MCP tool %s failed: %w", toolName, err)
        case "warn":
            agentflow.Logger().Warn().
                Err(err).
                Str("tool", toolName).
                Msg("MCP tool failed, using fallback")
            return a.executeFallback(toolName, params)
        case "silent":
            return a.executeFallback(toolName, params)
        default:
            return nil, fmt.Errorf("MCP tool %s failed: %w", toolName, err)
        }
    }
}
```

#### Testing MCP Integration
```go
// Test MCP integration with mock servers
func TestMCPIntegration(t *testing.T) {
    // Create mock MCP server configuration
    mockConfig := agentflow.MCPConfig{
        Enabled: true,
        Servers: []agentflow.MCPServerConfig{
            {
                Name:    "test-filesystem",
                Command: "mock-mcp-server",
                Args:    []string{"--type", "filesystem"},
                Env:     map[string]string{"TEST_MODE": "true"},
            },
        },
    }
    
    // Create test agent
    agent := &MCPTestAgent{
        fallbackCalled: false,
    }
    
    agents := map[string]agentflow.AgentHandler{
        "mcp-test": agent,
    }
    
    // Create runner with MCP support
    runner := agentflow.NewRunnerWithConfig(agentflow.RunnerConfig{
        Agents:      agents,
        QueueSize:   10,
        MCPConfig:   &mockConfig,
        TraceLogger: agentflow.NewInMemoryTraceLogger(),
    })
    
    ctx := context.Background()
    err := runner.Start(ctx)
    require.NoError(t, err)
    defer runner.Stop()
    
    // Test MCP tool availability
    toolRegistry := runner.ToolRegistry()
    require.NotNil(t, toolRegistry.GetTool("test-filesystem_read"))
    
    // Test agent execution with MCP tools
    event := agentflow.NewEvent("test", agentflow.EventData{
        "operation": "read_file",
        "path":      "/test/file.txt",
    }, map[string]string{
        agentflow.RouteMetadataKey: "mcp-test",
        agentflow.SessionIDKey:     "test-session",
    })
    
    err = runner.Emit(event)
    require.NoError(t, err)
    
    // Verify execution
    time.Sleep(100 * time.Millisecond)
    require.False(t, agent.fallbackCalled, "Should not have used fallback")
}

// Test fallback behavior when MCP is unavailable
func TestMCPFallback(t *testing.T) {
    // Create configuration with invalid MCP server
    invalidConfig := agentflow.MCPConfig{
        Enabled: true,
        Servers: []agentflow.MCPServerConfig{
            {
                Name:    "invalid-server",
                Command: "non-existent-command",
                Args:    []string{},
            },
        },
    }
    
    agent := &MCPTestAgent{
        fallbackCalled: false,
    }
    
    runner := agentflow.NewRunnerWithConfig(agentflow.RunnerConfig{
        Agents:      map[string]agentflow.AgentHandler{"test": agent},
        QueueSize:   10,
        MCPConfig:   &invalidConfig,
        TraceLogger: agentflow.NewInMemoryTraceLogger(),
    })
    
    ctx := context.Background()
    err := runner.Start(ctx)
    require.NoError(t, err)
    defer runner.Stop()
    
    // Execute with invalid MCP setup
    event := agentflow.NewEvent("test", agentflow.EventData{
        "operation": "read_file",
    }, map[string]string{
        agentflow.RouteMetadataKey: "test",
    })
    
    err = runner.Emit(event)
    require.NoError(t, err)
    
    time.Sleep(100 * time.Millisecond)
    require.True(t, agent.fallbackCalled, "Should have used fallback")
}

type MCPTestAgent struct {
    fallbackCalled bool
}

func (a *MCPTestAgent) Run(ctx context.Context, event agentflow.Event, state agentflow.State) (agentflow.AgentResult, error) {
    // Try MCP tool, fallback if not available
    if tool := runner.ToolRegistry().GetTool("test-filesystem_read"); tool != nil {
        result, err := tool.Execute(ctx, map[string]interface{}{
            "path": event.GetData()["path"],
        })
        if err == nil {
            state.Set("result", result)
            return agentflow.AgentResult{OutputState: state}, nil
        }
    }
    
    // Use fallback
    a.fallbackCalled = true
    state.Set("result", "fallback_data")
    return agentflow.AgentResult{OutputState: state}, nil
}
```

## Production Considerations

### File-Based Tracing

```go
traceLogger := agentflow.NewFileTraceLogger("./production-traces")
runner := agentflow.NewRunnerWithConfig(agentflow.RunnerConfig{
    Agents:      agents,
    QueueSize:   1000,
    TraceLogger: traceLogger,
})
```

### Monitoring and Metrics

```go
// Track custom metrics
type MetricsAgent struct {
    successCount int64
    errorCount   int64
}

func (a *MetricsAgent) Run(ctx context.Context, event agentflow.Event, state agentflow.State) (agentflow.AgentResult, error) {
    startTime := time.Now()
    
    // Process event
    result, err := a.processEvent(ctx, event, state)
    
    // Track metrics
    duration := time.Since(startTime)
    if err != nil {
        atomic.AddInt64(&a.errorCount, 1)
        agentflow.Logger().Error().
            Err(err).
            Dur("duration", duration).
            Msg("Agent execution failed")
    } else {
        atomic.AddInt64(&a.successCount, 1)
        agentflow.Logger().Info().
            Dur("duration", duration).
            Msg("Agent execution succeeded")
    }
      return result, err
}
```

### MCP Production Deployment

When deploying MCP-enabled AgentFlow applications to production, consider these factors:

#### 1. Server Management and Process Supervision

```go
// Production MCP configuration with supervision
type ProductionMCPConfig struct {
    Servers []MCPServerSpec `yaml:"servers"`
    Global  MCPGlobalConfig `yaml:"global"`
}

type MCPServerSpec struct {
    Name            string            `yaml:"name"`
    Command         string            `yaml:"command"`
    Args            []string          `yaml:"args"`
    Env             map[string]string `yaml:"env"`
    WorkingDir      string            `yaml:"working_dir"`
    RestartPolicy   string            `yaml:"restart_policy"`   // "always", "on-failure", "never"
    MaxRestarts     int               `yaml:"max_restarts"`
    HealthCheckPath string            `yaml:"health_check_path"`
    Timeout         time.Duration     `yaml:"timeout"`
}

type MCPGlobalConfig struct {
    HealthCheckInterval time.Duration `yaml:"health_check_interval"`
    ConnectionTimeout   time.Duration `yaml:"connection_timeout"`
    MaxConnections      int           `yaml:"max_connections"`
    LogLevel           string         `yaml:"log_level"`
}

// Production configuration example
// mcp-production.yaml
/*
servers:
  - name: "filesystem"
    command: "node"
    args: ["dist/filesystem-server.js"]
    working_dir: "/opt/mcp-servers/filesystem"
    restart_policy: "always"
    max_restarts: 5
    timeout: 30s
    env:
      NODE_ENV: "production"
      LOG_LEVEL: "info"
      
  - name: "web-search"
    command: "/usr/local/bin/web-search-server"
    args: ["--config", "/etc/web-search/config.json"]
    working_dir: "/opt/mcp-servers/web-search"
    restart_policy: "on-failure"
    max_restarts: 3
    timeout: 15s
    env:
      SEARCH_API_KEY: "${SEARCH_API_KEY}"
      RATE_LIMIT: "100"

global:
  health_check_interval: 30s
  connection_timeout: 10s
  max_connections: 10
  log_level: "info"
*/
```

#### 2. Resource Management and Monitoring

```go
// Monitor MCP server resource usage
type MCPResourceMonitor struct {
    processes map[string]*os.Process
    metrics   map[string]ResourceMetrics
    mu        sync.RWMutex
}

type ResourceMetrics struct {
    CPU       float64
    Memory    uint64
    Uptime    time.Duration
    Restarts  int
    LastCheck time.Time
}

func (m *MCPResourceMonitor) MonitorResources(ctx context.Context) {
    ticker := time.NewTicker(10 * time.Second)
    defer ticker.Stop()
    
    for {
        select {
        case <-ctx.Done():
            return
        case <-ticker.C:
            m.collectMetrics()
        }
    }
}

func (m *MCPResourceMonitor) collectMetrics() {
    m.mu.Lock()
    defer m.mu.Unlock()
    
    for serverName, process := range m.processes {
        if process == nil {
            continue
        }
        
        // Collect resource metrics (implementation depends on your monitoring system)
        metrics := m.getProcessMetrics(process)
        m.metrics[serverName] = metrics
        
        // Log high resource usage
        if metrics.CPU > 80.0 {
            agentflow.Logger().Warn().
                Str("server", serverName).
                Float64("cpu_percent", metrics.CPU).
                Msg("High CPU usage detected for MCP server")
        }
        
        if metrics.Memory > 1024*1024*1024 { // 1GB
            agentflow.Logger().Warn().
                Str("server", serverName).
                Uint64("memory_bytes", metrics.Memory).
                Msg("High memory usage detected for MCP server")
        }
    }
}
```

#### 3. Security and Access Control

```go
// Secure MCP server configuration
type SecureMCPConfig struct {
    // Network security
    AllowedHosts []string `yaml:"allowed_hosts"`
    UseTLS       bool     `yaml:"use_tls"`
    TLSCertPath  string   `yaml:"tls_cert_path"`
    TLSKeyPath   string   `yaml:"tls_key_path"`
    
    // Authentication
    RequireAuth  bool   `yaml:"require_auth"`
    AuthMethod   string `yaml:"auth_method"` // "token", "mutual-tls", "none"
    AuthTokens   []string `yaml:"auth_tokens"`
    
    // Resource limits
    MaxRequestSize    int64         `yaml:"max_request_size"`
    RequestTimeout    time.Duration `yaml:"request_timeout"`
    MaxConcurrentReqs int           `yaml:"max_concurrent_requests"`
}

// Apply security configuration
func configureMCPSecurity(config SecureMCPConfig) agentflow.MCPConfig {
    mcpConfig := agentflow.MCPConfig{
        Enabled: true,
        Security: agentflow.MCPSecurityConfig{
            RequireAuth:       config.RequireAuth,
            AuthMethod:        config.AuthMethod,
            MaxRequestSize:    config.MaxRequestSize,
            RequestTimeout:    config.RequestTimeout,
            MaxConcurrentReqs: config.MaxConcurrentReqs,
        },
    }
    
    // Add TLS configuration if enabled
    if config.UseTLS {
        mcpConfig.Security.TLS = &agentflow.MCPTLSConfig{
            CertPath: config.TLSCertPath,
            KeyPath:  config.TLSKeyPath,
        }
    }
    
    return mcpConfig
}
```

#### 4. High Availability and Load Balancing

```go
// Configure MCP with high availability
type HAMCPConfig struct {
    LoadBalancer LoadBalancerConfig   `yaml:"load_balancer"`
    Servers      []MCPServerCluster   `yaml:"server_clusters"`
}

type LoadBalancerConfig struct {
    Strategy     string `yaml:"strategy"`     // "round-robin", "least-connections", "health-based"
    HealthChecks bool   `yaml:"health_checks"`
    FailoverTime time.Duration `yaml:"failover_time"`
}

type MCPServerCluster struct {
    Name      string          `yaml:"name"`
    Primary   MCPServerSpec   `yaml:"primary"`
    Replicas  []MCPServerSpec `yaml:"replicas"`
    MinReplicas int           `yaml:"min_replicas"`
}

// Implementation with failover support
type HAMCPManager struct {
    clusters map[string]*MCPCluster
    router   *MCPLoadBalancer
}

func (m *HAMCPManager) ExecuteTool(ctx context.Context, toolName string, params map[string]interface{}) (map[string]interface{}, error) {
    // Route request to healthy server
    server := m.router.SelectServer(toolName)
    if server == nil {
        return nil, fmt.Errorf("no healthy servers available for tool %s", toolName)
    }
    
    result, err := server.ExecuteTool(ctx, toolName, params)
    if err != nil {
        // Try failover
        if fallbackServer := m.router.GetFallbackServer(toolName, server); fallbackServer != nil {
            agentflow.Logger().Warn().
                Str("tool", toolName).
                Str("failed_server", server.Name()).
                Str("fallback_server", fallbackServer.Name()).
                Msg("Using fallback MCP server")
                
            return fallbackServer.ExecuteTool(ctx, toolName, params)
        }
    }
    
    return result, err
}
```

#### 5. Performance Optimization

```go
// Connection pooling and caching for MCP
type OptimizedMCPClient struct {
    connectionPool *MCPConnectionPool
    resultCache    *MCPResultCache
    config         MCPOptimizationConfig
}

type MCPOptimizationConfig struct {
    // Connection pooling
    MaxConnections     int           `yaml:"max_connections"`
    ConnectionIdleTime time.Duration `yaml:"connection_idle_time"`
    
    // Result caching
    EnableCaching      bool          `yaml:"enable_caching"`
    CacheTTL          time.Duration `yaml:"cache_ttl"`
    MaxCacheSize      int           `yaml:"max_cache_size"`
    
    // Request optimization
    BatchRequests     bool          `yaml:"batch_requests"`
    MaxBatchSize      int           `yaml:"max_batch_size"`
    CompressionLevel  int           `yaml:"compression_level"`
}

func (c *OptimizedMCPClient) ExecuteTool(ctx context.Context, toolName string, params map[string]interface{}) (map[string]interface{}, error) {
    // Check cache first for cacheable operations
    if c.config.EnableCaching && c.isCacheable(toolName, params) {
        if result := c.resultCache.Get(toolName, params); result != nil {
            agentflow.Logger().Debug().
                Str("tool", toolName).
                Msg("Returning cached MCP result")
            return result, nil
        }
    }
    
    // Get connection from pool
    conn := c.connectionPool.GetConnection()
    defer c.connectionPool.ReturnConnection(conn)
    
    // Execute with timeout
    result, err := conn.ExecuteWithTimeout(ctx, toolName, params, c.config.RequestTimeout)
    if err != nil {
        return nil, err
    }
    
    // Cache result if enabled
    if c.config.EnableCaching && c.isCacheable(toolName, params) {
        c.resultCache.Set(toolName, params, result, c.config.CacheTTL)
    }
    
    return result, nil
}
```

#### 6. Logging and Observability

```go
// Comprehensive MCP logging for production
func configureMCPLogging() {
    // Set up structured logging for MCP operations
    agentflow.Logger().Info().Msg("Configuring MCP production logging")
    
    // Register MCP-specific log handlers
    mcpLogger := agentflow.Logger().With().Str("component", "mcp").Logger()
    
    // Log all MCP tool invocations
    agentflow.RegisterCallback(agentflow.HookBeforeMCPToolCall, "mcp-logger", func(ctx context.Context, data interface{}) error {
        if toolData, ok := data.(agentflow.MCPToolCallData); ok {
            mcpLogger.Info().
                Str("tool", toolData.ToolName).
                Str("server", toolData.ServerName).
                Interface("params", toolData.Parameters).
                Msg("MCP tool call initiated")
        }
        return nil
    })
    
    // Log tool results and errors
    agentflow.RegisterCallback(agentflow.HookAfterMCPToolCall, "mcp-result-logger", func(ctx context.Context, data interface{}) error {
        if resultData, ok := data.(agentflow.MCPToolResultData); ok {
            if resultData.Error != nil {
                mcpLogger.Error().
                    Err(resultData.Error).
                    Str("tool", resultData.ToolName).
                    Dur("duration", resultData.Duration).
                    Msg("MCP tool call failed")
            } else {
                mcpLogger.Info().
                    Str("tool", resultData.ToolName).
                    Dur("duration", resultData.Duration).
                    Msg("MCP tool call completed")
            }
        }
        return nil
    })
}
```

## Migration from Internal Usage

If you were using AgentFlow internally, update your imports:

**Old** (internal use):
```go
import "kunalkushwaha/agentflow/internal/core"
```

**New** (library use):
```go
import agentflow "github.com/kunalkushwaha/agentflow/core"
```

## Support and Community

- **Repository**: [github.com/kunalkushwaha/agentflow](https://github.com/kunalkushwaha/agentflow)
- **Examples**: See the `examples/` directory in the repository
- **Documentation**: Complete documentation in the `docs/` directory
- **Issues**: Report bugs or request features via GitHub Issues

## Conclusion

AgentFlow provides a powerful, flexible framework for building agent-based systems in Go. With proper design patterns and best practices, you can create sophisticated, scalable, and maintainable agent workflows that handle complex business logic with ease.

The library is production-ready and actively maintained. We welcome contributions and feedback from the community!
