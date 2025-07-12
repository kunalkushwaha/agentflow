---
title: "Single Agent Example"
description: "A complete working example showing how to build a simple agent that can search the web and answer questions"
weight: 10
---

This example demonstrates how to build a complete AgentFlow application with a single agent that can search the web, analyze information, and provide comprehensive answers to user questions.

## Complete Example

Here's a fully working example that you can run immediately:

### main.go

```go
package main

import (
    "context"
    "fmt"
    "log"
    "os"
    
    "github.com/joho/godotenv"
    agentflow "github.com/kunalkushwaha/agentflow/core"
)

type ResearchAgent struct {
    llm        agentflow.ModelProvider
    mcpManager agentflow.MCPManager
    name       string
}

func NewResearchAgent(name string, llm agentflow.ModelProvider, mcp agentflow.MCPManager) *ResearchAgent {
    return &ResearchAgent{
        name:       name,
        llm:        llm,
        mcpManager: mcp,
    }
}

func (a *ResearchAgent) Run(ctx context.Context, event agentflow.Event, state agentflow.State) (agentflow.AgentResult, error) {
    logger := agentflow.Logger()
    logger.Info().Str("agent", a.name).Msg("Starting research task")
    
    // Extract query from event
    eventData := event.GetData()
    query, ok := eventData["message"]
    if !ok {
        return agentflow.AgentResult{}, fmt.Errorf("no message in event data")
    }
    
    // Build research-focused system prompt
    systemPrompt := `You are a research agent specialized in finding and analyzing information.

Your capabilities:
- Search the web for current information
- Fetch content from specific URLs
- Analyze and synthesize multiple sources
- Provide well-structured, comprehensive answers

For each query:
1. First search for current, relevant information
2. If you find specific URLs, fetch their content for details
3. Analyze the information from multiple perspectives
4. Provide a comprehensive answer with sources

Always use tools when they can provide better, more current information than your training data.

When using tools, format calls exactly like this:
<tool_call>
{"name": "tool_name", "args": {"param": "value"}}
</tool_call>`
    
    // Get available tools
    toolPrompt := ""
    if a.mcpManager != nil {
        toolPrompt = agentflow.FormatToolsForPrompt(ctx, a.mcpManager)
        logger.Info().Msg("Tools discovered and formatted for prompt")
    }
    
    // Create full prompt
    fullPrompt := fmt.Sprintf("%s\n\n%s\n\nUser Query: %s", systemPrompt, toolPrompt, query)
    
    // Get initial LLM response
    logger.Info().Msg("Sending query to LLM")
    response, err := a.llm.Generate(ctx, fullPrompt)
    if err != nil {
        return agentflow.AgentResult{}, fmt.Errorf("LLM generation failed: %w", err)
    }
    
    // Execute any tool calls found in the response
    var finalResponse string
    if a.mcpManager != nil {
        logger.Info().Msg("Checking for tool calls in LLM response")
        toolResults := agentflow.ParseAndExecuteToolCalls(ctx, a.mcpManager, response)
        
        if len(toolResults) > 0 {
            logger.Info().Int("tool_calls", len(toolResults)).Msg("Executed tools, synthesizing results")
            
            // Synthesize tool results with original response
            synthesisPrompt := fmt.Sprintf(`Original response: %s

Tool execution results: %v

Please provide a comprehensive final answer that incorporates the tool results. Structure your response as:

## Summary
Brief overview of the key findings

## Detailed Analysis
Comprehensive analysis based on the research

## Sources
List the sources and their relevance

