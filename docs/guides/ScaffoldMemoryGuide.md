# Scaffold Memory System Guide

This guide explains how to use AgentFlow's enhanced scaffold system to create projects with intelligent memory configuration. The scaffold now includes automatic embedding model intelligence, configuration validation, and comprehensive troubleshooting support.

## Quick Start

### Basic Memory-Enabled Project

```bash
# Create a project with intelligent defaults
agentcli create myproject --memory-enabled

# The system automatically:
# - Selects nomic-embed-text:latest (768 dimensions)
# - Configures pgvector as memory provider
# - Sets up RAG with optimal parameters
# - Generates validation and troubleshooting code
```

### Production-Ready Setup

```bash
# Create a production-ready project
agentcli create myapp --memory-enabled \
  --memory-provider pgvector \
  --embedding-provider ollama \
  --embedding-model nomic-embed-text:latest \
  --rag-enabled \
  --hybrid-search \
  --session-memory
```

## Embedding Model Intelligence

The scaffold system now includes intelligent embedding model configuration:

### Automatic Dimension Detection

```bash
# System automatically detects dimensions
agentcli create myproject --memory-enabled --embedding-model nomic-embed-text:latest
# Output: ✓ Using embedding model: nomic-embed-text:latest (768 dimensions)
#         Excellent general-purpose embedding model with good performance
```

### Model Recommendations

```bash
# Get suggestions for unknown models
agentcli create myproject --memory-enabled --embedding-model unknown-model
# Output: ⚠️ Unknown embedding model: ollama/unknown-model
#         💡 Recommended models for this provider:
#            • nomic-embed-text:latest (768 dimensions) - Excellent general-purpose...
#            • mxbai-embed-large (1024 dimensions) - Larger model with better quality...
```

### Compatibility Validation

```bash
# System validates model compatibility
agentcli create myproject --memory-enabled \
  --memory-provider weaviate \
  --embedding-provider dummy
# Output: ⚠️ Compatibility warning: dummy embeddings are not recommended with Weaviate
```

## Enhanced Project Generation

### What Gets Generated

When you create a memory-enabled project, the scaffold generates:

1. **Intelligent Configuration** (`agentflow.toml`)
   - Correct `[agent_memory]` structure (not old `[memory]`)
   - Auto-calculated dimensions based on embedding model
   - Provider-specific optimizations

2. **Validated Main Code** (`main.go`)
   - Configuration validation at startup
   - Helpful error messages with troubleshooting steps
   - Graceful fallback handling

3. **Enhanced Agent Code** (`agent*.go`)
   - Robust memory integration with error handling
   - RAG context building with fallback behavior
   - Session management with validation

4. **Database Setup** (for pgvector/weaviate)
   - Docker Compose with correct vector dimensions
   - Database initialization scripts
   - Setup scripts for easy deployment

5. **Comprehensive Documentation** (`README.md`)
   - Setup instructions specific to your configuration
   - Troubleshooting guide for your providers
   - Performance optimization tips

### Generated Configuration Structure

The scaffold generates the modern configuration format:

```toml
# Generated agentflow.toml
[agent_memory]
provider = "pgvector"
connection = "postgres://user:password@localhost:5432/agentflow?sslmode=disable"
dimensions = 768  # Auto-calculated from embedding model
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

[agent_memory.search]
hybrid_search = true
keyword_weight = 0.3
semantic_weight = 0.7
```

## Configuration Validation

### Startup Validation

Generated projects include comprehensive validation:

```go
// Generated in main.go
func validateMemoryConfig(memoryConfig core.AgentMemoryConfig, expectedModel string) error {
    // Validate embedding dimensions
    expectedDimensions := 768
    if memoryConfig.Dimensions != expectedDimensions {
        return fmt.Errorf("%s requires %d dimensions, but %d configured\n💡 Solution: Update [agent_memory] dimensions = %d", 
            expectedModel, expectedDimensions, memoryConfig.Dimensions, expectedDimensions)
    }
    
    // Validate provider compatibility
    // ... comprehensive validation logic
}
```

