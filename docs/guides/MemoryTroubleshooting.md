# Memory System Troubleshooting Guide

This guide helps you troubleshoot common issues with the AgentFlow memory system, especially when using the scaffold-generated projects.

## Quick Diagnosis

Use the memory debug command to quickly diagnose issues:

```bash
# Basic overview and connection test
agentcli memory

# Detailed configuration validation
agentcli memory --validate

# Show current configuration
agentcli memory --config
```

## Common Issues and Solutions

### 1. Dimension Mismatch Errors

**Error Message:**
```
❌ Configuration Error: nomic-embed-text requires 768 dimensions, but 1536 configured
💡 Solution: Update agentflow.toml [agent_memory] dimensions = 768
```

**Cause:** The embedding model dimensions don't match the configured dimensions in your `agentflow.toml`.

**Solution:**
1. Check your embedding model's actual dimensions:
   ```bash
   agentcli memory --config
   ```

2. Update your `agentflow.toml`:
   ```toml
   [agent_memory]
   dimensions = 768  # Match your embedding model
   ```

3. If using pgvector, update your database schema:
   ```sql
   -- Connect to your database
   psql -h localhost -U user -d agentflow
   
   -- Drop and recreate tables with correct dimensions
   DROP TABLE IF EXISTS agent_memory;
   CREATE TABLE agent_memory (
       id SERIAL PRIMARY KEY,
       content TEXT NOT NULL,
       embedding vector(768),  -- Use correct dimensions
       tags TEXT[],
       metadata JSONB,
       created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
   );
   ```

### 2. Database Connection Issues

#### PostgreSQL/PgVector Issues

**Error Message:**
```
❌ Failed to connect to memory system: connection refused
```

**Solutions:**

1. **Start the database:**
   ```bash
   docker compose up -d
   ```

2. **Check if PostgreSQL is running:**
   ```bash
   docker ps | grep postgres
   ```

3. **Verify connection string in `agentflow.toml`:**
   ```toml
   [agent_memory]
   connection = "postgres://user:password@localhost:5432/agentflow?sslmode=disable"
   ```

4. **Test connection manually:**
   ```bash
   psql -h localhost -U user -d agentflow
   ```

5. **Check database exists:**
   ```sql
   \l  -- List databases
   \c agentflow  -- Connect to agentflow database
   \dt  -- List tables
   ```

6. **Recreate database if needed:**
   ```bash
   ./setup.sh  # or setup.bat on Windows
   ```

#### Weaviate Issues

**Error Message:**
```
❌ Cannot connect to Weaviate database
```

**Solutions:**

1. **Start Weaviate:**
   ```bash
   docker compose up -d
   ```

2. **Check Weaviate status:**
   ```bash
   curl http://localhost:8080/v1/meta
   ```

3. **Verify connection in `agentflow.toml`:**
   ```toml
   [agent_memory]
   connection = "http://localhost:8080"
   ```

### 3. Embedding Provider Issues

#### Ollama Issues

**Error Message:**
```
❌ Connection Error: Cannot connect to Ollama service
```

**Solutions:**

1. **Start Ollama:**
   ```bash
   ollama serve
   ```

2. **Pull the embedding model:**
   ```bash
   ollama pull nomic-embed-text:latest
   ```

3. **Test Ollama connection:**
   ```bash
   curl http://localhost:11434/api/tags
   ```

4. **Verify configuration:**
   ```toml
   [agent_memory.embedding]
   provider = "ollama"
   model = "nomic-embed-text:latest"
   base_url = "http://localhost:11434"
   ```

#### OpenAI Issues

**Error Message:**
```
❌ OpenAI API Error: authentication failed
```

**Solutions:**

1. **Set API key:**
   ```bash
   export OPENAI_API_KEY="your-api-key-here"
   ```

2. **Verify API key is valid:**
   ```bash
   curl -H "Authorization: Bearer $OPENAI_API_KEY" https://api.openai.com/v1/models
   ```

3. **Check model name:**
   ```toml
   [agent_memory.embedding]
   provider = "openai"
   model = "text-embedding-3-small"  # or text-embedding-3-large
   ```

### 4. Configuration Structure Issues

**Error Message:**
```
❌ Memory system not configured in agentflow.toml
```

**Cause:** Using old `[memory]` section instead of `[agent_memory]`.

**Solution:** Update your configuration structure:

```toml
# ❌ Old format (doesn't work)
[memory]
provider = "pgvector"

# ✅ New format (correct)
[agent_memory]
provider = "pgvector"
connection = "postgres://user:password@localhost:5432/agentflow?sslmode=disable"
dimensions = 768
auto_embed = true

[agent_memory.embedding]
provider = "ollama"
model = "nomic-embed-text:latest"
base_url = "http://localhost:11434"
```

### 5. RAG Configuration Issues

**Error Message:**
```
❌ RAG chunk overlap must be between 0 and chunk_size
```

**Solution:** Fix RAG parameters:

```toml
[agent_memory]
enable_rag = true
chunk_size = 1000
chunk_overlap = 100  # Must be less than chunk_size
knowledge_score_threshold = 0.7  # Between 0.0 and 1.0
```

