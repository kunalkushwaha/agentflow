package scaffold

import (
	"fmt"
	"os"
	"path/filepath"
)

// DockerComposeGenerator handles Docker Compose file generation for database providers
type DockerComposeGenerator struct {
	config ProjectConfig
}

// NewDockerComposeGenerator creates a new Docker Compose generator
func NewDockerComposeGenerator(config ProjectConfig) *DockerComposeGenerator {
	return &DockerComposeGenerator{config: config}
}

// GenerateDockerCompose creates Docker Compose files for database providers
func (dcg *DockerComposeGenerator) GenerateDockerCompose() error {
	if !dcg.config.MemoryEnabled {
		return nil // No Docker Compose needed if memory is not enabled
	}

	switch dcg.config.MemoryProvider {
	case "pgvector":
		return dcg.generatePgVectorCompose()
	case "weaviate":
		return dcg.generateWeaviateCompose()
	default:
		// No Docker Compose needed for in-memory provider
		return nil
	}
}

// generatePgVectorCompose creates Docker Compose file for PostgreSQL with pgvector
func (dcg *DockerComposeGenerator) generatePgVectorCompose() error {
	dockerComposeContent := `version: '3.8'

services:
  postgres:
    image: pgvector/pgvector:pg16
    container_name: agentflow-postgres
    environment:
      POSTGRES_DB: agentflow
      POSTGRES_USER: user
      POSTGRES_PASSWORD: password
    ports:
      - "15432:5432"
    volumes:
      - postgres_data:/var/lib/postgresql/data
      - ./init-db.sql:/docker-entrypoint-initdb.d/init-db.sql
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U user -d agentflow"]
      interval: 10s
      timeout: 5s
      retries: 5
    restart: unless-stopped

volumes:
  postgres_data:
    driver: local
`

	// Create Docker Compose file
	dockerComposePath := filepath.Join(dcg.config.Name, "docker-compose.yml")
	if err := os.WriteFile(dockerComposePath, []byte(dockerComposeContent), 0644); err != nil {
		return fmt.Errorf("failed to create docker-compose.yml: %w", err)
	}
	fmt.Printf("Created file: %s\n", dockerComposePath)

	// Create database initialization script
	if err := dcg.generatePgVectorInitScript(); err != nil {
		return err
	}

	// Create environment file template
	if err := dcg.generateEnvFile("pgvector"); err != nil {
		return err
	}

	return nil
}

// generateWeaviateCompose creates Docker Compose file for Weaviate
func (dcg *DockerComposeGenerator) generateWeaviateCompose() error {
	dockerComposeContent := `version: '3.8'

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
`

	// Create Docker Compose file
	dockerComposePath := filepath.Join(dcg.config.Name, "docker-compose.yml")
	if err := os.WriteFile(dockerComposePath, []byte(dockerComposeContent), 0644); err != nil {
		return fmt.Errorf("failed to create docker-compose.yml: %w", err)
	}
	fmt.Printf("Created file: %s\n", dockerComposePath)

	// Create environment file template
	if err := dcg.generateEnvFile("weaviate"); err != nil {
		return err
	}

	return nil
}

// generatePgVectorInitScript creates the database initialization script for pgvector
func (dcg *DockerComposeGenerator) generatePgVectorInitScript() error {
	// Get embedding dimensions from intelligence system
	dimensions := GetModelDimensions(dcg.config.EmbeddingProvider, dcg.config.EmbeddingModel)
	
	initScriptContent := fmt.Sprintf(`-- AgentFlow Database Initialization Script for pgvector
-- Configured for %d-dimensional vectors (compatible with %s embeddings)
-- This script sets up the database schema for AgentFlow memory system

-- Enable the pgvector extension
CREATE EXTENSION IF NOT EXISTS vector;

-- Create the agent_memory table for storing embeddings and content
CREATE TABLE IF NOT EXISTS agent_memory (
    id SERIAL PRIMARY KEY,
    content TEXT NOT NULL,
    embedding vector(%d),
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
    id VARCHAR(255) PRIMARY KEY,
    title TEXT,
    content TEXT NOT NULL,
    source TEXT,
    doc_type VARCHAR(50),
    metadata JSONB,
    tags TEXT[],
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    chunk_index INTEGER DEFAULT 0,
    chunk_total INTEGER DEFAULT 1
);

-- Create indexes for documents
CREATE INDEX IF NOT EXISTS idx_documents_source ON documents (source);
CREATE INDEX IF NOT EXISTS idx_documents_type ON documents (doc_type);
CREATE INDEX IF NOT EXISTS idx_documents_tags ON documents USING GIN (tags);
CREATE INDEX IF NOT EXISTS idx_documents_created_at ON documents (created_at);

-- Create a function to update the updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Create triggers to automatically update the updated_at column
CREATE TRIGGER update_agent_memory_updated_at BEFORE UPDATE ON agent_memory FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_documents_updated_at BEFORE UPDATE ON documents FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Grant permissions to the user
GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO "user";
GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA public TO "user";

-- Insert a test record to verify the setup
INSERT INTO agent_memory (content, tags, metadata) 
VALUES ('AgentFlow memory system initialized successfully', ARRAY['system', 'initialization'], '{"source": "init_script", "version": "1.0"}')
ON CONFLICT DO NOTHING;

-- Display setup completion message
DO $$
BEGIN
    RAISE NOTICE 'AgentFlow database setup completed successfully!';
    RAISE NOTICE 'Tables created: agent_memory, chat_history, documents';
    RAISE NOTICE 'Extensions enabled: vector (pgvector)';
    RAISE NOTICE 'Ready for AgentFlow memory operations.';
END $$;
`, dimensions, dcg.config.EmbeddingModel, dimensions, dimensions, dcg.config.EmbeddingModel, dimensions, dimensions, dcg.config.EmbeddingModel)

	initScriptPath := filepath.Join(dcg.config.Name, "init-db.sql")
	if err := os.WriteFile(initScriptPath, []byte(initScriptContent), 0644); err != nil {
		return fmt.Errorf("failed to create init-db.sql: %w", err)
	}
	fmt.Printf("Created file: %s\n", initScriptPath)

	return nil
}