### Runtime Error Handling

```go
// Generated error handling with specific troubleshooting
if err := memory.Store(ctx, testContent, "system-init"); err != nil {
    switch memoryConfig.Provider {
    case "pgvector":
        fmt.Printf("💡 PostgreSQL/PgVector Troubleshooting:\n")
        fmt.Printf("   1. Start database: docker compose up -d\n")
        fmt.Printf("   2. Run setup script: ./setup.sh\n")
        // ... more specific guidance
    }
}
```

## Memory Debug Integration

### Built-in Debug Support

Every generated project works with the memory debug command:

```bash
# Navigate to your generated project
cd myproject

# Debug your memory system
agentcli memory --validate
agentcli memory --stats
agentcli memory --search "test query"
```

### Project-Specific Troubleshooting

The generated README includes troubleshooting specific to your configuration:

```markdown
## Troubleshooting

### Memory System Issues

**Ollama Connection Failed:**
1. Start Ollama: `ollama serve`
2. Pull model: `ollama pull nomic-embed-text:latest`
3. Test connection: `curl http://localhost:11434/api/tags`

**Database Connection Failed:**
1. Start database: `docker compose up -d`
2. Run setup: `./setup.sh`
3. Check connection: `psql -h localhost -U user -d agentflow`
```

## Advanced Configuration Options

### Interactive Setup

```bash
# Use interactive mode for guided setup
agentcli create --interactive

# The system will guide you through:
# 1. Project name and basic settings
# 2. Memory provider selection with explanations
# 3. Embedding model selection with recommendations
# 4. RAG configuration with optimal defaults
# 5. Advanced features (hybrid search, sessions, etc.)
```

### Command Line Options

```bash
# Full configuration example
agentcli create myproject \
  --memory-enabled \
  --memory-provider pgvector \
  --embedding-provider ollama \
  --embedding-model nomic-embed-text:latest \
  --rag-enabled \
  --rag-chunk-size 1000 \
  --rag-overlap 100 \
  --rag-top-k 5 \
  --rag-score-threshold 0.7 \
  --hybrid-search \
  --session-memory
```

### Provider-Specific Optimizations

#### PostgreSQL/PgVector
```bash
agentcli create myproject --memory-enabled --memory-provider pgvector
# Generates:
# - Docker Compose with pgvector extension
# - Optimized connection pooling settings
# - Vector index creation scripts
# - Performance monitoring queries
```

#### Weaviate
```bash
agentcli create myproject --memory-enabled --memory-provider weaviate
# Generates:
# - Weaviate Docker configuration
# - Schema definitions for your embedding dimensions
# - Batch import optimization settings
# - Backup and restore scripts
```

#### In-Memory (Development)
```bash
agentcli create myproject --memory-enabled --memory-provider memory
# Generates:
# - Fast startup configuration
# - Development-optimized settings
# - Migration path to persistent storage
```

## Migration from Old Projects

### Automatic Detection

The scaffold can detect and help migrate old projects:

```bash
# If you have an old project with [memory] configuration
agentcli create myproject-new --memory-enabled
# Copy your agent logic from the old project
# The new project will have the correct configuration structure
```

### Configuration Migration

Old format:
```toml
[memory]  # ❌ Old format
provider = "pgvector"
dimensions = 1536
```

New format:
```toml
[agent_memory]  # ✅ New format
provider = "pgvector"
dimensions = 768  # Auto-calculated
connection = "postgres://..."

