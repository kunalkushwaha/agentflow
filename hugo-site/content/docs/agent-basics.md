---
title: "Agent Fundamentals"
description: "Understanding the AgentFlow Agent System - from basic interfaces to advanced patterns"
weight: 10
---

This guide covers the fundamental concepts of building agents in AgentFlow, from the basic interfaces to advanced patterns.

## Core Concepts

### AgentHandler Interface

The `AgentHandler` is the primary interface for implementing agent logic in AgentFlow:

```go
type AgentHandler interface {
    Run(ctx context.Context, event Event, state State) (AgentResult, error)
}
```

**Key Components:**
- **Event**: Contains the user input and metadata
- **State**: Thread-safe storage for agent data
- **AgentResult**: The agent's response and updated state

### Basic Agent Structure

Every agent follows this pattern:

```go
package main

import (
    "context"
    "fmt"
    agentflow "github.com/kunalkushwaha/agentflow/core"
)

type MyAgentHandler struct {
    llm        agentflow.ModelProvider
    mcpManager agentflow.MCPManager
    name       string
}

func NewMyAgent(name string, llm agentflow.ModelProvider, mcp agentflow.MCPManager) *MyAgentHandler {
    return &MyAgentHandler{
        name:       name,
        llm:        llm,
        mcpManager: mcp,
    }
}

func (a *MyAgentHandler) Run(ctx context.Context, event agentflow.Event, state agentflow.State) (agentflow.AgentResult, error) {
    logger := agentflow.Logger()
    logger.Info().Str("agent", a.name).Msg("Processing request")
    
    // 1. Extract input from event
    eventData := event.GetData()
    message, ok := eventData["message"]
    if !ok {
        return agentflow.AgentResult{}, fmt.Errorf("no message in event data")
    }
    
    // 2. Build system prompt
    systemPrompt := "You are a helpful assistant."
    
    // 3. Add available tools to prompt
    toolPrompt := ""
    if a.mcpManager != nil {
        toolPrompt = agentflow.FormatToolsForPrompt(ctx, a.mcpManager)
    }
    
    fullPrompt := fmt.Sprintf("%s\n%s\nUser: %s", systemPrompt, toolPrompt, message)
    
    // 4. Call LLM
    response, err := a.llm.Generate(ctx, fullPrompt)
    if err != nil {
        return agentflow.AgentResult{}, fmt.Errorf("LLM call failed: %w", err)
    }
    
    // 5. Execute any tool calls
    var finalResponse string
    if a.mcpManager != nil {
        toolResults := agentflow.ParseAndExecuteToolCalls(ctx, a.mcpManager, response)
        if len(toolResults) > 0 {
            // Synthesize tool results
            synthesisPrompt := fmt.Sprintf("Original response: %s\nTool results: %v\nProvide a comprehensive answer:", response, toolResults)
            finalResponse, _ = a.llm.Generate(ctx, synthesisPrompt)
        } else {
            finalResponse = response
        }
    } else {
        finalResponse = response
    }
    
    // 6. Update state and return
    state.Set("response", finalResponse)
    state.Set("processed_by", a.name)
    
    return agentflow.AgentResult{
        Result: finalResponse,
        State:  state,
    }, nil
}
```

## Agent Patterns

### 1. Information Gathering Agent

Specializes in research and data collection:

```go
type ResearchAgent struct {
    llm        agentflow.ModelProvider
    mcpManager agentflow.MCPManager
}

func (a *ResearchAgent) Run(ctx context.Context, event agentflow.Event, state agentflow.State) (agentflow.AgentResult, error) {
    message := event.GetData()["message"]
    
    systemPrompt := `You are a research agent. Your job is to gather comprehensive information using available tools.
    
