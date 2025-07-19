package core

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEnhancedErrorRouting(t *testing.T) {
	runner := NewRunner(10)
	callbackRegistry := NewCallbackRegistry()
	runner.SetCallbackRegistry(callbackRegistry)
	orch := NewRouteOrchestrator(callbackRegistry)
	runner.SetOrchestrator(orch)

	// Set up enhanced error routing configuration
	errorConfig := &ErrorRouterConfig{
		MaxRetries:           2,
		RetryDelayMs:         100,
		EnableCircuitBreaker: true,
		ErrorHandlerName:     "enhanced-error-handler",
		CategoryHandlers: map[string]string{
			ErrorCodeValidation: "validation-error-handler",
			ErrorCodeLLM:        "llm-error-handler",
		},
		SeverityHandlers: map[string]string{
			SeverityCritical: "critical-error-handler",
		},
	}
	runner.SetErrorRouterConfig(errorConfig)

	// Track which error handlers were called
	var errorHandlersCalled []string
	var errorEventData *ErrorEventData
	// Register enhanced error handler
	enhancedErrorHandler := AgentHandlerFunc(func(ctx context.Context, event Event, state State) (AgentResult, error) {
		errorHandlersCalled = append(errorHandlersCalled, "enhanced-error-handler")
		// Extract error data from state instead of event data
		if data, ok := state.Get("error_data"); ok {
			if errData, ok := data.(ErrorEventData); ok {
				errorEventData = &errData
			}
		}
		return AgentResult{OutputState: state}, nil
	})
	require.NoError(t, runner.RegisterAgent("enhanced-error-handler", enhancedErrorHandler))

	// Register validation error handler
	validationErrorHandler := AgentHandlerFunc(func(ctx context.Context, event Event, state State) (AgentResult, error) {
		errorHandlersCalled = append(errorHandlersCalled, "validation-error-handler")
		// Extract error data from state
		if data, ok := state.Get("error_data"); ok {
			if errData, ok := data.(ErrorEventData); ok {
				errorEventData = &errData
			}
		}
		return AgentResult{OutputState: state}, nil
	})
	require.NoError(t, runner.RegisterAgent("validation-error-handler", validationErrorHandler))

	// Register a failing agent that produces a validation error
	failingAgent := AgentHandlerFunc(func(ctx context.Context, event Event, state State) (AgentResult, error) {
		return AgentResult{}, fmt.Errorf("validation failed: required field missing")
	})
	require.NoError(t, runner.RegisterAgent("failing-agent", failingAgent))

	require.NoError(t, runner.Start(context.Background()))
	defer runner.Stop()

	// Emit event that will trigger validation error
	event := NewEvent("test", EventData{}, map[string]string{
		RouteMetadataKey: "failing-agent",
		SessionIDKey:     "test-session",
	})
	require.NoError(t, runner.Emit(event))

	// Wait for error handling to complete
	time.Sleep(200 * time.Millisecond)

	// Verify that the validation error handler was called
	assert.Contains(t, errorHandlersCalled, "validation-error-handler", "Validation error handler should be called")

	// Verify error event data structure
	assert.NotNil(t, errorEventData, "Error event data should be populated")
	if errorEventData != nil {
		assert.Equal(t, "failing-agent", errorEventData.FailedAgent)
		assert.Equal(t, ErrorCodeValidation, errorEventData.ErrorCode)
		assert.Equal(t, SeverityMedium, errorEventData.Severity)
		assert.Equal(t, "validation", errorEventData.ErrorCategory)
		assert.Equal(t, RecoveryTerminate, errorEventData.RecoveryAction)
		assert.Equal(t, "test-session", errorEventData.SessionID)
	}
}

func TestErrorCategorization(t *testing.T) {
	tests := []struct {
		name             string
		errorMsg         string
		expectedCode     string
		expectedSeverity string
		expectedCategory string
	}{
		{
			name:             "ValidationError",
			errorMsg:         "validation failed: required field missing",
			expectedCode:     ErrorCodeValidation,
			expectedSeverity: SeverityMedium,
			expectedCategory: "validation",
		},
		{
			name:             "TimeoutError",
			errorMsg:         "context deadline exceeded",
			expectedCode:     ErrorCodeTimeout,
			expectedSeverity: SeverityHigh,
			expectedCategory: "timeout",
		},
		{
			name:             "LLMError",
			errorMsg:         "openai completion failed",
			expectedCode:     ErrorCodeLLM,
			expectedSeverity: SeverityMedium,
			expectedCategory: "llm",
		},
		{
			name:             "NetworkError",
			errorMsg:         "dial tcp: connection refused",
			expectedCode:     ErrorCodeNetwork,
			expectedSeverity: SeverityHigh,
			expectedCategory: "network",
		},
		{
			name:             "AuthError",
			errorMsg:         "unauthorized access token",
			expectedCode:     ErrorCodeAuth,
			expectedSeverity: SeverityCritical,
			expectedCategory: "auth",
		},
		{
			name:             "UnknownError",
			errorMsg:         "something went wrong",
			expectedCode:     ErrorCodeUnknown,
			expectedSeverity: SeverityMedium,
			expectedCategory: "unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := fmt.Errorf("%s", tt.errorMsg)
			code, severity, category := categorizeError(err)

			assert.Equal(t, tt.expectedCode, code, "Error code mismatch")
			assert.Equal(t, tt.expectedSeverity, severity, "Severity mismatch")
			assert.Equal(t, tt.expectedCategory, category, "Category mismatch")
		})
	}
}

