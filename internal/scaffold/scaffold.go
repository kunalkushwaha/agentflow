package scaffold

import (
	"fmt"
	"os"
	"path/filepath"
	"text/template"

	"github.com/kunalkushwaha/agentflow/internal/scaffold/templates"
	"github.com/kunalkushwaha/agentflow/internal/scaffold/utils"
)

// CreateAgentProject creates a new AgentFlow project (alias for CreateAgentProjectModular)
func CreateAgentProject(config ProjectConfig) error {
	return CreateAgentProjectModular(config)
}

// CreateAgentProjectFromConfig creates a new AgentFlow project using ProjectConfig (alias)
func CreateAgentProjectFromConfig(config ProjectConfig) error {
	return CreateAgentProjectModular(config)
}

// CreateAgentProjectModular creates a new AgentFlow project using the modular template system
func CreateAgentProjectModular(config ProjectConfig) error {
	// Create the main project directory
	if err := os.Mkdir(config.Name, 0755); err != nil {
		return fmt.Errorf("failed to create project directory %s: %w", config.Name, err)
	}
	fmt.Printf("Created directory: %s\n", config.Name)

	// Create go.mod file
	if err := createGoMod(config); err != nil {
		return err
	}

	// Create README.md file
	if err := createReadme(config); err != nil {
		return err
	}

	// Create agent files using the new template-based approach
	if err := createAgentFilesWithTemplates(config); err != nil {
		return err
	}

	// Create main.go file
	if err := createMainGoWithTemplate(config); err != nil {
		return err
	}

	// Create agentflow.toml configuration file
	if err := createConfig(config); err != nil {
		return err
	}

	// Generate Docker Compose files for database providers
	if config.MemoryEnabled && (config.MemoryProvider == "pgvector" || config.MemoryProvider == "weaviate") {
		dockerGenerator := NewDockerComposeGenerator(config)
		if err := dockerGenerator.GenerateDockerCompose(); err != nil {
			return fmt.Errorf("failed to generate Docker Compose files: %w", err)
		}
		if err := dockerGenerator.GenerateSetupScript(); err != nil {
			return fmt.Errorf("failed to generate setup scripts: %w", err)
		}
	}

	// Generate workflow diagrams if requested
	if config.Visualize {
		if err := generateWorkflowDiagrams(config); err != nil {
			return err
		}
	}

	fmt.Printf("\n✅ Project '%s' created successfully using modular templates!\n", config.Name)
	fmt.Printf("📁 Directory: %s\n", config.Name)
	fmt.Printf("🚀 Run: cd %s && go mod tidy && go run . -m \"Your message\"\n", config.Name)

	return nil
}

// createGoMod creates the go.mod file
func createGoMod(config ProjectConfig) error {
	goModContent := fmt.Sprintf("module %s\n\ngo 1.21\n\nrequire github.com/kunalkushwaha/agentflow %s\n", config.Name, AgentFlowVersion)
	goModPath := filepath.Join(config.Name, "go.mod")
	if err := os.WriteFile(goModPath, []byte(goModContent), 0644); err != nil {
		return fmt.Errorf("failed to create go.mod: %w", err)
	}
	fmt.Printf("Created file: %s\n", goModPath)
	return nil
}

