package core

import (
	"context"
	"time"
)

// Memory is the central memory interface - replaces multiple interfaces
// Breaking change: Single, unified interface for all memory operations INCLUDING RAG
type Memory interface {
	// Personal memory operations (existing)
	Store(ctx context.Context, content string, tags ...string) error
	Query(ctx context.Context, query string, limit ...int) ([]Result, error)
	Remember(ctx context.Context, key string, value any) error
	Recall(ctx context.Context, key string) (any, error)

	// Chat history management (existing)
	AddMessage(ctx context.Context, role, content string) error
	GetHistory(ctx context.Context, limit ...int) ([]Message, error)

	// Session management (existing)
	NewSession() string
	SetSession(ctx context.Context, sessionID string) context.Context
	ClearSession(ctx context.Context) error
	Close() error

	// NEW: RAG-Enhanced Knowledge Base Operations
	// Breaking change: Add RAG capabilities to core memory interface
	IngestDocument(ctx context.Context, doc Document) error
	IngestDocuments(ctx context.Context, docs []Document) error
	SearchKnowledge(ctx context.Context, query string, options ...SearchOption) ([]KnowledgeResult, error)

	// NEW: Hybrid Search (Personal Memory + Knowledge Base)
	SearchAll(ctx context.Context, query string, options ...SearchOption) (*HybridResult, error)

	// NEW: RAG Context Assembly for LLM Prompts
	BuildContext(ctx context.Context, query string, options ...ContextOption) (*RAGContext, error)
}

// Result - simplified result structure
type Result struct {
	Content   string    `json:"content"`
	Score     float32   `json:"score"`
	Tags      []string  `json:"tags,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

// Message - conversation message
type Message struct {
	Role      string    `json:"role"` // user, assistant, system
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
}

// NEW: RAG-Enhanced Types for Knowledge Base and Document Management

// Document structure for knowledge ingestion
type Document struct {
	ID         string         `json:"id"`
	Title      string         `json:"title,omitempty"`
	Content    string         `json:"content"`
	Source     string         `json:"source,omitempty"` // URL, file path, etc.
	Type       DocumentType   `json:"type,omitempty"`   // PDF, TXT, WEB, etc.
	Metadata   map[string]any `json:"metadata,omitempty"`
	Tags       []string       `json:"tags,omitempty"`
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at,omitempty"`
	ChunkIndex int            `json:"chunk_index,omitempty"` // For chunked documents
	ChunkTotal int            `json:"chunk_total,omitempty"`
}

// DocumentType represents the type of document being ingested
type DocumentType string

const (
	DocumentTypePDF      DocumentType = "pdf"
	DocumentTypeText     DocumentType = "txt"
	DocumentTypeMarkdown DocumentType = "md"
	DocumentTypeWeb      DocumentType = "web"
	DocumentTypeCode     DocumentType = "code"
	DocumentTypeJSON     DocumentType = "json"
)

// KnowledgeResult represents search results from the knowledge base
type KnowledgeResult struct {
	Content    string         `json:"content"`
	Score      float32        `json:"score"`
	Source     string         `json:"source"`
	Title      string         `json:"title,omitempty"`
	DocumentID string         `json:"document_id"`
	Metadata   map[string]any `json:"metadata,omitempty"`
	Tags       []string       `json:"tags,omitempty"`
	CreatedAt  time.Time      `json:"created_at"`
	ChunkIndex int            `json:"chunk_index,omitempty"`
}

// HybridResult combines personal memory and knowledge base search results
type HybridResult struct {
	PersonalMemory []Result          `json:"personal_memory"`
	Knowledge      []KnowledgeResult `json:"knowledge"`
	Query          string            `json:"query"`
	TotalResults   int               `json:"total_results"`
	SearchTime     time.Duration     `json:"search_time"`
}

// RAGContext provides assembled context for LLM prompts
type RAGContext struct {
	Query          string            `json:"query"`
	PersonalMemory []Result          `json:"personal_memory"`
	Knowledge      []KnowledgeResult `json:"knowledge"`
	ChatHistory    []Message         `json:"chat_history"`
	ContextText    string            `json:"context_text"` // Formatted for LLM
	Sources        []string          `json:"sources"`      // Source attribution
	TokenCount     int               `json:"token_count"`  // Estimated tokens
	Timestamp      time.Time         `json:"timestamp"`
}

// Search and context configuration options
type SearchOption func(*SearchConfig)
type ContextOption func(*ContextConfig)

type SearchConfig struct {
	Limit            int            `json:"limit"`
	ScoreThreshold   float32        `json:"score_threshold"`
	Sources          []string       `json:"sources"`           // Filter by source
	DocumentTypes    []DocumentType `json:"document_types"`    // Filter by type
	Tags             []string       `json:"tags"`              // Filter by tags
	DateRange        *DateRange     `json:"date_range"`        // Filter by date
	HybridWeight     float32        `json:"hybrid_weight"`     // Semantic vs keyword weight
	IncludePersonal  bool           `json:"include_personal"`  // Include personal memory
	IncludeKnowledge bool           `json:"include_knowledge"` // Include knowledge base
}

type ContextConfig struct {
	MaxTokens       int     `json:"max_tokens"`       // Context size limit
	PersonalWeight  float32 `json:"personal_weight"`  // Weight for personal memory
	KnowledgeWeight float32 `json:"knowledge_weight"` // Weight for knowledge base
	HistoryLimit    int     `json:"history_limit"`    // Chat history messages
	IncludeSources  bool    `json:"include_sources"`  // Include source attribution
	FormatTemplate  string  `json:"format_template"`  // Custom context formatting
}

type DateRange struct {
	Start time.Time `json:"start"`
	End   time.Time `json:"end"`
}
