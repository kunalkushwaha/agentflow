package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/kunalkushwaha/agentflow/core"
	"github.com/spf13/cobra"
)

// Memory debug command flags
var (
	memoryStats     bool
	memoryList      bool
	memorySessions  bool
	memoryDocs      bool
	memorySearch    string
	memoryValidate  bool
	memoryClear     bool
	memoryConfig    bool
	memoryConfigPath string
)

// memoryCmd represents the memory debug command
var memoryCmd = &cobra.Command{
	Use:   "memory",
	Short: "Debug and inspect memory system state",
	Long: `Debug and inspect the current state of your AgentFlow memory system.

This command provides various tools to inspect, validate, and manage your memory system:

BASIC USAGE:
  # Show memory system overview and basic stats
  agentcli memory

  # Show detailed statistics
  agentcli memory --stats

  # List recent memories
  agentcli memory --list

  # Show active sessions
  agentcli memory --sessions

  # List knowledge base documents
  agentcli memory --docs

  # Test search functionality
  agentcli memory --search "your search query"

  # Validate configuration
  agentcli memory --validate

  # Show current configuration
  agentcli memory --config

  # Clear memory data (with confirmation)
  agentcli memory --clear

CONFIGURATION:
  # Use specific config file
  agentcli memory --config-path /path/to/agentflow.toml

The command automatically reads your agentflow.toml configuration and connects
to the configured memory provider to provide real-time information about your
memory system state.

REQUIREMENTS:
- Must be run from an AgentFlow project directory (containing agentflow.toml)
- Memory system must be configured in agentflow.toml
- Database/memory provider must be accessible`,
	RunE: runMemoryCommand,
}

func init() {
	rootCmd.AddCommand(memoryCmd)

	// Memory debug flags
	memoryCmd.Flags().BoolVar(&memoryStats, "stats", false, "Show detailed memory statistics")
	memoryCmd.Flags().BoolVar(&memoryList, "list", false, "List recent memories with previews")
	memoryCmd.Flags().BoolVar(&memorySessions, "sessions", false, "Show active sessions and their data")
	memoryCmd.Flags().BoolVar(&memoryDocs, "docs", false, "List knowledge base documents")
	memoryCmd.Flags().StringVar(&memorySearch, "search", "", "Test search functionality with query")
	memoryCmd.Flags().BoolVar(&memoryValidate, "validate", false, "Validate memory configuration")
	memoryCmd.Flags().BoolVar(&memoryClear, "clear", false, "Clear memory data (with confirmation)")
	memoryCmd.Flags().BoolVar(&memoryConfig, "config", false, "Show current memory configuration")
	memoryCmd.Flags().StringVar(&memoryConfigPath, "config-path", "", "Path to agentflow.toml file (default: ./agentflow.toml)")
}

func runMemoryCommand(cmd *cobra.Command, args []string) error {
	// Determine config file path
	configPath := "agentflow.toml"
	if memoryConfigPath != "" {
		configPath = memoryConfigPath
	}

	// Check if config file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return fmt.Errorf("❌ No agentflow.toml found in current directory\n💡 Run this command from your AgentFlow project root, or specify config:\n   agentcli memory --config-path /path/to/agentflow.toml")
	}

	// Load configuration
	config, err := core.LoadConfig(configPath)
	if err != nil {
		return fmt.Errorf("❌ Failed to load configuration: %v\n💡 Check your agentflow.toml file for syntax errors", err)
	}

	// Check if memory is configured
	if config.AgentMemory.Provider == "" {
		return fmt.Errorf("❌ Memory system not configured in agentflow.toml\n💡 Add [agent_memory] section to enable memory features")
	}

	// Create memory debugger
	debugger := &MemoryDebugger{
		config:     config,
		configPath: configPath,
	}

	// Initialize memory connection
	if err := debugger.Connect(); err != nil {
		return fmt.Errorf("❌ Failed to connect to memory system: %v\n%s", err, debugger.getTroubleshootingHelp())
	}
	defer debugger.Close()

	// Handle specific commands
	if memoryStats {
		return debugger.ShowStats()
	}
	if memoryList {
		return debugger.ListMemories()
	}
	if memorySessions {
		return debugger.ShowSessions()
	}
	if memoryDocs {
		return debugger.ListDocuments()
	}
	if memorySearch != "" {
		return debugger.TestSearch(memorySearch)
	}
	if memoryValidate {
		return debugger.ValidateConfig()
	}
	if memoryClear {
		return debugger.ClearData()
	}
	if memoryConfig {
		return debugger.ShowConfig()
	}

	// Default: show overview
	return debugger.ShowOverview()
}