// createReadme creates the README.md file
func createReadme(config ProjectConfig) error {
	content := fmt.Sprintf("# %s\n\n", config.Name)
	content += "**An AgentFlow project with intelligent multi-agent workflows**\n\n"
	content += "This project was generated using the AgentFlow CLI and includes production-ready features for building AI agent systems.\n\n"

	// Quick Start section
	content += "## 🚀 Quick Start\n\n"
	content += "```bash\n"
	content += "# Install dependencies\n"
	content += "go mod tidy\n\n"
	content += "# Run your agents with a message\n"
	content += "go run . -m \"Your message here\"\n\n"
	if config.MemoryEnabled && (config.MemoryProvider == "pgvector" || config.MemoryProvider == "weaviate") {
		content += "# Start the database (if using external memory provider)\n"
		if config.MemoryProvider == "pgvector" || config.MemoryProvider == "weaviate" {
			content += "docker compose up -d\n"
			content += "./setup.sh  # or setup.bat on Windows\n\n"
		}
	}
	content += "```\n\n"

	// Project Configuration section
	content += "## ⚙️ Project Configuration\n\n"
	content += "| Feature | Configuration |\n"
	content += "|---------|---------------|\n"
	content += fmt.Sprintf("| **Orchestration Mode** | %s |\n", config.OrchestrationMode)
	content += fmt.Sprintf("| **LLM Provider** | %s |\n", config.Provider)
	content += fmt.Sprintf("| **Number of Agents** | %d |\n", config.NumAgents)

	if config.MCPEnabled {
		content += "| **MCP Integration** | ✅ Enabled |\n"
		if config.MCPProduction {
			content += "| **MCP Features** | Production (caching, metrics, load balancing) |\n"
		}
	}

	if config.MemoryEnabled {
		content += "| **Memory System** | ✅ Enabled |\n"
		content += fmt.Sprintf("| **Memory Provider** | %s |\n", config.MemoryProvider)
		content += fmt.Sprintf("| **Embedding Provider** | %s |\n", config.EmbeddingProvider)
		if config.RAGEnabled {
			content += "| **RAG (Retrieval-Augmented Generation)** | ✅ Enabled |\n"
			content += fmt.Sprintf("| **RAG Configuration** | Chunk: %d tokens, Overlap: %d, Top-K: %d |\n", config.RAGChunkSize, config.RAGOverlap, config.RAGTopK)
		}
		if config.HybridSearch {
			content += "| **Hybrid Search** | ✅ Enabled (semantic + keyword) |\n"
		}
		if config.SessionMemory {
			content += "| **Session Memory** | ✅ Enabled |\n"
		}
	}

	if config.Visualize {
		content += "| **Workflow Visualization** | ✅ Enabled |\n"
	}

	content += "\n"

	// Agents section
	content += "## 🤖 Agents\n\n"
	agents := utils.ResolveAgentNames(convertToUtilsConfig(config))
	for i, agent := range agents {
		content += fmt.Sprintf("%d. **%s** (`%s`)\n", i+1, agent.DisplayName, agent.Name)
		content += fmt.Sprintf("   - **Purpose**: %s\n", agent.Purpose)
		content += fmt.Sprintf("   - **Role**: %s orchestration\n", config.OrchestrationMode)
		content += "\n"
	}

	// Usage Examples section
	content += "## 💡 Usage Examples\n\n"
	content += "### Basic Usage\n"
	content += "```bash\n"
	content += "# Ask a simple question\n"
	content += "go run . -m \"What is artificial intelligence?\"\n\n"
	content += "# Complex analysis request\n"
	content += "go run . -m \"Analyze the current trends in machine learning and provide recommendations\"\n"
	content += "```\n\n"

	if config.MemoryEnabled {
		content += "### Memory & RAG Usage\n"
		content += "```bash\n"
		content += "# The agents will automatically:\n"
		content += "# - Store conversation history\n"
		content += "# - Query relevant memories\n"
		if config.RAGEnabled {
			content += "# - Use RAG to enhance responses with knowledge base\n"
		}
		content += "# - Build context from previous interactions\n"
		content += "```\n\n"
	}

	if config.MemoryEnabled {
		content += "\n## Memory System\n\n"
		content += fmt.Sprintf("This project uses **%s** as the memory provider", config.MemoryProvider)
		if config.EmbeddingProvider == "openai" {
			content += " with **OpenAI embeddings**"
		}
		content += ".\n\n"

		if config.MemoryProvider == "pgvector" {
			content += "### PostgreSQL with pgvector Setup\n\n"
			content += "To use the pgvector memory provider, you need PostgreSQL with the pgvector extension:\n\n"
			content += "```bash\n"
			content += "# Using Docker\n"
			content += "docker run --name pgvector-db -e POSTGRES_PASSWORD=password -e POSTGRES_DB=agentflow -p 5432:5432 -d pgvector/pgvector:pg16\n"
			content += "\n"
			content += "# Update connection string in agentflow.toml:\n"
			content += "# Connection: \"postgres://user:password@localhost:15432/agentflow?sslmode=disable\"\n"
			content += "```\n\n"
		} else if config.MemoryProvider == "weaviate" {
			content += "### Weaviate Setup\n\n"
			content += "To use the Weaviate memory provider, you need Weaviate running:\n\n"
			content += "```bash\n"
			content += "# Using Docker\n"
			content += "docker run -d --name weaviate -p 8080:8080 -e QUERY_DEFAULTS_LIMIT=25 -e AUTHENTICATION_ANONYMOUS_ACCESS_ENABLED=true -e PERSISTENCE_DATA_PATH=/var/lib/weaviate -e DEFAULT_VECTORIZER_MODULE=none -e ENABLE_MODULES=text2vec-openai,text2vec-cohere,text2vec-huggingface,ref2vec-centroid,generative-openai,qna-openai semitechnologies/weaviate:latest\n"
			content += "```\n\n"
		}

		if config.EmbeddingProvider == "openai" {
			content += "### OpenAI API Key\n\n"
			content += "Set your OpenAI API key as an environment variable:\n\n"
			content += "```bash\n"
			content += "export OPENAI_API_KEY=\"your-api-key-here\"\n"
			content += "```\n\n"
		} else if config.EmbeddingProvider == "ollama" {
			content += "### Ollama Setup\n\n"
			content += "Make sure Ollama is running and the embedding model is installed:\n\n"
			content += "```bash\n"
			content += "# Start Ollama (if not already running)\n"
			content += "ollama serve\n"
			content += "\n"
			content += "# Install the embedding model\n"
			content += fmt.Sprintf("ollama pull %s\n", config.EmbeddingModel)
			content += "```\n\n"
		}

		if config.RAGEnabled {
			content += "### RAG Features\n\n"
			content += "This project includes RAG (Retrieval-Augmented Generation) capabilities:\n\n"
			content += fmt.Sprintf("- **Chunk Size**: %d tokens\n", config.RAGChunkSize)
			content += fmt.Sprintf("- **Overlap**: %d tokens\n", config.RAGOverlap)
			content += fmt.Sprintf("- **Top-K Results**: %d\n", config.RAGTopK)
			content += fmt.Sprintf("- **Score Threshold**: %.1f\n", config.RAGScoreThreshold)
			if config.HybridSearch {
				content += "- **Hybrid Search**: Enabled (semantic + keyword)\n"
			}
			if config.SessionMemory {
				content += "- **Session Memory**: Enabled\n"
			}
			content += "\n"
		}
	}

	if config.Visualize {
		content += "\n## 📊 Workflow Diagrams\n\n"
		content += fmt.Sprintf("Visual workflow diagrams have been generated in the `%s` directory. These diagrams show the orchestration pattern and agent interactions for this project.\n\n", config.VisualizeOutputDir)
		content += "The diagrams include:\n"
		content += "- **Workflow Overview**: High-level agent interaction patterns\n"
		content += "- **Agent Details**: Individual agent roles and responsibilities\n"
		content += "- **Configuration Summary**: Key settings and parameters\n\n"
	}

	// Documentation and Resources section
	content += "## 📚 Documentation & Resources\n\n"
	content += "### AgentFlow Documentation\n"
	content += "- **[AgentFlow Documentation](https://github.com/kunalkushwaha/agentflow/tree/main/docs)** - Complete framework documentation\n"
	content += "- **[Agent Basics Guide](https://github.com/kunalkushwaha/agentflow/blob/main/docs/guides/AgentBasics.md)** - Understanding AgentHandler interface and patterns\n"
	content += "- **[Configuration Guide](https://github.com/kunalkushwaha/agentflow/blob/main/docs/guides/Configuration.md)** - Managing agentflow.toml and environment setup\n"
	content += "- **[Examples & Tutorials](https://github.com/kunalkushwaha/agentflow/blob/main/docs/guides/Examples.md)** - Practical examples and code samples\n\n"

	if config.MemoryEnabled {
		content += "### Memory & RAG Documentation\n"
		content += "- **[Memory System Guide](https://github.com/kunalkushwaha/agentflow/blob/main/docs/guides/Memory.md)** - Complete memory implementation guide\n"
		content += "- **[RAG Configuration Guide](https://github.com/kunalkushwaha/agentflow/blob/main/docs/guides/RAGConfiguration.md)** - RAG configuration and best practices\n"
		content += "- **[Memory Quick Reference](https://github.com/kunalkushwaha/agentflow/blob/main/docs/memory_quick_reference.md)** - Essential memory API reference\n\n"
	}

	if config.MCPEnabled {
		content += "### MCP Integration Documentation\n"
		content += "- **[Tool Integration Guide](https://github.com/kunalkushwaha/agentflow/blob/main/docs/guides/ToolIntegration.md)** - MCP protocol and dynamic tool discovery\n"
		content += "- **[Custom Tools Guide](https://github.com/kunalkushwaha/agentflow/blob/main/docs/guides/CustomTools.md)** - Building your own MCP servers\n\n"
	}

	content += "### Advanced Topics\n"
	content += "- **[Multi-Agent Orchestration](https://github.com/kunalkushwaha/agentflow/blob/main/docs/multi_agent_orchestration.md)** - Advanced orchestration patterns and configuration\n"
	content += "- **[Production Deployment](https://github.com/kunalkushwaha/agentflow/blob/main/docs/guides/Production.md)** - Scaling, monitoring, and best practices\n"
	content += "- **[Performance Tuning](https://github.com/kunalkushwaha/agentflow/blob/main/docs/guides/Performance.md)** - Optimization and benchmarking\n"
	content += "- **[Error Handling](https://github.com/kunalkushwaha/agentflow/blob/main/docs/guides/ErrorHandling.md)** - Resilient agent workflows\n\n"

	// Troubleshooting section
	content += "## 🔧 Troubleshooting\n\n"
	content += "### Common Issues\n\n"
	
	if config.MemoryEnabled {
		content += "**Memory System Issues:**\n"
		if config.MemoryProvider == "pgvector" {
			content += "- **Database Connection Failed**: Ensure PostgreSQL with pgvector is running (`docker compose up -d`)\n"
			content += "- **Permission Denied**: Check database user permissions and connection string\n"
			content += "- **Vector Extension Missing**: Verify pgvector extension is installed in the database\n\n"
		} else if config.MemoryProvider == "weaviate" {
			content += "- **Weaviate Connection Failed**: Ensure Weaviate is running (`docker compose up -d`)\n"
			content += "- **Schema Issues**: Check Weaviate class configuration and vector dimensions\n\n"
		}
		
		if config.EmbeddingProvider == "openai" {
			content += "- **OpenAI API Errors**: Verify `OPENAI_API_KEY` environment variable is set\n"
			content += "- **Rate Limiting**: Implement retry logic or upgrade your OpenAI plan\n\n"
		} else if config.EmbeddingProvider == "ollama" {
			content += "- **Ollama Connection Failed**: Ensure Ollama is running (`ollama serve`)\n"
			content += fmt.Sprintf("- **Model Not Found**: Install the embedding model (`ollama pull %s`)\n\n", config.EmbeddingModel)
		}
	}

	content += "**General Issues:**\n"
	content += "- **LLM Provider Errors**: Check API keys and endpoint configurations in `agentflow.toml`\n"
	content += "- **Agent Not Responding**: Verify LLM provider credentials and network connectivity\n"
	content += "- **Build Errors**: Run `go mod tidy` to ensure all dependencies are installed\n\n"

	if config.MCPEnabled {
		content += "**MCP Integration Issues:**\n"
		content += "- **Tools Not Available**: Check MCP server configuration and connectivity\n"
		content += "- **Tool Execution Failed**: Verify tool permissions and required dependencies\n\n"
	}

	content += "### Getting Help\n\n"
	content += "- **[GitHub Issues](https://github.com/kunalkushwaha/agentflow/issues)** - Report bugs and request features\n"
	content += "- **[Discussions](https://github.com/kunalkushwaha/agentflow/discussions)** - Community support and questions\n"
	content += "- **[Documentation](https://github.com/kunalkushwaha/agentflow/tree/main/docs)** - Comprehensive guides and API reference\n\n"

	// Project Structure section
	content += "## 📁 Project Structure\n\n"
	content += "```\n"
	content += fmt.Sprintf("%s/\n", config.Name)
	content += "├── main.go                 # Application entry point\n"
	content += "├── agentflow.toml         # Configuration file\n"
	agents = utils.ResolveAgentNames(convertToUtilsConfig(config))
	for _, agent := range agents {
		content += fmt.Sprintf("├── %s              # %s agent implementation\n", agent.FileName, agent.DisplayName)
	}
	if config.MemoryEnabled && (config.MemoryProvider == "pgvector" || config.MemoryProvider == "weaviate") {
		content += "├── docker-compose.yml     # Database services\n"
		content += "├── init-db.sql           # Database initialization\n"
		content += "├── setup.sh/.bat         # Setup scripts\n"
		content += "├── .env.example          # Environment template\n"
	}
	if config.Visualize {
		content += fmt.Sprintf("├── %s/           # Workflow diagrams\n", config.VisualizeOutputDir)
	}
	content += "├── go.mod                 # Go module definition\n"
	content += "└── README.md              # This file\n"
	content += "```\n\n"

	// Next Steps section
	content += "## 🚀 Next Steps\n\n"
	content += "1. **Customize Agents**: Modify agent logic in the generated `.go` files\n"
	content += "2. **Configure Providers**: Update `agentflow.toml` with your API keys and settings\n"
	if config.MemoryEnabled {
		content += "3. **Add Knowledge**: Use the memory system to store and retrieve information\n"
		if config.RAGEnabled {
			content += "4. **Ingest Documents**: Add documents to your knowledge base for RAG\n"
		}
	}
	if config.MCPEnabled {
		content += "3. **Add Tools**: Configure additional MCP tools for enhanced capabilities\n"
	}
	content += fmt.Sprintf("%d. **Scale Up**: Add more agents or modify orchestration patterns as needed\n", 3 + (func() int { count := 0; if config.MemoryEnabled { count++; if config.RAGEnabled { count++ } }; if config.MCPEnabled { count++ }; return count })())
	content += fmt.Sprintf("%d. **Deploy**: Follow the [Production Guide](https://github.com/kunalkushwaha/agentflow/blob/main/docs/guides/Production.md) for deployment\n\n", 4 + (func() int { count := 0; if config.MemoryEnabled { count++; if config.RAGEnabled { count++ } }; if config.MCPEnabled { count++ }; return count })())

	content += "---\n\n"
	content += "*Generated with ❤️ by [AgentFlow](https://github.com/kunalkushwaha/agentflow)*\n"

	readmePath := filepath.Join(config.Name, "README.md")
	if err := os.WriteFile(readmePath, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to create README.md: %w", err)
	}
	fmt.Printf("Created file: %s\n", readmePath)
	return nil
}

