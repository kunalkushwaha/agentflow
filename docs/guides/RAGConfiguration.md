# RAG Configuration Guide

**Configuring Retrieval-Augmented Generation in AgentFlow**

AgentFlow provides comprehensive RAG (Retrieval-Augmented Generation) capabilities through flexible TOML configuration. This guide covers all configuration options for building knowledge-aware agents with document understanding and context assembly.

## 📚 Table of Contents

- [Overview](#overview)
- [Basic Configuration](#basic-configuration)
- [Advanced Configuration](#advanced-configuration)
- [Configuration Examples](#configuration-examples)
- [Best Practices](#best-practices)
- [Performance Tuning](#performance-tuning)
- [Migration Guide](#migration-guide)
- [Troubleshooting](#troubleshooting)

## 🎯 Overview

### What is RAG?

RAG (Retrieval-Augmented Generation) enhances LLM responses by:
- **Retrieving relevant information** from knowledge bases and personal memory
- **Augmenting prompts** with contextual information
- **Generating informed responses** based on retrieved knowledge
- **Providing source attribution** for transparency

### AgentFlow RAG Features

- **Hybrid Search**: Combines semantic similarity and keyword matching
- **Document Processing**: Automatic chunking and metadata extraction
- **Context Assembly**: Intelligent context building with token management
- **Session Memory**: Isolated memory per user or conversation
- **Multiple Providers**: Support for various vector databases

## ⚙️ Basic Configuration

### Minimal RAG Setup

```toml
[memory]
enabled = true
provider = "memory"
dimensions = 1536

[memory.embedding]
provider = "openai"
model = "text-embedding-3-small"

[memory.rag]
enabled = true
chunk_size = 1000
top_k = 5
```

### Core Memory Settings

```toml
[memory]
enabled = true
provider = "pgvector"                    # Options: memory, pgvector, weaviate
connection = "postgres://user:password@localhost:5432/agentflow"
max_results = 10                         # Maximum personal memory results
dimensions = 1536                        # Vector embedding dimensions
auto_embed = true                        # Automatically generate embeddings
```

### RAG Configuration

```toml
[memory.rag]
enabled = true                           # Enable RAG functionality
chunk_size = 1000                        # Document chunk size in characters
overlap = 200                            # Overlap between chunks
top_k = 5                               # Number of results to retrieve
score_threshold = 0.7                   # Minimum similarity score
hybrid_search = true                     # Enable hybrid search
session_memory = true                    # Enable session isolation
```

## 🔧 Advanced Configuration

### Document Processing

```toml
[memory.documents]
auto_chunk = true                        # Automatically chunk documents
supported_types = ["pdf", "txt", "md", "web", "code", "json"]
max_file_size = "10MB"                  # Maximum file size
enable_metadata_extraction = true        # Extract document metadata
enable_url_scraping = true              # Enable web content scraping
```

### Embedding Configuration

```toml
[memory.embedding]
provider = "openai"                      # Options: openai, azure, ollama, dummy
model = "text-embedding-3-small"        # Embedding model
api_key = "${OPENAI_API_KEY}"           # API key (environment variable)
base_url = ""                           # Custom endpoint (optional)
max_batch_size = 50                     # Batch size for embeddings
timeout_seconds = 30                    # Request timeout
cache_embeddings = true                 # Cache embeddings for performance
```

### Search Configuration

```toml
[memory.search]
hybrid_search = true                     # Enable hybrid search
keyword_weight = 0.3                    # Weight for keyword search (BM25)
semantic_weight = 0.7                   # Weight for semantic search
enable_reranking = false                # Enable result re-ranking
reranking_model = ""                    # Re-ranking model (optional)
enable_query_expansion = false          # Enable query expansion
```

### Context Assembly

```toml
[memory.context]
max_tokens = 4000                       # Maximum context tokens
personal_weight = 0.3                   # Weight for personal memory
knowledge_weight = 0.7                  # Weight for knowledge base
include_sources = true                  # Include source attribution
include_history = true                  # Include chat history
history_limit = 5                       # Number of history messages
```

## 📋 Configuration Examples

### Development Configuration

```toml
# agentflow.dev.toml - Development setup
[memory]
enabled = true
provider = "memory"                      # In-memory for fast iteration
max_results = 10
dimensions = 1536
auto_embed = true

[memory.embedding]
provider = "dummy"                       # Dummy embeddings for testing
model = "text-embedding-3-small"
cache_embeddings = true

[memory.rag]
enabled = true
chunk_size = 800                        # Smaller chunks for testing
overlap = 150
top_k = 3                               # Fewer results for faster testing
score_threshold = 0.5                   # Lower threshold for development
hybrid_search = false                   # Disable for simplicity
session_memory = false

[memory.documents]
auto_chunk = true
supported_types = ["txt", "md", "code"]
max_file_size = "5MB"
enable_metadata_extraction = false      # Disable for speed
enable_url_scraping = false
```

### Production Configuration

```toml
# agentflow.prod.toml - Production deployment
[memory]
enabled = true
provider = "pgvector"
connection = "postgres://user:password@localhost:5432/agentflow?sslmode=require"
max_results = 10
dimensions = 1536
auto_embed = true

[memory.pgvector]
table_name = "agent_memory"
connection_pool_size = 25

[memory.embedding]
provider = "openai"
model = "text-embedding-3-small"
api_key = "${OPENAI_API_KEY}"
cache_embeddings = true
max_batch_size = 100
timeout_seconds = 30

[memory.rag]
enabled = true
chunk_size = 1000
overlap = 200
top_k = 5
score_threshold = 0.7
hybrid_search = true
session_memory = true

[memory.documents]
auto_chunk = true
supported_types = ["pdf", "txt", "md", "web", "code", "json"]
max_file_size = "10MB"
enable_metadata_extraction = true
enable_url_scraping = true

[memory.context]
max_tokens = 4000
personal_weight = 0.3
knowledge_weight = 0.7
include_sources = true
include_history = true
history_limit = 5

[memory.search]
hybrid_search = true
keyword_weight = 0.3
semantic_weight = 0.7
enable_reranking = false
enable_query_expansion = false

[memory.advanced]
retry_max_attempts = 3
retry_base_delay = "100ms"
retry_max_delay = "5s"
health_check_interval = "1m"
```

### Enterprise Configuration

```toml
# agentflow.enterprise.toml - Large-scale deployment
[memory]
enabled = true
provider = "weaviate"
connection = "https://weaviate.company.com:8080"
max_results = 15
dimensions = 1536
auto_embed = true

[memory.weaviate]
api_key = "${WEAVIATE_API_KEY}"
class_name = "AgentMemory"
timeout = "30s"
max_retries = 3

[memory.embedding]
provider = "azure"
model = "text-embedding-ada-002"
api_key = "${AZURE_OPENAI_API_KEY}"
endpoint = "${AZURE_OPENAI_ENDPOINT}"
cache_embeddings = true
max_batch_size = 100
timeout_seconds = 45

[memory.rag]
enabled = true
chunk_size = 1200
overlap = 240
top_k = 8
score_threshold = 0.75
hybrid_search = true
session_memory = true

[memory.context]
max_tokens = 6000                       # Larger context for enterprise
personal_weight = 0.2
knowledge_weight = 0.8                  # Knowledge-focused
include_sources = true
include_history = true
history_limit = 10

[memory.search]
hybrid_search = true
keyword_weight = 0.25
semantic_weight = 0.75
enable_reranking = true
reranking_model = "cross-encoder/ms-marco-MiniLM-L-12-v2"
enable_query_expansion = true
```

## 🚀 Best Practices

### Provider Selection

| Use Case | Recommended Provider | Reason |
|----------|---------------------|---------|
| **Development** | `memory` | Fast iteration, no setup |
| **Production** | `pgvector` | Reliable, performant, mature |
| **Enterprise** | `weaviate` | Advanced features, clustering |
| **Prototyping** | `memory` | Quick testing, temporary data |

### Chunking Strategy

```toml
# Document type specific chunking
[memory.rag]
# For technical documentation
chunk_size = 1000
overlap = 200

# For code files
chunk_size = 800
overlap = 100

# For research papers
chunk_size = 1500
overlap = 300
```

**Guidelines:**
- **Small documents** (< 5KB): chunk_size = 500-800, overlap = 100-150
- **Medium documents** (5-50KB): chunk_size = 1000-1200, overlap = 200-250
- **Large documents** (> 50KB): chunk_size = 1200-1500, overlap = 250-300
- **Code files**: chunk_size = 600-1000, overlap = 50-100

### Context Weighting

```toml
[memory.context]
# Knowledge-focused (documentation, Q&A)
personal_weight = 0.2
knowledge_weight = 0.8

# Balanced (general purpose)
personal_weight = 0.5
knowledge_weight = 0.5

# Personal-focused (assistant, preferences)
personal_weight = 0.7
knowledge_weight = 0.3
```

### Score Thresholds

```toml
[memory.rag]
# High precision (strict matching)
score_threshold = 0.8

# Balanced (recommended)
score_threshold = 0.7

# High recall (permissive)
score_threshold = 0.5

# Development/testing
score_threshold = 0.3
```

### Context Limits by Model

```toml
[memory.context]
# GPT-3.5 Turbo
max_tokens = 3000

# GPT-4
max_tokens = 6000

# GPT-4 Turbo
max_tokens = 8000

# Claude 3
max_tokens = 10000
```

## ⚡ Performance Tuning

### Memory Provider Optimization

```toml
[memory]
# Reduce for faster queries
max_results = 5

[memory.rag]
# Optimize retrieval
top_k = 3                               # Fewer results = faster
score_threshold = 0.8                   # Higher threshold = fewer results
```

### Embedding Optimization

```toml
[memory.embedding]
cache_embeddings = true                 # Essential for performance
max_batch_size = 100                    # Larger batches = fewer API calls
timeout_seconds = 30                    # Reasonable timeout

# Provider-specific optimizations
[memory.embedding.openai]
max_requests_per_minute = 3000          # Respect rate limits

[memory.embedding.azure]
deployment_name = "text-embedding-ada-002"
api_version = "2023-05-15"
```

### Search Optimization

```toml
[memory.search]
hybrid_search = true                    # Best balance of speed/accuracy
keyword_weight = 0.3                    # Adjust based on query types
semantic_weight = 0.7                   # Higher for conceptual queries
enable_reranking = false                # Disable if not needed (faster)
```

### Database Optimization

```toml
# PostgreSQL + pgvector
[memory.pgvector]
connection = "postgres://user:password@localhost:5432/agentflow?pool_max_conns=25&pool_min_conns=5"

# Weaviate
[memory.weaviate]
timeout = "15s"                         # Shorter timeout for faster failures
max_retries = 2                         # Fewer retries
```

## 🔄 Migration Guide

### From Basic to RAG Configuration

#### Step 1: Enable RAG

```toml
# Add to existing configuration
[memory.rag]
enabled = true
chunk_size = 1000
top_k = 5
score_threshold = 0.7
```

#### Step 2: Configure Document Processing

```toml
[memory.documents]
auto_chunk = true
supported_types = ["txt", "md", "pdf"]
max_file_size = "10MB"
```

#### Step 3: Set Up Embeddings

```toml
[memory.embedding]
provider = "openai"
model = "text-embedding-3-small"
api_key = "${OPENAI_API_KEY}"
```

#### Step 4: Test Configuration

```bash
# Test with AgentFlow CLI
agentcli create test-rag --memory-enabled --rag-enabled --memory-provider pgvector

# Verify configuration
cd test-rag
go run . -m "Test RAG functionality"
```

### From Memory to PgVector

```toml
# Before
[memory]
provider = "memory"

# After
[memory]
provider = "pgvector"
connection = "postgres://user:password@localhost:5432/agentflow"

[memory.pgvector]
table_name = "agent_memory"
```

### Adding Hybrid Search

```toml
# Add to existing RAG configuration
[memory.search]
hybrid_search = true
keyword_weight = 0.3
semantic_weight = 0.7
```

## 🔧 Troubleshooting

### Common Configuration Issues

#### 1. Invalid Weight Configuration

```toml
# ❌ Wrong - weights don't sum to 1.0
[memory.context]
personal_weight = 0.5
knowledge_weight = 0.8

# ✅ Correct - weights sum to 1.0
[memory.context]
personal_weight = 0.3
knowledge_weight = 0.7
```

#### 2. Chunk Overlap Too Large

```toml
# ❌ Wrong - overlap >= chunk_size
[memory.rag]
chunk_size = 1000
overlap = 1000

# ✅ Correct - overlap < chunk_size
[memory.rag]
chunk_size = 1000
overlap = 200
```

#### 3. Missing Environment Variables

```bash
# Check required variables
echo $OPENAI_API_KEY
echo $DATABASE_URL

# Set if missing
export OPENAI_API_KEY="your-api-key"
export DATABASE_URL="postgres://user:password@localhost:5432/agentflow"
```

### Validation Errors

AgentFlow automatically validates configurations:

```bash
# Common validation messages
"RAG weights must sum to 1.0 (current: 1.3)"
"Chunk overlap must be less than chunk size"
"Score threshold must be between 0.0 and 1.0"
"Provider 'invalid' not supported"
```

### Performance Issues

#### Slow RAG Queries

```toml
# Reduce context size
[memory.context]
max_tokens = 2000                       # Smaller context

[memory.rag]
top_k = 3                              # Fewer results
score_threshold = 0.8                  # Higher threshold
```

#### High Memory Usage

```toml
# Optimize embedding cache
[memory.embedding]
cache_embeddings = false               # Disable if memory constrained
max_batch_size = 25                    # Smaller batches
```

### Debug Configuration

```toml
# Enable debug logging
[logging]
level = "debug"

# Test configuration
[memory.advanced]
health_check_interval = "30s"          # More frequent health checks
```

## 📊 Configuration Validation

AgentFlow provides automatic validation with helpful error messages:

### Weight Validation
- RAG weights (personal + knowledge) must sum to 1.0
- Search weights (keyword + semantic) must sum to 1.0

### Range Validation
- Score thresholds must be between 0.0 and 1.0
- Chunk overlap must be less than chunk size
- Token limits must be positive integers

### Provider Validation
- Ensures provider-specific settings are correct
- Validates connection strings and API keys
- Checks for required dependencies

## 🎛️ Environment Variables

### Required Variables

```bash
# LLM Provider
export OPENAI_API_KEY="your-openai-api-key"
# or
export AZURE_OPENAI_API_KEY="your-azure-api-key"
export AZURE_OPENAI_ENDPOINT="https://your-resource.openai.azure.com/"

# Database (if using pgvector)
export DATABASE_URL="postgres://user:password@localhost:5432/agentflow"

# Weaviate (if using weaviate)
export WEAVIATE_URL="http://localhost:8080"
export WEAVIATE_API_KEY="your-weaviate-api-key"
```

### Optional Variables

```bash
# Ollama (for local embeddings)
export OLLAMA_BASE_URL="http://localhost:11434"
export OLLAMA_MODEL="mxbai-embed-large"

# Configuration
export AGENTFLOW_CONFIG_PATH="/path/to/agentflow.toml"
export AGENTFLOW_LOG_LEVEL="debug"
```

### Using in Configuration

```toml
[memory.embedding]
api_key = "${OPENAI_API_KEY}"
endpoint = "${AZURE_OPENAI_ENDPOINT}"

[memory.pgvector]
connection = "${DATABASE_URL}"
```

## 📚 Complete Configuration Reference

```toml
[memory]
# Core settings
enabled = true
provider = "memory|pgvector|weaviate"
connection = "connection-string"
max_results = 10
dimensions = 1536
auto_embed = true

# RAG settings
[memory.rag]
enabled = true
chunk_size = 1000
overlap = 200
top_k = 5
score_threshold = 0.7
hybrid_search = true
session_memory = true

# Document processing
[memory.documents]
auto_chunk = true
supported_types = ["pdf", "txt", "md", "web", "code", "json"]
max_file_size = "10MB"
enable_metadata_extraction = true
enable_url_scraping = true

# Embedding service
[memory.embedding]
provider = "openai|azure|ollama|dummy"
model = "text-embedding-3-small"
api_key = "${API_KEY}"
endpoint = "${ENDPOINT}"
cache_embeddings = true
max_batch_size = 50
timeout_seconds = 30

# Search configuration
[memory.search]
hybrid_search = true
keyword_weight = 0.3
semantic_weight = 0.7
enable_reranking = false
reranking_model = ""
enable_query_expansion = false

# Context assembly
[memory.context]
max_tokens = 4000
personal_weight = 0.3
knowledge_weight = 0.7
include_sources = true
include_history = true
history_limit = 5

# Advanced settings
[memory.advanced]
retry_max_attempts = 3
retry_base_delay = "100ms"
retry_max_delay = "5s"
connection_pool_size = 25
health_check_interval = "1m"
```

---

## 🎯 Summary

AgentFlow's RAG configuration system provides:

✅ **Flexible Configuration**: TOML-based configuration with environment variable support  
✅ **Multiple Providers**: Support for in-memory, PostgreSQL, and Weaviate backends  
✅ **Advanced Features**: Hybrid search, context assembly, session isolation  
✅ **Performance Tuning**: Extensive optimization options for production use  
✅ **Validation**: Built-in configuration validation with helpful error messages  
✅ **Production Ready**: Enterprise-grade features with monitoring and health checks  

The RAG system enables building sophisticated knowledge-aware agents that can understand documents, maintain context, and provide informed responses with source attribution.

For more information:
- **[Memory System Guide](Memory.md)** - Complete memory API reference
- **[Memory Provider Setup](MemoryProviderSetup.md)** - Provider installation guides
- **[Configuration Guide](Configuration.md)** - General configuration patterns