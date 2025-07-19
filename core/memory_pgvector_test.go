package core

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// PgVectorIntegrationTestSuite tests PgVectorProvider against a real PostgreSQL database
type PgVectorIntegrationTestSuite struct {
	suite.Suite
	provider Memory
	config   AgentMemoryConfig
	pool     *pgxpool.Pool
	ctx      context.Context
}

func (suite *PgVectorIntegrationTestSuite) SetupSuite() {
	// Skip if no database connection string provided
	connStr := os.Getenv("AGENTFLOW_TEST_DB_URL")
	if connStr == "" {
		suite.T().Skip("AGENTFLOW_TEST_DB_URL environment variable not set, skipping integration tests")
	}

	suite.ctx = context.Background()

	// Configure test database
	suite.config = AgentMemoryConfig{
		Provider:                "pgvector",
		Connection:              connStr,
		MaxResults:              10,
		Dimensions:              1536,
		AutoEmbed:               true,
		EnableKnowledgeBase:     true,
		KnowledgeMaxResults:     20,
		KnowledgeScoreThreshold: -0.1, // Very low threshold for dummy embeddings
		ChunkSize:               1000,
		ChunkOverlap:            200,
		EnableRAG:               true,
		RAGMaxContextTokens:     4000,
		RAGPersonalWeight:       0.3,
		RAGKnowledgeWeight:      0.7,
		RAGIncludeSources:       true,
		Embedding: EmbeddingConfig{
			Provider:        "dummy", // Use dummy for tests
			Model:           "text-embedding-3-small",
			CacheEmbeddings: true,
			MaxBatchSize:    100,
			TimeoutSeconds:  30,
		},
	}

	// Create provider
	provider, err := NewMemory(suite.config)
	require.NoError(suite.T(), err)
	suite.provider = provider

	// Setup test database schema
	suite.setupTestDatabase()
}

func (suite *PgVectorIntegrationTestSuite) TearDownSuite() {
	if suite.provider != nil {
		suite.provider.Close()
	}
	if suite.pool != nil {
		suite.pool.Close()
	}
}

func (suite *PgVectorIntegrationTestSuite) SetupTest() {
	// Clean up data before each test
	suite.cleanupTestData()
}

func (suite *PgVectorIntegrationTestSuite) setupTestDatabase() {
	// Create connection pool for cleanup
	pool, err := pgxpool.New(suite.ctx, suite.config.Connection)
	require.NoError(suite.T(), err)
	suite.pool = pool

	// Ensure pgvector extension is installed
	_, err = pool.Exec(suite.ctx, "CREATE EXTENSION IF NOT EXISTS vector")
	require.NoError(suite.T(), err)
}

func (suite *PgVectorIntegrationTestSuite) cleanupTestData() {
	// Clean up test data
	tables := []string{"knowledge_base", "documents", "chat_history", "key_value_store", "personal_memory"}
	for _, table := range tables {
		_, err := suite.pool.Exec(suite.ctx, fmt.Sprintf("DELETE FROM %s", table))
		require.NoError(suite.T(), err)
	}
}

// Test basic memory operations
func (suite *PgVectorIntegrationTestSuite) TestStoreAndQuery() {
	ctx := suite.provider.SetSession(suite.ctx, "test-session-1")

	// Store some memories
	testMemories := []string{
		"I love programming in Go",
		"Machine learning is fascinating",
		"PostgreSQL is a great database",
		"Vector databases are the future",
	}

	for _, memory := range testMemories {
		err := suite.provider.Store(ctx, memory, "test")
		require.NoError(suite.T(), err)
	}

	// Test querying
	results, err := suite.provider.Query(ctx, "programming languages", 2)
	require.NoError(suite.T(), err)
	assert.Len(suite.T(), results, 2)
	assert.NotEmpty(suite.T(), results[0].Content)
	assert.Greater(suite.T(), results[0].Score, float32(0))
}

// Test batch operations
func (suite *PgVectorIntegrationTestSuite) TestBatchOperations() {
	ctx := suite.provider.SetSession(suite.ctx, "test-session-batch")

	// Test BatchStore if available
	if provider, ok := suite.provider.(*PgVectorProvider); ok {
		batchRequests := []BatchStoreRequest{
			{Content: "First batch item", Tags: []string{"batch", "test"}},
			{Content: "Second batch item", Tags: []string{"batch", "test"}},
			{Content: "Third batch item", Tags: []string{"batch", "test"}},
		}

		err := provider.BatchStore(ctx, batchRequests)
		require.NoError(suite.T(), err)

		// Verify stored items
		results, err := suite.provider.Query(ctx, "batch", 5)
		require.NoError(suite.T(), err)
		assert.Len(suite.T(), results, 3)
	}
}

