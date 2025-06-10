// WHO: GitHubMCPTestUtilities
// WHAT: Test Helper Functions
// WHEN: During test execution
// WHERE: MCP Bridge Layer Testing
// WHY: To provide reusable test utilities
// HOW: By implementing common test functions
// EXTENT: All test files
package ghmcp_test

// Canonical TNOS test imports: no duplicates, only what is used
import (
	"testing"

	"github.com/jubicudis/github-mcp-server/pkg/common"
	"github.com/jubicudis/github-mcp-server/pkg/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test constants
const (
	testNameParamNotPresent = "parameter not present"
)

// WHO: TestUtility
// WHAT: Helper function to check parameter validation
// WHEN: During test execution
// WHERE: Test context for OptionalParamOK
// WHY: To reduce cognitive complexity
// HOW: By extracting common test logic
// EXTENT: All parameter validation tests
func testTypedParam[T any](t *testing.T, args map[string]interface{},
	paramName string, expectedVal T, expectedOk bool,
	expectError bool, errorMsg string) {

	t.Helper()
	request := testutil.CreateMCPRequest(args)
	val, ok, err := common.OptionalParamOK[T](request, paramName)

	if expectError {
		require.Error(t, err)
		assert.Contains(t, err.Error(), errorMsg)
	} else {
		require.NoError(t, err)
	}
	assert.Equal(t, expectedOk, ok)
	assert.Equal(t, expectedVal, val)
}

// Test string parameters
func TestOptionalParamOKString(t *testing.T) {
	tests := []struct {
		name        string
		args        map[string]interface{}
		paramName   string
		expectedVal string
		expectedOk  bool
		expectError bool
		errorMsg    string
	}{
		{
			name:        "present and correct type (string)",
			args:        map[string]interface{}{"myParam": "hello"},
			paramName:   "myParam",
			expectedVal: "hello",
			expectedOk:  true,
			expectError: false,
		},
		{
			name:        "present but wrong type (string expected, got bool)",
			args:        map[string]interface{}{"myParam": true},
			paramName:   "myParam",
			expectedVal: "",
			expectedOk:  false,
			expectError: false,
		},
		{
			name:        testNameParamNotPresent,
			args:        map[string]interface{}{"anotherParam": "value"},
			paramName:   "myParam",
			expectedVal: "",
			expectedOk:  false,
			expectError: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			testTypedParam(t, tc.args, tc.paramName, tc.expectedVal, tc.expectedOk, tc.expectError, tc.errorMsg)
		})
	}
}

// Test bool parameters
func TestOptionalParamOKBool(t *testing.T) {
	tests := []struct {
		name        string
		args        map[string]interface{}
		paramName   string
		expectedVal bool
		expectedOk  bool
		expectError bool
		errorMsg    string
	}{
		{
			name:        "present and correct type (bool)",
			args:        map[string]interface{}{"myParam": true},
			paramName:   "myParam",
			expectedVal: true,
			expectedOk:  true,
			expectError: false,
		},
		{
			name:        "present but wrong type (bool expected, got string)",
			args:        map[string]interface{}{"myParam": "true"},
			paramName:   "myParam",
			expectedVal: false,
			expectedOk:  false,
			expectError: false,
		},
		{
			name:        testNameParamNotPresent,
			args:        map[string]interface{}{"anotherParam": "value"},
			paramName:   "myParam",
			expectedVal: false,
			expectedOk:  false,
			expectError: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			testTypedParam(t, tc.args, tc.paramName, tc.expectedVal, tc.expectedOk, tc.expectError, tc.errorMsg)
		})
	}
}

// Test float64 parameters
func TestOptionalParamOKFloat64(t *testing.T) {
	tests := []struct {
		name        string
		args        map[string]interface{}
		paramName   string
		expectedVal float64
		expectedOk  bool
		expectError bool
		errorMsg    string
	}{
		{
			name:        "present and correct type (number)",
			args:        map[string]interface{}{"myParam": float64(123)},
			paramName:   "myParam",
			expectedVal: 123,
			expectedOk:  true,
			expectError: false,
		},
		{
			name:        testNameParamNotPresent,
			args:        map[string]interface{}{"anotherParam": "value"},
			paramName:   "myParam",
			expectedVal: 0,
			expectedOk:  false,
			expectError: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			testTypedParam(t, tc.args, tc.paramName, tc.expectedVal, tc.expectedOk, tc.expectError, tc.errorMsg)
		})
	}
}

// Canonical test file for helper.go (if present)
// All tests must directly and robustly test the canonical logic in helper.go
// Remove all legacy, duplicate, or non-canonical tests
// Reference only helpers from /pkg/common and /pkg/testutil
// No import cycles, duplicate imports, or undefined helpers
// All test cases must match the actual signatures and logic of helper.go

// 7D Contextual Logging Stub (for future Helical Memory integration)
func log7DContext(t *testing.T, who, what, when, where, why, how, extent string) {
	// TODO: Integrate with Helical Memory system for 7D contextual logging
	// Example: helicalMemory.Log(who, what, when, where, why, how, extent)
}

// (File deprecated: all helpers moved to pkg/testutil/test_helpers.go)
// This file is intentionally left empty to avoid import cycles.
