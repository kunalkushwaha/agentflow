package scaffold

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// CreateAgentProject creates a new AgentFlow project scaffold.
func CreateAgentProject(agentName string, numAgents int, responsibleAI bool, errorHandler bool, provider string) error {
	// Create the main project directory
	if err := os.Mkdir(agentName, 0755); err != nil {
		return fmt.Errorf("failed to create project directory %s: %w", agentName, err)
	}
	fmt.Printf("Created directory: %s\n", agentName)

	// Create go.mod file
	goModContent := fmt.Sprintf("module %s\n\ngo 1.21\n\nrequire github.com/kunalkushwaha/agentflow v0.1.1\n", agentName)
	goModPath := filepath.Join(agentName, "go.mod")
	if err := os.WriteFile(goModPath, []byte(goModContent), 0644); err != nil {
		return fmt.Errorf("failed to create go.mod: %w", err)
	}
	fmt.Printf("Created file: %s\n", goModPath)

	// Create README.md file
	readmeContent := createReadmeContent(agentName, numAgents, responsibleAI, errorHandler, provider)
	readmePath := filepath.Join(agentName, "README.md")
	if err := os.WriteFile(readmePath, []byte(readmeContent), 0644); err != nil {
		return fmt.Errorf("failed to create README.md: %w", err)
	}
	fmt.Printf("Created file: %s\n", readmePath)

	// Create main.go file with provider-specific configuration
	mainGoContent := createMainGoContent(agentName, provider, numAgents, responsibleAI, errorHandler)
	mainGoPath := filepath.Join(agentName, "main.go")
	if err := os.WriteFile(mainGoPath, []byte(mainGoContent), 0644); err != nil {
		return fmt.Errorf("failed to create main.go: %w", err)
	}
	fmt.Printf("Created file: %s\n", mainGoPath)
	// Create agent files
	if numAgents == 1 {
		if err := createAgentFile(agentName, "agent.go", 1, numAgents, responsibleAI, errorHandler); err != nil {
			return err
		}
	} else {
		// For multiple agents, create all agents in separate files in the main directory
		for i := 1; i <= numAgents; i++ {
			filename := fmt.Sprintf("agent%d.go", i)
			if err := createAgentFile(agentName, filename, i, numAgents, responsibleAI, errorHandler); err != nil {
				return err
			}
		}
	}
	// Create error handler agent if requested
	if errorHandler {
		if err := createErrorHandlerAgent(agentName); err != nil {
			return err
		}
		// Also create specialized error handlers
		if err := createSpecializedErrorHandlers(agentName); err != nil {
			return err
		}
	}
	// Create responsible AI agent if requested
	if responsibleAI {
		if err := createResponsibleAIAgent(agentName); err != nil {
			return err
		}
	}

	// Always create workflow finalizer for proper completion detection
	if err := createWorkflowFinalizerAgent(agentName); err != nil {
		return err
	}

	// Create workflows directory
	workflowsDir := filepath.Join(agentName, "workflows")
	if err := os.Mkdir(workflowsDir, 0755); err != nil {
		return fmt.Errorf("failed to create workflows directory: %w", err)
	}
	fmt.Printf("Created directory: %s\n", workflowsDir)

	// Create workflow file
	workflowContent := createWorkflowContent(numAgents, responsibleAI, errorHandler)
	workflowPath := filepath.Join(workflowsDir, "main.workflow.md")
	if err := os.WriteFile(workflowPath, []byte(workflowContent), 0644); err != nil {
		return fmt.Errorf("failed to create workflow file: %w", err)
	}
	fmt.Printf("Created file: %s\n", workflowPath) // Create agentflow.toml config file
	configContent := createConfigContent(provider, errorHandler)
	configPath := filepath.Join(agentName, "agentflow.toml")
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		return fmt.Errorf("failed to create config file: %w", err)
	}
	fmt.Printf("Created file: %s\n", configPath)

	// Create workflow finalizer agent
	if err := createWorkflowFinalizerAgent(agentName); err != nil {
		return err
	}

	return nil
}