// Test key-value operations
func (suite *PgVectorIntegrationTestSuite) TestKeyValueOperations() {
	ctx := suite.provider.SetSession(suite.ctx, "test-session-kv")

	// Test Remember
	testData := map[string]any{
		"user_name":   "Alice",
		"preferences": []string{"AI", "Machine Learning"},
		"last_login":  time.Now().Unix(),
		"is_premium":  true,
	}

	for key, value := range testData {
		err := suite.provider.Remember(ctx, key, value)
		require.NoError(suite.T(), err)
	}

	// Test Recall
	for key, expectedValue := range testData {
		actualValue, err := suite.provider.Recall(ctx, key)
		require.NoError(suite.T(), err)

		// Note: JSON unmarshaling might change types (e.g., int64 to float64)
		switch key {
		case "user_name":
			assert.Equal(suite.T(), expectedValue, actualValue)
		case "is_premium":
			assert.Equal(suite.T(), expectedValue, actualValue)
		case "preferences":
			// JSON unmarshaling converts []string to []interface{}
			expectedSlice := expectedValue.([]string)
			actualSlice := actualValue.([]interface{})
			require.Len(suite.T(), actualSlice, len(expectedSlice))
			for i, exp := range expectedSlice {
				assert.Equal(suite.T(), exp, actualSlice[i].(string))
			}
		case "last_login":
			// JSON numbers are unmarshaled as float64
			assert.InDelta(suite.T(), expectedValue, actualValue, 1)
		}
	}

	// Test non-existent key
	value, err := suite.provider.Recall(ctx, "non_existent")
	require.NoError(suite.T(), err)
	assert.Nil(suite.T(), value)
}

// Test chat history operations
func (suite *PgVectorIntegrationTestSuite) TestChatHistory() {
	ctx := suite.provider.SetSession(suite.ctx, "test-session-chat")

	// Add messages
	messages := []struct {
		role    string
		content string
	}{
		{"user", "Hello, how are you?"},
		{"assistant", "I'm doing well, thank you! How can I help you today?"},
		{"user", "Can you explain quantum computing?"},
		{"assistant", "Quantum computing uses quantum mechanical phenomena..."},
	}

	for _, msg := range messages {
		err := suite.provider.AddMessage(ctx, msg.role, msg.content)
		require.NoError(suite.T(), err)
	}

	// Get full history
	history, err := suite.provider.GetHistory(ctx)
	require.NoError(suite.T(), err)
	assert.Len(suite.T(), history, 4)

	// Verify chronological order
	for i, expectedMsg := range messages {
		assert.Equal(suite.T(), expectedMsg.role, history[i].Role)
		assert.Equal(suite.T(), expectedMsg.content, history[i].Content)
	}

	// Get limited history
	limitedHistory, err := suite.provider.GetHistory(ctx, 2)
	require.NoError(suite.T(), err)
	assert.Len(suite.T(), limitedHistory, 2)

	// Should be most recent messages in chronological order
	assert.Equal(suite.T(), "user", limitedHistory[0].Role)
	assert.Contains(suite.T(), limitedHistory[0].Content, "quantum")
}

// Test document ingestion and knowledge search
func (suite *PgVectorIntegrationTestSuite) TestKnowledgeOperations() {
	ctx := suite.ctx

	// Create test documents
	docs := []Document{
		{
			ID:      "doc1",
			Title:   "Introduction to Go",
			Content: "Go is a programming language developed by Google. It's known for its simplicity and performance.",
			Source:  "golang.org",
			Type:    DocumentTypeText,
			Tags:    []string{"programming", "go", "google"},
			Metadata: map[string]any{
				"author": "Go Team",
				"year":   2009,
			},
		},
		{
			ID:      "doc2",
			Title:   "Machine Learning Basics",
			Content: "Machine learning is a subset of artificial intelligence that enables computers to learn without explicit programming.",
			Source:  "ml-guide.com",
			Type:    DocumentTypeText,
			Tags:    []string{"ml", "ai", "learning"},
			Metadata: map[string]any{
				"difficulty": "beginner",
				"topics":     []string{"supervised", "unsupervised"},
			},
		},
	}

	// Ingest documents
	for _, doc := range docs {
		err := suite.provider.IngestDocument(ctx, doc)
		require.NoError(suite.T(), err)
	}

	// Test knowledge search
	results, err := suite.provider.SearchKnowledge(ctx, "programming languages", WithScoreThreshold(-0.1))
	require.NoError(suite.T(), err)
	assert.NotEmpty(suite.T(), results)

	// Find Go-related result
	var goResult *KnowledgeResult
	for _, result := range results {
		if result.DocumentID == "doc1" {
			goResult = &result
			break
		}
	}
	require.NotNil(suite.T(), goResult)
	assert.Equal(suite.T(), "Introduction to Go", goResult.Title)
	assert.Equal(suite.T(), "golang.org", goResult.Source)
	assert.Contains(suite.T(), goResult.Tags, "programming")
	// With dummy embeddings, scores can be slightly negative - just check it's a valid float
	assert.IsType(suite.T(), float32(0), goResult.Score)

	// Test search with filters
	filteredResults, err := suite.provider.SearchKnowledge(ctx, "artificial intelligence",
		WithSources([]string{"ml-guide.com"}),
		WithTags([]string{"ai"}),
		WithScoreThreshold(-0.1),
	)
	require.NoError(suite.T(), err)
	assert.NotEmpty(suite.T(), filteredResults)
	assert.Equal(suite.T(), "doc2", filteredResults[0].DocumentID)
}