// MemoryDebugger provides debugging functionality for memory systems
type MemoryDebugger struct {
	config     *core.Config
	configPath string
	memory     core.Memory
}

// Connect establishes connection to the memory system
func (m *MemoryDebugger) Connect() error {
	memory, err := core.NewMemory(m.config.AgentMemory)
	if err != nil {
		return err
	}
	m.memory = memory
	return nil
}

// Close closes the memory connection
func (m *MemoryDebugger) Close() {
	if m.memory != nil {
		m.memory.Close()
	}
}

// ShowOverview displays basic memory system information
func (m *MemoryDebugger) ShowOverview() error {
	fmt.Printf("🧠 AgentFlow Memory System Debug\n")
	fmt.Printf("================================\n\n")

	// Basic configuration info
	fmt.Printf("📁 Config File: %s\n", m.configPath)
	fmt.Printf("🔧 Provider: %s\n", m.config.AgentMemory.Provider)
	fmt.Printf("🤖 Embedding: %s/%s (%d dimensions)\n", 
		m.config.AgentMemory.Embedding.Provider,
		m.config.AgentMemory.Embedding.Model,
		m.config.AgentMemory.Dimensions)

	// Connection status
	fmt.Printf("🔗 Connection: ")
	ctx := context.Background()
	testContent := fmt.Sprintf("Debug test at %s", time.Now().Format("15:04:05"))
	if err := m.memory.Store(ctx, testContent, "debug-test"); err != nil {
		fmt.Printf("❌ Failed (%v)\n", err)
		return nil
	}
	fmt.Printf("✅ Connected\n\n")

	// Quick stats
	fmt.Printf("📊 Quick Stats:\n")
	fmt.Printf("   Use --stats for detailed statistics\n")
	fmt.Printf("   Use --list to see recent memories\n")
	fmt.Printf("   Use --validate to check configuration\n")
	fmt.Printf("   Use --help for all available options\n")

	return nil
}

// getTroubleshootingHelp returns provider-specific troubleshooting help
func (m *MemoryDebugger) getTroubleshootingHelp() string {
	switch m.config.AgentMemory.Provider {
	case "pgvector":
		return `💡 PostgreSQL/PgVector Troubleshooting:
   1. Start database: docker compose up -d
   2. Check connection: psql -h localhost -U user -d agentflow
   3. Verify connection string in agentflow.toml
   4. Run setup script: ./setup.sh (or setup.bat on Windows)`
	case "weaviate":
		return `💡 Weaviate Troubleshooting:
   1. Start Weaviate: docker compose up -d
   2. Check status: curl http://localhost:8080/v1/meta
   3. Verify connection string in agentflow.toml`
	case "memory":
		return `💡 In-Memory Provider Issue:
   This shouldn't fail - check your configuration syntax`
	default:
		return `💡 Check your memory provider configuration in agentflow.toml`
	}
}

// ShowStats displays detailed memory statistics
func (m *MemoryDebugger) ShowStats() error {
	fmt.Printf("📊 Memory Statistics\n")
	fmt.Printf("===================\n\n")

	ctx := context.Background()

	// Basic configuration stats
	fmt.Printf("🔧 Configuration:\n")
	fmt.Printf("   Provider: %s\n", m.config.AgentMemory.Provider)
	fmt.Printf("   Connection: %s\n", m.config.AgentMemory.Connection)
	fmt.Printf("   Dimensions: %d\n", m.config.AgentMemory.Dimensions)
	fmt.Printf("   Embedding: %s/%s\n", m.config.AgentMemory.Embedding.Provider, m.config.AgentMemory.Embedding.Model)
	fmt.Printf("   RAG Enabled: %t\n", m.config.AgentMemory.EnableRAG)
	if m.config.AgentMemory.EnableRAG {
		fmt.Printf("   Chunk Size: %d\n", m.config.AgentMemory.ChunkSize)
		fmt.Printf("   Chunk Overlap: %d\n", m.config.AgentMemory.ChunkOverlap)
		fmt.Printf("   Score Threshold: %.2f\n", m.config.AgentMemory.KnowledgeScoreThreshold)
	}
	fmt.Printf("\n")

	// Try to get memory statistics
	fmt.Printf("📈 Memory Usage:\n")
	
	// Test basic memory operations
	testQuery := "test query for statistics"
	results, err := m.memory.Query(ctx, testQuery, 5)
	if err != nil {
		fmt.Printf("   ❌ Query test failed: %v\n", err)
	} else {
		fmt.Printf("   ✅ Query successful (%d results)\n", len(results))
	}

	// Test memory storage
	testContent := fmt.Sprintf("Statistics test at %s", time.Now().Format("2006-01-02 15:04:05"))
	if err := m.memory.Store(ctx, testContent, "stats-test"); err != nil {
		fmt.Printf("   ❌ Storage test failed: %v\n", err)
	} else {
		fmt.Printf("   ✅ Storage test successful\n")
	}

	// Test history retrieval
	history, err := m.memory.GetHistory(ctx, 3)
	if err != nil {
		fmt.Printf("   ❌ History retrieval failed: %v\n", err)
	} else {
		fmt.Printf("   ✅ History available (%d messages)\n", len(history))
	}

	fmt.Printf("\n")
	fmt.Printf("💡 Note: Detailed statistics depend on memory provider capabilities\n")
	fmt.Printf("   Use --list to see actual memory content\n")

	return nil
}

