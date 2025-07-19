# Weaviate Setup Guide

**Complete guide for setting up Weaviate vector database for AgentFlow memory**

This guide provides detailed instructions for setting up Weaviate as a vector database backend for AgentFlow's memory system, from development to production deployment.

## 🎯 Overview

Weaviate provides:
- **Purpose-built vector database** optimized for similarity search
- **GraphQL API** for flexible queries and data management
- **Built-in clustering** for horizontal scalability
- **Advanced search features** including hybrid search and filtering
- **Rich ecosystem** with integrations and modules

## 📋 Prerequisites

- Docker (recommended) OR Kubernetes
- 4GB+ RAM available (8GB+ recommended for production)
- Network access to Weaviate port (8080)
- Basic command line knowledge

## 🚀 Quick Start (Docker)

### Step 1: Run Weaviate

```bash
# Create and start Weaviate container
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
echo "Waiting for Weaviate to start..."
sleep 15

# Verify Weaviate is running
curl http://localhost:8080/v1/.well-known/ready
```

### Step 2: Test Connection

```bash
# Check Weaviate status
curl http://localhost:8080/v1/meta

# Expected response: JSON with version and modules info
```

### Step 3: Configure AgentFlow

Create `agentflow.toml`:

```toml
[memory]
enabled = true
provider = "weaviate"
max_results = 10
dimensions = 1536
auto_embed = true

[memory.weaviate]
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

Set environment variables:

```bash
export WEAVIATE_URL="http://localhost:8080"
export OPENAI_API_KEY="your-openai-api-key"
```

**✅ You're ready to use AgentFlow with Weaviate!**

---

## 🔧 Detailed Setup Options

### Option 1: Docker Compose (Recommended)

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
      start_period: 30s
    restart: unless-stopped

volumes:
  weaviate_data:
    driver: local
```

Start Weaviate:

```bash
# Start Weaviate
docker-compose up -d

# Check logs
docker-compose logs weaviate

# Test connection
curl http://localhost:8080/v1/.well-known/ready

# Stop when needed
docker-compose down
```

### Option 2: Production Docker Compose

Create `docker-compose.prod.yml`:

```yaml
version: '3.8'

services:
  weaviate:
    image: semitechnologies/weaviate:1.22.4
    container_name: agentflow-weaviate-prod
    ports:
      - "8080:8080"
    environment:
      # Core settings
      QUERY_DEFAULTS_LIMIT: 25
      PERSISTENCE_DATA_PATH: '/var/lib/weaviate'
      DEFAULT_VECTORIZER_MODULE: 'none'
      CLUSTER_HOSTNAME: 'node1'
      
      # Authentication (production)
      AUTHENTICATION_ANONYMOUS_ACCESS_ENABLED: 'false'
      AUTHENTICATION_APIKEY_ENABLED: 'true'
      AUTHENTICATION_APIKEY_ALLOWED_KEYS: 'your-secret-api-key'
      AUTHENTICATION_APIKEY_USERS: 'admin'
      
      # Modules
      ENABLE_MODULES: 'text2vec-openai,text2vec-cohere,text2vec-huggingface,ref2vec-centroid,generative-openai,qna-openai'
      
      # Performance settings
      LIMIT_RESOURCES: 'true'
      GOMEMLIMIT: '4GiB'
      
      # Backup settings
      BACKUP_FILESYSTEM_PATH: '/var/lib/weaviate/backups'
      
    volumes:
      - weaviate_data:/var/lib/weaviate
      - weaviate_backups:/var/lib/weaviate/backups
    healthcheck:
      test: ["CMD", "curl", "-f", "-H", "Authorization: Bearer your-secret-api-key", "http://localhost:8080/v1/.well-known/ready"]
      interval: 30s
      timeout: 10s
      retries: 3
      start_period: 60s
    restart: unless-stopped
    deploy:
      resources:
        limits:
          memory: 4G
          cpus: '2.0'
        reservations:
          memory: 2G
          cpus: '1.0'

volumes:
  weaviate_data:
    driver: local
  weaviate_backups:
    driver: local
```

### Option 3: Kubernetes Deployment

Create `weaviate-k8s.yaml`:

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: weaviate
  labels:
    app: weaviate