// createAgentFilesWithTemplates creates agent files using the template system
func createAgentFilesWithTemplates(config ProjectConfig) error {
	utilsConfig := convertToUtilsConfig(config)
	agents := utils.ResolveAgentNames(utilsConfig)

	if len(agents) == 0 {
		agents = append(agents, utils.CreateAgentInfo("agent1", config.OrchestrationMode))
	}

	// Validate agent names
	if err := utils.ValidateAgentNames(agents); err != nil {
		return fmt.Errorf("agent name validation failed: %w", err)
	}

	// Use the comprehensive template from templates package
	tmpl, err := template.New("agent").Parse(templates.AgentTemplate)
	if err != nil {
		return fmt.Errorf("failed to parse agent template: %w", err)
	}

	// Generate each agent file
	for i, agent := range agents {
		var nextAgent string
		var routingComment string

		if i < len(agents)-1 {
			// Route to next agent in the list
			nextAgent = agents[i+1].Name
			routingComment = fmt.Sprintf("Route to the next agent (%s) in the workflow", agents[i+1].DisplayName)
		} else if config.ResponsibleAI {
			// Last agent routes to responsible AI
			nextAgent = "responsible_ai"
			routingComment = "Route to Responsible AI for final content check"
		} else {
			// Last agent - workflow completion (no routing)
			nextAgent = ""
			routingComment = "Workflow completion"
		}

		// Create system prompt for this agent
		systemPrompt := utils.CreateSystemPrompt(agent, i, len(agents), config.OrchestrationMode)

		// Create template data structure that matches the comprehensive template
		templateData := struct {
			Config         ProjectConfig
			Agent          utils.AgentInfo
			Agents         []utils.AgentInfo
			AgentIndex     int
			TotalAgents    int
			NextAgent      string
			PrevAgent      string
			IsFirstAgent   bool
			IsLastAgent    bool
			SystemPrompt   string
			RoutingComment string
		}{
			Config:         config,
			Agent:          agent,
			Agents:         agents,
			AgentIndex:     i,
			TotalAgents:    len(agents),
			NextAgent:      nextAgent,
			IsFirstAgent:   i == 0,
			IsLastAgent:    i == len(agents)-1,
			SystemPrompt:   systemPrompt,
			RoutingComment: routingComment,
		}

		filePath := filepath.Join(config.Name, agent.FileName)
		file, err := os.Create(filePath)
		if err != nil {
			return fmt.Errorf("failed to create file %s: %w", filePath, err)
		}

		if err := tmpl.Execute(file, templateData); err != nil {
			file.Close()
			return fmt.Errorf("failed to execute template for %s: %w", agent.FileName, err)
		}
		file.Close()

		fmt.Printf("Created file: %s (%s agent)\n", filePath, agent.DisplayName)
	}

	return nil
}

