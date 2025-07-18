package templates

const AgentTemplate = `package main

import (
	"context"
	"fmt"
	"strings"

	agentflow "github.com/kunalkushwaha/agentflow/core"
)

// {{.Agent.DisplayName}}Handler represents the {{.Agent.Name}} agent handler
// Purpose: {{.Agent.Purpose}}
type {{.Agent.DisplayName}}Handler struct {
	llm agentflow.ModelProvider
}

// New{{.Agent.DisplayName}} creates a new {{.Agent.DisplayName}} instance
func New{{.Agent.DisplayName}}(llmProvider agentflow.ModelProvider) *{{.Agent.DisplayName}}Handler {
	return &{{.Agent.DisplayName}}Handler{llm: llmProvider}
}

// Run implements the agentflow.AgentHandler interface
func (a *{{.Agent.DisplayName}}Handler) Run(ctx context.Context, event agentflow.Event, state agentflow.State) (agentflow.AgentResult, error) {
	// Get logger for debug output
	logger := agentflow.Logger()
	logger.Debug().Str("agent", "{{.Agent.Name}}").Str("event_id", event.GetID()).Msg("Agent processing started")
	
	var inputToProcess interface{}
	var systemPrompt string
	
	{{if .IsFirstAgent}}
	// {{.Agent.DisplayName}} always processes the original input message
	eventData := event.GetData()
	if msg, ok := eventData["message"]; ok {
		inputToProcess = msg
	} else if stateMessage, exists := state.Get("message"); exists {
		inputToProcess = stateMessage
	} else {
		inputToProcess = "No message provided"
	}
	
	systemPrompt = ` + "`{{.SystemPrompt}}`" + `
	logger.Debug().Str("agent", "{{.Agent.Name}}").Interface("input", inputToProcess).Msg("Processing original message")
	{{else}}
	// Sequential processing: Use previous agent's output, with fallback chain
	found := false
	agents := []string{{"{"}}{{range $i, $agent := .Agents}}{{if gt $i 0}}, {{end}}"{{$agent.Name}}"{{end}}{{"}"}}
	
	for i := {{.AgentIndex}} - 1; i >= 0; i-- {
		if i < len(agents) {
			prevAgentName := agents[i]
			if agentResponse, exists := state.Get(fmt.Sprintf("%s_response", prevAgentName)); exists {
				inputToProcess = agentResponse
				logger.Debug().Str("agent", "{{.Agent.Name}}").Str("source_agent", prevAgentName).Interface("input", agentResponse).Msg("Processing previous agent's output")
				found = true
				break
			}
		}
	}
	
	if !found {
		// Final fallback to original message
		eventData := event.GetData()
		if msg, ok := eventData["message"]; ok {
			inputToProcess = msg
		} else if stateMessage, exists := state.Get("message"); exists {
			inputToProcess = stateMessage
		} else {
			inputToProcess = "No message provided"
		}
		logger.Debug().Str("agent", "{{.Agent.Name}}").Interface("input", inputToProcess).Msg("Processing original message (final fallback)")
	}
	
	systemPrompt = ` + "`{{.SystemPrompt}}`" + `
	{{end}}
	
	// Get available MCP tools to include in prompt
	var toolsPrompt string
	mcpManager := agentflow.GetMCPManager()
	if mcpManager != nil {
		availableTools := mcpManager.GetAvailableTools()
		logger.Debug().Str("agent", "{{.Agent.Name}}").Int("tool_count", len(availableTools)).Msg("MCP Tools discovered")
		toolsPrompt = agentflow.FormatToolsPromptForLLM(availableTools)
	} else {
		logger.Warn().Str("agent", "{{.Agent.Name}}").Msg("MCP Manager is not available")
	}
	
	// Create initial LLM prompt with available tools information
	userPrompt := fmt.Sprintf("User query: %v", inputToProcess)
	userPrompt += toolsPrompt
	
	prompt := agentflow.Prompt{
		System: systemPrompt,
		User:   userPrompt,
	}
	
	// Debug: Log the full prompt being sent to LLM
	logger.Debug().Str("agent", "{{.Agent.Name}}").Str("system_prompt", systemPrompt).Str("user_prompt", userPrompt).Msg("Full LLM prompt")
	
	// Call LLM to get initial response and potential tool calls
	response, err := a.llm.Call(ctx, prompt)
	if err != nil {
		return agentflow.AgentResult{}, fmt.Errorf("{{.Agent.DisplayName}} LLM call failed: %w", err)
	}
	
	logger.Debug().Str("agent", "{{.Agent.Name}}").Str("response", response.Content).Msg("Initial LLM response received")
	
	// Parse LLM response for tool calls using core function
	toolCalls := agentflow.ParseLLMToolCalls(response.Content)
	var mcpResults []string
	
	// Debug: Log the LLM response to see tool call format
	logger.Debug().Str("agent", "{{.Agent.Name}}").Str("llm_response", response.Content).Msg("LLM response for tool call analysis")
	logger.Debug().Str("agent", "{{.Agent.Name}}").Interface("parsed_tool_calls", toolCalls).Msg("Parsed tool calls from LLM response")
	
	// Execute any requested tools
	if len(toolCalls) > 0 && mcpManager != nil {
		logger.Info().Str("agent", "{{.Agent.Name}}").Int("tool_calls", len(toolCalls)).Msg("Executing LLM-requested tools")
		
		for _, toolCall := range toolCalls {
			if toolName, ok := toolCall["name"].(string); ok {
				var args map[string]interface{}
				if toolArgs, exists := toolCall["args"]; exists {
					if argsMap, ok := toolArgs.(map[string]interface{}); ok {
						args = argsMap
					} else {
						args = make(map[string]interface{})
					}
				} else {
					args = make(map[string]interface{})
				}
				
				logger.Info().Str("agent", "{{.Agent.Name}}").Str("tool_name", toolName).Interface("args", args).Msg("Executing tool as requested by LLM")
				
				// Execute tool using the global ExecuteMCPTool function
				result, err := agentflow.ExecuteMCPTool(ctx, toolName, args)
				if err != nil {
					logger.Error().Str("agent", "{{.Agent.Name}}").Str("tool_name", toolName).Err(err).Msg("Tool execution failed")
					mcpResults = append(mcpResults, fmt.Sprintf("Tool '%s' failed: %v", toolName, err))
				} else {
					if result.Success {
						logger.Info().Str("agent", "{{.Agent.Name}}").Str("tool_name", toolName).Msg("Tool execution successful")
						
						// Format the result content
						var resultContent string
						if len(result.Content) > 0 {
							resultContent = result.Content[0].Text
						} else {
							resultContent = "Tool executed successfully but returned no content"
						}
						
						mcpResults = append(mcpResults, fmt.Sprintf("Tool '%s' result: %s", toolName, resultContent))
					} else {
						logger.Error().Str("agent", "{{.Agent.Name}}").Str("tool_name", toolName).Msg("Tool execution was not successful")
						mcpResults = append(mcpResults, fmt.Sprintf("Tool '%s' was not successful", toolName))
					}
				}
			}
		}
	} else {
		logger.Debug().Str("agent", "{{.Agent.Name}}").Msg("No tool calls requested or MCP manager not available")
	}
	
	// Generate final response if tools were used
	var finalResponse string
	if len(mcpResults) > 0 {
		// Create enhanced prompt with tool results
		enhancedPrompt := agentflow.Prompt{
			System: systemPrompt,
			User:   fmt.Sprintf("Original query: %v\n\nTool results:\n%s\n\nPlease provide a comprehensive response incorporating these tool results:", inputToProcess, strings.Join(mcpResults, "\n")),
		}
		
		// Get final response from LLM
		finalLLMResponse, err := a.llm.Call(ctx, enhancedPrompt)
		if err != nil {
			return agentflow.AgentResult{}, fmt.Errorf("{{.Agent.DisplayName}} final LLM call failed: %w", err)
		}
		finalResponse = finalLLMResponse.Content
		logger.Info().Str("agent", "{{.Agent.Name}}").Str("final_response", finalResponse).Msg("Final response generated with tool results")
	} else {
		finalResponse = response.Content
		logger.Debug().Str("agent", "{{.Agent.Name}}").Msg("Using initial LLM response (no tools used)")
	}
	
	// Store agent response in state for potential use by subsequent agents
	outputState := agentflow.NewState()
	outputState.Set("{{.Agent.Name}}_response", finalResponse)
	outputState.Set("message", finalResponse)
	
	{{if .NextAgent}}
	// {{.RoutingComment}}
	outputState.SetMeta(agentflow.RouteMetadataKey, "{{.NextAgent}}")
	{{else}}
	// Workflow completion
	{{end}}
	
	logger.Info().Str("agent", "{{.Agent.Name}}").Msg("Agent processing completed successfully")
	
	return agentflow.AgentResult{
		OutputState: outputState,
	}, nil
}
`
