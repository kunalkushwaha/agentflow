package mcp

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// DefaultMCPServerManager implements MCPServerManager interface
type DefaultMCPServerManager struct {
	clients             map[string]MCPClient
	configs             map[string]ServerConfig
	healthStatus        map[string]HealthStatus
	clientFactory       MCPClientFactory
	autoDiscoveryTicker *time.Ticker
	autoDiscoveryCtx    context.Context
	autoDiscoveryCancel context.CancelFunc
	healthCheckInterval time.Duration
	healthTicker        *time.Ticker
	eventHandler        ServerEventHandler
	mu                  sync.RWMutex
}

// NewDefaultMCPServerManager creates a new default MCP server manager
func NewDefaultMCPServerManager(factory MCPClientFactory) *DefaultMCPServerManager {
	return &DefaultMCPServerManager{
		clients:             make(map[string]MCPClient),
		configs:             make(map[string]ServerConfig),
		healthStatus:        make(map[string]HealthStatus),
		clientFactory:       factory,
		healthCheckInterval: 30 * time.Second,
	}
}

// Server Management

func (m *DefaultMCPServerManager) AddServer(config ServerConfig) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Check if server already exists
	if _, exists := m.clients[config.ID]; exists {
		return fmt.Errorf("server %s already exists", config.ID)
	}

	// Create client using factory
	clientConfig := make(map[string]interface{})
	clientConfig["id"] = config.ID
	clientConfig["name"] = config.Name
	clientConfig["type"] = config.Type
	clientConfig["client_type"] = config.ClientType
	for k, v := range config.Config {
		clientConfig[k] = v
	}

	client, err := m.clientFactory.CreateClient(config.ClientType, clientConfig)
	if err != nil {
		return fmt.Errorf("failed to create client for server %s: %w", config.ID, err)
	}
	// Store the client and config
	m.clients[config.ID] = client
	m.configs[config.ID] = config
	m.healthStatus[config.ID] = HealthStatus{
		Status:    StatusDisconnected,
		LastCheck: time.Now(),
		Error:     "Not connected",
	}

	// Connect if enabled
	if config.Enabled {
		go m.connectServer(config.ID)
	}

	// Notify event handler
	if m.eventHandler != nil {
		m.eventHandler(ServerEvent{
			Type:      "server_added",
			ServerID:  config.ID,
			Data:      map[string]interface{}{"config": config},
			Timestamp: time.Now(),
		})
	}

	return nil
}

func (m *DefaultMCPServerManager) RemoveServer(serverID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	client, exists := m.clients[serverID]
	if !exists {
		return fmt.Errorf("server %s not found", serverID)
	}

	// Disconnect the client
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	client.Disconnect(ctx)

	// Remove from maps
	delete(m.clients, serverID)
	delete(m.configs, serverID)
	delete(m.healthStatus, serverID)

	// Notify event handler
	if m.eventHandler != nil {
		m.eventHandler(ServerEvent{
			Type:      "server_removed",
			ServerID:  serverID,
			Data:      map[string]interface{}{},
			Timestamp: time.Now(),
		})
	}

	return nil
}

func (m *DefaultMCPServerManager) GetServer(serverID string) (MCPClient, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	client, exists := m.clients[serverID]
	if !exists {
		return nil, fmt.Errorf("server %s not found", serverID)
	}

	return client, nil
}

func (m *DefaultMCPServerManager) ListServers() []ServerInfo {
	m.mu.RLock()
	defer m.mu.RUnlock()

	servers := make([]ServerInfo, 0, len(m.configs))
	for serverID, config := range m.configs {
		status := m.healthStatus[serverID]
		servers = append(servers, ServerInfo{
			ID:       serverID,
			Name:     config.Name,
			Version:  "1.0.0",
			Status:   status.Status,
			LastSeen: status.LastCheck,
			Capabilities: ServerCapabilities{
				Tools:     true,
				Resources: true,
				Prompts:   true,
				Logging:   true,
				Features:  []string{"agentflow"},
			},
			Metadata: map[string]string{
				"type": config.Type,
			},
		})
	}

	return servers
}

// Discovery

func (m *DefaultMCPServerManager) DiscoverServers(ctx context.Context) ([]ServerConfig, error) {
	// This is a placeholder implementation
	// In a real implementation, this would scan for available MCP servers
	// For example, by looking for MCP executables, checking configuration directories, etc.

	discovered := []ServerConfig{
		{
			ID:         "discovered-server-1",
			Name:       "Auto-discovered Server",
			Type:       "stdio",
			ClientType: "mark3labs",
			Connection: ConnectionConfig{
				Transport: "stdio",
				Command:   []string{"mcp-server-example"},
			},
			Enabled: false, // Don't auto-connect discovered servers
		},
	}

	return discovered, nil
}

