package core

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"
)

// InMemoryProvider - fast in-memory implementation for development/testing
type InMemoryProvider struct {
	mutex     sync.RWMutex
	vectors   map[string]vectorEntry
	keyValues map[string]any
	messages  map[string][]Message // sessionID -> messages
	sessionID string
	config    AgentMemoryConfig

	// NEW: Knowledge base storage (global, not session-scoped)
	knowledge map[string]knowledgeEntry // documentID -> document content
	documents map[string]Document       // documentID -> document metadata
}

type vectorEntry struct {
	Content   string
	Tags      []string
	CreatedAt time.Time
	// For in-memory, we'll use simple string matching instead of real embeddings
}

// NEW: Knowledge base entry for in-memory storage
type knowledgeEntry struct {
	Content   string
	Document  Document
	CreatedAt time.Time
}

func newInMemoryProvider(config AgentMemoryConfig) (Memory, error) {
	return &InMemoryProvider{
		vectors:   make(map[string]vectorEntry),
		keyValues: make(map[string]any),
		messages:  make(map[string][]Message),
		sessionID: "default",
		config:    config,
		knowledge: make(map[string]knowledgeEntry),
		documents: make(map[string]Document),
	}, nil
}

func (m *InMemoryProvider) Store(ctx context.Context, content string, tags ...string) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	sessionID := GetSessionID(ctx)
	key := sessionID + ":" + generateID()

	m.vectors[key] = vectorEntry{
		Content:   content,
		Tags:      tags,
		CreatedAt: time.Now(),
	}

	return nil
}

func (m *InMemoryProvider) Query(ctx context.Context, query string, limit ...int) ([]Result, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	maxResults := m.config.MaxResults
	if len(limit) > 0 && limit[0] > 0 {
		maxResults = limit[0]
	}

	sessionID := GetSessionID(ctx)
	var results []Result

	// Simple text matching for in-memory implementation
	sessionPrefix := sessionID + ":"
	for key, entry := range m.vectors {
		if strings.HasPrefix(key, sessionPrefix) {
			score := calculateScore(entry.Content, query)
			tagScore := float32(0)

			// Check tag matching
			for _, tag := range entry.Tags {
				if tagScore < calculateScore(tag, query) {
					tagScore = calculateScore(tag, query)
				}
			}

			// Use the best score
			finalScore := score
			if tagScore > finalScore {
				finalScore = tagScore
			}

			// Only include if there's a reasonable match
			if finalScore > 0.1 {
				results = append(results, Result{
					Content:   entry.Content,
					Score:     finalScore,
					Tags:      entry.Tags,
					CreatedAt: entry.CreatedAt,
				})

				if len(results) >= maxResults {
					break
				}
			}
		}
	}

	// Sort results by score (highest first)
	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})

	return results, nil
}

func (m *InMemoryProvider) Remember(ctx context.Context, key string, value any) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	sessionID := GetSessionID(ctx)
	fullKey := sessionID + ":" + key
	m.keyValues[fullKey] = value

	return nil
}

func (m *InMemoryProvider) Recall(ctx context.Context, key string) (any, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	sessionID := GetSessionID(ctx)
	fullKey := sessionID + ":" + key

	if value, exists := m.keyValues[fullKey]; exists {
		return value, nil
	}

	return nil, nil
}

func (m *InMemoryProvider) AddMessage(ctx context.Context, role, content string) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	sessionID := GetSessionID(ctx)

	message := Message{
		Role:      role,
		Content:   content,
		CreatedAt: time.Now(),
	}

	m.messages[sessionID] = append(m.messages[sessionID], message)

	return nil
}

func (m *InMemoryProvider) GetHistory(ctx context.Context, limit ...int) ([]Message, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	sessionID := GetSessionID(ctx)
	messages := m.messages[sessionID]

	if len(limit) > 0 && limit[0] > 0 && limit[0] < len(messages) {
		// Return last N messages
		start := len(messages) - limit[0]
		return messages[start:], nil
	}

	return messages, nil
}

func (m *InMemoryProvider) NewSession() string {
	return generateID()
}

func (m *InMemoryProvider) SetSession(ctx context.Context, sessionID string) context.Context {
	return WithMemory(ctx, m, sessionID)
}

func (m *InMemoryProvider) ClearSession(ctx context.Context) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	sessionID := GetSessionID(ctx)

	// Clear personal memory for this session
	sessionPrefix := sessionID + ":"
	for key := range m.vectors {
		if strings.HasPrefix(key, sessionPrefix) {
			delete(m.vectors, key)
		}
	}

	// Clear key-value store for this session
	for key := range m.keyValues {
		if strings.HasPrefix(key, sessionPrefix) {
			delete(m.keyValues, key)
		}
	}

	// Clear chat history for this session
	delete(m.messages, sessionID)

	return nil
}

func (m *InMemoryProvider) Close() error {
	// No cleanup needed for in-memory provider
	return nil
}

// RAG methods for InMemoryProvider