func createAgentFile(dir, filename string, agentNum int, totalAgents int, hasRAI bool, hasErrorHandler bool) error {
	// Determine next agent in the workflow chain
	var nextAgent string
	var routingComment string

	if agentNum < totalAgents {
		// Route to next numbered agent
		nextAgent = fmt.Sprintf("agent%d", agentNum+1)
		routingComment = fmt.Sprintf("// Route to the next agent (agent%d) in the workflow", agentNum+1)
	} else if hasRAI {
		// Last agent routes to responsible AI
		nextAgent = "responsible_ai"
		routingComment = "// Route to Responsible AI for final content check"
	} else {
		// Route to workflow finalizer to complete the workflow
		nextAgent = "workflow_finalizer"
		routingComment = "// Route to workflow finalizer to complete the workflow"
	}

	// Create routing code based on next agent
	var routingCode string
	if nextAgent != "" {
		routingCode = fmt.Sprintf(`
	%s
	outputState.SetMeta(agentflow.RouteMetadataKey, "%s")`, routingComment, nextAgent)
	} else {
		routingCode = fmt.Sprintf(`
	%s`, routingComment)
	}

	// Build content dynamically to avoid template escaping issues
	var content strings.Builder

	content.WriteString("package main\n\n")
	content.WriteString("import (\n")
	content.WriteString("\t\"context\"\n")
	content.WriteString("\t\"fmt\"\n\n")
	content.WriteString("\tagentflow \"github.com/kunalkushwaha/agentflow/core\"\n")
	content.WriteString(")\n\n")

	// Type definition
	content.WriteString(fmt.Sprintf("// Agent%dHandler represents the %d agent handler\n", agentNum, agentNum))
	content.WriteString(fmt.Sprintf("type Agent%dHandler struct {\n", agentNum))
	content.WriteString("\tllm agentflow.ModelProvider\n")
	content.WriteString("}\n\n")

	// Constructor
	content.WriteString(fmt.Sprintf("// NewAgent%d creates a new Agent%d instance\n", agentNum, agentNum))
	content.WriteString(fmt.Sprintf("func NewAgent%d(llmProvider agentflow.ModelProvider) *Agent%dHandler {\n", agentNum, agentNum))
	content.WriteString(fmt.Sprintf("\treturn &Agent%dHandler{llm: llmProvider}\n", agentNum))
	content.WriteString("}\n\n")
	// Run method
	content.WriteString("// Run implements the agentflow.AgentHandler interface\n")
	content.WriteString(fmt.Sprintf("func (a *Agent%dHandler) Run(ctx context.Context, event agentflow.Event, state agentflow.State) (agentflow.AgentResult, error) {\n", agentNum))
	content.WriteString("\t// Get logger for debug output\n")
	content.WriteString("\tlogger := agentflow.Logger()\n")
	content.WriteString(fmt.Sprintf("\tlogger.Debug().Str(\"agent\", \"agent%d\").Str(\"event_id\", event.GetID()).Msg(\"Agent processing started\")\n", agentNum))
	content.WriteString("\t\n")
	content.WriteString("\tvar inputToProcess interface{}\n")
	content.WriteString("\tvar systemPrompt string\n")
	content.WriteString("\t\n")

	if agentNum == 1 {
		// Agent1 logic
		content.WriteString("\t// Agent1 always processes the original input message\n")
		content.WriteString("\teventData := event.GetData()\n")
		content.WriteString("\tif msg, ok := eventData[\"message\"]; ok {\n")
		content.WriteString("\t\tinputToProcess = msg\n")
		content.WriteString("\t} else if stateMessage, exists := state.Get(\"message\"); exists {\n")
		content.WriteString("\t\tinputToProcess = stateMessage\n")
		content.WriteString("\t} else {\n")
		content.WriteString("\t\tinputToProcess = \"No message provided\"\n")
		content.WriteString("\t}\n")
		content.WriteString("\tsystemPrompt = \"You are Agent1, the first agent in a processing chain. Analyze and provide an initial response to the user input. Your output will be processed by subsequent agents.\"\n")
		content.WriteString(fmt.Sprintf("\tlogger.Debug().Str(\"agent\", \"agent%d\").Interface(\"input\", inputToProcess).Msg(\"Processing original message\")\n", agentNum))
	} else {
		// Sequential processing logic for other agents
		content.WriteString("\t// Sequential processing: Use previous agent's output, with fallback chain\n")
		content.WriteString("\tfound := false\n")
		content.WriteString(fmt.Sprintf("\tfor i := %d; i >= 1; i-- {\n", agentNum-1))
		content.WriteString("\t\tif agentResponse, exists := state.Get(fmt.Sprintf(\"agent%d_response\", i)); exists {\n")
		content.WriteString("\t\t\tinputToProcess = agentResponse\n")
		content.WriteString(fmt.Sprintf("\t\t\tlogger.Debug().Str(\"agent\", \"agent%d\").Int(\"source_agent\", i).Interface(\"input\", agentResponse).Msg(\"Processing previous agent's output\")\n", agentNum))
		content.WriteString("\t\t\tfound = true\n")
		content.WriteString("\t\t\tbreak\n")
		content.WriteString("\t\t}\n")
		content.WriteString("\t}\n")
		content.WriteString("\t\n")
		content.WriteString("\tif !found {\n")
		content.WriteString("\t\t// Final fallback to original message\n")
		content.WriteString("\t\teventData := event.GetData()\n")
		content.WriteString("\t\tif msg, ok := eventData[\"message\"]; ok {\n")
		content.WriteString("\t\t\tinputToProcess = msg\n")
		content.WriteString("\t\t} else if stateMessage, exists := state.Get(\"message\"); exists {\n")
		content.WriteString("\t\t\tinputToProcess = stateMessage\n")
		content.WriteString("\t\t} else {\n")
		content.WriteString("\t\t\tinputToProcess = \"No message provided\"\n")
		content.WriteString("\t\t}\n")
		content.WriteString(fmt.Sprintf("\t\tlogger.Debug().Str(\"agent\", \"agent%d\").Interface(\"input\", inputToProcess).Msg(\"Processing original message (final fallback)\")\n", agentNum))
		content.WriteString("\t}\n")
		content.WriteString("\t\n")
		content.WriteString("\t// Create specialized system prompt based on agent number\n")
		if agentNum == 2 {
			content.WriteString("\tsystemPrompt = \"You are Agent2, the second agent in a processing chain. Build upon the initial analysis from Agent1 and add your own insights and processing.\"\n")
		} else if agentNum == totalAgents {
			content.WriteString(fmt.Sprintf("\tsystemPrompt = \"You are Agent%d, the final regular agent in a processing chain before responsible AI review. Your role is to provide final synthesis, conclusions, and comprehensive output based on all previous agents' work.\"\n", agentNum))
		} else {
			content.WriteString(fmt.Sprintf("\tsystemPrompt = fmt.Sprintf(\"You are Agent%d, agent number %d in a processing chain. Your role is to build upon previous agents' work and add your own expertise and analysis.\", %d, %d)\n", agentNum, agentNum, agentNum, agentNum))
		}
	}

	content.WriteString("\t\n")
	content.WriteString("\t// Create LLM prompt\n")
	content.WriteString("\tprompt := agentflow.Prompt{\n")
	content.WriteString("\t\tSystem: systemPrompt,\n")
	content.WriteString("\t\tUser:   fmt.Sprintf(\"Previous agent's output: %v\", inputToProcess),\n")
	content.WriteString("\t}\n")
	content.WriteString("\t\n")
	content.WriteString("\t// Call LLM\n")
	content.WriteString("\tresponse, err := a.llm.Call(ctx, prompt)\n")
	content.WriteString("\tif err != nil {\n")
	content.WriteString(fmt.Sprintf("\t\treturn agentflow.AgentResult{}, fmt.Errorf(\"Agent%d LLM call failed: %%w\", err)\n", agentNum))
	content.WriteString("\t}\n")
	content.WriteString("\t\n")
	content.WriteString(fmt.Sprintf("\tlogger.Debug().Str(\"agent\", \"agent%d\").Str(\"response\", response.Content).Msg(\"LLM response received\")\n", agentNum))
	content.WriteString("\t\n")
	content.WriteString("\t// Create output state\n")
	content.WriteString("\toutputState := agentflow.NewState()\n")
	content.WriteString(fmt.Sprintf("\toutputState.Set(\"agent%d_response\", response.Content)\n", agentNum))
	content.WriteString(fmt.Sprintf("\toutputState.Set(\"processed_by\", \"agent%d\")\n", agentNum))
	content.WriteString("\t\n")
	content.WriteString("\t// Copy existing state data\n")
	content.WriteString("\tfor _, key := range state.Keys() {\n")
	content.WriteString("\t\tif value, exists := state.Get(key); exists {\n")
	content.WriteString("\t\t\toutputState.Set(key, value)\n")
	content.WriteString("\t\t}\n")
	content.WriteString("\t}")
	content.WriteString(routingCode)
	content.WriteString("\n\t\n")
	content.WriteString(fmt.Sprintf("\tlogger.Debug().Str(\"agent\", \"agent%d\").Msg(\"Agent completed processing\")\n", agentNum))
	content.WriteString("\t\n")
	content.WriteString("\treturn agentflow.AgentResult{OutputState: outputState}, nil\n")
	content.WriteString("}\n")

	filePath := filepath.Join(dir, filename)
	if err := os.WriteFile(filePath, []byte(content.String()), 0644); err != nil {
		return fmt.Errorf("failed to create %s: %w", filename, err)
	}
	fmt.Printf("Created file: %s\n", filePath)
	return nil
}

