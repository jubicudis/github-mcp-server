/*
 * WHO: CodeScanningProvider
 * WHAT: GitHub Code Scanning API integration for MCP
 * WHEN: During code scanning operations
 * WHERE: System Layer 6 (Integration)
 * WHY: To provide security vulnerability information
 * HOW: Using GitHub Code Scanning API
 * EXTENT: All code scanning alerts and operations
 */

package github

import (
	"context"
	"fmt"
	"github-mcp-server/pkg/translations"
	"io"
	"net/http"

	"github.com/google/go-github/v71/github"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// WHO: SecurityAlertRetriever
// WHAT: Get code scanning alert details
// WHEN: During security analysis
// WHERE: System Layer 6 (Integration)
// WHY: To provide vulnerability information
// HOW: Using GitHub Code Scanning API
// EXTENT: Single alert retrieval
func GetCodeScanningAlert(getClient GetClientFn, t translations.TranslationHelperFunc) (tool mcp.Tool, handler server.ToolHandlerFunc) {
	return mcp.NewTool("get_code_scanning_alert",
			mcp.WithDescription("Get details of a specific code scanning alert in a GitHub repository"),
			mcp.WithString("owner",
				mcp.Required(),
				mcp.Description("The owner of the repository."),
			),
			mcp.WithString("repo",
				mcp.Required(),
				mcp.Description("The name of the repository."),
			),
			mcp.WithNumber("alertNumber",
				mcp.Required(),
				mcp.Description("The number of the alert."),
			),
		),
		handleGetCodeScanningAlert(getClient)
}

// extractAlertParams extracts parameters for getting a specific code scanning alert
func extractAlertParams(request mcp.CallToolRequest) (owner, repo string, alertNumber int, err error) {
	owner, err = RequiredParam[string](request, "owner")
	if err != nil {
		return
	}

	repo, err = RequiredParam[string](request, "repo")
	if err != nil {
		return
	}

	alertNumberInt, err := RequiredInt(request, "alertNumber")
	if err != nil {
		return
	}

	return owner, repo, alertNumberInt, nil
}

// Refactor common logic for parameter extraction
func extractRequiredStringParam(request mcp.CallToolRequest, paramName string) (string, error) {
	return RequiredParam[string](request, paramName)
}

func extractOptionalStringParam(request mcp.CallToolRequest, paramName string) (string, error) {
	return OptionalParam[string](request, paramName)
}

// Refactor common logic for API response handling
func handleAPIResponse(resp *github.Response, err error) ([]byte, error) {
	if err != nil {
		return nil, fmt.Errorf("API call failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to read response body: %w", err)
		}
		return nil, fmt.Errorf("API call failed: %s", string(body))
	}

	return io.ReadAll(resp.Body)
}

// fetchCodeScanningAlert fetches a specific alert from GitHub API
func fetchCodeScanningAlert(ctx context.Context, client *github.Client, owner, repo string, alertNumber int64) ([]byte, error) {
	_, resp, err := client.CodeScanning.GetAlert(ctx, owner, repo, alertNumber)
	if err != nil {
		return nil, fmt.Errorf("failed to get alert: %w", err)
	}

	return handleAPIResponse(resp, err)
}

// handleGetCodeScanningAlert implements the handler logic for getting a code scanning alert
func handleGetCodeScanningAlert(getClient GetClientFn) server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		owner, repo, alertNumber, err := extractAlertParams(request)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		client, err := getClient(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to get GitHub client: %w", err)
		}

		alertJSON, err := fetchCodeScanningAlert(ctx, client, owner, repo, int64(alertNumber))
		if err != nil {
			// Check if this is already a formatted error
			if e, ok := err.(interface{ Error() string }); ok {
				return mcp.NewToolResultError(e.Error()), nil
			}
			return nil, err
		}

		return mcp.NewToolResultText(string(alertJSON)), nil
	}
}

// WHO: SecurityAlertLister
// WHAT: List code scanning alerts
// WHEN: During security analysis
// WHERE: System Layer 6 (Integration)
// WHY: To enumerate security vulnerabilities
// HOW: Using GitHub Code Scanning API
// EXTENT: Repository-wide alert enumeration
func ListCodeScanningAlerts(getClient GetClientFn, t translations.TranslationHelperFunc) (tool mcp.Tool, handler server.ToolHandlerFunc) {
	return mcp.NewTool("list_code_scanning_alerts",
			mcp.WithDescription("List code scanning alerts in a GitHub repository"),
			mcp.WithString("owner",
				mcp.Required(),
				mcp.Description("The owner of the repository."),
			),
			mcp.WithString("repo",
				mcp.Required(),
				mcp.Description("The name of the repository."),
			),
			mcp.WithString("ref",
				mcp.Description("The Git reference for the results you want to list."),
			),
			mcp.WithString("state",
				mcp.Description("State of the code scanning alerts to list. Set to closed to list only closed code scanning alerts. Default: open"),
				mcp.DefaultString("open"),
			),
			mcp.WithString("severity",
				mcp.Description("Only code scanning alerts with this severity will be returned. Possible values are: critical, high, medium, low, warning, note, error."),
			),
		),
		handleListCodeScanningAlerts(getClient)
}

// extractCodeScanningParams extracts parameters for code scanning API
func extractCodeScanningParams(request mcp.CallToolRequest) (owner, repo, ref, state string, err error) {
	owner, err = extractRequiredStringParam(request, "owner")
	if err != nil {
		return
	}

	repo, err = extractRequiredStringParam(request, "repo")
	if err != nil {
		return
	}

	ref, err = extractOptionalStringParam(request, "ref")
	if err != nil {
		return
	}

	state, err = extractOptionalStringParam(request, "state")
	if err != nil {
		return
	}

	return
}

// fetchCodeScanningAlerts fetches alerts from GitHub API
func fetchCodeScanningAlerts(ctx context.Context, client *github.Client, owner, repo, ref, state string) ([]byte, error) {
	opts := &github.AlertListOptions{
		Ref:   ref,
		State: state,
	}

	_, resp, err := client.CodeScanning.ListAlertsForRepo(ctx, owner, repo, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to list alerts: %w", err)
	}

	return handleAPIResponse(resp, err)
}

// handleListCodeScanningAlerts implements the handler logic for listing code scanning alerts
func handleListCodeScanningAlerts(getClient GetClientFn) server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		owner, repo, ref, state, err := extractCodeScanningParams(request)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		client, err := getClient(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to get GitHub client: %w", err)
		}

		alertsJSON, err := fetchCodeScanningAlerts(ctx, client, owner, repo, ref, state)
		if err != nil {
			// Check if this is already a formatted error message
			if e, ok := err.(interface{ Error() string }); ok {
				return mcp.NewToolResultError(e.Error()), nil
			}
			return nil, err
		}

		return mcp.NewToolResultText(string(alertsJSON)), nil
	}
}
