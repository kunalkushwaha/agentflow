package cmd

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/kunalkushwaha/agentflow/internal/scaffold"
	"github.com/spf13/cobra"
)

// createCmd represents the create command
var createCmd = &cobra.Command{
	Use:   "create [project-name]",
	Short: "Create a new AgentFlow project with multi-agent workflows, memory, and MCP integration",
	Long: `Create a new AgentFlow project with customizable multi-agent workflows.

This command generates a complete project structure including:
  * Multi-agent workflow implementation with various orchestration patterns
  * Memory system with RAG (Retrieval-Augmented Generation) capabilities
  * MCP (Model Context Protocol) tool integration
  * Configuration files (agentflow.toml) with intelligent defaults
  * Docker Compose files for database providers (PostgreSQL, Weaviate)
  * Error handling and responsible AI agents
  * Production-ready features (caching, metrics, load balancing)
  * Workflow visualization with Mermaid diagrams

BASIC EXAMPLES:
  # Simple project with 2 agents
  agentcli create myproject

  # Project with specific LLM provider and agent count
  agentcli create myproject --agents 3 --provider azure

  # Interactive mode for guided setup (recommended for beginners)
  agentcli create --interactive

MEMORY & RAG EXAMPLES:
  # Basic memory-enabled project (in-memory storage)
  agentcli create myproject --memory-enabled

  # PostgreSQL with vector search and RAG
  agentcli create myproject --memory-enabled --memory-provider pgvector --rag-enabled

  # Weaviate with OpenAI embeddings
  agentcli create myproject --memory-enabled --memory-provider weaviate --embedding-provider openai

  # Advanced RAG with custom settings
  agentcli create myproject --memory-enabled --memory-provider pgvector --rag-enabled \
    --rag-chunk-size 512 --rag-overlap 50 --rag-top-k 3 --rag-score-threshold 0.8

  # Hybrid search (semantic + keyword) with session memory
  agentcli create myproject --memory-enabled --memory-provider pgvector --rag-enabled \
    --hybrid-search --session-memory

  # Local embeddings with Ollama (recommended)
  agentcli create myproject --memory-enabled --memory-provider pgvector --rag-enabled \
    --embedding-provider ollama --embedding-model nomic-embed-text:latest

ORCHESTRATION EXAMPLES:
  # Collaborative workflow (agents work in parallel)
  agentcli create myworkflow --orchestration-mode collaborative \
    --collaborative-agents "analyzer,processor,validator"

  # Sequential pipeline (agents work in sequence)
  agentcli create mypipeline --orchestration-mode sequential \
    --sequential-agents "analyzer,transformer,validator"

  # Loop-based workflow (single agent with iterations)
  agentcli create myloop --orchestration-mode loop --loop-agent processor --max-iterations 5

  # Mixed orchestration with fault tolerance
  agentcli create myworkflow --orchestration-mode mixed \
    --collaborative-agents "analyzer,validator" --sequential-agents "processor,finalizer" \
    --failure-threshold 0.8 --max-concurrency 10

MCP INTEGRATION EXAMPLES:
  # Basic MCP with common tools
  agentcli create myproject --mcp-enabled

  # Production MCP with caching and metrics
  agentcli create myproject --mcp-production --with-cache --with-metrics

  # MCP with specific tools and servers
  agentcli create myproject --mcp-enabled \
    --mcp-tools "web_search,summarize,translate" --mcp-servers "docker,web-service"

VISUALIZATION EXAMPLES:
  # Generate workflow diagrams
  agentcli create myproject --visualize --visualize-output "docs/diagrams"

COMPLETE EXAMPLES:
  # Full-featured project with everything enabled
  agentcli create myproject --memory-enabled --memory-provider pgvector --rag-enabled \
    --mcp-enabled --visualize --orchestration-mode collaborative

  # Production-ready RAG system
  agentcli create knowledge-base --memory-enabled --memory-provider pgvector --rag-enabled \
    --embedding-provider openai --hybrid-search --session-memory --mcp-production

MEMORY PROVIDERS:
  * memory     - In-memory storage (fast, temporary, good for development)
  * pgvector   - PostgreSQL with vector extension (persistent, production-ready)
  * weaviate   - Dedicated vector database (scalable, advanced features)

EMBEDDING PROVIDERS:
  * dummy      - Simple embeddings for testing (default)
  * openai     - OpenAI embeddings (requires OPENAI_API_KEY)
  * ollama     - Local embeddings with Ollama (requires Ollama running)

For more information on setup and configuration, see the generated README.md file.`,
	Args: func(cmd *cobra.Command, args []string) error {
		interactive, _ := cmd.Flags().GetBool("interactive")
		if !interactive && len(args) != 1 {
			return fmt.Errorf("project name is required (or use --interactive)")
		}
		return nil
	},
	RunE: runCreateCommand,
}