// createMainGoWithTemplate creates main.go using templates
func createMainGoWithTemplate(config ProjectConfig) error {
	utilsConfig := convertToUtilsConfig(config)
	agents := utils.ResolveAgentNames(utilsConfig)

	// Use the comprehensive template from templates package
	tmpl, err := template.New("main").Parse(templates.MainTemplate)
	if err != nil {
		return fmt.Errorf("failed to parse main template: %w", err)
	}

	// Create template data structure that matches the comprehensive template
	templateData := struct {
		Config               ProjectConfig
		Agents               []utils.AgentInfo
		ProviderInitFunction string
		MCPInitFunction      string
		CacheInitFunction    string
	}{
		Config:               config,
		Agents:               agents,
		ProviderInitFunction: generateProviderInitFunction(config),
		MCPInitFunction:      generateMCPInitFunction(config),
		CacheInitFunction:    generateCacheInitFunction(config),
	}

	mainGoPath := filepath.Join(config.Name, "main.go")
	file, err := os.Create(mainGoPath)
	if err != nil {
		return fmt.Errorf("failed to create main.go: %w", err)
	}
	defer file.Close()

	if err := tmpl.Execute(file, templateData); err != nil {
		return fmt.Errorf("failed to execute main template: %w", err)
	}

	fmt.Printf("Created file: %s\n", mainGoPath)
	return nil
}

