package scaffold

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestEndToEndProjectGeneration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping end-to-end tests in short mode")
	}

	tempDir := t.TempDir()
	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(tempDir)

	tests := []struct {
		name           string
		config         ProjectConfig
		shouldCompile  bool
		skipGoModTidy  bool
	}{
		{
			name: "Basic project compilation",
			config: ProjectConfig{
				Name:          "e2e-basic",
				NumAgents:     2,
				Provider:      "mock", // Use mock to avoid external dependencies
				ResponsibleAI: true,
				ErrorHandler:  true,
				MemoryEnabled: false,
			},
			shouldCompile: true,
		},
		{
			name: "Memory-enabled project compilation",
			config: ProjectConfig{
				Name:                "e2e-memory",
				NumAgents:           1,
				Provider:            "mock",
				ResponsibleAI:       false,
				ErrorHandler:        false,
				MemoryEnabled:       true,
				MemoryProvider:      "memory", // Use in-memory to avoid database dependencies
				EmbeddingProvider:   "dummy",
				EmbeddingModel:      "dummy",
				EmbeddingDimensions: 1536,
				RAGEnabled:          true,
				RAGChunkSize:        1000,
				RAGOverlap:          100,
				RAGTopK:             5,
				RAGScoreThreshold:   0.7,
			},
			shouldCompile: true,
		},
		{
			name: "MCP-enabled project compilation",
			config: ProjectConfig{
				Name:          "e2e-mcp",
				NumAgents:     1,
				Provider:      "mock",
				ResponsibleAI: false,
				ErrorHandler:  false,
				MCPEnabled:    true,
				MCPProduction: false,
				WithCache:     false,
				WithMetrics:   false,
				MCPTools:      []string{"test_tool"},
				MCPServers:    []string{"test_server"},
			},
			shouldCompile: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Generate the project
			err := CreateAgentProjectModular(tt.config)
			if err != nil {
				t.Errorf("CreateAgentProjectModular() error = %v", err)
				return
			}

			projectPath := filepath.Join(tempDir, tt.config.Name)

			// Change to project directory
			originalProjectDir, _ := os.Getwd()
			defer os.Chdir(originalProjectDir)
			os.Chdir(projectPath)

			// Run go mod tidy
			if !tt.skipGoModTidy {
				cmd := exec.Command("go", "mod", "tidy")
				cmd.Dir = projectPath
				output, err := cmd.CombinedOutput()
				if err != nil {
					t.Errorf("go mod tidy failed: %v\nOutput: %s", err, output)
					return
				}
			}

			// Try to compile the project
			cmd := exec.Command("go", "build", "-o", "test-binary", ".")
			cmd.Dir = projectPath
			output, err := cmd.CombinedOutput()

			if tt.shouldCompile {
				if err != nil {
					t.Errorf("Project compilation failed: %v\nOutput: %s", err, output)
					return
				}

				// Verify binary was created
				binaryPath := filepath.Join(projectPath, "test-binary")
				if _, err := os.Stat(binaryPath); os.IsNotExist(err) {
					t.Errorf("Compiled binary not found: %s", binaryPath)
				}

				// Clean up binary
				os.Remove(binaryPath)
			} else {
				if err == nil {
					t.Errorf("Project compilation should have failed but succeeded")
				}
			}

			// Clean up project
			os.Chdir(originalProjectDir)
			os.RemoveAll(projectPath)
		})
	}
}

func TestGeneratedProjectStructure(t *testing.T) {
	tempDir := t.TempDir()
	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(tempDir)

	config := ProjectConfig{
		Name:                "structure-test",
		NumAgents:           3,
		Provider:            "azure",
		ResponsibleAI:       true,
		ErrorHandler:        true,
		MemoryEnabled:       true,
		MemoryProvider:      "pgvector",
		EmbeddingProvider:   "ollama",
		EmbeddingModel:      "nomic-embed-text:latest",
		EmbeddingDimensions: 768,
		RAGEnabled:          true,
		MCPEnabled:          true,
		Visualize:           true,
		VisualizeOutputDir:  "docs/workflows",
	}

	err := CreateAgentProjectModular(config)
	if err != nil {
		t.Errorf("CreateAgentProjectModular() error = %v", err)
		return
	}

	projectPath := filepath.Join(tempDir, config.Name)

	// Define expected project structure
	expectedStructure := map[string]bool{
		"go.mod":                    false, // file
		"README.md":                 false, // file
		"main.go":                   false, // file
		"agentflow.toml":            false, // file
		"agent1.go":                 false, // file
		"agent2.go":                 false, // file
		"agent3.go":                 false, // file
		"docker-compose.yml":        false, // file (for pgvector)
		"init-db.sql":               false, // file (for pgvector)
		"setup.sh":                  false, // file (for pgvector)
		"setup.bat":                 false, // file (for pgvector)
		".env.example":              false, // file (for pgvector)
		"docs/workflows":            true,  // directory (for visualization)
	}

	// Check each expected item
	for item, isDir := range expectedStructure {
		itemPath := filepath.Join(projectPath, item)
		stat, err := os.Stat(itemPath)
		
		if os.IsNotExist(err) {
			t.Errorf("Expected %s not found: %s", map[bool]string{true: "directory", false: "file"}[isDir], item)
			continue
		}
		
		if err != nil {
			t.Errorf("Error checking %s: %v", item, err)
			continue
		}

		if isDir && !stat.IsDir() {
			t.Errorf("Expected directory but found file: %s", item)
		} else if !isDir && stat.IsDir() {
			t.Errorf("Expected file but found directory: %s", item)
		}
	}

	// Verify README contains expected sections
	readmePath := filepath.Join(projectPath, "README.md")
	readmeContent, err := os.ReadFile(readmePath)
	if err != nil {
		t.Errorf("Failed to read README.md: %v", err)
	} else {
		readmeStr := string(readmeContent)
		expectedSections := []string{
			"Quick Start",
			"Project Configuration",
			"Agents",
			"Memory System",
			"Troubleshooting",
			"Next Steps",
		}
		
		for _, section := range expectedSections {
			if !strings.Contains(readmeStr, section) {
				t.Errorf("README.md missing expected section: %s", section)
			}
		}
	}
}