spec:
  replicas: 1
  selector:
    matchLabels:
      app: weaviate
  template:
    metadata:
      labels:
        app: weaviate
    spec:
      containers:
      - name: weaviate
        image: semitechnologies/weaviate:1.22.4
        ports:
        - containerPort: 8080
        env:
        - name: QUERY_DEFAULTS_LIMIT
          value: "25"
        - name: AUTHENTICATION_ANONYMOUS_ACCESS_ENABLED
          value: "true"
        - name: PERSISTENCE_DATA_PATH
          value: "/var/lib/weaviate"
        - name: DEFAULT_VECTORIZER_MODULE
          value: "none"
        - name: CLUSTER_HOSTNAME
          value: "node1"
        - name: ENABLE_MODULES
          value: "text2vec-openai,text2vec-cohere,text2vec-huggingface,ref2vec-centroid,generative-openai,qna-openai"
        volumeMounts:
        - name: weaviate-storage
          mountPath: /var/lib/weaviate
        resources:
          requests:
            memory: "2Gi"
            cpu: "1"
          limits:
            memory: "4Gi"
            cpu: "2"
        livenessProbe:
          httpGet:
            path: /v1/.well-known/ready
            port: 8080
          initialDelaySeconds: 30
          periodSeconds: 30
        readinessProbe:
          httpGet:
            path: /v1/.well-known/ready
            port: 8080
          initialDelaySeconds: 10
          periodSeconds: 10
      volumes:
      - name: weaviate-storage
        persistentVolumeClaim:
          claimName: weaviate-pvc
---
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: weaviate-pvc
spec:
  accessModes:
    - ReadWriteOnce
  resources:
    requests:
      storage: 20Gi
---
apiVersion: v1
kind: Service
metadata:
  name: weaviate-service
spec:
  selector:
    app: weaviate
  ports:
    - protocol: TCP
      port: 8080
      targetPort: 8080
  type: LoadBalancer
```

Deploy to Kubernetes:

```bash
# Apply the configuration
kubectl apply -f weaviate-k8s.yaml

# Check deployment status
kubectl get pods -l app=weaviate
kubectl get services weaviate-service

# Get external IP
kubectl get service weaviate-service
```

---

## ⚙️ Configuration

### Connection Options

```bash
# Basic connection
http://localhost:8080

# With authentication
http://localhost:8080 (with API key header)

# Remote connection
http://your-weaviate-host:8080

# HTTPS (production)
https://your-weaviate-host:8080
```

### AgentFlow Configuration

#### Basic Configuration

```toml
[memory]
enabled = true
provider = "weaviate"
max_results = 10
dimensions = 1536
auto_embed = true

[memory.weaviate]
connection = "http://localhost:8080"
class_name = "AgentMemory"

[memory.embedding]
provider = "openai"
model = "text-embedding-3-small"
```

#### Production Configuration

```toml
[memory]
enabled = true
provider = "weaviate"
max_results = 10
dimensions = 1536
auto_embed = true

[memory.weaviate]
connection = "https://your-weaviate-host:8080"
api_key = "${WEAVIATE_API_KEY}"
class_name = "AgentMemory"
timeout = "30s"
max_retries = 3

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
health_check_interval = "1m"
```

### Environment Variables

```bash
# Weaviate connection
export WEAVIATE_URL="http://localhost:8080"
export WEAVIATE_API_KEY="your-secret-api-key"  # For production
export WEAVIATE_CLASS_NAME="AgentMemory"

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

## 🚀 Advanced Features

### Authentication Setup

#### API Key Authentication

```yaml
# docker-compose.yml with authentication
services:
  weaviate:
    environment:
      AUTHENTICATION_ANONYMOUS_ACCESS_ENABLED: 'false'
      AUTHENTICATION_APIKEY_ENABLED: 'true'
      AUTHENTICATION_APIKEY_ALLOWED_KEYS: 'your-secret-key,another-key'
      AUTHENTICATION_APIKEY_USERS: 'admin,user'
```

```toml
# agentflow.toml with authentication
[memory.weaviate]
connection = "http://localhost:8080"
api_key = "your-secret-key"
class_name = "AgentMemory"
```

#### OIDC Authentication

```yaml
services:
  weaviate:
    environment:
      AUTHENTICATION_ANONYMOUS_ACCESS_ENABLED: 'false'
      AUTHENTICATION_OIDC_ENABLED: 'true'
      AUTHENTICATION_OIDC_ISSUER: 'https://your-oidc-provider.com'
      AUTHENTICATION_OIDC_CLIENT_ID: 'your-client-id'
      AUTHENTICATION_OIDC_USERNAME_CLAIM: 'email'
      AUTHENTICATION_OIDC_GROUPS_CLAIM: 'groups'
```

