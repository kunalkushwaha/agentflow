// Package core provides configuration loading for AgentFlow.
package core

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/BurntSushi/toml"
)

// Config represents the AgentFlow configuration structure
type Config struct {
	AgentFlow struct {
		Name     string `toml:"name"`
		Version  string `toml:"version"`
		Provider string `toml:"provider"`
	} `toml:"agent_flow"`

	Logging struct {
		Level  string `toml:"level"`
		Format string `toml:"format"`
	} `toml:"logging"`

	Runtime struct {
		MaxConcurrentAgents int `toml:"max_concurrent_agents"`
		TimeoutSeconds      int `toml:"timeout_seconds"`
	} `toml:"runtime"`

	// Breaking change: Agent memory configuration added
	AgentMemory AgentMemoryConfig `toml:"agent_memory"`

	// Error routing configuration
	ErrorRouting struct {
		Enabled              bool                     `toml:"enabled"`
		MaxRetries           int                      `toml:"max_retries"`
		RetryDelayMs         int                      `toml:"retry_delay_ms"`
		EnableCircuitBreaker bool                     `toml:"enable_circuit_breaker"`
		ErrorHandlerName     string                   `toml:"error_handler_name"`
		CategoryHandlers     map[string]string        `toml:"category_handlers"`
		SeverityHandlers     map[string]string        `toml:"severity_handlers"`
		CircuitBreaker       CircuitBreakerConfigToml `toml:"circuit_breaker"`
		Retry                RetryConfigToml          `toml:"retry"`
	} `toml:"error_routing"`
	Providers map[string]map[string]interface{} `toml:"providers"`

	// MCP configuration
	MCP MCPConfigToml `toml:"mcp"`
}

// MemoryConfig represents memory configuration in TOML
type MemoryConfig struct {
	Limit      string `toml:"limit"`
	Swap       string `toml:"swap"`
	Disable    bool   `toml:"disable"`
	Overcommit bool   `toml:"overcommit"`
}

// CircuitBreakerConfigToml represents circuit breaker configuration in TOML
type CircuitBreakerConfigToml struct {
	FailureThreshold int `toml:"failure_threshold"`
	SuccessThreshold int `toml:"success_threshold"`
	TimeoutMs        int `toml:"timeout_ms"`
	ResetTimeoutMs   int `toml:"reset_timeout_ms"`
	HalfOpenMaxCalls int `toml:"half_open_max_calls"`
}

// RetryConfigToml represents retry configuration in TOML
type RetryConfigToml struct {
	MaxRetries    int     `toml:"max_retries"`
	BaseDelayMs   int     `toml:"base_delay_ms"`
	MaxDelayMs    int     `toml:"max_delay_ms"`
	BackoffFactor float64 `toml:"backoff_factor"`
	EnableJitter  bool    `toml:"enable_jitter"`
}

// MCPConfigToml represents MCP configuration in TOML format
type MCPConfigToml struct {
	Enabled           bool                  `toml:"enabled"`
	EnableDiscovery   bool                  `toml:"enable_discovery"`
	DiscoveryTimeout  int                   `toml:"discovery_timeout_ms"`
	ScanPorts         []int                 `toml:"scan_ports"`
	ConnectionTimeout int                   `toml:"connection_timeout_ms"`
	MaxRetries        int                   `toml:"max_retries"`
	RetryDelay        int                   `toml:"retry_delay_ms"`
	EnableCaching     bool                  `toml:"enable_caching"`
	CacheTimeout      int                   `toml:"cache_timeout_ms"`
	MaxConnections    int                   `toml:"max_connections"`
	Servers           []MCPServerConfigToml `toml:"servers"`
}

// MCPServerConfigToml represents individual MCP server configuration in TOML
type MCPServerConfigToml struct {
	Name    string `toml:"name"`
	Type    string `toml:"type"` // tcp, stdio, docker, websocket
	Host    string `toml:"host,omitempty"`
	Port    int    `toml:"port,omitempty"`
	Command string `toml:"command,omitempty"` // for stdio transport
	Enabled bool   `toml:"enabled"`
}