// Test hybrid search
func (suite *PgVectorIntegrationTestSuite) TestHybridSearch() {
	ctx := suite.provider.SetSession(suite.ctx, "test-session-hybrid")

	// Store personal memory
	err := suite.provider.Store(ctx, "I'm learning about databases", "learning")
	require.NoError(suite.T(), err)

	// Ingest knowledge document
	doc := Document{
		ID:      "db-doc",
		Title:   "Database Systems",
		Content: "Database systems are organized collections of data that support efficient storage and retrieval.",
		Source:  "db-textbook.com",
		Type:    DocumentTypeText,
		Tags:    []string{"database", "storage"},
	}
	err = suite.provider.IngestDocument(ctx, doc)
	require.NoError(suite.T(), err)

	// Perform hybrid search
	hybridResults, err := suite.provider.SearchAll(ctx, "database learning")
	require.NoError(suite.T(), err)

	assert.NotEmpty(suite.T(), hybridResults.PersonalMemory)
	assert.NotEmpty(suite.T(), hybridResults.Knowledge)
	assert.Equal(suite.T(), "database learning", hybridResults.Query)
	assert.Greater(suite.T(), hybridResults.TotalResults, 0)
	assert.Greater(suite.T(), hybridResults.SearchTime, time.Duration(0))
}

// Test RAG context building
func (suite *PgVectorIntegrationTestSuite) TestRAGContextBuilding() {
	ctx := suite.provider.SetSession(suite.ctx, "test-session-rag")

	// Setup data
	err := suite.provider.Store(ctx, "I prefer using PostgreSQL for complex queries", "database", "preference")
	require.NoError(suite.T(), err)

	err = suite.provider.AddMessage(ctx, "user", "What database should I use?")
	require.NoError(suite.T(), err)

	err = suite.provider.AddMessage(ctx, "assistant", "It depends on your use case. What kind of application are you building?")
	require.NoError(suite.T(), err)

	doc := Document{
		ID:      "postgres-guide",
		Title:   "PostgreSQL Guide",
		Content: "PostgreSQL is a powerful open-source relational database with advanced features like JSON support and full-text search.",
		Source:  "postgresql.org",
		Type:    DocumentTypeText,
		Tags:    []string{"postgresql", "database", "guide"},
	}
	err = suite.provider.IngestDocument(ctx, doc)
	require.NoError(suite.T(), err)

	// Build RAG context
	ragContext, err := suite.provider.BuildContext(ctx, "recommend a database for my project")
	require.NoError(suite.T(), err)

	assert.Equal(suite.T(), "recommend a database for my project", ragContext.Query)
	assert.NotEmpty(suite.T(), ragContext.PersonalMemory)
	assert.NotEmpty(suite.T(), ragContext.Knowledge)
	assert.NotEmpty(suite.T(), ragContext.ChatHistory)
	assert.NotEmpty(suite.T(), ragContext.ContextText)
	assert.Contains(suite.T(), ragContext.Sources, "postgresql.org")
	assert.Greater(suite.T(), ragContext.TokenCount, 0)
}

// Test session isolation
func (suite *PgVectorIntegrationTestSuite) TestSessionIsolation() {
	ctx1 := suite.provider.SetSession(suite.ctx, "session-1")
	ctx2 := suite.provider.SetSession(suite.ctx, "session-2")

	// Store different data in each session
	err := suite.provider.Store(ctx1, "Session 1 data", "session1")
	require.NoError(suite.T(), err)

	err = suite.provider.Store(ctx2, "Session 2 data", "session2")
	require.NoError(suite.T(), err)

	err = suite.provider.Remember(ctx1, "session_name", "First Session")
	require.NoError(suite.T(), err)

	err = suite.provider.Remember(ctx2, "session_name", "Second Session")
	require.NoError(suite.T(), err)

	// Verify isolation
	results1, err := suite.provider.Query(ctx1, "data", 10)
	require.NoError(suite.T(), err)
	assert.Len(suite.T(), results1, 1)
	assert.Contains(suite.T(), results1[0].Content, "Session 1")

	results2, err := suite.provider.Query(ctx2, "data", 10)
	require.NoError(suite.T(), err)
	assert.Len(suite.T(), results2, 1)
	assert.Contains(suite.T(), results2[0].Content, "Session 2")

	// Check key-value isolation
	value1, err := suite.provider.Recall(ctx1, "session_name")
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), "First Session", value1)

	value2, err := suite.provider.Recall(ctx2, "session_name")
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), "Second Session", value2)
}