### Clustering Setup

#### Multi-Node Cluster

```yaml
# docker-compose.cluster.yml
version: '3.8'

services:
  weaviate-node1:
    image: semitechnologies/weaviate:1.22.4
    container_name: weaviate-node1
    ports:
      - "8080:8080"
    environment:
      CLUSTER_HOSTNAME: 'node1'
      CLUSTER_GOSSIP_BIND_PORT: '7100'
      CLUSTER_DATA_BIND_PORT: '7101'
      PERSISTENCE_DATA_PATH: '/var/lib/weaviate'
      DEFAULT_VECTORIZER_MODULE: 'none'
      ENABLE_MODULES: 'text2vec-openai,text2vec-cohere,text2vec-huggingface'
    volumes:
      - weaviate_node1_data:/var/lib/weaviate
    networks:
      - weaviate-cluster

  weaviate-node2:
    image: semitechnologies/weaviate:1.22.4
    container_name: weaviate-node2
    ports:
      - "8081:8080"
    environment:
      CLUSTER_HOSTNAME: 'node2'
      CLUSTER_GOSSIP_BIND_PORT: '7102'
      CLUSTER_DATA_BIND_PORT: '7103'
      CLUSTER_JOIN: 'node1:7100'
      PERSISTENCE_DATA_PATH: '/var/lib/weaviate'
      DEFAULT_VECTORIZER_MODULE: 'none'
      ENABLE_MODULES: 'text2vec-openai,text2vec-cohere,text2vec-huggingface'
    volumes:
      - weaviate_node2_data:/var/lib/weaviate
    networks:
      - weaviate-cluster
    depends_on:
      - weaviate-node1

  weaviate-node3:
    image: semitechnologies/weaviate:1.22.4
    container_name: weaviate-node3
    ports:
      - "8082:8080"
    environment:
      CLUSTER_HOSTNAME: 'node3'
      CLUSTER_GOSSIP_BIND_PORT: '7104'
      CLUSTER_DATA_BIND_PORT: '7105'
      CLUSTER_JOIN: 'node1:7100'
      PERSISTENCE_DATA_PATH: '/var/lib/weaviate'
      DEFAULT_VECTORIZER_MODULE: 'none'
      ENABLE_MODULES: 'text2vec-openai,text2vec-cohere,text2vec-huggingface'
    volumes:
      - weaviate_node3_data:/var/lib/weaviate
    networks:
      - weaviate-cluster
    depends_on:
      - weaviate-node1

volumes:
  weaviate_node1_data:
  weaviate_node2_data:
  weaviate_node3_data:

networks:
  weaviate-cluster:
    driver: bridge
```

### Backup and Restore

#### Automated Backups

```yaml
# Add backup service to docker-compose.yml
services:
  weaviate-backup:
    image: curlimages/curl:latest
    container_name: weaviate-backup
    volumes:
      - ./backups:/backups
      - ./scripts:/scripts
    command: /scripts/backup.sh
    depends_on:
      - weaviate
    restart: "no"
```

Create `scripts/backup.sh`:

```bash
#!/bin/sh
# backup.sh

WEAVIATE_URL="http://weaviate:8080"
BACKUP_DIR="/backups"
DATE=$(date +%Y%m%d_%H%M%S)

echo "Creating Weaviate backup..."

# Create backup
curl -X POST \
  "$WEAVIATE_URL/v1/backups/filesystem" \
  -H "Content-Type: application/json" \
  -d "{
    \"id\": \"backup_$DATE\",
    \"include\": [\"AgentMemory\", \"Documents\", \"ChatHistory\"]
  }"

# Wait for backup to complete
sleep 30

# Check backup status
curl "$WEAVIATE_URL/v1/backups/filesystem/backup_$DATE"

echo "Backup completed: backup_$DATE"
```

#### Manual Backup

```bash
# Create backup
curl -X POST \
  "http://localhost:8080/v1/backups/filesystem" \
  -H "Content-Type: application/json" \
  -d '{
    "id": "my-backup-2024",
    "include": ["AgentMemory"]
  }'

# Check backup status
curl "http://localhost:8080/v1/backups/filesystem/my-backup-2024"

# List all backups
curl "http://localhost:8080/v1/backups/filesystem"
```

#### Restore from Backup

```bash
# Restore backup
curl -X POST \
  "http://localhost:8080/v1/backups/filesystem/my-backup-2024/restore" \
  -H "Content-Type: application/json" \
  -d '{
    "include": ["AgentMemory"]
  }'

# Check restore status
curl "http://localhost:8080/v1/backups/filesystem/my-backup-2024/restore"
```