// ListMemories lists recent memories with content previews
func (m *MemoryDebugger) ListMemories() error {
	fmt.Printf("📝 Recent Memories\n")
	fmt.Printf("==================\n\n")

	ctx := context.Background()

	// Get chat history instead of using empty query which causes embedding issues
	messages, err := m.memory.GetHistory(ctx, 10)
	if err != nil {
		fmt.Printf("❌ Failed to retrieve memories: %v\n", err)
		return nil
	}

	if len(messages) == 0 {
		fmt.Printf("📭 No memories found\n")
		fmt.Printf("💡 Memories are created when agents store information or when you interact with the system\n")
		return nil
	}

	fmt.Printf("Found %d memories:\n\n", len(messages))

	for i, message := range messages {
		// Truncate content for preview
		content := message.Content
		if len(content) > 100 {
			content = content[:97] + "..."
		}

		fmt.Printf("%d. Role: %s\n", i+1, message.Role)
		fmt.Printf("   Content: %s\n", content)
		fmt.Printf("   Time: %s\n", message.CreatedAt.Format("2006-01-02 15:04:05"))
		fmt.Printf("\n")
	}

	fmt.Printf("💡 Use --search \"query\" to test semantic search functionality\n")
	return nil
}

// ShowSessions shows active sessions and their data
func (m *MemoryDebugger) ShowSessions() error {
	fmt.Printf("👥 Active Sessions\n")
	fmt.Printf("==================\n\n")

	ctx := context.Background()

	// Try to create a test session to demonstrate functionality
	sessionID := m.memory.NewSession()
	if sessionID == "" {
		fmt.Printf("❌ Session management not supported by this memory provider\n")
		fmt.Printf("💡 Session support depends on the memory provider implementation\n")
		return nil
	}

	fmt.Printf("✅ Session management is supported\n")
	fmt.Printf("🆔 Test session created: %s\n\n", sessionID)

	// Set session context and test
	sessionCtx := m.memory.SetSession(ctx, sessionID)
	
	// Store some test data in the session
	testContent := fmt.Sprintf("Session test data at %s", time.Now().Format("15:04:05"))
	if err := m.memory.Store(sessionCtx, testContent, "session-test"); err != nil {
		fmt.Printf("❌ Failed to store session data: %v\n", err)
	} else {
		fmt.Printf("✅ Session data storage successful\n")
	}

	// Try to retrieve session-specific data
	results, err := m.memory.Query(sessionCtx, "session test", 5)
	if err != nil {
		fmt.Printf("❌ Failed to query session data: %v\n", err)
	} else {
		fmt.Printf("✅ Session query successful (%d results)\n", len(results))
		if len(results) > 0 {
			fmt.Printf("   Latest session content: %s\n", results[0].Content)
		}
	}

	fmt.Printf("\n💡 Session Features:\n")
	fmt.Printf("   - Sessions isolate memory data between different conversations\n")
	fmt.Printf("   - Each session maintains its own context and history\n")
	fmt.Printf("   - Useful for multi-user or multi-conversation scenarios\n")

	return nil
}

