// WHO: CommonsManager
// WHAT: Shared types and utilities
// WHEN: Throughout GitHub MCP operations
// WHERE: GitHub MCP Server
// WHY: To provide shared functionality across components
// HOW: By defining common types and helper functions
// EXTENT: All GitHub MCP components
package github

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/go-github/v69/github"
	"github.com/mark3labs/mcp-go/mcp"
)

// WHO: PointerHelper
// WHAT: Generic Pointer Helper
// WHEN: During parameter construction
// WHERE: GitHub MCP Server
// WHY: To create pointers to literal values
// HOW: By using Go's generic type system
// EXTENT: All pointer creation operations
func Ptr[T any](v T) *T {
	return &v
}

// StringTranslationFunc defines a function type for string translations
type StringTranslationFunc func(key string, defaultValue string) string

// WHO: PaginationManager
// WHAT: Pagination Parameter Structure
// WHEN: During parameter processing
// WHERE: GitHub MCP Server
// WHY: To standardize pagination parameter handling
// HOW: By grouping pagination parameters in a struct
// EXTENT: All paginated requests
type PaginationParams struct {
	page    int
	perPage int
}

// WHO: PRParamsExtractor
// WHAT: Pull Request Parameter Extraction
// WHEN: During PR tool invocation
// WHERE: GitHub MCP Server
// WHY: To standardize PR parameter handling
// HOW: By grouping common parameter extraction logic
// EXTENT: All PR operation tools
type prParams struct {
	owner  string
	repo   string
	number int
}

// WHO: ParameterHelper
// WHAT: Optional Parameter Extraction
// WHEN: During parameter processing
// WHERE: GitHub MCP Server
// WHY: To safely extract optional parameters
// HOW: By checking parameter existence and type
// EXTENT: All tool parameter processing
func OptionalParamOK[T any](r mcp.CallToolRequest, p string) (value T, ok bool, err error) {
	// Check if the parameter is present in the request
	val, exists := r.Params.Arguments[p]
	if !exists {
		// Not present, return zero value, false, no error
		return
	}

	// Check if the parameter is of the expected type
	value, ok = val.(T)
	if !ok {
		// Present but wrong type
		err = fmt.Errorf("parameter %s is not of type %T, is %T", p, value, val)
		ok = true // Set ok to true because the parameter *was* present, even if wrong type
		return
	}

	// Present and correct type
	ok = true
	return
}

// WHO: ErrorChecker
// WHAT: Error Acceptance Checking
// WHEN: During error handling
// WHERE: GitHub MCP Server
// WHY: To identify acceptable errors
// HOW: By checking error type
// EXTENT: All error processing
func isAcceptedError(err error) bool {
	var acceptedError *github.AcceptedError
	return errors.As(err, &acceptedError)
}

// WHO: ParameterHelper
// WHAT: Required Parameter Extraction
// WHEN: During parameter processing
// WHERE: GitHub MCP Server
// WHY: To safely extract required parameters
// HOW: By checking parameter existence, type, and value
// EXTENT: All tool parameter processing
func RequiredParam[T comparable](r mcp.CallToolRequest, p string) (T, error) {
	var zero T

	// Check if the parameter is present in the request
	if _, ok := r.Params.Arguments[p]; !ok {
		return zero, fmt.Errorf("missing required parameter: %s", p)
	}

	// Check if the parameter is of the expected type
	if _, ok := r.Params.Arguments[p].(T); !ok {
		return zero, fmt.Errorf("parameter %s is not of type %T", p, zero)
	}

	if r.Params.Arguments[p].(T) == zero {
		return zero, fmt.Errorf("missing required parameter: %s", p)
	}

	return r.Params.Arguments[p].(T), nil
}

// WHO: ParameterHelper
// WHAT: Required Integer Extraction
// WHEN: During parameter processing
// WHERE: GitHub MCP Server
// WHY: To safely extract required integer parameters
// HOW: By checking parameter existence, type, and value
// EXTENT: All integer parameter processing
func RequiredIntParam(r mcp.CallToolRequest, p string) (int, error) {
	v, err := RequiredParam[float64](r, p)
	if err != nil {
		return 0, err
	}
	return int(v), nil
}

// WHO: ParameterHelper
// WHAT: Optional Parameter Extraction
// WHEN: During parameter processing
// WHERE: GitHub MCP Server
// WHY: To safely extract optional parameters
// HOW: By checking parameter existence and type
// EXTENT: All tool parameter processing
func ExtractOptionalParam[T any](r mcp.CallToolRequest, p string) (T, error) {
	var zero T

	// Check if the parameter is present in the request
	if _, ok := r.Params.Arguments[p]; !ok {
		return zero, nil
	}

	// Check if the parameter is of the expected type
	if _, ok := r.Params.Arguments[p].(T); !ok {
		return zero, fmt.Errorf("parameter %s is not of type %T, is %T", p, zero, r.Params.Arguments[p])
	}

	return r.Params.Arguments[p].(T), nil
}

