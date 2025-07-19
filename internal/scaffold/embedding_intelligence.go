package scaffold

import (
	"fmt"
	"strings"
)

// EmbeddingModelInfo contains information about a specific embedding model
type EmbeddingModelInfo struct {
	Provider    string
	Model       string
	Dimensions  int
	BaseURL     string
	Notes       string
	Recommended bool
}

// EmbeddingIntelligence provides smart configuration for embedding models
type EmbeddingIntelligence struct {
	knownModels map[string]map[string]EmbeddingModelInfo
}

// NewEmbeddingIntelligence creates a new embedding intelligence system
func NewEmbeddingIntelligence() *EmbeddingIntelligence {
	return &EmbeddingIntelligence{
		knownModels: initializeKnownModels(),
	}
}

// GetModelInfo returns information about a specific embedding model
func (ei *EmbeddingIntelligence) GetModelInfo(provider, model string) (*EmbeddingModelInfo, error) {
	provider = strings.ToLower(provider)
	
	if providerModels, exists := ei.knownModels[provider]; exists {
		if modelInfo, exists := providerModels[model]; exists {
			return &modelInfo, nil
		}
		
		// Try partial matching for common variations
		for knownModel, info := range providerModels {
			if strings.Contains(model, knownModel) || strings.Contains(knownModel, model) {
				return &info, nil
			}
		}
	}
	
	return nil, fmt.Errorf("unknown embedding model: %s/%s", provider, model)
}

// GetRecommendedModels returns recommended models for a provider
func (ei *EmbeddingIntelligence) GetRecommendedModels(provider string) []EmbeddingModelInfo {
	provider = strings.ToLower(provider)
	var recommended []EmbeddingModelInfo
	
	if providerModels, exists := ei.knownModels[provider]; exists {
		for _, info := range providerModels {
			if info.Recommended {
				recommended = append(recommended, info)
			}
		}
	}
	
	return recommended
}

// ValidateCompatibility checks if an embedding model is compatible with a memory provider
func (ei *EmbeddingIntelligence) ValidateCompatibility(provider, model, memoryProvider string) error {
	modelInfo, err := ei.GetModelInfo(provider, model)
	if err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}
	
	// Check for known incompatibilities
	if memoryProvider == "weaviate" && provider == "dummy" {
		return fmt.Errorf("dummy embeddings are not recommended with Weaviate - consider using 'openai' or 'ollama' for better search results")
	}
	
	// Validate dimensions are reasonable
	if modelInfo.Dimensions <= 0 || modelInfo.Dimensions > 4096 {
		return fmt.Errorf("embedding model %s has unusual dimensions (%d) - please verify this is correct", model, modelInfo.Dimensions)
	}
	
	return nil
}

// GetDimensionsForModel returns the dimensions for a specific model
func (ei *EmbeddingIntelligence) GetDimensionsForModel(provider, model string) (int, error) {
	modelInfo, err := ei.GetModelInfo(provider, model)
	if err != nil {
		return 0, err
	}
	return modelInfo.Dimensions, nil
}

// SuggestModel suggests an appropriate model based on provider and requirements
func (ei *EmbeddingIntelligence) SuggestModel(provider string, requirements string) (*EmbeddingModelInfo, error) {
	recommended := ei.GetRecommendedModels(provider)
	if len(recommended) == 0 {
		return nil, fmt.Errorf("no recommended models found for provider: %s", provider)
	}
	
	// For now, return the first recommended model
	// In the future, this could be more sophisticated based on requirements
	return &recommended[0], nil
}

// GetProviderDefaults returns default configuration for a provider
func (ei *EmbeddingIntelligence) GetProviderDefaults(provider string) map[string]interface{} {
	defaults := make(map[string]interface{})
	
	switch strings.ToLower(provider) {
	case "ollama":
		defaults["base_url"] = "http://localhost:11434"
		defaults["cache_embeddings"] = true
		defaults["timeout_seconds"] = 30
	case "openai":
		defaults["cache_embeddings"] = true
		defaults["max_batch_size"] = 100
		defaults["timeout_seconds"] = 30
	case "dummy":
		defaults["cache_embeddings"] = false
		defaults["timeout_seconds"] = 5
	}
	
	return defaults
}