// ListDocuments lists knowledge base documents
func (m *MemoryDebugger) ListDocuments() error {
	fmt.Printf("📚 Knowledge Base Documents\n")
	fmt.Printf("===========================\n\n")

	ctx := context.Background()

	// Check if RAG/knowledge base is enabled
	if !m.config.AgentMemory.EnableKnowledgeBase {
		fmt.Printf("❌ Knowledge base not enabled in configuration\n")
		fmt.Printf("💡 Enable knowledge base in agentflow.toml:\n")
		fmt.Printf("   [agent_memory]\n")
		fmt.Printf("   enable_knowledge_base = true\n")
		return nil
	}

	fmt.Printf("✅ Knowledge base is enabled\n")
	fmt.Printf("🔧 Configuration:\n")
	fmt.Printf("   Max Results: %d\n", m.config.AgentMemory.KnowledgeMaxResults)
	fmt.Printf("   Score Threshold: %.2f\n", m.config.AgentMemory.KnowledgeScoreThreshold)
	if m.config.AgentMemory.EnableRAG {
		fmt.Printf("   RAG Enabled: Yes\n")
		fmt.Printf("   Chunk Size: %d\n", m.config.AgentMemory.ChunkSize)
		fmt.Printf("   Chunk Overlap: %d\n", m.config.AgentMemory.ChunkOverlap)
	}
	fmt.Printf("\n")

	// Try to query for document-like content
	// This is a basic implementation - actual document listing would depend on the memory provider
	fmt.Printf("📄 Searching for document-like content...\n")
	
	// Search for various document indicators
	documentQueries := []string{"document", "file", "pdf", "text", "content", "knowledge"}
	totalDocuments := 0
	
	for _, query := range documentQueries {
		results, err := m.memory.Query(ctx, query, 5)
		if err != nil {
			continue
		}
		
		for _, result := range results {
			if result.Score > m.config.AgentMemory.KnowledgeScoreThreshold {
				totalDocuments++
				
				// Truncate content for display
				content := result.Content
				if len(content) > 150 {
					content = content[:147] + "..."
				}
				
				fmt.Printf("📄 Document %d (Score: %.3f):\n", totalDocuments, result.Score)
				fmt.Printf("   Content: %s\n", content)
				fmt.Printf("\n")
			}
		}
		
		if totalDocuments >= 10 { // Limit display
			break
		}
	}

	if totalDocuments == 0 {
		fmt.Printf("📭 No documents found in knowledge base\n")
		fmt.Printf("💡 Documents are typically added through:\n")
		fmt.Printf("   - Document ingestion APIs\n")
		fmt.Printf("   - Agent interactions that store knowledge\n")
		fmt.Printf("   - Manual content addition\n")
	} else {
		fmt.Printf("📊 Found %d document-like entries\n", totalDocuments)
		fmt.Printf("💡 Use --search \"query\" to find specific documents\n")
	}

	return nil
}

// TestSearch tests search functionality with similarity scores
func (m *MemoryDebugger) TestSearch(query string) error {
	fmt.Printf("🔍 Search Test: \"%s\"\n", query)
	fmt.Printf("========================\n\n")

	ctx := context.Background()
	startTime := time.Now()

	// Perform the search
	results, err := m.memory.Query(ctx, query, m.config.AgentMemory.MaxResults)
	if err != nil {
		fmt.Printf("❌ Search failed: %v\n", err)
		return nil
	}

	searchDuration := time.Since(startTime)
	fmt.Printf("⏱️  Search completed in %v\n", searchDuration)
	fmt.Printf("📊 Found %d results\n\n", len(results))

	if len(results) == 0 {
		fmt.Printf("📭 No results found for query: \"%s\"\n", query)
		fmt.Printf("💡 Try:\n")
		fmt.Printf("   - Using different keywords\n")
		fmt.Printf("   - Checking if data exists with --list\n")
		fmt.Printf("   - Lowering the score threshold in configuration\n")
		return nil
	}

	// Display results with similarity scores
	fmt.Printf("🎯 Search Results (sorted by relevance):\n")
	fmt.Printf("========================================\n\n")

	for i, result := range results {
		// Color-code scores
		var scoreIcon string
		switch {
		case result.Score >= 0.8:
			scoreIcon = "🟢" // High relevance
		case result.Score >= 0.6:
			scoreIcon = "🟡" // Medium relevance
		case result.Score >= 0.4:
			scoreIcon = "🟠" // Low relevance
		default:
			scoreIcon = "🔴" // Very low relevance
		}

		fmt.Printf("%s Result %d (Score: %.4f)\n", scoreIcon, i+1, result.Score)
		
		// Truncate content but show more than in list view
		content := result.Content
		if len(content) > 200 {
			content = content[:197] + "..."
		}
		
		fmt.Printf("   Content: %s\n", content)
		
		// Show relevance assessment
		if result.Score >= m.config.AgentMemory.KnowledgeScoreThreshold {
			fmt.Printf("   ✅ Above threshold (%.2f) - would be used in RAG\n", m.config.AgentMemory.KnowledgeScoreThreshold)
		} else {
			fmt.Printf("   ❌ Below threshold (%.2f) - would be filtered out\n", m.config.AgentMemory.KnowledgeScoreThreshold)
		}
		
		fmt.Printf("\n")
	}

	// Search analysis
	fmt.Printf("📈 Search Analysis:\n")
	fmt.Printf("==================\n")
	
	highRelevance := 0
	mediumRelevance := 0
	lowRelevance := 0
	aboveThreshold := 0
	
	for _, result := range results {
		if result.Score >= 0.8 {
			highRelevance++
		} else if result.Score >= 0.6 {
			mediumRelevance++
		} else {
			lowRelevance++
		}
		
		if result.Score >= m.config.AgentMemory.KnowledgeScoreThreshold {
			aboveThreshold++
		}
	}
	
	fmt.Printf("🟢 High relevance (≥0.8): %d\n", highRelevance)
	fmt.Printf("🟡 Medium relevance (≥0.6): %d\n", mediumRelevance)
	fmt.Printf("🟠 Low relevance (<0.6): %d\n", lowRelevance)
	fmt.Printf("✅ Above threshold (≥%.2f): %d\n", m.config.AgentMemory.KnowledgeScoreThreshold, aboveThreshold)
	
	if aboveThreshold == 0 {
		fmt.Printf("\n⚠️  No results above threshold - consider:\n")
		fmt.Printf("   - Lowering knowledge_score_threshold in agentflow.toml\n")
		fmt.Printf("   - Adding more relevant content to memory\n")
		fmt.Printf("   - Using different search terms\n")
	}

	return nil
}

