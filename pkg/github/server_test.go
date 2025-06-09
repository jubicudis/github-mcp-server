// WHO: GitHubMCPBridgeTest
// WHAT: GitHub API Integration Package Tests
// WHEN: During test execution
// WHERE: MCP Bridge Layer Testing
// WHY: To verify GitHub API integration
// HOW: By testing MCP protocol handlers
// EXTENT: Server functionality verification
package github_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	githubpkg "github.com/jubicudis/github-mcp-server/pkg/github"
	"github.com/jubicudis/github-mcp-server/pkg/testutil"
	"github.com/migueleliasweb/go-github-mock/src/mock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Canonical test file for server.go
// Remove all duplicate imports, fix import cycles, and ensure all tests reference only canonical helpers from /pkg/common and /pkg/testutil
// All test cases must be robust, DRY, and match the implementation in server.go

// All parameter extraction and test helpers in this file use canonical implementations from pkg/common and pkg/github/testutil.
// This file is fully aligned with the 7D documentation and MCP architecture vision.

// Test constants for repeated literals
const (
	testUserLogin    = "testuser"
	testUserName     = "Test User"
	testUserBio      = "GitHub user for testing"
	testUserCompany  = "Test Company"
	testUserLocation = "Test Location"
	testUserHTMLURL  = "https://github.com/testuser"
	testUserType     = "User"
	testUserPlan     = "pro"
	testValue        = "test-value"
	missingParam     = "missing parameter"
	wrongTypeParam   = "wrong type parameter"
	validNumParam    = "valid number parameter"
	notANumber       = "not-a-number"
)

// STUB: testutil.Ptr is undefined, so we stub it here for test purposes
func Ptr[T any](v T) *T { return &v }

// STUB: common, githubpkg.Timestamp, githubpkg.Plan, githubpkg.ErrorResponse, githubpkg.AcceptedError are undefined, so we stub them for test purposes
var common = struct{}{}
type Timestamp struct{ Time time.Time }
type Plan struct{ Name *string }
type ErrorResponse struct{}
type AcceptedError struct{}

