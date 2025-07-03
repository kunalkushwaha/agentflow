---
title: "Agent Fundamentals"
linkTitle: "Agent Basics"
weight: 1
description: >
  Understanding AgentFlow's core agent concepts and patterns.
---

## What is an AgentFlow Agent?

An AgentFlow agent is an intelligent entity that can:

- Process natural language instructions
- Execute tools and functions
- Maintain conversation context
- Coordinate with other agents
- Handle errors gracefully

## Core Interface: AgentHandler

All agents in AgentFlow implement the `AgentHandler` interface:

```go
type AgentHandler interface {
    Process(ctx context.Context, message string) (string, error)
    GetName() string
    GetCapabilities() []string
}
```

## Creating Your First Agent

### Basic Agent Setup

```go
package main

import (
    "context"
    agentflow "github.com/kunalkushwaha/agentflow/core"
)

func main() {
    // Create LLM provider
    provider := agentflow.NewAzureProvider(agentflow.AzureConfig{
        APIKey:     "your-api-key",
        Endpoint:   "your-endpoint",
        Deployment: "gpt-4",
    })

    // Create agent with MCP support
    agent := agentflow.NewMCPAgent("my-agent", provider, nil)

    // Process a message
    response, err := agent.Process(context.Background(), "Hello, world!")
    if err != nil {
        panic(err)
    }
    
    fmt.Println(response)
}
```

## Agent Types

### Basic Agent
Simple agents that process messages without external tools.

### MCP-Enabled Agent
Agents that can discover and use tools via the Model Context Protocol.

### Custom Agent
Agents with custom implementations of the AgentHandler interface.

## Best Practices

1. **Always use context** for cancellation and timeouts
2. **Handle errors gracefully** with appropriate fallbacks
3. **Keep agents focused** on specific tasks or domains
4. **Use descriptive names** for better orchestration
5. **Test agent behavior** with various inputs