---
title: "Configuration Management"
description: "Managing agentflow.toml and environment setup for production deployments"
weight: 30
---

AgentFlow uses a combination of configuration files and environment variables to manage settings for different deployment environments. This guide covers everything from basic setup to advanced production configurations.

## Configuration Files

### Basic agentflow.toml

The primary configuration file for AgentFlow projects:

```toml
# Basic AgentFlow configuration
[project]
name = "my-agent-project"
version = "1.0.0"

# LLM Provider Configuration
[provider]
type = "azure_openai"
api_key = "${AZURE_OPENAI_API_KEY}"
endpoint = "${AZURE_OPENAI_ENDPOINT}"
deployment = "gpt-4"
api_version = "2024-02-15-preview"

# Logging Configuration
[logging]
level = "info"
format = "json"
output = "stdout"

# MCP Configuration
[mcp]
enabled = true

[mcp.servers.search]
command = "npx"
args = ["-y", "@modelcontextprotocol/server-web-search"]
transport = "stdio"

[mcp.servers.filesystem]
command = "npx"
args = ["-y", "@modelcontextprotocol/server-filesystem"]
transport = "stdio"
```

### Production Configuration

For production deployments with advanced features:

```toml
[project]
name = "production-agent"
version = "2.1.0"
environment = "production"

# Azure OpenAI Configuration
[provider]
type = "azure_openai"
api_key = "${AZURE_OPENAI_API_KEY}"
endpoint = "${AZURE_OPENAI_ENDPOINT}"
deployment = "gpt-4"
api_version = "2024-02-15-preview"
max_tokens = 4000
temperature = 0.7
timeout = "30s"
retry_attempts = 3

# Advanced Logging
[logging]
level = "info"
format = "json"
output = "file"
file_path = "/var/log/agentflow/app.log"
rotation = true
max_size = "100MB"
max_backups = 5

# Production MCP with Caching
[mcp]
enabled = true
cache_enabled = true
cache_ttl = "5m"
connection_timeout = "30s"
max_retries = 3

[mcp.cache]
type = "memory"
max_size = 1000

# Production MCP Servers
[mcp.servers.search]
command = "npx"
args = ["-y", "@modelcontextprotocol/server-brave-search"]
transport = "stdio"
env = { "BRAVE_API_KEY" = "${BRAVE_API_KEY}" }

[mcp.servers.database]
command = "npx"
args = ["-y", "@modelcontextprotocol/server-postgres"]
transport = "stdio"
env = { "DATABASE_URL" = "${DATABASE_URL}" }

[mcp.servers.github]
command = "npx"
args = ["-y", "@modelcontextprotocol/server-github"]
transport = "stdio"
env = { "GITHUB_TOKEN" = "${GITHUB_TOKEN}" }

# Metrics and Monitoring
[metrics]
enabled = true
port = 8080
path = "/metrics"

# Circuit Breaker
[circuit_breaker]
enabled = true
failure_threshold = 5
reset_timeout = "60s"
max_requests = 100
```

## Environment Variable Management

### .env File Support

Create a `.env` file in your project root:

```bash
# .env
# LLM Provider
AZURE_OPENAI_API_KEY=your-azure-api-key
AZURE_OPENAI_ENDPOINT=https://your-resource.openai.azure.com
AZURE_OPENAI_DEPLOYMENT=gpt-4

# MCP Tools
SEARCH_API_KEY=your-search-api-key
DATABASE_URL=postgresql://user:pass@localhost/db
GITHUB_TOKEN=your-github-token
BRAVE_API_KEY=your-brave-api-key

# AWS (if using AWS MCP server)
AWS_ACCESS_KEY_ID=your-access-key
AWS_SECRET_ACCESS_KEY=your-secret-key
AWS_REGION=us-east-1
```

**Load environment variables:**
```go
import "github.com/joho/godotenv"

func init() {
    // Load .env file if it exists
    _ = godotenv.Load()
}
```

### Environment-Specific Configuration

Create different configuration files for different environments:

```bash
# Development
agentflow.dev.toml

# Staging
agentflow.staging.toml

# Production
agentflow.prod.toml
```

**Load based on environment:**
```go
env := os.Getenv("AGENTFLOW_ENV")
if env == "" {
    env = "development"
}

configFile := fmt.Sprintf("agentflow.%s.toml", env)
config, err := agentflow.LoadConfig(configFile)
```

## MCP Configuration

### Available MCP Servers

```toml
# Development Tools
[mcp.servers.filesystem]
command = "npx"
args = ["-y", "@modelcontextprotocol/server-filesystem"]
transport = "stdio"

[mcp.servers.docker]
command = "npx"
args = ["-y", "@modelcontextprotocol/server-docker"]
transport = "stdio"

# Web & Search
[mcp.servers.brave_search]
command = "npx"
args = ["-y", "@modelcontextprotocol/server-brave-search"]
transport = "stdio"
env = { "BRAVE_API_KEY" = "${BRAVE_API_KEY}" }

[mcp.servers.fetch]
command = "npx"
args = ["-y", "@modelcontextprotocol/server-fetch"]
transport = "stdio"

# Databases
[mcp.servers.postgres]
command = "npx"
args = ["-y", "@modelcontextprotocol/server-postgres"]
transport = "stdio"
env = { "DATABASE_URL" = "${DATABASE_URL}" }

[mcp.servers.sqlite]
command = "npx"
args = ["-y", "@modelcontextprotocol/server-sqlite"]
transport = "stdio"

# Cloud Services
[mcp.servers.aws]
command = "npx"
args = ["-y", "@modelcontextprotocol/server-aws"]
transport = "stdio"
env = { 
    "AWS_ACCESS_KEY_ID" = "${AWS_ACCESS_KEY_ID}",
    "AWS_SECRET_ACCESS_KEY" = "${AWS_SECRET_ACCESS_KEY}",
    "AWS_REGION" = "${AWS_REGION}"
}
```

### MCP Performance Tuning

```toml
[mcp]
enabled = true
cache_enabled = true
cache_ttl = "10m"
connection_timeout = "30s"
max_retries = 3
max_concurrent_tools = 5
tool_timeout = "60s"

[mcp.cache]
type = "memory"  # or "redis"
max_size = 2000
eviction_policy = "lru"

# Redis cache (if using)
[mcp.cache.redis]
host = "localhost"
port = 6379
password = "${REDIS_PASSWORD}"
db = 0
```

## LLM Provider Configuration

### Azure OpenAI

```toml
[provider]
type = "azure_openai"
api_key = "${AZURE_OPENAI_API_KEY}"
endpoint = "${AZURE_OPENAI_ENDPOINT}"
deployment = "gpt-4"
api_version = "2024-02-15-preview"
max_tokens = 4000
temperature = 0.7
timeout = "30s"
retry_attempts = 3
retry_delay = "1s"
```

### OpenAI

```toml
[provider]
type = "openai"
api_key = "${OPENAI_API_KEY}"
model = "gpt-4"
max_tokens = 4000
temperature = 0.7
timeout = "30s"
organization = "${OPENAI_ORG_ID}"  # Optional
```

### Ollama

```toml
[provider]
type = "ollama"
host = "http://localhost:11434"
model = "llama2"
temperature = 0.7
timeout = "60s"
keep_alive = "5m"
```

### Multiple Providers

```toml
# Primary provider
[provider]
type = "azure_openai"
api_key = "${AZURE_OPENAI_API_KEY}"
endpoint = "${AZURE_OPENAI_ENDPOINT}"
deployment = "gpt-4"

# Fallback providers
[[providers.fallback]]
type = "openai"
api_key = "${OPENAI_API_KEY}"
model = "gpt-4"

[[providers.fallback]]
type = "ollama"
host = "http://localhost:11434"
model = "llama2"
```

## Logging Configuration

### Basic Logging

```toml
[logging]
level = "info"          # debug, info, warn, error
format = "json"         # json, text
output = "stdout"       # stdout, stderr, file
```

### Advanced Logging