// createConfig creates the agentflow.toml configuration file
func createConfig(config ProjectConfig) error {
	configContent := fmt.Sprintf(`# AgentFlow Configuration

[agent_flow]
name = "%s"
version = "1.0.0"
provider = "%s"

[logging]
level = "info"
format = "json"

[runtime]
max_concurrent_agents = 10
timeout_seconds = 30

[providers.azure]
# API key will be read from AZURE_OPENAI_API_KEY environment variable
# Endpoint will be read from AZURE_OPENAI_ENDPOINT environment variable
# Deployment will be read from AZURE_OPENAI_DEPLOYMENT environment variable

[providers.openai]
# API key will be read from OPENAI_API_KEY environment variable

[providers.ollama]
base_url = "http://localhost:11434"
model = "llama2"

[providers.mock]
# Mock provider for testing - no configuration needed
`, config.Name, config.Provider)

	// Add MCP configuration if enabled
	if config.MCPEnabled {
		mcpConfig := `
[mcp]
enabled = true
enable_discovery = true
connection_timeout = 5000
max_retries = 3
retry_delay = 1000
enable_caching = true
cache_timeout = 300000
max_connections = 10

# Example MCP servers - configure as needed
[[mcp.servers]]
name = "docker"
type = "tcp"
host = "localhost"
port = 8811
enabled = false

[[mcp.servers]]
name = "filesystem"
type = "stdio"
command = "npx @modelcontextprotocol/server-filesystem /path/to/allowed/files"
enabled = false

[[mcp.servers]]
name = "brave-search"
type = "stdio"
command = "npx @modelcontextprotocol/server-brave-search"
enabled = false
`
		configContent += mcpConfig
	}

	// Add memory configuration if enabled
	if config.MemoryEnabled {
		// Get embedding dimensions from intelligence system
		dimensions := GetModelDimensions(config.EmbeddingProvider, config.EmbeddingModel)
		
		memoryConfig := fmt.Sprintf(`
[agent_memory]
provider = "%s"
connection = "%s"
max_results = %d
dimensions = %d
auto_embed = true
enable_knowledge_base = true
knowledge_max_results = %d
knowledge_score_threshold = %.1f
chunk_size = %d
chunk_overlap = %d
enable_rag = %t
rag_max_context_tokens = 4000
rag_personal_weight = 0.3
rag_knowledge_weight = 0.7
rag_include_sources = true

[agent_memory.embedding]
provider = "%s"
model = "%s"`,
			config.MemoryProvider,
			getConnectionString(config.MemoryProvider),
			config.RAGTopK,
			dimensions,
			config.RAGTopK,
			config.RAGScoreThreshold,
			config.RAGChunkSize,
			config.RAGOverlap,
			config.RAGEnabled,
			config.EmbeddingProvider,
			config.EmbeddingModel)

		// Add provider-specific embedding configuration
		if config.EmbeddingProvider == "ollama" {
			memoryConfig += `
base_url = "http://localhost:11434"`
		}
		
		memoryConfig += `
cache_embeddings = true
max_batch_size = 100
timeout_seconds = 30

[agent_memory.documents]
auto_chunk = true
supported_types = ["pdf", "txt", "md", "web", "code"]
max_file_size = "10MB"
enable_metadata_extraction = true
enable_url_scraping = true

[agent_memory.search]
hybrid_search = ` + fmt.Sprintf("%t", config.HybridSearch) + `
keyword_weight = 0.3
semantic_weight = 0.7
enable_reranking = false
enable_query_expansion = false
`

		configContent += memoryConfig
	}

	configPath := filepath.Join(config.Name, "agentflow.toml")
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		return fmt.Errorf("failed to create agentflow.toml: %w", err)
	}
	fmt.Printf("Created file: %s\n", configPath)
	return nil
}