// LoadConfig loads configuration from the specified TOML file path
func LoadConfig(path string) (*Config, error) {
	// Check if file exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil, fmt.Errorf("configuration file not found: %s", path)
	}

	// Read the file
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read configuration file %s: %w", path, err)
	}

	// Parse TOML
	var config Config
	if err := toml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse TOML configuration: %w", err)
	}

	// Set defaults if not specified
	if config.Logging.Level == "" {
		config.Logging.Level = "info"
	}
	if config.Logging.Format == "" {
		config.Logging.Format = "json"
	}
	if config.Runtime.MaxConcurrentAgents == 0 {
		config.Runtime.MaxConcurrentAgents = 10
	}
	if config.Runtime.TimeoutSeconds == 0 {
		config.Runtime.TimeoutSeconds = 30
	}

	// Set MCP defaults if not specified
	if !config.MCP.Enabled {
		// If no MCP config provided, set reasonable defaults but keep disabled
		config.MCP.Enabled = false
	}
	if config.MCP.DiscoveryTimeout == 0 {
		config.MCP.DiscoveryTimeout = 10000 // 10 seconds in ms
	}
	if config.MCP.ConnectionTimeout == 0 {
		config.MCP.ConnectionTimeout = 30000 // 30 seconds in ms
	}
	if config.MCP.MaxRetries == 0 {
		config.MCP.MaxRetries = 3
	}
	if config.MCP.RetryDelay == 0 {
		config.MCP.RetryDelay = 1000 // 1 second in ms
	}
	if config.MCP.CacheTimeout == 0 {
		config.MCP.CacheTimeout = 300000 // 5 minutes in ms
	}
	if config.MCP.MaxConnections == 0 {
		config.MCP.MaxConnections = 10
	}

	// Set agent memory defaults if not specified
	if config.AgentMemory.Provider == "" {
		config.AgentMemory.Provider = "memory" // Default to in-memory for simplicity
	}
	if config.AgentMemory.Connection == "" {
		config.AgentMemory.Connection = "memory"
	}
	if config.AgentMemory.MaxResults == 0 {
		config.AgentMemory.MaxResults = 10
	}
	if config.AgentMemory.Dimensions == 0 {
		config.AgentMemory.Dimensions = 1536
	}
	// AutoEmbed defaults to true
	config.AgentMemory.AutoEmbed = true

	// Set RAG defaults if not specified
	if config.AgentMemory.KnowledgeMaxResults == 0 {
		config.AgentMemory.KnowledgeMaxResults = 20
	}
	if config.AgentMemory.KnowledgeScoreThreshold == 0 {
		config.AgentMemory.KnowledgeScoreThreshold = 0.7
	}
	if config.AgentMemory.ChunkSize == 0 {
		config.AgentMemory.ChunkSize = 1000
	}
	if config.AgentMemory.ChunkOverlap == 0 {
		config.AgentMemory.ChunkOverlap = 200
	}
	if config.AgentMemory.RAGMaxContextTokens == 0 {
		config.AgentMemory.RAGMaxContextTokens = 4000
	}
	if config.AgentMemory.RAGPersonalWeight == 0 {
		config.AgentMemory.RAGPersonalWeight = 0.3
	}
	if config.AgentMemory.RAGKnowledgeWeight == 0 {
		config.AgentMemory.RAGKnowledgeWeight = 0.7
	}

	// Set document processing defaults
	if len(config.AgentMemory.Documents.SupportedTypes) == 0 {
		config.AgentMemory.Documents.SupportedTypes = []string{"pdf", "txt", "md", "web", "code"}
	}
	if config.AgentMemory.Documents.MaxFileSize == "" {
		config.AgentMemory.Documents.MaxFileSize = "10MB"
	}

	// Set embedding service defaults
	if config.AgentMemory.Embedding.Provider == "" {
		config.AgentMemory.Embedding.Provider = "azure"
	}
	if config.AgentMemory.Embedding.Model == "" {
		config.AgentMemory.Embedding.Model = "text-embedding-ada-002"
	}
	if config.AgentMemory.Embedding.MaxBatchSize == 0 {
		config.AgentMemory.Embedding.MaxBatchSize = 100
	}
	if config.AgentMemory.Embedding.TimeoutSeconds == 0 {
		config.AgentMemory.Embedding.TimeoutSeconds = 30
	}

	// Set search defaults
	if config.AgentMemory.Search.KeywordWeight == 0 {
		config.AgentMemory.Search.KeywordWeight = 0.3
	}
	if config.AgentMemory.Search.SemanticWeight == 0 {
		config.AgentMemory.Search.SemanticWeight = 0.7
	}

	// Set boolean defaults (these are false by default in Go)
	if !config.AgentMemory.EnableKnowledgeBase {
		config.AgentMemory.EnableKnowledgeBase = true
	}
	if !config.AgentMemory.EnableRAG {
		config.AgentMemory.EnableRAG = true
	}
	if !config.AgentMemory.RAGIncludeSources {
		config.AgentMemory.RAGIncludeSources = true
	}
	if !config.AgentMemory.Documents.AutoChunk {
		config.AgentMemory.Documents.AutoChunk = true
	}
	if !config.AgentMemory.Documents.EnableMetadataExtraction {
		config.AgentMemory.Documents.EnableMetadataExtraction = true
	}
	if !config.AgentMemory.Documents.EnableURLScraping {
		config.AgentMemory.Documents.EnableURLScraping = true
	}
	if !config.AgentMemory.Embedding.CacheEmbeddings {
		config.AgentMemory.Embedding.CacheEmbeddings = true
	}
	if !config.AgentMemory.Search.HybridSearch {
		config.AgentMemory.Search.HybridSearch = true
	}

	return &config, nil
}

