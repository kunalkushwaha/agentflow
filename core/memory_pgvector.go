package core

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"strings"
	"sync"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pgvector/pgvector-go"
)

// PgVectorProvider - production-ready PostgreSQL with pgvector
type PgVectorProvider struct {
	config           AgentMemoryConfig
	pool             *pgxpool.Pool
	embeddingService EmbeddingService
	mutex            sync.RWMutex
	retryConfig      RetryConfig
}

// RetryConfig defines retry behavior for database operations
type RetryConfig struct {
	MaxRetries      int
	BackoffDuration time.Duration
	MaxBackoff      time.Duration
}

// DefaultRetryConfig returns sensible defaults for retry behavior
func DefaultRetryConfig() RetryConfig {
	return RetryConfig{
		MaxRetries:      3,
		BackoffDuration: 100 * time.Millisecond,
		MaxBackoff:      5 * time.Second,
	}
}

// newPgVectorProvider creates a new PgVector provider
func newPgVectorProvider(config AgentMemoryConfig) (Memory, error) {
	provider := &PgVectorProvider{
		config:      config,
		retryConfig: DefaultRetryConfig(),
	}

	// Initialize embedding service
	var embeddingService EmbeddingService
	if config.Embedding.Provider == "openai" {
		if config.Embedding.APIKey == "" {
			return nil, fmt.Errorf("OpenAI API key is required for embedding service")
		}
		embeddingService = NewOpenAIEmbeddingService(config.Embedding.APIKey, config.Embedding.Model)
	} else if config.Embedding.Provider == "ollama" {
		embeddingService = NewOllamaEmbeddingService(config.Embedding.Model, config.Embedding.BaseURL)
	} else {
		// Use dummy embedding service for development
		embeddingService = NewDummyEmbeddingService(config.Dimensions)
	}
	provider.embeddingService = embeddingService

	// Initialize database with enhanced connection pooling
	if err := provider.initialize(); err != nil {
		return nil, fmt.Errorf("failed to initialize PgVector provider: %w", err)
	}

	return provider, nil
}

// Database schema and initialization
func (p *PgVectorProvider) initialize() error {
	// Create enhanced connection pool
	config, err := p.configurePool()
	if err != nil {
		return fmt.Errorf("failed to configure connection pool: %w", err)
	}

	pool, err := pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		return fmt.Errorf("failed to create connection pool: %w", err)
	}

	p.pool = pool

	// Test connection with retry logic
	if err := p.withRetry(context.Background(), "database ping", func() error {
		return pool.Ping(context.Background())
	}); err != nil {
		return fmt.Errorf("failed to ping database: %w", err)
	}

	// Create tables and indexes with retry logic
	if err := p.withRetry(context.Background(), "create tables", func() error {
		return p.createTables()
	}); err != nil {
		return fmt.Errorf("failed to create tables: %w", err)
	}

	return nil
}