// Command flags
var (
	// Basic project flags
	numAgents     int
	provider      string
	responsibleAI bool
	errorHandler  bool
	interactive   bool

	// Multi-agent orchestration flags
	orchestrationMode    string
	collaborativeAgents  string
	sequentialAgents     string
	loopAgent            string
	maxIterations        int
	orchestrationTimeout int
	failureThreshold     float64
	maxConcurrency       int

	// Visualization flags
	visualize          bool
	visualizeOutputDir string

	// MCP flags
	mcpEnabled         bool
	mcpProduction      bool
	withCache          bool
	withMetrics        bool
	mcpTools           string
	mcpServers         string
	cacheBackend       string
	metricsPort        int
	withLoadBalancer   bool
	connectionPoolSize int
	retryPolicy        string

	// Memory/RAG flags
	memoryEnabled     bool
	memoryProvider    string
	embeddingProvider string
	embeddingModel    string
	ragEnabled        bool
	ragChunkSize      int
	ragOverlap        int
	ragTopK           int
	ragScoreThreshold float64
	hybridSearch      bool
	sessionMemory     bool
)

func init() {
	rootCmd.AddCommand(createCmd)

	// Basic project flags
	createCmd.Flags().IntVarP(&numAgents, "agents", "a", 2, "Number of agents to create")
	createCmd.Flags().StringVarP(&provider, "provider", "p", "azure", "LLM provider (openai, azure, ollama, mock)")
	createCmd.Flags().BoolVar(&responsibleAI, "responsible-ai", true, "Include responsible AI agent")
	createCmd.Flags().BoolVar(&errorHandler, "error-handler", true, "Include error handling agents")
	createCmd.Flags().BoolVarP(&interactive, "interactive", "i", false, "Interactive mode for guided setup")

	// Multi-agent orchestration flags
	createCmd.Flags().StringVar(&orchestrationMode, "orchestration-mode", "sequential", "Orchestration mode (sequential, route, collaborative, loop, mixed)")
	createCmd.Flags().StringVar(&collaborativeAgents, "collaborative-agents", "", "Comma-separated list of agent names for parallel execution")
	createCmd.Flags().StringVar(&sequentialAgents, "sequential-agents", "", "Comma-separated list of agent names for sequential pipeline")
	createCmd.Flags().StringVar(&loopAgent, "loop-agent", "", "Single agent name for loop-based execution pattern")
	createCmd.Flags().IntVar(&maxIterations, "max-iterations", 5, "Maximum iterations for loop orchestration")
	createCmd.Flags().IntVar(&orchestrationTimeout, "orchestration-timeout", 30, "Timeout for orchestration operations (seconds)")
	createCmd.Flags().Float64Var(&failureThreshold, "failure-threshold", 0.5, "Failure threshold for stopping orchestration (0.0-1.0)")
	createCmd.Flags().IntVar(&maxConcurrency, "max-concurrency", 10, "Maximum concurrent agent executions")

	// Visualization flags
	createCmd.Flags().BoolVar(&visualize, "visualize", false, "Generate Mermaid workflow diagrams")
	createCmd.Flags().StringVar(&visualizeOutputDir, "visualize-output", "docs/workflows", "Output directory for generated diagrams")

	// MCP integration flags
	createCmd.Flags().BoolVar(&mcpEnabled, "mcp-enabled", false, "Enable MCP tool integration")
	createCmd.Flags().BoolVar(&mcpProduction, "mcp-production", false, "Include production MCP features (pooling, retry, metrics)")
	createCmd.Flags().BoolVar(&withCache, "with-cache", false, "Enable MCP result caching")
	createCmd.Flags().BoolVar(&withMetrics, "with-metrics", false, "Enable Prometheus metrics")
	createCmd.Flags().StringVar(&mcpTools, "mcp-tools", "web_search,summarize", "Comma-separated list of MCP tools")
	createCmd.Flags().StringVar(&mcpServers, "mcp-servers", "docker", "Comma-separated list of MCP server names")
	createCmd.Flags().StringVar(&cacheBackend, "cache-backend", "memory", "Cache backend (memory, redis)")
	createCmd.Flags().IntVar(&metricsPort, "metrics-port", 8080, "Metrics server port")
	createCmd.Flags().BoolVar(&withLoadBalancer, "with-load-balancer", false, "Enable MCP load balancing")
	createCmd.Flags().IntVar(&connectionPoolSize, "connection-pool-size", 5, "MCP connection pool size")
	createCmd.Flags().StringVar(&retryPolicy, "retry-policy", "exponential", "Retry policy (exponential, linear, fixed)")

	// Memory/RAG flags
	createCmd.Flags().BoolVar(&memoryEnabled, "memory-enabled", false, "Enable memory system for agents with persistent storage and retrieval")
	createCmd.Flags().StringVar(&memoryProvider, "memory-provider", "memory", "Memory provider: 'memory' (in-memory), 'pgvector' (PostgreSQL), 'weaviate' (vector DB)")
	createCmd.Flags().StringVar(&embeddingProvider, "embedding-provider", "ollama", "Embedding provider: 'openai' (requires API key), 'ollama' (local), 'dummy' (testing)")
	createCmd.Flags().StringVar(&embeddingModel, "embedding-model", "nomic-embed-text:latest", "Embedding model name (auto-selected based on provider if empty)")
	createCmd.Flags().BoolVar(&ragEnabled, "rag-enabled", false, "Enable RAG (Retrieval-Augmented Generation) for knowledge-aware responses")
	createCmd.Flags().IntVar(&ragChunkSize, "rag-chunk-size", 1000, "RAG document chunk size in tokens (recommended: 500-2000)")
	createCmd.Flags().IntVar(&ragOverlap, "rag-overlap", 100, "RAG chunk overlap size in tokens (recommended: 10-20% of chunk size)")
	createCmd.Flags().IntVar(&ragTopK, "rag-top-k", 5, "RAG top-k results to retrieve for context (recommended: 3-10)")
	createCmd.Flags().Float64Var(&ragScoreThreshold, "rag-score-threshold", 0.7, "RAG minimum similarity score threshold (0.0-1.0, recommended: 0.6-0.8)")
	createCmd.Flags().BoolVar(&hybridSearch, "hybrid-search", false, "Enable hybrid search combining semantic similarity and keyword matching")
	createCmd.Flags().BoolVar(&sessionMemory, "session-memory", false, "Enable session-based memory isolation for multi-user scenarios")

	// Mark MCP production dependencies
	createCmd.MarkFlagsMutuallyExclusive("mcp-production", "mcp-enabled")
}