func createResponsibleAIAgent(projectDir string) error {
	content := `package main

import (
	"context"
	"fmt"

	agentflow "github.com/kunalkushwaha/agentflow/core"
)

// ResponsibleAIHandler handles AI safety and compliance checks
type ResponsibleAIHandler struct {
	llm agentflow.ModelProvider
}

// NewResponsibleAIHandler creates a new ResponsibleAIHandler
func NewResponsibleAIHandler(llmProvider agentflow.ModelProvider) *ResponsibleAIHandler {
	return &ResponsibleAIHandler{llm: llmProvider}
}

// Run implements the agentflow.AgentHandler interface
func (a *ResponsibleAIHandler) Run(ctx context.Context, event agentflow.Event, state agentflow.State) (agentflow.AgentResult, error) {
	logger := agentflow.Logger()
	logger.Debug().Str("agent", "responsible_ai").Str("event_id", event.GetID()).Msg("ResponsibleAI agent processing started")
		// Get content to check from event or state
	var content interface{}
	eventData := event.GetData()
	
	// Check for specific content keys first
	if data, ok := eventData["content"]; ok {
		content = data
	} else if stateContent, exists := state.Get("content"); exists {
		content = stateContent	} else if agent1Response, exists := state.Get("agent1_response"); exists {
		content = agent1Response
	} else if agent2Response, exists := state.Get("agent2_response"); exists {
		content = agent2Response
	} else if message, exists := state.Get("message"); exists {
		content = message
	} else {
		content = "No specific content found - checking overall state data"
	}
	
	logger.Debug().Str("agent", "responsible_ai").Interface("content", content).Msg("ResponsibleAI checking content")
	
	// Create LLM prompt for responsible AI checking
	prompt := agentflow.Prompt{
		System: "You are a responsible AI assistant. Check the given content for safety, bias, and compliance with ethical AI guidelines. Respond with 'SAFE' if content is appropriate, or 'UNSAFE: reason' if not.",
		User:   fmt.Sprintf("Content to check: %v", content),
	}
	
	// Call LLM
	response, err := a.llm.Call(ctx, prompt)
	if err != nil {
		return agentflow.AgentResult{}, fmt.Errorf("ResponsibleAI LLM call failed: %w", err)
	}
	
	logger.Debug().Str("agent", "responsible_ai").Str("result", response.Content).Msg("ResponsibleAI result generated")
	
	// Create output state
	outputState := agentflow.NewState()
	outputState.Set("rai_check_result", response.Content)
	outputState.Set("processed_by", "responsible_ai")
	// Copy existing state data
	for _, key := range state.Keys() {
		if value, exists := state.Get(key); exists {
			outputState.Set(key, value)
		}
	}
		// Route to workflow finalizer to complete the workflow
	outputState.SetMeta(agentflow.RouteMetadataKey, "workflow_finalizer")
	
	logger.Debug().Str("agent", "responsible_ai").Msg("ResponsibleAI check completed - routing to workflow finalizer")
	
	return agentflow.AgentResult{OutputState: outputState}, nil
}
`

	filePath := filepath.Join(projectDir, "responsible_ai.go")
	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to create responsible_ai.go: %w", err)
	}
	fmt.Printf("Created file: %s\n", filePath)
	return nil
}

