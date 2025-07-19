//go:build integration

package tools

import (
	"context"
	"os"
	"strings"
	"testing"
)

func TestWebSearchTool_Integration(t *testing.T) {
	if os.Getenv("BRAVE_API_KEY") == "" {
		t.Skip("Skipping integration test: BRAVE_API_KEY not set")
	}

	t.Run("Live API call with valid query", func(t *testing.T) {
		tool, err := NewWebSearchTool()
		if err != nil {
			t.Fatalf("Failed to create tool: %v", err)
		}

		ctx := context.Background()
		result, err := tool.Call(ctx, map[string]any{"query": "what is the capital of France"})
		if err != nil {
			t.Fatalf("Expected no error on live API call, got: %v", err)
		}

		results, ok := result["results"].([]string)
		if !ok {
			t.Fatalf("Expected results to be []string, got %T", result["results"])
		}

		if len(results) == 0 {
			t.Fatal("Expected at least one result, but got none")
		}

		// Check that the results are meaningful
		if strings.Contains(results[0], "No results found.") {
			t.Errorf("Expected real search results, but got 'No results found.'")
		}

		// Check for expected content
		found := false
		for _, r := range results {
			if strings.Contains(strings.ToLower(r), "paris") {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected at least one result to mention 'Paris', got: %v", results)
		}
	})
}
