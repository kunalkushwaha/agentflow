// Package core provides orchestration capabilities for multi-agent systems
package core

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"
)

// =============================================================================
// ORCHESTRATION TYPES AND CONSTANTS
// =============================================================================

// OrchestrationMode defines how events are distributed to agents
type OrchestrationMode string

const (
	// OrchestrationRoute sends each event to a single agent based on routing metadata (default behavior)
	OrchestrationRoute OrchestrationMode = "route"
	// OrchestrationCollaborate sends each event to ALL registered agents in parallel
	OrchestrationCollaborate OrchestrationMode = "collaborate"
	// OrchestrationSequential processes agents one after another
	OrchestrationSequential OrchestrationMode = "sequential"
	// OrchestrationParallel processes agents in parallel (similar to collaborate but different semantics)
	OrchestrationParallel OrchestrationMode = "parallel"
	// OrchestrationLoop repeats processing with a single agent
	OrchestrationLoop OrchestrationMode = "loop"
	// OrchestrationMixed combines collaborative and sequential patterns in hybrid workflows
	OrchestrationMixed OrchestrationMode = "mixed"
)

// OrchestrationConfig contains configuration for orchestration behavior
type OrchestrationConfig struct {
	Timeout          time.Duration // Overall timeout for orchestration operations
	MaxConcurrency   int           // Maximum number of concurrent agent executions
	FailureThreshold float64       // Percentage of failures before stopping (0.0-1.0)
	RetryPolicy      *RetryPolicy  // Policy for retrying failed operations (uses existing RetryPolicy from core)
}

// DefaultOrchestrationConfig returns sensible defaults for orchestration configuration
func DefaultOrchestrationConfig() OrchestrationConfig {
	return OrchestrationConfig{
		Timeout:          30 * time.Second,
		MaxConcurrency:   10,
		FailureThreshold: 0.5, // Stop if 50% of agents fail
		RetryPolicy:      DefaultRetryPolicy(),
	}
}

// =============================================================================
// ENHANCED RUNNER CONFIGURATION
// =============================================================================

// EnhancedRunnerConfig extends RunnerConfig with orchestration options
type EnhancedRunnerConfig struct {
	RunnerConfig
	OrchestrationMode   OrchestrationMode   // Mode of orchestration (route, collaborate, mixed, etc.)
	Config              OrchestrationConfig // Orchestration-specific configuration
	CollaborativeAgents []string            // List of agent names for collaborative execution
	SequentialAgents    []string            // List of agent names for sequential execution
}

// =============================================================================
// ORCHESTRATION CONSTRUCTORS
// =============================================================================

// NewCollaborativeOrchestrator creates an orchestrator that runs all agents in parallel
// Each event is sent to ALL registered agents simultaneously
func NewCollaborativeOrchestrator(registry *CallbackRegistry) Orchestrator {
	return &collaborativeOrchestrator{
		handlers:         make(map[string]AgentHandler),
		callbackRegistry: registry,
	}
}

// NewSequentialOrchestrator creates an orchestrator that runs agents in sequence
func NewSequentialOrchestrator(registry *CallbackRegistry, agentNames []string) Orchestrator {
	return &sequentialOrchestrator{
		handlers:         make(map[string]AgentHandler),
		agentSequence:    agentNames,
		callbackRegistry: registry,
	}
}

// NewLoopOrchestrator creates an orchestrator that runs a single agent in a loop
func NewLoopOrchestrator(registry *CallbackRegistry, agentNames []string) Orchestrator {
	agentName := ""
	if len(agentNames) > 0 {
		agentName = agentNames[0] // Use first agent for loop
	}
	return &loopOrchestrator{
		handlers:         make(map[string]AgentHandler),
		agentName:        agentName,
		maxIterations:    5, // Default iterations
		callbackRegistry: registry,
	}
}