func createErrorHandlerAgent(projectDir string) error {
	content := `package main

import (
	"context"
	"fmt"

	agentflow "github.com/kunalkushwaha/agentflow/core"
)

// ErrorHandlerAgent handles errors and fallback logic
type ErrorHandlerAgent struct {
	llm agentflow.ModelProvider
}

// NewErrorHandler creates a new ErrorHandlerAgent
func NewErrorHandler(llmProvider agentflow.ModelProvider) *ErrorHandlerAgent {
	return &ErrorHandlerAgent{llm: llmProvider}
}

// Run implements the agentflow.AgentHandler interface
func (a *ErrorHandlerAgent) Run(ctx context.Context, event agentflow.Event, state agentflow.State) (agentflow.AgentResult, error) {
	logger := agentflow.Logger()
	logger.Debug().Str("agent", "error_handler").Str("event_id", event.GetID()).Msg("Error handler processing started")
	
	// Get error information from event or state
	var errorInfo interface{}
	eventData := event.GetData()
	if err, ok := eventData["error"]; ok {
		errorInfo = err
	} else if stateError, exists := state.Get("error"); exists {
		errorInfo = stateError
	} else {
		errorInfo = "No error information available"
	}
	
	logger.Debug().Str("agent", "error_handler").Interface("error_info", errorInfo).Msg("Error handler analyzing error")
	
	// Create LLM prompt for error handling
	prompt := agentflow.Prompt{
		System: "You are an error handling assistant. Analyze the given error and provide helpful suggestions for resolution.",
		User:   fmt.Sprintf("Error to analyze: %%v", errorInfo),
	}
	
	// Call LLM
	response, err := a.llm.Call(ctx, prompt)
	if err != nil {
		return agentflow.AgentResult{}, fmt.Errorf("Error Handler LLM call failed: %%w", err)
	}
	
	logger.Debug().Str("agent", "error_handler").Str("analysis", response.Content).Msg("Error handler analysis completed")
	
	// Create output state
	outputState := agentflow.NewState()
	outputState.Set("error_analysis", response.Content)
	outputState.Set("processed_by", "error_handler")
	
	// Copy existing state data
	for _, key := range state.Keys() {
		if value, exists := state.Get(key); exists {
			outputState.Set(key, value)
	}
	}
	
	logger.Debug().Str("agent", "error_handler").Msg("Error handling completed")
	
	return agentflow.AgentResult{OutputState: outputState}, nil
}
`
	filePath := filepath.Join(projectDir, "error_handler.go")
	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to create error_handler.go: %w", err)
	}
	fmt.Printf("Created file: %s\n", filePath)
	return nil
}

