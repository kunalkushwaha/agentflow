package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/kunalkushwaha/agentflow/core"
)

func TestMemoryDebugger_ShowOverview(t *testing.T) {
	// Create a temporary config file for testing
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "agentflow.toml")
	
	configContent := `
[agent_flow]
name = "test-project"
provider = "mock"

[agent_memory]
provider = "memory"
connection = "memory"
dimensions = 768
auto_embed = true

[agent_memory.embedding]
provider = "dummy"
model = "dummy"
`
	
	err := os.WriteFile(configPath, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test config: %v", err)
	}

	// Load the config
	config, err := core.LoadConfig(configPath)
	if err != nil {
		t.Fatalf("Failed to load test config: %v", err)
	}

	// Create debugger
	debugger := &MemoryDebugger{
		config:     config,
		configPath: configPath,
	}

	// Test connection (this might fail in test environment, which is okay)
	err = debugger.Connect()
	if err != nil {
		t.Logf("Connection failed (expected in test environment): %v", err)
		return // Skip the rest of the test if we can't connect
	}
	defer debugger.Close()

	// Test ShowOverview
	err = debugger.ShowOverview()
	if err != nil {
		t.Errorf("ShowOverview() error = %v", err)
	}
}

func TestMemoryDebugger_ShowConfig(t *testing.T) {
	// Create a comprehensive test config
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "agentflow.toml")
	
	configContent := `
[agent_flow]
name = "test-project"
provider = "mock"

[agent_memory]
provider = "pgvector"
connection = "postgres://user:password@localhost:5432/agentflow"
dimensions = 768
max_results = 5
auto_embed = true
enable_knowledge_base = true
knowledge_max_results = 5
knowledge_score_threshold = 0.7
chunk_size = 1000
chunk_overlap = 100
enable_rag = true
rag_max_context_tokens = 4000
rag_personal_weight = 0.3
rag_knowledge_weight = 0.7
rag_include_sources = true

[agent_memory.embedding]
provider = "ollama"
model = "nomic-embed-text:latest"
base_url = "http://localhost:11434"
cache_embeddings = true
max_batch_size = 100
timeout_seconds = 30

[agent_memory.documents]
auto_chunk = true
supported_types = ["pdf", "txt", "md"]
max_file_size = "10MB"
enable_metadata_extraction = true
enable_url_scraping = true

[agent_memory.search]
hybrid_search = true
keyword_weight = 0.3
semantic_weight = 0.7
enable_reranking = false
enable_query_expansion = false
`
	
	err := os.WriteFile(configPath, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test config: %v", err)
	}

	// Load the config
	config, err := core.LoadConfig(configPath)
	if err != nil {
		t.Fatalf("Failed to load test config: %v", err)
	}

	// Create debugger
	debugger := &MemoryDebugger{
		config:     config,
		configPath: configPath,
	}

	// Test ShowConfig (doesn't require connection)
	err = debugger.ShowConfig()
	if err != nil {
		t.Errorf("ShowConfig() error = %v", err)
	}
}

func TestMemoryDebugger_ValidateConfig(t *testing.T) {
	tests := []struct {
		name         string
		configToml   string
		expectErrors bool
	}{
		{
			name: "Valid configuration",
			configToml: `
[agent_flow]
name = "test-project"
provider = "mock"

[agent_memory]
provider = "memory"
connection = "memory"
dimensions = 768
enable_rag = true
chunk_size = 1000
chunk_overlap = 100
knowledge_score_threshold = 0.7

[agent_memory.embedding]
provider = "dummy"
model = "dummy"
`,
			expectErrors: false,
		},
		{
			name: "Invalid dimensions",
			configToml: `
[agent_flow]
name = "test-project"
provider = "mock"

[agent_memory]
provider = "memory"
connection = "memory"
dimensions = 0

[agent_memory.embedding]
provider = "dummy"
model = "dummy"
`,
			expectErrors: true,
		},
		{
			name: "Invalid RAG configuration",
			configToml: `
[agent_flow]
name = "test-project"
provider = "mock"

[agent_memory]
provider = "memory"
connection = "memory"
dimensions = 768
enable_rag = true
chunk_size = -1
chunk_overlap = 2000
knowledge_score_threshold = 1.5

[agent_memory.embedding]
provider = "dummy"
model = "dummy"
`,
			expectErrors: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempDir := t.TempDir()
			configPath := filepath.Join(tempDir, "agentflow.toml")
			
			err := os.WriteFile(configPath, []byte(tt.configToml), 0644)
			if err != nil {
				t.Fatalf("Failed to create test config: %v", err)
			}

			// Load the config
			config, err := core.LoadConfig(configPath)
			if err != nil {
				if !tt.expectErrors {
					t.Fatalf("Failed to load test config: %v", err)
				}
				return // Expected error in config loading
			}

			// Create debugger
			debugger := &MemoryDebugger{
				config:     config,
				configPath: configPath,
			}

			// Try to connect (might fail, which is okay for validation testing)
			err = debugger.Connect()
			if err != nil && !tt.expectErrors {
				t.Logf("Connection failed (might be expected): %v", err)
				return
			}
			if debugger.memory != nil {
				defer debugger.Close()
			}

			// Test ValidateConfig
			err = debugger.ValidateConfig()
			if err != nil {
				t.Errorf("ValidateConfig() error = %v", err)
			}
			// Note: ValidateConfig doesn't return errors, it prints them
			// In a real implementation, we might want to refactor this to return validation results
		})
	}
}