Key behaviors:
- Use search tools for current information
- Use fetch_content for specific URLs
- Gather multiple perspectives
- Organize findings clearly`
    
    // Include tools and generate research-focused response
    toolPrompt := agentflow.FormatToolsForPrompt(ctx, a.mcpManager)
    prompt := fmt.Sprintf("%s\n%s\nResearch query: %s", systemPrompt, toolPrompt, message)
    
    response, err := a.llm.Generate(ctx, prompt)
    if err != nil {
        return agentflow.AgentResult{}, err
    }
    
    // Execute research tools
    toolResults := agentflow.ParseAndExecuteToolCalls(ctx, a.mcpManager, response)
    
    // Compile research findings
    if len(toolResults) > 0 {
        compilationPrompt := fmt.Sprintf(`Research findings: %v
        
Please compile these findings into a structured research report with:
1. Key findings
2. Sources
3. Important details
4. Areas for further investigation`, toolResults)
        
        response, _ = a.llm.Generate(ctx, compilationPrompt)
    }
    
    state.Set("research_findings", response)
    return agentflow.AgentResult{Result: response, State: state}, nil
}
```

### 2. Analysis Agent

Processes information and draws insights:

```go
type AnalysisAgent struct {
    llm agentflow.ModelProvider
}

func (a *AnalysisAgent) Run(ctx context.Context, event agentflow.Event, state agentflow.State) (agentflow.AgentResult, error) {
    // Get previous research findings
    findings, exists := state.Get("research_findings")
    if !exists {
        return agentflow.AgentResult{}, fmt.Errorf("no research findings to analyze")
    }
    
    message := event.GetData()["message"]
    
    systemPrompt := `You are an analysis agent. Your job is to analyze information and provide insights.
    
Key behaviors:
- Identify patterns and trends
- Draw meaningful conclusions  
- Highlight important implications
- Provide actionable insights`
    
    prompt := fmt.Sprintf(`%s

Original query: %s
Research findings: %s

Please provide a thorough analysis with insights and implications.`, systemPrompt, message, findings)
    
    analysis, err := a.llm.Generate(ctx, prompt)
    if err != nil {
        return agentflow.AgentResult{}, err
    }
    
    state.Set("analysis", analysis)
    return agentflow.AgentResult{Result: analysis, State: state}, nil
}
```

### 3. Synthesis Agent

Combines multiple inputs into final output:

```go
type SynthesisAgent struct {
    llm agentflow.ModelProvider
}

func (a *SynthesisAgent) Run(ctx context.Context, event agentflow.Event, state agentflow.State) (agentflow.AgentResult, error) {
    // Gather all previous work
    research, _ := state.Get("research_findings")
    analysis, _ := state.Get("analysis")
    message := event.GetData()["message"]
    
    systemPrompt := `You are a synthesis agent. Your job is to create comprehensive, well-structured final responses.
    
Key behaviors:
- Integrate multiple information sources
- Create coherent, flowing narrative
- Ensure completeness and accuracy
- Provide clear, actionable conclusions`
    
    prompt := fmt.Sprintf(`%s

Original query: %s
Research findings: %s
Analysis: %s

Please synthesize this into a comprehensive, well-structured response that fully addresses the original query.`, 
        systemPrompt, message, research, analysis)
    
    synthesis, err := a.llm.Generate(ctx, prompt)
    if err != nil {
        return agentflow.AgentResult{}, err
    }
    
    state.Set("final_response", synthesis)
    return agentflow.AgentResult{Result: synthesis, State: state}, nil
}
```

## State Management

### Using State for Data Flow

State allows agents to share data across the workflow:

```go
// Agent 1: Store research data
state.Set("research_data", researchResults)
state.Set("sources", sourceList)
state.SetMeta("research_agent", "agent1")

// Agent 2: Access research data
researchData, exists := state.Get("research_data")
if exists {
    // Process the research data
}

// Access metadata
researchAgent, _ := state.GetMeta("research_agent")
```

### State Best Practices

1. **Use descriptive keys**: `"user_preferences"` not `"prefs"`
2. **Store structured data**: Use structs or maps for complex data
3. **Set metadata**: Track which agent processed what
4. **Handle missing data**: Always check if data exists before using

