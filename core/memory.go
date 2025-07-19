package core

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strings"
)

// AgentMemoryConfig - enhanced configuration for agent memory storage with RAG support
type AgentMemoryConfig struct {
	// Core memory settings
	Provider   string `toml:"provider"`    // pgvector, weaviate, memory
	Connection string `toml:"connection"`  // postgres://..., http://..., or "memory"
	MaxResults int    `toml:"max_results"` // default: 10
	Dimensions int    `toml:"dimensions"`  // default: 1536
	AutoEmbed  bool   `toml:"auto_embed"`  // default: true

	// RAG-enhanced settings
	EnableKnowledgeBase     bool    `toml:"enable_knowledge_base"`     // default: true
	KnowledgeMaxResults     int     `toml:"knowledge_max_results"`     // default: 20
	KnowledgeScoreThreshold float32 `toml:"knowledge_score_threshold"` // default: 0.7
	ChunkSize               int     `toml:"chunk_size"`                // default: 1000
	ChunkOverlap            int     `toml:"chunk_overlap"`             // default: 200

	// RAG context assembly settings
	EnableRAG           bool    `toml:"enable_rag"`             // default: true
	RAGMaxContextTokens int     `toml:"rag_max_context_tokens"` // default: 4000
	RAGPersonalWeight   float32 `toml:"rag_personal_weight"`    // default: 0.3
	RAGKnowledgeWeight  float32 `toml:"rag_knowledge_weight"`   // default: 0.7
	RAGIncludeSources   bool    `toml:"rag_include_sources"`    // default: true

	// Document processing settings
	Documents DocumentConfig `toml:"documents"`

	// Embedding service settings
	Embedding EmbeddingConfig `toml:"embedding"`

	// Search settings
	Search SearchConfigToml `toml:"search"`
}

// DocumentConfig represents document processing configuration
type DocumentConfig struct {
	AutoChunk                bool     `toml:"auto_chunk"`                 // default: true
	SupportedTypes           []string `toml:"supported_types"`            // default: ["pdf", "txt", "md", "web", "code"]
	MaxFileSize              string   `toml:"max_file_size"`              // default: "10MB"
	EnableMetadataExtraction bool     `toml:"enable_metadata_extraction"` // default: true
	EnableURLScraping        bool     `toml:"enable_url_scraping"`        // default: true
}

// EmbeddingConfig represents embedding service configuration
type EmbeddingConfig struct {
	Provider        string `toml:"provider"`         // openai, ollama, dummy
	Model           string `toml:"model"`            // text-embedding-ada-002, mxbai-embed-large, etc.
	CacheEmbeddings bool   `toml:"cache_embeddings"` // default: true
	APIKey          string `toml:"api_key"`          // API key for service
	BaseURL         string `toml:"base_url"`         // Base URL for service (e.g., Ollama endpoint)
	Endpoint        string `toml:"endpoint"`         // Custom endpoint (deprecated, use BaseURL)
	MaxBatchSize    int    `toml:"max_batch_size"`   // default: 100
	TimeoutSeconds  int    `toml:"timeout_seconds"`  // default: 30
}

// SearchConfigToml represents search configuration
type SearchConfigToml struct {
	HybridSearch         bool    `toml:"hybrid_search"`          // default: true
	KeywordWeight        float32 `toml:"keyword_weight"`         // default: 0.3
	SemanticWeight       float32 `toml:"semantic_weight"`        // default: 0.7
	EnableReranking      bool    `toml:"enable_reranking"`       // default: false
	RerankingModel       string  `toml:"reranking_model"`        // Model for reranking
	EnableQueryExpansion bool    `toml:"enable_query_expansion"` // default: false
}

// NewMemory creates a new memory instance based on configuration
func NewMemory(config AgentMemoryConfig) (Memory, error) {
	// Set core memory defaults
	if config.MaxResults == 0 {
		config.MaxResults = 10
	}
	if config.Dimensions == 0 {
		config.Dimensions = 1536
	}
	if config.Connection == "" && config.Provider == "memory" {
		config.Connection = "memory"
	}

	// Set RAG defaults
	if config.KnowledgeMaxResults == 0 {
		config.KnowledgeMaxResults = 20
	}
	if config.KnowledgeScoreThreshold == 0 {
		// Use lower threshold for dummy embeddings, higher for real embeddings
		if config.Embedding.Provider == "dummy" {
			config.KnowledgeScoreThreshold = 0.0 // No filtering for dummy embeddings
		} else {
			config.KnowledgeScoreThreshold = 0.7 // Standard threshold for real embeddings
		}
	}
	if config.ChunkSize == 0 {
		config.ChunkSize = 1000
	}
	if config.ChunkOverlap == 0 {
		config.ChunkOverlap = 200
	}
	if config.RAGMaxContextTokens == 0 {
		config.RAGMaxContextTokens = 4000
	}
	if config.RAGPersonalWeight == 0 {
		config.RAGPersonalWeight = 0.3
	}
	if config.RAGKnowledgeWeight == 0 {
		config.RAGKnowledgeWeight = 0.7
	}

	// Set document processing defaults
	if len(config.Documents.SupportedTypes) == 0 {
		config.Documents.SupportedTypes = []string{"pdf", "txt", "md", "web", "code"}
	}
	if config.Documents.MaxFileSize == "" {
		config.Documents.MaxFileSize = "10MB"
	}

	// Set embedding service defaults
	if config.Embedding.Provider == "" {
		config.Embedding.Provider = "dummy"
	}
	if config.Embedding.Model == "" {
		config.Embedding.Model = "text-embedding-3-small"
	}
	if config.Embedding.MaxBatchSize == 0 {
		config.Embedding.MaxBatchSize = 100
	}
	if config.Embedding.TimeoutSeconds == 0 {
		config.Embedding.TimeoutSeconds = 30
	}

	// Set search defaults
	if config.Search.KeywordWeight == 0 {
		config.Search.KeywordWeight = 0.3
	}
	if config.Search.SemanticWeight == 0 {
		config.Search.SemanticWeight = 0.7
	}

	switch config.Provider {
	case "memory":
		return newInMemoryProvider(config)
	case "pgvector":
		return newPgVectorProvider(config)
	case "weaviate":
		return newWeaviateProvider(config)
	default:
		return nil, fmt.Errorf("unsupported memory provider: %s", config.Provider)
	}
}

