/*
 * WHO: ClientAdapterTester
 * WHAT: Integration test for GitHub client adapter pattern
 * WHEN: During testing operations
 * WHERE: System Layer 6 (Integration)
 * WHY: To verify adapter functionality
 * HOW: Using Go testing framework
 * EXTENT: All client adapter operations
 */

// Change test package to 'github_test' to avoid import cycle
package github_test

import (
	"testing"

	"github.com/jubicudis/github-mcp-server/pkg/common"
	"github.com/jubicudis/github-mcp-server/pkg/github"
)

// TestClientAdapter verifies that the adapter pattern works correctly
func TestClientAdapter(t *testing.T) {
	// WHO: AdapterTester
	// WHAT: Test adapter functionality
	// WHEN: During test execution
	// WHERE: System Layer 6 (Integration)
	// WHY: To validate implementation
	// HOW: Using test assertions
	// EXTENT: Adapter verification

	// GitHub token from environment or use a placeholder for test
	token := "test_token"

	// Canonical: Use ConnectionOptions struct from pkg/common
	opts := common.ConnectionOptions{Credentials: map[string]string{"token": token}}
	// Create legacy client via canonical adapter
	legacyClient := github.NewClientCompatibilityAdapter(opts)

	if legacyClient == nil {
		t.Fatal("Failed to create legacy client adapter")
	}

	// Test context data
	ctx := legacyClient.GetContext()

	if ctx.Who == "" {
		t.Error("Missing WHO dimension in context")
	}

	// Verify context contains expected 7D dimensions
	if ctx.What == "" {
		t.Error("Missing WHAT dimension in context")
	}
	if ctx.When == 0 {
		t.Error("Missing WHEN dimension in context")
	}
	if ctx.Where == "" {
		t.Error("Missing WHERE dimension in context")
	}
	if ctx.Why == "" {
		t.Error("Missing WHY dimension in context")
	}
	if ctx.How == "" {
		t.Error("Missing HOW dimension in context")
	}
	if ctx.Extent == 0 {
		t.Error("Missing EXTENT dimension in context")
	}

	// Test getting a user
	user, err := legacyClient.GetUser("testuser")
	if err != nil {
		t.Errorf("Failed to get user: %v", err)
	}
	if user == nil {
		t.Error("Expected user object, got nil")
	} else if user.Login == nil || *user.Login != "testuser" {
		t.Errorf("Expected username 'testuser', got '%v'", user.Login)
	}

	// Test getting a repository
	repo, err := legacyClient.GetRepository("testowner", "testrepo")
	if err != nil {
		t.Errorf("Failed to get repository: %v", err)
	}
	if repo == nil {
		t.Error("Expected repository object, got nil")
	} else if repo.Name == nil || *repo.Name != "testrepo" {
		t.Errorf("Expected repository name 'testrepo', got '%v'", repo.Name)
	}

	// Test parsing a resource URI
	uri := "repo://testowner/testrepo/contents/path/to/file"
	resourceURI, err := legacyClient.ParseResourceURI(uri)
	if err != nil {
		t.Errorf("Failed to parse resource URI: %v", err)
	}
	if resourceURI == nil {
		t.Error("Expected resourceURI object, got nil")
	} else {
		if resourceURI.Owner != "testowner" {
			t.Errorf("Expected owner 'testowner', got '%s'", resourceURI.Owner)
		}
		if resourceURI.Repo != "testrepo" {
			t.Errorf("Expected repo 'testrepo', got '%s'", resourceURI.Repo)
		}
		if resourceURI.Path != "path/to/file" {
			t.Errorf("Expected path 'path/to/file', got '%s'", resourceURI.Path)
		}
	}

	// Test Möbius compression
	testData := map[string]interface{}{
		"key1": "value1",
		"key2": 123,
		"key3": true,
		"key4": []string{"a", "b", "c"},
	}

	compressed, err := legacyClient.ApplyMobiusCompression(testData)
	if err != nil {
		t.Errorf("Failed to apply Möbius compression: %v", err)
	}
	if compressed == nil {
		t.Error("Expected compressed object, got nil")
	}

	// Test creating new context
	newContext := legacyClient.CreateContext("TestOperation", "Testing", 0.5)
	if newContext.What != "TestOperation" {
		t.Errorf("Expected What='TestOperation', got '%s'", newContext.What)
	}
	if newContext.Why != "Testing" {
		t.Errorf("Expected Why='Testing', got '%s'", newContext.Why)
	}
	if newContext.Extent != 0.5 {
		t.Errorf("Expected Extent=0.5, got '%f'", newContext.Extent)
	}

	// Test creating client with context
	clientWithContext := legacyClient.WithContext(newContext)
	if clientWithContext == nil {
		t.Error("Failed to create client with context")
	}

	updatedContext := clientWithContext.GetContext()
	if updatedContext.What != "TestOperation" {
		t.Errorf("Context was not properly updated: expected What='TestOperation', got '%s'", updatedContext.What)
	}
}

// Canonical test file for ClientCompatibilityAdapter
// Remove all duplicate imports, fix import cycles, and ensure all tests reference only canonical helpers from /pkg/common and /pkg/testutil
// All test cases must be robust, DRY, and match the implementation in client_adapter.go