// ValidateConfig validates memory configuration with specific error reporting
func (m *MemoryDebugger) ValidateConfig() error {
	fmt.Printf("✅ Configuration Validation\n")
	fmt.Printf("===========================\n\n")

	ctx := context.Background()
	validationErrors := []string{}
	validationWarnings := []string{}

	// Basic configuration validation
	fmt.Printf("🔧 Basic Configuration:\n")
	
	if m.config.AgentMemory.Provider == "" {
		validationErrors = append(validationErrors, "Memory provider not specified")
	} else {
		fmt.Printf("   ✅ Provider: %s\n", m.config.AgentMemory.Provider)
	}

	if m.config.AgentMemory.Dimensions <= 0 {
		validationErrors = append(validationErrors, "Invalid dimensions configuration")
	} else {
		fmt.Printf("   ✅ Dimensions: %d\n", m.config.AgentMemory.Dimensions)
	}

	if m.config.AgentMemory.Embedding.Provider == "" {
		validationErrors = append(validationErrors, "Embedding provider not specified")
	} else {
		fmt.Printf("   ✅ Embedding Provider: %s\n", m.config.AgentMemory.Embedding.Provider)
	}

	if m.config.AgentMemory.Embedding.Model == "" {
		validationErrors = append(validationErrors, "Embedding model not specified")
	} else {
		fmt.Printf("   ✅ Embedding Model: %s\n", m.config.AgentMemory.Embedding.Model)
	}

	// Connection validation
	fmt.Printf("\n🔗 Connection Validation:\n")
	
	// Test basic connection
	testContent := fmt.Sprintf("Validation test at %s", time.Now().Format("15:04:05"))
	if err := m.memory.Store(ctx, testContent, "validation-test"); err != nil {
		validationErrors = append(validationErrors, fmt.Sprintf("Connection test failed: %v", err))
		fmt.Printf("   ❌ Connection: Failed (%v)\n", err)
	} else {
		fmt.Printf("   ✅ Connection: Successful\n")
	}

	// Test query functionality
	if results, err := m.memory.Query(ctx, "validation test", 1); err != nil {
		validationErrors = append(validationErrors, fmt.Sprintf("Query test failed: %v", err))
		fmt.Printf("   ❌ Query: Failed (%v)\n", err)
	} else {
		fmt.Printf("   ✅ Query: Successful (%d results)\n", len(results))
	}

	// Provider-specific validation
	fmt.Printf("\n🏗️  Provider-Specific Validation:\n")
	
	switch m.config.AgentMemory.Provider {
	case "pgvector":
		if m.config.AgentMemory.Connection == "" {
			validationErrors = append(validationErrors, "PgVector requires connection string")
		} else if !strings.Contains(m.config.AgentMemory.Connection, "postgres://") {
			validationWarnings = append(validationWarnings, "PgVector connection string should start with 'postgres://'")
		} else {
			fmt.Printf("   ✅ PgVector connection string format valid\n")
		}
		
		if m.config.AgentMemory.Dimensions > 2000 {
			validationWarnings = append(validationWarnings, "Large dimensions may impact PgVector performance")
		}

	case "weaviate":
		if m.config.AgentMemory.Connection == "" {
			validationErrors = append(validationErrors, "Weaviate requires connection URL")
		} else if !strings.Contains(m.config.AgentMemory.Connection, "http") {
			validationWarnings = append(validationWarnings, "Weaviate connection should be HTTP URL")
		} else {
			fmt.Printf("   ✅ Weaviate connection URL format valid\n")
		}

	case "memory":
		fmt.Printf("   ✅ In-memory provider requires no additional configuration\n")
		if m.config.AgentMemory.EnableRAG {
			validationWarnings = append(validationWarnings, "RAG with in-memory provider - data won't persist")
		}

	default:
		validationErrors = append(validationErrors, fmt.Sprintf("Unknown memory provider: %s", m.config.AgentMemory.Provider))
	}

	// Summary
	fmt.Printf("\n📋 Validation Summary:\n")
	fmt.Printf("======================\n")

	if len(validationErrors) == 0 && len(validationWarnings) == 0 {
		fmt.Printf("🎉 All validations passed! Your memory configuration looks good.\n")
	} else {
		if len(validationErrors) > 0 {
			fmt.Printf("❌ Errors found (%d):\n", len(validationErrors))
			for i, err := range validationErrors {
				fmt.Printf("   %d. %s\n", i+1, err)
			}
		}

		if len(validationWarnings) > 0 {
			fmt.Printf("\n⚠️  Warnings (%d):\n", len(validationWarnings))
			for i, warning := range validationWarnings {
				fmt.Printf("   %d. %s\n", i+1, warning)
			}
		}

		if len(validationErrors) > 0 {
			fmt.Printf("\n💡 Fix the errors above to ensure proper memory system operation.\n")
		}
	}

	return nil
}