```go
// Good: Structured data storage
type UserProfile struct {
    Name        string
    Preferences []string
    Context     map[string]interface{}
}

profile := UserProfile{
    Name:        "John",
    Preferences: []string{"technical", "detailed"},
    Context:     map[string]interface{}{"industry": "software"},
}
state.Set("user_profile", profile)

// Good: Metadata tracking
state.SetMeta("processed_by", "agent1")
state.SetMeta("processing_time", time.Now().Format(time.RFC3339))
state.SetMeta("data_sources", "research,analysis")
```

## Error Handling

### Graceful Error Management

```go
func (a *MyAgent) Run(ctx context.Context, event agentflow.Event, state agentflow.State) (agentflow.AgentResult, error) {
    // Validate inputs
    message, ok := event.GetData()["message"]
    if !ok {
        return agentflow.AgentResult{}, fmt.Errorf("missing required field: message")
    }
    
    // Handle LLM errors
    response, err := a.llm.Generate(ctx, prompt)
    if err != nil {
        // Log error for debugging
        agentflow.Logger().Error().Err(err).Msg("LLM generation failed")
        
        // Return graceful fallback
        fallbackResponse := "I apologize, but I'm having trouble processing your request right now. Please try again."
        state.Set("error", err.Error())
        state.Set("fallback_used", true)
        
        return agentflow.AgentResult{
            Result: fallbackResponse,
            State:  state,
        }, nil // Don't propagate error, handle gracefully
    }
    
    // Handle tool execution errors
    toolResults := agentflow.ParseAndExecuteToolCalls(ctx, a.mcpManager, response)
    if len(toolResults) == 0 && strings.Contains(response, "tool_call") {
        // Tool call was attempted but failed
        agentflow.Logger().Warn().Msg("Tool calls failed, proceeding without tools")
        // Continue with original response
    }
    
    return agentflow.AgentResult{Result: response, State: state}, nil
}
```

## Testing Agents

### Unit Testing Agent Logic

```go
package main

import (
    "context"
    "testing"
    agentflow "github.com/kunalkushwaha/agentflow/core"
)

func TestMyAgent(t *testing.T) {
    // Setup
    mockLLM := &MockModelProvider{}
    mockMCP := &MockMCPManager{}
    agent := NewMyAgent("test-agent", mockLLM, mockMCP)
    
    // Create test event
    eventData := agentflow.EventData{"message": "Hello, world!"}
    event := agentflow.NewEvent("test", eventData, nil)
    state := agentflow.NewState()
    
    // Execute
    result, err := agent.Run(context.Background(), event, state)
    
    // Assert
    if err != nil {
        t.Fatalf("Expected no error, got %v", err)
    }
    
    if result.Result == "" {
        t.Error("Expected non-empty result")
    }
    
    // Check state was updated
    response, exists := result.State.Get("response")
    if !exists {
        t.Error("Expected response to be set in state")
    }
    
    if response != result.Result {
        t.Error("State response should match result")
    }
}

// Mock implementations for testing
type MockModelProvider struct{}
func (m *MockModelProvider) Generate(ctx context.Context, prompt string) (string, error) {
    return "Mock response", nil
}
func (m *MockModelProvider) Name() string { return "mock" }

type MockMCPManager struct{}
func (m *MockMCPManager) ListTools(ctx context.Context) ([]agentflow.ToolSchema, error) { return nil, nil }
func (m *MockMCPManager) CallTool(ctx context.Context, name string, args map[string]interface{}) (interface{}, error) { return nil, nil }
// ... implement other required methods
```

## Next Steps

- **[Tool Integration](../tool-integration/)** - Learn how to use MCP tools
- **[Configuration Management](../configuration/)** - Configure LLM providers and settings
- **[Architecture Overview](../architecture/)** - Understand AgentFlow's design principles