// NewRunnerWithOrchestration creates a runner with specified orchestration mode
func NewRunnerWithOrchestration(cfg EnhancedRunnerConfig) Runner {
	// Create base runner with standard configuration
	runner := NewRunnerWithConfig(cfg.RunnerConfig)

	// Override orchestrator based on mode
	callbackRegistry := runner.GetCallbackRegistry()

	var orch Orchestrator
	switch cfg.OrchestrationMode {
	case OrchestrationCollaborate:
		orch = NewCollaborativeOrchestrator(callbackRegistry)
	case OrchestrationMixed:
		orch = NewMixedOrchestrator(callbackRegistry, cfg.CollaborativeAgents, cfg.SequentialAgents)
	case OrchestrationSequential:
		orch = NewSequentialOrchestrator(callbackRegistry, cfg.SequentialAgents)
	case OrchestrationLoop:
		orch = NewLoopOrchestrator(callbackRegistry, cfg.SequentialAgents)
	default:
		orch = NewRouteOrchestrator(callbackRegistry)
	}

	// Set the orchestrator on the runner
	if runnerImpl, ok := runner.(*RunnerImpl); ok {
		runnerImpl.SetOrchestrator(orch)

		// Re-register all agents with the new orchestrator since SetOrchestrator replaces it
		for name, agent := range cfg.RunnerConfig.Agents {
			if err := orch.RegisterAgent(name, agent); err != nil {
				Logger().Error().Str("agent", name).Err(err).Msg("Failed to register agent with new orchestrator")
			}
		}

		// Re-register the default error handler if it wasn't provided
		if _, exists := cfg.RunnerConfig.Agents["error-handler"]; !exists {
			orch.RegisterAgent("error-handler", AgentHandlerFunc(
				func(ctx context.Context, event Event, state State) (AgentResult, error) {
					state.SetMeta(RouteMetadataKey, "")
					return AgentResult{OutputState: state}, nil
				},
			))
		}
	}

	return runner
}

// =============================================================================
// ORCHESTRATION BUILDER PATTERN
// =============================================================================

// OrchestrationBuilder provides fluent interface for orchestration setup
type OrchestrationBuilder struct {
	mode   OrchestrationMode
	agents map[string]AgentHandler
	config OrchestrationConfig
}

// NewOrchestrationBuilder creates a new orchestration builder with the specified mode
func NewOrchestrationBuilder(mode OrchestrationMode) *OrchestrationBuilder {
	return &OrchestrationBuilder{
		mode:   mode,
		agents: make(map[string]AgentHandler),
		config: DefaultOrchestrationConfig(),
	}
}

// WithAgent adds a single agent to the orchestration
func (ob *OrchestrationBuilder) WithAgent(name string, handler AgentHandler) *OrchestrationBuilder {
	ob.agents[name] = handler
	return ob
}

// WithAgents adds multiple agents to the orchestration from a map
func (ob *OrchestrationBuilder) WithAgents(agents map[string]AgentHandler) *OrchestrationBuilder {
	for name, handler := range agents {
		ob.agents[name] = handler
	}
	return ob
}

// WithTimeout sets the orchestration timeout
func (ob *OrchestrationBuilder) WithTimeout(timeout time.Duration) *OrchestrationBuilder {
	ob.config.Timeout = timeout
	return ob
}

// WithMaxConcurrency sets the maximum number of concurrent agents
func (ob *OrchestrationBuilder) WithMaxConcurrency(max int) *OrchestrationBuilder {
	ob.config.MaxConcurrency = max
	return ob
}

// WithFailureThreshold sets the failure threshold (0.0-1.0)
// When this percentage of agents fail, the orchestration will stop
func (ob *OrchestrationBuilder) WithFailureThreshold(threshold float64) *OrchestrationBuilder {
	ob.config.FailureThreshold = threshold
	return ob
}

