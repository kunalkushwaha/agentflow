---
title: "Single Agent Example"
weight: 1
description: >
  Basic example of implementing a single agent with AgentFlow.
---

## Creating Your First Agent

This example demonstrates how to create a simple agent that can process natural language instructions and respond intelligently.

### Basic Setup

```go
package main

import (
    "context"
    "fmt"
    "log"
    
    agentflow "github.com/kunalkushwaha/agentflow/core"
)

func main() {
    // Configure Azure OpenAI provider
    provider := agentflow.NewAzureProvider(agentflow.AzureConfig{
        APIKey:     "your-api-key",
        Endpoint:   "https://your-resource.openai.azure.com",
        Deployment: "gpt-4",
    })

    // Create a basic agent
    agent := agentflow.NewBasicAgent("assistant", provider)

    // Process a simple request
    ctx := context.Background()
    response, err := agent.Process(ctx, "Hello! Can you help me understand Go interfaces?")
    if err != nil {
        log.Fatal(err)
    }

    fmt.Println("Agent Response:", response)
}
```

### Configuration File

Create an `agentflow.toml` file in your project:

```toml
[provider]
type = "azure"
api_key = "${AZURE_OPENAI_API_KEY}"
endpoint = "${AZURE_OPENAI_ENDPOINT}"
deployment = "gpt-4"
model = "gpt-4"
max_tokens = 2000
temperature = 0.7
```

### Environment Variables

Set up your environment:

```bash
export AZURE_OPENAI_API_KEY="your-secret-key"
export AZURE_OPENAI_ENDPOINT="https://your-resource.openai.azure.com"
```

### Running the Example

```bash
# Install dependencies
go mod init my-agent-example
go mod tidy

# Run the agent
go run main.go
```

## Expected Output

```
Agent Response: Hello! I'd be happy to help you understand Go interfaces. 

An interface in Go is a type that specifies a method set. Any type that 
implements all the methods in the interface automatically satisfies that 
interface, without explicitly declaring it...
```

## Next Steps

- Try different prompts and see how the agent responds
- Experiment with different model parameters in the configuration
- Add error handling and logging for production use
- Explore [Multi-Agent coordination](../multi-agent/) for more complex workflows