---

## 📊 Monitoring and Maintenance

### Health Checks

```bash
#!/bin/bash
# health-check.sh

WEAVIATE_URL="http://localhost:8080"
API_KEY=""  # Set if using authentication

echo "🔍 Checking Weaviate health..."

# Test connection
if curl -f "$WEAVIATE_URL/v1/.well-known/ready" > /dev/null 2>&1; then
    echo "✅ Weaviate connection successful"
else
    echo "❌ Weaviate connection failed"
    exit 1
fi

# Check cluster status
NODES=$(curl -s "$WEAVIATE_URL/v1/nodes" | jq -r '.nodes | length')
if [ "$NODES" -gt 0 ]; then
    echo "✅ Cluster has $NODES node(s)"
else
    echo "⚠️  No cluster nodes found"
fi

# Check schema
CLASSES=$(curl -s "$WEAVIATE_URL/v1/schema" | jq -r '.classes | length')
echo "ℹ️  Schema has $CLASSES class(es)"

# Check disk space (Docker)
if command -v docker &> /dev/null; then
    CONTAINER_ID=$(docker ps -q -f name=weaviate)
    if [ -n "$CONTAINER_ID" ]; then
        DISK_USAGE=$(docker exec "$CONTAINER_ID" df -h /var/lib/weaviate | awk 'NR==2 {print $5}' | sed 's/%//')
        if [ "$DISK_USAGE" -gt 80 ]; then
            echo "⚠️  Disk usage high: ${DISK_USAGE}%"
        else
            echo "✅ Disk usage OK: ${DISK_USAGE}%"
        fi
    fi
fi

echo "✅ All health checks passed"
```

### Performance Monitoring

```bash
# Monitor Weaviate metrics
curl "http://localhost:8080/v1/meta" | jq '.'

# Check cluster statistics
curl "http://localhost:8080/v1/nodes" | jq '.nodes[] | {name: .name, status: .status, stats: .stats}'

# Monitor specific class
curl "http://localhost:8080/v1/schema/AgentMemory" | jq '.'

# Get object count
curl "http://localhost:8080/v1/objects?class=AgentMemory&limit=0" | jq '.totalResults'
```

### GraphQL Queries

```bash
# Query objects with GraphQL
curl -X POST \
  "http://localhost:8080/v1/graphql" \
  -H "Content-Type: application/json" \
  -d '{
    "query": "{ Get { AgentMemory(limit: 10) { content tags _additional { id creationTimeUnix } } } }"
  }'

# Aggregate queries
curl -X POST \
  "http://localhost:8080/v1/graphql" \
  -H "Content-Type: application/json" \
  -d '{
    "query": "{ Aggregate { AgentMemory { meta { count } } } }"
  }'

# Vector search with GraphQL
curl -X POST \
  "http://localhost:8080/v1/graphql" \
  -H "Content-Type: application/json" \
  -d '{
    "query": "{ Get { AgentMemory(nearVector: {vector: [0.1, 0.2, 0.3]}, limit: 5) { content _additional { distance } } } }"
  }'
```

### Maintenance Tasks

```bash
#!/bin/bash
# maintenance.sh

WEAVIATE_URL="http://localhost:8080"

echo "🔧 Running Weaviate maintenance tasks..."

# Check cluster health
echo "Checking cluster health..."
curl -s "$WEAVIATE_URL/v1/nodes" | jq '.nodes[] | {name: .name, status: .status}'

# Optimize indexes (if needed)
echo "Checking index status..."
curl -s "$WEAVIATE_URL/v1/schema" | jq '.classes[] | {class: .class, vectorIndexType: .vectorIndexType}'

# Clean up old backups (keep last 7)
echo "Cleaning up old backups..."
BACKUPS=$(curl -s "$WEAVIATE_URL/v1/backups/filesystem" | jq -r '.[] | select(.status == "SUCCESS") | .id' | sort -r | tail -n +8)
for backup in $BACKUPS; do
    echo "Deleting old backup: $backup"
    curl -X DELETE "$WEAVIATE_URL/v1/backups/filesystem/$backup"
done

echo "✅ Maintenance completed"
```

---

## 🔧 Troubleshooting

### Common Issues

#### 1. Connection Refused