func TestRecoveryActionLogic(t *testing.T) {
	tests := []struct {
		name           string
		errorCode      string
		retryCount     int
		maxRetries     int
		expectedAction string
	}{
		{
			name:           "AuthErrorNoRetry",
			errorCode:      ErrorCodeAuth,
			retryCount:     0,
			maxRetries:     3,
			expectedAction: RecoveryEscalate,
		},
		{
			name:           "ValidationErrorNoRetry",
			errorCode:      ErrorCodeValidation,
			retryCount:     0,
			maxRetries:     3,
			expectedAction: RecoveryTerminate,
		},
		{
			name:           "NetworkErrorRetry",
			errorCode:      ErrorCodeNetwork,
			retryCount:     1,
			maxRetries:     3,
			expectedAction: RecoveryRetry,
		},
		{
			name:           "NetworkErrorMaxRetries",
			errorCode:      ErrorCodeNetwork,
			retryCount:     3,
			maxRetries:     3,
			expectedAction: RecoveryFallback,
		},
		{
			name:           "UnknownErrorRetry",
			errorCode:      ErrorCodeUnknown,
			retryCount:     1,
			maxRetries:     3,
			expectedAction: RecoveryRetry,
		},
		{
			name:           "UnknownErrorMaxRetries",
			errorCode:      ErrorCodeUnknown,
			retryCount:     3,
			maxRetries:     3,
			expectedAction: RecoveryEscalate,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			action := determineRecoveryAction(tt.errorCode, tt.retryCount, tt.maxRetries)
			assert.Equal(t, tt.expectedAction, action, "Recovery action mismatch")
		})
	}
}

func TestErrorHandlerSelection(t *testing.T) {
	config := &ErrorRouterConfig{
		MaxRetries:           3,
		RetryDelayMs:         1000,
		EnableCircuitBreaker: true,
		ErrorHandlerName:     "default-error-handler",
		CategoryHandlers: map[string]string{
			ErrorCodeValidation: "validation-error-handler",
			ErrorCodeLLM:        "llm-error-handler",
		},
		SeverityHandlers: map[string]string{
			SeverityCritical: "critical-error-handler",
		},
	}

	tests := []struct {
		name            string
		errorData       ErrorEventData
		expectedHandler string
	}{
		{
			name: "SeverityHandlerPriority",
			errorData: ErrorEventData{
				ErrorCode: ErrorCodeAuth,
				Severity:  SeverityCritical,
			},
			expectedHandler: "critical-error-handler",
		},
		{
			name: "CategoryHandler",
			errorData: ErrorEventData{
				ErrorCode: ErrorCodeValidation,
				Severity:  SeverityMedium,
			},
			expectedHandler: "validation-error-handler",
		},
		{
			name: "DefaultHandler",
			errorData: ErrorEventData{
				ErrorCode: ErrorCodeUnknown,
				Severity:  SeverityMedium,
			},
			expectedHandler: "default-error-handler",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := determineErrorHandler(tt.errorData, config)
			assert.Equal(t, tt.expectedHandler, handler, "Handler selection mismatch")
		})
	}
}

func TestRetryCountIncrement(t *testing.T) {
	// Create original event
	originalEvent := NewEvent("test", EventData{}, map[string]string{
		RouteMetadataKey: "test-agent",
		SessionIDKey:     "test-session",
	})

	// Increment retry count
	retriedEvent := IncrementRetryCount(originalEvent)

	// Verify retry count was set to 1
	retryCountStr, ok := retriedEvent.GetMetadataValue("retry_count")
	assert.True(t, ok, "Retry count should be present")
	assert.Equal(t, "1", retryCountStr, "Retry count should be 1")

	// Increment again
	retriedEvent2 := IncrementRetryCount(retriedEvent)
	retryCountStr2, ok := retriedEvent2.GetMetadataValue("retry_count")
	assert.True(t, ok, "Retry count should be present")
	assert.Equal(t, "2", retryCountStr2, "Retry count should be 2")
}