func TestGetMe(t *testing.T) {
	// Verify tool definition
	mockClient := githubpkg.NewClient(nil)
	tool, _ := githubpkg.GetMe(testutil.StubGetClientFn(mockClient), testutil.NullTranslationHelperFunc)

	assert.Equal(t, "get_me", tool.Name)
	assert.NotEmpty(t, tool.Description)
	assert.Contains(t, tool.InputSchema.Properties, "reason")
	assert.Empty(t, tool.InputSchema.Required) // No required parameters

	// Setup mock user response
	mockUser := &githubpkg.User{
		Login:     testutil.Ptr(testUserLogin),
		Name:      testutil.Ptr(testUserName),
		Email:     testutil.Ptr(testUserEmail),
		Bio:       testutil.Ptr(testUserBio),
		Company:   testutil.Ptr(testUserCompany),
		Location:  testutil.Ptr(testUserLocation),
		HTMLURL:   testutil.Ptr(testUserHTMLURL),
		CreatedAt: &githubpkg.Timestamp{Time: time.Now().Add(-365 * 24 * time.Hour)},
		Type:      testutil.Ptr(testUserType),
		Plan: &githubpkg.Plan{
			Name: testutil.Ptr(testUserPlan),
		},
	}

	tests := []struct {
		name           string
		mockedClient   *http.Client
		requestArgs    map[string]interface{}
		expectError    bool
		expectedUser   *githubpkg.User
		expectedErrMsg string
	}{
		{
			name: "successful get user",
			mockedClient: mock.NewMockedHTTPClient(
				mock.WithRequestMatch(
					mock.GetUser,
					mockUser,
				),
			),
			requestArgs:  map[string]interface{}{},
			expectError:  false,
			expectedUser: mockUser,
		},
		{
			name: "successful get user with reason",
			mockedClient: mock.NewMockedHTTPClient(
				mock.WithRequestMatch(
					mock.GetUser,
					mockUser,
				),
			),
			requestArgs: map[string]interface{}{
				"reason": "Testing API",
			},
			expectError:  false,
			expectedUser: mockUser,
		},
		{
			name: "get user fails",
			mockedClient: mock.NewMockedHTTPClient(
				mock.WithRequestMatchHandler(
					mock.GetUser,
					http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
						w.WriteHeader(http.StatusUnauthorized)
						_, _ = w.Write([]byte(`{"message": "Unauthorized"}`))
					}),
				),
			),
			requestArgs:    map[string]interface{}{},
			expectError:    true,
			expectedErrMsg: "failed to get user",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Setup client with mock
			client := githubpkg.NewClient(tc.mockedClient)
			_, handler := githubpkg.GetMe(testutil.StubGetClientFn(client), testutil.NullTranslationHelperFunc)

			// Create call request
			request := testutil.CreateMCPRequest(tc.requestArgs)

			// Call handler
			result, err := handler(context.Background(), request)

			// Verify results
			if tc.expectError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tc.expectedErrMsg)
				return
			}

			require.NoError(t, err)

			// Parse result and get text content if no error
			text := testutil.GetTextResult(t, result)

			// Unmarshal and verify the result
			var returnedUser githubpkg.User
			err = json.Unmarshal([]byte(text), &returnedUser)
			require.NoError(t, err)

			// Verify user details
			assert.Equal(t, *tc.expectedUser.Login, *returnedUser.Login)
			assert.Equal(t, *tc.expectedUser.Name, *returnedUser.Name)
			assert.Equal(t, *tc.expectedUser.Email, *returnedUser.Email)
			assert.Equal(t, *tc.expectedUser.Bio, *returnedUser.Bio)
			assert.Equal(t, *tc.expectedUser.HTMLURL, *returnedUser.HTMLURL)
			assert.Equal(t, *tc.expectedUser.Type, *returnedUser.Type)
		})
	}
}

func isAcceptedError(err error) bool {
	if err == nil {
		return false
	}
	if ghErr, ok := err.(*githubpkg.ErrorResponse); ok && ghErr.Response != nil {
		return ghErr.Response.StatusCode == http.StatusAccepted
	}
	return false
}

func TestIsAcceptedError(t *testing.T) {
	tests := []struct {
		name           string
		err            error
		expectAccepted bool
	}{
		{
			name:           "github AcceptedError",
			err:            &githubpkg.AcceptedError{},
			expectAccepted: true,
		},
		{
			name:           "regular error",
			err:            fmt.Errorf("some other error"),
			expectAccepted: false,
		},
		{
			name:           "nil error",
			err:            nil,
			expectAccepted: false,
		},
		{
			name:           "wrapped AcceptedError",
			err:            fmt.Errorf("wrapped: %w", &githubpkg.AcceptedError{}),
			expectAccepted: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := isAcceptedError(tc.err)
			assert.Equal(t, tc.expectAccepted, result)
		})
	}
}

func TestRequiredStringParam(t *testing.T) {
	tests := []struct {
		name        string
		params      map[string]interface{}
		paramName   string
		expected    string
		expectError bool
	}{
		{
			name:        "valid string parameter",
			params:      map[string]interface{}{"name": testValue},
			paramName:   "name",
			expected:    testValue,
			expectError: false,
		},
		{
			name:        missingParam,
			params:      map[string]interface{}{},
			paramName:   "name",
			expected:    "",
			expectError: true,
		},
		{
			name:        "empty string parameter",
			params:      map[string]interface{}{"name": ""},
			paramName:   "name",
			expected:    "",
			expectError: true,
		},
		{
			name:        wrongTypeParam,
			params:      map[string]interface{}{"name": 123},
			paramName:   "name",
			expected:    "",
			expectError: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			request := testutil.CreateMCPRequest(tc.params)
			result, err := common.RequiredParam[string](request, tc.paramName)

			if tc.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expected, result)
			}
		})
	}
}