// WHO: ParameterHelper
// WHAT: Optional Integer Extraction
// WHEN: During parameter processing
// WHERE: GitHub MCP Server
// WHY: To safely extract optional integer parameters
// HOW: By checking parameter existence and type
// EXTENT: All integer parameter processing
func OptionalIntParam(r mcp.CallToolRequest, p string) (int, error) {
	v, err := ExtractOptionalParam[float64](r, p)
	if err != nil {
		return 0, err
	}
	return int(v), nil
}

// WHO: ParameterHelper
// WHAT: Optional Integer with Default
// WHEN: During parameter processing
// WHERE: GitHub MCP Server
// WHY: To extract integer parameters with defaults
// HOW: By checking parameter and providing default if needed
// EXTENT: All integer parameter processing
func OptionalIntParamWithDefault(r mcp.CallToolRequest, p string, d int) (int, error) {
	v, err := OptionalIntParam(r, p)
	if err != nil {
		return 0, err
	}
	if v == 0 {
		return d, nil
	}
	return v, nil
}

// WHO: ParameterHelper
// WHAT: Optional String Array Extraction
// WHEN: During parameter processing
// WHERE: GitHub MCP Server
// WHY: To safely extract optional string array parameters
// HOW: By checking parameter existence and converting types
// EXTENT: All string array parameter processing
func OptionalStringArrayParam(r mcp.CallToolRequest, p string) ([]string, error) {
	// Check if the parameter is present in the request
	if _, ok := r.Params.Arguments[p]; !ok {
		return []string{}, nil
	}

	switch v := r.Params.Arguments[p].(type) {
	case nil:
		return []string{}, nil
	case []string:
		return v, nil
	case []any:
		strSlice := make([]string, len(v))
		for i, v := range v {
			s, ok := v.(string)
			if !ok {
				return []string{}, fmt.Errorf("parameter %s is not of type string, is %T", p, v)
			}
			strSlice[i] = s
		}
		return strSlice, nil
	default:
		return []string{}, fmt.Errorf("parameter %s could not be coerced to []string, is %T", p, r.Params.Arguments[p])
	}
}

// WHO: ToolOptionProvider
// WHAT: Pagination Option Provider
// WHEN: During tool definition
// WHERE: GitHub MCP Server
// WHY: To add standardized pagination parameters
// HOW: By adding page and perPage parameters to tool
// EXTENT: All paginated tools
func WithPagination() mcp.ToolOption {
	return func(tool *mcp.Tool) {
		mcp.WithNumber("page",
			mcp.Description("Page number for pagination (min 1)"),
			mcp.Min(1),
		)(tool)

		mcp.WithNumber("perPage",
			mcp.Description("Results per page for pagination (min 1, max 100)"),
			mcp.Min(1),
			mcp.Max(100),
		)(tool)
	}
}

// WHO: PaginationHelper
// WHAT: Pagination Parameter Extraction
// WHEN: During parameter processing
// WHERE: GitHub MCP Server
// WHY: To safely extract pagination parameters
// HOW: By checking parameters and providing defaults
// EXTENT: All paginated requests
func OptionalPaginationParams(r mcp.CallToolRequest) (PaginationParams, error) {
	page, err := OptionalIntParamWithDefault(r, "page", 1)
	if err != nil {
		return PaginationParams{}, err
	}
	perPage, err := OptionalIntParamWithDefault(r, "perPage", 30)
	if err != nil {
		return PaginationParams{}, err
	}
	return PaginationParams{
		page:    page,
		perPage: perPage,
	}, nil
}

// Extract common PR parameters from a request
func extractPRParams(r mcp.CallToolRequest) (prParams, error) {
	// Extract required parameters
	owner, err := RequiredParam[string](r, "owner")
	if err != nil {
		return prParams{}, err
	}

	repo, err := RequiredParam[string](r, "repo")
	if err != nil {
		return prParams{}, err
	}

	number, err := RequiredIntParam(r, "pullNumber")
	if err != nil {
		return prParams{}, err
	}

	return prParams{
		owner:  owner,
		repo:   repo,
		number: number,
	}, nil
}

// WHO: ClientPRParamsPreparation
// WHAT: Client and PR Parameters Preparation
// WHEN: During PR operations
// WHERE: GitHub MCP Server
// WHY: To standardize client and parameter extraction
// HOW: By combining client and parameter extraction
// EXTENT: All PR operations
func prepareClientAndPRParams(ctx context.Context, getClient GetClientFn, r mcp.CallToolRequest) (*github.Client, prParams, error) {
	// Get GitHub client
	client, err := getClient(ctx)
	if err != nil {
		return nil, prParams{}, fmt.Errorf(ErrGetGitHubClient, err)
	}

	// Extract PR parameters
	params, err := extractPRParams(r)
	if err != nil {
		return nil, prParams{}, err
	}

	return client, params, nil
}