func createSpecializedErrorHandlers(projectDir string) error {
	// Create validation error handler with simple retry logic
	validationErrorContent := `package main

import (
	"context"
	"fmt"
	"time"

	agentflow "github.com/kunalkushwaha/agentflow/core"
)

// ValidationErrorHandler handles validation errors with simple retry logic
type ValidationErrorHandler struct {
	llm agentflow.ModelProvider
	maxRetries int
	retryDelay time.Duration
}

// NewValidationErrorHandler creates a new ValidationErrorHandler
func NewValidationErrorHandler(llmProvider agentflow.ModelProvider) *ValidationErrorHandler {
	return &ValidationErrorHandler{
		llm: llmProvider,
		maxRetries: 2,
		retryDelay: time.Second,
	}
}

// Run implements the agentflow.AgentHandler interface
func (a *ValidationErrorHandler) Run(ctx context.Context, event agentflow.Event, state agentflow.State) (agentflow.AgentResult, error) {
	logger := agentflow.Logger()
	logger.Debug().Str("agent", "validation_error_handler").Str("event_id", event.GetID()).Msg("Validation error handler processing started")
	
	// Extract error data
	eventData := event.GetData()
	var errorData map[string]interface{}
	if data, ok := eventData["error_data"].(map[string]interface{}); ok {
		errorData = data
	}
	
	var errorAnalysis string
	var err error
	
	// Simple retry logic for LLM calls
	for attempt := 0; attempt <= a.maxRetries; attempt++ {
		if attempt > 0 {
			time.Sleep(a.retryDelay * time.Duration(attempt))
		}
		
		// Create validation error analysis prompt
		prompt := agentflow.Prompt{
			System: "You are a validation error specialist. Analyze validation errors and provide specific correction guidance.",
			User:   fmt.Sprintf("Validation error details: %v. Provide clear steps to fix this validation issue.", errorData),
		}
				// Call LLM for validation analysis
		response, callErr := a.llm.Call(ctx, prompt)
		if callErr == nil {
			errorAnalysis = response.Content
			err = nil
			break
		}
		err = callErr
		logger.Debug().Str("agent", "validation_error_handler").Int("attempt", attempt+1).Err(callErr).Msg("Validation handler attempt failed")
	}
	
	// Create output state
	outputState := agentflow.NewState()
		if err != nil {
		// Use fallback response if all retries failed
		logger.Debug().Str("agent", "validation_error_handler").Err(err).Msg("Validation handler using fallback")
		outputState.Set("validation_fix_suggestions", "Unable to generate specific suggestions due to service issues. Please check input format and try again.")
		outputState.Set("recovery_action", "manual_review_required")
		outputState.Set("fallback_used", true)
	} else {
		outputState.Set("validation_fix_suggestions", errorAnalysis)
		outputState.Set("recovery_action", "retry_with_corrections")
		outputState.Set("fallback_used", false)
	}
		outputState.Set("processed_by", "validation_error_handler")
	outputState.Set("error_category", "validation")
	
	// Copy existing state
	for _, key := range state.Keys() {
		if value, exists := state.Get(key); exists {
			outputState.Set(key, value)
		}
	}
	
	logger.Debug().Str("agent", "validation_error_handler").Msg("Validation error handling completed")
	return agentflow.AgentResult{OutputState: outputState}, nil
}
`
	// Create timeout error handler with backoff strategy
	timeoutErrorContent := `package main

import (
	"context"
	"fmt"
	"time"

	agentflow "github.com/kunalkushwaha/agentflow/core"
)

// TimeoutErrorHandler handles timeout errors with exponential backoff
type TimeoutErrorHandler struct {
	llm agentflow.ModelProvider
	maxRetries int
	baseDelay time.Duration
}

// NewTimeoutErrorHandler creates a new TimeoutErrorHandler
func NewTimeoutErrorHandler(llmProvider agentflow.ModelProvider) *TimeoutErrorHandler {
	return &TimeoutErrorHandler{
		llm: llmProvider,
		maxRetries: 1, // Conservative retries for timeout scenarios
		baseDelay: 2 * time.Second,
	}
}

// Run implements the agentflow.AgentHandler interface
func (a *TimeoutErrorHandler) Run(ctx context.Context, event agentflow.Event, state agentflow.State) (agentflow.AgentResult, error) {
	logger := agentflow.Logger()
	logger.Debug().Str("agent", "timeout_error_handler").Str("event_id", event.GetID()).Msg("Timeout error handler processing started")
	
	// Extract timeout information
	eventData := event.GetData()
	var retryCount int
	if count, ok := eventData["retry_count"].(int); ok {
		retryCount = count
	}
	
	var errorData map[string]interface{}
	if data, ok := eventData["error_data"].(map[string]interface{}); ok {
		errorData = data
	}
	
	var suggestions string
	var err error
	
	// Try LLM analysis with short timeout
	timeoutCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	
	prompt := agentflow.Prompt{
		System: "You are a timeout error specialist. Analyze timeout errors and suggest optimization strategies.",
		User:   fmt.Sprintf("Timeout error after %d attempts. Error details: %v. Suggest timeout optimization and recovery strategies.", retryCount, errorData),
	}
		response, err := a.llm.Call(timeoutCtx, prompt)
	if err != nil {
		logger.Debug().Str("agent", "timeout_error_handler").Err(err).Msg("Timeout handler LLM call failed")
		suggestions = "Unable to generate specific timeout optimization due to service issues. Consider increasing timeout values or reducing operation complexity."
	} else {
		suggestions = response.Content
	}
	
	// Determine retry strategy based on attempt count
	var recoveryAction string
	var retryDelay time.Duration
	
	if retryCount < 2 {
		recoveryAction = "retry_with_extended_timeout"
		retryDelay = a.baseDelay * time.Duration(retryCount+1)
	} else if retryCount < 3 {
		recoveryAction = "retry_with_optimized_timeout"
		retryDelay = a.baseDelay * time.Duration(retryCount+1)
	} else {
		recoveryAction = "escalate_to_fallback"
		retryDelay = 0
	}
	
	outputState := agentflow.NewState()
	outputState.Set("recovery_action", recoveryAction)
	outputState.Set("retry_delay", retryDelay)
	outputState.Set("retry_count", retryCount+1)
	outputState.Set("timeout_optimization_suggestions", suggestions)
	outputState.Set("processed_by", "timeout_error_handler")
	outputState.Set("error_category", "timeout")
	outputState.Set("timeout_strategy", fmt.Sprintf("Attempt %d: %s (delay: %v)", retryCount+1, recoveryAction, retryDelay))
	outputState.Set("fallback_used", err != nil)
	
	// Copy existing state
	for _, key := range state.Keys() {
		if value, exists := state.Get(key); exists {
			outputState.Set(key, value)
		}
	}
	
	logger.Debug().Str("agent", "timeout_error_handler").Str("recovery_action", recoveryAction).Msg("Timeout error handling completed")
	return agentflow.AgentResult{OutputState: outputState}, nil
}
`
	// Create critical error handler with immediate fallback
	criticalErrorContent := `package main

import (
	"context"
	"fmt"
	"log"
	"time"

	agentflow "github.com/kunalkushwaha/agentflow/core"
)

// CriticalErrorHandler handles critical system errors with immediate fallback
type CriticalErrorHandler struct {
	llm agentflow.ModelProvider
	llmTimeout time.Duration
}

// NewCriticalErrorHandler creates a new CriticalErrorHandler
func NewCriticalErrorHandler(llmProvider agentflow.ModelProvider) *CriticalErrorHandler {
	return &CriticalErrorHandler{
		llm: llmProvider,
		llmTimeout: 5 * time.Second, // Very short timeout for critical scenarios
	}
}

// Run implements the agentflow.AgentHandler interface
func (a *CriticalErrorHandler) Run(ctx context.Context, event agentflow.Event, state agentflow.State) (agentflow.AgentResult, error) {
	logger := agentflow.Logger()
	logger.Debug().Str("agent", "critical_error_handler").Str("event_id", event.GetID()).Msg("Critical error handler processing started")
	
	// Extract critical error information
	eventData := event.GetData()
	errorMsg := "Unknown critical error"
	if msg, ok := eventData["error_message"].(string); ok {
		errorMsg = msg
	}
	
	var errorData map[string]interface{}
	if data, ok := eventData["error_data"].(map[string]interface{}); ok {
		errorData = data
	}
	
	// Log critical error immediately for monitoring
	log.Printf("CRITICAL ERROR: %s", errorMsg)
	
	var errorAnalysis string
	var recommendedAction string
	
	// Attempt LLM call with very short timeout
	criticalCtx, cancel := context.WithTimeout(ctx, a.llmTimeout)
	defer cancel()
	
	prompt := agentflow.Prompt{
		System: "You are a critical error analyst. Provide immediate emergency response recommendations for critical system errors.",
		User:   fmt.Sprintf("CRITICAL ERROR: %s. Error data: %v. Provide immediate emergency response and system protection recommendations.", errorMsg, errorData),
	}
		response, err := a.llm.Call(criticalCtx, prompt)
	if err != nil {
		// Use emergency fallback immediately
		logger.Debug().Str("agent", "critical_error_handler").Err(err).Msg("Critical handler using emergency fallback")
		errorAnalysis = "LLM analysis unavailable due to service issues. Emergency fallback activated."
		recommendedAction = "immediate_system_shutdown"
	} else {
		errorAnalysis = response.Content
		recommendedAction = "guided_emergency_response"
	}
	
	// Create emergency response state
	outputState := agentflow.NewState()
	outputState.Set("recovery_action", "terminate_workflow")
	outputState.Set("processed_by", "critical_error_handler")
	outputState.Set("error_category", "critical")
	outputState.Set("alert_level", "emergency")
	outputState.Set("critical_error_logged", true)
	outputState.Set("workflow_status", "terminated_due_to_critical_error")
	outputState.Set("error_analysis", errorAnalysis)
	outputState.Set("recommended_action", recommendedAction)
	outputState.Set("emergency_timestamp", time.Now().UTC().Format(time.RFC3339))
	outputState.Set("fallback_used", err != nil)
	
	// Preserve error context for post-mortem analysis
	outputState.Set("error_context", map[string]interface{}{
		"original_error": errorMsg,
		"error_data":     errorData,
		"event_id":       event.GetID(),
	})
	
	// Copy existing state for analysis
	for _, key := range state.Keys() {
		if value, exists := state.Get(key); exists {
			outputState.Set(key, value)
		}
	}
		logger.Debug().
		Str("agent", "critical_error_handler").
		Str("action", recommendedAction).
		Str("analysis", errorAnalysis).
		Str("event_id", event.GetID()).
		Msg("Critical error handling completed")
	
	return agentflow.AgentResult{OutputState: outputState}, nil
}
`
	// Write validation error handler
	validationPath := filepath.Join(projectDir, "validation_error_handler.go")
	if err := os.WriteFile(validationPath, []byte(validationErrorContent), 0644); err != nil {
		return fmt.Errorf("failed to create validation_error_handler.go: %w", err)
	}
	fmt.Printf("Created file: %s\n", validationPath)

	// Write timeout error handler
	timeoutPath := filepath.Join(projectDir, "timeout_error_handler.go")
	if err := os.WriteFile(timeoutPath, []byte(timeoutErrorContent), 0644); err != nil {
		return fmt.Errorf("failed to create timeout_error_handler.go: %w", err)
	}
	fmt.Printf("Created file: %s\n", timeoutPath)
	// Write critical error handler
	criticalPath := filepath.Join(projectDir, "critical_error_handler.go")
	if err := os.WriteFile(criticalPath, []byte(criticalErrorContent), 0644); err != nil {
		return fmt.Errorf("failed to create critical_error_handler.go: %w", err)
	}
	fmt.Printf("Created file: %s\n", criticalPath)

	return nil
}