```bash
# Check if Weaviate is running
docker ps | grep weaviate

# Check logs
docker logs agentflow-weaviate

# Check port availability
netstat -tlnp | grep 8080

# Test connection
curl http://localhost:8080/v1/.well-known/ready
```

#### 2. Authentication Errors

```bash
# Test with API key
curl -H "Authorization: Bearer your-api-key" \
  "http://localhost:8080/v1/.well-known/ready"

# Check authentication configuration
docker exec agentflow-weaviate env | grep AUTH
```

#### 3. Schema Issues

```bash
# Check current schema
curl "http://localhost:8080/v1/schema" | jq '.'

# Delete class (careful!)
curl -X DELETE "http://localhost:8080/v1/schema/AgentMemory"

# Schema will be recreated automatically by AgentFlow
```

#### 4. Performance Issues

```bash
# Check resource usage
docker stats agentflow-weaviate

# Monitor memory usage
curl "http://localhost:8080/v1/meta" | jq '.hostname, .version'

# Check for memory leaks
docker exec agentflow-weaviate ps aux
```

#### 5. Cluster Issues

```bash
# Check cluster status
curl "http://localhost:8080/v1/nodes" | jq '.nodes[] | {name: .name, status: .status}'

# Check gossip protocol
docker logs agentflow-weaviate | grep -i gossip

# Restart problematic nodes
docker restart weaviate-node2
```

### Debug Mode

Enable debug logging:

```yaml
# docker-compose.yml
services:
  weaviate:
    environment:
      LOG_LEVEL: 'debug'
      PROMETHEUS_MONITORING_ENABLED: 'true'
```

### Recovery Procedures

#### Reset Weaviate

```bash
# ⚠️ WARNING: This will delete all data!

# Stop Weaviate
docker-compose down

# Remove data volume
docker volume rm $(docker volume ls -q | grep weaviate)

# Restart Weaviate
docker-compose up -d

# Schema and data will be recreated by AgentFlow
```

#### Cluster Recovery

```bash
# If cluster is in split-brain state:

# Stop all nodes
docker-compose down

# Start nodes one by one
docker-compose up -d weaviate-node1
sleep 30
docker-compose up -d weaviate-node2
sleep 30
docker-compose up -d weaviate-node3

# Check cluster status
curl "http://localhost:8080/v1/nodes"
```

---

## 🎯 Production Deployment

### Security Checklist

```bash
# ✅ Enable authentication
AUTHENTICATION_ANONYMOUS_ACCESS_ENABLED: 'false'
AUTHENTICATION_APIKEY_ENABLED: 'true'

# ✅ Use HTTPS in production
# Configure reverse proxy (nginx, traefik, etc.)

# ✅ Restrict network access
# Use firewall rules or security groups

# ✅ Regular backups
# Automated backup schedule

# ✅ Monitor resource usage
# Set up alerts for CPU, memory, disk

# ✅ Update regularly
# Keep Weaviate version up to date
```

### Load Balancer Setup

```nginx
# nginx.conf
upstream weaviate_cluster {
    server weaviate-node1:8080;
    server weaviate-node2:8080;
    server weaviate-node3:8080;
}

server {
    listen 80;
    server_name your-weaviate-domain.com;
    
    location / {
        proxy_pass http://weaviate_cluster;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```

### Monitoring Setup

```yaml
# Add Prometheus monitoring
services:
  weaviate:
    environment:
      PROMETHEUS_MONITORING_ENABLED: 'true'
      PROMETHEUS_MONITORING_PORT: '2112'
    ports:
      - "2112:2112"  # Prometheus metrics

  prometheus:
    image: prom/prometheus
    ports:
      - "9090:9090"
    volumes:
      - ./prometheus.yml:/etc/prometheus/prometheus.yml
```

---

## 🎉 Summary

You now have a complete Weaviate setup for AgentFlow:

✅ **Vector Database**: Purpose-built Weaviate vector database  
✅ **Scalability**: Clustering and load balancing support  
✅ **Security**: Authentication and access control  
✅ **Monitoring**: Health checks, metrics, and maintenance  
✅ **Production Ready**: Backup, recovery, and deployment procedures  

### Next Steps

1. **Test your setup** with the provided health check script
2. **Configure AgentFlow** with your Weaviate connection
3. **Set up monitoring** and backup procedures
4. **Scale horizontally** with clustering if needed

For more information:
- **[Memory System Guide](Memory.md)** - Complete API reference
- **[Memory Provider Setup](MemoryProviderSetup.md)** - All provider options
- **[Configuration Guide](Configuration.md)** - Advanced configuration