// generateEnvFile creates environment variable template files
func (dcg *DockerComposeGenerator) generateEnvFile(provider string) error {
	var envContent string

	switch provider {
	case "pgvector":
		envContent = `# AgentFlow Environment Configuration
# Copy this file to .env and update the values

# Database Configuration
POSTGRES_DB=agentflow
POSTGRES_USER=user
POSTGRES_PASSWORD=password
DATABASE_URL=postgres://user:password@localhost:15432/agentflow?sslmode=disable

# LLM Provider Configuration (choose one)
# Azure OpenAI
AZURE_OPENAI_API_KEY=your-azure-openai-api-key
AZURE_OPENAI_ENDPOINT=https://your-resource.openai.azure.com/
AZURE_OPENAI_DEPLOYMENT=your-deployment-name

# OpenAI
OPENAI_API_KEY=your-openai-api-key

# Embedding Provider Configuration
# For OpenAI embeddings (if using OpenAI embedding provider)
# OPENAI_API_KEY=your-openai-api-key

# For Ollama embeddings (default)
OLLAMA_BASE_URL=http://localhost:11434
OLLAMA_MODEL=mxbai-embed-large

# AgentFlow Configuration
AGENTFLOW_LOG_LEVEL=info
AGENTFLOW_SESSION_ID=default
`
	case "weaviate":
		envContent = `# AgentFlow Environment Configuration
# Copy this file to .env and update the values

# Weaviate Configuration
WEAVIATE_URL=http://localhost:8080
WEAVIATE_CLASS_NAME=AgentMemory

# LLM Provider Configuration (choose one)
# Azure OpenAI
AZURE_OPENAI_API_KEY=your-azure-openai-api-key
AZURE_OPENAI_ENDPOINT=https://your-resource.openai.azure.com/
AZURE_OPENAI_DEPLOYMENT=your-deployment-name

# OpenAI
OPENAI_API_KEY=your-openai-api-key

# Embedding Provider Configuration
# For OpenAI embeddings (if using OpenAI embedding provider)
# OPENAI_API_KEY=your-openai-api-key

# For Ollama embeddings (default)
OLLAMA_BASE_URL=http://localhost:11434
OLLAMA_MODEL=mxbai-embed-large

# AgentFlow Configuration
AGENTFLOW_LOG_LEVEL=info
AGENTFLOW_SESSION_ID=default
`
	}

	envExamplePath := filepath.Join(dcg.config.Name, ".env.example")
	if err := os.WriteFile(envExamplePath, []byte(envContent), 0644); err != nil {
		return fmt.Errorf("failed to create .env.example: %w", err)
	}
	fmt.Printf("Created file: %s\n", envExamplePath)

	return nil
}

// GenerateSetupScript creates setup scripts for easy database initialization
func (dcg *DockerComposeGenerator) GenerateSetupScript() error {
	if !dcg.config.MemoryEnabled {
		return nil
	}

	switch dcg.config.MemoryProvider {
	case "pgvector":
		return dcg.generatePgVectorSetupScript()
	case "weaviate":
		return dcg.generateWeaviateSetupScript()
	default:
		return nil
	}
}