func (p *PgVectorProvider) createTables() error {
	ctx := context.Background()

	// Enable pgvector extension
	if _, err := p.pool.Exec(ctx, "CREATE EXTENSION IF NOT EXISTS vector"); err != nil {
		return fmt.Errorf("failed to create vector extension: %w", err)
	}

	// Create personal memory table
	personalMemorySchema := `
		CREATE TABLE IF NOT EXISTS personal_memory (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			session_id VARCHAR(255) NOT NULL,
			content TEXT NOT NULL,
			embedding vector(%d),
			tags TEXT[],
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		);
		CREATE INDEX IF NOT EXISTS idx_personal_memory_session ON personal_memory(session_id);
		CREATE INDEX IF NOT EXISTS idx_personal_memory_embedding ON personal_memory USING ivfflat (embedding vector_cosine_ops);
		CREATE INDEX IF NOT EXISTS idx_personal_memory_tags ON personal_memory USING gin(tags);
	`

	if _, err := p.pool.Exec(ctx, fmt.Sprintf(personalMemorySchema, p.config.Dimensions)); err != nil {
		return fmt.Errorf("failed to create personal_memory table: %w", err)
	}

	// Create key-value store table
	keyValueSchema := `
		CREATE TABLE IF NOT EXISTS key_value_store (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			session_id VARCHAR(255) NOT NULL,
			key VARCHAR(255) NOT NULL,
			value JSONB NOT NULL,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			UNIQUE(session_id, key)
		);
		CREATE INDEX IF NOT EXISTS idx_kv_session_key ON key_value_store(session_id, key);
	`

	if _, err := p.pool.Exec(ctx, keyValueSchema); err != nil {
		return fmt.Errorf("failed to create key_value_store table: %w", err)
	}

	// Create chat history table
	chatHistorySchema := `
		CREATE TABLE IF NOT EXISTS chat_history (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			session_id VARCHAR(255) NOT NULL,
			role VARCHAR(50) NOT NULL,
			content TEXT NOT NULL,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		);
		CREATE INDEX IF NOT EXISTS idx_chat_history_session ON chat_history(session_id, created_at);
	`

	if _, err := p.pool.Exec(ctx, chatHistorySchema); err != nil {
		return fmt.Errorf("failed to create chat_history table: %w", err)
	}

	// Create documents table (knowledge base)
	documentsSchema := `
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
		CREATE INDEX IF NOT EXISTS idx_documents_source ON documents(source);
		CREATE INDEX IF NOT EXISTS idx_documents_type ON documents(doc_type);
		CREATE INDEX IF NOT EXISTS idx_documents_tags ON documents USING gin(tags);
	`

	if _, err := p.pool.Exec(ctx, documentsSchema); err != nil {
		return fmt.Errorf("failed to create documents table: %w", err)
	}

	// Create knowledge base table with embeddings
	knowledgeSchema := `
		CREATE TABLE IF NOT EXISTS knowledge_base (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			document_id VARCHAR(255) NOT NULL REFERENCES documents(id) ON DELETE CASCADE,
			content TEXT NOT NULL,
			embedding vector(%d),
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		);
		CREATE INDEX IF NOT EXISTS idx_knowledge_document ON knowledge_base(document_id);
		CREATE INDEX IF NOT EXISTS idx_knowledge_embedding ON knowledge_base USING ivfflat (embedding vector_cosine_ops);
	`

	if _, err := p.pool.Exec(ctx, fmt.Sprintf(knowledgeSchema, p.config.Dimensions)); err != nil {
		return fmt.Errorf("failed to create knowledge_base table: %w", err)
	}

	return nil
}

func (p *PgVectorProvider) Store(ctx context.Context, content string, tags ...string) error {
	sessionID := GetSessionID(ctx)

	// Generate embedding using the embedding service
	embedding, err := p.embeddingService.GenerateEmbedding(ctx, content)
	if err != nil {
		return fmt.Errorf("failed to generate embedding: %w", err)
	}

	// Use retry logic for database operations
	return p.withRetry(ctx, "store memory", func() error {
		query := `
			INSERT INTO personal_memory (session_id, content, embedding, tags)
			VALUES ($1, $2, $3, $4)
		`

		_, err := p.pool.Exec(ctx, query, sessionID, content, pgvector.NewVector(embedding), tags)
		if err != nil {
			return fmt.Errorf("failed to store memory: %w", err)
		}

		return nil
	})
}

func (p *PgVectorProvider) Query(ctx context.Context, query string, limit ...int) ([]Result, error) {
	sessionID := GetSessionID(ctx)
	maxResults := p.config.MaxResults
	if len(limit) > 0 && limit[0] > 0 {
		maxResults = limit[0]
	}

	// Generate query embedding
	queryEmbedding, err := p.embeddingService.GenerateEmbedding(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to generate query embedding: %w", err)
	}

	// Use cosine similarity search
	sqlQuery := `
		SELECT content, tags, created_at, 
			   1 - (embedding <=> $1) as similarity_score
		FROM personal_memory 
		WHERE session_id = $2
		ORDER BY embedding <=> $1
		LIMIT $3
	`

	rows, err := p.pool.Query(ctx, sqlQuery, pgvector.NewVector(queryEmbedding), sessionID, maxResults)
	if err != nil {
		return nil, fmt.Errorf("failed to query memory: %w", err)
	}
	defer rows.Close()

	var results []Result
	for rows.Next() {
		var content string
		var tags []string
		var createdAt time.Time
		var score float32

		if err := rows.Scan(&content, &tags, &createdAt, &score); err != nil {
			return nil, fmt.Errorf("failed to scan result: %w", err)
		}

		results = append(results, Result{
			Content:   content,
			Score:     score,
			Tags:      tags,
			CreatedAt: createdAt,
		})
	}

	return results, nil
}