```toml
[logging]
level = "info"
format = "json"
output = "file"
file_path = "/var/log/agentflow/app.log"
rotation = true
max_size = "100MB"
max_backups = 5
max_age = "30d"
compress = true

# Structured logging fields
[logging.fields]
service = "agentflow"
version = "1.0.0"
environment = "production"
```

### Log Levels by Component

```toml
[logging]
level = "info"

[logging.components]
mcp = "debug"
llm = "info"
agents = "info"
tools = "warn"
```

## Best Practices

### 1. Environment Variable Naming

Use consistent prefixes:

```bash
# AgentFlow settings
AGENTFLOW_LOG_LEVEL=debug
AGENTFLOW_QUEUE_SIZE=200

# Provider settings  
AZURE_OPENAI_API_KEY=...
OPENAI_API_KEY=...
OLLAMA_HOST=...

# Tool settings
SEARCH_API_KEY=...
DATABASE_URL=...
```

### 2. Security

**Never commit secrets:**
```bash
# .gitignore
.env
*.key
agentflow.prod.toml  # If it contains secrets
```

**Use environment variables for secrets:**
```toml
[provider]
api_key = "${AZURE_OPENAI_API_KEY}"  # Good
# api_key = "sk-actual-key-here"      # Never do this
```

### 3. Configuration Organization

**Separate concerns:**
```toml
# Core application config
[project]
name = "my-app"

# External service configs
[provider]
type = "azure_openai"

[mcp]
enabled = true

# Infrastructure configs
[logging]
level = "info"

[metrics]
enabled = true
```

### 4. Documentation

**Document all configuration options:**
```toml
# agentflow.example.toml
[project]
name = "example-app"              # Application name
version = "1.0.0"                 # Application version

[provider]
type = "azure_openai"             # LLM provider: azure_openai, openai, ollama
api_key = "${AZURE_OPENAI_API_KEY}"  # API key from environment
deployment = "gpt-4"              # Model deployment name
max_tokens = 4000                 # Maximum tokens per request
temperature = 0.7                 # Response randomness (0.0-1.0)
```

## Configuration Loading

### Go Implementation

```go
package main

import (
    "context"
    "fmt"
    "os"
    agentflow "github.com/kunalkushwaha/agentflow/core"
)

func main() {
    // Load configuration
    config, err := agentflow.LoadConfig("agentflow.toml")
    if err != nil {
        // Try environment-specific config
        env := os.Getenv("AGENTFLOW_ENV")
        if env != "" {
            configFile := fmt.Sprintf("agentflow.%s.toml", env)
            config, err = agentflow.LoadConfig(configFile)
        }
        if err != nil {
            panic(fmt.Sprintf("Failed to load config: %v", err))
        }
    }
    
    // Initialize services based on config
    ctx := context.Background()
    
    // Initialize LLM provider
    llm, err := agentflow.InitializeLLMProvider(config.Provider)
    if err != nil {
        panic(fmt.Sprintf("Failed to initialize LLM: %v", err))
    }
    
    // Initialize MCP if enabled
    var mcpManager agentflow.MCPManager
    if config.MCP.Enabled {
        mcpManager, err = agentflow.InitializeMCP(ctx, config.MCP)
        if err != nil {
            panic(fmt.Sprintf("Failed to initialize MCP: %v", err))
        }
    }
    
    // Create and run your agents...
}
```

### Configuration Validation

```go
func validateConfig(config *agentflow.Config) error {
    if config.Provider.Type == "" {
        return fmt.Errorf("provider.type is required")
    }
    
    if config.Provider.APIKey == "" {
        return fmt.Errorf("provider.api_key is required")
    }
    
    if config.MCP.Enabled && len(config.MCP.Servers) == 0 {
        return fmt.Errorf("mcp.servers is required when MCP is enabled")
    }
    
    return nil
}
```

## Next Steps

- **[Architecture Overview](../architecture/)** - Understand AgentFlow's design principles
- **[Agent Fundamentals](../agent-basics/)** - Learn how to build agents with these configurations