func createWorkflowFinalizerAgent(projectDir string) error {
	content := `package main

import (
	"context"
	"fmt"
	"sync"

	agentflow "github.com/kunalkushwaha/agentflow/core"
)

// WorkflowFinalizerHandler handles workflow completion and signals the WaitGroup
type WorkflowFinalizerHandler struct {
	wg *sync.WaitGroup
}

// NewWorkflowFinalizer creates a new WorkflowFinalizerHandler
func NewWorkflowFinalizer(wg *sync.WaitGroup) *WorkflowFinalizerHandler {
	return &WorkflowFinalizerHandler{wg: wg}
}

// Run implements the agentflow.AgentHandler interface
func (h *WorkflowFinalizerHandler) Run(ctx context.Context, event agentflow.Event, state agentflow.State) (agentflow.AgentResult, error) {
	logger := agentflow.Logger()
	logger.Debug().Str("event_id", event.GetID()).Msg("Workflow finalizer processing event")
	
	// Log the final state for debugging
	logger.Debug().Interface("state_keys", state.Keys()).Msg("Final workflow state")
	
	// Display clean final output to user
	fmt.Println("\n=== WORKFLOW RESULTS ===")
	
	// Find and display the final agent's response
	var finalResponse string
	var foundFinalResponse bool
	
	// Look for the highest numbered agent response
	for i := 10; i >= 1; i-- {
		responseKey := fmt.Sprintf("agent%d_response", i)
		if response, exists := state.Get(responseKey); exists {
			finalResponse = fmt.Sprintf("%v", response)
			foundFinalResponse = true
			logger.Debug().Str("agent", "workflow_finalizer").Int("final_agent", i).Str("response", finalResponse).Msg("Found final agent response")
			break
		}
	}
	
	// Fallback to original message if no agent responses found
	if !foundFinalResponse {
		if originalMsg, exists := state.Get("message"); exists {
			finalResponse = fmt.Sprintf("%v", originalMsg)
			logger.Debug().Str("agent", "workflow_finalizer").Interface("original_message", originalMsg).Msg("Using original message as fallback")
		}
	}
	
	// Clean user-facing output
	fmt.Println("=== WORKFLOW RESULTS ===")
	fmt.Printf("%s\n", finalResponse)
	fmt.Println("=========================")
	
	// Create final output state
	outputState := agentflow.NewState()
	outputState.Set("workflow_completed", true)
	outputState.Set("completion_time", fmt.Sprintf("%v", event.GetTimestamp()))
	
	// Copy all final results from state
	for _, key := range state.Keys() {
		if value, exists := state.Get(key); exists {
			outputState.Set(key, value)
		}
	}
	
	logger.Debug().Msg("Workflow completed successfully, signaling completion")
	
	// Signal workflow completion
	h.wg.Done()
	
	return agentflow.AgentResult{OutputState: outputState}, nil
}
`
	filePath := filepath.Join(projectDir, "workflow_finalizer.go")
	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to create workflow_finalizer.go: %w", err)
	}
	fmt.Printf("Created file: %s\n", filePath)
	return nil
}

func createReadmeContent(projectName string, numAgents int, responsibleAI bool, errorHandler bool, provider string) string {
	return fmt.Sprintf(`# %s

An AgentFlow multi-agent system with %d agents.

## Configuration

- **Provider**: %s
- **Agents**: %d
- **Responsible AI**: %t
- **Error Handler**: %t

## Setup

1. Install dependencies:
`+"```bash"+`
   go mod tidy
`+"```"+`

2. Configure your LLM provider:
   - For OpenAI: Set OPENAI_API_KEY environment variable
   - For Azure: Set AZURE_OPENAI_API_KEY, AZURE_OPENAI_ENDPOINT, AZURE_OPENAI_DEPLOYMENT
   - For Ollama: Ensure Ollama is running locally

3. Run the application:
`+"```bash"+`
   go run .
`+"```"+`

## Project Structure

- `+"`main.go`"+` - Main application entry point
- `+"`agent*.go`"+` - Individual agent implementations
- `+"`workflows/`"+` - Workflow definitions
- `+"`agentflow.toml`"+` - Configuration file

## Usage

This project implements a multi-agent system using the AgentFlow framework. Each agent can process events and maintain state throughout the workflow execution.

Generated with AgentFlow CLI v0.1.0
`, projectName, numAgents, provider, numAgents, responsibleAI, errorHandler)
}

