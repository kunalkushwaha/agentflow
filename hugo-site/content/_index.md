---
title: "AgentFlow Documentation"
description: "The Go SDK for building production-ready multi-agent AI systems"
---

AgentFlow makes it incredibly simple to build and prototype AI agent workflows in Go. From a single intelligent agent to complex multi-agent orchestrations, AgentFlow provides the SDK foundation and scaffolding you need to develop AI applications with production-ready patterns.

## What Makes AgentFlow Special?

- **30-Second Setup**: Generate working multi-agent systems with a single CLI command
- **LLM-Driven Tool Discovery**: Agents automatically find and use the right tools via MCP protocol  
- **Production-First**: Built-in error handling, observability, and enterprise patterns
- **Unified API**: One clean interface for all LLM providers and tool integrations
- **Zero Dependencies**: Pure Go with minimal external requirements
- **Developer Experience**: From prototype to production without rewriting code

## Quick Start

Get started with AgentFlow in just a few commands:

```bash
# Install the CLI
go install github.com/kunalkushwaha/agentflow/cmd/agentcli@latest

# Create a new agent project
agentcli create my-agent

# Run your first agent
cd my-agent && go run main.go
```

## Perfect for

- **Building intelligent, event-driven workflows** with configurable orchestration patterns
- **Integrating multiple agents and tools** into cohesive, observable systems
- **Leveraging any LLM provider** (OpenAI, Azure OpenAI, Ollama) through unified interfaces
- **Creating modular, extensible AI systems** that scale from prototype to production
- **Focusing on business logic** while AgentFlow handles the infrastructure complexity