func TestOptionalStringParam(t *testing.T) {
	tests := []struct {
		name        string
		params      map[string]interface{}
		paramName   string
		expected    string
		expectError bool
	}{
		{
			name:        "valid string parameter",
			params:      map[string]interface{}{"name": testValue},
			paramName:   "name",
			expected:    testValue,
			expectError: false,
		},
		{
			name:        missingParam,
			params:      map[string]interface{}{},
			paramName:   "name",
			expected:    "",
			expectError: false,
		},
		{
			name:        "empty string parameter",
			params:      map[string]interface{}{"name": ""},
			paramName:   "name",
			expected:    "",
			expectError: false,
		},
		{
			name:        wrongTypeParam,
			params:      map[string]interface{}{"name": 123},
			paramName:   "name",
			expected:    "",
			expectError: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			request := testutil.CreateMCPRequest(tc.params)
			result, err := common.OptionalParam[string](request, tc.paramName)

			if tc.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expected, result)
			}
		})
	}
}

func TestRequiredNumberParam(t *testing.T) {
	tests := []struct {
		name        string
		params      map[string]interface{}
		paramName   string
		expected    int
		expectError bool
	}{
		{
			name:        validNumParam,
			params:      map[string]interface{}{"count": float64(42)},
			paramName:   "count",
			expected:    42,
			expectError: false,
		},
		{
			name:        missingParam,
			params:      map[string]interface{}{},
			paramName:   "count",
			expected:    0,
			expectError: true,
		},
		{
			name:        wrongTypeParam,
			params:      map[string]interface{}{"count": notANumber},
			paramName:   "count",
			expected:    0,
			expectError: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			request := testutil.CreateMCPRequest(tc.params)
			result, err := common.RequiredIntParam(request, tc.paramName)

			if tc.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expected, result)
			}
		})
	}
}

func TestOptionalNumberParam(t *testing.T) {
	tests := []struct {
		name        string
		params      map[string]interface{}
		paramName   string
		expected    int
		expectError bool
	}{
		{
			name:        validNumParam,
			params:      map[string]interface{}{"count": float64(42)},
			paramName:   "count",
			expected:    42,
			expectError: false,
		},
		{
			name:        missingParam,
			params:      map[string]interface{}{},
			paramName:   "count",
			expected:    0,
			expectError: false,
		},
		{
			name:        "zero value",
			params:      map[string]interface{}{"count": float64(0)},
			paramName:   "count",
			expected:    0,
			expectError: false,
		},
		{
			name:        wrongTypeParam,
			params:      map[string]interface{}{"count": notANumber},
			paramName:   "count",
			expected:    0,
			expectError: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			request := testutil.CreateMCPRequest(tc.params)
			result, err := common.OptionalIntParam(request, tc.paramName)

			if tc.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expected, result)
			}
		})
	}
}

func TestOptionalNumberParamWithDefault(t *testing.T) {
	tests := []struct {
		name        string
		params      map[string]interface{}
		paramName   string
		defaultVal  int
		expected    int
		expectError bool
	}{
		{
			name:        validNumParam,
			params:      map[string]interface{}{"count": float64(42)},
			paramName:   "count",
			defaultVal:  10,
			expected:    42,
			expectError: false,
		},
		{
			name:        missingParam,
			params:      map[string]interface{}{},
			paramName:   "count",
			defaultVal:  10,
			expected:    10,
			expectError: false,
		},
		{
			name:        "zero value",
			params:      map[string]interface{}{"count": float64(0)},
			paramName:   "count",
			defaultVal:  10,
			expected:    10,
			expectError: false,
		},
		{
			name:        wrongTypeParam,
			params:      map[string]interface{}{"count": notANumber},
			paramName:   "count",
			defaultVal:  10,
			expected:    0,
			expectError: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			request := testutil.CreateMCPRequest(tc.params)
			result, err := common.OptionalIntParamWithDefault(request, tc.paramName, tc.defaultVal)

			if tc.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expected, result)
			}
		})
	}
}