## Conclusion
Key takeaways and implications`, response, toolResults)
            
            finalResponse, err = a.llm.Generate(ctx, synthesisPrompt)
            if err != nil {
                logger.Warn().Err(err).Msg("Tool result synthesis failed, using original response")
                finalResponse = response
            }
        } else {
            logger.Info().Msg("No tools were executed")
            finalResponse = response
        }
    } else {
        finalResponse = response
    }
    
    // Update state with research results
    state.Set("research_query", query)
    state.Set("research_response", finalResponse)
    state.Set("agent_name", a.name)
    state.Set("tools_used", len(toolResults) > 0)
    
    if len(toolResults) > 0 {
        state.Set("tool_results", toolResults)
    }
    
    logger.Info().Str("agent", a.name).Msg("Research task completed")
    
    return agentflow.AgentResult{
        Result: finalResponse,
        State:  state,
    }, nil
}

func main() {
    // Load environment variables from .env file
    if err := godotenv.Load(); err != nil {
        log.Println("No .env file found, using system environment variables")
    }
    
    // Load configuration
    config, err := agentflow.LoadConfig("agentflow.toml")
    if err != nil {
        log.Fatalf("Failed to load config: %v", err)
    }
    
    ctx := context.Background()
    
    // Initialize LLM provider
    llm, err := agentflow.InitializeLLMProvider(config.Provider)
    if err != nil {
        log.Fatalf("Failed to initialize LLM provider: %v", err)
    }
    
    // Initialize MCP manager for tool integration
    var mcpManager agentflow.MCPManager
    if config.MCP.Enabled {
        mcpManager, err = agentflow.InitializeProductionMCP(ctx, config.MCP)
        if err != nil {
            log.Fatalf("Failed to initialize MCP: %v", err)
        }
        defer mcpManager.Disconnect()
    }
    
    // Create our research agent
    agent := NewResearchAgent("research-agent", llm, mcpManager)
    
    // Get query from command line or use default
    query := "What are the latest developments in Go programming language?"
    if len(os.Args) > 1 {
        query = os.Args[1]
    }
    
    fmt.Printf("🔍 Research Query: %s\n\n", query)
    
    // Create event and state
    eventData := agentflow.EventData{"message": query}
    event := agentflow.NewEvent("research_request", eventData, nil)
    state := agentflow.NewState()
    
    // Run the agent
    fmt.Println("🤖 Agent is researching...")
    result, err := agent.Run(ctx, event, state)
    if err != nil {
        log.Fatalf("Agent execution failed: %v", err)
    }
    
    // Display results
    fmt.Println("📋 Research Results:")
    fmt.Println("=" + string(rune('=')*80))
    fmt.Println(result.Result)
    fmt.Println("=" + string(rune('=')*80))
    
    // Show metadata
    if toolsUsed, exists := result.State.Get("tools_used"); exists && toolsUsed.(bool) {
        fmt.Println("\n🔧 Tools were used to gather current information")
    }
    
    fmt.Println("\n✅ Research completed successfully!")
}
```

### agentflow.toml

```toml
[project]
name = "research-agent-example"
version = "1.0.0"

# Azure OpenAI Configuration (recommended)
[provider]
type = "azure_openai"
api_key = "${AZURE_OPENAI_API_KEY}"
endpoint = "${AZURE_OPENAI_ENDPOINT}"
deployment = "gpt-4"
api_version = "2024-02-15-preview"
max_tokens = 4000
temperature = 0.7
timeout = "60s"

# Alternative: OpenAI Configuration
# [provider]
# type = "openai"
# api_key = "${OPENAI_API_KEY}"
# model = "gpt-4"
# max_tokens = 4000
# temperature = 0.7

# Alternative: Ollama Configuration (for local development)
# [provider]
# type = "ollama"
# host = "http://localhost:11434"
# model = "llama2"
# temperature = 0.7

# MCP Configuration for Tools
[mcp]
enabled = true
cache_enabled = true
cache_ttl = "10m"
connection_timeout = "30s"
max_retries = 3

# Web Search Tool
[mcp.servers.search]
command = "npx"
args = ["-y", "@modelcontextprotocol/server-brave-search"]
transport = "stdio"
env = { "BRAVE_API_KEY" = "${BRAVE_API_KEY}" }

# Alternative: Basic web search (no API key required)
# [mcp.servers.search]
# command = "npx"
# args = ["-y", "@modelcontextprotocol/server-web-search"]
# transport = "stdio"

# URL Content Fetcher
[mcp.servers.fetch]
command = "npx"
args = ["-y", "@modelcontextprotocol/server-fetch"]
transport = "stdio"

# File System Access (optional)
[mcp.servers.filesystem]
command = "npx"
args = ["-y", "@modelcontextprotocol/server-filesystem"]
transport = "stdio"

# Logging Configuration
[logging]
level = "info"
format = "json"
output = "stdout"
```

### .env

```bash
# Azure OpenAI (recommended)
AZURE_OPENAI_API_KEY=your-azure-openai-api-key
AZURE_OPENAI_ENDPOINT=https://your-resource.openai.azure.com
AZURE_OPENAI_DEPLOYMENT=gpt-4

# OR OpenAI
# OPENAI_API_KEY=your-openai-api-key

# Optional: Brave Search API for better web search
BRAVE_API_KEY=your-brave-search-api-key

# Optional: Other service API keys
GITHUB_TOKEN=your-github-token
DATABASE_URL=postgresql://user:pass@localhost/db
```

### go.mod

```go
module research-agent-example

go 1.21

require (
    github.com/joho/godotenv v1.5.1
    github.com/kunalkushwaha/agentflow v0.1.0
)
```

## How to Run

### 1. Prerequisites