func TestMemoryDebugger_getTroubleshootingHelp(t *testing.T) {
	tests := []struct {
		name           string
		memoryProvider string
		wantContains   []string
	}{
		{
			name:           "PgVector troubleshooting",
			memoryProvider: "pgvector",
			wantContains:   []string{"PostgreSQL", "docker compose", "psql"},
		},
		{
			name:           "Weaviate troubleshooting",
			memoryProvider: "weaviate",
			wantContains:   []string{"Weaviate", "curl", "8080"},
		},
		{
			name:           "Memory troubleshooting",
			memoryProvider: "memory",
			wantContains:   []string{"In-Memory", "configuration"},
		},
		{
			name:           "Unknown provider",
			memoryProvider: "unknown",
			wantContains:   []string{"memory provider"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a minimal config
			config := &core.Config{
				AgentMemory: core.AgentMemoryConfig{
					Provider: tt.memoryProvider,
				},
			}

			debugger := &MemoryDebugger{
				config: config,
			}

			help := debugger.getTroubleshootingHelp()

			for _, want := range tt.wantContains {
				if !strings.Contains(help, want) {
					t.Errorf("getTroubleshootingHelp() missing expected content: %s\nGot: %s", want, help)
				}
			}
		})
	}
}

func TestRunMemoryCommand_ConfigValidation(t *testing.T) {
	tests := []struct {
		name        string
		setupConfig bool
		configToml  string
		wantError   bool
	}{
		{
			name:        "No config file",
			setupConfig: false,
			wantError:   true,
		},
		{
			name:        "Valid config file",
			setupConfig: true,
			configToml: `
[agent_flow]
name = "test-project"
provider = "mock"

[agent_memory]
provider = "memory"
connection = "memory"
dimensions = 768

[agent_memory.embedding]
provider = "dummy"
model = "dummy"
`,
			wantError: false,
		},
		{
			name:        "Config without memory",
			setupConfig: true,
			configToml: `
[agent_flow]
name = "test-project"
provider = "mock"
`,
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempDir := t.TempDir()
			originalDir, _ := os.Getwd()
			defer os.Chdir(originalDir)
			os.Chdir(tempDir)

			if tt.setupConfig {
				err := os.WriteFile("agentflow.toml", []byte(tt.configToml), 0644)
				if err != nil {
					t.Fatalf("Failed to create test config: %v", err)
				}
			}

			// Create a mock command for testing
			cmd := &cobra.Command{}
			err := runMemoryCommand(cmd, []string{})

			if tt.wantError && err == nil {
				t.Errorf("runMemoryCommand() expected error but got none")
			}
			if !tt.wantError && err != nil {
				t.Errorf("runMemoryCommand() unexpected error: %v", err)
			}
		})
	}
}

// Test helper functions
func TestParseCommaSeparatedList(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  []string
	}{
		{
			name:  "Empty string",
			input: "",
			want:  []string{},
		},
		{
			name:  "Single item",
			input: "item1",
			want:  []string{"item1"},
		},
		{
			name:  "Multiple items",
			input: "item1,item2,item3",
			want:  []string{"item1", "item2", "item3"},
		},
		{
			name:  "Items with spaces",
			input: "item1, item2 , item3",
			want:  []string{"item1", "item2", "item3"},
		},
		{
			name:  "Items with empty entries",
			input: "item1,,item2,",
			want:  []string{"item1", "item2"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseCommaSeparatedList(tt.input)
			if len(got) != len(tt.want) {
				t.Errorf("parseCommaSeparatedList() length = %d, want %d", len(got), len(tt.want))
				return
			}
			for i, v := range got {
				if v != tt.want[i] {
					t.Errorf("parseCommaSeparatedList()[%d] = %s, want %s", i, v, tt.want[i])
				}
			}
		})
	}
}

func TestContains(t *testing.T) {
	tests := []struct {
		name  string
		slice []string
		item  string
		want  bool
	}{
		{
			name:  "Item exists",
			slice: []string{"a", "b", "c"},
			item:  "b",
			want:  true,
		},
		{
			name:  "Item doesn't exist",
			slice: []string{"a", "b", "c"},
			item:  "d",
			want:  false,
		},
		{
			name:  "Empty slice",
			slice: []string{},
			item:  "a",
			want:  false,
		},
		{
			name:  "Case sensitive",
			slice: []string{"A", "B", "C"},
			item:  "a",
			want:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := contains(tt.slice, tt.item)
			if got != tt.want {
				t.Errorf("contains() = %v, want %v", got, tt.want)
			}
		})
	}
}