func runCreateCommand(cmd *cobra.Command, args []string) error {
	var projectName string

	if interactive {
		config, err := interactiveSetup()
		if err != nil {
			return fmt.Errorf("interactive setup failed: %w", err)
		}
		return scaffold.CreateAgentProject(config)
	}

	// Non-interactive mode
	projectName = args[0]

	// Validate MCP flag combinations
	if err := validateMCPFlags(); err != nil {
		return err
	}

	// Validate orchestration configuration
	if err := validateOrchestrationFlags(); err != nil {
		return err
	}

	// Validate memory configuration with embedding intelligence
	if err := validateMemoryFlagsWithIntelligence(); err != nil {
		return err
	}

	// Parse tool and server lists
	toolList := parseCommaSeparatedList(mcpTools)
	serverList := parseCommaSeparatedList(mcpServers)

	// Use embedding intelligence to calculate dimensions and validate configuration
	var embeddingDimensions int
	if memoryEnabled {
		dimensions, err := scaffold.EmbeddingIntel.GetDimensionsForModel(embeddingProvider, embeddingModel)
		if err != nil {
			// Use fallback logic for unknown models
			dimensions = scaffold.GetModelDimensions(embeddingProvider, embeddingModel)
			fmt.Printf("⚠️  Unknown embedding model %s/%s - using default dimensions: %d\n", embeddingProvider, embeddingModel, dimensions)
		}
		embeddingDimensions = dimensions
		
		// Show embedding model information
		if modelInfo, err := scaffold.EmbeddingIntel.GetModelInfo(embeddingProvider, embeddingModel); err == nil {
			fmt.Printf("✓ Using embedding model: %s (%d dimensions)\n", modelInfo.Model, modelInfo.Dimensions)
			if modelInfo.Notes != "" {
				fmt.Printf("  %s\n", modelInfo.Notes)
			}
		}
	}

	// Create project configuration
	config := scaffold.ProjectConfig{
		Name:          projectName,
		NumAgents:     numAgents,
		Provider:      provider,
		ResponsibleAI: responsibleAI,
		ErrorHandler:  errorHandler,

		// Orchestration configuration
		OrchestrationMode:    orchestrationMode,
		CollaborativeAgents:  parseCommaSeparatedList(collaborativeAgents),
		SequentialAgents:     parseCommaSeparatedList(sequentialAgents),
		LoopAgent:            loopAgent,
		MaxIterations:        maxIterations,
		OrchestrationTimeout: orchestrationTimeout,
		FailureThreshold:     failureThreshold,
		MaxConcurrency:       maxConcurrency,

		// Visualization configuration
		Visualize:          visualize,
		VisualizeOutputDir: visualizeOutputDir,

		// MCP configuration
		MCPEnabled:         mcpEnabled || mcpProduction,
		MCPProduction:      mcpProduction,
		WithCache:          withCache,
		WithMetrics:        withMetrics,
		MCPTools:           toolList,
		MCPServers:         serverList,
		CacheBackend:       cacheBackend,
		MetricsPort:        metricsPort,
		WithLoadBalancer:   withLoadBalancer,
		ConnectionPoolSize: connectionPoolSize,
		RetryPolicy:        retryPolicy,

		// Memory/RAG configuration with intelligent defaults
		MemoryEnabled:       memoryEnabled,
		MemoryProvider:      memoryProvider,
		EmbeddingProvider:   embeddingProvider,
		EmbeddingModel:      embeddingModel,
		EmbeddingDimensions: embeddingDimensions,
		RAGEnabled:          ragEnabled,
		RAGChunkSize:        ragChunkSize,
		RAGOverlap:          ragOverlap,
		RAGTopK:             ragTopK,
		RAGScoreThreshold:   ragScoreThreshold,
		HybridSearch:        hybridSearch,
		SessionMemory:       sessionMemory,
	}

	// Create the project
	fmt.Printf("Creating AgentFlow project '%s'...\n", projectName)
	if config.MCPEnabled {
		fmt.Printf("✓ MCP integration enabled\n")
		if config.MCPProduction {
			fmt.Printf("✓ Production MCP features enabled\n")
		}
		if config.WithCache {
			fmt.Printf("✓ MCP caching enabled (%s)\n", config.CacheBackend)
		}
		if config.WithMetrics {
			fmt.Printf("✓ Metrics enabled on port %d\n", config.MetricsPort)
		}
	}

	if config.Visualize {
		fmt.Printf("✓ Workflow visualization enabled\n")
		fmt.Printf("✓ Diagrams will be generated in %s/\n", config.VisualizeOutputDir)
	}

	if config.MemoryEnabled {
		fmt.Printf("✓ Memory system enabled (%s)\n", config.MemoryProvider)
		if config.RAGEnabled {
			fmt.Printf("✓ RAG enabled (chunk size: %d, overlap: %d, top-k: %d)\n",
				config.RAGChunkSize, config.RAGOverlap, config.RAGTopK)
		}
		if config.HybridSearch {
			fmt.Printf("✓ Hybrid search enabled\n")
		}
		if config.SessionMemory {
			fmt.Printf("✓ Session memory enabled\n")
		}
		fmt.Printf("✓ Embedding provider: %s\n", config.EmbeddingProvider)
	}

	// Use modular template system for better maintainability
	return scaffold.CreateAgentProjectModular(config)
}

