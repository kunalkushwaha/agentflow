package core

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	mathrand "math/rand"
	"net/http"
	"strings"
	"time"
)

// EmbeddingService interface for generating embeddings
type EmbeddingService interface {
	GenerateEmbedding(ctx context.Context, text string) ([]float32, error)
	GenerateEmbeddings(ctx context.Context, texts []string) ([][]float32, error)
	GetDimensions() int
}

// OpenAIEmbeddingService implements EmbeddingService using OpenAI API
type OpenAIEmbeddingService struct {
	apiKey     string
	model      string
	baseURL    string
	dimensions int
	client     *http.Client
}

// OpenAI API request/response structures
type openAIEmbeddingRequest struct {
	Input          interface{} `json:"input"`
	Model          string      `json:"model"`
	EncodingFormat string      `json:"encoding_format,omitempty"`
	Dimensions     int         `json:"dimensions,omitempty"`
}

type openAIEmbeddingResponse struct {
	Object string `json:"object"`
	Data   []struct {
		Object    string    `json:"object"`
		Embedding []float32 `json:"embedding"`
		Index     int       `json:"index"`
	} `json:"data"`
	Model string `json:"model"`
	Usage struct {
		PromptTokens int `json:"prompt_tokens"`
		TotalTokens  int `json:"total_tokens"`
	} `json:"usage"`
}

// NewOpenAIEmbeddingService creates a new OpenAI embedding service
func NewOpenAIEmbeddingService(apiKey, model string) *OpenAIEmbeddingService {
	dimensions := 1536 // Default for text-embedding-3-small
	if model == "text-embedding-3-large" {
		dimensions = 3072
	} else if model == "text-embedding-ada-002" {
		dimensions = 1536
	}

	return &OpenAIEmbeddingService{
		apiKey:     apiKey,
		model:      model,
		baseURL:    "https://api.openai.com/v1/embeddings",
		dimensions: dimensions,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// GenerateEmbedding generates a single embedding
func (s *OpenAIEmbeddingService) GenerateEmbedding(ctx context.Context, text string) ([]float32, error) {
	embeddings, err := s.GenerateEmbeddings(ctx, []string{text})
	if err != nil {
		return nil, err
	}
	if len(embeddings) == 0 {
		return nil, fmt.Errorf("no embeddings returned")
	}
	return embeddings[0], nil
}

// GenerateEmbeddings generates multiple embeddings in batch
func (s *OpenAIEmbeddingService) GenerateEmbeddings(ctx context.Context, texts []string) ([][]float32, error) {
	if len(texts) == 0 {
		return [][]float32{}, nil
	}

	// Prepare request
	request := openAIEmbeddingRequest{
		Input:          texts,
		Model:          s.model,
		EncodingFormat: "float",
	}

	// Marshal request
	requestBody, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, "POST", s.baseURL, bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+s.apiKey)

	// Execute request
	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	// Read response
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Check for errors
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API error %d: %s", resp.StatusCode, string(responseBody))
	}

	// Parse response
	var response openAIEmbeddingResponse
	if err := json.Unmarshal(responseBody, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	// Extract embeddings
	embeddings := make([][]float32, len(response.Data))
	for _, data := range response.Data {
		if data.Index >= len(embeddings) {
			return nil, fmt.Errorf("invalid embedding index %d", data.Index)
		}
		embeddings[data.Index] = data.Embedding
	}

	return embeddings, nil
}

// GetDimensions returns the embedding dimensions
func (s *OpenAIEmbeddingService) GetDimensions() int {
	return s.dimensions
}

// OllamaEmbeddingService implements EmbeddingService using Ollama API
type OllamaEmbeddingService struct {
	model      string
	baseURL    string
	dimensions int
	client     *http.Client
}

// Ollama API request/response structures
type ollamaEmbeddingRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
}

type ollamaEmbeddingResponse struct {
	Embedding []float32 `json:"embedding"`
}