func (p *PgVectorProvider) Remember(ctx context.Context, key string, value any) error {
	sessionID := GetSessionID(ctx)

	// Convert value to JSON
	jsonValue, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal value: %w", err)
	}

	query := `
		INSERT INTO key_value_store (session_id, key, value)
		VALUES ($1, $2, $3)
		ON CONFLICT (session_id, key) 
		DO UPDATE SET value = $3, updated_at = NOW()
	`

	_, err = p.pool.Exec(ctx, query, sessionID, key, jsonValue)
	if err != nil {
		return fmt.Errorf("failed to store key-value: %w", err)
	}

	return nil
}

func (p *PgVectorProvider) Recall(ctx context.Context, key string) (any, error) {
	sessionID := GetSessionID(ctx)

	query := `SELECT value FROM key_value_store WHERE session_id = $1 AND key = $2`

	var jsonValue []byte
	err := p.pool.QueryRow(ctx, query, sessionID, key).Scan(&jsonValue)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to recall key-value: %w", err)
	}

	var value any
	if err := json.Unmarshal(jsonValue, &value); err != nil {
		return nil, fmt.Errorf("failed to unmarshal value: %w", err)
	}

	return value, nil
}

func (p *PgVectorProvider) AddMessage(ctx context.Context, role, content string) error {
	sessionID := GetSessionID(ctx)

	query := `
		INSERT INTO chat_history (session_id, role, content)
		VALUES ($1, $2, $3)
	`

	_, err := p.pool.Exec(ctx, query, sessionID, role, content)
	if err != nil {
		return fmt.Errorf("failed to add message: %w", err)
	}

	return nil
}

func (p *PgVectorProvider) GetHistory(ctx context.Context, limit ...int) ([]Message, error) {
	sessionID := GetSessionID(ctx)

	var query string
	var args []any

	if len(limit) > 0 && limit[0] > 0 {
		query = `
			SELECT role, content, created_at 
			FROM chat_history 
			WHERE session_id = $1 
			ORDER BY created_at DESC 
			LIMIT $2
		`
		args = []any{sessionID, limit[0]}
	} else {
		query = `
			SELECT role, content, created_at 
			FROM chat_history 
			WHERE session_id = $1 
			ORDER BY created_at ASC
		`
		args = []any{sessionID}
	}

	rows, err := p.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get history: %w", err)
	}
	defer rows.Close()

	var messages []Message
	for rows.Next() {
		var role, content string
		var createdAt time.Time

		if err := rows.Scan(&role, &content, &createdAt); err != nil {
			return nil, fmt.Errorf("failed to scan message: %w", err)
		}

		messages = append(messages, Message{
			Role:      role,
			Content:   content,
			CreatedAt: createdAt,
		})
	}

	// If we used LIMIT, reverse the order to get chronological order
	if len(limit) > 0 && limit[0] > 0 {
		for i := len(messages)/2 - 1; i >= 0; i-- {
			opp := len(messages) - 1 - i
			messages[i], messages[opp] = messages[opp], messages[i]
		}
	}

	return messages, nil
}

func (p *PgVectorProvider) NewSession() string {
	return generateID()
}

func (p *PgVectorProvider) SetSession(ctx context.Context, sessionID string) context.Context {
	return WithMemory(ctx, p, sessionID)
}