func validateMCPFlags() error {
	// MCP production implies MCP enabled
	if mcpProduction {
		mcpEnabled = true
	}

	// Cache and metrics require MCP
	if (withCache || withMetrics) && !mcpEnabled && !mcpProduction {
		return fmt.Errorf("--with-cache and --with-metrics require --mcp-enabled or --mcp-production")
	}

	// Load balancer requires production
	if withLoadBalancer && !mcpProduction {
		return fmt.Errorf("--with-load-balancer requires --mcp-production")
	}

	// Validate provider
	validProviders := []string{"openai", "azure", "ollama", "mock"}
	if !contains(validProviders, provider) {
		return fmt.Errorf("invalid provider: %s. Valid options: %s", provider, strings.Join(validProviders, ", "))
	}

	// Validate cache backend
	validBackends := []string{"memory", "redis"}
	if !contains(validBackends, cacheBackend) {
		return fmt.Errorf("invalid cache backend: %s. Valid options: %s", cacheBackend, strings.Join(validBackends, ", "))
	}

	// Validate retry policy
	validPolicies := []string{"exponential", "linear", "fixed"}
	if !contains(validPolicies, retryPolicy) {
		return fmt.Errorf("invalid retry policy: %s. Valid options: %s", retryPolicy, strings.Join(validPolicies, ", "))
	}

	return nil
}

func validateOrchestrationFlags() error {
	// Validate orchestration mode
	validModes := []string{"sequential", "route", "collaborative", "loop", "mixed"}
	if !contains(validModes, orchestrationMode) {
		return fmt.Errorf("invalid orchestration mode: %s. Valid options: %s", orchestrationMode, strings.Join(validModes, ", "))
	}

	// Collaborative mode validations
	if orchestrationMode == "collaborative" {
		if numAgents < 2 {
			return fmt.Errorf("collaborative orchestration requires at least 2 agents")
		}
		if collaborativeAgents != "" {
			agents := parseCommaSeparatedList(collaborativeAgents)
			if len(agents) < 2 {
				return fmt.Errorf("collaborative orchestration requires at least 2 agent names")
			}
		}
	}

	// Sequential mode validations
	if orchestrationMode == "sequential" {
		if numAgents < 2 {
			return fmt.Errorf("sequential orchestration requires at least 2 agents")
		}
		if sequentialAgents != "" {
			agents := parseCommaSeparatedList(sequentialAgents)
			if len(agents) < 2 {
				return fmt.Errorf("sequential orchestration requires at least 2 agent names in sequence")
			}
		}
	}

	// Loop mode validations
	if orchestrationMode == "loop" {
		if numAgents != 1 {
			return fmt.Errorf("loop orchestration requires exactly 1 agent")
		}
		if loopAgent == "" {
			return fmt.Errorf("loop orchestration requires --loop-agent to be specified")
		}
	}

	// Mixed mode validations
	if orchestrationMode == "mixed" {
		if collaborativeAgents == "" && sequentialAgents == "" {
			return fmt.Errorf("mixed orchestration requires at least one of --collaborative-agents or --sequential-agents")
		}
	}

	// Cross-mode validations - ensure only relevant flags are used
	if orchestrationMode != "collaborative" && orchestrationMode != "mixed" && collaborativeAgents != "" {
		return fmt.Errorf("--collaborative-agents can only be used with --orchestration-mode collaborative or mixed")
	}
	if orchestrationMode != "sequential" && orchestrationMode != "mixed" && sequentialAgents != "" {
		return fmt.Errorf("--sequential-agents can only be used with --orchestration-mode sequential or mixed")
	}
	if orchestrationMode != "loop" && loopAgent != "" {
		return fmt.Errorf("--loop-agent can only be used with --orchestration-mode loop")
	}

	// Max iterations validation (primarily for loop mode)
	if maxIterations <= 0 {
		return fmt.Errorf("max iterations must be a positive integer")
	}
	if orchestrationMode == "loop" && maxIterations > 100 {
		return fmt.Errorf("max iterations for loop mode should not exceed 100 to prevent infinite loops")
	}

	// Timeout must be positive
	if orchestrationTimeout <= 0 {
		return fmt.Errorf("orchestration timeout must be a positive integer")
	}

	// Failure threshold must be between 0 and 1
	if failureThreshold < 0.0 || failureThreshold > 1.0 {
		return fmt.Errorf("failure threshold must be between 0.0 and 1.0")
	}

	// Max concurrency must be positive
	if maxConcurrency <= 0 {
		return fmt.Errorf("max concurrency must be a positive integer")
	}
	if maxConcurrency > 100 {
		return fmt.Errorf("max concurrency should not exceed 100 for performance reasons")
	}

	return nil
}

