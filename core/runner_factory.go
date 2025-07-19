// Package core provides public factory functions for creating runners and tool registries in AgentFlow.
package core

import (
	"context"
	"fmt"
	"log"
)

// RunnerConfig allows customization but provides sensible defaults.
// Breaking change: Memory is now required
type RunnerConfig struct {
	QueueSize    int
	Orchestrator Orchestrator
	Agents       map[string]AgentHandler
	Memory       Memory      // REQUIRED: Memory is now central to the system
	SessionID    string      // REQUIRED: Session ID for memory operations
	TraceLogger  TraceLogger // Optional trace logger
	ConfigPath   string      // Path to agentflow.toml config file
	Config       *Config     // Pre-loaded configuration (optional)
}

// NewRunnerWithConfig wires up everything, registers agents, and returns a ready-to-use runner.
// Breaking change: Memory and SessionID are now required
func NewRunnerWithConfig(cfg RunnerConfig) Runner {
	// Breaking change: Memory is required
	if cfg.Memory == nil {
		log.Fatal("Memory is required in RunnerConfig - use NewRunnerFromConfig() for automatic memory setup")
	}

	// Breaking change: SessionID is required
	if cfg.SessionID == "" {
		log.Fatal("SessionID is required in RunnerConfig - use NewRunnerFromConfig() for automatic session setup")
	}

	// Load configuration if specified
	var config *Config
	if cfg.Config != nil {
		config = cfg.Config
	} else if cfg.ConfigPath != "" {
		var err error
		config, err = LoadConfig(cfg.ConfigPath)
		if err != nil {
			log.Printf("Warning: Failed to load config from %s: %v", cfg.ConfigPath, err)
		}
	} else {
		// Try to load from working directory
		var err error
		config, err = LoadConfigFromWorkingDir()
		if err != nil {
			log.Printf("Info: No agentflow.toml found in working directory: %v", err)
		}
	}

	// Apply configuration settings
	if config != nil {
		config.ApplyLoggingConfig()

		// Use configuration values if not specified in RunnerConfig
		if cfg.QueueSize <= 0 && config.Runtime.MaxConcurrentAgents > 0 {
			cfg.QueueSize = config.Runtime.MaxConcurrentAgents
		}

		Logger().Info().
			Str("config_name", config.AgentFlow.Name).
			Str("config_version", config.AgentFlow.Version).
			Str("config_provider", config.AgentFlow.Provider).
			Str("session_id", cfg.SessionID).
			Str("memory_provider", getMemoryProviderName(cfg.Memory)).
			Str("log_level", config.Logging.Level).
			Msg("Loaded AgentFlow configuration with memory")
	}

	queueSize := cfg.QueueSize
	if queueSize <= 0 {
		queueSize = 10
	}
	runner := NewRunner(queueSize)

	// Callbacks and tracing
	callbackRegistry := NewCallbackRegistry()

	// Use provided trace logger or create default
	var traceLogger TraceLogger
	if cfg.TraceLogger != nil {
		traceLogger = cfg.TraceLogger
	} else {
		traceLogger = NewInMemoryTraceLogger()
	}

	runner.SetCallbackRegistry(callbackRegistry)
	runner.SetTraceLogger(traceLogger)
	RegisterTraceHooks(callbackRegistry, traceLogger)

	// Orchestrator
	var orch Orchestrator
	if cfg.Orchestrator != nil {
		orch = cfg.Orchestrator
	} else {
		orch = NewRouteOrchestrator(callbackRegistry)
	}
	runner.SetOrchestrator(orch)

	// Breaking change: Wrap all agent handlers to include memory in context
	wrappedAgents := make(map[string]AgentHandler)
	for name, agent := range cfg.Agents {
		wrappedAgents[name] = &MemoryAwareAgentHandler{
			handler:   agent,
			memory:    cfg.Memory,
			sessionID: cfg.SessionID,
		}
	}

	// Register wrapped agents
	for name, agent := range wrappedAgents {
		if err := runner.RegisterAgent(name, agent); err != nil {
			log.Fatalf("Failed to register agent %s: %v", name, err)
		}
	}

	// Register a default no-op error handler if not present
	if _, ok := cfg.Agents["error-handler"]; !ok {
		runner.RegisterAgent("error-handler", &MemoryAwareAgentHandler{
			handler: AgentHandlerFunc(
				func(ctx context.Context, event Event, state State) (AgentResult, error) {
					state.SetMeta(RouteMetadataKey, "")
					return AgentResult{OutputState: state}, nil
				},
			),
			memory:    cfg.Memory,
			sessionID: cfg.SessionID,
		})
	}

	// Automatically configure error routing based on available error handlers
	configureErrorRouting(runner, cfg.Agents)

	return runner
}

// Breaking change: New convenience constructor that auto-initializes memory
// NewRunnerFromConfig creates a runner with automatic memory setup from config
func NewRunnerFromConfig(configPath string) (Runner, error) {
	config, err := LoadConfig(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load config from %s: %w", configPath, err)
	}

	// Auto-initialize memory based on config
	memory, err := NewMemory(config.AgentMemory)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize memory: %w", err)
	}

	// Generate session ID
	sessionID := GenerateSessionID()

	// Create runner config with memory
	runnerConfig := RunnerConfig{
		Config:    config,
		Memory:    memory,
		SessionID: sessionID,
		Agents:    make(map[string]AgentHandler), // Empty - agents added separately
	}

	return NewRunnerWithConfig(runnerConfig), nil
}

