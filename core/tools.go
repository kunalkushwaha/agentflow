// Package core provides the public ToolRegistry interface and related types for AgentFlow.
package core

import (
	"context"

	"github.com/kunalkushwaha/agentflow/internal/tools"
)

// ToolRegistry defines the interface for managing and accessing tools.
type ToolRegistry interface {
	// List returns the names of all registered tools.
	List() []string

	// CallTool looks up a tool by name and executes it with the given arguments.
	CallTool(ctx context.Context, name string, args map[string]any) (map[string]any, error)

	// HasTool checks if a tool with the given name is registered.
	HasTool(name string) bool
}

// ToolRegistryAdapter wraps an internal tools.ToolRegistry to implement the public ToolRegistry interface.
type ToolRegistryAdapter struct {
	registry *tools.ToolRegistry
}

// NewToolRegistryAdapter creates a new adapter for the internal tool registry.
func NewToolRegistryAdapter(registry *tools.ToolRegistry) *ToolRegistryAdapter {
	return &ToolRegistryAdapter{registry: registry}
}

// List returns the names of all registered tools.
func (a *ToolRegistryAdapter) List() []string {
	return a.registry.List()
}

// CallTool looks up a tool by name and executes it with the given arguments.
func (a *ToolRegistryAdapter) CallTool(ctx context.Context, name string, args map[string]any) (map[string]any, error) {
	return a.registry.CallTool(ctx, name, args)
}

// HasTool checks if a tool with the given name is registered.
func (a *ToolRegistryAdapter) HasTool(name string) bool {
	_, exists := a.registry.Get(name)
	return exists
}