func validateMemoryFlags() error {
	// Auto-enable memory when RAG is enabled with warning message
	if ragEnabled && !memoryEnabled {
		fmt.Println("⚠️  RAG requires memory - automatically enabling memory")
		memoryEnabled = true
	}

	// Auto-enable memory when session memory is enabled
	if sessionMemory && !memoryEnabled {
		fmt.Println("⚠️  Session memory requires memory system - automatically enabling memory")
		memoryEnabled = true
	}

	// Auto-enable memory when hybrid search is enabled
	if hybridSearch && !memoryEnabled {
		fmt.Println("⚠️  Hybrid search requires memory system - automatically enabling memory")
		memoryEnabled = true
	}

	// Validate memory provider options
	if memoryEnabled {
		validProviders := []string{"memory", "pgvector", "weaviate"}
		if !contains(validProviders, memoryProvider) {
			return fmt.Errorf("invalid memory provider: %s. Valid options: %s", memoryProvider, strings.Join(validProviders, ", "))
		}
	}

	// Validate embedding provider options
	if memoryEnabled {
		validEmbeddingProviders := []string{"openai", "ollama", "dummy"}
		if !contains(validEmbeddingProviders, embeddingProvider) {
			return fmt.Errorf("invalid embedding provider: %s. Valid options: %s", embeddingProvider, strings.Join(validEmbeddingProviders, ", "))
		}
	}

	// Warn about Docker requirements for database providers
	if memoryEnabled && (memoryProvider == "pgvector" || memoryProvider == "weaviate") {
		fmt.Printf("ℹ️  Database provider '%s' selected - Docker Compose file will be generated\n", memoryProvider)
		fmt.Println("   Run 'docker-compose up -d' (or 'docker compose up -d') to start the database")
		fmt.Println("   Setup scripts (setup.sh/setup.bat) will be created for easy initialization")
	}

	// Warn about API key requirements for OpenAI embeddings
	if memoryEnabled && embeddingProvider == "openai" {
		fmt.Println("ℹ️  OpenAI embeddings selected - set OPENAI_API_KEY environment variable")
		fmt.Println("   Example: export OPENAI_API_KEY=\"your-api-key-here\"")
	}

	// Warn about Ollama requirements
	if memoryEnabled && embeddingProvider == "ollama" {
		fmt.Printf("ℹ️  Ollama embeddings selected - ensure Ollama is running on http://localhost:11434\n")
		fmt.Printf("   Make sure the embedding model is installed: ollama pull %s\n", embeddingModel)
	}

	// Validate RAG parameters with helpful error messages
	if ragEnabled {
		if ragChunkSize <= 0 {
			return fmt.Errorf("RAG chunk size must be a positive integer (current: %d). Recommended: 500-2000", ragChunkSize)
		}
		if ragOverlap < 0 || ragOverlap >= ragChunkSize {
			return fmt.Errorf("RAG overlap must be non-negative and less than chunk size (current: %d, chunk size: %d). Recommended: 10-20%% of chunk size", ragOverlap, ragChunkSize)
		}
		if ragTopK <= 0 {
			return fmt.Errorf("RAG top-k must be a positive integer (current: %d). Recommended: 3-10", ragTopK)
		}
		if ragScoreThreshold < 0.0 || ragScoreThreshold > 1.0 {
			return fmt.Errorf("RAG score threshold must be between 0.0 and 1.0 (current: %.2f). Recommended: 0.6-0.8", ragScoreThreshold)
		}
	}

	// Validate compatibility between memory and embedding providers
	if memoryEnabled {
		// Check for potential compatibility issues
		if memoryProvider == "weaviate" && embeddingProvider == "dummy" {
			fmt.Println("⚠️  Using dummy embeddings with Weaviate may not provide meaningful search results")
			fmt.Println("   Consider using 'openai' or 'ollama' embedding provider for better performance")
		}
		
		if memoryProvider == "memory" && ragEnabled {
			fmt.Println("ℹ️  Using in-memory provider with RAG - data will not persist between restarts")
			fmt.Println("   Consider using 'pgvector' or 'weaviate' for persistent storage")
		}
	}

	return nil
}