// Breaking change: Convenience constructor for quick setup
// NewRunner creates a runner with memory and sessionID - replaces old constructor
func NewRunnerWithMemory(agents map[string]AgentHandler, memory Memory, sessionID string) Runner {
	if memory == nil {
		memory = QuickMemory() // Use in-memory as fallback
	}
	if sessionID == "" {
		sessionID = GenerateSessionID()
	}

	return NewRunnerWithConfig(RunnerConfig{
		Agents:    agents,
		Memory:    memory,
		SessionID: sessionID,
	})
}

// DEPRECATED: Use NewRunnerFromConfig() instead
// NewRunnerWithConfigFile creates a runner by loading configuration from the specified file
func NewRunnerWithConfigFile(configPath string, agents map[string]AgentHandler) Runner {
	log.Println("WARNING: NewRunnerWithConfigFile is deprecated. Use NewRunnerFromConfig() instead for proper memory support.")

	// For backward compatibility, create with in-memory provider
	memory := QuickMemory()
	sessionID := GenerateSessionID()

	return NewRunnerWithConfig(RunnerConfig{
		ConfigPath: configPath,
		Agents:     agents,
		Memory:     memory,
		SessionID:  sessionID,
	})
}

// DEPRECATED: Use NewRunnerFromConfig() instead
// NewRunnerFromWorkingDir creates a runner by loading agentflow.toml from the working directory
func NewRunnerFromWorkingDir(agents map[string]AgentHandler) Runner {
	log.Println("WARNING: NewRunnerFromWorkingDir is deprecated. Use NewRunnerFromConfig() instead for proper memory support.")

	// For backward compatibility, create with in-memory provider
	memory := QuickMemory()
	sessionID := GenerateSessionID()

	return NewRunnerWithConfig(RunnerConfig{
		Agents:    agents,
		Memory:    memory,
		SessionID: sessionID,
	})
}

// NewProviderFromConfig creates a ModelProvider from the loaded configuration
func NewProviderFromConfig(configPath string) (ModelProvider, error) {
	config, err := LoadConfig(configPath)
	if err != nil {
		return nil, err
	}
	return config.InitializeProvider()
}

// NewProviderFromWorkingDir creates a ModelProvider from agentflow.toml in the working directory
func NewProviderFromWorkingDir() (ModelProvider, error) {
	config, err := LoadConfigFromWorkingDir()
	if err != nil {
		return nil, err
	}
	return config.InitializeProvider()
}

// MemoryAwareAgentHandler wraps an agent handler to provide memory through context
// Breaking change: All agents now have automatic memory access
type MemoryAwareAgentHandler struct {
	handler   AgentHandler
	memory    Memory
	sessionID string
}

func (m *MemoryAwareAgentHandler) Run(ctx context.Context, event Event, state State) (AgentResult, error) {
	// Breaking change: Always inject memory into context
	ctx = WithMemory(ctx, m.memory, m.sessionID)
	return m.handler.Run(ctx, event, state)
}

// NewMemoryAwareAgentHandler creates a new memory-aware agent handler wrapper
func NewMemoryAwareAgentHandler(handler AgentHandler, memory Memory, sessionID string) *MemoryAwareAgentHandler {
	return &MemoryAwareAgentHandler{
		handler:   handler,
		memory:    memory,
		sessionID: sessionID,
	}
}

// getMemoryProviderName returns a string representation of the memory provider type
func getMemoryProviderName(memory Memory) string {
	if memory == nil {
		return "none"
	}

	switch memory.(type) {
	case *InMemoryProvider:
		return "memory"
	case *NoOpMemory:
		return "noop"
	default:
		return "unknown"
	}
}

// configureErrorRouting automatically configures error routing based on available error handlers
func configureErrorRouting(runner Runner, agents map[string]AgentHandler) {
	if runner == nil {
		return
	}

	// Create a minimal error config that only uses handlers that actually exist
	errorConfig := DefaultErrorRouterConfig()

	// Clear all the default handlers first to avoid referencing non-existent handlers
	errorConfig.CategoryHandlers = make(map[string]string)
	errorConfig.SeverityHandlers = make(map[string]string)

	// Only configure handlers that actually exist
	if agents != nil {
		// Set the default error handler if available
		if _, exists := agents["error-handler"]; exists {
			errorConfig.ErrorHandlerName = "error-handler"
		} else if _, exists := agents["error_handler"]; exists {
			errorConfig.ErrorHandlerName = "error_handler"
		} else {
			// If no error handler exists, use the default one we registered
			errorConfig.ErrorHandlerName = "error-handler"
		}

		// Only add specialized handlers if they actually exist
		if _, exists := agents["validation-error-handler"]; exists {
			errorConfig.CategoryHandlers[ErrorCodeValidation] = "validation-error-handler"
		}
		if _, exists := agents["timeout-error-handler"]; exists {
			errorConfig.CategoryHandlers[ErrorCodeTimeout] = "timeout-error-handler"
		}
		if _, exists := agents["critical-error-handler"]; exists {
			errorConfig.SeverityHandlers[SeverityCritical] = "critical-error-handler"
		}
		if _, exists := agents["network-error-handler"]; exists {
			errorConfig.CategoryHandlers[ErrorCodeNetwork] = "network-error-handler"
		}
		if _, exists := agents["llm-error-handler"]; exists {
			errorConfig.CategoryHandlers[ErrorCodeLLM] = "llm-error-handler"
		}
		if _, exists := agents["auth-error-handler"]; exists {
			errorConfig.CategoryHandlers[ErrorCodeAuth] = "auth-error-handler"
		}
	} else {
		// No agents provided, just use the default error handler
		errorConfig.ErrorHandlerName = "error-handler"
	}

	// Apply the configuration to the runner
	if runnerImpl, ok := runner.(*RunnerImpl); ok {
		runnerImpl.SetErrorRouterConfig(errorConfig)
	}
}