// generateWorkflowDiagrams creates Mermaid workflow diagrams for the project
func generateWorkflowDiagrams(config ProjectConfig) error {
	// Create the output directory for diagrams
	diagramsDir := filepath.Join(config.Name, config.VisualizeOutputDir)
	if err := os.MkdirAll(diagramsDir, 0755); err != nil {
		return fmt.Errorf("failed to create diagrams directory: %w", err)
	}
	fmt.Printf("Created directory: %s\n", diagramsDir)

	// Generate diagram based on orchestration mode
	var diagram string
	var title string

	switch config.OrchestrationMode {
	case "collaborative":
		diagram, title = generateCollaborativeDiagram(config)
	case "sequential":
		diagram, title = generateSequentialDiagram(config)
	case "loop":
		diagram, title = generateLoopDiagram(config)
	case "mixed":
		diagram, title = generateMixedDiagram(config)
	default:
		diagram, title = generateRouteDiagram(config)
	}

	// Create the diagram file
	diagramPath := filepath.Join(diagramsDir, "workflow.md")
	content := fmt.Sprintf(`# %s Workflow

## Overview
This diagram shows the %s orchestration pattern used in this project.

## Workflow Diagram

%s

## Configuration
- **Orchestration Mode**: %s
- **Number of Agents**: %d
- **Timeout**: %d seconds
- **Max Concurrency**: %d
- **Failure Threshold**: %.2f

## Agent Details
%s
`, title, config.OrchestrationMode, diagram, config.OrchestrationMode, config.NumAgents, config.OrchestrationTimeout, config.MaxConcurrency, config.FailureThreshold, generateAgentDetails(config))

	if err := os.WriteFile(diagramPath, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to create workflow diagram: %w", err)
	}
	fmt.Printf("Created workflow diagram: %s\n", diagramPath)

	return nil
}

// generateCollaborativeDiagram creates a collaborative orchestration diagram
func generateCollaborativeDiagram(config ProjectConfig) (string, string) {
	agents := config.CollaborativeAgents
	if len(agents) == 0 {
		// Use default agent names if none specified
		agents = make([]string, config.NumAgents)
		for i := 0; i < config.NumAgents; i++ {
			agents[i] = fmt.Sprintf("agent%d", i+1)
		}
	}

	diagram := "```mermaid\n---\ntitle: Collaborative Orchestration\n---\nflowchart TD\n"
	diagram += "    EVENT[\"📨 Input Event\"]\n"
	diagram += "    ORCHESTRATOR[\"🎯 Collaborative Orchestrator\"]\n"
	diagram += "    AGGREGATOR[\"📊 Result Aggregator\"]\n"
	diagram += "    RESULT[\"📤 Final Result\"]\n\n"
	diagram += "    EVENT --> ORCHESTRATOR\n"

	for i, agent := range agents {
		agentId := fmt.Sprintf("AGENT%d", i+1)
		diagram += fmt.Sprintf("    %s[\"🤖 %s\"]\n", agentId, agent)
		diagram += fmt.Sprintf("    ORCHESTRATOR --> %s\n", agentId)
		diagram += fmt.Sprintf("    %s --> AGGREGATOR\n", agentId)
	}

	diagram += "    AGGREGATOR --> RESULT\n"
	diagram += "```"

	return diagram, "Collaborative"
}

// generateSequentialDiagram creates a sequential orchestration diagram
func generateSequentialDiagram(config ProjectConfig) (string, string) {
	agents := config.SequentialAgents
	if len(agents) == 0 {
		// Use default agent names if none specified
		agents = make([]string, config.NumAgents)
		for i := 0; i < config.NumAgents; i++ {
			agents[i] = fmt.Sprintf("agent%d", i+1)
		}
	}

	diagram := "```mermaid\n---\ntitle: Sequential Pipeline\n---\nflowchart TD\n"
	diagram += "    INPUT[\"📨 Input Event\"]\n"

	var prevNode = "INPUT"
	for i, agent := range agents {
		agentId := fmt.Sprintf("AGENT%d", i+1)
		diagram += fmt.Sprintf("    %s[\"🤖 %s\"]\n", agentId, agent)
		diagram += fmt.Sprintf("    %s --> %s\n", prevNode, agentId)
		prevNode = agentId
	}

	diagram += "    OUTPUT[\"📤 Final Result\"]\n"
	diagram += fmt.Sprintf("    %s --> OUTPUT\n", prevNode)
	diagram += "```"

	return diagram, "Sequential Pipeline"
}

