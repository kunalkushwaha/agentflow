package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

// OllamaAdapter implements the LLMAdapter interface for Ollama's API.
type OllamaAdapter struct {
	apiKey      string
	model       string
	maxTokens   int
	temperature float32
}

// NewOllamaAdapter creates a new OllamaAdapter instance.
func NewOllamaAdapter(apiKey, model string, maxTokens int, temperature float32) (*OllamaAdapter, error) {
	if apiKey == "" {
		return nil, errors.New("API key cannot be empty")
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
		apiKey:      apiKey,
		model:       model,
		maxTokens:   maxTokens,
		temperature: temperature,
	}, nil
}

// Complete sends a prompt to Ollama's API and returns the completion.
func (o *OllamaAdapter) Complete(ctx context.Context, systemPrompt string, userPrompt string) (string, error) {
	if systemPrompt == "" && userPrompt == "" {
		return "", errors.New("both systemPrompt and userPrompt cannot be empty")
	}

	// Prepare the request payload
	requestBody := map[string]interface{}{
		"model":       o.model,
		"messages":    []map[string]string{{"role": "system", "content": systemPrompt}, {"role": "user", "content": userPrompt}},
		"max_tokens":  o.maxTokens,
		"temperature": o.temperature,
		"stream":      false,
	}

	payload, err := json.Marshal(requestBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request body: %w", err)
	}

	// Make the HTTP request
	client := &http.Client{}
	req, err := http.NewRequestWithContext(ctx, "POST", "http://localhost:11434/api/chat", bytes.NewBuffer(payload))
	if err != nil {
		return "", fmt.Errorf("failed to create HTTP request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	//req.Header.Set("Authorization", "Bearer "+o.apiKey)

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		return "", fmt.Errorf("Ollama API error: %s", string(body))
	}

	var response struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	content := response.Message.Content
	if content == "" {
		return "", errors.New("response content is empty")
	}

	// Check if the content is a JSON object and parse it if necessary
	var parsedContent map[string]interface{}
	if err := json.Unmarshal([]byte(content), &parsedContent); err == nil {
		// If parsing succeeds, return the JSON as a string
		parsedJSON, err := json.Marshal(parsedContent)
		if err != nil {
			return "", fmt.Errorf("failed to marshal parsed JSON content: %w", err)
		}
		return string(parsedJSON), nil
	}

	// Otherwise, return the raw content
	return strings.TrimSpace(content), nil
}