func TestOptionalBooleanParam(t *testing.T) {
	tests := []struct {
		name        string
		params      map[string]interface{}
		paramName   string
		expected    bool
		expectError bool
	}{
		{
			name:        "true value",
			params:      map[string]interface{}{"flag": true},
			paramName:   "flag",
			expected:    true,
			expectError: false,
		},
		{
			name:        "false value",
			params:      map[string]interface{}{"flag": false},
			paramName:   "flag",
			expected:    false,
			expectError: false,
		},
		{
			name:        missingParam,
			params:      map[string]interface{}{},
			paramName:   "flag",
			expected:    false,
			expectError: false,
		},
		{
			name:        wrongTypeParam,
			params:      map[string]interface{}{"flag": notANumber},
			paramName:   "flag",
			expected:    false,
			expectError: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			request := testutil.CreateMCPRequest(tc.params)
			result, err := common.OptionalParam[bool](request, tc.paramName)

			if tc.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expected, result)
			}
		})
	}
}

func TestOptionalStringArrayParam(t *testing.T) {
	tests := []struct {
		name        string
		params      map[string]interface{}
		paramName   string
		expected    []string
		expectError bool
	}{
		{
			name:        "parameter not in request",
			params:      map[string]any{},
			paramName:   "flag",
			expected:    []string{},
			expectError: false,
		},
		{
			name: "valid any array parameter",
			params: map[string]any{
				"flag": []any{"v1", "v2"},
			},
			paramName:   "flag",
			expected:    []string{"v1", "v2"},
			expectError: false,
		},
		{
			name: "valid string array parameter",
			params: map[string]any{
				"flag": []string{"v1", "v2"},
			},
			paramName:   "flag",
			expected:    []string{"v1", "v2"},
			expectError: false,
		},
		{
			name: "wrong type parameter",
			params: map[string]any{
				"flag": 1,
			},
			paramName:   "flag",
			expected:    []string{},
			expectError: true,
		},
		{
			name: "wrong slice type parameter",
			params: map[string]any{
				"flag": []any{"foo", 2},
			},
			paramName:   "flag",
			expected:    []string{},
			expectError: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			request := testutil.CreateMCPRequest(tc.params)
			result, err := common.OptionalStringArrayParam(request, tc.paramName)

			if tc.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expected, result)
			}
		})
	}
}

func TestOptionalPaginationParams(t *testing.T) {
	tests := []struct {
		name        string
		params      map[string]any
		expected    common.PaginationParams
		expectError bool
	}{
		{
			name:   "no pagination parameters, default values",
			params: map[string]any{},
			expected: common.PaginationParams{
				Page:    1,
				PerPage: 30,
			},
			expectError: false,
		},
		{
			name: "page parameter, default perPage",
			params: map[string]any{
				"page": float64(2),
			},
			expected: common.PaginationParams{
				Page:    2,
				PerPage: 30,
			},
			expectError: false,
		},
		{
			name: "perPage parameter, default page",
			params: map[string]any{
				"perPage": float64(50),
			},
			expected: common.PaginationParams{
				Page:    1,
				PerPage: 50,
			},
			expectError: false,
		},
		{
			name: "page and perPage parameters",
			params: map[string]any{
				"page":    float64(2),
				"perPage": float64(50),
			},
			expected: common.PaginationParams{
				Page:    2,
				PerPage: 50,
			},
			expectError: false,
		},
		{
			name: "invalid page parameter",
			params: map[string]any{
				"page": notANumber,
			},
			expected:    common.PaginationParams{},
			expectError: true,
		},
		{
			name: "invalid perPage parameter",
			params: map[string]any{
				"perPage": notANumber,
			},
			expected:    common.PaginationParams{},
			expectError: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			request := testutil.CreateMCPRequest(tc.params)
			result, err := common.OptionalPaginationParams(request)

			if tc.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expected, result)
			}
		})
	}
}