// getConnectionString returns the appropriate connection string for a memory provider
func getConnectionString(memoryProvider string) string {
	switch memoryProvider {
	case "pgvector":
		return "postgres://user:password@localhost:15432/agentflow?sslmode=disable"
	case "weaviate":
		return "http://localhost:8080"
	default:
		return "memory"
	}
}

// generateLoopDiagram creates a loop orchestration diagram
func generateLoopDiagram(config ProjectConfig) (string, string) {
	agentName := config.LoopAgent
	if agentName == "" {
		agentName = "processor"
	}

	diagram := "```mermaid\n---\ntitle: Loop Processing\n---\nflowchart TD\n"
	diagram += "    INPUT[\"📨 Input Event\"]\n"
	diagram += "    AGENT[\"🤖 " + agentName + "\"]\n"
	diagram += "    CONDITION{\"🔄 Continue Loop?\"}\n"
	diagram += "    OUTPUT[\"📤 Final Result\"]\n\n"
	diagram += "    INPUT --> AGENT\n"
	diagram += "    AGENT --> CONDITION\n"
	diagram += "    CONDITION -->|Yes| AGENT\n"
	diagram += "    CONDITION -->|No| OUTPUT\n"
	diagram += fmt.Sprintf("    CONDITION -.->|Max %d iterations| OUTPUT\n", config.MaxIterations)
	diagram += "```"

	return diagram, "Loop Processing"
}

// generateMixedDiagram creates a mixed orchestration diagram
func generateMixedDiagram(config ProjectConfig) (string, string) {
	diagram := "```mermaid\n---\ntitle: Mixed Orchestration\n---\nflowchart TD\n"
	diagram += "    INPUT[\"📨 Input Event\"]\n"
	diagram += "    PHASE1[\"🤝 Collaborative Phase\"]\n"
	diagram += "    PHASE2[\"🎭 Sequential Phase\"]\n"
	diagram += "    OUTPUT[\"📤 Final Result\"]\n\n"
	diagram += "    INPUT --> PHASE1\n"

	// Add collaborative agents
	if len(config.CollaborativeAgents) > 0 {
		for i, agent := range config.CollaborativeAgents {
			agentId := fmt.Sprintf("COLLAB%d", i+1)
			diagram += fmt.Sprintf("    %s[\"🤖 %s\"]\n", agentId, agent)
			diagram += fmt.Sprintf("    PHASE1 --> %s\n", agentId)
			diagram += fmt.Sprintf("    %s --> PHASE2\n", agentId)
		}
	}

	// Add sequential agents
	if len(config.SequentialAgents) > 0 {
		var prevNode = "PHASE2"
		for i, agent := range config.SequentialAgents {
			agentId := fmt.Sprintf("SEQ%d", i+1)
			diagram += fmt.Sprintf("    %s[\"🤖 %s\"]\n", agentId, agent)
			diagram += fmt.Sprintf("    %s --> %s\n", prevNode, agentId)
			prevNode = agentId
		}
		diagram += fmt.Sprintf("    %s --> OUTPUT\n", prevNode)
	} else {
		diagram += "    PHASE2 --> OUTPUT\n"
	}

	diagram += "```"

	return diagram, "Mixed Orchestration"
}

// generateRouteDiagram creates a route orchestration diagram
func generateRouteDiagram(config ProjectConfig) (string, string) {
	diagram := "```mermaid\n---\ntitle: Route Orchestration\n---\nflowchart TD\n"
	diagram += "    INPUT[\"📨 Input Event\"]\n"
	diagram += "    ROUTER[\"🎯 Event Router\"]\n"
	diagram += "    OUTPUT[\"📤 Result\"]\n\n"
	diagram += "    INPUT --> ROUTER\n"

	for i := 0; i < config.NumAgents; i++ {
		agentId := fmt.Sprintf("AGENT%d", i+1)
		agentName := fmt.Sprintf("agent%d", i+1)
		diagram += fmt.Sprintf("    %s[\"🤖 %s\"]\n", agentId, agentName)
		diagram += fmt.Sprintf("    ROUTER -.->|Route| %s\n", agentId)
		diagram += fmt.Sprintf("    %s --> OUTPUT\n", agentId)
	}

	diagram += "```"

	return diagram, "Route Orchestration"
}