// ClearData clears memory data with confirmation prompts and selective clearing
func (m *MemoryDebugger) ClearData() error {
	fmt.Printf("🗑️  Clear Memory Data\n")
	fmt.Printf("====================\n\n")

	ctx := context.Background()

	// Show current data overview
	fmt.Printf("📊 Current Memory Overview:\n")
	
	// Get current memory count
	results, err := m.memory.Query(ctx, "test", 100) // Get up to 100 items for counting
	if err != nil {
		fmt.Printf("❌ Failed to query current data: %v\n", err)
		return nil
	}

	fmt.Printf("   📝 Total memories found: %d\n", len(results))

	// Get history count
	history, err := m.memory.GetHistory(ctx, 100)
	if err == nil {
		fmt.Printf("   💬 Chat history messages: %d\n", len(history))
	}

	// Check for sessions
	sessionID := m.memory.NewSession()
	if sessionID != "" {
		fmt.Printf("   👥 Session support: Available\n")
	}

	fmt.Printf("\n⚠️  WARNING: This operation will permanently delete data!\n")
	fmt.Printf("🔒 Available clearing options:\n")
	fmt.Printf("   1. Clear all memories (stored content)\n")
	fmt.Printf("   2. Clear chat history\n")
	fmt.Printf("   3. Clear everything (memories + history)\n")
	fmt.Printf("   4. Cancel operation\n")

	fmt.Printf("\nSelect option (1-4): ")
	var choice string
	fmt.Scanln(&choice)

	switch choice {
	case "1":
		return m.clearMemories(ctx)
	case "2":
		return m.clearHistory(ctx)
	case "3":
		return m.clearEverything(ctx)
	case "4":
		fmt.Printf("✅ Operation cancelled\n")
		return nil
	default:
		fmt.Printf("❌ Invalid choice. Operation cancelled.\n")
		return nil
	}
}

