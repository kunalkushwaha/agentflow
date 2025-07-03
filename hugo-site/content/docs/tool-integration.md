---
title: "Tool Integration"
linkTitle: "Tools"
weight: 3
description: >
  Understanding MCP protocol and dynamic tool discovery in AgentFlow.
---

## Model Context Protocol (MCP)

AgentFlow leverages the Model Context Protocol (MCP) to enable agents to discover and use external tools dynamically. This allows agents to interact with various services, APIs, and local resources seamlessly.

## Enabling MCP in Your Agent

### Basic MCP Configuration

```toml
# agentflow.toml
[mcp]
enabled = true
cache_enabled = true
cache_ttl = "5m"

# Web search capability
[mcp.servers.search]
command = "npx"
args = ["-y", "@modelcontextprotocol/server-web-search"]
transport = "stdio"

# File system access
[mcp.servers.filesystem]
command = "npx"
args = ["-y", "@modelcontextprotocol/server-filesystem"]
transport = "stdio"
```

### Creating an MCP-Enabled Agent

```go
package main

import (
    "context"
    "fmt"
    "log"
    
    agentflow "github.com/kunalkushwaha/agentflow/core"
)

func main() {
    // Create provider
    provider := agentflow.NewAzureProvider(agentflow.AzureConfig{
        APIKey:     "your-api-key",
        Endpoint:   "your-endpoint",
        Deployment: "gpt-4",
    })

    // Create MCP configuration
    mcpConfig := agentflow.MCPConfig{
        Enabled: true,
        Servers: map[string]agentflow.MCPServerConfig{
            "search": {
                Command:   "npx",
                Args:      []string{"-y", "@modelcontextprotocol/server-web-search"},
                Transport: "stdio",
            },
        },
    }

    // Create MCP-enabled agent
    agent := agentflow.NewMCPAgent("research-agent", provider, &mcpConfig)

    // The agent can now use web search tools
    response, err := agent.Process(context.Background(), 
        "Search for the latest news about Go programming language and summarize the key developments")
    
    if err != nil {
        log.Fatal(err)
    }

    fmt.Println("Research Result:", response)
}
```

## Available MCP Servers

### Web and Search Tools

```toml
# Brave Search
[mcp.servers.brave_search]
command = "npx"
args = ["-y", "@modelcontextprotocol/server-brave-search"]
transport = "stdio"
env = { "BRAVE_API_KEY" = "${BRAVE_API_KEY}" }

# Web Fetch
[mcp.servers.fetch]
command = "npx"
args = ["-y", "@modelcontextprotocol/server-fetch"]
transport = "stdio"
```

### Development Tools

```toml
# Git operations
[mcp.servers.git]
command = "npx"
args = ["-y", "@modelcontextprotocol/server-git"]
transport = "stdio"

# Docker management
[mcp.servers.docker]
command = "npx"
args = ["-y", "@modelcontextprotocol/server-docker"]
transport = "stdio"
```

### Database Integration

```toml
# PostgreSQL
[mcp.servers.postgres]
command = "npx"
args = ["-y", "@modelcontextprotocol/server-postgres"]
transport = "stdio"
env = { "DATABASE_URL" = "${DATABASE_URL}" }

# SQLite
[mcp.servers.sqlite]
command = "npx"
args = ["-y", "@modelcontextprotocol/server-sqlite"]
transport = "stdio"
```

## Tool Discovery Process

1. **Agent Initialization**: MCP servers are started when the agent is created
2. **Capability Discovery**: Agent queries each server for available tools/resources
3. **Dynamic Invocation**: Agent can call tools based on user requests
4. **Result Integration**: Tool outputs are incorporated into agent responses

## Best Practices

### Security Considerations

- **Limit server permissions** to only what's necessary
- **Use environment variables** for sensitive configuration
- **Validate tool outputs** before using in responses
- **Monitor tool usage** for unexpected behavior

### Performance Optimization

- **Enable caching** for frequently used tools
- **Set appropriate timeouts** for external calls
- **Limit concurrent tool usage** to prevent resource exhaustion
- **Cache tool discovery results** to speed up initialization

### Error Handling

```go
// Example of robust tool usage
response, err := agent.Process(ctx, "Search for Go tutorials")
if err != nil {
    // Check if it's a tool-related error
    if mcpErr, ok := err.(*agentflow.MCPError); ok {
        log.Printf("Tool error: %s (tool: %s)", mcpErr.Message, mcpErr.Tool)
        // Fallback to basic response without tools
        return agent.ProcessWithoutMCP(ctx, "Provide general information about Go tutorials")
    }
    return "", err
}
```

## Production Configuration

```toml
[mcp]
enabled = true
cache_enabled = true
cache_ttl = "10m"
connection_timeout = "30s"
max_retries = 3
max_concurrent_connections = 10

[mcp.cache]
type = "memory"          # memory, redis (future)
max_size = 1000
cleanup_interval = "1m"

# Production-ready servers with environment variables
[mcp.servers.search]
command = "npx"
args = ["-y", "@modelcontextprotocol/server-web-search"]
transport = "stdio"
env = { "SEARCH_API_KEY" = "${SEARCH_API_KEY}" }
```

## Next Steps

- Explore [custom MCP servers](../custom-tools/) for specific use cases
- Learn about [multi-agent coordination](../multi-agent/) with shared tools
- See [production deployment](../production/) patterns for MCP-enabled agents