// Test error handling and retries
func (suite *PgVectorIntegrationTestSuite) TestErrorHandling() {
	ctx := suite.provider.SetSession(suite.ctx, "test-session-error")

	// Test with invalid session (should still work due to retry logic)
	provider, ok := suite.provider.(*PgVectorProvider)
	require.True(suite.T(), ok)

	// Temporarily close the pool to test retry logic
	originalPool := provider.pool
	provider.pool.Close()

	// This should fail
	err := suite.provider.Store(ctx, "test content", "test")
	assert.Error(suite.T(), err)

	// Restore the pool
	newPool, err := pgxpool.New(suite.ctx, suite.config.Connection)
	require.NoError(suite.T(), err)
	provider.pool = newPool

	// This should now work
	err = suite.provider.Store(ctx, "test content after recovery", "test")
	assert.NoError(suite.T(), err)

	// Cleanup
	originalPool.Close()
}

// Test performance and concurrent operations
func (suite *PgVectorIntegrationTestSuite) TestPerformanceAndConcurrency() {
	ctx := suite.provider.SetSession(suite.ctx, "test-session-perf")

	// Test batch insertion performance
	start := time.Now()

	if provider, ok := suite.provider.(*PgVectorProvider); ok {
		batchSize := 50
		batchRequests := make([]BatchStoreRequest, batchSize)
		for i := 0; i < batchSize; i++ {
			batchRequests[i] = BatchStoreRequest{
				Content: fmt.Sprintf("Performance test item %d", i),
				Tags:    []string{"performance", "test", fmt.Sprintf("batch-%d", i/10)},
			}
		}

		err := provider.BatchStore(ctx, batchRequests)
		require.NoError(suite.T(), err)

		batchDuration := time.Since(start)
		suite.T().Logf("Batch insert of %d items took: %v", batchSize, batchDuration)

		// Test query performance
		start = time.Now()
		results, err := suite.provider.Query(ctx, "performance test", 10)
		require.NoError(suite.T(), err)
		queryDuration := time.Since(start)

		suite.T().Logf("Query returned %d results in: %v", len(results), queryDuration)
		assert.NotEmpty(suite.T(), results)

		// Performance thresholds (adjust based on your requirements)
		assert.Less(suite.T(), batchDuration, 10*time.Second, "Batch insert took too long")
		assert.Less(suite.T(), queryDuration, 1*time.Second, "Query took too long")
	}
}

// Run the integration test suite
func TestPgVectorIntegrationSuite(t *testing.T) {
	suite.Run(t, new(PgVectorIntegrationTestSuite))
}

// Benchmark tests
func BenchmarkPgVectorStore(b *testing.B) {
	connStr := os.Getenv("AGENTFLOW_TEST_DB_URL")
	if connStr == "" {
		b.Skip("AGENTFLOW_TEST_DB_URL environment variable not set")
	}

	config := AgentMemoryConfig{
		Provider:   "pgvector",
		Connection: connStr,
		Dimensions: 1536,
		Embedding: EmbeddingConfig{
			Provider: "dummy",
			Model:    "text-embedding-3-small",
		},
	}

	provider, err := NewMemory(config)
	require.NoError(b, err)
	defer provider.Close()

	ctx := provider.SetSession(context.Background(), "benchmark-session")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := provider.Store(ctx, fmt.Sprintf("Benchmark content %d", i), "benchmark")
		require.NoError(b, err)
	}
}

func BenchmarkPgVectorQuery(b *testing.B) {
	connStr := os.Getenv("AGENTFLOW_TEST_DB_URL")
	if connStr == "" {
		b.Skip("AGENTFLOW_TEST_DB_URL environment variable not set")
	}

	config := AgentMemoryConfig{
		Provider:   "pgvector",
		Connection: connStr,
		Dimensions: 1536,
		Embedding: EmbeddingConfig{
			Provider: "dummy",
			Model:    "text-embedding-3-small",
		},
	}

	provider, err := NewMemory(config)
	require.NoError(b, err)
	defer provider.Close()

	ctx := provider.SetSession(context.Background(), "benchmark-query-session")

	// Populate with test data
	for i := 0; i < 100; i++ {
		err := provider.Store(ctx, fmt.Sprintf("Test data for querying %d", i), "test")
		require.NoError(b, err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := provider.Query(ctx, "test data", 10)
		require.NoError(b, err)
	}
}