func (m *InMemoryProvider) IngestDocument(ctx context.Context, doc Document) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// Generate ID if not provided
	if doc.ID == "" {
		doc.ID = generateID()
	}

	// Set timestamps
	if doc.CreatedAt.IsZero() {
		doc.CreatedAt = time.Now()
	}
	doc.UpdatedAt = time.Now()

	// Store document metadata
	m.documents[doc.ID] = doc

	// Store in knowledge base
	m.knowledge[doc.ID] = knowledgeEntry{
		Content:   doc.Content,
		Document:  doc,
		CreatedAt: doc.CreatedAt,
	}

	return nil
}

func (m *InMemoryProvider) IngestDocuments(ctx context.Context, docs []Document) error {
	for _, doc := range docs {
		if err := m.IngestDocument(ctx, doc); err != nil {
			return fmt.Errorf("failed to ingest document %s: %w", doc.ID, err)
		}
	}
	return nil
}

func (m *InMemoryProvider) SearchKnowledge(ctx context.Context, query string, options ...SearchOption) ([]KnowledgeResult, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	// Apply search options
	config := &SearchConfig{
		Limit:            m.config.KnowledgeMaxResults,
		ScoreThreshold:   m.config.KnowledgeScoreThreshold,
		IncludeKnowledge: true,
	}
	for _, opt := range options {
		opt(config)
	}

	var results []KnowledgeResult

	// Search through knowledge base
	for docID, entry := range m.knowledge {
		doc := entry.Document

		// Apply filters
		if len(config.Sources) > 0 && !contains(doc.Source, config.Sources[0]) {
			continue
		}

		if len(config.DocumentTypes) > 0 && doc.Type != config.DocumentTypes[0] {
			continue
		}

		if len(config.Tags) > 0 && !containsAnyTag(doc.Tags, config.Tags[0]) {
			continue
		}

		if config.DateRange != nil {
			if doc.CreatedAt.Before(config.DateRange.Start) || doc.CreatedAt.After(config.DateRange.End) {
				continue
			}
		}

		// Calculate score using simple text matching
		score := calculateScore(entry.Content, query)
		titleScore := calculateScore(doc.Title, query)
		if titleScore > score {
			score = titleScore
		}

		// Apply score threshold
		if score < config.ScoreThreshold {
			continue
		}

		if score > 0.1 { // Basic relevance threshold
			results = append(results, KnowledgeResult{
				Content:    entry.Content,
				Score:      score,
				Source:     doc.Source,
				Title:      doc.Title,
				DocumentID: docID,
				Metadata:   doc.Metadata,
				Tags:       doc.Tags,
				CreatedAt:  entry.CreatedAt,
				ChunkIndex: doc.ChunkIndex,
			})

			if len(results) >= config.Limit {
				break
			}
		}
	}

	// Sort results by score (highest first)
	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})

	return results, nil
}

func (m *InMemoryProvider) SearchAll(ctx context.Context, query string, options ...SearchOption) (*HybridResult, error) {
	start := time.Now()

	// Apply search options
	config := &SearchConfig{
		Limit:            m.config.MaxResults,
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
		personalResults, err := m.Query(ctx, query, config.Limit)
		if err != nil {
			return nil, fmt.Errorf("failed to search personal memory: %w", err)
		}
		result.PersonalMemory = personalResults
	}

	// Search knowledge base if enabled
	if config.IncludeKnowledge {
		knowledgeResults, err := m.SearchKnowledge(ctx, query, options...)
		if err != nil {
			return nil, fmt.Errorf("failed to search knowledge base: %w", err)
		}
		result.Knowledge = knowledgeResults
	}

	result.TotalResults = len(result.PersonalMemory) + len(result.Knowledge)
	result.SearchTime = time.Since(start)

	return result, nil
}

func (m *InMemoryProvider) BuildContext(ctx context.Context, query string, options ...ContextOption) (*RAGContext, error) {
	// Apply context options
	config := &ContextConfig{
		MaxTokens:       m.config.RAGMaxContextTokens,
		PersonalWeight:  m.config.RAGPersonalWeight,
		KnowledgeWeight: m.config.RAGKnowledgeWeight,
		HistoryLimit:    5,
		IncludeSources:  m.config.RAGIncludeSources,
		FormatTemplate:  "", // Use default formatting
	}
	for _, opt := range options {
		opt(config)
	}

	// Get hybrid search results
	searchResults, err := m.SearchAll(ctx, query,
		WithLimit(config.MaxTokens/100), // Rough estimate: 100 tokens per result
		WithIncludePersonal(config.PersonalWeight > 0),
		WithIncludeKnowledge(config.KnowledgeWeight > 0),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to search for context: %w", err)
	}

	// Get chat history
	history, err := m.GetHistory(ctx, config.HistoryLimit)
	if err != nil {
		return nil, fmt.Errorf("failed to get chat history: %w", err)
	}

	// Build context text
	contextText := m.formatContextText(query, searchResults, history, config)

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

// Helper function for InMemory context formatting
func (m *InMemoryProvider) formatContextText(query string, results *HybridResult, history []Message, config *ContextConfig) string {
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
