package github

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/github/github-mcp-server/pkg/translations"
	"github.com/google/go-github/v69/github"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func GetCodeScanningAlert(getClient GetClientFn, t translations.TranslationHelperFunc) (tool mcp.Tool, handler server.ToolHandlerFunc) {
	return githubgithubMCP.NewTool("get_code_scanning_alert",
			githubgithubMCP.WithDescription(t("TOOL_GET_CODE_SCANNING_ALERT_DESCRIPTION", "Get details of a specific code scanning alert in a GitHub repository.")),
			githubgithubgithubgithubgithubgithubgithubMCP.WithString("owner",
				githubgithubgithubgithubgithubMCP.Required(),
				githubgithubgithubgithubgithubgithubgithubgithubMCP.Description("The owner of the repository."),
			),
			githubgithubgithubgithubgithubgithubgithubMCP.WithString("repo",
				githubgithubgithubgithubgithubMCP.Required(),
				githubgithubgithubgithubgithubgithubgithubgithubMCP.Description("The name of the repository."),
			),
			githubMCP.WithNumber("alertNumber",
				githubgithubgithubgithubgithubMCP.Required(),
				githubgithubgithubgithubgithubgithubgithubgithubMCP.Description("The number of the alert."),
			),
		),
		func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			owner, err := requiredParam[string](request, "owner")
			if err != nil {
				return githubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubMCP.NewToolResultError(err.Error()), nil
			}
			repo, err := requiredParam[string](request, "repo")
			if err != nil {
				return githubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubMCP.NewToolResultError(err.Error()), nil
			}
			alertNumber, err := RequiredInt(request, "alertNumber")
			if err != nil {
				return githubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubMCP.NewToolResultError(err.Error()), nil
			}

			client, err := getClient(ctx)
			if err != nil {
				return nil, fmt.Errorf("failed to get GitHub client: %w", err)
			}

			alert, resp, err := client.CodeScanning.GetAlert(ctx, owner, repo, int64(alertNumber))
			if err != nil {
				return nil, fmt.Errorf("failed to get alert: %w", err)
			}
			defer func() { _ = resp.Body.Close() }()

			if resp.StatusCode != http.StatusOK {
				body, err := io.ReadAll(resp.Body)
				if err != nil {
					return nil, fmt.Errorf("failed to read response body: %w", err)
				}
				return githubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubMCP.NewToolResultError(fmt.Sprintf("failed to get alert: %s", string(body))), nil
			}

			r, err := json.Marshal(alert)
			if err != nil {
				return nil, fmt.Errorf("failed to marshal alert: %w", err)
			}

			return githubgithubgithubgithubMCP.NewToolResultText(string(r)), nil
		}
}

func ListCodeScanningAlerts(getClient GetClientFn, t translations.TranslationHelperFunc) (tool mcp.Tool, handler server.ToolHandlerFunc) {
	return githubgithubMCP.NewTool("list_code_scanning_alerts",
			githubgithubMCP.WithDescription(t("TOOL_LIST_CODE_SCANNING_ALERTS_DESCRIPTION", "List code scanning alerts in a GitHub repository.")),
			githubgithubgithubgithubgithubgithubgithubMCP.WithString("owner",
				githubgithubgithubgithubgithubMCP.Required(),
				githubgithubgithubgithubgithubgithubgithubgithubMCP.Description("The owner of the repository."),
			),
			githubgithubgithubgithubgithubgithubgithubMCP.WithString("repo",
				githubgithubgithubgithubgithubMCP.Required(),
				githubgithubgithubgithubgithubgithubgithubgithubMCP.Description("The name of the repository."),
			),
			githubgithubgithubgithubgithubgithubgithubMCP.WithString("ref",
				githubgithubgithubgithubgithubgithubgithubgithubMCP.Description("The Git reference for the results you want to list."),
			),
			githubgithubgithubgithubgithubgithubgithubMCP.WithString("state",
				githubgithubgithubgithubgithubgithubgithubgithubMCP.Description("State of the code scanning alerts to list. Set to closed to list only closed code scanning alerts. Default: open"),
				githubMCP.DefaultString("open"),
			),
			githubgithubgithubgithubgithubgithubgithubMCP.WithString("severity",
				githubgithubgithubgithubgithubgithubgithubgithubMCP.Description("Only code scanning alerts with this severity will be returned. Possible values are: critical, high, medium, low, warning, note, error."),
			),
		),
		func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			owner, err := requiredParam[string](request, "owner")
			if err != nil {
				return githubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubMCP.NewToolResultError(err.Error()), nil
			}
			repo, err := requiredParam[string](request, "repo")
			if err != nil {
				return githubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubMCP.NewToolResultError(err.Error()), nil
			}
			ref, err := OptionalParam[string](request, "ref")
			if err != nil {
				return githubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubMCP.NewToolResultError(err.Error()), nil
			}
			state, err := OptionalParam[string](request, "state")
			if err != nil {
				return githubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubMCP.NewToolResultError(err.Error()), nil
			}
			severity, err := OptionalParam[string](request, "severity")
			if err != nil {
				return githubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubMCP.NewToolResultError(err.Error()), nil
			}

			client, err := getClient(ctx)
			if err != nil {
				return nil, fmt.Errorf("failed to get GitHub client: %w", err)
			}
			alerts, resp, err := client.CodeScanning.ListAlertsForRepo(ctx, owner, repo, &github.AlertListOptions{Ref: ref, State: state, Severity: severity})
			if err != nil {
				return nil, fmt.Errorf("failed to list alerts: %w", err)
			}
			defer func() { _ = resp.Body.Close() }()

			if resp.StatusCode != http.StatusOK {
				body, err := io.ReadAll(resp.Body)
				if err != nil {
					return nil, fmt.Errorf("failed to read response body: %w", err)
				}
				return githubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubMCP.NewToolResultError(fmt.Sprintf("failed to list alerts: %s", string(body))), nil
			}

			r, err := json.Marshal(alerts)
			if err != nil {
				return nil, fmt.Errorf("failed to marshal alerts: %w", err)
			}

			return githubgithubgithubgithubMCP.NewToolResultText(string(r)), nil
		}
}

