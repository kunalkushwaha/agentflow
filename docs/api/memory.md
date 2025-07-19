# Memory API Reference

**Complete API reference for AgentFlow's memory system**

This document provides comprehensive API documentation for AgentFlow's memory system, including interfaces, types, methods, and usage examples.

## 📚 Table of Contents

- [Overview](#overview)
- [Core Interfaces](#core-interfaces)
- [Memory Providers](#memory-providers)
- [Data Types](#data-types)
- [Configuration Types](#configuration-types)
- [Search Options](#search-options)
- [Error Types](#error-types)
- [Usage Examples](#usage-examples)

## 🎯 Overview

The AgentFlow memory system provides a unified interface for persistent storage, vector search, and RAG (Retrieval-Augmented Generation) capabilities. The API is designed to be provider-agnostic, allowing seamless switching between different storage backends.

### Key Features

- **Unified Interface**: Single API for all memory operations
- **Multiple Providers**: Support for in-memory, PostgreSQL+pgvector, and Weaviate
- **Session Isolation**: Complete data separation between users/sessions
- **RAG Capabilities**: Document ingestion, knowledge search, context building
- **Vector Search**: Semantic similarity search with configurable scoring
- **Batch Operations**: Efficient bulk operations for large datasets

## 🔧 Core Interfaces

### Memory Interface

The primary interface for all memory operations.

```go
type Memory interface {
    // Personal memory operations
    Store(ctx context.Context, content string, tags ...string) error
    Query(ctx context.Context, query string, limit ...int) ([]Result, error)
    
    // Key-value storage
    Remember(ctx context.Context, key string, value any) error
    Recall(ctx context.Context, key string) (any, error)
    
    // Chat history
    AddMessage(ctx context.Context, role, content string) error
    GetHistory(ctx context.Context, limit ...int) ([]Message, error)
    
    // Knowledge base (RAG)
    IngestDocument(ctx context.Context, doc Document) error
    IngestDocuments(ctx context.Context, docs []Document) error
    SearchKnowledge(ctx context.Context, query string, options ...SearchOption) ([]KnowledgeResult, error)
    SearchAll(ctx context.Context, query string, options ...SearchOption) (*HybridResult, error)
    BuildContext(ctx context.Context, query string, options ...ContextOption) (*RAGContext, error)
    
    // Session management
    NewSession() string
    SetSession(ctx context.Context, sessionID string) context.Context
    ClearSession(ctx context.Context) error
    Close() error
}
```

### EmbeddingProvider Interface

Interface for embedding generation services.

```go
type EmbeddingProvider interface {
    // Generate embeddings for text
    GenerateEmbedding(ctx context.Context, text string) ([]float32, error)
    GenerateEmbeddings(ctx context.Context, texts []string) ([][]float32, error)
    
    // Provider information
    GetDimensions() int
    GetModel() string
    GetProvider() string
    
    // Lifecycle
    Close() error
}
```

### MemoryProvider Interface

Base interface for memory storage providers.

```go
type MemoryProvider interface {
    // Core operations
    Store(ctx context.Context, sessionID, content string, embedding []float32, tags []string) error
    Query(ctx context.Context, sessionID string, queryEmbedding []float32, limit int) ([]Result, error)
    
    // Key-value operations
    Set(ctx context.Context, sessionID, key string, value []byte) error
    Get(ctx context.Context, sessionID, key string) ([]byte, error)
    Delete(ctx context.Context, sessionID, key string) error
    
    // Session management
    ClearSession(ctx context.Context, sessionID string) error
    
    // Lifecycle
    Close() error
}
```

## 💾 Memory Providers

### NewMemory Function

Creates a new memory instance with the specified configuration.

```go
func NewMemory(config AgentMemoryConfig) (Memory, error)
```

**Parameters:**
- `config`: Configuration for the memory system

**Returns:**
- `Memory`: Memory interface implementation
- `error`: Error if initialization fails

**Example:**
```go
config := AgentMemoryConfig{
    Provider:   "pgvector",
    Connection: "postgres://user:password@localhost:5432/agentflow",
    Dimensions: 1536,
    Embedding: EmbeddingConfig{
        Provider: "openai",
        APIKey:   "your-api-key",
        Model:    "text-embedding-3-small",
    },
}

memory, err := NewMemory(config)
if err != nil {
    log.Fatal(err)
}
defer memory.Close()
```

### Provider-Specific Types

#### InMemoryProvider

```go
type InMemoryProvider struct {
    // Implementation details are internal
}

// NewInMemoryProvider creates a new in-memory provider
func NewInMemoryProvider(config AgentMemoryConfig) (*InMemoryProvider, error)
```

#### PgVectorProvider

```go
type PgVectorProvider struct {
    // Implementation details are internal
}

// NewPgVectorProvider creates a new PostgreSQL+pgvector provider
func NewPgVectorProvider(config AgentMemoryConfig) (*PgVectorProvider, error)

// Enhanced methods for batch operations
func (p *PgVectorProvider) BatchStore(ctx context.Context, requests []BatchStoreRequest) error
func (p *PgVectorProvider) BatchIngestDocuments(ctx context.Context, docs []Document) error
```

#### WeaviateProvider

```go
type WeaviateProvider struct {
    // Implementation details are internal
}

// NewWeaviateProvider creates a new Weaviate provider
func NewWeaviateProvider(config AgentMemoryConfig) (*WeaviateProvider, error)
```

## 📊 Data Types

### Result

Represents a personal memory search result.

```go
type Result struct {
    Content   string      `json:"content"`    // The stored content
    Score     float32     `json:"score"`      // Similarity score (0.0-1.0)
    Tags      []string    `json:"tags"`       // Associated tags
    CreatedAt time.Time   `json:"created_at"` // When the memory was created
}
```

### KnowledgeResult

Represents a knowledge base search result.

```go
type KnowledgeResult struct {
    Content    string         `json:"content"`     // Document content
    Score      float32        `json:"score"`       // Similarity score
    Source     string         `json:"source"`      // Document source
    Title      string         `json:"title"`       // Document title
    DocumentID string         `json:"document_id"` // Unique document ID
    Metadata   map[string]any `json:"metadata"`    // Additional metadata
    Tags       []string       `json:"tags"`        // Document tags
    CreatedAt  time.Time      `json:"created_at"`  // Creation timestamp
    ChunkIndex int            `json:"chunk_index"` // Chunk position in document
}
```

### Document

Represents a document for knowledge base ingestion.

```go
type Document struct {
    ID         string         `json:"id"`          // Unique document identifier
    Title      string         `json:"title"`       // Document title
    Content    string         `json:"content"`     // Document content
    Source     string         `json:"source"`      // Source URL or path
    Type       DocumentType   `json:"type"`        // Document type
    Metadata   map[string]any `json:"metadata"`    // Additional metadata
    Tags       []string       `json:"tags"`        // Document tags
    CreatedAt  time.Time      `json:"created_at"`  // Creation timestamp
    UpdatedAt  time.Time      `json:"updated_at"`  // Last update timestamp
    ChunkIndex int            `json:"chunk_index"` // Chunk index (for chunked documents)
    ChunkTotal int            `json:"chunk_total"` // Total chunks in document
}
```

### DocumentType

Enumeration of supported document types.

```go
type DocumentType string

const (
    DocumentTypeText     DocumentType = "text"
    DocumentTypePDF      DocumentType = "pdf"
    DocumentTypeMarkdown DocumentType = "markdown"
    DocumentTypeHTML     DocumentType = "html"
    DocumentTypeCode     DocumentType = "code"
    DocumentTypeJSON     DocumentType = "json"
    DocumentTypeWeb      DocumentType = "web"
)
```

### Message

Represents a chat history message.

```go
type Message struct {
    Role      string         `json:"role"`       // "user", "assistant", or "system"
    Content   string         `json:"content"`    // Message content
    Metadata  map[string]any `json:"metadata"`   // Additional metadata
    CreatedAt time.Time      `json:"created_at"` // Message timestamp
}
```

### HybridResult

Represents combined search results from personal memory and knowledge base.

```go
type HybridResult struct {
    PersonalMemory []Result          `json:"personal_memory"` // Personal memory results
    Knowledge      []KnowledgeResult `json:"knowledge"`       // Knowledge base results
    TotalResults   int               `json:"total_results"`   // Total number of results
    SearchTime     time.Duration     `json:"search_time"`     // Time taken for search
}
```

### RAGContext

Represents assembled context for RAG operations.

```go
type RAGContext struct {
    Query           string            `json:"query"`            // Original query
    ContextText     string            `json:"context_text"`     // Assembled context text
    PersonalMemory  []Result          `json:"personal_memory"`  // Relevant personal memories
    Knowledge       []KnowledgeResult `json:"knowledge"`        // Relevant knowledge
    ChatHistory     []Message         `json:"chat_history"`     // Recent chat history
    Sources         []string          `json:"sources"`          // Source documents
    TokenCount      int               `json:"token_count"`      // Estimated token count
    AssemblyTime    time.Duration     `json:"assembly_time"`    // Time taken to assemble
}
```

## ⚙️ Configuration Types

### AgentMemoryConfig

Main configuration structure for the memory system.

```go
type AgentMemoryConfig struct {
    // Core settings
    Provider   string `toml:"provider"`   // "memory", "pgvector", "weaviate"
    Connection string `toml:"connection"` // Provider-specific connection string
    Dimensions int    `toml:"dimensions"` // Vector embedding dimensions
    AutoEmbed  bool   `toml:"auto_embed"` // Automatically generate embeddings
    
    // Embedding configuration
    Embedding EmbeddingConfig `toml:"embedding"`
    
    // RAG settings
    RAG RAGConfig `toml:"rag"`
    
    // Provider-specific settings
    PgVector PgVectorConfig `toml:"pgvector"`
    Weaviate WeaviateConfig `toml:"weaviate"`
    
    // Advanced settings
    Advanced AdvancedConfig `toml:"advanced"`
}
```

### EmbeddingConfig

Configuration for embedding providers.

```go
type EmbeddingConfig struct {
    Provider       string        `toml:"provider"`         // "openai", "azure", "ollama", "dummy"
    APIKey         string        `toml:"api_key"`          // API key for the provider
    Model          string        `toml:"model"`            // Embedding model name
    BaseURL        string        `toml:"base_url"`         // Custom base URL
    MaxBatchSize   int           `toml:"max_batch_size"`   // Maximum batch size
    TimeoutSeconds int           `toml:"timeout_seconds"`  // Request timeout
    CacheEnabled   bool          `toml:"cache_enabled"`    // Enable embedding cache
}
```

### RAGConfig

Configuration for RAG functionality.

```go
type RAGConfig struct {
    Enabled         bool    `toml:"enabled"`           // Enable RAG functionality
    ChunkSize       int     `toml:"chunk_size"`        // Document chunk size
    Overlap         int     `toml:"overlap"`           // Chunk overlap size
    TopK            int     `toml:"top_k"`             // Number of results to retrieve
    ScoreThreshold  float32 `toml:"score_threshold"`   // Minimum similarity score
    HybridSearch    bool    `toml:"hybrid_search"`     // Enable hybrid search
    SessionMemory   bool    `toml:"session_memory"`    // Enable session isolation
}
```

### PgVectorConfig

PostgreSQL+pgvector specific configuration.

```go
type PgVectorConfig struct {
    Connection         string `toml:"connection"`           // PostgreSQL connection string
    TableName          string `toml:"table_name"`           // Table name for memories
    ConnectionPoolSize int    `toml:"connection_pool_size"` // Connection pool size
}
```

### WeaviateConfig

Weaviate specific configuration.

```go
type WeaviateConfig struct {
    Connection string `toml:"connection"` // Weaviate URL
    APIKey     string `toml:"api_key"`    // API key for authentication
    ClassName  string `toml:"class_name"` // Weaviate class name
    Timeout    string `toml:"timeout"`    // Request timeout
    MaxRetries int    `toml:"max_retries"` // Maximum retry attempts
}
```

### AdvancedConfig

Advanced configuration options.

```go
type AdvancedConfig struct {
    RetryMaxAttempts     int           `toml:"retry_max_attempts"`      // Maximum retry attempts
    RetryBaseDelay       time.Duration `toml:"retry_base_delay"`        // Base delay between retries
    RetryMaxDelay        time.Duration `toml:"retry_max_delay"`         // Maximum delay between retries
    ConnectionPoolSize   int           `toml:"connection_pool_size"`    // Database connection pool size
    HealthCheckInterval  time.Duration `toml:"health_check_interval"`   // Health check interval
}
```

## 🔍 Search Options

### SearchOption

Functional options for customizing search behavior.

```go
type SearchOption func(*SearchConfig)

// Available search options
func WithLimit(limit int) SearchOption
func WithScoreThreshold(threshold float32) SearchOption
func WithSources(sources []string) SearchOption
func WithTags(tags []string) SearchOption
func WithDocumentTypes(types []DocumentType) SearchOption
func WithDateRange(start, end time.Time) SearchOption
func WithIncludePersonal(include bool) SearchOption
func WithIncludeKnowledge(include bool) SearchOption
```

### ContextOption

Functional options for RAG context building.

```go
type ContextOption func(*ContextConfig)

// Available context options
func WithMaxTokens(tokens int) ContextOption
func WithPersonalWeight(weight float32) ContextOption
func WithKnowledgeWeight(weight float32) ContextOption
func WithHistoryLimit(limit int) ContextOption
func WithIncludeSources(include bool) ContextOption
```

### BatchStoreRequest

Request structure for batch store operations.

```go
type BatchStoreRequest struct {
    Content string   `json:"content"` // Content to store
    Tags    []string `json:"tags"`    // Associated tags
}
```

## ❌ Error Types

### Common Errors

```go
var (
    ErrMemoryNotInitialized = errors.New("memory system not initialized")
    ErrInvalidProvider      = errors.New("invalid memory provider")
    ErrInvalidConfiguration = errors.New("invalid memory configuration")
    ErrSessionNotFound      = errors.New("session not found")
    ErrDocumentNotFound     = errors.New("document not found")
    ErrEmbeddingFailed      = errors.New("embedding generation failed")
    ErrConnectionFailed     = errors.New("connection to memory provider failed")
)
```

### Provider-Specific Errors

```go
// PostgreSQL errors
var (
    ErrPgVectorExtensionMissing = errors.New("pgvector extension not installed")
    ErrPgVectorConnectionFailed = errors.New("failed to connect to PostgreSQL")
)

// Weaviate errors
var (
    ErrWeaviateConnectionFailed = errors.New("failed to connect to Weaviate")
    ErrWeaviateSchemaError      = errors.New("Weaviate schema error")
)

// Embedding errors
var (
    ErrOpenAIAPIKeyMissing = errors.New("OpenAI API key not provided")
    ErrOllamaNotRunning    = errors.New("Ollama server not running")
)
```

## 💡 Usage Examples

### Basic Memory Operations

```go
package main

import (
    "context"
    "log"
    
    agentflow "github.com/kunalkushwaha/agentflow/core"
)

func basicMemoryExample() {
    // Create memory configuration
    config := agentflow.AgentMemoryConfig{
        Provider:   "memory",
        Connection: "memory",
        Dimensions: 1536,
        Embedding: agentflow.EmbeddingConfig{
            Provider: "dummy",
            Model:    "text-embedding-3-small",
        },
    }
    
    // Initialize memory
    memory, err := agentflow.NewMemory(config)
    if err != nil {
        log.Fatal(err)
    }
    defer memory.Close()
    
    // Create session
    ctx := memory.SetSession(context.Background(), "user-123")
    
    // Store memories
    err = memory.Store(ctx, "I love programming in Go", "programming", "preference")
    if err != nil {
        log.Fatal(err)
    }
    
    // Query memories
    results, err := memory.Query(ctx, "programming languages", 5)
    if err != nil {
        log.Fatal(err)
    }
    
    for _, result := range results {
        log.Printf("Memory: %s (Score: %.2f)", result.Content, result.Score)
    }
}
```

### RAG Operations

```go
func ragExample() {
    // Configure memory with RAG
    config := agentflow.AgentMemoryConfig{
        Provider:   "pgvector",
        Connection: "postgres://user:password@localhost:5432/agentflow",
        Dimensions: 1536,
        Embedding: agentflow.EmbeddingConfig{
            Provider: "openai",
            APIKey:   "your-api-key",
            Model:    "text-embedding-3-small",
        },
        RAG: agentflow.RAGConfig{
            Enabled:        true,
            ChunkSize:      1000,
            Overlap:        200,
            TopK:           5,
            ScoreThreshold: 0.7,
        },
    }
    
    memory, err := agentflow.NewMemory(config)
    if err != nil {
        log.Fatal(err)
    }
    defer memory.Close()
    
    ctx := context.Background()
    
    // Ingest document
    doc := agentflow.Document{
        ID:      "guide-001",
        Title:   "AgentFlow Guide",
        Content: "AgentFlow is a Go SDK for building AI agent systems...",
        Source:  "docs/guide.md",
        Type:    agentflow.DocumentTypeMarkdown,
        Tags:    []string{"documentation", "guide"},
    }
    
    err = memory.IngestDocument(ctx, doc)
    if err != nil {
        log.Fatal(err)
    }
    
    // Search knowledge base
    results, err := memory.SearchKnowledge(ctx, "How to build agents?",
        agentflow.WithScoreThreshold(0.8),
        agentflow.WithLimit(3),
    )
    if err != nil {
        log.Fatal(err)
    }
    
    for _, result := range results {
        log.Printf("Knowledge: %s (Score: %.2f, Source: %s)", 
            result.Content, result.Score, result.Source)
    }
    
    // Build RAG context
    ragContext, err := memory.BuildContext(ctx, "How to build agents?",
        agentflow.WithMaxTokens(3000),
        agentflow.WithIncludeSources(true),
    )
    if err != nil {
        log.Fatal(err)
    }
    
    log.Printf("RAG Context (%d tokens):\n%s", 
        ragContext.TokenCount, ragContext.ContextText)
}
```

### Advanced Search

```go
func advancedSearchExample() {
    memory, err := agentflow.NewMemory(config)
    if err != nil {
        log.Fatal(err)
    }
    defer memory.Close()
    
    ctx := memory.SetSession(context.Background(), "user-456")
    
    // Hybrid search across personal memory and knowledge base
    hybridResult, err := memory.SearchAll(ctx, "machine learning best practices",
        agentflow.WithScoreThreshold(0.7),
        agentflow.WithLimit(10),
        agentflow.WithTags([]string{"ml", "best-practices"}),
        agentflow.WithIncludePersonal(true),
        agentflow.WithIncludeKnowledge(true),
    )
    if err != nil {
        log.Fatal(err)
    }
    
    log.Printf("Found %d personal memories and %d knowledge results", 
        len(hybridResult.PersonalMemory), len(hybridResult.Knowledge))
    log.Printf("Search completed in %v", hybridResult.SearchTime)
    
    // Process personal memories
    for _, memory := range hybridResult.PersonalMemory {
        log.Printf("Personal: %s (Score: %.2f)", memory.Content, memory.Score)
    }
    
    // Process knowledge results
    for _, knowledge := range hybridResult.Knowledge {
        log.Printf("Knowledge: %s (Score: %.2f, Source: %s)", 
            knowledge.Content, knowledge.Score, knowledge.Source)
    }
}
```

### Session Management

```go
func sessionManagementExample() {
    memory, err := agentflow.NewMemory(config)
    if err != nil {
        log.Fatal(err)
    }
    defer memory.Close()
    
    // Create multiple sessions
    session1 := memory.NewSession()
    session2 := memory.NewSession()
    
    ctx1 := memory.SetSession(context.Background(), session1)
    ctx2 := memory.SetSession(context.Background(), session2)
    
    // Store different data in each session
    memory.Store(ctx1, "I prefer coffee", "beverage")
    memory.Store(ctx2, "I prefer tea", "beverage")
    
    // Query each session independently
    results1, _ := memory.Query(ctx1, "beverage preference", 1)
    results2, _ := memory.Query(ctx2, "beverage preference", 1)
    
    log.Printf("Session 1: %s", results1[0].Content) // "I prefer coffee"
    log.Printf("Session 2: %s", results2[0].Content) // "I prefer tea"
    
    // Clear a session
    err = memory.ClearSession(ctx1)
    if err != nil {
        log.Printf("Failed to clear session: %v", err)
    }
}
```

### Batch Operations (PgVector)

```go
func batchOperationsExample() {
    config := agentflow.AgentMemoryConfig{
        Provider:   "pgvector",
        Connection: "postgres://user:password@localhost:5432/agentflow",
        // ... other config
    }
    
    memory, err := agentflow.NewMemory(config)
    if err != nil {
        log.Fatal(err)
    }
    defer memory.Close()
    
    // Cast to PgVectorProvider for batch operations
    if provider, ok := memory.(*agentflow.PgVectorProvider); ok {
        ctx := memory.SetSession(context.Background(), "batch-session")
        
        // Batch store multiple memories
        batchRequests := []agentflow.BatchStoreRequest{
            {Content: "First memory", Tags: []string{"batch", "test"}},
            {Content: "Second memory", Tags: []string{"batch", "test"}},
            {Content: "Third memory", Tags: []string{"batch", "test"}},
        }
        
        err = provider.BatchStore(ctx, batchRequests)
        if err != nil {
            log.Fatal(err)
        }
        
        log.Printf("Batch stored %d memories", len(batchRequests))
        
        // Batch ingest documents
        docs := []agentflow.Document{
            {ID: "doc1", Title: "Document 1", Content: "Content 1"},
            {ID: "doc2", Title: "Document 2", Content: "Content 2"},
            {ID: "doc3", Title: "Document 3", Content: "Content 3"},
        }
        
        err = provider.BatchIngestDocuments(ctx, docs)
        if err != nil {
            log.Fatal(err)
        }
        
        log.Printf("Batch ingested %d documents", len(docs))
    }
}
```

### Error Handling

```go
func errorHandlingExample() {
    config := agentflow.AgentMemoryConfig{
        Provider:   "pgvector",
        Connection: "invalid-connection-string",
    }
    
    memory, err := agentflow.NewMemory(config)
    if err != nil {
        // Handle initialization errors
        switch {
        case errors.Is(err, agentflow.ErrInvalidConfiguration):
            log.Printf("Configuration error: %v", err)
        case errors.Is(err, agentflow.ErrConnectionFailed):
            log.Printf("Connection error: %v", err)
        case errors.Is(err, agentflow.ErrPgVectorExtensionMissing):
            log.Printf("pgvector extension not installed: %v", err)
        default:
            log.Printf("Unknown error: %v", err)
        }
        return
    }
    defer memory.Close()
    
    ctx := memory.SetSession(context.Background(), "error-session")
    
    // Handle operation errors
    err = memory.Store(ctx, "test content", "test")
    if err != nil {
        switch {
        case errors.Is(err, agentflow.ErrEmbeddingFailed):
            log.Printf("Embedding generation failed: %v", err)
        case errors.Is(err, agentflow.ErrSessionNotFound):
            log.Printf("Session not found: %v", err)
        default:
            log.Printf("Store operation failed: %v", err)
        }
    }
}
```

---

## 🎯 Summary

The AgentFlow Memory API provides:

✅ **Unified Interface**: Single API for all memory operations across providers  
✅ **Type Safety**: Comprehensive type definitions with clear documentation  
✅ **Flexible Configuration**: Extensive configuration options for all use cases  
✅ **Error Handling**: Well-defined error types with clear error messages  
✅ **Performance**: Batch operations and optimization options  
✅ **Production Ready**: Advanced features for enterprise deployment  

This API reference covers all public interfaces and types in the AgentFlow memory system. For implementation guides and examples, see:

- **[Memory System Guide](../guides/Memory.md)** - Complete implementation guide
- **[RAG Configuration Guide](../guides/RAGConfiguration.md)** - RAG setup and configuration
- **[Memory Provider Setup](../guides/MemoryProviderSetup.md)** - Provider installation guides