// WithRetryPolicy sets the retry policy for failed agents
func (ob *OrchestrationBuilder) WithRetryPolicy(policy *RetryPolicy) *OrchestrationBuilder {
	ob.config.RetryPolicy = policy
	return ob
}

// WithConfig sets the complete orchestration configuration
func (ob *OrchestrationBuilder) WithConfig(config OrchestrationConfig) *OrchestrationBuilder {
	ob.config = config
	return ob
}

// Build creates the configured runner with the specified orchestration mode
func (ob *OrchestrationBuilder) Build() Runner {
	// Ensure we have memory and sessionID to satisfy Runner requirements
	memory := QuickMemory()
	sessionID := GenerateSessionID()

	return NewRunnerWithOrchestration(EnhancedRunnerConfig{
		RunnerConfig: RunnerConfig{
			Agents:    ob.agents,
			Memory:    memory,
			SessionID: sessionID,
		},
		OrchestrationMode: ob.mode,
		Config:            ob.config,
	})
}

// =============================================================================
// CONVENIENCE FUNCTIONS
// =============================================================================

// CreateCollaborativeRunner creates a runner where all agents process events in parallel
// Each event is sent to ALL registered agents simultaneously
func CreateCollaborativeRunner(agents map[string]AgentHandler, timeout time.Duration) Runner {
	return NewOrchestrationBuilder(OrchestrationCollaborate).
		WithAgents(agents).
		WithTimeout(timeout).
		Build()
}

// CreateRouteRunner creates a standard routing runner (existing behavior)
// Each event is sent to a single agent based on routing metadata
func CreateRouteRunner(agents map[string]AgentHandler) Runner {
	return NewOrchestrationBuilder(OrchestrationRoute).
		WithAgents(agents).
		Build()
}

// CreateHighThroughputRunner creates a collaborative runner optimized for high throughput
// Uses higher concurrency limits and more tolerant failure thresholds
func CreateHighThroughputRunner(agents map[string]AgentHandler) Runner {
	return NewOrchestrationBuilder(OrchestrationCollaborate).
		WithAgents(agents).
		WithMaxConcurrency(50).
		WithTimeout(60 * time.Second).
		WithFailureThreshold(0.8). // Tolerate 80% failures
		Build()
}

// CreateFaultTolerantRunner creates a collaborative runner with aggressive retry policies
// Designed for environments where transient failures are common
func CreateFaultTolerantRunner(agents map[string]AgentHandler) Runner {
	retryPolicy := RetryPolicy{
		MaxRetries:    5,
		BackoffFactor: 1.5,
		MaxDelay:      30 * time.Second,
	}

	return NewOrchestrationBuilder(OrchestrationCollaborate).
		WithAgents(agents).
		WithRetryPolicy(&retryPolicy).
		WithFailureThreshold(0.9). // Very tolerant of failures
		Build()
}

// CreateLoadBalancedRunner creates a runner that distributes load across multiple agent instances
// Useful for scaling horizontally with multiple instances of the same agent type
func CreateLoadBalancedRunner(agents map[string]AgentHandler, maxConcurrency int) Runner {
	return NewOrchestrationBuilder(OrchestrationRoute).
		WithAgents(agents).
		WithMaxConcurrency(maxConcurrency).
		WithTimeout(30 * time.Second).
		Build()
}

// =============================================================================
// ORCHESTRATION UTILITIES
// =============================================================================

// ConvertAgentToHandler converts an Agent to an AgentHandler for use in orchestration
// This is a utility function to bridge between Agent and AgentHandler interfaces
func ConvertAgentToHandler(agent Agent) AgentHandler {
	return AgentHandlerFunc(func(ctx context.Context, event Event, state State) (AgentResult, error) {
		// Merge event data into state
		if eventData := event.GetData(); eventData != nil {
			for key, value := range eventData {
				state.Set(key, value)
			}
		}

		// Run the agent
		outputState, err := agent.Run(ctx, state)
		if err != nil {
			return AgentResult{Error: err.Error()}, err
		}

		return AgentResult{OutputState: outputState}, nil
	})
}

