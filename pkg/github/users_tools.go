package ghmcp

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/google/go-github/v71/github"
	"github.com/jubicudis/Tranquility-Neuro-OS/github-mcp-server/pkg/common"
	"github.com/jubicudis/Tranquility-Neuro-OS/github-mcp-server/pkg/translations"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// GetUser defines the tool for retrieving a GitHub user.
// If no username is provided, it fetches the authenticated user.
func GetUser(
	clientFn common.GetClientFn,
	translateFn translations.TranslationHelperFunc,
) (mcp.Tool, server.ToolHandlerFunc) {
	tool := mcp.NewTool("get_user",
		mcp.WithDescription(translateFn("tool.get_user.description", "Retrieves information about a GitHub user. If no username is specified, it retrieves information about the authenticated user.")),
		mcp.WithString("username",
			mcp.Description(translateFn("tool.get_user.input.username.description", "The username of the GitHub user to retrieve. If omitted, retrieves the authenticated user.")),
		),
	)

	handler := func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		client, err := clientFn(ctx)
		if err != nil {
			return common.CreateErrorResponse(translateFn, "error.client_initialization_failed", "Failed to initialize GitHub client: %v", err)
		}

		username, usernameOk, err := common.OptionalParamOK[string](request, "username")
		if err != nil {
			return common.CreateErrorResponse(translateFn, "error.invalid_input_username", "Invalid username parameter: %v", err)
		}

		var user *github.User
		var resp *github.Response
		var apiErr error

		if usernameOk && username != "" {
			user, resp, apiErr = client.Users.Get(ctx, username)
		} else {
			user, resp, apiErr = client.Users.Get(ctx, "") // Empty string for authenticated user
		}

		if apiErr != nil {
			statusCode := 0
			if resp != nil {
				statusCode = resp.StatusCode
			}
			if resp != nil && resp.StatusCode == http.StatusNotFound {
				return common.CreateErrorResponse(translateFn, "error.user_not_found", "User not found: %s", username)
			}
			return common.CreateErrorResponse(translateFn, "error.get_user_failed", "Failed to get user (status: %d): %v", statusCode, apiErr)
		}
		
		if resp.StatusCode != http.StatusOK {
			return common.CreateErrorResponse(translateFn, "error.get_user_failed_status", "Failed to get user, received status code %d", resp.StatusCode)
		}


		jsonResult, err := json.Marshal(user)
		if err != nil {
			return common.CreateErrorResponse(translateFn, "error.json_marshal_failed", "Failed to marshal user data: %v", err)
		}

		result := mcp.NewToolResultText(string(jsonResult))
		return result, nil
	}

	return tool, handler
}
