# PostgreSQL + pgvector Setup Guide

**Complete guide for setting up PostgreSQL with pgvector for AgentFlow memory**

This guide provides detailed instructions for setting up PostgreSQL with the pgvector extension for production AgentFlow deployments.

## 🎯 Overview

PostgreSQL with pgvector provides:
- **Persistent storage** for agent memories and knowledge base
- **Vector similarity search** with excellent performance (~45ms queries)
- **ACID transactions** for data consistency
- **Production-ready** scalability and reliability
- **Advanced indexing** for optimal query performance

## 📋 Prerequisites

- Docker (recommended) OR PostgreSQL 12+
- 2GB+ RAM available
- Network access to PostgreSQL port (5432)
- Basic command line knowledge

## 🚀 Quick Start (Docker)

### Step 1: Run PostgreSQL with pgvector

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
echo "Waiting for PostgreSQL to start..."
sleep 10

# Verify container is running
docker ps | grep agentflow-postgres
```

### Step 2: Initialize Database

```bash
# Connect to database and enable pgvector
docker exec -it agentflow-postgres psql -U agentflow -d agentflow -c "
CREATE EXTENSION IF NOT EXISTS vector;
SELECT extname, extversion FROM pg_extension WHERE extname = 'vector';
"

# Expected output: vector | 0.5.1 (or similar)
```

### Step 3: Test Connection

```bash
# Test connection string
docker exec -it agentflow-postgres psql -U agentflow -d agentflow -c "
SELECT version();
SELECT * FROM pg_extension WHERE extname = 'vector';
"
```

### Step 4: Configure AgentFlow

Create `agentflow.toml`:

```toml
[memory]
enabled = true
provider = "pgvector"
max_results = 10
dimensions = 1536
auto_embed = true

[memory.pgvector]
connection = "postgres://agentflow:password@localhost:5432/agentflow?sslmode=disable"
table_name = "agent_memory"

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

Set environment variables:

```bash
export DATABASE_URL="postgres://agentflow:password@localhost:5432/agentflow?sslmode=disable"
export OPENAI_API_KEY="your-openai-api-key"
```

**✅ You're ready to use AgentFlow with pgvector!**

---

## 🔧 Detailed Setup Options

### Option 1: Docker Compose (Recommended for Development)

Create `docker-compose.yml`:

```yaml
version: '3.8'

services:
  postgres:
    image: pgvector/pgvector:pg16
    container_name: agentflow-postgres
    environment:
      POSTGRES_DB: agentflow
      POSTGRES_USER: agentflow
      POSTGRES_PASSWORD: password
    ports:
      - "5432:5432"
    volumes:
      - postgres_data:/var/lib/postgresql/data
      - ./init-db.sql:/docker-entrypoint-initdb.d/init-db.sql
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U agentflow -d agentflow"]
      interval: 10s
      timeout: 5s
      retries: 5
    restart: unless-stopped

volumes:
  postgres_data:
    driver: local
```

Create `init-db.sql`:

```sql
-- AgentFlow Database Initialization Script for pgvector
-- This script sets up the database schema for AgentFlow memory system

-- Enable the pgvector extension
CREATE EXTENSION IF NOT EXISTS vector;

-- Create the agent_memory table for storing embeddings and content
CREATE TABLE IF NOT EXISTS agent_memory (
    id SERIAL PRIMARY KEY,
    content TEXT NOT NULL,
    embedding vector(1536),
    tags TEXT[],
    metadata JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create indexes for better performance
CREATE INDEX IF NOT EXISTS idx_agent_memory_embedding ON agent_memory USING ivfflat (embedding vector_cosine_ops);
CREATE INDEX IF NOT EXISTS idx_agent_memory_tags ON agent_memory USING GIN (tags);
CREATE INDEX IF NOT EXISTS idx_agent_memory_created_at ON agent_memory (created_at);
CREATE INDEX IF NOT EXISTS idx_agent_memory_metadata ON agent_memory USING GIN (metadata);

-- Create the chat_history table for storing conversation history
CREATE TABLE IF NOT EXISTS chat_history (
    id SERIAL PRIMARY KEY,
    session_id VARCHAR(255) NOT NULL,
    role VARCHAR(50) NOT NULL CHECK (role IN ('user', 'assistant', 'system')),
    content TEXT NOT NULL,
    metadata JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create indexes for chat history
CREATE INDEX IF NOT EXISTS idx_chat_history_session_id ON chat_history (session_id);
CREATE INDEX IF NOT EXISTS idx_chat_history_created_at ON chat_history (created_at);

-- Create the documents table for RAG document storage
CREATE TABLE IF NOT EXISTS documents (
    id SERIAL PRIMARY KEY,
    title VARCHAR(500),
    content TEXT NOT NULL,
    source VARCHAR(1000),
    document_type VARCHAR(50),
    chunk_index INTEGER DEFAULT 0,
    chunk_total INTEGER DEFAULT 1,
    embedding vector(1536),
    metadata JSONB,
    tags TEXT[],
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create indexes for documents
CREATE INDEX IF NOT EXISTS idx_documents_embedding ON documents USING ivfflat (embedding vector_cosine_ops);
CREATE INDEX IF NOT EXISTS idx_documents_source ON documents (source);
CREATE INDEX IF NOT EXISTS idx_documents_type ON documents (document_type);
CREATE INDEX IF NOT EXISTS idx_documents_tags ON documents USING GIN (tags);
CREATE INDEX IF NOT EXISTS idx_documents_created_at ON documents (created_at);

-- Grant permissions to the user
GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO agentflow;
GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA public TO agentflow;

-- Insert a test record to verify the setup
INSERT INTO agent_memory (content, tags, metadata) 
VALUES ('AgentFlow memory system initialized successfully', ARRAY['system', 'initialization'], '{"source": "init_script", "version": "1.0"}')
ON CONFLICT DO NOTHING;
```