func TestConfigurationConsistency(t *testing.T) {
	tempDir := t.TempDir()
	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(tempDir)

	tests := []struct {
		name   string
		config ProjectConfig
		checks []func(projectPath string) error
	}{
		{
			name: "Memory configuration consistency",
			config: ProjectConfig{
				Name:                "consistency-test",
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
			},
			checks: []func(string) error{
				func(projectPath string) error {
					// Check TOML has correct dimensions
					tomlPath := filepath.Join(projectPath, "agentflow.toml")
					content, err := os.ReadFile(tomlPath)
					if err != nil {
						return err
					}
					if !strings.Contains(string(content), "dimensions = 768") {
						return fmt.Errorf("TOML missing correct dimensions")
					}
					return nil
				},
				func(projectPath string) error {
					// Check main.go has validation for correct dimensions
					mainPath := filepath.Join(projectPath, "main.go")
					content, err := os.ReadFile(mainPath)
					if err != nil {
						return err
					}
					if !strings.Contains(string(content), "768") {
						return fmt.Errorf("main.go missing dimension validation")
					}
					return nil
				},
				func(projectPath string) error {
					// Check SQL has correct vector dimensions
					sqlPath := filepath.Join(projectPath, "init-db.sql")
					content, err := os.ReadFile(sqlPath)
					if err != nil {
						return err
					}
					if !strings.Contains(string(content), "vector(768)") {
						return fmt.Errorf("SQL missing correct vector dimensions")
					}
					return nil
				},
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

			projectPath := filepath.Join(tempDir, tt.config.Name)

			// Run all consistency checks
			for i, check := range tt.checks {
				if err := check(projectPath); err != nil {
					t.Errorf("Consistency check %d failed: %v", i+1, err)
				}
			}

			// Clean up
			os.RemoveAll(projectPath)
		})
	}
}

func TestMemoryValidationInGeneratedCode(t *testing.T) {
	tempDir := t.TempDir()
	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(tempDir)

	config := ProjectConfig{
		Name:                "validation-test",
		NumAgents:           1,
		Provider:            "mock",
		MemoryEnabled:       true,
		MemoryProvider:      "memory",
		EmbeddingProvider:   "dummy",
		EmbeddingModel:      "dummy",
		EmbeddingDimensions: 1536,
		RAGEnabled:          true,
	}

	err := CreateAgentProjectModular(config)
	if err != nil {
		t.Errorf("CreateAgentProjectModular() error = %v", err)
		return
	}

	projectPath := filepath.Join(tempDir, config.Name)

	// Read main.go and verify validation function exists
	mainPath := filepath.Join(projectPath, "main.go")
	mainContent, err := os.ReadFile(mainPath)
	if err != nil {
		t.Errorf("Failed to read main.go: %v", err)
		return
	}

	mainStr := string(mainContent)

	// Check for validation function
	if !strings.Contains(mainStr, "func validateMemoryConfig") {
		t.Errorf("main.go missing validateMemoryConfig function")
	}

	// Check for validation call
	if !strings.Contains(mainStr, "validateMemoryConfig(memoryConfig") {
		t.Errorf("main.go missing validation function call")
	}

	// Check for error handling
	if !strings.Contains(mainStr, "Configuration Error") {
		t.Errorf("main.go missing configuration error handling")
	}

	// Check for troubleshooting help
	if !strings.Contains(mainStr, "Troubleshooting") {
		t.Errorf("main.go missing troubleshooting guidance")
	}
}

// Benchmark tests for performance
func BenchmarkProjectGeneration(b *testing.B) {
	tempDir := b.TempDir()
	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(tempDir)

	config := ProjectConfig{
		Name:          "benchmark-test",
		NumAgents:     2,
		Provider:      "azure",
		ResponsibleAI: true,
		ErrorHandler:  true,
		MemoryEnabled: false,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		config.Name = fmt.Sprintf("benchmark-test-%d", i)
		err := CreateAgentProjectModular(config)
		if err != nil {
			b.Errorf("CreateAgentProjectModular() error = %v", err)
		}
		// Clean up immediately to avoid disk space issues
		os.RemoveAll(filepath.Join(tempDir, config.Name))
	}
}

func BenchmarkMemoryProjectGeneration(b *testing.B) {
	tempDir := b.TempDir()
	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(tempDir)

	config := ProjectConfig{
		Name:                "benchmark-memory-test",
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
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		config.Name = fmt.Sprintf("benchmark-memory-test-%d", i)
		err := CreateAgentProjectModular(config)
		if err != nil {
			b.Errorf("CreateAgentProjectModular() error = %v", err)
		}
		// Clean up immediately
		os.RemoveAll(filepath.Join(tempDir, config.Name))
	}
}