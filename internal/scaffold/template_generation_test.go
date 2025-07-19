package scaffold

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCreateAgentProjectModular(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()
	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(tempDir)

	tests := []struct {
		name   string
		config ProjectConfig
	}{
		{
			name: "Basic project without memory",
			config: ProjectConfig{
				Name:          "test-basic",
				NumAgents:     2,
				Provider:      "azure",
				ResponsibleAI: true,
				ErrorHandler:  true,
				MemoryEnabled: false,
			},
		},
		{
			name: "Project with memory and RAG",
			config: ProjectConfig{
				Name:                "test-memory",
				NumAgents:           2,
				Provider:            "azure",
				ResponsibleAI:       true,
				ErrorHandler:        true,
				MemoryEnabled:       true,
				MemoryProvider:      "pgvector",
				EmbeddingProvider:   "ollama",
				EmbeddingModel:      "nomic-embed-text:latest",
				EmbeddingDimensions: 768,
				RAGEnabled:          true,
				RAGChunkSize:        1000,
				RAGOverlap:          100,
				RAGTopK:             5,
				RAGScoreThreshold:   0.7,
			},
		},
		{
			name: "Project with MCP integration",
			config: ProjectConfig{
				Name:          "test-mcp",
				NumAgents:     1,
				Provider:      "openai",
				ResponsibleAI: false,
				ErrorHandler:  false,
				MCPEnabled:    true,
				MCPProduction: false,
				WithCache:     true,
				WithMetrics:   false,
				MCPTools:      []string{"web_search", "summarize"},
				MCPServers:    []string{"docker"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := CreateAgentProjectModular(tt.config)
			if err != nil {
				t.Errorf("CreateAgentProjectModular() error = %v", err)
				return
			}

			// Verify project directory was created
			projectPath := filepath.Join(tempDir, tt.config.Name)
			if _, err := os.Stat(projectPath); os.IsNotExist(err) {
				t.Errorf("Project directory not created: %s", projectPath)
				return
			}

			// Verify essential files were created
			essentialFiles := []string{
				"go.mod",
				"README.md",
				"main.go",
				"agentflow.toml",
			}

			for _, file := range essentialFiles {
				filePath := filepath.Join(projectPath, file)
				if _, err := os.Stat(filePath); os.IsNotExist(err) {
					t.Errorf("Essential file not created: %s", file)
				}
			}

			// Verify agent files were created
			expectedAgents := tt.config.NumAgents
			if expectedAgents == 0 {
				expectedAgents = 1 // Default
			}

			for i := 1; i <= expectedAgents; i++ {
				agentFile := filepath.Join(projectPath, fmt.Sprintf("agent%d.go", i))
				if _, err := os.Stat(agentFile); os.IsNotExist(err) {
					t.Errorf("Agent file not created: agent%d.go", i)
				}
			}

			// Memory-specific validations
			if tt.config.MemoryEnabled {
				// Check if Docker files were created for database providers
				if tt.config.MemoryProvider == "pgvector" || tt.config.MemoryProvider == "weaviate" {
					dockerFile := filepath.Join(projectPath, "docker-compose.yml")
					if _, err := os.Stat(dockerFile); os.IsNotExist(err) {
						t.Errorf("Docker compose file not created for memory provider: %s", tt.config.MemoryProvider)
					}

					setupScript := filepath.Join(projectPath, "setup.sh")
					if _, err := os.Stat(setupScript); os.IsNotExist(err) {
						t.Errorf("Setup script not created for memory provider: %s", tt.config.MemoryProvider)
					}
				}

				// Verify TOML contains memory configuration
				tomlPath := filepath.Join(projectPath, "agentflow.toml")
				tomlContent, err := os.ReadFile(tomlPath)
				if err != nil {
					t.Errorf("Failed to read agentflow.toml: %v", err)
				} else {
					tomlStr := string(tomlContent)
					if !strings.Contains(tomlStr, "[agent_memory]") {
						t.Errorf("agentflow.toml missing [agent_memory] section")
					}
					if !strings.Contains(tomlStr, fmt.Sprintf("dimensions = %d", tt.config.EmbeddingDimensions)) {
						t.Errorf("agentflow.toml missing correct dimensions configuration")
					}
				}

				// Verify main.go contains memory validation
				mainPath := filepath.Join(projectPath, "main.go")
				mainContent, err := os.ReadFile(mainPath)
				if err != nil {
					t.Errorf("Failed to read main.go: %v", err)
				} else {
					mainStr := string(mainContent)
					if !strings.Contains(mainStr, "validateMemoryConfig") {
						t.Errorf("main.go missing memory validation function")
					}
					if !strings.Contains(mainStr, "config.AgentMemory") {
						t.Errorf("main.go not using config.AgentMemory pattern")
					}
				}
			}

			// Clean up for next test
			os.RemoveAll(projectPath)
		})
	}
}

func TestTemplateGeneration(t *testing.T) {
	tests := []struct {
		name           string
		config         ProjectConfig
		expectedInMain []string
		expectedInToml []string
	}{
		{
			name: "Memory configuration in templates",
			config: ProjectConfig{
				Name:                "test-template",
				NumAgents:           1,
				Provider:            "azure",
				MemoryEnabled:       true,
				MemoryProvider:      "pgvector",
				EmbeddingProvider:   "ollama",
				EmbeddingModel:      "nomic-embed-text:latest",
				EmbeddingDimensions: 768,
				RAGEnabled:          true,
				RAGChunkSize:        1000,
				RAGOverlap:          100,
				RAGTopK:             5,
				RAGScoreThreshold:   0.7,
			},
			expectedInMain: []string{
				"validateMemoryConfig",
				"config.AgentMemory",
				"Memory configuration validation failed",
				"PostgreSQL/PgVector Troubleshooting",
				"Ollama Troubleshooting",
			},
			expectedInToml: []string{
				"[agent_memory]",
				"provider = \"pgvector\"",
				"dimensions = 768",
				"[agent_memory.embedding]",
				"provider = \"ollama\"",
				"model = \"nomic-embed-text:latest\"",
				"enable_rag = true",
				"chunk_size = 1000",
				"chunk_overlap = 100",
			},
		},
		{
			name: "OpenAI embedding configuration",
			config: ProjectConfig{
				Name:                "test-openai",
				NumAgents:           1,
				Provider:            "openai",
				MemoryEnabled:       true,
				MemoryProvider:      "weaviate",
				EmbeddingProvider:   "openai",
				EmbeddingModel:      "text-embedding-3-small",
				EmbeddingDimensions: 1536,
				RAGEnabled:          false,
			},
			expectedInMain: []string{
				"validateMemoryConfig",
				"text-embedding-3-small requires 1536 dimensions",
			},
			expectedInToml: []string{
				"[agent_memory]",
				"provider = \"weaviate\"",
				"dimensions = 1536",
				"[agent_memory.embedding]",
				"provider = \"openai\"",
				"model = \"text-embedding-3-small\"",
				"enable_rag = false",
			},
		},
	}

	tempDir := t.TempDir()
	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(tempDir)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := CreateAgentProjectModular(tt.config)
			if err != nil {
				t.Errorf("CreateAgentProjectModular() error = %v", err)
				return
			}

			projectPath := filepath.Join(tempDir, tt.config.Name)

			// Check main.go content
			mainPath := filepath.Join(projectPath, "main.go")
			mainContent, err := os.ReadFile(mainPath)
			if err != nil {
				t.Errorf("Failed to read main.go: %v", err)
			} else {
				mainStr := string(mainContent)
				for _, expected := range tt.expectedInMain {
					if !strings.Contains(mainStr, expected) {
						t.Errorf("main.go missing expected content: %s", expected)
					}
				}
			}

			// Check agentflow.toml content
			tomlPath := filepath.Join(projectPath, "agentflow.toml")
			tomlContent, err := os.ReadFile(tomlPath)
			if err != nil {
				t.Errorf("Failed to read agentflow.toml: %v", err)
			} else {
				tomlStr := string(tomlContent)
				for _, expected := range tt.expectedInToml {
					if !strings.Contains(tomlStr, expected) {
						t.Errorf("agentflow.toml missing expected content: %s", expected)
					}
				}
			}

			// Clean up
			os.RemoveAll(projectPath)
		})
	}
}

