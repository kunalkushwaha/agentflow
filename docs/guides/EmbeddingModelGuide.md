# Embedding Model Intelligence Guide

AgentFlow includes an intelligent embedding model system that automatically configures appropriate settings based on your embedding model choice. This guide explains how to use and benefit from this system.

## Overview

The Embedding Model Intelligence system:
- **Automatically configures dimensions** based on your chosen embedding model
- **Validates compatibility** between embedding models and memory providers
- **Provides recommendations** for optimal model selection
- **Offers troubleshooting guidance** for common issues

## Supported Models

### Ollama Models (Recommended for Local Development)

#### nomic-embed-text:latest ⭐ **Recommended**
- **Dimensions:** 768
- **Provider:** ollama
- **Notes:** Excellent general-purpose embedding model with good performance
- **Best for:** Local development, privacy-focused applications

```bash
# Install and use
ollama pull nomic-embed-text:latest
agentcli create myproject --memory-enabled --embedding-provider ollama --embedding-model nomic-embed-text:latest
```

#### mxbai-embed-large
- **Dimensions:** 1024
- **Provider:** ollama
- **Notes:** Larger model with better quality, requires more resources
- **Best for:** Applications requiring higher embedding quality

```bash
ollama pull mxbai-embed-large
agentcli create myproject --memory-enabled --embedding-provider ollama --embedding-model mxbai-embed-large
```

#### all-minilm
- **Dimensions:** 384
- **Provider:** ollama
- **Notes:** Lightweight and fast, good for development
- **Best for:** Resource-constrained environments, rapid prototyping

```bash
ollama pull all-minilm
agentcli create myproject --memory-enabled --embedding-provider ollama --embedding-model all-minilm
```

### OpenAI Models (Production Ready)

#### text-embedding-3-small ⭐ **Recommended**
- **Dimensions:** 1536
- **Provider:** openai
- **Notes:** Cost-effective OpenAI embedding model with good performance
- **Best for:** Production applications with budget considerations

```bash
export OPENAI_API_KEY="your-api-key"
agentcli create myproject --memory-enabled --embedding-provider openai --embedding-model text-embedding-3-small
```

#### text-embedding-3-large
- **Dimensions:** 3072
- **Provider:** openai
- **Notes:** Highest quality OpenAI embedding model, more expensive
- **Best for:** Applications requiring maximum embedding quality

```bash
agentcli create myproject --memory-enabled --embedding-provider openai --embedding-model text-embedding-3-large
```

#### text-embedding-ada-002 (Legacy)
- **Dimensions:** 1536
- **Provider:** openai
- **Notes:** Legacy OpenAI model, use text-embedding-3-small instead
- **Status:** Not recommended for new projects

### Testing Models

#### dummy
- **Dimensions:** 1536
- **Provider:** dummy
- **Notes:** Simple embeddings for testing only, not suitable for production
- **Best for:** Testing, development without external dependencies

## Automatic Configuration

When you create a project, the system automatically:

1. **Detects model dimensions:**
   ```bash
   agentcli create myproject --memory-enabled --embedding-model nomic-embed-text:latest
   # Automatically configures 768 dimensions
   ```

2. **Validates compatibility:**
   ```bash
   # System checks if the model works with your memory provider
   # Warns about potential issues
   ```

3. **Provides helpful information:**
   ```
   ✓ Using embedding model: nomic-embed-text:latest (768 dimensions)
     Excellent general-purpose embedding model with good performance
   ```

## Model Selection Guide

### For Local Development

**Recommended:** `nomic-embed-text:latest` with Ollama
- No API costs
- Good performance
- Privacy-friendly
- Easy to set up

```bash
# Setup
ollama serve
ollama pull nomic-embed-text:latest

# Create project
agentcli create myproject --memory-enabled \
  --memory-provider pgvector \
  --embedding-provider ollama \
  --embedding-model nomic-embed-text:latest
```

### For Production Applications

**Budget-conscious:** `text-embedding-3-small` with OpenAI
- Good quality-to-cost ratio
- Reliable and fast
- Well-supported

**High-quality:** `text-embedding-3-large` with OpenAI
- Best embedding quality
- Higher cost
- Suitable for critical applications

```bash
# Setup
export OPENAI_API_KEY="your-api-key"

# Create project
agentcli create myproject --memory-enabled \
  --memory-provider pgvector \
  --embedding-provider openai \
  --embedding-model text-embedding-3-small
```

### For Resource-Constrained Environments

**Lightweight:** `all-minilm` with Ollama
- Smallest dimensions (384)
- Fast processing
- Lower memory usage

```bash
agentcli create myproject --memory-enabled \
  --memory-provider memory \
  --embedding-provider ollama \
  --embedding-model all-minilm
```

## Compatibility Matrix

| Embedding Model | Dimensions | Memory Provider | Compatibility | Notes |
|----------------|------------|-----------------|---------------|-------|
| nomic-embed-text:latest | 768 | pgvector | ✅ Excellent | Recommended combination |
| nomic-embed-text:latest | 768 | weaviate | ✅ Excellent | Good for large scale |
| nomic-embed-text:latest | 768 | memory | ✅ Good | Development only |
| text-embedding-3-small | 1536 | pgvector | ✅ Excellent | Production ready |
| text-embedding-3-large | 3072 | pgvector | ⚠️ Good | May impact performance |
| text-embedding-3-large | 3072 | weaviate | ✅ Excellent | Handles large dimensions well |
| dummy | 1536 | weaviate | ❌ Poor | Not recommended |
| all-minilm | 384 | pgvector | ✅ Excellent | Very fast |

## Configuration Examples

### Complete Configuration Examples