Make sure you have the following installed:
- Go 1.21 or later
- Node.js (for MCP servers)
- An LLM provider API key (Azure OpenAI, OpenAI, or local Ollama)

### 2. Setup

```bash
# Clone or create the project directory
mkdir research-agent-example
cd research-agent-example

# Initialize Go module
go mod init research-agent-example

# Create the files above (main.go, agentflow.toml, .env)

# Install dependencies
go mod tidy

# Install MCP servers (they'll be auto-installed on first run)
npm install -g @modelcontextprotocol/server-brave-search
npm install -g @modelcontextprotocol/server-fetch
```

### 3. Configuration

Edit your `.env` file with your actual API keys:

```bash
# For Azure OpenAI
AZURE_OPENAI_API_KEY=your-actual-api-key
AZURE_OPENAI_ENDPOINT=https://your-resource.openai.azure.com

# For OpenAI
OPENAI_API_KEY=your-actual-api-key

# Optional but recommended for better search
BRAVE_API_KEY=your-brave-api-key
```

### 4. Run the Example

```bash
# Run with default query
go run main.go

# Run with custom query
go run main.go "What are the latest AI breakthroughs in 2024?"

# Run with a specific research question
go run main.go "How does quantum computing impact cryptography?"
```

## Example Output

```
🔍 Research Query: What are the latest developments in Go programming language?

🤖 Agent is researching...

📋 Research Results:
================================================================================
## Summary
Based on the latest information gathered, Go continues to evolve with significant 
improvements in performance, developer experience, and ecosystem growth in 2024.

## Detailed Analysis
The Go team has released several important updates this year:

1. **Go 1.22 Features**: Enhanced performance improvements, better garbage collection,
   and improved tooling for dependency management.

2. **Generics Evolution**: Continued refinement of the generics implementation 
   introduced in Go 1.18, with better type inference and performance optimizations.

3. **Developer Tooling**: Significant improvements to the Go language server (gopls)
   and enhanced debugging capabilities.

## Sources
- Official Go blog posts from 2024
- Go team announcements on GitHub
- Community discussions and benchmarks

## Conclusion
Go maintains its position as a leading language for cloud-native development,
with ongoing improvements that enhance both performance and developer productivity.
================================================================================

🔧 Tools were used to gather current information

✅ Research completed successfully!
```

## Key Features Demonstrated

### 1. LLM Integration
- Supports multiple providers (Azure OpenAI, OpenAI, Ollama)
- Configurable parameters (temperature, max tokens, etc.)
- Error handling and fallbacks

### 2. Tool Integration via MCP
- Dynamic tool discovery
- Automatic tool execution based on LLM decisions
- Web search and content fetching capabilities
- Extensible tool ecosystem

### 3. Configuration Management
- Environment-based configuration
- Secure API key management
- Multiple deployment environments

### 4. Agent Architecture
- Clean separation of concerns
- Structured prompt engineering
- State management for data flow
- Comprehensive logging

### 5. Error Handling
- Graceful fallbacks when tools fail
- Detailed error reporting
- Connection retry logic

## Customization Options

### Different LLM Providers

Switch between providers by changing the `[provider]` section in `agentflow.toml`:

```toml
# For local development with Ollama
[provider]
type = "ollama"
host = "http://localhost:11434"
model = "llama2"
temperature = 0.7
```

### Additional Tools

Add more MCP servers to expand capabilities:

```toml
# Database integration
[mcp.servers.postgres]
command = "npx"
args = ["-y", "@modelcontextprotocol/server-postgres"]
transport = "stdio"
env = { "DATABASE_URL" = "${DATABASE_URL}" }

# GitHub integration
[mcp.servers.github]
command = "npx"
args = ["-y", "@modelcontextprotocol/server-github"]
transport = "stdio"
env = { "GITHUB_TOKEN" = "${GITHUB_TOKEN}" }
```

### Agent Behavior

Modify the system prompt in `main.go` to change agent behavior:

```go
systemPrompt := `You are a technical analysis agent specialized in software development.

Focus on:
- Code examples and implementations
- Best practices and patterns
- Performance implications
- Security considerations

Provide detailed, actionable insights for developers.`
```

## Next Steps

1. **Extend the Agent**: Add more sophisticated reasoning or specialized prompts
2. **Add More Tools**: Integrate databases, APIs, or custom tools
3. **Multi-Agent Workflow**: Combine multiple agents for complex tasks
4. **Production Deployment**: Add monitoring, scaling, and deployment configurations

This example provides a solid foundation for building more complex AgentFlow applications. The modular design makes it easy to extend and customize for your specific use cases.