// generateAgentDetails creates detailed agent information
func generateAgentDetails(config ProjectConfig) string {
	details := ""

	switch config.OrchestrationMode {
	case "collaborative":
		if len(config.CollaborativeAgents) > 0 {
			details += "### Collaborative Agents\n"
			for i, agent := range config.CollaborativeAgents {
				details += fmt.Sprintf("%d. **%s**: Processes events in parallel with other agents\n", i+1, agent)
			}
		}
	case "sequential":
		if len(config.SequentialAgents) > 0 {
			details += "### Sequential Agents\n"
			for i, agent := range config.SequentialAgents {
				details += fmt.Sprintf("%d. **%s**: Processes events in pipeline order\n", i+1, agent)
			}
		}
	case "loop":
		if config.LoopAgent != "" {
			details += "### Loop Agent\n"
			details += fmt.Sprintf("1. **%s**: Processes events iteratively up to %d times\n", config.LoopAgent, config.MaxIterations)
		}
	case "mixed":
		if len(config.CollaborativeAgents) > 0 {
			details += "### Collaborative Agents (Phase 1)\n"
			for i, agent := range config.CollaborativeAgents {
				details += fmt.Sprintf("%d. **%s**: Processes events in parallel\n", i+1, agent)
			}
		}
		if len(config.SequentialAgents) > 0 {
			details += "\n### Sequential Agents (Phase 2)\n"
			for i, agent := range config.SequentialAgents {
				details += fmt.Sprintf("%d. **%s**: Processes events in pipeline order\n", i+1, agent)
			}
		}
	default:
		details += "### Route Agents\n"
		for i := 0; i < config.NumAgents; i++ {
			details += fmt.Sprintf("%d. **agent%d**: Processes events based on routing logic\n", i+1, i+1)
		}
	}

	return details
}

// convertToUtilsConfig converts ProjectConfig to utils.ProjectConfig
func convertToUtilsConfig(config ProjectConfig) utils.ProjectConfig {
	return utils.ProjectConfig{
		Name:                 config.Name,
		NumAgents:            config.NumAgents,
		Provider:             config.Provider,
		ResponsibleAI:        config.ResponsibleAI,
		ErrorHandler:         config.ErrorHandler,
		MCPEnabled:           config.MCPEnabled,
		MCPProduction:        config.MCPProduction,
		WithCache:            config.WithCache,
		WithMetrics:          config.WithMetrics,
		MCPTools:             config.MCPTools,
		MCPServers:           config.MCPServers,
		CacheBackend:         config.CacheBackend,
		MetricsPort:          config.MetricsPort,
		WithLoadBalancer:     config.WithLoadBalancer,
		ConnectionPoolSize:   config.ConnectionPoolSize,
		RetryPolicy:          config.RetryPolicy,
		OrchestrationMode:    config.OrchestrationMode,
		CollaborativeAgents:  config.CollaborativeAgents,
		SequentialAgents:     config.SequentialAgents,
		LoopAgent:            config.LoopAgent,
		MaxIterations:        config.MaxIterations,
		OrchestrationTimeout: config.OrchestrationTimeout,
		FailureThreshold:     config.FailureThreshold,
		MaxConcurrency:       config.MaxConcurrency,
		Visualize:            config.Visualize,
		VisualizeOutputDir:   config.VisualizeOutputDir,
	}
}

// generateProviderInitFunction generates the provider initialization function code
func generateProviderInitFunction(config ProjectConfig) string {
	return `func initializeProvider(providerType string) (core.ModelProvider, error) {
	// Use the config-based provider initialization
	return core.NewProviderFromWorkingDir()
}`
}

// generateMCPInitFunction generates the MCP initialization function code
func generateMCPInitFunction(config ProjectConfig) string {
	return `func initializeMCP() (core.MCPManager, error) {
	// Load configuration from agentflow.toml in current directory
	config, err := core.LoadConfigFromWorkingDir()
	if err != nil {
		return nil, fmt.Errorf("failed to load configuration: %w", err)
	}

	// Check if MCP is enabled in configuration
	if !config.MCP.Enabled {
		return nil, fmt.Errorf("MCP is not enabled in agentflow.toml")
	}

	// Convert TOML config to MCP config
	mcpConfig := core.MCPConfig{
		EnableDiscovery:   config.MCP.EnableDiscovery,
		ConnectionTimeout: time.Duration(config.MCP.ConnectionTimeout) * time.Millisecond,
		MaxRetries:        config.MCP.MaxRetries,
		RetryDelay:        time.Duration(config.MCP.RetryDelay) * time.Millisecond,
		EnableCaching:     config.MCP.EnableCaching,
		CacheTimeout:      time.Duration(config.MCP.CacheTimeout) * time.Millisecond,
		MaxConnections:    config.MCP.MaxConnections,
		Servers:           make([]core.MCPServerConfig, len(config.MCP.Servers)),
	}

	// Convert server configurations
	for i, server := range config.MCP.Servers {
		mcpConfig.Servers[i] = core.MCPServerConfig{
			Name:    server.Name,
			Type:    server.Type,
			Host:    server.Host,
			Port:    server.Port,
			Command: server.Command,
			Enabled: server.Enabled,
		}
	}

	// Initialize MCP manager with configuration from TOML
	err = core.InitializeMCP(mcpConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize MCP: %w", err)
	}

	// Get the initialized MCP manager
	manager := core.GetMCPManager()
	if manager == nil {
		return nil, fmt.Errorf("MCP manager not available after initialization")
	}

	return manager, nil
}`
}

// generateCacheInitFunction generates the cache initialization function code
func generateCacheInitFunction(config ProjectConfig) string {
	return `func initializeCache() error {
	// Cache initialization placeholder
	return nil
}`
}