// CreateMixedOrchestration creates a runner that combines multiple orchestration patterns
// This allows for complex workflows where different agent groups use different orchestration modes
func CreateMixedOrchestration(routeAgents, collaborativeAgents map[string]AgentHandler) Runner {
	// Combine all agents into a single map
	allAgents := make(map[string]AgentHandler)
	for name, handler := range routeAgents {
		allAgents[name] = handler
	}
	for name, handler := range collaborativeAgents {
		allAgents[name] = handler
	}

	// For now, use route orchestration as the base
	// In the future, this could be enhanced to support mixed modes
	return CreateRouteRunner(allAgents)
}

// =============================================================================
// COLLABORATIVE ORCHESTRATOR IMPLEMENTATION
// =============================================================================

// collaborativeOrchestrator implements the Orchestrator interface for collaborative mode
type collaborativeOrchestrator struct {
	handlers         map[string]AgentHandler
	callbackRegistry *CallbackRegistry
	mu               sync.RWMutex
}

// RegisterAgent adds an agent handler to the collaborative orchestrator
func (o *collaborativeOrchestrator) RegisterAgent(name string, handler AgentHandler) error {
	o.mu.Lock()
	defer o.mu.Unlock()

	if _, exists := o.handlers[name]; exists {
		return fmt.Errorf("agent with name '%s' already registered", name)
	}

	o.handlers[name] = handler
	Logger().Info().
		Str("agent", name).
		Msg("CollaborativeOrchestrator: Registered agent")
	return nil
}

// GetCallbackRegistry returns the callback registry
func (o *collaborativeOrchestrator) GetCallbackRegistry() *CallbackRegistry {
	return o.callbackRegistry
}

// Stop halts the orchestrator
func (o *collaborativeOrchestrator) Stop() {
	// Implementation for stopping the orchestrator
	Logger().Info().Msg("CollaborativeOrchestrator: Stopped")
}

// Dispatch sends the event to all registered agent handlers concurrently
func (o *collaborativeOrchestrator) Dispatch(ctx context.Context, event Event) (AgentResult, error) {
	if event == nil {
		Logger().Warn().Msg("CollaborativeOrchestrator: Received nil event, skipping dispatch.")
		err := errors.New("cannot dispatch nil event")
		return AgentResult{Error: err.Error()}, err
	}

	o.mu.RLock()
	defer o.mu.RUnlock()

	if len(o.handlers) == 0 {
		Logger().Warn().Msg("CollaborativeOrchestrator: No agents registered")
		err := errors.New("no agents registered")
		return AgentResult{Error: err.Error()}, err
	}

	// Create a channel to collect results from all agents
	resultChan := make(chan AgentResult, len(o.handlers))
	var wg sync.WaitGroup

	// Extract the current state from the event data
	currentState := NewState()
	if event != nil {
		// Create state from event data - using event metadata and data
		for k, v := range event.GetMetadata() {
			currentState.SetMeta(k, v)
		}
		// Add event data to state
		eventData := event.GetData()
		for key, value := range eventData {
			currentState.Set(key, value)
		}
	}

	// Launch goroutines for each agent handler
	for name, handler := range o.handlers {
		wg.Add(1)
		go func(agentName string, h AgentHandler) {
			defer wg.Done()
			Logger().Debug().
				Str("agent", agentName).
				Msg("CollaborativeOrchestrator: Dispatching to agent")

			result, err := h.Run(ctx, event, currentState)
			if err != nil {
				result.Error = err.Error()
			}
			resultChan <- result
		}(name, handler)
	}

	// Wait for all agents to complete
	go func() {
		wg.Wait()
		close(resultChan)
	}()

	// Collect all results
	var results []AgentResult
	var errors []string
	hasSuccess := false
	combinedState := NewState()

	for result := range resultChan {
		results = append(results, result)
		if result.Error != "" {
			errors = append(errors, result.Error)
		} else {
			hasSuccess = true
			// Merge output states (iterate through keys since State is an interface)
			for _, key := range result.OutputState.Keys() {
				if value, ok := result.OutputState.Get(key); ok {
					combinedState.Set(key, value)
				}
			}
			// Also merge metadata
			for _, key := range result.OutputState.MetaKeys() {
				if value, ok := result.OutputState.GetMeta(key); ok {
					combinedState.SetMeta(key, value)
				}
			}
		}
	}

	// Create combined result
	combinedResult := AgentResult{
		OutputState: combinedState,
		StartTime:   time.Now(),
		EndTime:     time.Now(),
	}

	// If all agents failed, return error
	if !hasSuccess {
		combinedResult.Error = fmt.Sprintf("all agents failed: %v", errors)
		return combinedResult, fmt.Errorf("collaborative dispatch failed: all agents returned errors")
	}

	Logger().Info().
		Int("total_agents", len(o.handlers)).
		Int("successful", len(results)-len(errors)).
		Int("failed", len(errors)).
		Msg("CollaborativeOrchestrator: Dispatch completed")

	return combinedResult, nil
}