// initializeKnownModels creates the database of known embedding models
func initializeKnownModels() map[string]map[string]EmbeddingModelInfo {
	return map[string]map[string]EmbeddingModelInfo{
		"ollama": {
			"nomic-embed-text:latest": {
				Provider:    "ollama",
				Model:       "nomic-embed-text:latest",
				Dimensions:  768,
				BaseURL:     "http://localhost:11434",
				Notes:       "Excellent general-purpose embedding model with good performance",
				Recommended: true,
			},
			"nomic-embed-text": {
				Provider:    "ollama",
				Model:       "nomic-embed-text",
				Dimensions:  768,
				BaseURL:     "http://localhost:11434",
				Notes:       "Excellent general-purpose embedding model with good performance",
				Recommended: true,
			},
			"mxbai-embed-large": {
				Provider:    "ollama",
				Model:       "mxbai-embed-large",
				Dimensions:  1024,
				BaseURL:     "http://localhost:11434",
				Notes:       "Larger model with better quality, requires more resources",
				Recommended: false,
			},
			"all-minilm": {
				Provider:    "ollama",
				Model:       "all-minilm",
				Dimensions:  384,
				BaseURL:     "http://localhost:11434",
				Notes:       "Lightweight and fast, good for development",
				Recommended: false,
			},
		},
		"openai": {
			"text-embedding-3-small": {
				Provider:    "openai",
				Model:       "text-embedding-3-small",
				Dimensions:  1536,
				Notes:       "Cost-effective OpenAI embedding model with good performance",
				Recommended: true,
			},
			"text-embedding-3-large": {
				Provider:    "openai",
				Model:       "text-embedding-3-large",
				Dimensions:  3072,
				Notes:       "Highest quality OpenAI embedding model, more expensive",
				Recommended: false,
			},
			"text-embedding-ada-002": {
				Provider:    "openai",
				Model:       "text-embedding-ada-002",
				Dimensions:  1536,
				Notes:       "Legacy OpenAI model, use text-embedding-3-small instead",
				Recommended: false,
			},
		},
		"dummy": {
			"dummy": {
				Provider:    "dummy",
				Model:       "dummy",
				Dimensions:  1536,
				Notes:       "Simple embeddings for testing only, not suitable for production",
				Recommended: false,
			},
		},
	}
}

// Global instance for easy access
var EmbeddingIntel = NewEmbeddingIntelligence()

// Helper functions for backward compatibility and ease of use

// GetModelDimensions is a convenience function to get dimensions for a model
func GetModelDimensions(provider, model string) int {
	dimensions, err := EmbeddingIntel.GetDimensionsForModel(provider, model)
	if err != nil {
		// Return sensible defaults based on provider
		switch strings.ToLower(provider) {
		case "ollama":
			if strings.Contains(strings.ToLower(model), "nomic") {
				return 768
			}
			if strings.Contains(strings.ToLower(model), "mxbai") {
				return 1024
			}
			return 768 // Default for Ollama
		case "openai":
			if strings.Contains(strings.ToLower(model), "large") {
				return 3072
			}
			return 1536 // Default for OpenAI
		default:
			return 1536 // Global default
		}
	}
	return dimensions
}

// ValidateEmbeddingConfig validates an embedding configuration
func ValidateEmbeddingConfig(provider, model, memoryProvider string) []string {
	var warnings []string
	
	// Check if model is known
	_, err := EmbeddingIntel.GetModelInfo(provider, model)
	if err != nil {
		warnings = append(warnings, fmt.Sprintf("Unknown embedding model: %s/%s - using default dimensions", provider, model))
	}
	
	// Check compatibility
	if err := EmbeddingIntel.ValidateCompatibility(provider, model, memoryProvider); err != nil {
		warnings = append(warnings, fmt.Sprintf("Compatibility issue: %v", err))
	}
	
	// Provider-specific warnings
	switch strings.ToLower(provider) {
	case "ollama":
		warnings = append(warnings, "Ensure Ollama is running: ollama serve")
		warnings = append(warnings, fmt.Sprintf("Ensure model is pulled: ollama pull %s", model))
	case "openai":
		warnings = append(warnings, "Ensure OPENAI_API_KEY environment variable is set")
	case "dummy":
		warnings = append(warnings, "Dummy embeddings are for testing only - not suitable for production")
	}
	
	return warnings
}

// GetEmbeddingModelSuggestions returns suggestions for embedding models
func GetEmbeddingModelSuggestions(provider string) []string {
	recommended := EmbeddingIntel.GetRecommendedModels(provider)
	var suggestions []string
	
	for _, model := range recommended {
		suggestion := fmt.Sprintf("%s (%d dimensions) - %s", model.Model, model.Dimensions, model.Notes)
		suggestions = append(suggestions, suggestion)
	}
	
	return suggestions
}