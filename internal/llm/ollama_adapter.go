package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

// OllamaAdapter implements the LLMAdapter interface for Ollama's API.
type OllamaAdapter struct {
	baseURL     string
	model       string
	maxTokens   int
	temperature float32
}

// NewOllamaAdapter creates a new OllamaAdapter instance.
// baseURL should include scheme and host, e.g. http://localhost:11434
func NewOllamaAdapter(baseURL, model string, maxTokens int, temperature float32) (*OllamaAdapter, error) {
	if baseURL == "" {
		baseURL = "http://localhost:11434"
	}
	if model == "" {
		model = "gemma3:latest" // Replace with Ollama's default model
	}
	if maxTokens == 0 {
		maxTokens = 150 // Default max tokens
	}
	if temperature == 0 {
		temperature = 0.7 // Default temperature
	}

	return &OllamaAdapter{
		baseURL:     baseURL,
		model:       model,
		maxTokens:   maxTokens,
		temperature: temperature,
	}, nil
}

// Call implements the ModelProvider interface for a single request/response.
func (o *OllamaAdapter) Call(ctx context.Context, prompt Prompt) (Response, error) {
	if prompt.System == "" && prompt.User == "" {
		return Response{}, errors.New("both system and user prompts cannot be empty")
	}

	// Determine final parameters, preferring explicit prompt settings
	var finalMaxTokens int
if prompt.Parameters.MaxTokens != nil && *prompt.Parameters.MaxTokens > 0 {
	finalMaxTokens = int(*prompt.Parameters.MaxTokens)
} else {
	finalMaxTokens = o.maxTokens
}

var finalTemperature float32
if prompt.Parameters.Temperature != nil && *prompt.Parameters.Temperature > 0 {
	finalTemperature = *prompt.Parameters.Temperature
} else {
	finalTemperature = o.temperature
}

	// Prepare the request payload
	requestBody := map[string]interface{}{
		"model":       o.model,
		"messages":    []map[string]string{{"role": "system", "content": prompt.System}, {"role": "user", "content": prompt.User}},
		"max_tokens":  finalMaxTokens,
		"temperature": finalTemperature,
		"stream":      false,
	}

	payload, err := json.Marshal(requestBody)
	if err != nil {
		return Response{}, fmt.Errorf("failed to marshal request body: %w", err)
	}
	// Make the HTTP request with timeout
	client := &http.Client{
		Timeout: 30 * time.Second, // Add 30 second timeout
	}
	req, err := http.NewRequestWithContext(ctx, "POST", fmt.Sprintf("%s/api/chat", o.baseURL), bytes.NewBuffer(payload))
	if err != nil {
		return Response{}, fmt.Errorf("failed to create HTTP request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return Response{}, fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		return Response{}, fmt.Errorf("Ollama API error: %s", string(body))
	}

	var apiResp struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return Response{}, fmt.Errorf("failed to decode response: %w", err)
	}

	return Response{
		Content: apiResp.Message.Content,
	}, nil
}

// Stream implements the ModelProvider interface for streaming responses.
func (o *OllamaAdapter) Stream(ctx context.Context, prompt Prompt) (<-chan Token, error) {
	return nil, errors.New("streaming not supported by OllamaAdapter")
}

// Embeddings implements the ModelProvider interface for generating embeddings.
func (o *OllamaAdapter) Embeddings(ctx context.Context, texts []string) ([][]float64, error) {
	return nil, errors.New("embeddings not supported by OllamaAdapter")
}