// =============================================================================
// MIXED ORCHESTRATOR IMPLEMENTATION
// =============================================================================

// mixedOrchestrator implements hybrid orchestration combining collaborative and sequential patterns
type mixedOrchestrator struct {
	handlers                map[string]AgentHandler
	collaborativeAgents     map[string]AgentHandler // Agents that run in parallel
	collaborativeAgentNames []string                // Names of collaborative agents
	sequentialAgents        []string                // Agent names that run in sequence
	callbackRegistry        *CallbackRegistry
	mu                      sync.RWMutex
}

// NewMixedOrchestrator creates an orchestrator that combines collaborative and sequential execution
func NewMixedOrchestrator(registry *CallbackRegistry, collaborativeAgentNames, sequentialAgentNames []string) Orchestrator {
	return &mixedOrchestrator{
		handlers:                make(map[string]AgentHandler),
		collaborativeAgents:     make(map[string]AgentHandler),
		collaborativeAgentNames: collaborativeAgentNames,
		sequentialAgents:        sequentialAgentNames,
		callbackRegistry:        registry,
	}
}

// RegisterAgent adds an agent to the appropriate execution group
func (o *mixedOrchestrator) RegisterAgent(name string, handler AgentHandler) error {
	o.mu.Lock()
	defer o.mu.Unlock()

	if handler == nil {
		return fmt.Errorf("handler cannot be nil for agent %s", name)
	}

	o.handlers[name] = handler

	// Determine which group this agent belongs to
	isCollaborative := false
	for _, collabName := range o.getCollaborativeAgentNames() {
		if collabName == name {
			o.collaborativeAgents[name] = handler
			isCollaborative = true
			break
		}
	}

	Logger().Debug().
		Str("agent", name).
		Bool("collaborative", isCollaborative).
		Bool("sequential", !isCollaborative && o.isSequentialAgent(name)).
		Msg("MixedOrchestrator: Agent registered")

	return nil
}