func TestDatabaseSchemaGeneration(t *testing.T) {
	tests := []struct {
		name       string
		config     ProjectConfig
		wantFiles  []string
		wantInSQL  []string
	}{
		{
			name: "PgVector with 768 dimensions",
			config: ProjectConfig{
				Name:                "test-pgvector",
				MemoryEnabled:       true,
				MemoryProvider:      "pgvector",
				EmbeddingDimensions: 768,
			},
			wantFiles: []string{
				"docker-compose.yml",
				"init-db.sql",
				"setup.sh",
				"setup.bat",
			},
			wantInSQL: []string{
				"vector(768)",
				"CREATE EXTENSION IF NOT EXISTS vector",
				"agent_memory",
			},
		},
		{
			name: "PgVector with 1536 dimensions",
			config: ProjectConfig{
				Name:                "test-pgvector-1536",
				MemoryEnabled:       true,
				MemoryProvider:      "pgvector",
				EmbeddingDimensions: 1536,
			},
			wantFiles: []string{
				"docker-compose.yml",
				"init-db.sql",
			},
			wantInSQL: []string{
				"vector(1536)",
			},
		},
		{
			name: "Weaviate configuration",
			config: ProjectConfig{
				Name:                "test-weaviate",
				MemoryEnabled:       true,
				MemoryProvider:      "weaviate",
				EmbeddingDimensions: 3072,
			},
			wantFiles: []string{
				"docker-compose.yml",
				"setup.sh",
				"setup.bat",
			},
		},
	}

	tempDir := t.TempDir()
	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(tempDir)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set required fields
			tt.config.NumAgents = 1
			tt.config.Provider = "azure"

			err := CreateAgentProjectModular(tt.config)
			if err != nil {
				t.Errorf("CreateAgentProjectModular() error = %v", err)
				return
			}

			projectPath := filepath.Join(tempDir, tt.config.Name)

			// Check that expected files were created
			for _, file := range tt.wantFiles {
				filePath := filepath.Join(projectPath, file)
				if _, err := os.Stat(filePath); os.IsNotExist(err) {
					t.Errorf("Expected file not created: %s", file)
				}
			}

			// Check SQL content if applicable
			if tt.config.MemoryProvider == "pgvector" && len(tt.wantInSQL) > 0 {
				sqlPath := filepath.Join(projectPath, "init-db.sql")
				if sqlContent, err := os.ReadFile(sqlPath); err == nil {
					sqlStr := string(sqlContent)
					for _, expected := range tt.wantInSQL {
						if !strings.Contains(sqlStr, expected) {
							t.Errorf("init-db.sql missing expected content: %s", expected)
						}
					}
				}
			}

			// Clean up
			os.RemoveAll(projectPath)
		})
	}
}