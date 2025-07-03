---
title: "Documentation"
description: "Core documentation, architecture, and API references for AgentFlow"
---

Welcome to the complete AgentFlow documentation. Here you'll find everything you need to build production-ready multi-agent AI systems in Go.

## Getting Started

- **[Agent Fundamentals](agent-basics/)** - Understanding AgentHandler interface and patterns
- **[Tool Integration](tool-integration/)** - MCP protocol and dynamic tool discovery
- **[Configuration Management](configuration/)** - Managing agentflow.toml and environment setup
- **[Architecture Overview](architecture/)** - Core vs Internal package design principles

## Core Concepts

AgentFlow is designed around a few key concepts that make building AI agent systems straightforward:

- **Agents**: The core building blocks that process events and produce results
- **Events**: Data structures that flow between agents
- **Tools**: External capabilities that agents can discover and use via MCP
- **Orchestration**: Patterns for coordinating multiple agents
- **State Management**: Maintaining context across agent interactions

## API Reference

For detailed API documentation, see our [API Reference](https://pkg.go.dev/github.com/kunalkushwaha/agentflow) on pkg.go.dev.