Start the services:

```bash
# Start PostgreSQL
docker-compose up -d

# Check logs
docker-compose logs postgres

# Test connection
docker-compose exec postgres psql -U agentflow -d agentflow -c "SELECT version();"
```

### Option 2: Native Installation

#### Ubuntu/Debian

```bash
# Update package list
sudo apt update

# Install PostgreSQL
sudo apt install postgresql postgresql-contrib

# Install pgvector
sudo apt install postgresql-15-pgvector

# Start and enable PostgreSQL
sudo systemctl start postgresql
sudo systemctl enable postgresql

# Create database and user
sudo -u postgres createuser agentflow
sudo -u postgres createdb agentflow -O agentflow
sudo -u postgres psql -c "ALTER USER agentflow PASSWORD 'password';"

# Enable pgvector extension
sudo -u postgres psql -d agentflow -c "CREATE EXTENSION IF NOT EXISTS vector;"

# Test connection
psql "postgres://agentflow:password@localhost:5432/agentflow" -c "SELECT version();"
```

#### CentOS/RHEL/Fedora

```bash
# Install PostgreSQL
sudo dnf install postgresql postgresql-server postgresql-contrib

# Install pgvector (may need to compile from source)
git clone https://github.com/pgvector/pgvector.git
cd pgvector
make
sudo make install

# Initialize database
sudo postgresql-setup --initdb

# Start and enable PostgreSQL
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

# Start PostgreSQL service
brew services start postgresql

# Create database
createdb agentflow

# Enable pgvector
psql agentflow -c "CREATE EXTENSION IF NOT EXISTS vector;"

# Create user (optional)
psql agentflow -c "CREATE USER agentflow WITH PASSWORD 'password';"
psql agentflow -c "GRANT ALL PRIVILEGES ON DATABASE agentflow TO agentflow;"
```

---

## ⚙️ Configuration

### Connection String Options

```bash
# Basic connection
postgres://username:password@host:port/database

# With SSL (production)
postgres://username:password@host:port/database?sslmode=require

# With connection pooling
postgres://username:password@host:port/database?pool_max_conns=25&pool_min_conns=5

# Complete example
postgres://agentflow:password@localhost:5432/agentflow?sslmode=disable&pool_max_conns=25&pool_min_conns=5&pool_max_conn_lifetime=1h
```

### AgentFlow Configuration

#### Basic Configuration

```toml
[memory]
enabled = true
provider = "pgvector"
max_results = 10
dimensions = 1536
auto_embed = true

[memory.pgvector]
connection = "postgres://agentflow:password@localhost:5432/agentflow?sslmode=disable"
table_name = "agent_memory"

[memory.embedding]
provider = "openai"
model = "text-embedding-3-small"
```

#### Production Configuration

```toml
[memory]
enabled = true
provider = "pgvector"
max_results = 10
dimensions = 1536
auto_embed = true

[memory.pgvector]
connection = "postgres://agentflow:password@localhost:5432/agentflow?sslmode=require&pool_max_conns=25&pool_min_conns=5"
table_name = "agent_memory"

[memory.embedding]
provider = "openai"
model = "text-embedding-3-small"
cache_embeddings = true
max_batch_size = 50
timeout_seconds = 30

[memory.rag]
enabled = true
chunk_size = 1000
overlap = 100
top_k = 5
score_threshold = 0.7
hybrid_search = true
session_memory = true

[memory.advanced]
retry_max_attempts = 3
retry_base_delay = "100ms"
retry_max_delay = "5s"
connection_pool_size = 25
health_check_interval = "1m"
```