func (p *PgVectorProvider) ClearSession(ctx context.Context) error {
	sessionID := GetSessionID(ctx)

	// Start a transaction
	tx, err := p.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to start transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// Clear personal memory
	if _, err := tx.Exec(ctx, "DELETE FROM personal_memory WHERE session_id = $1", sessionID); err != nil {
		return fmt.Errorf("failed to clear personal memory: %w", err)
	}

	// Clear key-value store
	if _, err := tx.Exec(ctx, "DELETE FROM key_value_store WHERE session_id = $1", sessionID); err != nil {
		return fmt.Errorf("failed to clear key-value store: %w", err)
	}

	// Clear chat history
	if _, err := tx.Exec(ctx, "DELETE FROM chat_history WHERE session_id = $1", sessionID); err != nil {
		return fmt.Errorf("failed to clear chat history: %w", err)
	}

	// Commit transaction
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func (p *PgVectorProvider) Close() error {
	if p.pool != nil {
		p.pool.Close()
	}
	return nil
}

// RAG methods for PgVectorProvider

func (p *PgVectorProvider) IngestDocument(ctx context.Context, doc Document) error {
	// Start a transaction
	tx, err := p.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to start transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// Generate ID if not provided
	if doc.ID == "" {
		doc.ID = generateID()
	}

	// Set timestamps
	if doc.CreatedAt.IsZero() {
		doc.CreatedAt = time.Now()
	}
	doc.UpdatedAt = time.Now()

	// Insert document metadata
	documentQuery := `
		INSERT INTO documents (id, title, content, source, doc_type, metadata, tags, created_at, updated_at, chunk_index, chunk_total)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		ON CONFLICT (id) DO UPDATE SET
			title = $2, content = $3, source = $4, doc_type = $5, metadata = $6, tags = $7, updated_at = $9, chunk_index = $10, chunk_total = $11
	`

	metadataJSON, err := json.Marshal(doc.Metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	_, err = tx.Exec(ctx, documentQuery,
		doc.ID, doc.Title, doc.Content, doc.Source, string(doc.Type),
		metadataJSON, doc.Tags, doc.CreatedAt, doc.UpdatedAt,
		doc.ChunkIndex, doc.ChunkTotal)
	if err != nil {
		return fmt.Errorf("failed to insert document: %w", err)
	}

	// Generate embedding for the document content
	embedding, err := p.embeddingService.GenerateEmbedding(ctx, doc.Content)
	if err != nil {
		return fmt.Errorf("failed to generate document embedding: %w", err)
	}

	// Insert into knowledge base with embedding
	// First, delete any existing entries for this document
	_, err = tx.Exec(ctx, "DELETE FROM knowledge_base WHERE document_id = $1", doc.ID)
	if err != nil {
		return fmt.Errorf("failed to delete existing knowledge base entry: %w", err)
	}

	// Then insert the new entry
	knowledgeQuery := `
		INSERT INTO knowledge_base (document_id, content, embedding, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)
	`

	_, err = tx.Exec(ctx, knowledgeQuery,
		doc.ID, doc.Content, pgvector.NewVector(embedding),
		doc.CreatedAt, doc.UpdatedAt)
	if err != nil {
		return fmt.Errorf("failed to insert knowledge base entry: %w", err)
	}

	// Commit transaction
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func (p *PgVectorProvider) IngestDocuments(ctx context.Context, docs []Document) error {
	for _, doc := range docs {
		if err := p.IngestDocument(ctx, doc); err != nil {
			return fmt.Errorf("failed to ingest document %s: %w", doc.ID, err)
		}
	}
	return nil
}

func (p *PgVectorProvider) SearchKnowledge(ctx context.Context, query string, options ...SearchOption) ([]KnowledgeResult, error) {
	// Apply search options
	config := &SearchConfig{
		Limit:            p.config.KnowledgeMaxResults,
		ScoreThreshold:   p.config.KnowledgeScoreThreshold,
		IncludeKnowledge: true,
	}
	for _, opt := range options {
		opt(config)
	}

	// Generate query embedding
	queryEmbedding, err := p.embeddingService.GenerateEmbedding(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to generate query embedding: %w", err)
	}

	// Build base SQL query
	baseQuery := `
		SELECT 
			kb.content,
			1 - (kb.embedding <=> $1) as similarity_score,
			d.source,
			d.title,
			d.id as document_id,
			d.metadata,
			d.tags,
			kb.created_at,
			d.chunk_index
		FROM knowledge_base kb
		JOIN documents d ON kb.document_id = d.id
		WHERE 1=1
	`

	args := []any{pgvector.NewVector(queryEmbedding)}
	argIndex := 2

	// Apply filters
	if len(config.Sources) > 0 {
		baseQuery += fmt.Sprintf(" AND d.source = $%d", argIndex)
		args = append(args, config.Sources[0])
		argIndex++
	}

	if len(config.DocumentTypes) > 0 {
		baseQuery += fmt.Sprintf(" AND d.doc_type = $%d", argIndex)
		args = append(args, string(config.DocumentTypes[0]))
		argIndex++
	}

	if len(config.Tags) > 0 {
		baseQuery += fmt.Sprintf(" AND d.tags && $%d", argIndex)
		args = append(args, config.Tags)
		argIndex++
	}

	if config.DateRange != nil {
		baseQuery += fmt.Sprintf(" AND d.created_at BETWEEN $%d AND $%d", argIndex, argIndex+1)
		args = append(args, config.DateRange.Start, config.DateRange.End)
		argIndex += 2
	}

	// Add score threshold filter
	if config.ScoreThreshold > 0 {
		baseQuery += fmt.Sprintf(" AND (1 - (kb.embedding <=> $1)) >= $%d", argIndex)
		args = append(args, config.ScoreThreshold)
		argIndex++
	}

	// Add ordering and limit
	baseQuery += " ORDER BY kb.embedding <=> $1 LIMIT $" + fmt.Sprintf("%d", argIndex)
	args = append(args, config.Limit)

	// Execute query
	rows, err := p.pool.Query(ctx, baseQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to search knowledge base: %w", err)
	}
	defer rows.Close()

	var results []KnowledgeResult
	for rows.Next() {
		var content, source, title, documentID string
		var score float32
		var metadataJSON []byte
		var tags []string
		var createdAt time.Time
		var chunkIndex int

		err := rows.Scan(&content, &score, &source, &title, &documentID,
			&metadataJSON, &tags, &createdAt, &chunkIndex)
		if err != nil {
			return nil, fmt.Errorf("failed to scan knowledge result: %w", err)
		}

		// Parse metadata
		var metadata map[string]any
		if len(metadataJSON) > 0 {
			if err := json.Unmarshal(metadataJSON, &metadata); err != nil {
				return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
			}
		}

		results = append(results, KnowledgeResult{
			Content:    content,
			Score:      score,
			Source:     source,
			Title:      title,
			DocumentID: documentID,
			Metadata:   metadata,
			Tags:       tags,
			CreatedAt:  createdAt,
			ChunkIndex: chunkIndex,
		})
	}

	return results, nil
}

func (p *PgVectorProvider) SearchAll(ctx context.Context, query string, options ...SearchOption) (*HybridResult, error) {
	start := time.Now()

	// Apply search options
	config := &SearchConfig{
		Limit:            p.config.MaxResults,
		ScoreThreshold:   0.0,
		IncludePersonal:  true,
		IncludeKnowledge: true,
	}
	for _, opt := range options {
		opt(config)
	}

	result := &HybridResult{
		Query:          query,
		PersonalMemory: []Result{},
		Knowledge:      []KnowledgeResult{},
	}

	// Search personal memory if enabled
	if config.IncludePersonal {
		personalResults, err := p.Query(ctx, query, config.Limit)
		if err != nil {
			return nil, fmt.Errorf("failed to search personal memory: %w", err)
		}
		result.PersonalMemory = personalResults
	}

	// Search knowledge base if enabled
	if config.IncludeKnowledge {
		knowledgeResults, err := p.SearchKnowledge(ctx, query, options...)
		if err != nil {
			return nil, fmt.Errorf("failed to search knowledge base: %w", err)
		}
		result.Knowledge = knowledgeResults
	}

	result.TotalResults = len(result.PersonalMemory) + len(result.Knowledge)
	result.SearchTime = time.Since(start)

	return result, nil
}

func (p *PgVectorProvider) BuildContext(ctx context.Context, query string, options ...ContextOption) (*RAGContext, error) {
	// Apply context options
	config := &ContextConfig{
		MaxTokens:       p.config.RAGMaxContextTokens,
		PersonalWeight:  p.config.RAGPersonalWeight,
		KnowledgeWeight: p.config.RAGKnowledgeWeight,
		HistoryLimit:    5,
		IncludeSources:  p.config.RAGIncludeSources,
		FormatTemplate:  "", // Use default formatting
	}
	for _, opt := range options {
		opt(config)
	}

	// Get hybrid search results
	searchResults, err := p.SearchAll(ctx, query,
		WithLimit(config.MaxTokens/100), // Rough estimate: 100 tokens per result
		WithIncludePersonal(config.PersonalWeight > 0),
		WithIncludeKnowledge(config.KnowledgeWeight > 0),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to search for context: %w", err)
	}

	// Get chat history
	history, err := p.GetHistory(ctx, config.HistoryLimit)
	if err != nil {
		return nil, fmt.Errorf("failed to get chat history: %w", err)
	}

	// Build context text
	contextText := p.formatContextText(query, searchResults, history, config)

	// Collect sources
	sources := []string{}
	for _, result := range searchResults.Knowledge {
		if result.Source != "" {
			sources = append(sources, result.Source)
		}
	}

	// Remove duplicates
	sources = removeDuplicates(sources)

	return &RAGContext{
		Query:          query,
		PersonalMemory: searchResults.PersonalMemory,
		Knowledge:      searchResults.Knowledge,
		ChatHistory:    history,
		ContextText:    contextText,
		Sources:        sources,
		TokenCount:     estimateTokenCount(contextText),
		Timestamp:      time.Now(),
	}, nil
}

// Helper function for PgVector context formatting
func (p *PgVectorProvider) formatContextText(query string, results *HybridResult, history []Message, config *ContextConfig) string {
	if config.FormatTemplate != "" {
		// TODO: Implement custom template formatting
		return config.FormatTemplate
	}

	var builder strings.Builder

	// Add query
	builder.WriteString(fmt.Sprintf("Query: %s\n\n", query))

	// Add personal memory context
	if len(results.PersonalMemory) > 0 {
		builder.WriteString("Personal Memory:\n")
		for i, result := range results.PersonalMemory {
			builder.WriteString(fmt.Sprintf("%d. %s\n", i+1, result.Content))
		}
		builder.WriteString("\n")
	}

	// Add knowledge base context
	if len(results.Knowledge) > 0 {
		builder.WriteString("Knowledge Base:\n")
		for i, result := range results.Knowledge {
			source := ""
			if config.IncludeSources && result.Source != "" {
				source = fmt.Sprintf(" (Source: %s)", result.Source)
			}
			builder.WriteString(fmt.Sprintf("%d. %s%s\n", i+1, result.Content, source))
		}
		builder.WriteString("\n")
	}

	// Add recent chat history
	if len(history) > 0 {
		builder.WriteString("Recent Conversation:\n")
		for _, msg := range history {
			builder.WriteString(fmt.Sprintf("%s: %s\n", msg.Role, msg.Content))
		}
	}

	return builder.String()
}

// Enhanced retry logic and methods from memory_pgvector_enhanced.go

// withRetry executes a function with exponential backoff retry logic
func (p *PgVectorProvider) withRetry(ctx context.Context, operation string, fn func() error) error {
	config := p.retryConfig
	var lastErr error

	for attempt := 1; attempt <= config.MaxRetries; attempt++ {
		if err := fn(); err != nil {
			lastErr = err

			// Check if error is retryable
			if !isRetryableError(err) {
				return fmt.Errorf("%s failed (non-retryable): %w", operation, err)
			}

			// Last attempt failed
			if attempt == config.MaxRetries {
				return fmt.Errorf("%s failed after %d attempts: %w", operation, config.MaxRetries, err)
			}

			// Calculate delay with exponential backoff
			delay := time.Duration(float64(config.BackoffDuration) * math.Pow(2.0, float64(attempt-1)))
			if delay > config.MaxBackoff {
				delay = config.MaxBackoff
			}

			// Wait before retry
			select {
			case <-ctx.Done():
				return fmt.Errorf("%s cancelled during retry: %w", operation, ctx.Err())
			case <-time.After(delay):
				// Continue to next attempt
			}
		} else {
			// Success
			return nil
		}
	}

	return fmt.Errorf("%s failed after %d attempts: %w", operation, config.MaxRetries, lastErr)
}

// isRetryableError determines if an error is worth retrying
func isRetryableError(err error) bool {
	if err == nil {
		return false
	}

	errStr := err.Error()

	// Network-related errors that are typically transient
	retryableErrors := []string{
		"connection refused",
		"connection reset",
		"timeout",
		"temporary failure",
		"server is not ready",
		"too many connections",
		"connection lost",
		"network unreachable",
		"host unreachable",
	}

	for _, retryable := range retryableErrors {
		if strings.Contains(errStr, retryable) {
			return true
		}
	}

	return false
}

// BatchStoreRequest represents a single item in a batch store operation
type BatchStoreRequest struct {
	Content string
	Tags    []string
}

// BatchStore stores multiple memories in a single transaction
func (p *PgVectorProvider) BatchStore(ctx context.Context, requests []BatchStoreRequest) error {
	if len(requests) == 0 {
		return nil
	}

	sessionID := GetSessionID(ctx)

	// Generate embeddings in batch
	texts := make([]string, len(requests))
	for i, req := range requests {
		texts[i] = req.Content
	}

	embeddings, err := p.embeddingService.GenerateEmbeddings(ctx, texts)
	if err != nil {
		return fmt.Errorf("failed to generate batch embeddings: %w", err)
	}

	if len(embeddings) != len(requests) {
		return fmt.Errorf("embedding count mismatch: got %d, expected %d", len(embeddings), len(requests))
	}

	// Store in database with transaction
	return p.withRetry(ctx, "batch store", func() error {
		tx, err := p.pool.Begin(ctx)
		if err != nil {
			return fmt.Errorf("failed to begin transaction: %w", err)
		}
		defer tx.Rollback(ctx)

		// Prepare batch insert
		query := `
			INSERT INTO personal_memory (session_id, content, embedding, tags)
			VALUES ($1, $2, $3, $4)
		`

		for i, req := range requests {
			_, err = tx.Exec(ctx, query, sessionID, req.Content, pgvector.NewVector(embeddings[i]), req.Tags)
			if err != nil {
				return fmt.Errorf("failed to insert batch item %d: %w", i, err)
			}
		}

		return tx.Commit(ctx)
	})
}

// BatchIngestDocuments ingests multiple documents efficiently
func (p *PgVectorProvider) BatchIngestDocuments(ctx context.Context, docs []Document) error {
	if len(docs) == 0 {
		return nil
	}

	// Generate embeddings in batch
	texts := make([]string, len(docs))
	for i, doc := range docs {
		texts[i] = doc.Content
	}

	embeddings, err := p.embeddingService.GenerateEmbeddings(ctx, texts)
	if err != nil {
		return fmt.Errorf("failed to generate batch embeddings: %w", err)
	}

	// Store documents in batches to avoid overwhelming the database
	batchSize := p.config.Embedding.MaxBatchSize
	if batchSize <= 0 {
		batchSize = 50 // Default batch size
	}

	for i := 0; i < len(docs); i += batchSize {
		end := i + batchSize
		if end > len(docs) {
			end = len(docs)
		}

		err := p.batchIngestChunk(ctx, docs[i:end], embeddings[i:end])
		if err != nil {
			return fmt.Errorf("failed to ingest batch chunk %d-%d: %w", i, end-1, err)
		}
	}

	return nil
}

func (p *PgVectorProvider) batchIngestChunk(ctx context.Context, docs []Document, embeddings [][]float32) error {
	return p.withRetry(ctx, "batch ingest chunk", func() error {
		tx, err := p.pool.Begin(ctx)
		if err != nil {
			return fmt.Errorf("failed to begin transaction: %w", err)
		}
		defer tx.Rollback(ctx)

		now := time.Now()

		// Batch insert documents
		documentQuery := `
			INSERT INTO documents (id, title, content, source, doc_type, metadata, tags, created_at, updated_at, chunk_index, chunk_total)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
			ON CONFLICT (id) DO UPDATE SET
				title = EXCLUDED.title, content = EXCLUDED.content, source = EXCLUDED.source, 
				doc_type = EXCLUDED.doc_type, metadata = EXCLUDED.metadata, tags = EXCLUDED.tags, 
				updated_at = EXCLUDED.updated_at, chunk_index = EXCLUDED.chunk_index, chunk_total = EXCLUDED.chunk_total
		`

		knowledgeQuery := `
			INSERT INTO knowledge_base (document_id, content, embedding, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5)
		`

		for i, doc := range docs {
			// Set default values
			if doc.ID == "" {
				doc.ID = generateID()
			}
			if doc.CreatedAt.IsZero() {
				doc.CreatedAt = now
			}
			doc.UpdatedAt = now

			// First, delete any existing knowledge base entries for this document
			_, err = tx.Exec(ctx, "DELETE FROM knowledge_base WHERE document_id = $1", doc.ID)
			if err != nil {
				return fmt.Errorf("failed to delete existing knowledge base entries for doc %s: %w", doc.ID, err)
			}

			// Insert document
			metadataJSON, err := marshalMetadata(doc.Metadata)
			if err != nil {
				return fmt.Errorf("failed to marshal metadata for doc %s: %w", doc.ID, err)
			}

			_, err = tx.Exec(ctx, documentQuery,
				doc.ID, doc.Title, doc.Content, doc.Source, string(doc.Type),
				metadataJSON, doc.Tags, doc.CreatedAt, doc.UpdatedAt,
				doc.ChunkIndex, doc.ChunkTotal)
			if err != nil {
				return fmt.Errorf("failed to insert document %s: %w", doc.ID, err)
			}

			// Insert knowledge base entry
			_, err = tx.Exec(ctx, knowledgeQuery,
				doc.ID, doc.Content, pgvector.NewVector(embeddings[i]),
				doc.CreatedAt, doc.UpdatedAt)
			if err != nil {
				return fmt.Errorf("failed to insert knowledge base entry for doc %s: %w", doc.ID, err)
			}
		}

		return tx.Commit(ctx)
	})
}

// Enhanced connection pool configuration
func (p *PgVectorProvider) configurePool() (*pgxpool.Config, error) {
	config, err := pgxpool.ParseConfig(p.config.Connection)
	if err != nil {
		return nil, fmt.Errorf("failed to parse connection string: %w", err)
	}

	// Configure connection pool settings for optimal performance
	config.MaxConns = 25                       // Maximum number of connections
	config.MinConns = 5                        // Minimum number of connections
	config.MaxConnLifetime = time.Hour         // Maximum lifetime of a connection
	config.MaxConnIdleTime = 30 * time.Minute  // Maximum idle time
	config.HealthCheckPeriod = 1 * time.Minute // Health check interval

	// Set statement timeout for long-running vector operations
	config.ConnConfig.RuntimeParams = map[string]string{
		"statement_timeout":                   "30s",
		"idle_in_transaction_session_timeout": "60s",
	}

	return config, nil
}

// UpdatePoolConfig updates the connection pool with optimal settings
func (p *PgVectorProvider) UpdatePoolConfig() error {
	if p.pool != nil {
		p.pool.Close()
	}

	config, err := p.configurePool()
	if err != nil {
		return err
	}

	pool, err := pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		return fmt.Errorf("failed to create enhanced connection pool: %w", err)
	}

	p.pool = pool

	// Test the connection
	return p.withRetry(context.Background(), "pool connection test", func() error {
		return p.pool.Ping(context.Background())
	})
}

// Helper function for metadata marshaling
func marshalMetadata(metadata map[string]any) ([]byte, error) {
	if metadata == nil {
		return []byte("{}"), nil
	}
	return json.Marshal(metadata)
}