// generatePgVectorSetupScript creates a setup script for pgvector
func (dcg *DockerComposeGenerator) generatePgVectorSetupScript() error {
	setupScriptContent := `#!/bin/bash
# AgentFlow PostgreSQL with pgvector Setup Script

echo "🚀 Setting up AgentFlow with PostgreSQL + pgvector..."

# Check if Docker is installed
if ! command -v docker &> /dev/null; then
    echo "❌ Docker is not installed. Please install Docker first."
    echo "   Visit: https://docs.docker.com/get-docker/"
    exit 1
fi

# Check if Docker Compose is available
if ! command -v docker-compose &> /dev/null && ! docker compose version &> /dev/null; then
    echo "❌ Docker Compose is not available. Please install Docker Compose."
    exit 1
fi

echo "✅ Docker and Docker Compose are available"

# Create .env file if it doesn't exist
if [ ! -f .env ]; then
    echo "📝 Creating .env file from template..."
    cp .env.example .env
    echo "⚠️  Please update the .env file with your actual API keys and configuration"
fi

# Start the database
echo "🐘 Starting PostgreSQL with pgvector..."
if command -v docker-compose &> /dev/null; then
    docker-compose up -d
else
    docker compose up -d
fi

# Wait for database to be ready
echo "⏳ Waiting for database to be ready..."
sleep 10

# Check if database is ready
echo "🔍 Checking database connection..."
if docker exec agentflow-postgres pg_isready -U user -d agentflow > /dev/null 2>&1; then
    echo "✅ Database is ready!"
    echo "📊 Database URL: postgres://user:password@localhost:15432/agentflow"
    echo ""
    echo "🎉 Setup complete! You can now run your AgentFlow project:"
    echo "   go mod tidy"
    echo "   go run . -m \"Your message here\""
    echo ""
    echo "📚 Useful commands:"
    echo "   - View logs: docker logs agentflow-postgres"
    echo "   - Stop database: docker-compose down"
    echo "   - Connect to database: docker exec -it agentflow-postgres psql -U user -d agentflow"
else
    echo "❌ Database is not ready. Please check the logs:"
    echo "   docker logs agentflow-postgres"
    exit 1
fi
`

	setupScriptPath := filepath.Join(dcg.config.Name, "setup.sh")
	if err := os.WriteFile(setupScriptPath, []byte(setupScriptContent), 0755); err != nil {
		return fmt.Errorf("failed to create setup.sh: %w", err)
	}
	fmt.Printf("Created file: %s (executable)\n", setupScriptPath)

	// Also create a Windows batch script
	windowsSetupContent := `@echo off
REM AgentFlow PostgreSQL with pgvector Setup Script for Windows

echo 🚀 Setting up AgentFlow with PostgreSQL + pgvector...

REM Check if Docker is installed
docker --version >nul 2>&1
if %errorlevel% neq 0 (
    echo ❌ Docker is not installed. Please install Docker Desktop first.
    echo    Visit: https://docs.docker.com/desktop/windows/
    pause
    exit /b 1
)

echo ✅ Docker is available

REM Create .env file if it doesn't exist
if not exist .env (
    echo 📝 Creating .env file from template...
    copy .env.example .env
    echo ⚠️  Please update the .env file with your actual API keys and configuration
)

REM Start the database
echo 🐘 Starting PostgreSQL with pgvector...
docker compose up -d

REM Wait for database to be ready
echo ⏳ Waiting for database to be ready...
timeout /t 10 /nobreak >nul

REM Check if database is ready
echo 🔍 Checking database connection...
docker exec agentflow-postgres pg_isready -U user -d agentflow >nul 2>&1
if %errorlevel% equ 0 (
    echo ✅ Database is ready!
    echo 📊 Database URL: postgres://user:password@localhost:15432/agentflow
    echo.
    echo 🎉 Setup complete! You can now run your AgentFlow project:
    echo    go mod tidy
    echo    go run . -m "Your message here"
    echo.
    echo 📚 Useful commands:
    echo    - View logs: docker logs agentflow-postgres
    echo    - Stop database: docker compose down
    echo    - Connect to database: docker exec -it agentflow-postgres psql -U user -d agentflow
) else (
    echo ❌ Database is not ready. Please check the logs:
    echo    docker logs agentflow-postgres
    pause
    exit /b 1
)

pause
`

	windowsSetupPath := filepath.Join(dcg.config.Name, "setup.bat")
	if err := os.WriteFile(windowsSetupPath, []byte(windowsSetupContent), 0644); err != nil {
		return fmt.Errorf("failed to create setup.bat: %w", err)
	}
	fmt.Printf("Created file: %s\n", windowsSetupPath)

	return nil
}