func createMainGoContent(projectName, provider string, numAgents int, responsibleAI bool, errorHandler bool) string {
	var content strings.Builder

	// Build agents map for factory pattern
	agentMappings := ""
	for i := 1; i <= numAgents; i++ {
		agentMappings += fmt.Sprintf("\t\t\"agent%d\": NewAgent%d(llmProvider),\n", i, i)
	}
	if responsibleAI {
		agentMappings += "\t\t\"responsible_ai\": NewResponsibleAIHandler(llmProvider),\n"
	}
	if errorHandler {
		agentMappings += "\t\t\"error_handler\": NewErrorHandler(llmProvider),\n"
		// Add specialized error handlers if error handling is enabled
		agentMappings += "\t\t\"validation-error-handler\": NewValidationErrorHandler(llmProvider),\n"
		agentMappings += "\t\t\"timeout-error-handler\": NewTimeoutErrorHandler(llmProvider),\n"
		agentMappings += "\t\t\"critical-error-handler\": NewCriticalErrorHandler(llmProvider),\n"
	}
	// Add workflow finalizer for proper completion detection
	agentMappings += "\t\t\"workflow_finalizer\": NewWorkflowFinalizer(&wg),\n"

	// Determine the initial agent to send the event to.
	initialAgent := "agent1"
	if numAgents == 0 && responsibleAI {
		initialAgent = "responsible_ai"
	} else if numAgents == 0 && errorHandler {
		initialAgent = "error_handler"
	}
	content.WriteString("package main\n\n")
	content.WriteString("import (\n")
	content.WriteString("\t\"context\"\n")
	content.WriteString("\t\"flag\"\n")
	content.WriteString("\t\"fmt\"\n")
	content.WriteString("\t\"os\"\n")
	content.WriteString("\t\"sync\"\n")
	content.WriteString("\t\"time\"\n\n")
	content.WriteString("\t\"github.com/kunalkushwaha/agentflow/core\"\n")
	content.WriteString(")\n\n")
	content.WriteString("func main() {\n")
	content.WriteString("\tctx := context.Background()\n\n")
	content.WriteString("\t// Configure AgentFlow logging level\n")
	content.WriteString("\t// Options: DEBUG, INFO, WARN, ERROR\n")
	content.WriteString("\tcore.SetLogLevel(core.INFO) // Default to INFO\n\n")
	content.WriteString("\t// Optional: Get logger for custom logging\n")
	content.WriteString("\tlogger := core.Logger()\n")
	content.WriteString("\tlogger.Info().Msg(\"Starting multi-agent system...\")\n\n")
	content.WriteString("\t// Parse command line flags\n")
	content.WriteString("\tmessageFlag := flag.String(\"m\", \"\", \"Message to process by the multi-agent system\")\n")
	content.WriteString("\tflag.Parse()\n\n")
	content.WriteString("\t// Get input message from flag or interactive input\n")
	content.WriteString("\tvar inputMessage string\n")
	content.WriteString("\tif *messageFlag != \"\" {\n")
	content.WriteString("\t\t// Use message from -m flag\n")
	content.WriteString("\t\tinputMessage = *messageFlag\n")
	content.WriteString("\t\tlogger.Info().Str(\"input\", inputMessage).Msg(\"Using message from -m flag\")\n")
	content.WriteString("\t} else {\n")
	content.WriteString("\t\t// Prompt user for input if no flag provided\n")
	content.WriteString("\t\tfmt.Print(\"Enter a message for the multi-agent system: \")\n")
	content.WriteString("\t\tfmt.Scanln(&inputMessage)\n")
	content.WriteString("\t\tif inputMessage == \"\" {\n")
	content.WriteString(fmt.Sprintf("\t\t\tinputMessage = \"Hello from %s!\"\n", projectName))
	content.WriteString("\t\t\tlogger.Info().Msg(\"No input provided, using default message\")\n")
	content.WriteString("\t\t}\n")
	content.WriteString("\t}\n\n")
	content.WriteString("\t// Initialize WaitGroup for workflow completion detection\n")
	content.WriteString("\tvar wg sync.WaitGroup\n\n")
	content.WriteString("\t// Initialize LLM provider from working directory configuration\n")
	content.WriteString("\tllmProvider, err := core.NewProviderFromWorkingDir()\n")
	content.WriteString("\tif err != nil {\n")
	content.WriteString("\t\tlogger.Error().Err(err).Msg(\"Failed to initialize LLM provider from agentflow.toml\")\n")
	content.WriteString("\t\tos.Exit(1)\n")
	content.WriteString("\t}\n\n")
	content.WriteString("\t// Create agents map using the modern factory pattern\n")
	content.WriteString("\tagents := map[string]core.AgentHandler{\n")
	content.WriteString(agentMappings)
	content.WriteString("\t}\n\n")
	content.WriteString("\t// Create runner using the factory pattern - automatically wires up everything\n")
	content.WriteString("\trunner := core.NewRunnerFromWorkingDir(agents)\n\n")
	content.WriteString("\t// Start the runner\n")
	content.WriteString("\tif err := runner.Start(ctx); err != nil {\n")
	content.WriteString("\t\tlogger.Error().Err(err).Msg(\"Error starting runner\")\n")
	content.WriteString("\t\tos.Exit(1)\n")
	content.WriteString("\t}\n")
	content.WriteString("\tdefer runner.Stop()\n\n")
	content.WriteString("\t// Generate unique session ID for this workflow execution\n")
	content.WriteString("\tsessionID := \"session-\" + time.Now().Format(\"20060102-150405\")\n\n")
	content.WriteString("\t// Create an initial event with routing metadata and session ID\n")
	content.WriteString(fmt.Sprintf("\tinitialEvent := core.NewEvent(\"%s\", map[string]interface{}{\n", initialAgent))
	content.WriteString("\t\t\"message\": inputMessage,\n")
	content.WriteString("\t}, map[string]string{\n")
	content.WriteString(fmt.Sprintf("\t\tcore.RouteMetadataKey: \"%s\",\n", initialAgent))
	content.WriteString("\t\tcore.SessionIDKey:     sessionID,\n")
	content.WriteString("\t})\n\n")
	content.WriteString("\t// Emit the initial event\n")
	content.WriteString("\tlogger.Info().Str(\"session_id\", sessionID).Str(\"input\", inputMessage).Msg(\"Emitting initial event to start workflow\")\n\n")
	content.WriteString("\t// Add 1 to WaitGroup before emitting event - will be decremented by workflow_finalizer\n")
	content.WriteString("\twg.Add(1)\n\n")
	content.WriteString("\tif err := runner.Emit(initialEvent); err != nil {\n")
	content.WriteString("\t\tlogger.Error().Err(err).Msg(\"Failed to emit initial event\")\n")
	content.WriteString("\t\tos.Exit(1)\n")
	content.WriteString("\t}\n\n")
	content.WriteString("\t// Wait for workflow completion using WaitGroup pattern\n")
	content.WriteString("\t// The workflow_finalizer agent will call wg.Done() when the workflow is complete\n")
	content.WriteString("\tlogger.Info().Msg(\"Waiting for multi-agent workflow to complete...\")\n")
	content.WriteString("\twg.Wait()\n\n")
	content.WriteString("\tlogger.Info().Str(\"session_id\", sessionID).Msg(\"Workflow completed, shutting down...\")\n")
	content.WriteString("}\n")

	return content.String()
}