### Environment Variables

```bash
# Database connection
export DATABASE_URL="postgres://agentflow:password@localhost:5432/agentflow?sslmode=disable"

# OpenAI API (if using OpenAI embeddings)
export OPENAI_API_KEY="your-openai-api-key"

# Azure OpenAI (alternative)
export AZURE_OPENAI_API_KEY="your-azure-api-key"
export AZURE_OPENAI_ENDPOINT="https://your-resource.openai.azure.com/"

# Ollama (for local embeddings)
export OLLAMA_BASE_URL="http://localhost:11434"
export OLLAMA_MODEL="mxbai-embed-large"
```

---

## 🚀 Performance Optimization

### PostgreSQL Configuration

Add to `postgresql.conf`:

```ini
# Memory settings
shared_buffers = 256MB                    # 25% of RAM
effective_cache_size = 1GB                # 75% of RAM
work_mem = 64MB                          # For sorting and hashing
maintenance_work_mem = 256MB             # For maintenance operations

# Vector-specific settings
max_parallel_workers_per_gather = 2      # Parallel query execution
max_parallel_workers = 8                 # Total parallel workers

# Connection settings
max_connections = 100                    # Adjust based on your needs
shared_preload_libraries = 'pg_stat_statements'  # Query statistics

# Logging (for debugging)
log_statement = 'all'                    # Log all statements (development only)
log_duration = on                        # Log query duration
log_min_duration_statement = 1000        # Log slow queries (>1s)
```

### Index Optimization

```sql
-- Create optimal indexes for vector operations
CREATE INDEX CONCURRENTLY idx_agent_memory_embedding_cosine 
ON agent_memory USING ivfflat (embedding vector_cosine_ops) 
WITH (lists = 100);

CREATE INDEX CONCURRENTLY idx_agent_memory_embedding_l2 
ON agent_memory USING ivfflat (embedding vector_l2_ops) 
WITH (lists = 100);

-- Indexes for filtering
CREATE INDEX CONCURRENTLY idx_agent_memory_session_tags 
ON agent_memory (session_id, tags);

CREATE INDEX CONCURRENTLY idx_agent_memory_created_at_desc 
ON agent_memory (created_at DESC);

-- Composite indexes for common queries
CREATE INDEX CONCURRENTLY idx_agent_memory_session_created 
ON agent_memory (session_id, created_at DESC);
```

### Query Optimization

```sql
-- Analyze query performance
EXPLAIN ANALYZE 
SELECT content, embedding <-> '[0.1,0.2,...]'::vector AS distance
FROM agent_memory 
WHERE session_id = 'user-123'
ORDER BY embedding <-> '[0.1,0.2,...]'::vector 
LIMIT 10;

-- Update table statistics
ANALYZE agent_memory;

-- Vacuum regularly
VACUUM ANALYZE agent_memory;
```

### Connection Pooling

Use connection pooling for production:

```bash
# Install pgbouncer
sudo apt install pgbouncer

# Configure pgbouncer
# /etc/pgbouncer/pgbouncer.ini
[databases]
agentflow = host=localhost port=5432 dbname=agentflow

[pgbouncer]
listen_port = 6432
listen_addr = localhost
auth_type = md5
auth_file = /etc/pgbouncer/userlist.txt
pool_mode = transaction
max_client_conn = 100
default_pool_size = 25
```

---

## 📊 Monitoring and Maintenance

### Health Checks

```bash
#!/bin/bash
# health-check.sh

DB_URL="postgres://agentflow:password@localhost:5432/agentflow"

echo "🔍 Checking PostgreSQL health..."

# Test connection
if psql "$DB_URL" -c "SELECT 1;" > /dev/null 2>&1; then
    echo "✅ Database connection successful"
else
    echo "❌ Database connection failed"
    exit 1
fi

# Check pgvector extension
if psql "$DB_URL" -c "SELECT * FROM pg_extension WHERE extname = 'vector';" | grep -q vector; then
    echo "✅ pgvector extension installed"
else
    echo "❌ pgvector extension missing"
    exit 1
fi

# Check table existence
TABLES=("agent_memory" "chat_history" "documents")
for table in "${TABLES[@]}"; do
    if psql "$DB_URL" -c "\dt $table" | grep -q "$table"; then
        echo "✅ Table $table exists"
    else
        echo "⚠️  Table $table missing (will be created automatically)"
    fi
done

# Check disk space
DISK_USAGE=$(df -h /var/lib/postgresql/data | awk 'NR==2 {print $5}' | sed 's/%//')
if [ "$DISK_USAGE" -gt 80 ]; then
    echo "⚠️  Disk usage high: ${DISK_USAGE}%"
else
    echo "✅ Disk usage OK: ${DISK_USAGE}%"
fi

echo "✅ All health checks passed"
```

