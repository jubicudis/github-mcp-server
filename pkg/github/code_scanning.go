/*
 * WHO: CodeScanningProvider
 * WHAT: GitHub Code Scanning API integration for MCP
 * WHEN: During code scanning operations
 * WHERE: System Layer 6 (Integration)
 * WHY: To provide security vulnerability information
 * HOW: Using GitHub Code Scanning API
 * EXTENT: All code scanning alerts and operations
 */

package ghmcp

import (
	"context"
	"fmt"
	"io"
	"net/http"

	"github.com/jubicudis/github-mcp-server/pkg/common"
	"github.com/jubicudis/github-mcp-server/pkg/translations"

	"github.com/google/go-github/v71/github"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// Canonical code scanning logic for GitHub MCP server
// Remove all stubs, placeholders, and incomplete logic
// All types and methods must be robust, DRY, and reference only canonical helpers from /pkg/common
// All bridge and event logic must be fully implemented

// WHO: SecurityAlertRetriever
// WHAT: Get code scanning alert details
// WHEN: During security analysis
// WHERE: System Layer 6 (Integration)
// WHY: To provide vulnerability information
// HOW: Using GitHub Code Scanning API
// EXTENT: Single alert retrieval
func GetCodeScanningAlert(getClient common.GetClientFn, t translations.TranslationHelperFunc) (mcp.Tool, server.ToolHandlerFunc) {
	tool := mcp.NewTool("get_code_scanning_alert",
		mcp.WithDescription(t("tool.get_code_scanning_alert.description", "Get details of a specific code scanning alert in a GitHub repository")),
		mcp.WithString("owner",
			mcp.Required(),
			mcp.Description(t("tool.get_code_scanning_alert.input.owner.description", "The owner of the repository.")),
		),
		mcp.WithString("repo",
			mcp.Required(),
			mcp.Description(t("tool.get_code_scanning_alert.input.repo.description", "The name of the repository.")),
		),
		mcp.WithNumber("alertNumber",
			mcp.Required(),
			mcp.Description(t("tool.get_code_scanning_alert.input.alertNumber.description", "The number of the alert.")),
		),
	)
	handler := handleGetCodeScanningAlert(getClient, t) // Pass 't' to the handler
	return tool, handler
}

// extractAlertParams extracts parameters for getting a specific code scanning alert
func extractAlertParams(request mcp.CallToolRequest) (owner, repo string, alertNumber int, err error) {
	owner, err = common.RequiredParam[string](request, "owner")
	if err != nil {
		return
	}

	repo, err = common.RequiredParam[string](request, "repo")
	if err != nil {
		return
	}

	alertNumber, err = common.RequiredIntParam(request, "alertNumber")
	if err != nil {
		return
	}

	return owner, repo, alertNumber, nil
}

// Refactor common logic for parameter extraction
func extractRequiredStringParam(request mcp.CallToolRequest, paramName string) (string, error) {
	return common.RequiredParam[string](request, paramName)
}

func extractOptionalStringParam(request mcp.CallToolRequest, paramName string) (string, error) {
	val, _, err := common.OptionalParamOK[string](request, paramName)
	return val, err
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
func handleGetCodeScanningAlert(getClient common.GetClientFn, t translations.TranslationHelperFunc) server.ToolHandlerFunc { // Add 't' parameter
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		owner, repo, alertNumber, err := extractAlertParams(request)
		if err != nil {
			// Use common.CreateErrorResponse for consistent error handling
			return common.CreateErrorResponse(t, "error.invalid_input", "Invalid input: %v", err)
		}

		client, err := getClient(ctx)
		if err != nil {
			// Use common.CreateErrorResponse for consistent error handling
			return common.CreateErrorResponse(t, "error.client_initialization_failed", "Failed to initialize GitHub client: %v", err)
		}

		alertJSON, err := fetchCodeScanningAlert(ctx, client, owner, repo, int64(alertNumber))
		if err != nil {
			// Use common.CreateErrorResponse for consistent error handling
			return common.CreateErrorResponse(t, "error.fetch_alert_failed", "Failed to fetch code scanning alert: %v", err)
		}

		result := mcp.NewToolResultText(string(alertJSON))
		return result, nil // Return the result directly
	}
}

// WHO: SecurityAlertLister
// WHAT: List code scanning alerts
// WHEN: During security analysis
// WHERE: System Layer 6 (Integration)
// WHY: To enumerate security vulnerabilities
// HOW: Using GitHub Code Scanning API
// EXTENT: Repository-wide alert enumeration
func ListCodeScanningAlerts(getClient common.GetClientFn, t translations.TranslationHelperFunc) (mcp.Tool, server.ToolHandlerFunc) {
	tool := mcp.NewTool("list_code_scanning_alerts",
		mcp.WithDescription(t("tool.list_code_scanning_alerts.description", "List code scanning alerts in a GitHub repository")),
		mcp.WithString("owner",
			mcp.Required(),
			mcp.Description(t("tool.list_code_scanning_alerts.input.owner.description", "The owner of the repository.")),
		),
		mcp.WithString("repo",
			mcp.Required(),
			mcp.Description(t("tool.list_code_scanning_alerts.input.repo.description", "The name of the repository.")),
		),
		mcp.WithString("ref",
			mcp.Description(t("tool.list_code_scanning_alerts.input.ref.description", "The Git reference for the results you want to list.")),
		),
		mcp.WithString("state",
			mcp.Description(t("tool.list_code_scanning_alerts.input.state.description", "State of the code scanning alerts to list. Set to closed to list only closed code scanning alerts. Default: open")),
			mcp.DefaultString("open"),
		),
		mcp.WithString("severity",
			mcp.Description(t("tool.list_code_scanning_alerts.input.severity.description", "Only code scanning alerts with this severity will be returned. Possible values are: critical, high, medium, low, warning, note, error.")),
		),
	)
	handler := handleListCodeScanningAlerts(getClient, t) // Pass 't' to the handler
	return tool, handler
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
func handleListCodeScanningAlerts(getClient common.GetClientFn, t translations.TranslationHelperFunc) server.ToolHandlerFunc { // Add 't' parameter
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		owner, repo, ref, state, err := extractCodeScanningParams(request)
		if err != nil {
			// Use common.CreateErrorResponse for consistent error handling
			return common.CreateErrorResponse(t, "error.invalid_input", "Invalid input: %v", err)
		}

		client, err := getClient(ctx)
		if err != nil {
			// Use common.CreateErrorResponse for consistent error handling
			return common.CreateErrorResponse(t, "error.client_initialization_failed", "Failed to initialize GitHub client: %v", err)
		}

		alertsJSON, err := fetchCodeScanningAlerts(ctx, client, owner, repo, ref, state)
		if err != nil {
			// Use common.CreateErrorResponse for consistent error handling
			return common.CreateErrorResponse(t, "error.fetch_alerts_failed", "Failed to fetch code scanning alerts: %v", err)
		}

		result := mcp.NewToolResultText(string(alertsJSON))
		return result, nil // Return the result directly
	}
}
