package templates

const MainTemplate = `package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	{{if or (eq .Config.OrchestrationMode "collaborative") .Config.MemoryEnabled}}
	"strings"
	{{end}}
	"sync"
	"time"

	"github.com/kunalkushwaha/agentflow/core"
)

{{if .Config.MemoryEnabled}}
// Global memory instance for access by agents
var memory core.Memory
{{end}}

func main() {
	ctx := context.Background()
	core.SetLogLevel(core.INFO)
	logger := core.Logger()
	logger.Info().Msg("Starting {{.Config.Name}} multi-agent system...")

	messageFlag := flag.String("m", "", "Message to process")
	flag.Parse()

	// Read provider from config
	config, err := core.LoadConfig("agentflow.toml")
	if err != nil {
		fmt.Printf("Failed to load configuration: %v\n", err)
		os.Exit(1)
	}

	llmProvider, err := initializeProvider(config.AgentFlow.Provider)
	if err != nil {
		fmt.Printf("Failed to initialize LLM provider '%s': %v\n", config.AgentFlow.Provider, err)
		fmt.Printf("Make sure you have set the appropriate environment variables:\n")
		switch config.AgentFlow.Provider {
		case "azure":
			fmt.Printf("  AZURE_OPENAI_API_KEY=your-api-key\n")
			fmt.Printf("  AZURE_OPENAI_ENDPOINT=https://your-resource.openai.azure.com/\n")
			fmt.Printf("  AZURE_OPENAI_DEPLOYMENT=your-deployment-name\n")
		case "openai":
			fmt.Printf("  OPENAI_API_KEY=your-api-key\n")
		case "ollama":
			fmt.Printf("  Ollama should be running on localhost:11434\n")
		}
		os.Exit(1)
	}

	{{if .Config.MCPEnabled}}
	// Initialize MCP manager for tool integration with timeout handling
	logger.Info().Msg("Initializing MCP with timeout handling...")
	mcpInitCtx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	var mcpManager core.MCPManager
	mcpDone := make(chan bool, 1)
	var mcpErr error

	go func() {
		mcpManager, mcpErr = initializeMCP()
		mcpDone <- true
	}()

	select {
	case <-mcpDone:
		if mcpErr != nil {
			logger.Warn().Err(mcpErr).Msg("MCP initialization failed, continuing without MCP")
			mcpManager = nil
		}
	case <-mcpInitCtx.Done():
		logger.Warn().Msg("MCP initialization timed out, continuing without MCP")
		mcpManager = nil
		mcpErr = fmt.Errorf("MCP initialization timeout")
	}

	if mcpManager != nil {
		logger.Info().Msg("MCP manager initialized successfully - agents can access tools via core.GetMCPManager()")

		// Initialize MCP tool registry with timeout
		registryCtx, registryCancel := context.WithTimeout(ctx, 10*time.Second)
		defer registryCancel()

		registryDone := make(chan error, 1)
		go func() {
			registryDone <- core.InitializeMCPToolRegistry()
		}()

		select {
		case err := <-registryDone:
			if err != nil {
				logger.Warn().Err(err).Msg("Failed to initialize MCP tool registry")
			} else {
				logger.Info().Msg("MCP tool registry initialized successfully")
			}
		case <-registryCtx.Done():
			logger.Warn().Msg("MCP tool registry initialization timed out")
		}

		// Register MCP tools with the registry with timeout
		toolsCtx, toolsCancel := context.WithTimeout(ctx, 10*time.Second)
		defer toolsCancel()

		toolsDone := make(chan error, 1)
		go func() {
			toolsDone <- core.RegisterMCPToolsWithRegistry(toolsCtx)
		}()

		select {
		case err := <-toolsDone:
			if err != nil {
				logger.Warn().Err(err).Msg("Failed to register MCP tools with registry")
			} else {
				logger.Info().Msg("MCP tools registered with registry successfully")
			}
		case <-toolsCtx.Done():
			logger.Warn().Msg("MCP tools registration timed out")
		}
	}
	{{end}}

	{{if .Config.MemoryEnabled}}
	// Initialize memory system
	fmt.Println("🧠 Initializing memory system...")
	
	// Create memory configuration from agentflow.toml settings
	memoryConfig := config.AgentMemory
	
	// Validate configuration before initializing memory
	fmt.Println("🔍 Validating memory configuration...")
	if err := validateMemoryConfig(memoryConfig, "{{.Config.EmbeddingModel}}"); err != nil {
		logger.Error().Err(err).Msg("Memory configuration validation failed")
		fmt.Printf("❌ Configuration Error: %v\n", err)
		os.Exit(1)
	}
	
	logger.Info().Msg("Memory configuration validation passed")
	fmt.Println("✅ Configuration validated!")
	
	memory, err := core.NewMemory(memoryConfig)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to initialize memory")
		fmt.Printf("Memory initialization failed: %v\n", err)
		
		// Provide specific troubleshooting based on provider
		switch memoryConfig.Provider {
		case "pgvector":
			fmt.Printf("\n💡 PostgreSQL/PgVector Troubleshooting:\n")
			fmt.Printf("   1. Start database: docker compose up -d\n")
			fmt.Printf("   2. Run setup script: ./setup.sh (or setup.bat on Windows)\n")
			fmt.Printf("   3. Check connection string in agentflow.toml\n")
			fmt.Printf("   4. Verify database exists: psql -h localhost -U user -d agentflow\n")
		case "weaviate":
			fmt.Printf("\n💡 Weaviate Troubleshooting:\n")
			fmt.Printf("   1. Start Weaviate: docker compose up -d\n")
			fmt.Printf("   2. Check Weaviate is running: curl http://localhost:8080/v1/meta\n")
			fmt.Printf("   3. Verify connection string in agentflow.toml\n")
		case "memory":
			fmt.Printf("\n💡 In-Memory Provider Issue:\n")
			fmt.Printf("   This shouldn't fail - check your configuration\n")
		}
		
		// Check embedding provider availability
		if memoryConfig.Embedding.Provider == "ollama" {
			fmt.Printf("\n💡 Ollama Troubleshooting:\n")
			fmt.Printf("   1. Start Ollama: ollama serve\n")
			fmt.Printf("   2. Pull model: ollama pull %s\n", memoryConfig.Embedding.Model)
			fmt.Printf("   3. Test connection: curl http://localhost:11434/api/tags\n")
		} else if memoryConfig.Embedding.Provider == "openai" {
			fmt.Printf("\n💡 OpenAI Troubleshooting:\n")
			fmt.Printf("   1. Set API key: export OPENAI_API_KEY=\"your-key\"\n")
			fmt.Printf("   2. Verify key is valid and has credits\n")
		}
		
		os.Exit(1)
	}
	defer memory.Close()
	
	// Test memory connection
	testContent := fmt.Sprintf("System initialized at %s", time.Now().Format("2006-01-02 15:04:05"))
	if err := memory.Store(ctx, testContent, "system-init"); err != nil {
		logger.Warn().Err(err).Msg("Memory connection test failed, continuing anyway")
		fmt.Printf("⚠️  Memory connection test failed: %v\n", err)
		fmt.Printf("Your agents will still work, but memory features may be limited\n")
	} else {
		logger.Info().Msg("Memory system initialized successfully")
		fmt.Printf("✅ Memory system ready!\n")
	}
	{{end}}

	agents := make(map[string]core.AgentHandler)
	results := make([]AgentOutput, 0)
	var resultsMutex sync.Mutex

	{{range .Agents}}
	// Create {{.DisplayName}} handler with result collection
	{{if $.Config.MemoryEnabled}}
	{{.Name}} := New{{.DisplayName}}(llmProvider, memory)
	{{else}}
	{{.Name}} := New{{.DisplayName}}(llmProvider)
	{{end}}
	wrapped{{.DisplayName}} := &ResultCollectorHandler{
		originalHandler: {{.Name}},
		agentName:       "{{.Name}}",
		outputs:         &results,
		mutex:           &resultsMutex,
	}
	agents["{{.Name}}"] = wrapped{{.DisplayName}}
	{{end}}

	// Create basic error handlers to prevent routing errors
	// These use the first agent as a fallback handler for simplicity
	{{if .Agents}}
	firstAgent := agents["{{(index .Agents 0).Name}}"]
	if firstAgent != nil {
		agents["error-handler"] = firstAgent
		agents["validation-error-handler"] = firstAgent
		agents["timeout-error-handler"] = firstAgent
		agents["critical-error-handler"] = firstAgent
		agents["high-priority-error-handler"] = firstAgent
		agents["network-error-handler"] = firstAgent
		agents["llm-error-handler"] = firstAgent
		agents["auth-error-handler"] = firstAgent
	}
	{{end}}

	// Create orchestrated runner
	{{if .Config.MemoryEnabled}}
	// Use NewRunnerFromConfig for automatic memory setup
	runner, err := core.NewRunnerFromConfig("agentflow.toml")
	if err != nil {
		logger.Error().Err(err).Msg("Failed to create runner from config")
		fmt.Printf("Error creating runner: %v\n", err)
		os.Exit(1)
	}
	
	// Register agents with the runner
	for name, handler := range agents {
		if err := runner.RegisterAgent(name, handler); err != nil {
			logger.Error().Err(err).Str("agent", name).Msg("Failed to register agent")
			fmt.Printf("Error registering agent %s: %v\n", name, err)
			os.Exit(1)
		}
	}
	{{else}}
	{{if eq .Config.OrchestrationMode "collaborative"}}
	runner := core.CreateCollaborativeRunner(agents, 30*time.Second)
	{{else if eq .Config.OrchestrationMode "sequential"}}
	runner := core.NewRunnerWithOrchestration(core.EnhancedRunnerConfig{
		RunnerConfig: core.RunnerConfig{
			Agents: agents,
		},
		OrchestrationMode: core.OrchestrationSequential,
		Config:            core.DefaultOrchestrationConfig(),
		SequentialAgents: []string{
			{{- range $i, $agent := .Agents}}
			{{- if $i}}, {{end}}"{{$agent.Name}}"
			{{- end}}
		},
	})
	{{else if eq .Config.OrchestrationMode "loop"}}
	runner := core.NewRunnerWithOrchestration(core.EnhancedRunnerConfig{
		RunnerConfig: core.RunnerConfig{
			Agents: agents,
		},
		OrchestrationMode: core.OrchestrationLoop,
		Config:            core.DefaultOrchestrationConfig(),
		SequentialAgents: []string{"{{(index .Agents 0).Name}}"}, // Loop uses single agent
	})
	{{else if eq .Config.OrchestrationMode "mixed"}}
	// Create mixed mode orchestration with CLI-specified collaborative and sequential agent groups
	{{if and (gt (len .Config.CollaborativeAgents) 0) (gt (len .Config.SequentialAgents) 0)}}
	// Use CLI-specified agent groups
	collaborativeAgents := []string{
		{{- range $i, $agent := .Config.CollaborativeAgents}}
		{{- if $i}}, {{end}}"{{$agent}}"
		{{- end}},
	}
	sequentialAgents := []string{
		{{- range $i, $agent := .Config.SequentialAgents}}
		{{- if $i}}, {{end}}"{{$agent}}"
		{{- end}},
	}
	{{else if gt (len .Agents) 1}}
	// Fallback: Split agents automatically when no specific groups are provided
	halfPoint := len([]string{
		{{- range $i, $agent := .Agents}}
		{{- if $i}}, {{end}}"{{$agent.Name}}"
		{{- end}}
	}) / 2
	allAgentNames := []string{
		{{- range $i, $agent := .Agents}}
		{{- if $i}}, {{end}}"{{$agent.Name}}"
		{{- end}}
	}
	
	var collaborativeAgents, sequentialAgents []string
	if halfPoint > 0 {
		collaborativeAgents = allAgentNames[:halfPoint]
		sequentialAgents = allAgentNames[halfPoint:]
	} else {
		// If only one agent, make it sequential
		sequentialAgents = allAgentNames
	}
	{{else}}
	// Single agent - use sequential mode
	sequentialAgents := []string{"{{(index .Agents 0).Name}}"}
	var collaborativeAgents []string
	{{end}}
	
	runner := core.NewRunnerWithOrchestration(core.EnhancedRunnerConfig{
		RunnerConfig: core.RunnerConfig{
			Agents: agents,
		},
		OrchestrationMode:   core.OrchestrationMixed,
		Config:              core.DefaultOrchestrationConfig(),
		CollaborativeAgents: collaborativeAgents,
		SequentialAgents:    sequentialAgents,
	})
	{{else}}
	// Default collaborative mode
	runner := core.CreateCollaborativeRunner(agents, 30*time.Second)
	{{end}}
	{{end}}

	{{if and (eq .Config.OrchestrationMode "collaborative") (not .Config.MemoryEnabled)}}
	// Result collection system
	var agentOutputs []AgentOutput
	var outputMutex sync.Mutex

	// Create a result collector by wrapping the existing agents (only for collaborative mode)
	for name, handler := range agents {
		if strings.Contains(name, "error-handler") {
			continue // Skip error handlers for result collection
		}

		// Wrap the original handler to capture outputs
		originalHandler := handler
		wrappedHandler := &ResultCollectorHandler{
			originalHandler: originalHandler,
			agentName:       name,
			outputs:         &agentOutputs,
			mutex:           &outputMutex,
		}
		agents[name] = wrappedHandler
	}

	// Recreate runner with wrapped agents
	runner = core.CreateCollaborativeRunner(agents, 30*time.Second)
	{{end}}

	var message string
	if *messageFlag != "" {
		message = *messageFlag
	} else {
		fmt.Print("Enter your message: ")
		fmt.Scanln(&message)
	}

	if message == "" {
		message = "Hello! Please provide information about current topics."
	}

	logger.Info().Str("message", message).Msg("Processing user message")

	{{if eq .Config.OrchestrationMode "collaborative"}}
	// Start the collaborative runner
	runner.Start(ctx)
	defer runner.Stop()

	{{if .Agents}}
	event := core.NewEvent("{{(index .Agents 0).Name}}", core.EventData{
		"message": message,
	}, map[string]string{
		"route": "{{(index .Agents 0).Name}}",
	})
	{{else}}
	event := core.NewEvent("user_request", core.EventData{
		"message": message,
	}, map[string]string{
		"route": "user_request",
	})
	{{end}}

	if err := runner.Emit(event); err != nil {
		logger.Error().Err(err).Msg("Workflow execution failed")
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	// Extended wait to allow all agents to complete their work
	logger.Info().Msg("Waiting for agents to complete processing...")
	time.Sleep(10 * time.Second)

	// Display collected agent outputs
	outputMutex.Lock()
	if len(agentOutputs) > 0 {
		fmt.Printf("\n=== Agent Results ===\n")
		for _, output := range agentOutputs {
			fmt.Printf("\n[%s] %s:\n", output.Timestamp.Format("15:04:05"), output.AgentName)
			fmt.Printf("%s\n", output.Content)
			fmt.Printf("%s\n", strings.Repeat("-", 50))
		}
	} else {
		logger.Debug().Msg("No agent outputs captured")
	}
	outputMutex.Unlock()

	logger.Info().Msg("Workflow completed successfully")
	{{else}}
	// Start the runner (non-blocking)
	runner.Start(ctx)

	{{if .Agents}}
	event := core.NewEvent("{{(index .Agents 0).Name}}", core.EventData{
		"message": message,
	}, map[string]string{
		"route": "{{(index .Agents 0).Name}}",
	})
	{{else}}
	event := core.NewEvent("user_request", core.EventData{
		"message": message,
	}, map[string]string{
		"route": "user_request",
	})
	{{end}}

	if err := runner.Emit(event); err != nil {
		logger.Error().Err(err).Msg("Workflow execution failed")
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	// Wait for processing to complete BEFORE printing results.
	// We call runner.Stop() explicitly here (instead of using defer runner.Stop()).
	// A deferred call would execute only when main() returns—after the result-printing
	// code below—so we could print an empty "Agent Responses" section while the
	// agents are still working.  Calling Stop() now closes the queue and blocks
	// until the event-processing goroutine has finished handling all queued events,
	// guaranteeing the results slice is fully populated.
	logger.Info().Msg("Waiting for agents to complete processing...")
	runner.Stop()

	// Display collected results
	fmt.Printf("\n=== Agent Responses ===\n")
	resultsMutex.Lock()
	if len(results) > 0 {
		for _, result := range results {
			fmt.Printf("\n🤖 %s:\n%s\n", result.AgentName, result.Content)
			fmt.Printf("⏰ %s\n", result.Timestamp.Format("15:04:05"))
		}
	} else {
		fmt.Printf("No agent responses captured. This might indicate:\n")
		fmt.Printf("- LLM provider credentials are not configured\n")
		fmt.Printf("- The agent encountered an error during processing\n")
		fmt.Printf("- The LLM provider is not responding\n")
	}
	resultsMutex.Unlock()

	fmt.Printf("\n=== Workflow Completed ===\n")
	fmt.Printf("Check the logs above for detailed agent execution results.\n")
	{{end}}

	logger.Info().Msg("Workflow completed successfully")
}

{{.ProviderInitFunction}}

{{if .Config.MCPEnabled}}
{{.MCPInitFunction}}
{{end}}

{{if .Config.WithCache}}
{{.CacheInitFunction}}
{{end}}

// ResultCollectorHandler wraps an agent handler to capture its outputs
type ResultCollectorHandler struct {
	originalHandler core.AgentHandler
	agentName       string
	outputs         *[]AgentOutput
	mutex           *sync.Mutex
}

// AgentOutput holds the output from an agent
type AgentOutput struct {
	AgentName string
	Content   string
	Timestamp time.Time
}

// Run implements the AgentHandler interface and captures the output
func (r *ResultCollectorHandler) Run(ctx context.Context, event core.Event, state core.State) (core.AgentResult, error) {
	// Call the original handler
	result, err := r.originalHandler.Run(ctx, event, state)

	// Extract meaningful content from the result
	var content string
	if err != nil {
		content = fmt.Sprintf("Error: %v", err)
	} else if result.Error != "" {
		content = fmt.Sprintf("Agent Error: %s", result.Error)
	} else {
		// Try to extract content from the result's output state
		if result.OutputState != nil {
			if responseData, exists := result.OutputState.Get("response"); exists {
				if responseStr, ok := responseData.(string); ok {
					content = responseStr
				}
			}
			if content == "" {
				if outputData, exists := result.OutputState.Get("output"); exists {
					if outputStr, ok := outputData.(string); ok {
						content = outputStr
					}
				}
			}
			if content == "" {
				if messageData, exists := result.OutputState.Get("message"); exists {
					if messageStr, ok := messageData.(string); ok {
						content = messageStr
					}
				}
			}
		}
	}

	// If we still don't have content, create a summary
	if content == "" {
		content = fmt.Sprintf("Agent %s completed processing successfully", r.agentName)
	}

	// Store the output
	r.mutex.Lock()
	*r.outputs = append(*r.outputs, AgentOutput{
		AgentName: r.agentName,
		Content:   content,
		Timestamp: time.Now(),
	})
	r.mutex.Unlock()

	return result, err
}

{{if .Config.MemoryEnabled}}
// validateMemoryConfig validates the memory configuration against expected values
func validateMemoryConfig(memoryConfig core.AgentMemoryConfig, expectedModel string) error {
	// Validate embedding dimensions
	expectedDimensions := {{.Config.EmbeddingDimensions}}
	if memoryConfig.Dimensions != expectedDimensions {
		return fmt.Errorf("%s requires %d dimensions, but %d configured in agentflow.toml\n💡 Solution: Update [agent_memory] dimensions = %d", 
			expectedModel, expectedDimensions, memoryConfig.Dimensions, expectedDimensions)
	}
	
	// Validate embedding provider and model
	expectedProvider := "{{.Config.EmbeddingProvider}}"
	expectedModelName := "{{.Config.EmbeddingModel}}"
	
	if memoryConfig.Embedding.Provider != expectedProvider {
		return fmt.Errorf("embedding provider mismatch: expected '%s', got '%s'\n💡 Solution: Update [agent_memory.embedding] provider = \"%s\"", 
			expectedProvider, memoryConfig.Embedding.Provider, expectedProvider)
	}
	
	if memoryConfig.Embedding.Model != expectedModelName {
		return fmt.Errorf("embedding model mismatch: expected '%s', got '%s'\n💡 Solution: Update [agent_memory.embedding] model = \"%s\"", 
			expectedModelName, memoryConfig.Embedding.Model, expectedModelName)
	}
	
	// Validate memory provider configuration
	switch memoryConfig.Provider {
	case "pgvector":
		if memoryConfig.Connection == "" {
			return fmt.Errorf("pgvector provider requires a connection string\n💡 Solution: Set [agent_memory] connection = \"postgres://user:password@localhost:15432/agentflow?sslmode=disable\"")
		}
		if !strings.Contains(memoryConfig.Connection, "postgres://") {
			return fmt.Errorf("pgvector connection string should start with 'postgres://'\n💡 Current: %s", memoryConfig.Connection)
		}
	case "weaviate":
		if memoryConfig.Connection == "" {
			return fmt.Errorf("weaviate provider requires a connection string\n💡 Solution: Set [agent_memory] connection = \"http://localhost:8080\"")
		}
		if !strings.Contains(memoryConfig.Connection, "http") {
			return fmt.Errorf("weaviate connection string should be an HTTP URL\n💡 Current: %s", memoryConfig.Connection)
		}
	case "memory":
		// In-memory provider doesn't need connection validation
	default:
		return fmt.Errorf("unknown memory provider: %s\n💡 Valid options: memory, pgvector, weaviate", memoryConfig.Provider)
	}
	
	// Validate RAG configuration if enabled
	{{if .Config.RAGEnabled}}
	if memoryConfig.EnableRAG {
		if memoryConfig.ChunkSize <= 0 {
			return fmt.Errorf("RAG chunk size must be positive, got %d\n💡 Solution: Set [agent_memory] chunk_size = 1000", memoryConfig.ChunkSize)
		}
		if memoryConfig.ChunkOverlap < 0 || memoryConfig.ChunkOverlap >= memoryConfig.ChunkSize {
			return fmt.Errorf("RAG chunk overlap must be between 0 and chunk_size (%d), got %d\n💡 Solution: Set [agent_memory] chunk_overlap = 100", 
				memoryConfig.ChunkSize, memoryConfig.ChunkOverlap)
		}
		if memoryConfig.KnowledgeScoreThreshold < 0.0 || memoryConfig.KnowledgeScoreThreshold > 1.0 {
			return fmt.Errorf("RAG score threshold must be between 0.0 and 1.0, got %.2f\n💡 Solution: Set [agent_memory] knowledge_score_threshold = 0.7", 
				memoryConfig.KnowledgeScoreThreshold)
		}
	}
	{{end}}
	
	// Validate embedding provider specific settings
	switch memoryConfig.Embedding.Provider {
	case "ollama":
		if memoryConfig.Embedding.BaseURL == "" {
			return fmt.Errorf("ollama embedding provider requires base_url\n💡 Solution: Set [agent_memory.embedding] base_url = \"http://localhost:11434\"")
		}
	case "openai":
		// OpenAI uses environment variables, so we can't validate API key here
		// But we can check if the model name looks reasonable
		if !strings.Contains(memoryConfig.Embedding.Model, "embedding") {
			return fmt.Errorf("OpenAI model '%s' doesn't look like an embedding model\n💡 Recommended: text-embedding-3-small or text-embedding-3-large", 
				memoryConfig.Embedding.Model)
		}
	case "dummy":
		// Dummy provider doesn't need validation
	default:
		return fmt.Errorf("unknown embedding provider: %s\n💡 Valid options: openai, ollama, dummy", memoryConfig.Embedding.Provider)
	}
	
	return nil
}
{{end}}
`