// LoadConfigFromWorkingDir loads agentflow.toml from the current working directory
func LoadConfigFromWorkingDir() (*Config, error) {
	wd, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("failed to get working directory: %w", err)
	}

	configPath := filepath.Join(wd, "agentflow.toml")
	return LoadConfig(configPath)
}

// InitializeProvider creates a ModelProvider based on the configuration
func (c *Config) InitializeProvider() (ModelProvider, error) {
	provider := c.AgentFlow.Provider
	if provider == "" {
		return nil, fmt.Errorf("no provider specified in configuration")
	}

	// Get provider-specific configuration
	providerConfig, exists := c.Providers[provider]
	if !exists {
		return nil, fmt.Errorf("no configuration found for provider: %s", provider)
	}

	switch provider {
	case "openai":
		return c.initializeOpenAIProvider(providerConfig)
	case "azure":
		return c.initializeAzureProvider(providerConfig)
	case "ollama":
		return c.initializeOllamaProvider(providerConfig)
	default:
		return nil, fmt.Errorf("unsupported provider: %s", provider)
	}
}

// initializeOpenAIProvider creates an OpenAI provider from configuration
func (c *Config) initializeOpenAIProvider(config map[string]interface{}) (ModelProvider, error) {
	// First try to get from config, then fall back to environment variables
	apiKey := c.getStringValue(config, "api_key")
	if apiKey == "" {
		apiKey = os.Getenv("OPENAI_API_KEY")
		if apiKey == "" {
			return nil, fmt.Errorf("OpenAI API key not found in configuration or OPENAI_API_KEY environment variable")
		}
	}

	model := c.getStringValue(config, "model")
	if model == "" {
		model = "gpt-4o"
	}

	maxTokens := c.getIntValue(config, "max_tokens")
	if maxTokens == 0 {
		maxTokens = 1000
	}

	temperature := c.getFloatValue(config, "temperature")
	if temperature == 0 {
		temperature = 0.7
	}

	return NewOpenAIAdapter(apiKey, model, maxTokens, float32(temperature))
}

