# AgentFlow Memory System - Complete Implementation Guide

## 🎯 Overview

AgentFlow provides a powerful, production-ready memory system that enables agents to maintain persistent context, conversation history, and knowledge bases. The system supports multiple storage backends, RAG (Retrieval-Augmented Generation) capabilities, and advanced features like vector embeddings, batch operations, and intelligent retry logic.

## 📚 Table of Contents

- [Quick Start](#quick-start)
- [Core Concepts](#core-concepts)
- [Configuration](#configuration)
- [Memory Providers](#memory-providers)
- [API Reference](#api-reference)
- [RAG (Retrieval-Augmented Generation)](#rag-retrieval-augmented-generation)
- [Advanced Features](#advanced-features)
- [Examples](#examples)
- [Performance & Optimization](#performance--optimization)
- [Troubleshooting](#troubleshooting)

## 🚀 Quick Start

### 1. Basic Setup

```go
package main

import (
    "context"
    "log"
    
    "github.com/kunalkushwaha/agentflow/core"
)

func main() {
    // Create memory configuration
    config := core.AgentMemoryConfig{
        Provider:   "memory",    // In-memory for development
        Connection: "memory",
        Dimensions: 1536,
        Embedding: core.EmbeddingConfig{
            Provider: "dummy",   // For testing
            Model:    "text-embedding-3-small",
        },
    }
    
    // Create memory provider
    memory, err := core.NewMemory(config)
    if err != nil {
        log.Fatal(err)
    }
    defer memory.Close()
    
    // Create context with session
    ctx := memory.SetSession(context.Background(), "user-123")
    
    // Store a memory
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
        fmt.Printf("Content: %s, Score: %.2f\n", result.Content, result.Score)
    }
}
```

### 2. Configuration File Setup

Create an `agentflow.toml` configuration file:

```toml
[agent_memory]
# Basic settings
provider = "pgvector"
connection = "postgres://user:password@localhost:5432/agentflow"
max_results = 10
dimensions = 1536
auto_embed = true

# RAG settings
enable_knowledge_base = true
knowledge_max_results = 20
knowledge_score_threshold = 0.7
enable_rag = true
rag_max_context_tokens = 4000
rag_personal_weight = 0.3
rag_knowledge_weight = 0.7
rag_include_sources = true

# Embedding configuration
[agent_memory.embedding]
provider = "openai"
api_key = "${OPENAI_API_KEY}"
model = "text-embedding-3-small"
max_batch_size = 50
timeout_seconds = 30
```

## 🧠 Core Concepts

### Memory Interface

The unified Memory interface provides all memory operations:

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

### Session Management

All memory operations are session-scoped, providing complete isolation between different users or conversation threads:

```go
// Create new session
sessionID := memory.NewSession()
ctx := memory.SetSession(context.Background(), sessionID)

// Different sessions see different data
ctx1 := memory.SetSession(context.Background(), "user-1")
ctx2 := memory.SetSession(context.Background(), "user-2")

memory.Store(ctx1, "I like coffee", "preference")
memory.Store(ctx2, "I like tea", "preference")

// Queries are isolated by session
results1, _ := memory.Query(ctx1, "beverages", 5)  // Returns coffee preference
results2, _ := memory.Query(ctx2, "beverages", 5)  // Returns tea preference
```

### Data Types

#### Personal Memory Result
```go
type Result struct {
    Content   string      `json:"content"`
    Score     float32     `json:"score"`
    Tags      []string    `json:"tags"`
    CreatedAt time.Time   `json:"created_at"`
}
```

#### Knowledge Base Result
```go
type KnowledgeResult struct {
    Content    string         `json:"content"`
    Score      float32        `json:"score"`
    Source     string         `json:"source"`
    Title      string         `json:"title"`
    DocumentID string         `json:"document_id"`
    Metadata   map[string]any `json:"metadata"`
    Tags       []string       `json:"tags"`
    CreatedAt  time.Time      `json:"created_at"`
    ChunkIndex int            `json:"chunk_index"`
}
```

#### Document Structure
```go
type Document struct {
    ID         string         `json:"id"`
    Title      string         `json:"title"`
    Content    string         `json:"content"`
    Source     string         `json:"source"`
    Type       DocumentType   `json:"type"`
    Metadata   map[string]any `json:"metadata"`
    Tags       []string       `json:"tags"`
    CreatedAt  time.Time      `json:"created_at"`
    UpdatedAt  time.Time      `json:"updated_at"`
    ChunkIndex int            `json:"chunk_index"`
    ChunkTotal int            `json:"chunk_total"`
}
```

## ⚙️ Configuration

### Complete Configuration Reference

```toml
[agent_memory]
# Provider Selection
provider = "pgvector"  # Options: memory, pgvector, weaviate
connection = "postgres://user:password@localhost:5432/agentflow"

# Core Settings
max_results = 10                    # Maximum results for personal memory queries
dimensions = 1536                   # Vector embedding dimensions
auto_embed = true                   # Automatically generate embeddings

# Knowledge Base Settings
enable_knowledge_base = true        # Enable document ingestion and search
knowledge_max_results = 20          # Maximum results from knowledge base
knowledge_score_threshold = 0.7     # Minimum relevance score (0.0-1.0)
chunk_size = 1000                  # Document chunk size in characters
chunk_overlap = 200                # Overlap between chunks in characters

# RAG Context Assembly
enable_rag = true                  # Enable RAG context building
rag_max_context_tokens = 4000      # Maximum tokens in assembled context
rag_personal_weight = 0.3          # Weight for personal memory (0.0-1.0)
rag_knowledge_weight = 0.7         # Weight for knowledge base (0.0-1.0)
rag_include_sources = true         # Include source attribution in context

# Embedding Service Configuration
[agent_memory.embedding]
provider = "openai"                # Options: openai, azure, ollama, dummy
api_key = "${OPENAI_API_KEY}"      # API key (supports environment variables)
model = "text-embedding-3-small"   # Embedding model
base_url = ""                      # Custom base URL (for local/custom endpoints)
max_batch_size = 50                # Maximum items per batch request
timeout_seconds = 30               # Request timeout
cache_embeddings = true            # Cache embeddings for repeated content

# Advanced Settings (Optional)
[agent_memory.advanced]
retry_max_attempts = 3             # Maximum retry attempts for failed operations
retry_base_delay = "100ms"         # Base delay between retries
retry_max_delay = "5s"             # Maximum delay between retries
connection_pool_size = 25          # Database connection pool size
health_check_interval = "1m"       # Health check interval for connections
```

### Environment Variables

```bash
# Database connection
export AGENTFLOW_DB_URL="postgres://user:password@localhost:5432/agentflow"

# Embedding service
export OPENAI_API_KEY="your-openai-api-key"
export AZURE_OPENAI_ENDPOINT="https://your-resource.openai.azure.com/"
export AZURE_OPENAI_API_KEY="your-azure-api-key"

# Optional: Custom configuration file
export AGENTFLOW_CONFIG_PATH="/path/to/agentflow.toml"
```

## 💾 Memory Providers

### 1. In-Memory Provider (`memory`)

**Best for**: Development, testing, temporary sessions

```go
config := core.AgentMemoryConfig{
    Provider:   "memory",
    Connection: "memory",
    Dimensions: 1536,
}
```

**Features**:
- ✅ All memory operations
- ✅ Session isolation
- ✅ RAG capabilities
- ❌ No persistence (data lost on restart)
- ❌ No scaling across instances

### 2. PostgreSQL + pgvector (`pgvector`)

**Best for**: Production deployments, persistent storage

```go
config := core.AgentMemoryConfig{
    Provider:   "pgvector",
    Connection: "postgres://user:password@localhost:5432/agentflow",
    Dimensions: 1536,
}
```

**Features**:
- ✅ Full persistence
- ✅ Excellent performance (~45ms queries)
- ✅ Enhanced retry logic with exponential backoff
- ✅ Advanced connection pooling
- ✅ Batch operations for efficiency
- ✅ ACID transactions
- ✅ Vector similarity search with indexing

**Setup Requirements**:
```sql
-- Install pgvector extension
CREATE EXTENSION IF NOT EXISTS vector;

-- Tables are created automatically by AgentFlow
```

### 3. Weaviate (`weaviate`) 

**Best for**: Large-scale vector operations, advanced search

```go
config := core.AgentMemoryConfig{
    Provider:   "weaviate",
    Connection: "http://localhost:8080",
    Dimensions: 1536,
}
```

**Note**: Currently in development. Basic implementation available.

## 📖 API Reference

### Personal Memory Operations

#### Store Memory
```go
// Store with tags
err := memory.Store(ctx, "I prefer dark roast coffee", "preference", "coffee")

// Store without tags
err := memory.Store(ctx, "Meeting scheduled for tomorrow")
```

#### Query Memory
```go
// Basic query
results, err := memory.Query(ctx, "coffee preferences", 5)

// Query with custom limit
results, err := memory.Query(ctx, "meetings")  // Uses default limit from config
```

### Key-Value Storage

#### Remember/Recall
```go
// Store structured data
err := memory.Remember(ctx, "user_preferences", map[string]interface{}{
    "theme": "dark",
    "language": "en",
    "timezone": "UTC",
})

// Retrieve data
prefs, err := memory.Recall(ctx, "user_preferences")
if prefs != nil {
    preferences := prefs.(map[string]interface{})
    theme := preferences["theme"].(string)
}
```

### Chat History

#### Add Messages
```go
// Add user message
err := memory.AddMessage(ctx, "user", "How do I configure memory in AgentFlow?")

// Add assistant response
err := memory.AddMessage(ctx, "assistant", "You can configure memory using the agent_memory section in your TOML file...")
```

#### Get History
```go
// Get all history
history, err := memory.GetHistory(ctx)

// Get recent N messages
recent, err := memory.GetHistory(ctx, 10)
```

### Knowledge Base Operations

#### Document Ingestion
```go
// Single document
doc := core.Document{
    ID:      "guide-001",
    Title:   "AgentFlow Memory Guide",
    Content: "AgentFlow provides a powerful memory system...",
    Source:  "docs/memory-guide.md",
    Type:    core.DocumentTypeText,
    Tags:    []string{"documentation", "memory"},
    Metadata: map[string]any{
        "author": "AgentFlow Team",
        "version": "1.0",
    },
}

err := memory.IngestDocument(ctx, doc)

// Multiple documents (more efficient)
docs := []core.Document{doc1, doc2, doc3}
err := memory.IngestDocuments(ctx, docs)
```

#### Knowledge Search
```go
// Basic search
results, err := memory.SearchKnowledge(ctx, "memory configuration")

// Advanced search with options
results, err := memory.SearchKnowledge(ctx, "vector databases",
    core.WithScoreThreshold(0.8),
    core.WithSources([]string{"docs/"}),
    core.WithTags([]string{"database", "vector"}),
    core.WithLimit(10),
)
```

### Advanced Search Options

```go
// Search options
type SearchOption func(*SearchConfig)

func WithLimit(limit int) SearchOption
func WithScoreThreshold(threshold float32) SearchOption
func WithSources(sources []string) SearchOption
func WithTags(tags []string) SearchOption
func WithDocumentTypes(types []DocumentType) SearchOption
func WithDateRange(start, end time.Time) SearchOption
func WithIncludePersonal(include bool) SearchOption
func WithIncludeKnowledge(include bool) SearchOption
```

## 🔄 RAG (Retrieval-Augmented Generation)

### Hybrid Search

Combine personal memory and knowledge base searches:

```go
// Hybrid search across both personal memory and knowledge base
result, err := memory.SearchAll(ctx, "database best practices")

fmt.Printf("Personal memories: %d\n", len(result.PersonalMemory))
fmt.Printf("Knowledge results: %d\n", len(result.Knowledge))
fmt.Printf("Total results: %d\n", result.TotalResults)
fmt.Printf("Search time: %v\n", result.SearchTime)
```

### RAG Context Building

Automatically assemble context for LLM prompts:

```go
// Build RAG context
ragContext, err := memory.BuildContext(ctx, "How do I optimize database performance?")

// Use in LLM prompt
prompt := fmt.Sprintf(`
Context:
%s

Question: %s

Please provide a comprehensive answer based on the context above.
`, ragContext.ContextText, ragContext.Query)

// ragContext contains:
// - PersonalMemory: Relevant personal memories
// - Knowledge: Relevant knowledge base entries  
// - ChatHistory: Recent conversation
// - ContextText: Formatted text ready for LLM
// - Sources: List of source documents
// - TokenCount: Estimated token count
```

### Context Configuration

```go
// Configure context building
ragContext, err := memory.BuildContext(ctx, "database question",
    core.WithMaxTokens(3000),
    core.WithPersonalWeight(0.4),      // 40% personal memory
    core.WithKnowledgeWeight(0.6),     // 60% knowledge base
    core.WithHistoryLimit(5),          // Include last 5 messages
    core.WithIncludeSources(true),     // Include source attribution
)
```

## 🚀 Advanced Features

### Batch Operations (PgVectorProvider)

For improved performance with large datasets:

```go
// Cast to access enhanced features
if provider, ok := memory.(*core.PgVectorProvider); ok {
    // Batch store multiple memories
    batchRequests := []core.BatchStoreRequest{
        {Content: "First item", Tags: []string{"batch", "test"}},
        {Content: "Second item", Tags: []string{"batch", "test"}},
        {Content: "Third item", Tags: []string{"batch", "test"}},
    }
    
    err := provider.BatchStore(ctx, batchRequests)
    
    // Batch ingest documents
    docs := []core.Document{doc1, doc2, doc3}
    err := provider.BatchIngestDocuments(ctx, docs)
}
```

### Retry Logic and Error Handling

The PgVectorProvider includes sophisticated retry logic:

```go
// Automatic retry with exponential backoff
// - 3 retry attempts by default
// - 100ms → 200ms → 400ms → 5s max delay
// - Intelligent error classification
// - Context cancellation support

// Configure retry behavior
retryConfig := core.DefaultRetryConfig()
retryConfig.MaxRetries = 5
retryConfig.BackoffDuration = 200 * time.Millisecond
retryConfig.MaxBackoff = 10 * time.Second
```

### Connection Pooling

Optimized database connections:

```go
// PgVectorProvider automatically configures:
// - MaxConns: 25 connections
// - MinConns: 5 connections  
// - Connection lifetime: 1 hour
// - Health checks: every minute
// - Statement timeouts: 30 seconds
```

### Session Management

```go
// Create new session
sessionID := memory.NewSession()

// Set session context
ctx := memory.SetSession(context.Background(), sessionID)

// Get current session ID
currentSession := core.GetSessionID(ctx)

// Clear all data for session
err := memory.ClearSession(ctx)
```

## 💡 Examples

### Example 1: Personal Assistant with Memory

```go
package main

import (
    "context"
    "fmt"
    "log"
    
    "github.com/kunalkushwaha/agentflow/core"
)

func personalAssistantExample() {
    // Setup memory
    config := core.AgentMemoryConfig{
        Provider:   "pgvector",
        Connection: "postgres://user:password@localhost:5432/agentflow",
        Dimensions: 1536,
        Embedding: core.EmbeddingConfig{
            Provider: "openai",
            APIKey:   "your-api-key",
            Model:    "text-embedding-3-small",
        },
    }
    
    memory, err := core.NewMemory(config)
    if err != nil {
        log.Fatal(err)
    }
    defer memory.Close()
    
    // User session
    ctx := memory.SetSession(context.Background(), "alice-123")
    
    // Store user preferences
    memory.Store(ctx, "I prefer morning meetings", "scheduling", "preference")
    memory.Store(ctx, "I work in Pacific Time Zone", "timezone", "preference")
    memory.Remember(ctx, "notification_preferences", map[string]interface{}{
        "email": true,
        "slack": false,
        "phone": false,
    })
    
    // User asks a question
    memory.AddMessage(ctx, "user", "Schedule a meeting for me")
    
    // Query relevant memories
    results, _ := memory.Query(ctx, "scheduling preferences", 3)
    
    // Build context for LLM
    ragContext, _ := memory.BuildContext(ctx, "schedule meeting preferences")
    
    // Generate response using LLM with context
    response := "Based on your preferences for morning meetings and Pacific Time Zone..."
    
    // Store assistant response
    memory.AddMessage(ctx, "assistant", response)
    
    fmt.Println("Personal assistant conversation stored in memory!")
}
```

### Example 2: Document-Based Q&A System

```go
func documentQAExample() {
    memory, err := core.NewMemory(config)
    if err != nil {
        log.Fatal(err)
    }
    defer memory.Close()
    
    ctx := context.Background()
    
    // Ingest documentation
    docs := []core.Document{
        {
            ID:      "memory-guide",
            Title:   "Memory System Guide",
            Content: "AgentFlow memory system provides persistent storage...",
            Source:  "docs/memory.md",
            Type:    core.DocumentTypeText,
            Tags:    []string{"documentation", "memory"},
        },
        {
            ID:      "config-guide", 
            Title:   "Configuration Guide",
            Content: "Configure AgentFlow using TOML files...",
            Source:  "docs/config.md",
            Type:    core.DocumentTypeText,
            Tags:    []string{"documentation", "configuration"},
        },
    }
    
    err = memory.IngestDocuments(ctx, docs)
    if err != nil {
        log.Fatal(err)
    }
    
    // Answer user questions using knowledge base
    questions := []string{
        "How do I configure memory?",
        "What providers are available?",
        "How does RAG work?",
    }
    
    for _, question := range questions {
        // Build RAG context
        ragContext, err := memory.BuildContext(ctx, question)
        if err != nil {
            log.Printf("Error building context: %v", err)
            continue
        }
        
        fmt.Printf("Question: %s\n", question)
        fmt.Printf("Context (%d tokens):\n%s\n", ragContext.TokenCount, ragContext.ContextText)
        fmt.Printf("Sources: %v\n\n", ragContext.Sources)
    }
}
```

### Example 3: Multi-User Chat Application

```go
func multiUserChatExample() {
    memory, err := core.NewMemory(config)
    if err != nil {
        log.Fatal(err)
    }
    defer memory.Close()
    
    // Different users with isolated sessions
    users := map[string]string{
        "alice": memory.NewSession(),
        "bob":   memory.NewSession(),
        "charlie": memory.NewSession(),
    }
    
    // Each user stores different preferences
    for username, sessionID := range users {
        userCtx := memory.SetSession(context.Background(), sessionID)
        
        memory.Store(userCtx, fmt.Sprintf("I am %s", username), "identity")
        memory.Remember(userCtx, "username", username)
        
        // Chat history is isolated per user
        memory.AddMessage(userCtx, "user", "Hello, I'm "+username)
        memory.AddMessage(userCtx, "assistant", "Hello "+username+"! How can I help you?")
    }
    
    // Query each user's isolated data
    for username, sessionID := range users {
        userCtx := memory.SetSession(context.Background(), sessionID)
        
        results, _ := memory.Query(userCtx, "identity", 1)
        storedUsername, _ := memory.Recall(userCtx, "username")
        history, _ := memory.GetHistory(userCtx)
        
        fmt.Printf("User: %s\n", username)
        fmt.Printf("  Memory: %s\n", results[0].Content)
        fmt.Printf("  Stored username: %s\n", storedUsername)
        fmt.Printf("  Messages: %d\n\n", len(history))
    }
}
```

## ⚡ Performance & Optimization

### Performance Benchmarks

Based on comprehensive testing:

| Operation | Provider | Performance | Notes |
|-----------|----------|-------------|--------|
| Single Store | PgVector | ~50ms | With retry logic |
| Batch Store (50 items) | PgVector | ~2.3s | ~22 ops/sec |
| Query (10 results) | PgVector | ~45ms | With vector similarity |
| Knowledge Search | PgVector | ~45ms | Indexed search |
| RAG Context Build | PgVector | ~90ms | Full hybrid search |
| Document Ingest | PgVector | ~150ms | Per document |
| Batch Document Ingest | PgVector | ~150ms | Per 3 documents |

### Optimization Tips

#### 1. Use Batch Operations
```go
// Instead of multiple Store() calls
for _, item := range items {
    memory.Store(ctx, item.Content, item.Tags...)  // Slower
}

// Use batch operations (PgVectorProvider)
if provider, ok := memory.(*core.PgVectorProvider); ok {
    provider.BatchStore(ctx, batchRequests)  // Much faster
}
```

#### 2. Configure Appropriate Limits
```toml
[agent_memory]
max_results = 10              # Don't fetch more than needed
knowledge_max_results = 20    # Limit knowledge base results
rag_max_context_tokens = 4000 # Balance context size vs performance
```

#### 3. Use Score Thresholds
```go
// Filter low-quality results
results, err := memory.SearchKnowledge(ctx, "query",
    core.WithScoreThreshold(0.7),  // Only high-relevance results
)
```

#### 4. Embedding Service Optimization
```toml
[agent_memory.embedding]
max_batch_size = 50           # Batch embedding requests
cache_embeddings = true       # Cache repeated content
timeout_seconds = 30          # Reasonable timeout
```

#### 5. Connection Pooling (PgVector)
```toml
[agent_memory.advanced]
connection_pool_size = 25     # Optimize for your workload
health_check_interval = "1m"  # Monitor connection health
```

### Memory Usage Guidelines

#### Session Management
```go
// Clean up unused sessions periodically
err := memory.ClearSession(ctx)  // Removes all session data
```

#### Document Chunking
```toml
[agent_memory]
chunk_size = 1000     # Smaller chunks = more granular search
chunk_overlap = 200   # Overlap preserves context across chunks
```

## 🔧 Troubleshooting

### Common Issues

#### 1. Connection Errors
```go
// Error: failed to ping database
// Solution: Check connection string and database accessibility
config := core.AgentMemoryConfig{
    Provider:   "pgvector",
    Connection: "postgres://user:password@localhost:5432/agentflow?sslmode=disable",
}
```

#### 2. Embedding Service Errors
```go
// Error: failed to generate embedding
// Solution: Verify API key and provider configuration
config.Embedding = core.EmbeddingConfig{
    Provider: "openai",
    APIKey:   os.Getenv("OPENAI_API_KEY"),  // Ensure API key is set
    Model:    "text-embedding-3-small",
}
```

#### 3. Low Search Scores
```go
// Issue: All similarity scores are very low
// Solution: Use appropriate score thresholds for your embedding provider

// For dummy embeddings (testing)
core.WithScoreThreshold(-0.1)

// For real embeddings
core.WithScoreThreshold(0.7)
```

#### 4. Performance Issues
```go
// Issue: Slow queries
// Solutions:
// 1. Reduce result limits
results, _ := memory.Query(ctx, "query", 5)  // Instead of 50

// 2. Use batch operations
provider.BatchStore(ctx, requests)  // Instead of individual stores

// 3. Optimize embedding batch size
config.Embedding.MaxBatchSize = 25  // Tune for your provider
```

### Debugging Tips

#### Enable Debug Logging
```go
import "log"

// Set up debug logging
log.SetFlags(log.LstdFlags | log.Lshortfile)
```

#### Check Session Isolation
```go
// Verify session ID
sessionID := core.GetSessionID(ctx)
fmt.Printf("Current session: %s\n", sessionID)

// Test with known session
ctx = memory.SetSession(context.Background(), "test-session")
```

#### Validate Configuration
```go
// Test basic memory operations
err := memory.Store(ctx, "test content", "test")
if err != nil {
    log.Printf("Store failed: %v", err)
}

results, err := memory.Query(ctx, "test", 1)
if err != nil {
    log.Printf("Query failed: %v", err)
} else {
    log.Printf("Found %d results", len(results))
}
```

### Database Setup (PgVector)

#### PostgreSQL with pgvector
```sql
-- Install pgvector extension
CREATE EXTENSION IF NOT EXISTS vector;

-- Verify installation
SELECT * FROM pg_extension WHERE extname = 'vector';

-- Check tables (created automatically by AgentFlow)
\dt

-- Tables created:
-- - personal_memory (session-scoped memories)
-- - key_value_store (session-scoped key-value pairs)
-- - chat_history (conversation messages)
-- - documents (document metadata)
-- - knowledge_base (document embeddings)
```

#### Docker Setup
```bash
# Run PostgreSQL with pgvector
docker run -d \
  --name agentflow-postgres \
  -e POSTGRES_DB=agentflow \
  -e POSTGRES_USER=agentflow \
  -e POSTGRES_PASSWORD=password \
  -p 5432:5432 \
  pgvector/pgvector:pg16

# Test connection
psql "postgres://agentflow:password@localhost:5432/agentflow"
```

### Environment Variables Reference

```bash
# Required
export AGENTFLOW_DB_URL="postgres://user:password@localhost:5432/agentflow"
export OPENAI_API_KEY="your-openai-api-key"

# Optional
export AGENTFLOW_CONFIG_PATH="/path/to/agentflow.toml"
export AGENTFLOW_LOG_LEVEL="debug"
export AGENTFLOW_EMBEDDING_CACHE="true"

# Azure OpenAI (if using Azure)
export AZURE_OPENAI_ENDPOINT="https://your-resource.openai.azure.com/"
export AZURE_OPENAI_API_KEY="your-azure-api-key"
export AZURE_OPENAI_DEPLOYMENT_NAME="your-deployment-name"

# Testing
export AGENTFLOW_TEST_DB_URL="postgres://agentflow:password@localhost:5432/agentflow_test"
```

## 📊 Monitoring & Metrics

### Built-in Metrics

AgentFlow memory operations provide timing and performance metrics:

```go
// Hybrid search includes timing
result, err := memory.SearchAll(ctx, "query")
fmt.Printf("Search took: %v\n", result.SearchTime)
fmt.Printf("Total results: %d\n", result.TotalResults)

// RAG context includes token counting
ragContext, err := memory.BuildContext(ctx, "query")
fmt.Printf("Context tokens: %d\n", ragContext.TokenCount)
```

### Health Checks

```go
// Test memory health
func checkMemoryHealth(memory core.Memory) error {
    ctx := memory.SetSession(context.Background(), "health-check")
    
    // Test basic operations
    err := memory.Store(ctx, "health check", "test")
    if err != nil {
        return fmt.Errorf("store failed: %w", err)
    }
    
    results, err := memory.Query(ctx, "health", 1)
    if err != nil {
        return fmt.Errorf("query failed: %w", err)
    }
    
    if len(results) == 0 {
        return fmt.Errorf("no results returned")
    }
    
    // Cleanup
    memory.ClearSession(ctx)
    
    return nil
}
```

---

## 🎯 Summary

The AgentFlow Memory System provides:

✅ **Unified Interface**: Single API for all memory operations  
✅ **Multiple Providers**: In-memory, PostgreSQL+pgvector, Weaviate  
✅ **Session Isolation**: Complete data separation between users/sessions  
✅ **RAG Capabilities**: Document ingestion, knowledge search, context building  
✅ **Advanced Features**: Batch operations, retry logic, connection pooling  
✅ **Production Ready**: Comprehensive error handling, performance optimization  
✅ **Easy Configuration**: TOML-based configuration with environment variable support  
✅ **Comprehensive Testing**: Full test coverage with integration tests  

The system is designed for production use with enterprise-grade features while maintaining simplicity for development and testing scenarios.

For more examples and advanced usage, see the `/examples` directory in the AgentFlow repository.

For detailed RAG configuration options, see the **[RAG Configuration Guide](RAGConfiguration.md)**.