// QuickMemory creates an in-memory provider for quick testing
func QuickMemory() Memory {
	config := AgentMemoryConfig{
		Provider:   "memory",
		Connection: "memory",
		MaxResults: 10,
		AutoEmbed:  true,
	}

	memory, err := NewMemory(config)
	if err != nil {
		// Return no-op memory instead of panicking
		return &NoOpMemory{}
	}

	return memory
}

// Utility functions
func generateID() string {
	bytes := make([]byte, 8)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

func contains(text, query string) bool {
	if len(text) == 0 || len(query) == 0 {
		return false
	}

	// Convert to lowercase for case-insensitive matching
	text = strings.ToLower(text)
	query = strings.ToLower(query)

	// Direct substring match
	return strings.Contains(text, query)
}

func containsAnyTag(tags []string, query string) bool {
	for _, tag := range tags {
		if contains(tag, query) {
			return true
		}
	}
	return false
}

func removeDuplicates(slice []string) []string {
	keys := make(map[string]bool)
	result := []string{}

	for _, item := range slice {
		if !keys[item] {
			keys[item] = true
			result = append(result, item)
		}
	}

	return result
}

func estimateTokenCount(text string) int {
	// Rough estimation: ~4 characters per token
	return len(text) / 4
}

// Simple scoring function for in-memory provider
func calculateScore(content, query string) float32 {
	if len(content) == 0 || len(query) == 0 {
		return 0
	}

	content = strings.ToLower(content)
	query = strings.ToLower(query)

	// Simple substring matching with basic scoring
	if strings.Contains(content, query) {
		// Higher score for exact matches
		if content == query {
			return 1.0
		}
		// Medium score for substring matches
		return 0.7
	}

	// Lower score for partial word matches
	words := strings.Fields(query)
	matchCount := 0
	for _, word := range words {
		if strings.Contains(content, word) {
			matchCount++
		}
	}

	if matchCount > 0 {
		return float32(matchCount) / float32(len(words)) * 0.5
	}

	return 0
}

// Search option constructors
func WithLimit(limit int) SearchOption {
	return func(config *SearchConfig) {
		config.Limit = limit
	}
}

func WithScoreThreshold(threshold float32) SearchOption {
	return func(config *SearchConfig) {
		config.ScoreThreshold = threshold
	}
}

func WithSources(sources []string) SearchOption {
	return func(config *SearchConfig) {
		config.Sources = sources
	}
}

func WithDocumentTypes(types []DocumentType) SearchOption {
	return func(config *SearchConfig) {
		config.DocumentTypes = types
	}
}

func WithTags(tags []string) SearchOption {
	return func(config *SearchConfig) {
		config.Tags = tags
	}
}

func WithIncludePersonal(include bool) SearchOption {
	return func(config *SearchConfig) {
		config.IncludePersonal = include
	}
}

func WithIncludeKnowledge(include bool) SearchOption {
	return func(config *SearchConfig) {
		config.IncludeKnowledge = include
	}
}

// Context option constructors
func WithMaxTokens(maxTokens int) ContextOption {
	return func(config *ContextConfig) {
		config.MaxTokens = maxTokens
	}
}

func WithPersonalWeight(weight float32) ContextOption {
	return func(config *ContextConfig) {
		config.PersonalWeight = weight
	}
}

func WithKnowledgeWeight(weight float32) ContextOption {
	return func(config *ContextConfig) {
		config.KnowledgeWeight = weight
	}
}

func WithHistoryLimit(limit int) ContextOption {
	return func(config *ContextConfig) {
		config.HistoryLimit = limit
	}
}

func WithIncludeSources(include bool) ContextOption {
	return func(config *ContextConfig) {
		config.IncludeSources = include
	}
}

func WithFormatTemplate(template string) ContextOption {
	return func(config *ContextConfig) {
		config.FormatTemplate = template
	}
}