// initializeAzureProvider creates an Azure OpenAI provider from configuration
func (c *Config) initializeAzureProvider(config map[string]interface{}) (ModelProvider, error) {
	// Try to get from config, then fall back to environment variables
	apiKey := c.getStringValue(config, "api_key")
	if apiKey == "" {
		apiKey = os.Getenv("AZURE_OPENAI_API_KEY")
		if apiKey == "" {
			return nil, fmt.Errorf("Azure OpenAI API key not found in configuration or AZURE_OPENAI_API_KEY environment variable")
		}
	}

	endpoint := c.getStringValue(config, "endpoint")
	if endpoint == "" {
		endpoint = os.Getenv("AZURE_OPENAI_ENDPOINT")
		if endpoint == "" {
			return nil, fmt.Errorf("Azure OpenAI endpoint not found in configuration or AZURE_OPENAI_ENDPOINT environment variable")
		}
	}

	chatDeployment := c.getStringValue(config, "chat_deployment")
	if chatDeployment == "" {
		chatDeployment = os.Getenv("AZURE_OPENAI_CHAT_DEPLOYMENT")
		if chatDeployment == "" {
			return nil, fmt.Errorf("Azure OpenAI chat deployment not found in configuration or AZURE_OPENAI_CHAT_DEPLOYMENT environment variable")
		}
	}

	embeddingDeployment := c.getStringValue(config, "embedding_deployment")
	if embeddingDeployment == "" {
		embeddingDeployment = os.Getenv("AZURE_OPENAI_EMBEDDING_DEPLOYMENT")
		if embeddingDeployment == "" {
			embeddingDeployment = "text-embedding-ada-002" // default
		}
	}

	return NewAzureOpenAIAdapter(AzureOpenAIAdapterOptions{
		Endpoint:            endpoint,
		APIKey:              apiKey,
		ChatDeployment:      chatDeployment,
		EmbeddingDeployment: embeddingDeployment,
	})
}

// initializeOllamaProvider creates an Ollama provider from configuration
func (c *Config) initializeOllamaProvider(config map[string]interface{}) (ModelProvider, error) {
	baseURL := c.getStringValue(config, "base_url")
	if baseURL == "" {
		baseURL = c.getStringValue(config, "endpoint") // alias support
	}
	if baseURL == "" {
		baseURL = os.Getenv("OLLAMA_BASE_URL")
		if baseURL == "" {
			baseURL = "http://localhost:11434"
		}
	}

	model := c.getStringValue(config, "model")
	if model == "" {
		model = os.Getenv("OLLAMA_MODEL")
		if model == "" {
			model = "llama3.2:latest"
		}
	}

	maxTokens := c.getIntValue(config, "max_tokens")
	if maxTokens == 0 {
		maxTokens = 1000
	}

	temperature := c.getFloatValue(config, "temperature")
	if temperature == 0 {
		temperature = 0.7
	}

	return NewOllamaAdapter(baseURL, model, maxTokens, float32(temperature))
}

