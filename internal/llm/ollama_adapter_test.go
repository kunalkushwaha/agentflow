package llm

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOllamaAdapter_Complete(t *testing.T) {
	tests := []struct {
		name           string
		responseBody   string
		responseStatus int
		systemPrompt   string
		userPrompt     string
		expectedResult string
		expectedError  string
	}{
		{
			name:           "Valid prompt returns response containing expected substring",
			responseBody:   `{"message": {"content": "The answer to 2+2 is 4."}}`,
			responseStatus: http.StatusOK,
			systemPrompt:   "System message",
			userPrompt:     "What is 2+2?",
			expectedResult: "4", // Check for substring in the result
			expectedError:  "",
		},
		{
			name:           "Empty prompt is invalid",
			responseBody:   "",
			responseStatus: http.StatusBadRequest,
			systemPrompt:   "",
			userPrompt:     "",
			expectedResult: "",
			expectedError:  "both systemPrompt and userPrompt cannot be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json") // Ensure correct content type
				w.WriteHeader(tt.responseStatus)
				w.Write([]byte(tt.responseBody))
			}))
			defer server.Close()

			adapter := &OllamaAdapter{
				model:       "gemma3:latest",
				maxTokens:   100,
				temperature: 0.7,
			}

			result, err := adapter.Complete(context.Background(), tt.systemPrompt, tt.userPrompt)

			if tt.expectedError != "" {
				assert.EqualError(t, err, tt.expectedError)
			} else {
				assert.NoError(t, err)
			}

			if tt.expectedResult != "" {
				assert.Contains(t, result, tt.expectedResult)
			} else {
				assert.Equal(t, tt.expectedResult, result)
			}
		})
	}
}
