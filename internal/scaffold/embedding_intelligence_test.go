package scaffold

import (
	"strings"
	"testing"
)

func TestEmbeddingIntelligence_GetModelInfo(t *testing.T) {
	ei := NewEmbeddingIntelligence()

	tests := []struct {
		name         string
		provider     string
		model        string
		wantDims     int
		wantProvider string
		wantError    bool
	}{
		{
			name:         "Valid Ollama nomic-embed-text",
			provider:     "ollama",
			model:        "nomic-embed-text:latest",
			wantDims:     768,
			wantProvider: "ollama",
			wantError:    false,
		},
		{
			name:         "Valid Ollama nomic-embed-text without tag",
			provider:     "ollama",
			model:        "nomic-embed-text",
			wantDims:     768,
			wantProvider: "ollama",
			wantError:    false,
		},
		{
			name:         "Valid OpenAI text-embedding-3-small",
			provider:     "openai",
			model:        "text-embedding-3-small",
			wantDims:     1536,
			wantProvider: "openai",
			wantError:    false,
		},
		{
			name:         "Valid OpenAI text-embedding-3-large",
			provider:     "openai",
			model:        "text-embedding-3-large",
			wantDims:     3072,
			wantProvider: "openai",
			wantError:    false,
		},
		{
			name:         "Valid dummy model",
			provider:     "dummy",
			model:        "dummy",
			wantDims:     1536,
			wantProvider: "dummy",
			wantError:    false,
		},
		{
			name:      "Unknown provider",
			provider:  "unknown",
			model:     "some-model",
			wantError: true,
		},
		{
			name:      "Unknown model for valid provider",
			provider:  "ollama",
			model:     "unknown-model",
			wantError: true,
		},
		{
			name:         "Case insensitive provider",
			provider:     "OLLAMA",
			model:        "nomic-embed-text:latest",
			wantDims:     768,
			wantProvider: "ollama", // Provider is normalized to lowercase
			wantError:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			info, err := ei.GetModelInfo(tt.provider, tt.model)
			
			if tt.wantError {
				if err == nil {
					t.Errorf("GetModelInfo() expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("GetModelInfo() unexpected error: %v", err)
				return
			}

			if info.Dimensions != tt.wantDims {
				t.Errorf("GetModelInfo() dimensions = %d, want %d", info.Dimensions, tt.wantDims)
			}

			if info.Provider != tt.wantProvider {
				t.Errorf("GetModelInfo() provider = %s, want %s", info.Provider, tt.wantProvider)
			}
		})
	}
}

func TestEmbeddingIntelligence_GetRecommendedModels(t *testing.T) {
	ei := NewEmbeddingIntelligence()

	tests := []struct {
		name         string
		provider     string
		wantMinCount int
		wantMaxCount int
	}{
		{
			name:         "Ollama recommended models",
			provider:     "ollama",
			wantMinCount: 1,
			wantMaxCount: 10,
		},
		{
			name:         "OpenAI recommended models",
			provider:     "openai",
			wantMinCount: 1,
			wantMaxCount: 10,
		},
		{
			name:         "Unknown provider",
			provider:     "unknown",
			wantMinCount: 0,
			wantMaxCount: 0,
		},
		{
			name:         "Case insensitive provider",
			provider:     "OLLAMA",
			wantMinCount: 1,
			wantMaxCount: 10,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			models := ei.GetRecommendedModels(tt.provider)
			
			if len(models) < tt.wantMinCount {
				t.Errorf("GetRecommendedModels() returned %d models, want at least %d", len(models), tt.wantMinCount)
			}

			if len(models) > tt.wantMaxCount {
				t.Errorf("GetRecommendedModels() returned %d models, want at most %d", len(models), tt.wantMaxCount)
			}

			// Verify all returned models are marked as recommended
			for _, model := range models {
				if !model.Recommended {
					t.Errorf("GetRecommendedModels() returned non-recommended model: %s", model.Model)
				}
			}
		})
	}
}

func TestEmbeddingIntelligence_ValidateCompatibility(t *testing.T) {
	ei := NewEmbeddingIntelligence()

	tests := []struct {
		name           string
		provider       string
		model          string
		memoryProvider string
		wantError      bool
	}{
		{
			name:           "Valid Ollama with pgvector",
			provider:       "ollama",
			model:          "nomic-embed-text:latest",
			memoryProvider: "pgvector",
			wantError:      false,
		},
		{
			name:           "Valid OpenAI with weaviate",
			provider:       "openai",
			model:          "text-embedding-3-small",
			memoryProvider: "weaviate",
			wantError:      false,
		},
		{
			name:           "Dummy with weaviate (should warn)",
			provider:       "dummy",
			model:          "dummy",
			memoryProvider: "weaviate",
			wantError:      true,
		},
		{
			name:           "Unknown model",
			provider:       "ollama",
			model:          "unknown-model",
			memoryProvider: "pgvector",
			wantError:      true,
		},
		{
			name:           "Valid model with memory provider",
			provider:       "ollama",
			model:          "nomic-embed-text:latest",
			memoryProvider: "memory",
			wantError:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ei.ValidateCompatibility(tt.provider, tt.model, tt.memoryProvider)
			
			if tt.wantError && err == nil {
				t.Errorf("ValidateCompatibility() expected error but got none")
			}
			
			if !tt.wantError && err != nil {
				t.Errorf("ValidateCompatibility() unexpected error: %v", err)
			}
		})
	}
}