// validateMemoryFlagsWithIntelligence validates memory configuration using embedding intelligence
func validateMemoryFlagsWithIntelligence() error {
	// First run the basic validation
	if err := validateMemoryFlags(); err != nil {
		return err
	}

	// Skip intelligence validation if memory is not enabled
	if !memoryEnabled {
		return nil
	}

	// Use embedding intelligence to suggest better models if current one is unknown
	if _, err := scaffold.EmbeddingIntel.GetModelInfo(embeddingProvider, embeddingModel); err != nil {
		fmt.Printf("⚠️  Unknown embedding model: %s/%s\n", embeddingProvider, embeddingModel)
		
		// Suggest recommended models for the provider
		suggestions := scaffold.GetEmbeddingModelSuggestions(embeddingProvider)
		if len(suggestions) > 0 {
			fmt.Println("💡 Recommended models for this provider:")
			for _, suggestion := range suggestions {
				fmt.Printf("   • %s\n", suggestion)
			}
		}
		
		// Continue with fallback dimensions but warn user
		fmt.Printf("   Continuing with default dimensions, but consider using a recommended model\n")
	}

	// Validate compatibility between embedding model and memory provider
	if err := scaffold.EmbeddingIntel.ValidateCompatibility(embeddingProvider, embeddingModel, memoryProvider); err != nil {
		fmt.Printf("⚠️  Compatibility warning: %v\n", err)
		// Don't fail, just warn - let user decide
	}

	// Show additional warnings from the validation system
	warnings := scaffold.ValidateEmbeddingConfig(embeddingProvider, embeddingModel, memoryProvider)
	for _, warning := range warnings {
		fmt.Printf("ℹ️  %s\n", warning)
	}

	// Validate that embedding model dimensions are reasonable for the memory provider
	dimensions := scaffold.GetModelDimensions(embeddingProvider, embeddingModel)
	if dimensions > 3072 {
		fmt.Printf("⚠️  Large embedding dimensions (%d) may impact performance\n", dimensions)
		fmt.Println("   Consider using a smaller model for better performance")
	}

	// Specific validation for pgvector
	if memoryProvider == "pgvector" && dimensions > 2000 {
		fmt.Printf("ℹ️  PgVector with %d dimensions - ensure your PostgreSQL instance has sufficient resources\n", dimensions)
	}

	return nil
}