### Performance Monitoring

```sql
-- Monitor query performance
SELECT 
    query,
    calls,
    total_time,
    mean_time,
    rows
FROM pg_stat_statements 
WHERE query LIKE '%agent_memory%'
ORDER BY total_time DESC 
LIMIT 10;

-- Monitor table statistics
SELECT 
    schemaname,
    tablename,
    n_tup_ins as inserts,
    n_tup_upd as updates,
    n_tup_del as deletes,
    n_tup_hot_upd as hot_updates,
    n_live_tup as live_tuples,
    n_dead_tup as dead_tuples
FROM pg_stat_user_tables 
WHERE tablename IN ('agent_memory', 'chat_history', 'documents');

-- Monitor index usage
SELECT 
    indexrelname,
    idx_scan,
    idx_tup_read,
    idx_tup_fetch
FROM pg_stat_user_indexes 
WHERE indexrelname LIKE '%memory%'
ORDER BY idx_scan DESC;
```

### Backup and Recovery

```bash
#!/bin/bash
# backup.sh

DB_URL="postgres://agentflow:password@localhost:5432/agentflow"
BACKUP_DIR="/backups/agentflow"
DATE=$(date +%Y%m%d_%H%M%S)

# Create backup directory
mkdir -p "$BACKUP_DIR"

# Full database backup
echo "Creating full backup..."
pg_dump "$DB_URL" > "$BACKUP_DIR/agentflow_full_$DATE.sql"

# Compressed backup
echo "Creating compressed backup..."
pg_dump "$DB_URL" | gzip > "$BACKUP_DIR/agentflow_compressed_$DATE.sql.gz"

# Schema-only backup
echo "Creating schema backup..."
pg_dump --schema-only "$DB_URL" > "$BACKUP_DIR/agentflow_schema_$DATE.sql"

# Data-only backup
echo "Creating data backup..."
pg_dump --data-only "$DB_URL" > "$BACKUP_DIR/agentflow_data_$DATE.sql"

# Cleanup old backups (keep last 7 days)
find "$BACKUP_DIR" -name "*.sql*" -mtime +7 -delete

echo "✅ Backup completed: $BACKUP_DIR"
```

### Maintenance Tasks

```bash
#!/bin/bash
# maintenance.sh

DB_URL="postgres://agentflow:password@localhost:5432/agentflow"

echo "🔧 Running maintenance tasks..."

# Update statistics
echo "Updating table statistics..."
psql "$DB_URL" -c "ANALYZE;"

# Vacuum tables
echo "Vacuuming tables..."
psql "$DB_URL" -c "VACUUM ANALYZE agent_memory;"
psql "$DB_URL" -c "VACUUM ANALYZE chat_history;"
psql "$DB_URL" -c "VACUUM ANALYZE documents;"

# Reindex if needed (run during low traffic)
echo "Checking index bloat..."
psql "$DB_URL" -c "
SELECT 
    schemaname, 
    tablename, 
    indexname,
    pg_size_pretty(pg_relation_size(indexrelid)) as size
FROM pg_stat_user_indexes 
WHERE indexrelname LIKE '%memory%'
ORDER BY pg_relation_size(indexrelid) DESC;
"

echo "✅ Maintenance completed"
```

---

## 🔧 Troubleshooting

### Common Issues

#### 1. Connection Refused

```bash
# Check if PostgreSQL is running
docker ps | grep postgres
# or for native installation
sudo systemctl status postgresql

# Check port availability
netstat -tlnp | grep 5432

# Test connection
telnet localhost 5432
```

#### 2. pgvector Extension Missing

```sql
-- Check if extension is available
SELECT * FROM pg_available_extensions WHERE name = 'vector';

-- Install extension
CREATE EXTENSION IF NOT EXISTS vector;

-- Verify installation
SELECT * FROM pg_extension WHERE extname = 'vector';
```

#### 3. Permission Denied

```sql
-- Grant necessary permissions
GRANT ALL PRIVILEGES ON DATABASE agentflow TO agentflow;
GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO agentflow;
GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA public TO agentflow;
```

#### 4. Slow Queries

