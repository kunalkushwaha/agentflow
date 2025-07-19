# Memory Provider Setup Guide

**Complete setup guides for AgentFlow memory providers**

AgentFlow supports multiple memory providers for different use cases. This guide provides step-by-step setup instructions for each provider, from development to production deployment.

## 📚 Table of Contents

- [Overview](#overview)
- [In-Memory Provider](#in-memory-provider)
- [PostgreSQL + pgvector Setup](#postgresql--pgvector-setup)
- [Weaviate Setup](#weaviate-setup)
- [Provider Comparison](#provider-comparison)
- [Migration Guide](#migration-guide)
- [Troubleshooting](#troubleshooting)

## 🎯 Overview

### Provider Selection Guide

| Provider | Best For | Persistence | Scalability | Setup Complexity |
|----------|----------|-------------|-------------|------------------|
| **memory** | Development, Testing | ❌ No | ⚠️ Single instance | ✅ Minimal |
| **pgvector** | Production, Enterprise | ✅ Yes | ✅ High | ⚠️ Moderate |
| **weaviate** | Large-scale Vector Ops | ✅ Yes | ✅ Very High | ⚠️ Moderate |

### Quick Start Recommendations

- **Just getting started?** → Use `memory` provider
- **Building a production app?** → Use `pgvector` provider  
- **Need advanced vector features?** → Use `weaviate` provider

## 💾 In-Memory Provider

**Perfect for development, testing, and temporary sessions**

### Configuration

```toml
[memory]
enabled = true
provider = "memory"
max_results = 10
dimensions = 1536
auto_embed = true

[memory.embedding]
provider = "dummy"  # For testing
model = "text-embedding-3-small"
```

### Usage Example

```go
package main

import (
    "context"
    "log"
    
    agentflow "github.com/kunalkushwaha/agentflow/core"
)

func main() {
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
    
    // Create memory provider
    memory, err := agentflow.NewMemory(config)
    if err != nil {
        log.Fatal(err)
    }
    defer memory.Close()
    
    // Use memory system
    ctx := memory.SetSession(context.Background(), "test-session")
    
    // Store and query
    err = memory.Store(ctx, "I love programming in Go", "programming")
    if err != nil {
        log.Fatal(err)
    }
    
    results, err := memory.Query(ctx, "programming languages", 5)
    if err != nil {
        log.Fatal(err)
    }
    
    log.Printf("Found %d results", len(results))
}
```

### Pros & Cons

**✅ Advantages:**
- Zero setup required
- Fast performance
- Perfect for development
- No external dependencies

**❌ Limitations:**
- No persistence (data lost on restart)
- Single instance only
- Limited to available RAM
- Not suitable for production

---

## 🐘 PostgreSQL + pgvector Setup

**Production-ready persistent memory with excellent performance**

### Prerequisites

- PostgreSQL 12+ 
- pgvector extension
- Docker (recommended) or native PostgreSQL installation

### Option 1: Docker Setup (Recommended)

#### Step 1: Run PostgreSQL with pgvector

```bash
# Create and start PostgreSQL container with pgvector
docker run -d \
  --name agentflow-postgres \
  -e POSTGRES_DB=agentflow \
  -e POSTGRES_USER=agentflow \
  -e POSTGRES_PASSWORD=password \
  -p 5432:5432 \
  -v postgres_data:/var/lib/postgresql/data \
  pgvector/pgvector:pg16

# Wait for container to start
sleep 10

# Verify container is running
docker ps | grep agentflow-postgres
```

#### Step 2: Initialize Database

```bash
# Connect to database
docker exec -it agentflow-postgres psql -U agentflow -d agentflow

# Enable pgvector extension
CREATE EXTENSION IF NOT EXISTS vector;

# Verify extension
SELECT * FROM pg_extension WHERE extname = 'vector';

# Exit psql
\q
```

#### Step 3: Test Connection

```bash
# Test connection string
psql "postgres://agentflow:password@localhost:5432/agentflow" -c "SELECT version();"
```

### Option 2: Native Installation

#### Ubuntu/Debian

```bash
# Install PostgreSQL
sudo apt update
sudo apt install postgresql postgresql-contrib

# Install pgvector
sudo apt install postgresql-15-pgvector

# Start PostgreSQL
sudo systemctl start postgresql
sudo systemctl enable postgresql

# Create database and user
sudo -u postgres createuser agentflow
sudo -u postgres createdb agentflow -O agentflow
sudo -u postgres psql -c "ALTER USER agentflow PASSWORD 'password';"

# Enable pgvector
sudo -u postgres psql -d agentflow -c "CREATE EXTENSION IF NOT EXISTS vector;"
```

#### macOS (Homebrew)

```bash
# Install PostgreSQL
brew install postgresql

# Install pgvector
brew install pgvector

# Start PostgreSQL
brew services start postgresql

# Create database
createdb agentflow
psql agentflow -c "CREATE EXTENSION IF NOT EXISTS vector;"
```

### Configuration

#### agentflow.toml

```toml
[memory]
enabled = true
provider = "pgvector"
max_results = 10
dimensions = 1536
auto_embed = true

[memory.pgvector]
# Update with your PostgreSQL connection details
connection = "postgres://agentflow:password@localhost:5432/agentflow?sslmode=disable"
table_name = "agent_memory"

[memory.embedding]
provider = "openai"  # or "ollama" for local
model = "text-embedding-3-small"

[memory.rag]
enabled = true
chunk_size = 1000
overlap = 100
top_k = 5
score_threshold = 0.7
```

#### Environment Variables

```bash
# Database connection
export DATABASE_URL="postgres://agentflow:password@localhost:5432/agentflow?sslmode=disable"

# OpenAI API (if using OpenAI embeddings)
export OPENAI_API_KEY="your-openai-api-key"

# Or Azure OpenAI
export AZURE_OPENAI_API_KEY="your-azure-api-key"
export AZURE_OPENAI_ENDPOINT="https://your-resource.openai.azure.com/"
```

### Usage Example

```go
package main

import (
    "context"
    "log"
    "os"
    
    agentflow "github.com/kunalkushwaha/agentflow/core"
)

func main() {
    // Create pgvector configuration
    config := agentflow.AgentMemoryConfig{
        Provider:   "pgvector",
        Connection: os.Getenv("DATABASE_URL"),
        Dimensions: 1536,
        Embedding: agentflow.EmbeddingConfig{
            Provider: "openai",
            APIKey:   os.Getenv("OPENAI_API_KEY"),
            Model:    "text-embedding-3-small",
        },
    }
    
    // Create memory provider
    memory, err := agentflow.NewMemory(config)
    if err != nil {
        log.Fatal(err)
    }
    defer memory.Close()
    
    // Create session
    ctx := memory.SetSession(context.Background(), "user-123")
    
    // Store memories
    err = memory.Store(ctx, "I prefer morning meetings", "scheduling", "preference")
    if err != nil {
        log.Fatal(err)
    }
    
    // Query memories
    results, err := memory.Query(ctx, "meeting preferences", 5)
    if err != nil {
        log.Fatal(err)
    }
    
    for _, result := range results {
        log.Printf("Memory: %s (Score: %.2f)", result.Content, result.Score)
    }
    
    // RAG example - ingest document
    doc := agentflow.Document{
        ID:      "meeting-guide",
        Title:   "Meeting Best Practices",
        Content: "Effective meetings should start on time, have clear agendas...",
        Source:  "docs/meetings.md",
        Type:    agentflow.DocumentTypeText,
        Tags:    []string{"meetings", "productivity"},
    }
    
    err = memory.IngestDocument(ctx, doc)
    if err != nil {
        log.Fatal(err)
    }
    
    // Build RAG context
    ragContext, err := memory.BuildContext(ctx, "How to run effective meetings?")
    if err != nil {
        log.Fatal(err)
    }
    
    log.Printf("RAG Context (%d tokens):\n%s", ragContext.TokenCount, ragContext.ContextText)
}
```

### Production Optimization

#### Connection Pooling

```toml
[memory.pgvector]
connection = "postgres://agentflow:password@localhost:5432/agentflow?sslmode=disable&pool_max_conns=25&pool_min_conns=5"
```

#### Performance Tuning

```sql
-- Optimize PostgreSQL for vector operations
-- Add to postgresql.conf

# Memory settings
shared_buffers = 256MB
effective_cache_size = 1GB
work_mem = 64MB

# Vector-specific settings
max_parallel_workers_per_gather = 2
max_parallel_workers = 8

# Connection settings
max_connections = 100
```

#### Monitoring

```sql
-- Monitor vector operations
SELECT 
    schemaname,
    tablename,
    n_tup_ins as inserts,
    n_tup_upd as updates,
    n_tup_del as deletes
FROM pg_stat_user_tables 
WHERE tablename LIKE '%memory%';

-- Check index usage
SELECT 
    indexrelname,
    idx_scan,
    idx_tup_read,
    idx_tup_fetch
FROM pg_stat_user_indexes 
WHERE indexrelname LIKE '%memory%';
```

### Pros & Cons

**✅ Advantages:**
- Full persistence
- Excellent performance (~45ms queries)
- ACID transactions
- Mature ecosystem
- Advanced indexing
- Production-ready

**❌ Considerations:**
- Requires PostgreSQL setup
- More complex than in-memory
- Database maintenance needed

---

## 🔍 Weaviate Setup

**Dedicated vector database for large-scale operations**

### Prerequisites

- Docker or Kubernetes
- 4GB+ RAM recommended
- Network access for Weaviate

### Docker Setup

#### Step 1: Run Weaviate

```bash
# Create Weaviate container
docker run -d \
  --name agentflow-weaviate \
  -p 8080:8080 \
  -e QUERY_DEFAULTS_LIMIT=25 \
  -e AUTHENTICATION_ANONYMOUS_ACCESS_ENABLED='true' \
  -e PERSISTENCE_DATA_PATH='/var/lib/weaviate' \
  -e DEFAULT_VECTORIZER_MODULE='none' \
  -e CLUSTER_HOSTNAME='node1' \
  -e ENABLE_MODULES='text2vec-openai,text2vec-cohere,text2vec-huggingface,ref2vec-centroid,generative-openai,qna-openai' \
  -v weaviate_data:/var/lib/weaviate \
  semitechnologies/weaviate:1.22.4

# Wait for startup
sleep 15

# Verify Weaviate is running
curl http://localhost:8080/v1/.well-known/ready
```

#### Step 2: Test Connection

```bash
# Check Weaviate status
curl http://localhost:8080/v1/meta

# Expected response: JSON with version info
```

### Docker Compose Setup

Create `docker-compose.yml`:

```yaml
version: '3.8'

services:
  weaviate:
    image: semitechnologies/weaviate:1.22.4
    container_name: agentflow-weaviate
    ports:
      - "8080:8080"
    environment:
      QUERY_DEFAULTS_LIMIT: 25
      AUTHENTICATION_ANONYMOUS_ACCESS_ENABLED: 'true'
      PERSISTENCE_DATA_PATH: '/var/lib/weaviate'
      DEFAULT_VECTORIZER_MODULE: 'none'
      CLUSTER_HOSTNAME: 'node1'
      ENABLE_MODULES: 'text2vec-openai,text2vec-cohere,text2vec-huggingface,ref2vec-centroid,generative-openai,qna-openai'
    volumes:
      - weaviate_data:/var/lib/weaviate
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:8080/v1/.well-known/ready"]
      interval: 10s
      timeout: 5s
      retries: 5
    restart: unless-stopped

volumes:
  weaviate_data:
    driver: local
```

Start with Docker Compose:

```bash
# Start Weaviate
docker-compose up -d

# Check logs
docker-compose logs weaviate

# Stop when needed
docker-compose down
```

### Configuration

#### agentflow.toml

```toml
[memory]
enabled = true
provider = "weaviate"
max_results = 10
dimensions = 1536
auto_embed = true

[memory.weaviate]
# Update with your Weaviate connection details
connection = "http://localhost:8080"
class_name = "AgentMemory"

[memory.embedding]
provider = "openai"
model = "text-embedding-3-small"

[memory.rag]
enabled = true
chunk_size = 1000
overlap = 100
top_k = 5
score_threshold = 0.7
```

#### Environment Variables

```bash
# Weaviate connection
export WEAVIATE_URL="http://localhost:8080"
export WEAVIATE_CLASS_NAME="AgentMemory"

# OpenAI API (for embeddings)
export OPENAI_API_KEY="your-openai-api-key"
```

### Usage Example

```go
package main

import (
    "context"
    "log"
    "os"
    
    agentflow "github.com/kunalkushwaha/agentflow/core"
)

func main() {
    // Create Weaviate configuration
    config := agentflow.AgentMemoryConfig{
        Provider:   "weaviate",
        Connection: "http://localhost:8080",
        Dimensions: 1536,
        Embedding: agentflow.EmbeddingConfig{
            Provider: "openai",
            APIKey:   os.Getenv("OPENAI_API_KEY"),
            Model:    "text-embedding-3-small",
        },
    }
    
    // Create memory provider
    memory, err := agentflow.NewMemory(config)
    if err != nil {
        log.Fatal(err)
    }
    defer memory.Close()
    
    // Create session
    ctx := memory.SetSession(context.Background(), "user-456")
    
    // Store and query memories
    err = memory.Store(ctx, "I work remotely from San Francisco", "location", "work")
    if err != nil {
        log.Fatal(err)
    }
    
    results, err := memory.Query(ctx, "work location", 3)
    if err != nil {
        log.Fatal(err)
    }
    
    for _, result := range results {
        log.Printf("Memory: %s (Score: %.2f)", result.Content, result.Score)
    }
}
```

### Advanced Configuration

#### Authentication (Production)

```yaml
# docker-compose.yml with authentication
services:
  weaviate:
    environment:
      AUTHENTICATION_ANONYMOUS_ACCESS_ENABLED: 'false'
      AUTHENTICATION_APIKEY_ENABLED: 'true'
      AUTHENTICATION_APIKEY_ALLOWED_KEYS: 'your-secret-key'
      AUTHENTICATION_APIKEY_USERS: 'admin'
```

```toml
# agentflow.toml with authentication
[memory.weaviate]
connection = "http://localhost:8080"
api_key = "your-secret-key"
class_name = "AgentMemory"
```

#### Clustering (Production)

```yaml
# Multi-node Weaviate cluster
services:
  weaviate-node1:
    image: semitechnologies/weaviate:1.22.4
    environment:
      CLUSTER_HOSTNAME: 'node1'
      CLUSTER_GOSSIP_BIND_PORT: '7100'
      CLUSTER_DATA_BIND_PORT: '7101'
      
  weaviate-node2:
    image: semitechnologies/weaviate:1.22.4
    environment:
      CLUSTER_HOSTNAME: 'node2'
      CLUSTER_GOSSIP_BIND_PORT: '7102'
      CLUSTER_DATA_BIND_PORT: '7103'
      CLUSTER_JOIN: 'node1:7100'
```

### Monitoring

#### Health Checks

```bash
# Check Weaviate health
curl http://localhost:8080/v1/.well-known/ready

# Check cluster status
curl http://localhost:8080/v1/nodes

# Check schema
curl http://localhost:8080/v1/schema
```

#### Performance Monitoring

```bash
# Check metrics
curl http://localhost:8080/v1/meta

# Monitor resource usage
docker stats agentflow-weaviate
```

### Pros & Cons

**✅ Advantages:**
- Purpose-built for vectors
- Excellent scalability
- Advanced search features
- Built-in clustering
- GraphQL API
- Rich ecosystem

**❌ Considerations:**
- More complex setup
- Higher resource usage
- Learning curve
- Newer ecosystem

---

## 📊 Provider Comparison

### Performance Comparison

| Operation | Memory | PgVector | Weaviate |
|-----------|--------|----------|----------|
| **Store** | ~1ms | ~50ms | ~75ms |
| **Query** | ~5ms | ~45ms | ~60ms |
| **Batch Store** | ~10ms | ~2.3s | ~3.1s |
| **RAG Context** | ~15ms | ~90ms | ~120ms |

### Feature Comparison

| Feature | Memory | PgVector | Weaviate |
|---------|--------|----------|----------|
| **Persistence** | ❌ | ✅ | ✅ |
| **Scalability** | ❌ | ✅ | ✅ |
| **ACID Transactions** | ❌ | ✅ | ❌ |
| **Vector Indexing** | ❌ | ✅ | ✅ |
| **Clustering** | ❌ | ⚠️ | ✅ |
| **GraphQL API** | ❌ | ❌ | ✅ |
| **Setup Complexity** | ✅ Easy | ⚠️ Moderate | ⚠️ Moderate |

### Use Case Recommendations

#### Choose **Memory** when:
- Developing and testing
- Prototyping quickly
- Temporary sessions only
- Minimal setup required

#### Choose **PgVector** when:
- Building production applications
- Need ACID transactions
- Have PostgreSQL expertise
- Want mature ecosystem

#### Choose **Weaviate** when:
- Large-scale vector operations
- Need advanced search features
- Building vector-first applications
- Want purpose-built vector DB

---

## 🔄 Migration Guide

### From Memory to PgVector

1. **Setup PostgreSQL with pgvector**
2. **Update configuration:**
   ```toml
   [memory]
   provider = "pgvector"  # Changed from "memory"
   
   [memory.pgvector]
   connection = "postgres://user:password@localhost:5432/agentflow"
   ```
3. **Add environment variables**
4. **Test connection**
5. **Migrate data** (if needed)

### From PgVector to Weaviate

1. **Setup Weaviate**
2. **Update configuration:**
   ```toml
   [memory]
   provider = "weaviate"  # Changed from "pgvector"
   
   [memory.weaviate]
   connection = "http://localhost:8080"
   ```
3. **Export data from PostgreSQL**
4. **Import data to Weaviate**
5. **Test functionality**

### Data Migration Script

```go
func migrateMemoryProvider(oldMemory, newMemory agentflow.Memory) error {
    // Get all sessions (implementation depends on provider)
    sessions := []string{"session1", "session2"} // Get from old provider
    
    for _, sessionID := range sessions {
        ctx := oldMemory.SetSession(context.Background(), sessionID)
        newCtx := newMemory.SetSession(context.Background(), sessionID)
        
        // Migrate personal memories
        results, err := oldMemory.Query(ctx, "", 1000) // Get all
        if err != nil {
            return err
        }
        
        for _, result := range results {
            err = newMemory.Store(newCtx, result.Content, result.Tags...)
            if err != nil {
                return err
            }
        }
        
        // Migrate chat history
        history, err := oldMemory.GetHistory(ctx, 1000)
        if err != nil {
            return err
        }
        
        for _, msg := range history {
            err = newMemory.AddMessage(newCtx, msg.Role, msg.Content)
            if err != nil {
                return err
            }
        }
    }
    
    return nil
}
```

---

## 🔧 Troubleshooting

### Common Issues

#### Connection Problems

**PostgreSQL Connection Failed:**
```bash
# Check if PostgreSQL is running
docker ps | grep postgres
# or
sudo systemctl status postgresql

# Test connection
psql "postgres://user:password@localhost:5432/dbname" -c "SELECT 1;"

# Check firewall
sudo ufw status
```

**Weaviate Connection Failed:**
```bash
# Check if Weaviate is running
docker ps | grep weaviate
curl http://localhost:8080/v1/.well-known/ready

# Check logs
docker logs agentflow-weaviate
```

#### Performance Issues

**Slow Queries:**
```sql
-- PostgreSQL: Check query performance
EXPLAIN ANALYZE SELECT * FROM agent_memory 
WHERE embedding <-> '[0.1,0.2,...]' < 0.5;

-- Add indexes if needed
CREATE INDEX CONCURRENTLY idx_memory_embedding 
ON agent_memory USING ivfflat (embedding vector_cosine_ops);
```

**High Memory Usage:**
```bash
# Monitor resource usage
docker stats

# Adjust memory limits
docker run --memory=2g --name agentflow-postgres ...
```

#### Data Issues

**Missing Extensions:**
```sql
-- PostgreSQL: Install pgvector
CREATE EXTENSION IF NOT EXISTS vector;

-- Check extensions
SELECT * FROM pg_extension WHERE extname = 'vector';
```

**Schema Issues:**
```bash
# Weaviate: Check schema
curl http://localhost:8080/v1/schema

# Reset schema (careful!)
curl -X DELETE http://localhost:8080/v1/schema
```

### Debug Mode

Enable debug logging:

```toml
[logging]
level = "debug"
```

```go
// Enable debug logging in code
import "log"
log.SetFlags(log.LstdFlags | log.Lshortfile)
```

### Health Check Scripts

#### PostgreSQL Health Check

```bash
#!/bin/bash
# check-postgres.sh

DB_URL="postgres://agentflow:password@localhost:5432/agentflow"

echo "Checking PostgreSQL connection..."
if psql "$DB_URL" -c "SELECT 1;" > /dev/null 2>&1; then
    echo "✅ PostgreSQL connection successful"
else
    echo "❌ PostgreSQL connection failed"
    exit 1
fi

echo "Checking pgvector extension..."
if psql "$DB_URL" -c "SELECT * FROM pg_extension WHERE extname = 'vector';" | grep -q vector; then
    echo "✅ pgvector extension installed"
else
    echo "❌ pgvector extension missing"
    exit 1
fi

echo "✅ All checks passed"
```

#### Weaviate Health Check

```bash
#!/bin/bash
# check-weaviate.sh

WEAVIATE_URL="http://localhost:8080"

echo "Checking Weaviate connection..."
if curl -f "$WEAVIATE_URL/v1/.well-known/ready" > /dev/null 2>&1; then
    echo "✅ Weaviate connection successful"
else
    echo "❌ Weaviate connection failed"
    exit 1
fi

echo "Checking Weaviate schema..."
if curl -f "$WEAVIATE_URL/v1/schema" > /dev/null 2>&1; then
    echo "✅ Weaviate schema accessible"
else
    echo "❌ Weaviate schema not accessible"
    exit 1
fi

echo "✅ All checks passed"
```

---

## 🎯 Summary

This guide covered complete setup for all AgentFlow memory providers:

✅ **In-Memory**: Perfect for development and testing  
✅ **PostgreSQL + pgvector**: Production-ready with excellent performance  
✅ **Weaviate**: Advanced vector database for large-scale operations  

### Next Steps

1. **Choose your provider** based on your use case
2. **Follow the setup guide** for your chosen provider
3. **Configure AgentFlow** with the appropriate settings
4. **Test your setup** with the provided examples
5. **Monitor and optimize** for production use

For more information:
- **[Memory System Guide](Memory.md)** - Complete API reference
- **[Configuration Guide](Configuration.md)** - Advanced configuration options
- **[RAG Configuration Guide](../RAG_CONFIGURATION_GUIDE.md)** - RAG-specific settings