func TestEmbeddingIntelligence_GetDimensionsForModel(t *testing.T) {
	ei := NewEmbeddingIntelligence()

	tests := []struct {
		name      string
		provider  string
		model     string
		wantDims  int
		wantError bool
	}{
		{
			name:      "Valid nomic-embed-text",
			provider:  "ollama",
			model:     "nomic-embed-text:latest",
			wantDims:  768,
			wantError: false,
		},
		{
			name:      "Valid OpenAI small",
			provider:  "openai",
			model:     "text-embedding-3-small",
			wantDims:  1536,
			wantError: false,
		},
		{
			name:      "Valid OpenAI large",
			provider:  "openai",
			model:     "text-embedding-3-large",
			wantDims:  3072,
			wantError: false,
		},
		{
			name:      "Unknown model",
			provider:  "ollama",
			model:     "unknown-model",
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dims, err := ei.GetDimensionsForModel(tt.provider, tt.model)
			
			if tt.wantError {
				if err == nil {
					t.Errorf("GetDimensionsForModel() expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("GetDimensionsForModel() unexpected error: %v", err)
				return
			}

			if dims != tt.wantDims {
				t.Errorf("GetDimensionsForModel() = %d, want %d", dims, tt.wantDims)
			}
		})
	}
}

func TestGetModelDimensions(t *testing.T) {
	tests := []struct {
		name     string
		provider string
		model    string
		want     int
	}{
		{
			name:     "Known Ollama nomic model",
			provider: "ollama",
			model:    "nomic-embed-text:latest",
			want:     768,
		},
		{
			name:     "Known OpenAI small model",
			provider: "openai",
			model:    "text-embedding-3-small",
			want:     1536,
		},
		{
			name:     "Known OpenAI large model",
			provider: "openai",
			model:    "text-embedding-3-large",
			want:     3072,
		},
		{
			name:     "Unknown Ollama model with nomic in name",
			provider: "ollama",
			model:    "nomic-custom-model",
			want:     768,
		},
		{
			name:     "Unknown Ollama model with mxbai in name",
			provider: "ollama",
			model:    "mxbai-custom-model",
			want:     1024,
		},
		{
			name:     "Unknown Ollama model",
			provider: "ollama",
			model:    "unknown-model",
			want:     768, // Default for Ollama
		},
		{
			name:     "Unknown OpenAI model with large in name",
			provider: "openai",
			model:    "custom-large-model",
			want:     3072,
		},
		{
			name:     "Unknown OpenAI model",
			provider: "openai",
			model:    "unknown-model",
			want:     1536, // Default for OpenAI
		},
		{
			name:     "Unknown provider",
			provider: "unknown",
			model:    "some-model",
			want:     1536, // Global default
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetModelDimensions(tt.provider, tt.model)
			if got != tt.want {
				t.Errorf("GetModelDimensions() = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestValidateEmbeddingConfig(t *testing.T) {
	tests := []struct {
		name           string
		provider       string
		model          string
		memoryProvider string
		wantWarnings   int
	}{
		{
			name:           "Valid Ollama configuration",
			provider:       "ollama",
			model:          "nomic-embed-text:latest",
			memoryProvider: "pgvector",
			wantWarnings:   2, // Ollama serve + model pull warnings
		},
		{
			name:           "Valid OpenAI configuration",
			provider:       "openai",
			model:          "text-embedding-3-small",
			memoryProvider: "weaviate",
			wantWarnings:   1, // API key warning
		},
		{
			name:           "Dummy provider",
			provider:       "dummy",
			model:          "dummy",
			memoryProvider: "memory",
			wantWarnings:   1, // Testing only warning
		},
		{
			name:           "Unknown model",
			provider:       "ollama",
			model:          "unknown-model",
			memoryProvider: "pgvector",
			wantWarnings:   4, // Unknown model + compatibility issue + Ollama warnings
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			warnings := ValidateEmbeddingConfig(tt.provider, tt.model, tt.memoryProvider)
			
			if len(warnings) != tt.wantWarnings {
				t.Errorf("ValidateEmbeddingConfig() returned %d warnings, want %d. Warnings: %v", 
					len(warnings), tt.wantWarnings, warnings)
			}
		})
	}
}

func TestGetEmbeddingModelSuggestions(t *testing.T) {
	tests := []struct {
		name         string
		provider     string
		wantMinCount int
	}{
		{
			name:         "Ollama suggestions",
			provider:     "ollama",
			wantMinCount: 1,
		},
		{
			name:         "OpenAI suggestions",
			provider:     "openai",
			wantMinCount: 1,
		},
		{
			name:         "Unknown provider",
			provider:     "unknown",
			wantMinCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			suggestions := GetEmbeddingModelSuggestions(tt.provider)
			
			if len(suggestions) < tt.wantMinCount {
				t.Errorf("GetEmbeddingModelSuggestions() returned %d suggestions, want at least %d", 
					len(suggestions), tt.wantMinCount)
			}

			// Verify suggestions contain dimensions and notes
			for _, suggestion := range suggestions {
				if !strings.Contains(suggestion, "dimensions") {
					t.Errorf("GetEmbeddingModelSuggestions() suggestion missing dimensions: %s", suggestion)
				}
			}
		})
	}
}