func createWorkflowContent(numAgents int, responsibleAI bool, errorHandler bool) string {
	workflow := `# Multi-Agent Workflow

This workflow demonstrates the interaction between multiple agents in the system.

## Workflow Diagram

`

	// Create a simple mermaid diagram
	workflow += "```mermaid\n"
	workflow += "graph TD\n"
	workflow += "    Start([Start Event]) --> A1[Agent 1]\n"

	if numAgents > 1 {
		for i := 2; i <= numAgents; i++ {
			workflow += fmt.Sprintf("    A%d[Agent %d] --> A%d[Agent %d]\n", i-1, i-1, i, i)
		}
	}

	if responsibleAI {
		workflow += fmt.Sprintf("    A%d --> RAI[Responsible AI Check]\n", numAgents)
	}

	if errorHandler {
		workflow += "    RAI --> EH[Error Handler]\n"
		workflow += "    EH --> End([End])\n"
	} else if responsibleAI {
		workflow += "    RAI --> End([End])\n"
	} else {
		workflow += fmt.Sprintf("    A%d --> End([End])\n", numAgents)
	}

	workflow += "```\n\n"

	workflow += "## Agent Descriptions\n\n"
	for i := 1; i <= numAgents; i++ {
		workflow += fmt.Sprintf("- **Agent %d**: Processes input and generates responses\n", i)
	}

	if responsibleAI {
		workflow += "- **Responsible AI**: Checks content for safety and compliance\n"
	}

	if errorHandler {
		workflow += "- **Error Handler**: Manages errors and provides fallback logic\n"
	}

	return workflow
}

func createConfigContent(provider string, errorHandler bool) string {
	var providerConfig string
	switch provider {
	case "openai":
		providerConfig = `[providers.openai]
# OpenAI provider configuration
# You can set these values here or use environment variables
# api_key = "your-openai-api-key"  # or set OPENAI_API_KEY
model = "gpt-3.5-turbo"
max_tokens = 1000
temperature = 0.7`
	case "azure":
		providerConfig = `[providers.azure]
# Azure OpenAI provider configuration
# You can set these values here or use environment variables
# api_key = "your-azure-api-key"          # or set AZURE_OPENAI_API_KEY
# endpoint = "https://your-resource.openai.azure.com"  # or set AZURE_OPENAI_ENDPOINT
# chat_deployment = "gpt-35-turbo"        # or set AZURE_OPENAI_CHAT_DEPLOYMENT
# embedding_deployment = "text-embedding-ada-002"     # or set AZURE_OPENAI_EMBEDDING_DEPLOYMENT`
	case "ollama":
		providerConfig = `[providers.ollama]
# Ollama provider configuration
# You can set these values here or use environment variables
# base_url = "http://localhost:11434"     # or set OLLAMA_BASE_URL
# model = "llama2"                        # or set OLLAMA_MODEL
max_tokens = 1000
temperature = 0.7`
	default:
		providerConfig = fmt.Sprintf(`[providers.%s]
# Provider-specific configuration
# Add your configuration here based on the selected provider`, provider)
	}

	// Add error routing configuration if error handling is enabled
	var errorRoutingConfig string
	if errorHandler {
		errorRoutingConfig = `

# Error Routing Configuration
[error_routing]
enabled = true
auto_detect_handlers = true

# Circuit Breaker Configuration
[error_routing.circuit_breaker]
failure_threshold = 3
success_threshold = 2
timeout = "30s"
reset_timeout = "60s"
half_open_max_calls = 2

# Retry Configuration
[error_routing.retry]
max_retries = 2
base_delay = "1s"
max_delay = "10s"
backoff_factor = 2.0
enable_jitter = true

# Error Handler Routing
[error_routing.handlers]
# Category-based error routing
[error_routing.handlers.category]
validation = "validation_error_handler"
timeout = "timeout_error_handler"
critical = "critical_error_handler"

# Severity-based error routing
[error_routing.handlers.severity]
high = "critical_error_handler"
medium = "timeout_error_handler"
low = "validation_error_handler"

# Default error handler
default_handler = "error_handler"`
	}

	return fmt.Sprintf(`[agent_flow]
name = "Multi-Agent System"
version = "1.0.0"
provider = "%s"

[logging]
level = "info"
format = "json"

[runtime]
max_concurrent_agents = 10
timeout_seconds = 30

%s%s
`, provider, providerConfig, errorRoutingConfig)
}