// clearMemories clears stored memories
func (m *MemoryDebugger) clearMemories(ctx context.Context) error {
	fmt.Printf("\n🗑️  Clearing Memories\n")
	fmt.Printf("====================\n")

	fmt.Printf("⚠️  This will delete all stored memories (content, embeddings, metadata)\n")
	fmt.Printf("💬 Chat history will be preserved\n")
	fmt.Printf("\nType 'DELETE' to confirm: ")
	
	var confirmation string
	fmt.Scanln(&confirmation)
	
	if confirmation != "DELETE" {
		fmt.Printf("❌ Confirmation failed. Operation cancelled.\n")
		return nil
	}

	// Note: The core Memory interface doesn't have a Clear method
	// This is a limitation we need to work around
	fmt.Printf("⚠️  Direct memory clearing not supported by current Memory interface\n")
	fmt.Printf("💡 To clear memories, you can:\n")
	fmt.Printf("   1. Restart with a fresh database (for pgvector/weaviate)\n")
	fmt.Printf("   2. Restart the application (for in-memory provider)\n")
	fmt.Printf("   3. Manually clear the database tables\n")
	
	if m.config.AgentMemory.Provider == "pgvector" {
		fmt.Printf("\n🐘 PostgreSQL/PgVector clearing commands:\n")
		fmt.Printf("   psql -h localhost -U user -d agentflow -c \"TRUNCATE TABLE agent_memory;\"\n")
		fmt.Printf("   psql -h localhost -U user -d agentflow -c \"TRUNCATE TABLE documents;\"\n")
	} else if m.config.AgentMemory.Provider == "weaviate" {
		fmt.Printf("\n🔍 Weaviate clearing:\n")
		fmt.Printf("   Use Weaviate API or restart the Weaviate container\n")
	}

	return nil
}

// clearHistory clears chat history
func (m *MemoryDebugger) clearHistory(ctx context.Context) error {
	fmt.Printf("\n💬 Clearing Chat History\n")
	fmt.Printf("========================\n")

	fmt.Printf("⚠️  This will delete all chat history\n")
	fmt.Printf("📝 Stored memories will be preserved\n")
	fmt.Printf("\nType 'DELETE' to confirm: ")
	
	var confirmation string
	fmt.Scanln(&confirmation)
	
	if confirmation != "DELETE" {
		fmt.Printf("❌ Confirmation failed. Operation cancelled.\n")
		return nil
	}

	// Note: Similar limitation - no ClearHistory method in Memory interface
	fmt.Printf("⚠️  Direct history clearing not supported by current Memory interface\n")
	fmt.Printf("💡 Chat history is typically stored in the same tables as memories\n")
	fmt.Printf("   Consider using the database-specific clearing commands above\n")

	return nil
}

// clearEverything clears all data
func (m *MemoryDebugger) clearEverything(ctx context.Context) error {
	fmt.Printf("\n🗑️  Clearing Everything\n")
	fmt.Printf("======================\n")

	fmt.Printf("⚠️  This will delete ALL memory data:\n")
	fmt.Printf("   - All stored memories\n")
	fmt.Printf("   - All chat history\n")
	fmt.Printf("   - All embeddings\n")
	fmt.Printf("   - All metadata\n")
	fmt.Printf("\nType 'DELETE EVERYTHING' to confirm: ")
	
	var confirmation string
	fmt.Scanln(&confirmation)
	
	if confirmation != "DELETE EVERYTHING" {
		fmt.Printf("❌ Confirmation failed. Operation cancelled.\n")
		return nil
	}

	fmt.Printf("⚠️  Complete data clearing not supported by current Memory interface\n")
	fmt.Printf("💡 To clear all data:\n")
	
	if m.config.AgentMemory.Provider == "pgvector" {
		fmt.Printf("\n🐘 PostgreSQL/PgVector - Complete reset:\n")
		fmt.Printf("   docker compose down\n")
		fmt.Printf("   docker volume rm $(docker volume ls -q | grep postgres)\n")
		fmt.Printf("   docker compose up -d\n")
		fmt.Printf("   ./setup.sh\n")
	} else if m.config.AgentMemory.Provider == "weaviate" {
		fmt.Printf("\n🔍 Weaviate - Complete reset:\n")
		fmt.Printf("   docker compose down\n")
		fmt.Printf("   docker volume rm $(docker volume ls -q | grep weaviate)\n")
		fmt.Printf("   docker compose up -d\n")
	} else {
		fmt.Printf("\n🧠 In-memory provider:\n")
		fmt.Printf("   Restart your application to clear all data\n")
	}

	return nil
}

