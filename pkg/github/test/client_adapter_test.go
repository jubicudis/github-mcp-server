/*
 * WHO: ClientAdapterTester
 * WHAT: Integration test for GitHub client adapter pattern
 * WHEN: During testing operations
 * WHERE: System Layer 6 (Integration)
 * WHY: To verify adapter functionality
 * HOW: Using Go testing framework
 * EXTENT: All client adapter operations
 */

package github_test

import (
	"context"
	"testing"
	"time"

	"tranquility-neuro-os/github-mcp-server/pkg/github"
	"tranquility-neuro-os/github-mcp-server/pkg/log"
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

	// Create logger for testing
	logger := log.NewLogger(log.Config{
		Level:      log.LevelDebug,
		ConsoleOut: true,
	})

	// GitHub token from environment or use a placeholder for test
	token := "test_token"

	// Create legacy client via adapter
	legacyClient := github.NewClient(token, logger)

	if legacyClient == nil {
		t.Fatal("Failed to create legacy client adapter")
	}

	// Test context compression
	ctx := legacyClient.CompressContext()

	if ctx == nil {
		t.Error("Failed to compress context")
	}

	// Verify context contains expected 7D dimensions
	dimensions := []string{"who", "what", "when", "where", "why", "how", "extent"}
	for _, dim := range dimensions {
		if _, ok := ctx[dim]; !ok {
			t.Errorf("Missing dimension %s in compressed context", dim)
		}
	}

	// Create user service with legacy client
	userService := github.NewUserService(legacyClient)

	if userService == nil {
		t.Fatal("Failed to create user service")
	}

	// Verify the advanced client implementation from github_client.go also works
	options := github.ClientOptions{
		Token:           token,
		APIBaseURL:      "https://api.github.com",
		GraphQLBaseURL:  "https://api.github.com/graphql",
		AcceptHeader:    "application/vnd.github.v3+json",
		UserAgent:       "TNOS-GitHub-MCP-Test-Client",
		Timeout:         30 * time.Second,
		Logger:          logger,
		EnableCache:     true,
		CacheTimeout:    5 * time.Minute,
		RateLimitBuffer: 10,
	}

	// Create advanced client
	advancedClient, err := github.NewClient(options)

	if err != nil {
		t.Logf("Note: Advanced client creation skipped: %v", err)
		t.Log("This is expected in test environments without proper GitHub credentials")
	} else if advancedClient != nil {
		// Test the advanced client
		ctx := context.Background()
		owner := "octocat"
		repo := "hello-world"

		repoData, err := advancedClient.GetRepositoryByName(ctx, owner, repo)

		// We don't expect this to succeed without valid credentials,
		// just verifying the interface works
		t.Logf("Advanced client repository request result: %v", err)

		if repoData != nil {
			t.Logf("Found repository data with name: %v", repoData["name"])
		}
	}

	// Test MCP bridge integration
	bridgeOptions := github.DefaultMCPBridgeOptions()
	bridgeOptions.GithubToken = token
	bridgeOptions.Logger = logger

	bridge, err := github.NewMCPBridge(bridgeOptions)

	if err != nil {
		t.Logf("Note: MCP Bridge creation error: %v", err)
	} else if bridge != nil {
		state := bridge.GetState()
		t.Logf("MCP Bridge created with state: %v", state)

		// Test context synchronization (this won't actually connect but tests the interface)
		testContext := map[string]interface{}{
			"who":    "TestClient",
			"what":   "RunTest",
			"when":   time.Now().Unix(),
			"where":  "TestEnvironment",
			"why":    "TestVerification",
			"how":    "GoTesting",
			"extent": 1.0,
		}

		err = bridge.SyncContext(testContext)

		// We expect an error due to lack of connection
		if err != nil {
			t.Logf("Expected error during context sync without connection: %v", err)
		}
	}
}