**Recommended RAG Settings:**
- **Chunk Size:** 500-2000 tokens
- **Chunk Overlap:** 10-20% of chunk size
- **Score Threshold:** 0.6-0.8
- **Top-K Results:** 3-10

### 6. Performance Issues

#### Slow Queries

**Symptoms:** Memory queries taking too long.

**Solutions:**

1. **Reduce dimensions if possible:**
   ```toml
   [agent_memory]
   dimensions = 768  # Instead of 1536 or 3072
   ```

2. **Optimize database (PostgreSQL):**
   ```sql
   -- Create index on embedding column
   CREATE INDEX ON agent_memory USING ivfflat (embedding vector_cosine_ops);
   
   -- Analyze table for better query planning
   ANALYZE agent_memory;
   ```

3. **Reduce result count:**
   ```toml
   [agent_memory]
   max_results = 5  # Instead of 10 or more
   ```

#### High Memory Usage

**Solutions:**

1. **Enable embedding caching:**
   ```toml
   [agent_memory.embedding]
   cache_embeddings = true
   ```

2. **Reduce batch size:**
   ```toml
   [agent_memory.embedding]
   max_batch_size = 50  # Instead of 100
   ```

3. **Use smaller embedding model:**
   ```toml
   [agent_memory.embedding]
   model = "all-minilm"  # 384 dimensions instead of 768
   ```

## Diagnostic Commands

### Memory Debug Commands

```bash
# Show overview and basic stats
agentcli memory

# Detailed statistics
agentcli memory --stats

# List recent memories
agentcli memory --list

# Test search functionality
agentcli memory --search "your query"

# Validate configuration
agentcli memory --validate

# Show current configuration
agentcli memory --config

# Show active sessions
agentcli memory --sessions

# List knowledge base documents
agentcli memory --docs
```

### Database Diagnostic Commands

#### PostgreSQL
```bash
# Connect to database
psql -h localhost -U user -d agentflow

# Check table structure
\d agent_memory

# Count records
SELECT COUNT(*) FROM agent_memory;

# Check embedding dimensions
SELECT array_length(embedding, 1) FROM agent_memory LIMIT 1;

# Check recent entries
SELECT id, content, created_at FROM agent_memory ORDER BY created_at DESC LIMIT 5;
```

#### Weaviate
```bash
# Check Weaviate status
curl http://localhost:8080/v1/meta

# List classes
curl http://localhost:8080/v1/schema

# Check objects count
curl http://localhost:8080/v1/objects
```

## Migration Guide

### Upgrading from Old Configuration Format

If you have an existing project with the old `[memory]` configuration:

1. **Backup your data** (if using persistent storage)

2. **Update configuration structure:**
   ```bash
   # Create backup
   cp agentflow.toml agentflow.toml.backup
   
   # Update to new format (manual edit required)
   # Change [memory] to [agent_memory]
   # Add [agent_memory.embedding] section
   ```

3. **Regenerate project files:**
   ```bash
   # Generate new project with correct structure
   agentcli create myproject-new --memory-enabled --memory-provider pgvector \
     --embedding-provider ollama --embedding-model nomic-embed-text:latest
   
   # Copy your custom agent logic to new project
   ```

4. **Migrate data** (if needed):
   ```sql
   -- Export data from old format
   pg_dump -h localhost -U user -d agentflow -t agent_memory > backup.sql
   
   -- Import to new database
   psql -h localhost -U user -d agentflow < backup.sql
   ```

### Changing Embedding Models

When changing embedding models with different dimensions:

1. **Update configuration:**
   ```toml
   [agent_memory]
   dimensions = 1536  # New model dimensions
   
   [agent_memory.embedding]
   model = "text-embedding-3-small"  # New model
   ```

2. **Migrate database schema:**
   ```sql
   -- For PostgreSQL
   ALTER TABLE agent_memory ALTER COLUMN embedding TYPE vector(1536);
   
   -- You may need to recreate the table if this fails
   ```

3. **Re-embed existing content:**
   ```bash
   # This would require custom script or clearing and re-adding content
   agentcli memory --clear  # Use with caution!
   ```

## Getting Help

If you're still experiencing issues:

1. **Check the logs** for detailed error messages
2. **Use the validation command:** `agentcli memory --validate`
3. **Check the GitHub issues:** [AgentFlow Issues](https://github.com/kunalkushwaha/agentflow/issues)
4. **Create a new issue** with:
   - Your `agentflow.toml` configuration
   - Error messages
   - Output of `agentcli memory --validate`
   - Your operating system and Docker version

## Best Practices

### Configuration
- Always use the new `[agent_memory]` configuration format
- Match embedding dimensions with your chosen model
- Use recommended RAG parameters
- Enable embedding caching for better performance

### Development
- Start with in-memory provider for development
- Use pgvector for production
- Test configuration with `agentcli memory --validate`
- Monitor performance with `agentcli memory --stats`

### Production
- Use persistent storage (pgvector or weaviate)
- Set up proper database backups
- Monitor memory usage and query performance
- Use appropriate embedding models for your use case