// ShowConfig displays current memory configuration in readable format
func (m *MemoryDebugger) ShowConfig() error {
	fmt.Printf("⚙️  Memory Configuration\n")
	fmt.Printf("========================\n\n")

	fmt.Printf("📁 Configuration File: %s\n\n", m.configPath)

	// Basic memory configuration
	fmt.Printf("🧠 Memory Provider Configuration:\n")
	fmt.Printf("   Provider: %s\n", m.config.AgentMemory.Provider)
	fmt.Printf("   Connection: %s\n", m.config.AgentMemory.Connection)
	fmt.Printf("   Max Results: %d\n", m.config.AgentMemory.MaxResults)
	fmt.Printf("   Dimensions: %d\n", m.config.AgentMemory.Dimensions)
	fmt.Printf("   Auto Embed: %t\n", m.config.AgentMemory.AutoEmbed)
	fmt.Printf("\n")

	// Embedding configuration
	fmt.Printf("🤖 Embedding Configuration:\n")
	fmt.Printf("   Provider: %s\n", m.config.AgentMemory.Embedding.Provider)
	fmt.Printf("   Model: %s\n", m.config.AgentMemory.Embedding.Model)
	if m.config.AgentMemory.Embedding.BaseURL != "" {
		fmt.Printf("   Base URL: %s\n", m.config.AgentMemory.Embedding.BaseURL)
	}
	if m.config.AgentMemory.Embedding.APIKey != "" {
		fmt.Printf("   API Key: %s\n", maskAPIKey(m.config.AgentMemory.Embedding.APIKey))
	}
	fmt.Printf("   Cache Embeddings: %t\n", m.config.AgentMemory.Embedding.CacheEmbeddings)
	fmt.Printf("   Max Batch Size: %d\n", m.config.AgentMemory.Embedding.MaxBatchSize)
	fmt.Printf("   Timeout: %ds\n", m.config.AgentMemory.Embedding.TimeoutSeconds)
	fmt.Printf("\n")

	// Knowledge base configuration
	fmt.Printf("📚 Knowledge Base Configuration:\n")
	fmt.Printf("   Enabled: %t\n", m.config.AgentMemory.EnableKnowledgeBase)
	if m.config.AgentMemory.EnableKnowledgeBase {
		fmt.Printf("   Max Results: %d\n", m.config.AgentMemory.KnowledgeMaxResults)
		fmt.Printf("   Score Threshold: %.2f\n", m.config.AgentMemory.KnowledgeScoreThreshold)
	}
	fmt.Printf("\n")

	// RAG configuration
	fmt.Printf("🔍 RAG Configuration:\n")
	fmt.Printf("   Enabled: %t\n", m.config.AgentMemory.EnableRAG)
	if m.config.AgentMemory.EnableRAG {
		fmt.Printf("   Chunk Size: %d\n", m.config.AgentMemory.ChunkSize)
		fmt.Printf("   Chunk Overlap: %d\n", m.config.AgentMemory.ChunkOverlap)
		fmt.Printf("   Max Context Tokens: %d\n", m.config.AgentMemory.RAGMaxContextTokens)
		fmt.Printf("   Personal Weight: %.1f\n", m.config.AgentMemory.RAGPersonalWeight)
		fmt.Printf("   Knowledge Weight: %.1f\n", m.config.AgentMemory.RAGKnowledgeWeight)
		fmt.Printf("   Include Sources: %t\n", m.config.AgentMemory.RAGIncludeSources)
	}
	fmt.Printf("\n")

	// Document processing configuration
	fmt.Printf("📄 Document Processing:\n")
	fmt.Printf("   Auto Chunk: %t\n", m.config.AgentMemory.Documents.AutoChunk)
	fmt.Printf("   Supported Types: %v\n", m.config.AgentMemory.Documents.SupportedTypes)
	fmt.Printf("   Max File Size: %s\n", m.config.AgentMemory.Documents.MaxFileSize)
	fmt.Printf("   Metadata Extraction: %t\n", m.config.AgentMemory.Documents.EnableMetadataExtraction)
	fmt.Printf("   URL Scraping: %t\n", m.config.AgentMemory.Documents.EnableURLScraping)
	fmt.Printf("\n")

	// Search configuration
	fmt.Printf("🔎 Search Configuration:\n")
	fmt.Printf("   Hybrid Search: %t\n", m.config.AgentMemory.Search.HybridSearch)
	if m.config.AgentMemory.Search.HybridSearch {
		fmt.Printf("   Keyword Weight: %.1f\n", m.config.AgentMemory.Search.KeywordWeight)
		fmt.Printf("   Semantic Weight: %.1f\n", m.config.AgentMemory.Search.SemanticWeight)
		fmt.Printf("   Enable Reranking: %t\n", m.config.AgentMemory.Search.EnableReranking)
		fmt.Printf("   Query Expansion: %t\n", m.config.AgentMemory.Search.EnableQueryExpansion)
	}

	return nil
}

// maskAPIKey masks API key for display
func maskAPIKey(key string) string {
	if len(key) <= 8 {
		return "***"
	}
	return key[:4] + "..." + key[len(key)-4:]
}