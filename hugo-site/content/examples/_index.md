---
title: "Examples"
linkTitle: "Examples"
weight: 30
menu:
  main:
    weight: 30
    pre: <i class='fa fa-code'></i>
---

## AgentFlow Examples

This section contains practical examples and code snippets to help you get started with AgentFlow.

## Getting Started Examples

### Quick Start

```bash
# Install the CLI
go install github.com/kunalkushwaha/agentflow/cmd/agentcli@latest

# Create your first project
agentcli create my-agent-app --agents 2 --mcp-enabled
cd my-agent-app

# Run your agents
go run . -m "search for the latest Go tutorials and summarize them"
```

## Example Workflows

- **[Single Agent Workflow](single-agent/)** - Basic agent implementation
- **[Multi-Agent Coordination](multi-agent/)** - Orchestrating multiple agents
- **[Tool Integration Examples](tool-integration/)** - Using MCP servers and tools
- **[Custom LLM Providers](custom-providers/)** - Implementing custom providers

## Code Samples

- **[Basic Configuration](basic-config/)** - Setting up agentflow.toml
- **[Error Handling Patterns](error-handling/)** - Robust error handling
- **[Production Deployment](production/)** - Deployment and scaling examples