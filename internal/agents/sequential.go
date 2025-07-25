package agents

import (
	"context"
	"fmt"

	agenticgokit "github.com/kunalkushwaha/agenticgokit/internal/core"
)

// SequentialAgent runs a series of sub-agents one after another.
type SequentialAgent struct {
	name   string
	agents []agenticgokit.Agent
}

// Name returns the name of the sequential agent.
func (a *SequentialAgent) Name() string {
	return a.name
}

// NewSequentialAgent creates a new SequentialAgent.
// It filters out any nil agents provided in the list.
func NewSequentialAgent(name string, agents ...agenticgokit.Agent) *SequentialAgent {
	validAgents := make([]agenticgokit.Agent, 0, len(agents))
	for i, agent := range agents {
		if agent == nil {
			agenticgokit.Logger().Warn().
				Str("sequential_agent", name).
				Int("index", i).
				Msg("SequentialAgent: received a nil agent, skipping.")
			continue
		}
		validAgents = append(validAgents, agent)
	}
	return &SequentialAgent{
		agents: validAgents,
		name:   name,
	}
}

// Run executes the sequence of sub-agents.
// It iterates through the configured agents, passing state sequentially.
// Execution halts immediately if a sub-agent returns an error or if the context is cancelled.
func (s *SequentialAgent) Run(ctx context.Context, initialState agenticgokit.State) (agenticgokit.State, error) {
	if len(s.agents) == 0 {
		agenticgokit.Logger().Warn().
			Str("sequential_agent", s.name).
			Msg("SequentialAgent: No sub-agents to run.")
		return initialState, nil // Return input state if no agents
	}

	var err error
	nextState := initialState // Start with the initial state

	for i, agent := range s.agents {
		// Check for context cancellation before running each sub-agent
		select {
		case <-ctx.Done():
			agenticgokit.Logger().Warn().
				Str("sequential_agent", s.name).
				Int("agent_index", i).
				Msg("SequentialAgent: Context cancelled before running agent.")
			return nextState, fmt.Errorf("SequentialAgent '%s': context cancelled: %w", s.name, ctx.Err())
		default:
			// Context is not cancelled, proceed
		}

		// It's crucial to clone the state before passing it to the next agent
		// to prevent unintended side effects if agents modify the state concurrently
		// or if the caller reuses the initial state.
		inputState := nextState.Clone()

		// Run the sub-agent
		outputState, agentErr := agent.Run(ctx, inputState)
		if agentErr != nil {
			err = fmt.Errorf("SequentialAgent '%s': error in sub-agent %d: %w", s.name, i, agentErr)
			agenticgokit.Logger().Error().
				Str("sequential_agent", s.name).
				Int("agent_index", i).
				Err(agentErr).
				Msg("SequentialAgent: Error in sub-agent.")
			// Return the state *before* the error occurred and the error itself
			return nextState, err
		}
		// Update the state for the next iteration
		nextState = outputState
	}

	// Return the final state after all agents completed successfully
	return nextState, nil
}