// Dispatch implements hybrid execution: collaborative agents run in parallel, then sequential agents run in order
func (o *mixedOrchestrator) Dispatch(ctx context.Context, event Event) (AgentResult, error) {
	Logger().Info().
		Str("event_id", event.GetID()).
		Int("collaborative_agents", len(o.collaborativeAgents)).
		Int("sequential_agents", len(o.sequentialAgents)).
		Msg("MixedOrchestrator: Starting hybrid dispatch")

	// Get initial state from event data
	combinedState := NewState()

	// Copy event data into state
	for key, value := range event.GetData() {
		combinedState.Set(key, value)
	}

	// Phase 1: Execute collaborative agents in parallel
	if len(o.collaborativeAgents) > 0 {
		Logger().Info().
			Int("agents", len(o.collaborativeAgents)).
			Msg("MixedOrchestrator: Phase 1 - Collaborative execution")

		collabResult, err := o.executeCollaborativePhase(ctx, event, combinedState)
		if err != nil {
			return AgentResult{}, fmt.Errorf("collaborative phase failed: %w", err)
		}

		// Merge collaborative results into combined state
		for _, key := range collabResult.OutputState.Keys() {
			if value, ok := collabResult.OutputState.Get(key); ok {
				combinedState.Set(key, value)
			}
		}
		for _, key := range collabResult.OutputState.MetaKeys() {
			if value, ok := collabResult.OutputState.GetMeta(key); ok {
				combinedState.SetMeta(key, value)
			}
		}
	}

	// Phase 2: Execute sequential agents in order
	if len(o.sequentialAgents) > 0 {
		Logger().Info().
			Int("agents", len(o.sequentialAgents)).
			Msg("MixedOrchestrator: Phase 2 - Sequential execution")

		seqResult, err := o.executeSequentialPhase(ctx, event, combinedState)
		if err != nil {
			return AgentResult{}, fmt.Errorf("sequential phase failed: %w", err)
		}

		// Merge sequential result into combined state
		for _, key := range seqResult.OutputState.Keys() {
			if value, ok := seqResult.OutputState.Get(key); ok {
				combinedState.Set(key, value)
			}
		}
		for _, key := range seqResult.OutputState.MetaKeys() {
			if value, ok := seqResult.OutputState.GetMeta(key); ok {
				combinedState.SetMeta(key, value)
			}
		}
	}

	Logger().Info().
		Str("event_id", event.GetID()).
		Msg("MixedOrchestrator: Hybrid dispatch completed successfully")

	return AgentResult{
		OutputState: combinedState,
		StartTime:   time.Now(),
		EndTime:     time.Now(),
	}, nil
}

// executeCollaborativePhase runs all collaborative agents in parallel
func (o *mixedOrchestrator) executeCollaborativePhase(ctx context.Context, event Event, state State) (AgentResult, error) {
	if len(o.collaborativeAgents) == 0 {
		return AgentResult{OutputState: state}, nil
	}

	var wg sync.WaitGroup
	resultChan := make(chan AgentResult, len(o.collaborativeAgents))

	// Execute all collaborative agents in parallel
	for name, handler := range o.collaborativeAgents {
		wg.Add(1)
		go func(agentName string, h AgentHandler) {
			defer wg.Done()

			result, err := h.Run(ctx, event, state)
			if err != nil {
				result = AgentResult{
					OutputState: state,
					Error:       fmt.Sprintf("Agent %s failed: %v", agentName, err),
				}
			}

			resultChan <- result
		}(name, handler)
	}

	wg.Wait()
	close(resultChan)

	// Collect and merge results
	combinedState := NewState()
	var errors []string
	hasSuccess := false

	for result := range resultChan {
		if result.Error != "" {
			errors = append(errors, result.Error)
		} else {
			hasSuccess = true
			// Merge successful results
			for _, key := range result.OutputState.Keys() {
				if value, ok := result.OutputState.Get(key); ok {
					combinedState.Set(key, value)
				}
			}
			for _, key := range result.OutputState.MetaKeys() {
				if value, ok := result.OutputState.GetMeta(key); ok {
					combinedState.SetMeta(key, value)
				}
			}
		}
	}

	if !hasSuccess && len(errors) > 0 {
		return AgentResult{}, fmt.Errorf("all collaborative agents failed: %v", errors)
	}

	return AgentResult{OutputState: combinedState}, nil
}