// generateWeaviateSetupScript creates a setup script for Weaviate
func (dcg *DockerComposeGenerator) generateWeaviateSetupScript() error {
	setupScriptContent := `#!/bin/bash
# AgentFlow Weaviate Setup Script

echo "🚀 Setting up AgentFlow with Weaviate..."

# Check if Docker is installed
if ! command -v docker &> /dev/null; then
    echo "❌ Docker is not installed. Please install Docker first."
    echo "   Visit: https://docs.docker.com/get-docker/"
    exit 1
fi

# Check if Docker Compose is available
if ! command -v docker-compose &> /dev/null && ! docker compose version &> /dev/null; then
    echo "❌ Docker Compose is not available. Please install Docker Compose."
    exit 1
fi

echo "✅ Docker and Docker Compose are available"

# Create .env file if it doesn't exist
if [ ! -f .env ]; then
    echo "📝 Creating .env file from template..."
    cp .env.example .env
    echo "⚠️  Please update the .env file with your actual API keys and configuration"
fi

# Start Weaviate
echo "🔍 Starting Weaviate..."
if command -v docker-compose &> /dev/null; then
    docker-compose up -d
else
    docker compose up -d
fi

# Wait for Weaviate to be ready
echo "⏳ Waiting for Weaviate to be ready..."
sleep 15

# Check if Weaviate is ready
echo "🔍 Checking Weaviate connection..."
if curl -f http://localhost:8080/v1/.well-known/ready > /dev/null 2>&1; then
    echo "✅ Weaviate is ready!"
    echo "🌐 Weaviate URL: http://localhost:8080"
    echo "📊 Weaviate Console: http://localhost:8080/v1/console"
    echo ""
    echo "🎉 Setup complete! You can now run your AgentFlow project:"
    echo "   go mod tidy"
    echo "   go run . -m \"Your message here\""
    echo ""
    echo "📚 Useful commands:"
    echo "   - View logs: docker logs agentflow-weaviate"
    echo "   - Stop Weaviate: docker-compose down"
    echo "   - Check status: curl http://localhost:8080/v1/.well-known/ready"
else
    echo "❌ Weaviate is not ready. Please check the logs:"
    echo "   docker logs agentflow-weaviate"
    exit 1
fi
`

	setupScriptPath := filepath.Join(dcg.config.Name, "setup.sh")
	if err := os.WriteFile(setupScriptPath, []byte(setupScriptContent), 0755); err != nil {
		return fmt.Errorf("failed to create setup.sh: %w", err)
	}
	fmt.Printf("Created file: %s (executable)\n", setupScriptPath)

	// Also create a Windows batch script
	windowsSetupContent := `@echo off
REM AgentFlow Weaviate Setup Script for Windows

echo 🚀 Setting up AgentFlow with Weaviate...

REM Check if Docker is installed
docker --version >nul 2>&1
if %errorlevel% neq 0 (
    echo ❌ Docker is not installed. Please install Docker Desktop first.
    echo    Visit: https://docs.docker.com/desktop/windows/
    pause
    exit /b 1
)

echo ✅ Docker is available

REM Create .env file if it doesn't exist
if not exist .env (
    echo 📝 Creating .env file from template...
    copy .env.example .env
    echo ⚠️  Please update the .env file with your actual API keys and configuration
)

REM Start Weaviate
echo 🔍 Starting Weaviate...
docker compose up -d

REM Wait for Weaviate to be ready
echo ⏳ Waiting for Weaviate to be ready...
timeout /t 15 /nobreak >nul

REM Check if Weaviate is ready
echo 🔍 Checking Weaviate connection...
curl -f http://localhost:8080/v1/.well-known/ready >nul 2>&1
if %errorlevel% equ 0 (
    echo ✅ Weaviate is ready!
    echo 🌐 Weaviate URL: http://localhost:8080
    echo 📊 Weaviate Console: http://localhost:8080/v1/console
    echo.
    echo 🎉 Setup complete! You can now run your AgentFlow project:
    echo    go mod tidy
    echo    go run . -m "Your message here"
    echo.
    echo 📚 Useful commands:
    echo    - View logs: docker logs agentflow-weaviate
    echo    - Stop Weaviate: docker compose down
    echo    - Check status: curl http://localhost:8080/v1/.well-known/ready
) else (
    echo ❌ Weaviate is not ready. Please check the logs:
    echo    docker logs agentflow-weaviate
    pause
    exit /b 1
)

pause
`

	windowsSetupPath := filepath.Join(dcg.config.Name, "setup.bat")
	if err := os.WriteFile(windowsSetupPath, []byte(windowsSetupContent), 0644); err != nil {
		return fmt.Errorf("failed to create setup.bat: %w", err)
	}
	fmt.Printf("Created file: %s\n", windowsSetupPath)

	return nil
}