#### Local Development Setup
```toml
[agent_memory]
provider = "pgvector"
connection = "postgres://user:password@localhost:5432/agentflow?sslmode=disable"
dimensions = 768
auto_embed = true
enable_rag = true
chunk_size = 1000
chunk_overlap = 100

[agent_memory.embedding]
provider = "ollama"
model = "nomic-embed-text:latest"
base_url = "http://localhost:11434"
cache_embeddings = true
timeout_seconds = 30
```

#### Production Setup
```toml
[agent_memory]
provider = "pgvector"
connection = "postgres://user:password@prod-db:5432/agentflow?sslmode=require"
dimensions = 1536
auto_embed = true
enable_rag = true
chunk_size = 1000
chunk_overlap = 100
knowledge_score_threshold = 0.75

[agent_memory.embedding]
provider = "openai"
model = "text-embedding-3-small"
cache_embeddings = true
max_batch_size = 100
timeout_seconds = 30
```

#### High-Performance Setup
```toml
[agent_memory]
provider = "weaviate"
connection = "http://weaviate:8080"
dimensions = 3072
auto_embed = true
enable_rag = true
chunk_size = 2000
chunk_overlap = 200

[agent_memory.embedding]
provider = "openai"
model = "text-embedding-3-large"
cache_embeddings = true
max_batch_size = 50
timeout_seconds = 60
```

## Validation and Troubleshooting

### Using the Intelligence System

The system provides automatic validation:

```bash
# Check if your model is recognized
agentcli create myproject --memory-enabled --embedding-model unknown-model
# Output: ⚠️ Unknown embedding model: ollama/unknown-model
#         💡 Recommended models for this provider:
#            • nomic-embed-text:latest (768 dimensions) - Excellent general-purpose...
```

### Manual Validation

```bash
# Validate your configuration
agentcli memory --validate

# Check current configuration
agentcli memory --config

# Test search with your embeddings
agentcli memory --search "test query"
```

### Common Issues and Solutions

#### Unknown Model Warning
```
⚠️ Unknown embedding model: ollama/custom-model
```

**Solution:** Use a recognized model or verify your custom model's dimensions:
```bash
# Check available models
agentcli create --help | grep -A 10 "EMBEDDING PROVIDERS"

# Or use a recommended model
agentcli create myproject --embedding-model nomic-embed-text:latest
```

#### Dimension Mismatch
```
❌ Configuration Error: nomic-embed-text requires 768 dimensions, but 1536 configured
```

**Solution:** The system automatically prevents this, but if you encounter it:
```toml
[agent_memory]
dimensions = 768  # Match your embedding model
```

#### Performance Warnings
```
⚠️ Large embedding dimensions (3072) may impact performance
```

**Solution:** Consider using a smaller model for better performance:
```bash
# Instead of text-embedding-3-large (3072 dims)
agentcli create myproject --embedding-model text-embedding-3-small  # 1536 dims
```

## Advanced Usage

### Custom Model Integration

If you need to use a custom embedding model:

1. **Add it to the intelligence system** (for developers):
   ```go
   // In internal/scaffold/embedding_intelligence.go
   "custom-provider": {
       "custom-model": {
           Provider: "custom-provider",
           Model: "custom-model",
           Dimensions: 512,
           Notes: "Custom embedding model",
       },
   }
   ```

2. **Use fallback dimensions:**
   ```bash
   # System will use provider defaults
   agentcli create myproject --embedding-provider ollama --embedding-model custom-model
   ```

### Performance Optimization

#### Dimension Selection
- **384 dimensions:** Fastest, lowest memory usage
- **768 dimensions:** Good balance of speed and quality
- **1536 dimensions:** Standard quality, moderate performance
- **3072 dimensions:** Highest quality, slower performance

#### Provider Selection
- **Ollama:** Best for local development, no API costs
- **OpenAI:** Best for production, consistent quality
- **Dummy:** Only for testing, no real embeddings

#### Memory Provider Pairing
- **Small dimensions (384-768) + PgVector:** Excellent performance
- **Large dimensions (1536-3072) + Weaviate:** Better handling of high-dimensional vectors
- **Any dimensions + Memory:** Fast but non-persistent

## Migration Between Models

### Changing Embedding Models

When switching embedding models:

1. **Update configuration:**
   ```toml
   [agent_memory]
   dimensions = 1536  # New model dimensions
   
   [agent_memory.embedding]
   model = "text-embedding-3-small"  # New model
   ```

2. **Handle existing data:**
   ```bash
   # Option 1: Clear and restart (loses data)
   agentcli memory --clear
   
   # Option 2: Re-embed existing content (custom script needed)
   # Option 3: Keep old embeddings (may reduce search quality)
   ```

3. **Test the new setup:**
   ```bash
   agentcli memory --validate
   agentcli memory --search "test query"
   ```

## Best Practices

### Model Selection
1. **Start with recommended models** (nomic-embed-text:latest or text-embedding-3-small)
2. **Consider your use case** (local vs. production, cost vs. quality)
3. **Test performance** with your specific data and queries
4. **Monitor resource usage** and adjust as needed

### Configuration
1. **Use the intelligence system** - let it configure dimensions automatically
2. **Validate your setup** with `agentcli memory --validate`
3. **Test thoroughly** before production deployment
4. **Monitor performance** and adjust settings as needed

### Development Workflow
1. **Start local** with Ollama and nomic-embed-text:latest
2. **Test functionality** with in-memory or pgvector
3. **Optimize configuration** based on your data
4. **Deploy to production** with appropriate model and provider

This guide should help you make informed decisions about embedding models and get the most out of AgentFlow's intelligent configuration system.