// executeSequentialPhase runs sequential agents one after another
func (o *mixedOrchestrator) executeSequentialPhase(ctx context.Context, event Event, initialState State) (AgentResult, error) {
	if len(o.sequentialAgents) == 0 {
		return AgentResult{OutputState: initialState}, nil
	}

	currentState := initialState // State interface, not *SimpleState

	for i, agentName := range o.sequentialAgents {
		o.mu.RLock()
		handler, exists := o.handlers[agentName]
		o.mu.RUnlock()

		if !exists {
			Logger().Warn().
				Str("agent", agentName).
				Int("position", i).
				Msg("MixedOrchestrator: Sequential agent not found, skipping")
			continue
		}

		Logger().Debug().
			Str("agent", agentName).
			Int("position", i).
			Int("total", len(o.sequentialAgents)).
			Msg("MixedOrchestrator: Executing sequential agent")

		result, err := handler.Run(ctx, event, currentState)
		if err != nil {
			return AgentResult{}, fmt.Errorf("sequential agent %s (position %d) failed: %w", agentName, i, err)
		}

		// Use this agent's output as input for the next agent
		currentState = result.OutputState
	}

	return AgentResult{OutputState: currentState}, nil
}

// Helper methods
func (o *mixedOrchestrator) getCollaborativeAgentNames() []string {
	return o.collaborativeAgentNames
}

func (o *mixedOrchestrator) isSequentialAgent(name string) bool {
	for _, seqName := range o.sequentialAgents {
		if seqName == name {
			return true
		}
	}
	return false
}

// GetCallbackRegistry returns the callback registry for this orchestrator
func (o *mixedOrchestrator) GetCallbackRegistry() *CallbackRegistry {
	return o.callbackRegistry
}

// Stop halts the mixed orchestrator (cleanup if needed)
func (o *mixedOrchestrator) Stop() {
	o.mu.Lock()
	defer o.mu.Unlock()

	// Clear handlers
	o.handlers = make(map[string]AgentHandler)
	o.collaborativeAgents = make(map[string]AgentHandler)
	o.sequentialAgents = nil

	Logger().Debug().Msg("MixedOrchestrator: Stopped and cleaned up")
}

// =============================================================================
// SEQUENTIAL ORCHESTRATOR IMPLEMENTATION
// =============================================================================

// sequentialOrchestrator implements sequential execution of agents
type sequentialOrchestrator struct {
	handlers         map[string]AgentHandler
	agentSequence    []string
	callbackRegistry *CallbackRegistry
	mu               sync.RWMutex
}

// RegisterAgent adds an agent to the sequential orchestrator
func (o *sequentialOrchestrator) RegisterAgent(name string, handler AgentHandler) error {
	o.mu.Lock()
	defer o.mu.Unlock()

	if handler == nil {
		return fmt.Errorf("handler cannot be nil for agent %s", name)
	}

	o.handlers[name] = handler
	Logger().Debug().Str("agent", name).Msg("SequentialOrchestrator: Agent registered")
	return nil
}

// Dispatch executes agents in the specified sequence
func (o *sequentialOrchestrator) Dispatch(ctx context.Context, event Event) (AgentResult, error) {
	o.mu.RLock()
	defer o.mu.RUnlock()

	if len(o.agentSequence) == 0 {
		return AgentResult{}, fmt.Errorf("no agent sequence defined")
	}

	// Initialize state from event data
	currentState := NewState()
	for key, value := range event.GetData() {
		currentState.Set(key, value)
	}
	for key, value := range event.GetMetadata() {
		currentState.SetMeta(key, value)
	}

	var state State = currentState // Use State interface for chaining

	// Execute agents in sequence
	for i, agentName := range o.agentSequence {
		handler, exists := o.handlers[agentName]
		if !exists {
			Logger().Warn().Str("agent", agentName).Msg("SequentialOrchestrator: Agent not found, skipping")
			continue
		}

		Logger().Debug().
			Str("agent", agentName).
			Int("position", i).
			Msg("SequentialOrchestrator: Executing agent")

		result, err := handler.Run(ctx, event, state)
		if err != nil {
			return AgentResult{}, fmt.Errorf("sequential agent %s failed: %w", agentName, err)
		}

		// Pass output state to next agent
		state = result.OutputState
	}

	return AgentResult{OutputState: state}, nil
}