func (m *DefaultMCPServerManager) StartAutoDiscovery(ctx context.Context, interval time.Duration) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.autoDiscoveryTicker != nil {
		return fmt.Errorf("auto-discovery already running")
	}

	m.autoDiscoveryCtx, m.autoDiscoveryCancel = context.WithCancel(ctx)
	m.autoDiscoveryTicker = time.NewTicker(interval)

	go m.runAutoDiscovery()

	return nil
}

func (m *DefaultMCPServerManager) StopAutoDiscovery() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.autoDiscoveryTicker == nil {
		return fmt.Errorf("auto-discovery not running")
	}

	m.autoDiscoveryTicker.Stop()
	m.autoDiscoveryCancel()
	m.autoDiscoveryTicker = nil
	m.autoDiscoveryCancel = nil

	return nil
}

// Health Monitoring

func (m *DefaultMCPServerManager) HealthCheck(ctx context.Context) map[string]HealthStatus {
	m.mu.RLock()
	defer m.mu.RUnlock()

	status := make(map[string]HealthStatus)
	for serverID := range m.clients {
		status[serverID] = m.checkServerHealth(ctx, serverID)
	}

	return status
}

func (m *DefaultMCPServerManager) SetHealthCheckInterval(interval time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.healthCheckInterval = interval

	// Restart health checking with new interval
	if m.healthTicker != nil {
		m.healthTicker.Stop()
		m.healthTicker = time.NewTicker(interval)
		go m.runHealthChecking()
	}
}

// Event Handling

func (m *DefaultMCPServerManager) SetServerEventHandler(handler ServerEventHandler) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.eventHandler = handler
}

// Private methods

func (m *DefaultMCPServerManager) connectServer(serverID string) {
	m.mu.RLock()
	client := m.clients[serverID]
	config := m.configs[serverID]
	m.mu.RUnlock()

	if client == nil {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), config.Timeout)
	defer cancel()

	startTime := time.Now()
	err := client.Connect(ctx, config)
	responseTime := time.Since(startTime)
	m.mu.Lock()
	if err != nil {
		m.healthStatus[serverID] = HealthStatus{
			Status:       StatusError,
			LastCheck:    time.Now(),
			ResponseTime: responseTime,
			Error:        err.Error(),
		}
	} else {
		m.healthStatus[serverID] = HealthStatus{
			Status:       StatusConnected,
			LastCheck:    time.Now(),
			ResponseTime: responseTime,
		}
	}
	m.mu.Unlock()

	// Notify event handler
	if m.eventHandler != nil {
		eventType := "server_connected"
		if err != nil {
			eventType = "server_connection_failed"
		}

		m.eventHandler(ServerEvent{
			Type:     eventType,
			ServerID: serverID,
			Data: map[string]interface{}{
				"error":         err,
				"response_time": responseTime,
			},
			Timestamp: time.Now(),
		})
	}
}

func (m *DefaultMCPServerManager) checkServerHealth(ctx context.Context, serverID string) HealthStatus {
	client := m.clients[serverID]
	if client == nil {
		return HealthStatus{
			Status:    StatusDisconnected,
			LastCheck: time.Now(),
			Error:     "Client not found",
		}
	}

	startTime := time.Now()
	err := client.Ping(ctx)
	responseTime := time.Since(startTime)
	if err != nil {
		return HealthStatus{
			Status:       StatusError,
			LastCheck:    time.Now(),
			ResponseTime: responseTime,
			Error:        err.Error(),
		}
	}

	return HealthStatus{
		Status:       StatusConnected,
		LastCheck:    time.Now(),
		ResponseTime: responseTime,
	}
}

func (m *DefaultMCPServerManager) runAutoDiscovery() {
	for {
		select {
		case <-m.autoDiscoveryCtx.Done():
			return
		case <-m.autoDiscoveryTicker.C:
			discovered, err := m.DiscoverServers(m.autoDiscoveryCtx)
			if err != nil {
				continue
			}

			// Add any newly discovered servers
			for _, config := range discovered {
				m.mu.RLock()
				_, exists := m.clients[config.ID]
				m.mu.RUnlock()

				if !exists {
					// Only add if not already managed
					m.AddServer(config)
				}
			}
		}
	}
}

func (m *DefaultMCPServerManager) runHealthChecking() {
	for {
		m.mu.RLock()
		ticker := m.healthTicker
		m.mu.RUnlock()

		if ticker == nil {
			return // Health checking was stopped
		}

		select {
		case <-ticker.C:
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			m.HealthCheck(ctx)
			cancel()
		}
	}
}

// StartHealthChecking starts automatic health checking
func (m *DefaultMCPServerManager) StartHealthChecking() {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.healthTicker != nil {
		return // Already running
	}

	m.healthTicker = time.NewTicker(m.healthCheckInterval)
	go m.runHealthChecking()
}

// StopHealthChecking stops automatic health checking
func (m *DefaultMCPServerManager) StopHealthChecking() {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.healthTicker != nil {
		m.healthTicker.Stop()
		m.healthTicker = nil
	}
}
