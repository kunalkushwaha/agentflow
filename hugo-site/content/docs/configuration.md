---
title: "Configuration Management"
linkTitle: "Configuration"
weight: 4
description: >
  Managing agentflow.toml and environment setup.
---

## Configuration File Structure

AgentFlow uses TOML configuration files for setup:

```toml
# agentflow.toml

# Primary LLM provider for all agents
[provider]
type = "azure"                    # Using Azure OpenAI for enterprise compliance
api_key = "${AZURE_OPENAI_API_KEY}"
deployment = "gpt-4"              # GPT-4 deployment for high-quality responses

# Tools available to all agents
[mcp]
enabled = true

# Web search for research agents
[mcp.servers.search]
command = "npx"
args = ["-y", "@modelcontextprotocol/server-web-search"]
transport = "stdio"

# Docker management for DevOps agents
[mcp.servers.docker]
command = "npx"
args = ["-y", "@modelcontextprotocol/server-docker"]
transport = "stdio"
```

## Environment Variables

### Security Best Practices

**Use environment variables for secrets:**
```bash
export AZURE_OPENAI_API_KEY="your-secret-key"
export AZURE_OPENAI_ENDPOINT="https://your-resource.openai.azure.com"
```

### Naming Conventions

- Use UPPERCASE_SNAKE_CASE for environment variables
- Prefix with service name: `AZURE_`, `OPENAI_`, `OLLAMA_`
- Suffix with type: `_API_KEY`, `_ENDPOINT`, `_URL`

## MCP Configuration

### Basic MCP Setup

```toml
[mcp]
enabled = true
cache_enabled = true
cache_ttl = "5m"

# Web search tools
[mcp.servers.search]
command = "npx"
args = ["-y", "@modelcontextprotocol/server-web-search"]
transport = "stdio"
```

### Available MCP Servers

```toml
# Development Tools
[mcp.servers.filesystem]
command = "npx"
args = ["-y", "@modelcontextprotocol/server-filesystem"]
transport = "stdio"

# Databases
[mcp.servers.postgres]
command = "npx"
args = ["-y", "@modelcontextprotocol/server-postgres"]
transport = "stdio"
env = { "DATABASE_URL" = "${DATABASE_URL}" }
```

## Configuration Organization

### Environment-Specific Configs

```toml
# Base configuration
[provider]
type = "azure"
model = "gpt-4"

# Development overrides in agentflow.dev.toml
[provider]
type = "mock"
response = "Development response"

# Production overrides in agentflow.prod.toml  
[provider]
max_tokens = 4000
timeout = "60s"
```