[agent_memory.embedding]
provider = "ollama"
model = "nomic-embed-text:latest"
```

## Best Practices

### Development Workflow

1. **Start Local:**
   ```bash
   agentcli create myproject --memory-enabled
   # Uses Ollama + nomic-embed-text by default
   ```

2. **Test Configuration:**
   ```bash
   cd myproject
   agentcli memory --validate
   ```

3. **Develop and Test:**
   ```bash
   go mod tidy
   go run . -m "test message"
   ```

4. **Debug Issues:**
   ```bash
   agentcli memory --stats
   agentcli memory --search "your query"
   ```

### Production Deployment

1. **Create Production Project:**
   ```bash
   agentcli create myapp-prod --memory-enabled \
     --memory-provider pgvector \
     --embedding-provider openai \
     --embedding-model text-embedding-3-small
   ```

2. **Configure Environment:**
   ```bash
   export OPENAI_API_KEY="your-key"
   # Update connection strings in agentflow.toml
   ```

3. **Deploy Database:**
   ```bash
   docker compose up -d
   ./setup.sh
   ```

4. **Validate Production Setup:**
   ```bash
   agentcli memory --validate
   agentcli memory --stats
   ```

### Performance Optimization

1. **Choose Appropriate Dimensions:**
   - 384 dims: Fastest (all-minilm)
   - 768 dims: Balanced (nomic-embed-text)
   - 1536 dims: Standard (OpenAI small)
   - 3072 dims: Highest quality (OpenAI large)

2. **Optimize RAG Settings:**
   ```bash
   # For faster responses
   --rag-chunk-size 500 --rag-top-k 3
   
   # For better quality
   --rag-chunk-size 1500 --rag-top-k 7
   ```

3. **Enable Caching:**
   ```toml
   [agent_memory.embedding]
   cache_embeddings = true
   max_batch_size = 100
   ```

## Troubleshooting

### Common Issues

#### Dimension Mismatch
```
❌ Configuration Error: nomic-embed-text requires 768 dimensions, but 1536 configured
```
**Solution:** The scaffold prevents this, but if you encounter it, regenerate the project or update dimensions manually.

#### Wrong Configuration Format
```
❌ Memory system not configured in agentflow.toml
```
**Solution:** Ensure you're using `[agent_memory]` not `[memory]`. Regenerate with the new scaffold.

#### Provider Not Available
```
❌ Cannot connect to Ollama service
```
**Solution:** Follow the generated troubleshooting guide in your project's README.

### Getting Help

1. **Use the debug command:**
   ```bash
   agentcli memory --validate
   ```

2. **Check the generated README** for project-specific guidance

3. **Refer to the troubleshooting guide:** `docs/guides/MemoryTroubleshooting.md`

4. **Create an issue** with your configuration and error messages

## Examples

### Complete Examples

#### Local Development Project
```bash
agentcli create dev-project --memory-enabled \
  --memory-provider pgvector \
  --embedding-provider ollama \
  --embedding-model nomic-embed-text:latest \
  --rag-enabled

cd dev-project
docker compose up -d
./setup.sh
go mod tidy
go run . -m "Hello, world!"
```

#### Production RAG System
```bash
agentcli create rag-system --memory-enabled \
  --memory-provider pgvector \
  --embedding-provider openai \
  --embedding-model text-embedding-3-small \
  --rag-enabled \
  --hybrid-search \
  --session-memory \
  --rag-chunk-size 1000 \
  --rag-top-k 5

cd rag-system
export OPENAI_API_KEY="your-key"
# Update connection string in agentflow.toml for production database
docker compose up -d
./setup.sh
agentcli memory --validate
go run . -m "Analyze this document..."
```

#### High-Performance Setup
```bash
agentcli create high-perf --memory-enabled \
  --memory-provider weaviate \
  --embedding-provider openai \
  --embedding-model text-embedding-3-large \
  --rag-enabled \
  --hybrid-search

cd high-perf
export OPENAI_API_KEY="your-key"
docker compose up -d
agentcli memory --validate
agentcli memory --stats
```

The enhanced scaffold system makes it easy to create production-ready AgentFlow projects with intelligent memory configuration, comprehensive validation, and built-in troubleshooting support.