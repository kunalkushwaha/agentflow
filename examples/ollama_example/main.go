package main

import (
	"context"
	"errors"
	"time"

	agentflow "kunalkushwaha/agentflow/internal/core"
	"kunalkushwaha/agentflow/internal/llm"
)

// OllamaAgent implements agentflow.AgentHandler
type OllamaAgent struct {
	adapter *llm.OllamaAdapter
}

func (a *OllamaAgent) Run(ctx context.Context, event agentflow.Event, state agentflow.State) (agentflow.AgentResult, error) {
	systemPrompt := "You are a helpful assistant."

	// Fetch the user prompt from the event payload
	userPrompt, ok := event.GetData()["user_prompt"].(string)
	if !ok || userPrompt == "" {
		return agentflow.AgentResult{}, errors.New("user_prompt is missing or invalid in the event payload")
	}

	response, err := a.adapter.Complete(ctx, systemPrompt, userPrompt)
	if err != nil {
		return agentflow.AgentResult{}, err
	}

	agentflow.Logger().Info().Msgf("Ollama Response: %s", response)
	return agentflow.AgentResult{OutputState: state}, nil
}

func main() {
	agentflow.SetLogLevel(agentflow.INFO)

	// Fetch the Ollama API key from the environment variable
	// apiKey := os.Getenv("OLLAMA_API_KEY")
	// if apiKey == "" {
	// 	log.Fatal("OLLAMA_API_KEY environment variable is not set")
	// }

	adapter, err := llm.NewOllamaAdapter("test-key", "gemma3:latest", 100, 0.7)
	if err != nil {
		agentflow.Logger().Error().Msgf("Failed to create OllamaAdapter: %v", err)
	}

	agent := &OllamaAgent{adapter: adapter}

	// Create a SimpleEvent with the user prompt
	event := &agentflow.SimpleEvent{
		Data: agentflow.EventData{"user_prompt": "What is the capital of France?"},
	}

	// Create a SimpleState instance
	state := agentflow.NewState()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	result, err := agent.Run(ctx, event, state)
	if err != nil {
		agentflow.Logger().Error().Msgf("Agent execution failed: %v", err)
	}

	agentflow.Logger().Debug().Msgf("Agent execution succeeded: %+v", result)

}