// Helper methods to safely extract values from the configuration map
func (c *Config) getStringValue(config map[string]interface{}, key string) string {
	if val, exists := config[key]; exists {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return ""
}

func (c *Config) getIntValue(config map[string]interface{}, key string) int {
	if val, exists := config[key]; exists {
		switch v := val.(type) {
		case int:
			return v
		case int64:
			return int(v)
		case float64:
			return int(v)
		}
	}
	return 0
}

func (c *Config) getFloatValue(config map[string]interface{}, key string) float64 {
	if val, exists := config[key]; exists {
		switch v := val.(type) {
		case float64:
			return v
		case int:
			return float64(v)
		case int64:
			return float64(v)
		}
	}
	return 0
}

// GetLogLevel returns the logging level from configuration
func (c *Config) GetLogLevel() LogLevel {
	switch c.Logging.Level {
	case "debug":
		return DEBUG
	case "info":
		return INFO
	case "warn":
		return WARN
	case "error":
		return ERROR
	default:
		return INFO
	}
}

// ApplyLoggingConfig applies the logging configuration
func (c *Config) ApplyLoggingConfig() {
	SetLogLevel(c.GetLogLevel())
}

// GetErrorRoutingConfig converts TOML configuration to runtime ErrorRouterConfig
func (c *Config) GetErrorRoutingConfig() *ErrorRouterConfig {
	if !c.ErrorRouting.Enabled {
		return nil
	}

	config := &ErrorRouterConfig{
		MaxRetries:           c.ErrorRouting.MaxRetries,
		RetryDelayMs:         c.ErrorRouting.RetryDelayMs,
		EnableCircuitBreaker: c.ErrorRouting.EnableCircuitBreaker,
		ErrorHandlerName:     c.ErrorRouting.ErrorHandlerName,
		CategoryHandlers:     make(map[string]string),
		SeverityHandlers:     make(map[string]string),
	}

	// Set defaults if not specified
	if config.MaxRetries == 0 {
		config.MaxRetries = 3
	}
	if config.RetryDelayMs == 0 {
		config.RetryDelayMs = 1000
	}

	// Copy category handlers
	for category, handler := range c.ErrorRouting.CategoryHandlers {
		config.CategoryHandlers[category] = handler
	}

	// Copy severity handlers
	for severity, handler := range c.ErrorRouting.SeverityHandlers {
		config.SeverityHandlers[severity] = handler
	}

	return config
}

// GetCircuitBreakerConfig converts TOML configuration to runtime CircuitBreakerConfig
func (c *Config) GetCircuitBreakerConfig() *CircuitBreakerConfig {
	cb := &c.ErrorRouting.CircuitBreaker

	config := &CircuitBreakerConfig{
		FailureThreshold:   cb.FailureThreshold,
		SuccessThreshold:   cb.SuccessThreshold,
		Timeout:            time.Duration(cb.TimeoutMs) * time.Millisecond,
		MaxConcurrentCalls: cb.HalfOpenMaxCalls,
	}

	// Set defaults if not specified
	if config.FailureThreshold == 0 {
		config.FailureThreshold = 5
	}
	if config.SuccessThreshold == 0 {
		config.SuccessThreshold = 3
	}
	if config.Timeout == 0 {
		config.Timeout = 30 * time.Second
	}
	if config.MaxConcurrentCalls == 0 {
		config.MaxConcurrentCalls = 3
	}

	return config
}

// GetRetryConfig converts TOML configuration to runtime RetryPolicy
func (c *Config) GetRetryConfig() *RetryPolicy {
	r := &c.ErrorRouting.Retry

	config := &RetryPolicy{
		MaxRetries:    r.MaxRetries,
		InitialDelay:  time.Duration(r.BaseDelayMs) * time.Millisecond,
		MaxDelay:      time.Duration(r.MaxDelayMs) * time.Millisecond,
		BackoffFactor: r.BackoffFactor,
		Jitter:        r.EnableJitter,
	}

	// Set defaults if not specified
	if config.MaxRetries == 0 {
		config.MaxRetries = 3
	}
	if config.InitialDelay == 0 {
		config.InitialDelay = 1000 * time.Millisecond
	}
	if config.MaxDelay == 0 {
		config.MaxDelay = 30 * time.Second
	}
	if config.BackoffFactor == 0 {
		config.BackoffFactor = 2.0
	}

	return config
}

// ToMCPConfig converts MCPConfigToml to the runtime MCPConfig
func (c *MCPConfigToml) ToMCPConfig() MCPConfig {
	config := MCPConfig{
		EnableDiscovery:   c.EnableDiscovery,
		DiscoveryTimeout:  time.Duration(c.DiscoveryTimeout) * time.Millisecond,
		ScanPorts:         c.ScanPorts,
		ConnectionTimeout: time.Duration(c.ConnectionTimeout) * time.Millisecond,
		MaxRetries:        c.MaxRetries,
		RetryDelay:        time.Duration(c.RetryDelay) * time.Millisecond,
		EnableCaching:     c.EnableCaching,
		CacheTimeout:      time.Duration(c.CacheTimeout) * time.Millisecond,
		MaxConnections:    c.MaxConnections,
		Servers:           make([]MCPServerConfig, len(c.Servers)),
	}

	// Convert server configurations
	for i, server := range c.Servers {
		config.Servers[i] = MCPServerConfig{
			Name:    server.Name,
			Type:    server.Type,
			Host:    server.Host,
			Port:    server.Port,
			Command: server.Command,
			Enabled: server.Enabled,
		}
	}

	return config
}

// GetMCPConfig returns the MCP configuration from the main config
func (c *Config) GetMCPConfig() MCPConfig {
	return c.MCP.ToMCPConfig()
}