```sql
-- Check for missing indexes
SELECT 
    schemaname,
    tablename,
    attname,
    n_distinct,
    correlation
FROM pg_stats 
WHERE tablename = 'agent_memory';

-- Create missing indexes
CREATE INDEX CONCURRENTLY idx_agent_memory_session 
ON agent_memory (session_id);
```

#### 5. High Memory Usage

```bash
# Check PostgreSQL memory usage
ps aux | grep postgres

# Monitor with htop
htop

# Adjust PostgreSQL settings
# Edit postgresql.conf:
shared_buffers = 128MB  # Reduce if needed
work_mem = 32MB         # Reduce if needed
```

### Debug Mode

Enable detailed logging:

```sql
-- Enable query logging
ALTER SYSTEM SET log_statement = 'all';
ALTER SYSTEM SET log_duration = on;
ALTER SYSTEM SET log_min_duration_statement = 0;

-- Reload configuration
SELECT pg_reload_conf();

-- Check logs
-- Docker: docker logs agentflow-postgres
-- Native: tail -f /var/log/postgresql/postgresql-*.log
```

### Recovery Procedures

#### Restore from Backup

```bash
# Stop applications using the database
# Restore full backup
psql "$DB_URL" < /backups/agentflow/agentflow_full_20240101_120000.sql

# Or restore compressed backup
gunzip -c /backups/agentflow/agentflow_compressed_20240101_120000.sql.gz | psql "$DB_URL"
```

#### Reset Database

```bash
# ⚠️ WARNING: This will delete all data!

# Drop and recreate database
docker exec -it agentflow-postgres psql -U agentflow -c "
DROP DATABASE IF EXISTS agentflow;
CREATE DATABASE agentflow;
"

# Reconnect and setup
docker exec -it agentflow-postgres psql -U agentflow -d agentflow -c "
CREATE EXTENSION IF NOT EXISTS vector;
"

# Tables will be recreated automatically by AgentFlow
```

---

## 🎯 Production Deployment

### Docker Production Setup

Create `docker-compose.prod.yml`:

```yaml
version: '3.8'

services:
  postgres:
    image: pgvector/pgvector:pg16
    container_name: agentflow-postgres-prod
    environment:
      POSTGRES_DB: agentflow
      POSTGRES_USER: agentflow
      POSTGRES_PASSWORD_FILE: /run/secrets/postgres_password
    ports:
      - "5432:5432"
    volumes:
      - postgres_data:/var/lib/postgresql/data
      - ./postgresql.conf:/etc/postgresql/postgresql.conf
      - ./init-db.sql:/docker-entrypoint-initdb.d/init-db.sql
    command: postgres -c config_file=/etc/postgresql/postgresql.conf
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U agentflow -d agentflow"]
      interval: 30s
      timeout: 10s
      retries: 3
      start_period: 60s
    restart: unless-stopped
    secrets:
      - postgres_password
    deploy:
      resources:
        limits:
          memory: 2G
          cpus: '1.0'
        reservations:
          memory: 1G
          cpus: '0.5'

secrets:
  postgres_password:
    file: ./secrets/postgres_password.txt

volumes:
  postgres_data:
    driver: local
```

### Security Considerations

```bash
# Use strong passwords
openssl rand -base64 32 > ./secrets/postgres_password.txt

# Restrict file permissions
chmod 600 ./secrets/postgres_password.txt

# Use SSL in production
# Add to postgresql.conf:
ssl = on
ssl_cert_file = '/etc/ssl/certs/server.crt'
ssl_key_file = '/etc/ssl/private/server.key'
```

### Monitoring Setup

```yaml
# Add to docker-compose.prod.yml
  postgres-exporter:
    image: prometheuscommunity/postgres-exporter
    environment:
      DATA_SOURCE_NAME: "postgresql://agentflow:password@postgres:5432/agentflow?sslmode=disable"
    ports:
      - "9187:9187"
    depends_on:
      - postgres
```

---

## 🎉 Summary

You now have a complete PostgreSQL + pgvector setup for AgentFlow:

✅ **Database Setup**: PostgreSQL with pgvector extension  
✅ **Performance Optimization**: Indexes, connection pooling, configuration tuning  
✅ **Monitoring**: Health checks, performance monitoring, maintenance tasks  
✅ **Production Ready**: Security, backups, recovery procedures  

### Next Steps

1. **Test your setup** with the provided health check script
2. **Configure AgentFlow** with your database connection
3. **Set up monitoring** and backup procedures
4. **Optimize performance** based on your workload

For more information:
- **[Memory System Guide](Memory.md)** - Complete API reference
- **[Memory Provider Setup](MemoryProviderSetup.md)** - All provider options
- **[Configuration Guide](Configuration.md)** - Advanced configuration