// NewOllamaEmbeddingService creates a new Ollama embedding service
func NewOllamaEmbeddingService(model, baseURL string) *OllamaEmbeddingService {
	if baseURL == "" {
		baseURL = "http://localhost:11434"
	}

	dimensions := 1024 // Default for mxbai-embed-large
	if strings.Contains(model, "mxbai-embed-large") {
		dimensions = 1024
	} else if strings.Contains(model, "nomic-embed") {
		dimensions = 768
	}

	return &OllamaEmbeddingService{
		model:      model,
		baseURL:    baseURL,
		dimensions: dimensions,
		client: &http.Client{
			Timeout: 60 * time.Second, // Ollama can be slower
		},
	}
}

// GenerateEmbedding generates a single embedding
func (s *OllamaEmbeddingService) GenerateEmbedding(ctx context.Context, text string) ([]float32, error) {
	// Prepare request
	request := ollamaEmbeddingRequest{
		Model:  s.model,
		Prompt: text,
	}

	// Marshal request
	requestBody, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	httpReq, err := http.NewRequestWithContext(ctx, "POST", s.baseURL+"/api/embeddings", bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	// Send request
	resp, err := s.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Read response
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(responseBody))
	}

	// Parse response
	var response ollamaEmbeddingResponse
	if err := json.Unmarshal(responseBody, &response); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if len(response.Embedding) == 0 {
		return nil, fmt.Errorf("no embedding returned from Ollama")
	}

	return response.Embedding, nil
}

// GenerateEmbeddings generates multiple embeddings in batch (sequential for Ollama)
func (s *OllamaEmbeddingService) GenerateEmbeddings(ctx context.Context, texts []string) ([][]float32, error) {
	if len(texts) == 0 {
		return [][]float32{}, nil
	}

	embeddings := make([][]float32, len(texts))
	for i, text := range texts {
		embedding, err := s.GenerateEmbedding(ctx, text)
		if err != nil {
			return nil, fmt.Errorf("failed to generate embedding for text %d: %w", i, err)
		}
		embeddings[i] = embedding
	}

	return embeddings, nil
}

// GetDimensions returns the embedding dimensions
func (s *OllamaEmbeddingService) GetDimensions() int {
	return s.dimensions
}

// DummyEmbeddingService for testing and development
type DummyEmbeddingService struct {
	dimensions int
}

// NewDummyEmbeddingService creates a dummy embedding service for testing
func NewDummyEmbeddingService(dimensions int) *DummyEmbeddingService {
	if dimensions <= 0 {
		dimensions = 1536 // Default
	}
	return &DummyEmbeddingService{dimensions: dimensions}
}

// GenerateEmbedding generates a dummy embedding
func (s *DummyEmbeddingService) GenerateEmbedding(ctx context.Context, text string) ([]float32, error) {
	embedding := make([]float32, s.dimensions)
	// Use text hash for consistent dummy embeddings
	textHash := simpleHash(text)
	mathrand.Seed(int64(textHash))
	for i := range embedding {
		embedding[i] = mathrand.Float32()*2 - 1 // Random value between -1 and 1
	}
	return embedding, nil
}

// GenerateEmbeddings generates multiple dummy embeddings
func (s *DummyEmbeddingService) GenerateEmbeddings(ctx context.Context, texts []string) ([][]float32, error) {
	embeddings := make([][]float32, len(texts))
	for idx, text := range texts {
		embedding, err := s.GenerateEmbedding(ctx, text)
		if err != nil {
			return nil, err
		}
		embeddings[idx] = embedding
	}
	return embeddings, nil
}

// GetDimensions returns the embedding dimensions
func (s *DummyEmbeddingService) GetDimensions() int {
	return s.dimensions
}

// Simple hash function for consistent dummy embeddings
func simpleHash(s string) uint32 {
	var hash uint32 = 5381
	for _, c := range s {
		hash = ((hash << 5) + hash) + uint32(c)
	}
	return hash
}