// GetCallbackRegistry returns the callback registry
func (o *sequentialOrchestrator) GetCallbackRegistry() *CallbackRegistry {
	return o.callbackRegistry
}

// Stop halts the sequential orchestrator
func (o *sequentialOrchestrator) Stop() {
	o.mu.Lock()
	defer o.mu.Unlock()
	o.handlers = make(map[string]AgentHandler)
	Logger().Debug().Msg("SequentialOrchestrator: Stopped")
}

// =============================================================================
// LOOP ORCHESTRATOR IMPLEMENTATION
// =============================================================================

// loopOrchestrator implements loop execution with a single agent
type loopOrchestrator struct {
	handlers         map[string]AgentHandler
	agentName        string
	maxIterations    int
	callbackRegistry *CallbackRegistry
	mu               sync.RWMutex
}

// RegisterAgent adds an agent to the loop orchestrator
func (o *loopOrchestrator) RegisterAgent(name string, handler AgentHandler) error {
	o.mu.Lock()
	defer o.mu.Unlock()

	if handler == nil {
		return fmt.Errorf("handler cannot be nil for agent %s", name)
	}

	o.handlers[name] = handler
	Logger().Debug().Str("agent", name).Msg("LoopOrchestrator: Agent registered")
	return nil
}

// Dispatch executes the specified agent in a loop
func (o *loopOrchestrator) Dispatch(ctx context.Context, event Event) (AgentResult, error) {
	o.mu.RLock()
	defer o.mu.RUnlock()

	if o.agentName == "" {
		return AgentResult{}, fmt.Errorf("no agent specified for loop")
	}

	handler, exists := o.handlers[o.agentName]
	if !exists {
		return AgentResult{}, fmt.Errorf("loop agent %s not found", o.agentName)
	}

	// Initialize state from event data
	currentState := NewState()
	for key, value := range event.GetData() {
		currentState.Set(key, value)
	}
	for key, value := range event.GetMetadata() {
		currentState.SetMeta(key, value)
	}

	var state State = currentState // Use State interface for chaining

	// Execute agent in loop
	for i := 0; i < o.maxIterations; i++ {
		Logger().Debug().
			Str("agent", o.agentName).
			Int("iteration", i+1).
			Int("max_iterations", o.maxIterations).
			Msg("LoopOrchestrator: Executing agent iteration")

		result, err := handler.Run(ctx, event, state)
		if err != nil {
			return AgentResult{}, fmt.Errorf("loop agent %s (iteration %d) failed: %w", o.agentName, i+1, err)
		}

		// Check for completion signal in state
		if completed, ok := result.OutputState.Get("loop_completed"); ok {
			if completedBool, isBool := completed.(bool); isBool && completedBool {
				Logger().Info().
					Str("agent", o.agentName).
					Int("iteration", i+1).
					Msg("LoopOrchestrator: Agent signaled completion")
				return AgentResult{OutputState: result.OutputState}, nil
			}
		}

		// Pass output state to next iteration
		state = result.OutputState
	}

	Logger().Info().
		Str("agent", o.agentName).
		Int("iterations", o.maxIterations).
		Msg("LoopOrchestrator: Completed all iterations")

	return AgentResult{OutputState: state}, nil
}

// GetCallbackRegistry returns the callback registry
func (o *loopOrchestrator) GetCallbackRegistry() *CallbackRegistry {
	return o.callbackRegistry
}

// Stop halts the loop orchestrator
func (o *loopOrchestrator) Stop() {
	o.mu.Lock()
	defer o.mu.Unlock()
	o.handlers = make(map[string]AgentHandler)
	Logger().Debug().Msg("LoopOrchestrator: Stopped")
}