func parseCommaSeparatedList(input string) []string {
	if input == "" {
		return []string{}
	}
	items := strings.Split(input, ",")
	result := make([]string, 0, len(items))
	for _, item := range items {
		if trimmed := strings.TrimSpace(item); trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func interactiveSetup() (scaffold.ProjectConfig, error) {
	config := scaffold.ProjectConfig{}

	fmt.Println("🚀 AgentFlow Project Setup")
	fmt.Println("==========================")

	// Project name
	fmt.Print("Project name: ")
	fmt.Scanln(&config.Name)

	// Basic configuration
	fmt.Print("Number of agents (default 2): ")
	var agentsInput string
	fmt.Scanln(&agentsInput)
	if agentsInput != "" {
		if parsed, err := strconv.Atoi(agentsInput); err == nil {
			config.NumAgents = parsed
		} else {
			config.NumAgents = 2
		}
	} else {
		config.NumAgents = 2
	}

	// Provider selection
	fmt.Println("Select LLM provider:")
	fmt.Println("1. OpenAI (default)")
	fmt.Println("2. Azure OpenAI")
	fmt.Println("3. Ollama (local)")
	fmt.Println("4. Mock (testing)")
	fmt.Print("Choice (1-4): ")
	var providerChoice string
	fmt.Scanln(&providerChoice)

	providers := map[string]string{
		"1": "openai", "2": "azure", "3": "ollama", "4": "mock",
		"": "openai", // default
	}
	if p, exists := providers[providerChoice]; exists {
		config.Provider = p
	} else {
		config.Provider = "openai"
	}

	// MCP integration
	fmt.Print("Enable MCP integration? (y/N): ")
	var mcpChoice string
	fmt.Scanln(&mcpChoice)
	config.MCPEnabled = strings.ToLower(mcpChoice) == "y" || strings.ToLower(mcpChoice) == "yes"

	if config.MCPEnabled {
		// MCP feature selection
		fmt.Print("Enable production MCP features? (y/N): ")
		var prodChoice string
		fmt.Scanln(&prodChoice)
		config.MCPProduction = strings.ToLower(prodChoice) == "y" || strings.ToLower(prodChoice) == "yes"

		fmt.Print("Enable MCP caching? (y/N): ")
		var cacheChoice string
		fmt.Scanln(&cacheChoice)
		config.WithCache = strings.ToLower(cacheChoice) == "y" || strings.ToLower(cacheChoice) == "yes"

		fmt.Print("Enable metrics? (y/N): ")
		var metricsChoice string
		fmt.Scanln(&metricsChoice)
		config.WithMetrics = strings.ToLower(metricsChoice) == "y" || strings.ToLower(metricsChoice) == "yes"

		// MCP tools
		fmt.Print("MCP tools (comma-separated, default: web_search,summarize): ")
		var toolsInput string
		fmt.Scanln(&toolsInput)
		if toolsInput != "" {
			config.MCPTools = parseCommaSeparatedList(toolsInput)
		} else {
			config.MCPTools = []string{"web_search", "summarize"}
		}

		// MCP servers
		fmt.Print("MCP servers (comma-separated, default: docker): ")
		var serversInput string
		fmt.Scanln(&serversInput)
		if serversInput != "" {
			config.MCPServers = parseCommaSeparatedList(serversInput)
		} else {
			config.MCPServers = []string{"docker"}
		}

		// Set defaults
		config.CacheBackend = "memory"
		config.MetricsPort = 8080
		config.ConnectionPoolSize = 5
		config.RetryPolicy = "exponential"
	}

	// Responsible AI and error handling (defaults to true)
	config.ResponsibleAI = true
	config.ErrorHandler = true

	// Orchestration mode selection
	fmt.Println("\nSelect orchestration mode:")
	fmt.Println("1. Route (default) - Events routed to specific agents")
	fmt.Println("2. Collaborative - All agents process events in parallel")
	fmt.Println("3. Sequential - Agents process events in pipeline order")
	fmt.Println("4. Loop - Single agent processes with iterations")
	fmt.Println("5. Mixed - Combination of collaborative and sequential")
	fmt.Print("Choice (1-5): ")
	var orchestrationChoice string
	fmt.Scanln(&orchestrationChoice)

	orchestrationModes := map[string]string{
		"1": "route", "2": "collaborative", "3": "sequential", "4": "loop", "5": "mixed",
		"": "route", // default
	}
	if mode, exists := orchestrationModes[orchestrationChoice]; exists {
		config.OrchestrationMode = mode
	} else {
		config.OrchestrationMode = "route"
	}

	// Set orchestration defaults
	config.MaxIterations = 5
	config.OrchestrationTimeout = 30
	config.FailureThreshold = 0.5
	config.MaxConcurrency = 10

	// Mode-specific configuration
	switch config.OrchestrationMode {
	case "collaborative":
		fmt.Print("Collaborative agents (comma-separated names, leave empty for auto-generated): ")
		var collabInput string
		fmt.Scanln(&collabInput)
		if collabInput != "" {
			config.CollaborativeAgents = parseCommaSeparatedList(collabInput)
		}
	case "sequential":
		fmt.Print("Sequential agents (comma-separated names in order, leave empty for auto-generated): ")
		var seqInput string
		fmt.Scanln(&seqInput)
		if seqInput != "" {
			config.SequentialAgents = parseCommaSeparatedList(seqInput)
		}
	case "loop":
		fmt.Print("Loop agent name (leave empty for auto-generated): ")
		var loopInput string
		fmt.Scanln(&loopInput)
		if loopInput != "" {
			config.LoopAgent = loopInput
		}
		fmt.Print("Max iterations (default 5): ")
		var iterInput string
		fmt.Scanln(&iterInput)
		if iterInput != "" {
			if parsed, err := strconv.Atoi(iterInput); err == nil && parsed > 0 {
				config.MaxIterations = parsed
			}
		}
	case "mixed":
		fmt.Print("Collaborative agents (comma-separated, leave empty to skip): ")
		var collabInput string
		fmt.Scanln(&collabInput)
		if collabInput != "" {
			config.CollaborativeAgents = parseCommaSeparatedList(collabInput)
		}
		fmt.Print("Sequential agents (comma-separated, leave empty to skip): ")
		var seqInput string
		fmt.Scanln(&seqInput)
		if seqInput != "" {
			config.SequentialAgents = parseCommaSeparatedList(seqInput)
		}
	}

	// Visualization options
	fmt.Print("\nGenerate workflow diagrams? (y/N): ")
	var visualizeChoice string
	fmt.Scanln(&visualizeChoice)
	config.Visualize = strings.ToLower(visualizeChoice) == "y" || strings.ToLower(visualizeChoice) == "yes"

	if config.Visualize {
		fmt.Print("Diagram output directory (default: docs/workflows): ")
		var outputDir string
		fmt.Scanln(&outputDir)
		if outputDir != "" {
			config.VisualizeOutputDir = outputDir
		} else {
			config.VisualizeOutputDir = "docs/workflows"
		}
	}

	// Memory and RAG options
	fmt.Print("\nEnable memory system? (y/N): ")
	var memoryChoice string
	fmt.Scanln(&memoryChoice)
	config.MemoryEnabled = strings.ToLower(memoryChoice) == "y" || strings.ToLower(memoryChoice) == "yes"

	if config.MemoryEnabled {
		// Memory provider selection
		fmt.Println("Select memory provider:")
		fmt.Println("1. In-Memory (default) - Fast, temporary storage")
		fmt.Println("2. PgVector - PostgreSQL with vector search")
		fmt.Println("3. Weaviate - Dedicated vector database")
		fmt.Print("Choice (1-3): ")
		var memoryProviderChoice string
		fmt.Scanln(&memoryProviderChoice)

		memoryProviders := map[string]string{
			"1": "memory", "2": "pgvector", "3": "weaviate",
			"": "memory", // default
		}
		if p, exists := memoryProviders[memoryProviderChoice]; exists {
			config.MemoryProvider = p
		} else {
			config.MemoryProvider = "memory"
		}

		// Embedding provider selection
		fmt.Println("Select embedding provider:")
		fmt.Println("1. Ollama (recommended) - Local embeddings with nomic-embed-text")
		fmt.Println("2. OpenAI - Production-ready embeddings (requires API key)")
		fmt.Println("3. Dummy - For testing/development only")
		fmt.Print("Choice (1-3): ")
		var embeddingChoice string
		fmt.Scanln(&embeddingChoice)

		embeddingProviders := map[string]string{
			"1": "ollama", "2": "openai", "3": "dummy",
			"": "ollama", // default changed to ollama
		}
		if p, exists := embeddingProviders[embeddingChoice]; exists {
			config.EmbeddingProvider = p
		} else {
			config.EmbeddingProvider = "ollama"
		}

		// Embedding model with intelligent suggestions
		if config.EmbeddingProvider == "openai" {
			fmt.Println("Available OpenAI models:")
			suggestions := scaffold.GetEmbeddingModelSuggestions("openai")
			for i, suggestion := range suggestions {
				fmt.Printf("  %d. %s\n", i+1, suggestion)
			}
			fmt.Print("OpenAI embedding model (default: text-embedding-3-small): ")
			var modelInput string
			fmt.Scanln(&modelInput)
			if modelInput != "" {
				config.EmbeddingModel = modelInput
			} else {
				config.EmbeddingModel = "text-embedding-3-small"
			}
		} else if config.EmbeddingProvider == "ollama" {
			fmt.Println("Available Ollama models:")
			suggestions := scaffold.GetEmbeddingModelSuggestions("ollama")
			for i, suggestion := range suggestions {
				fmt.Printf("  %d. %s\n", i+1, suggestion)
			}
			fmt.Print("Ollama embedding model (default: nomic-embed-text:latest): ")
			var modelInput string
			fmt.Scanln(&modelInput)
			if modelInput != "" {
				config.EmbeddingModel = modelInput
			} else {
				config.EmbeddingModel = "nomic-embed-text:latest"
			}
		} else {
			config.EmbeddingModel = "dummy"
		}

		// Calculate and show embedding dimensions
		dimensions := scaffold.GetModelDimensions(config.EmbeddingProvider, config.EmbeddingModel)
		config.EmbeddingDimensions = dimensions
		fmt.Printf("✓ Using %d-dimensional embeddings\n", dimensions)

		// RAG options
		fmt.Print("Enable RAG (Retrieval-Augmented Generation)? (y/N): ")
		var ragChoice string
		fmt.Scanln(&ragChoice)
		config.RAGEnabled = strings.ToLower(ragChoice) == "y" || strings.ToLower(ragChoice) == "yes"

		if config.RAGEnabled {
			// RAG chunk size
			fmt.Print("RAG chunk size (default: 1000): ")
			var chunkInput string
			fmt.Scanln(&chunkInput)
			if chunkInput != "" {
				if parsed, err := strconv.Atoi(chunkInput); err == nil && parsed > 0 {
					config.RAGChunkSize = parsed
				} else {
					config.RAGChunkSize = 1000
				}
			} else {
				config.RAGChunkSize = 1000
			}

			// RAG overlap
			fmt.Print("RAG chunk overlap (default: 100): ")
			var overlapInput string
			fmt.Scanln(&overlapInput)
			if overlapInput != "" {
				if parsed, err := strconv.Atoi(overlapInput); err == nil && parsed >= 0 {
					config.RAGOverlap = parsed
				} else {
					config.RAGOverlap = 100
				}
			} else {
				config.RAGOverlap = 100
			}

			// RAG top-k
			fmt.Print("RAG top-k results (default: 5): ")
			var topkInput string
			fmt.Scanln(&topkInput)
			if topkInput != "" {
				if parsed, err := strconv.Atoi(topkInput); err == nil && parsed > 0 {
					config.RAGTopK = parsed
				} else {
					config.RAGTopK = 5
				}
			} else {
				config.RAGTopK = 5
			}

			// RAG score threshold
			fmt.Print("RAG score threshold (default: 0.7): ")
			var thresholdInput string
			fmt.Scanln(&thresholdInput)
			if thresholdInput != "" {
				if parsed, err := strconv.ParseFloat(thresholdInput, 64); err == nil && parsed >= 0.0 && parsed <= 1.0 {
					config.RAGScoreThreshold = parsed
				} else {
					config.RAGScoreThreshold = 0.7
				}
			} else {
				config.RAGScoreThreshold = 0.7
			}
		}

		// Additional memory features
		fmt.Print("Enable hybrid search (semantic + keyword)? (y/N): ")
		var hybridChoice string
		fmt.Scanln(&hybridChoice)
		config.HybridSearch = strings.ToLower(hybridChoice) == "y" || strings.ToLower(hybridChoice) == "yes"

		fmt.Print("Enable session-based memory? (y/N): ")
		var sessionChoice string
		fmt.Scanln(&sessionChoice)
		config.SessionMemory = strings.ToLower(sessionChoice) == "y" || strings.ToLower(sessionChoice) == "yes"
	}

	fmt.Println("\n✓ Configuration complete!")
	return config, nil
}
