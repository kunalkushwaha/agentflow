# AgentFlow Documentation

**The Complete Guide to Building AI Agent Systems in Go**

AgentFlow is a production-ready Go framework for building intelligent agent workflows with dynamic tool integration, multi-provider LLM support, and enterprise-grade patterns.

## 📚 For AgentFlow Users

### **Getting Started**
- **[Quick Start Guide](#quick-start)** - Get running in 5 minutes
- **[Installation & Setup](#installation)** - Go module setup and CLI installation
- **[Your First Agent](#first-agent)** - Build a simple agent from scratch
- **[Multi-Agent Orchestration](#multi-agent)** - Collaborative, sequential, and mixed workflows
- **[Workflow Visualization](visualization_guide.md)** - Generate Mermaid diagrams automatically

### **Core Concepts**  
- **[Agent Fundamentals](guides/AgentBasics.md)** - Understanding AgentHandler interface and patterns
- **[Memory & RAG](guides/Memory.md)** - Persistent memory, vector search, and knowledge bases
- **[Multi-Agent Orchestration](multi_agent_orchestration.md)** - Orchestration patterns and API reference
- **[Examples & Tutorials](guides/Examples.md)** - Practical examples and code samples
- **[Tool Integration](guides/ToolIntegration.md)** - MCP protocol and dynamic tool discovery
- **[LLM Providers](guides/Providers.md)** - Azure, OpenAI, Ollama, and custom providers
- **[Configuration](guides/Configuration.md)** - Managing agentflow.toml and environment setup

### **Advanced Usage**
- **[Multi-Agent Orchestration](multi_agent_orchestration.md)** - Advanced orchestration patterns and configuration
- **[RAG Configuration](guides/RAGConfiguration.md)** - Retrieval-Augmented Generation setup and tuning
- **[Memory Provider Setup](guides/MemoryProviderSetup.md)** - PostgreSQL, Weaviate, and in-memory setup guides
- **[Workflow Visualization](visualization_guide.md)** - Generate and customize Mermaid diagrams
- **[Production Deployment](guides/Production.md)** - Scaling, monitoring, and best practices  
- **[Error Handling](guides/ErrorHandling.md)** - Resilient agent workflows
- **[Custom Tools](guides/CustomTools.md)** - Building your own MCP servers
- **[Performance Tuning](guides/Performance.md)** - Optimization and benchmarking

### **API Reference**
- **[Core Package API](api/core.md)** - Complete public API reference
- **[Agent Interface](api/agents.md)** - AgentHandler and related types
- **[Memory API](api/memory.md)** - Memory system and RAG APIs
- **[MCP Integration](api/mcp.md)** - Tool discovery and execution APIs
- **[CLI Commands](api/cli.md)** - agentcli reference

## 🔧 For AgentFlow Contributors

### **Development Setup**
- **[Contributor Guide](contributors/ContributorGuide.md)** - Getting started with development
- **[Architecture Deep Dive](contributors/Architecture.md)** - Internal structure and design decisions
- **[Testing Strategy](contributors/Testing.md)** - Unit tests, integration tests, and benchmarks
- **[Release Process](contributors/ReleaseProcess.md)** - How releases are managed

### **Codebase Structure**
- **[Core vs Internal](contributors/CoreVsInternal.md)** - Public API vs implementation
- **[Adding Features](contributors/AddingFeatures.md)** - How to extend AgentFlow
- **[Code Style](contributors/CodeStyle.md)** - Go standards and project conventions
- **[Documentation Standards](contributors/DocsStandards.md)** - Writing user-focused docs

---

## Quick Start

### Installation
```bash
# Install the CLI
go install github.com/kunalkushwaha/agentflow/cmd/agentcli@latest

# Create a collaborative multi-agent system
agentcli create research-system \
  --orchestration-mode collaborative \
  --collaborative-agents "researcher,analyzer,validator" \
  --visualize \
  --mcp-enabled

cd research-system

# Run with any message - agents work together intelligently
go run . -m "research AI trends and provide comprehensive analysis"
```

### Multi-Agent Orchestration
```bash
# Sequential processing pipeline
agentcli create data-pipeline \
  --orchestration-mode sequential \
  --sequential-agents "collector,processor,formatter" \
  --visualize

# Loop-based workflow with conditions
agentcli create quality-loop \
  --orchestration-mode loop \
  --loop-agent "quality-checker" \
  --max-iterations 5 \
  --visualize

# Mixed collaborative + sequential workflow
agentcli create complex-workflow \
  --orchestration-mode mixed \
  --collaborative-agents "analyzer,validator" \
  --sequential-agents "processor,reporter" \
  --visualize-output "docs/diagrams"
```

### First Agent
```bash
# Generate a single agent project
agentcli create simple-agent --visualize

# The generated agent1.go will look like this:
```

```go
package main

import (
    "context"
    "fmt"
    agentflow "github.com/kunalkushwaha/agentflow/core"
)

type Agent1Handler struct {
    llm        agentflow.ModelProvider
    mcpManager agentflow.MCPManager
}

func NewAgent1(llm agentflow.ModelProvider, mcp agentflow.MCPManager) *Agent1Handler {
    return &Agent1Handler{llm: llm, mcpManager: mcp}
}

func (a *Agent1Handler) Run(ctx context.Context, event agentflow.Event, state agentflow.State) (agentflow.AgentResult, error) {
    // Extract user message
    message := event.GetData()["message"]
    
    // Build prompt with available tools
    systemPrompt := "You are a helpful assistant that uses tools when needed."
    toolPrompt := agentflow.FormatToolsForPrompt(ctx, a.mcpManager)
    fullPrompt := fmt.Sprintf("%s\n%s\nUser: %s", systemPrompt, toolPrompt, message)
    
    // Get LLM response
    response, err := a.llm.Generate(ctx, fullPrompt)
    if err != nil {
        return agentflow.AgentResult{}, err
    }
    
    // Execute any tool calls
    toolResults := agentflow.ParseAndExecuteToolCalls(ctx, a.mcpManager, response)
    if len(toolResults) > 0 {
        // Synthesize tool results with response
        finalPrompt := fmt.Sprintf("Response: %s\nTool Results: %v\nProvide final answer:", response, toolResults)
        response, _ = a.llm.Generate(ctx, finalPrompt)
    }
    
    // Return result
    state.Set("response", response)
    return agentflow.AgentResult{Result: response, State: state}, nil
}
```

### Multi-Agent
```bash
# Generate a collaborative multi-agent workflow
agentcli create research-system \
  --orchestration-mode collaborative \
  --collaborative-agents "researcher,analyzer,validator" \
  --visualize

# This creates:
# - researcher.go (Research agent - gathers information)
# - analyzer.go (Analysis agent - processes data)  
# - validator.go (Validation agent - ensures quality)
# - main.go (Collaborative orchestration)
# - workflow.mmd (Mermaid diagram)
```

**Collaborative Orchestration Code:**
```go
package main

import (
    "context"
    "fmt"
    "log"
    "time"
    
    "github.com/kunalkushwaha/agentflow/core"
)

func main() {
    // Initialize agents
    agents := map[string]core.AgentHandler{
        "researcher": NewResearcher(),
        "analyzer":   NewAnalyzer(),
        "validator":  NewValidator(),
    }
    
    // Create collaborative orchestration
    runner := core.NewOrchestrationBuilder(core.OrchestrationCollaborate).
        WithAgents(agents).
        WithTimeout(2 * time.Minute).
        WithFailureThreshold(0.8).
        WithMaxConcurrency(10).
        Build()
    
    // Create event
    event := core.NewEvent("all", map[string]interface{}{
        "task": "research AI trends and provide comprehensive analysis",
    }, nil)
    
    // All agents process the event in parallel
    result, err := runner.Run(context.Background(), event)
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Printf("Collaborative Result: %s\n", result.GetResult())
}
```

## Why AgentFlow?

### **For Users:**
- **⚡ Fast Setup**: Working agents in 5 minutes with CLI scaffolding
- **🔧 Tool-Rich**: Dynamic tool discovery via MCP protocol
- **🌐 Provider Agnostic**: Works with any LLM (Azure, OpenAI, Ollama)
- **🏗️ Production Ready**: Built-in error handling, monitoring, scaling patterns

### **For Contributors:**
- **🎯 Clear Architecture**: Separation between core (public API) and internal (implementation)
- **📝 Documentation First**: Every feature documented with examples
- **🧪 Test Coverage**: Comprehensive unit and integration tests
- **🔄 Continuous Integration**: Automated testing and release workflows

---

## Contributing

We welcome contributions! See our [Contributor Guide](contributors/ContributorGuide.md) for details.

```bash
# Quick start for contributors
git clone https://github.com/kunalkushwaha/agentflow.git
cd agentflow
go mod tidy
go test ./...

# Generate docs
go run tools/docgen/main.go
```

## Community

- **[GitHub Discussions](https://github.com/kunalkushwaha/agentflow/discussions)** - Q&A and community
- **[Issues](https://github.com/kunalkushwaha/agentflow/issues)** - Bug reports and feature requests
- **[Contributing](CONTRIBUTING.md)** - How to contribute code and documentation

---

**[⭐ Star us on GitHub](https://github.com/kunalkushwaha/agentflow)** | **[📖 Full Documentation](https://agentflow.dev)** | **[🚀 Examples](examples/)**
