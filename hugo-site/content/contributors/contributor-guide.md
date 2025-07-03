---
title: "Contributor Guide"
weight: 1
description: >
  Getting started with AgentFlow development and contribution.
---

## Welcome to AgentFlow Development

This guide will help you get started with contributing to the AgentFlow project, understanding the codebase, and following our development practices.

## Quick Start for Contributors

### 1. Development Setup

```bash
# Clone the repository
git clone https://github.com/kunalkushwaha/agentflow.git
cd agentflow

# Install dependencies
go mod tidy

# Run tests to ensure everything works
go test ./...

# Install development tools
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
```

### 2. Project Structure

```
agentflow/
├── cmd/                    # Main applications
│   └── agentcli/          # CLI application
├── core/                   # Public API
│   ├── agent.go           # Core interfaces
│   ├── runner.go          # Public runner interface
│   └── *.go               # Other public APIs
├── internal/               # Private implementation
│   ├── agents/            # Agent implementations
│   ├── mcp/               # MCP implementation
│   ├── llm/               # LLM provider implementations
│   └── */                 # Other internal packages
├── docs/                   # Documentation
└── hugo-site/              # Hugo documentation site
```

## Development Workflow

1. **Fork the repository** on GitHub
2. **Create a feature branch** from `main`
3. **Make your changes** with appropriate tests
4. **Run the full test suite** to ensure nothing breaks
5. **Submit a pull request** with a clear description

## Code Standards

- Follow standard Go conventions
- Write tests for new functionality
- Update documentation for user-facing changes
- Use descriptive commit messages

## Getting Help

- **GitHub Discussions**: General questions about contributing
- **GitHub Issues**: Bug reports and feature requests
- **Code Reviews**: Ask questions in PR comments

For more detailed information, see our other contributor guides:

- [Architecture Deep Dive](../architecture/) - Internal structure and design decisions
- [Testing Strategy](../testing/) - Unit tests, integration tests, and benchmarks
- [Code Style Guide](../code-style/) - Go standards and project conventions