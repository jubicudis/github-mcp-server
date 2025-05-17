package github

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/tranquility-dev/github-mcp-server/pkg/translations"

	"github.com/google/go-github/v49/github"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func GetCommit(getClient GetClientFn, t translations.TranslationHelperFunc) (tool mcp.Tool, handler server.ToolHandlerFunc) {
	return mcp.NewTool("get_commit",
			mcp.WithDescription(t("TOOL_GET_COMMITS_DESCRIPTION", "Get details for a commit from a GitHub repository")),
			mcp.WithString("owner",
				mcp.Required(),
				mcp.Description("Repository owner"),
			),
			mcp.WithString("repo",
				mcp.Required(),
				mcp.Description("Repository name"),
			),
			mcp.WithString("sha",
				mcp.Required(),
				mcp.Description("Commit SHA, branch name, or tag name"),
			),
			WithPagination(),
		),
		func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			owner, err := requiredParam[string](request, "owner")
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			repo, err := requiredParam[string](request, "repo")
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			sha, err := requiredParam[string](request, "sha")
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			pagination, err := OptionalPaginationParams(request)
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}

			opts := &github.ListOptions{
				Page:    pagination.page,
				PerPage: pagination.perPage,
			}

			client, err := getClient(ctx)
			if err != nil {
				return nil, fmt.Errorf("failed to get GitHub client: %w", err)
			}
			commit, resp, err := client.Repositories.GetCommit(ctx, owner, repo, sha, opts)
			if err != nil {
				return nil, fmt.Errorf("failed to get commit: %w", err)
			}
			defer func() { _ = resp.Body.Close() }()

			if resp.StatusCode != 200 {
				body, err := io.ReadAll(resp.Body)
				if err != nil {
					return nil, fmt.Errorf("failed to read response body: %w", err)
				}
				return mcp.NewToolResultError(fmt.Sprintf("failed to get commit: %s", string(body))), nil
			}

			r, err := json.Marshal(commit)
			if err != nil {
				return nil, fmt.Errorf("failed to marshal response: %w", err)
			}

			return mcp.NewToolResultText(string(r)), nil
		}
}

// ListCommits creates a tool to get commits of a branch in a repository.
func ListCommits(getClient GetClientFn, t translations.TranslationHelperFunc) (tool mcp.Tool, handler server.ToolHandlerFunc) {
	return mcp.NewTool("list_commits",
			mcp.WithDescription(t("TOOL_LIST_COMMITS_DESCRIPTION", "Get list of commits of a branch in a GitHub repository")),
			mcp.WithString("owner",
				mcp.Required(),
				mcp.Description("Repository owner"),
			),
			mcp.WithString("repo",
				mcp.Required(),
				mcp.Description("Repository name"),
			),
			mcp.WithString("sha",
				mcp.Description("Branch name"),
			),
			WithPagination(),
		),
		func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			owner, err := requiredParam[string](request, "owner")
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			repo, err := requiredParam[string](request, "repo")
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}

			opts := &github.CommitsListOptions{}

			sha, err := OptionalParam[string](request, "sha")
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			if sha != "" {
				opts.SHA = sha
			}

			pagination, err := OptionalPaginationParams(request)
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			opts.ListOptions = github.ListOptions{
				Page:    pagination.page,
				PerPage: pagination.perPage,
			}

			client, err := getClient(ctx)
			if err != nil {
				return nil, fmt.Errorf("failed to get GitHub client: %w", err)
			}
			commits, resp, err := client.Repositories.ListCommits(ctx, owner, repo, opts)
			if err != nil {
				return nil, fmt.Errorf("failed to list commits: %w", err)
			}
			defer func() { _ = resp.Body.Close() }()

			if resp.StatusCode != http.StatusOK {
				body, err := io.ReadAll(resp.Body)
				if err != nil {
					return nil, fmt.Errorf("failed to read response body: %w", err)
				}
				return mcp.NewToolResultError(fmt.Sprintf("failed to list commits: %s", string(body))), nil
			}

			r, err := json.Marshal(commits)
			if err != nil {
				return nil, fmt.Errorf("failed to marshal response: %w", err)
			}

			return mcp.NewToolResultText(string(r)), nil
		}
}

// ListBranches creates a tool to list branches in a GitHub repository.
func ListBranches(getClient GetClientFn, t translations.TranslationHelperFunc) (tool mcp.Tool, handler server.ToolHandlerFunc) {
	return mcp.NewTool("list_branches",
			mcp.WithDescription(t("TOOL_LIST_BRANCHES_DESCRIPTION", "List branches in a GitHub repository")),
			mcp.WithString("owner",
				mcp.Required(),
				mcp.Description("Repository owner"),
			),
			mcp.WithString("repo",
				mcp.Required(),
				mcp.Description("Repository name"),
			),
			mcp.WithNumber("page",
				mcp.Description("Page number"),
			),
			mcp.WithNumber("perPage",
				mcp.Description("Number of results per page"),
			),
		),
		func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			owner, err := requiredParam[string](request, "owner")
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			repo, err := requiredParam[string](request, "repo")
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}

			page, err := OptionalIntParamWithDefault(request, "page", 1)
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			perPage, err := OptionalIntParamWithDefault(request, "perPage", 30)
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}

			opts := &github.BranchListOptions{
				ListOptions: github.ListOptions{
					Page:    page,
					PerPage: perPage,
				},
			}

			client, err := getClient(ctx)
			if err != nil {
				return nil, fmt.Errorf("failed to get GitHub client: %w", err)
			}
			branches, resp, err := client.Repositories.ListBranches(ctx, owner, repo, opts)
			if err != nil {
				return nil, fmt.Errorf("failed to list branches: %w", err)
			}
			defer func() { _ = resp.Body.Close() }()

			if resp.StatusCode != http.StatusOK {
				body, err := io.ReadAll(resp.Body)
				if err != nil {
					return nil, fmt.Errorf("failed to read response body: %w", err)
				}
				return mcp.NewToolResultError(fmt.Sprintf("failed to list branches: %s", string(body))), nil
			}

			r, err := json.Marshal(branches)
			if err != nil {
				return nil, fmt.Errorf("failed to marshal response: %w", err)
			}

			return mcp.NewToolResultText(string(r)), nil
		}
}

// CreateOrUpdateFile creates a tool to create or update a file in a GitHub repository.
func CreateOrUpdateFile(getClient GetClientFn, t translations.TranslationHelperFunc) (tool mcp.Tool, handler server.ToolHandlerFunc) {
	return mcp.NewTool("create_or_update_file",
			mcp.WithDescription(t("TOOL_CREATE_OR_UPDATE_FILE_DESCRIPTION", "Create or update a file in a GitHub repository")),
			mcp.WithString("owner",
				mcp.Required(),
				mcp.Description("Repository owner"),
			),
			mcp.WithString("repo",
				mcp.Required(),
				mcp.Description("Repository name"),
			),
			mcp.WithString("path",
				mcp.Required(),
				mcp.Description("Path in the repository including filename"),
			),
			mcp.WithString("content",
				mcp.Required(),
				mcp.Description("File content"),
			),
			mcp.WithString("message",
				mcp.Required(),
				mcp.Description("Commit message"),
			),
			mcp.WithString("branch",
				mcp.Required(),
				mcp.Description("Branch name"),
			),
			mcp.WithString("sha",
				mcp.Description("The blob SHA of the file being replaced (required for updating an existing file)"),
			),
		),
		func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			owner, err := requiredParam[string](request, "owner")
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			repo, err := requiredParam[string](request, "repo")
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			path, err := requiredParam[string](request, "path")
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			content, err := requiredParam[string](request, "content")
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			message, err := requiredParam[string](request, "message")
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			branch, err := requiredParam[string](request, "branch")
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			sha, err := OptionalParam[string](request, "sha")
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}

			opts := &github.RepositoryContentFileOptions{
				Message: Ptr(message),
				Content: []byte(content),
				Branch:  Ptr(branch),
			}

			if sha != "" {
				opts.SHA = Ptr(sha)
			}

			client, err := getClient(ctx)
			if err != nil {
				return nil, fmt.Errorf("failed to get GitHub client: %w", err)
			}
			contentResponse, resp, err := client.Repositories.CreateFile(ctx, owner, repo, path, opts)
			if err != nil {
				return nil, fmt.Errorf("failed to create/update file: %w", err)
			}
			defer func() { _ = resp.Body.Close() }()

			if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
				body, err := io.ReadAll(resp.Body)
				if err != nil {
					return nil, fmt.Errorf("failed to read response body: %w", err)
				}
				return mcp.NewToolResultError(fmt.Sprintf("failed to create/update file: %s", string(body))), nil
			}

			r, err := json.Marshal(contentResponse)
			if err != nil {
				return nil, fmt.Errorf("failed to marshal response: %w", err)
			}

			return mcp.NewToolResultText(string(r)), nil
		}
}

// CreateRepository creates a tool to create a new GitHub repository.
func CreateRepository(getClient GetClientFn, t translations.TranslationHelperFunc) (tool mcp.Tool, handler server.ToolHandlerFunc) {
	return mcp.NewTool("create_repository",
			mcp.WithDescription(t("TOOL_CREATE_REPOSITORY_DESCRIPTION", "Create a new GitHub repository")),
			mcp.WithString("name",
				mcp.Required(),
				mcp.Description("Repository name"),
			),
			mcp.WithString("description",
				mcp.Description("Repository description"),
			),
			mcp.WithBoolean("private",
				mcp.Description("Whether the repository should be private"),
			),
			mcp.WithBoolean("autoInit",
				mcp.Description("Create an initial commit with README"),
			),
		),
		func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			name, err := requiredParam[string](request, "name")
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}

			description, err := OptionalParam[string](request, "description")
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}

			private, err := OptionalParamWithDefault[bool](request, "private", false)
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}

			autoInit, err := OptionalParamWithDefault[bool](request, "autoInit", false)
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}

			opts := &github.Repository{
				Name:        Ptr(name),
				Description: Ptr(description),
				Private:     Ptr(private),
				AutoInit:    Ptr(autoInit),
			}

			client, err := getClient(ctx)
			if err != nil {
				return nil, fmt.Errorf("failed to get GitHub client: %w", err)
			}
			repo, resp, err := client.Repositories.Create(ctx, "", opts)
			if err != nil {
				return nil, fmt.Errorf("failed to create repository: %w", err)
			}
			defer func() { _ = resp.Body.Close() }()

			if resp.StatusCode != http.StatusCreated {
				body, err := io.ReadAll(resp.Body)
				if err != nil {
					return nil, fmt.Errorf("failed to read response body: %w", err)
				}
				return mcp.NewToolResultError(fmt.Sprintf("failed to create repository: %s", string(body))), nil
			}

			r, err := json.Marshal(repo)
			if err != nil {
				return nil, fmt.Errorf("failed to marshal response: %w", err)
			}

			return mcp.NewToolResultText(string(r)), nil
		}
}

// GetFileContents creates a tool to get the contents of a file or directory from a GitHub repository.
func GetFileContents(getClient GetClientFn, t translations.TranslationHelperFunc) (tool mcp.Tool, handler server.ToolHandlerFunc) {
	return mcp.NewTool("get_file_contents",
			mcp.WithDescription(t("TOOL_GET_FILE_CONTENTS_DESCRIPTION", "Get the contents of a file or directory from a GitHub repository")),
			mcp.WithString("owner",
				mcp.Required(),
				mcp.Description("Repository owner"),
			),
			mcp.WithString("repo",
				mcp.Required(),
				mcp.Description("Repository name"),
			),
			mcp.WithString("path",
				mcp.Required(),
				mcp.Description("Path in the repository including filename"),
			),
			mcp.WithString("branch",
				mcp.Description("Branch name, tag, or commit SHA"),
			),
		),
		func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			owner, err := requiredParam[string](request, "owner")
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			repo, err := requiredParam[string](request, "repo")
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			path, err := requiredParam[string](request, "path")
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			branch, err := OptionalParam[string](request, "branch")
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}

			opts := &github.RepositoryContentGetOptions{}
			if branch != "" {
				opts.Ref = branch
			}

			client, err := getClient(ctx)
			if err != nil {
				return nil, fmt.Errorf("failed to get GitHub client: %w", err)
			}

			// Get the file contents
			fileContent, directoryContent, resp, err := client.Repositories.GetContents(ctx, owner, repo, path, opts)
			if err != nil {
				return nil, fmt.Errorf("failed to get file contents: %w", err)
			}
			defer func() { _ = resp.Body.Close() }()

			if resp.StatusCode != http.StatusOK {
				body, err := io.ReadAll(resp.Body)
				if err != nil {
					return nil, fmt.Errorf("failed to read response body: %w", err)
				}
				return mcp.NewToolResultError(fmt.Sprintf("failed to get file contents: %s", string(body))), nil
			}

			// If it's a file, return the file content, otherwise return the directory content
			var r []byte
			var marshallErr error
			if fileContent != nil {
				r, marshallErr = json.Marshal(fileContent)
			} else {
				r, marshallErr = json.Marshal(directoryContent)
			}

			if marshallErr != nil {
				return nil, fmt.Errorf("failed to marshal response: %w", marshallErr)
			}

			return mcp.NewToolResultText(string(r)), nil
		}
}

// ForkRepository creates a tool to fork a repository.
func ForkRepository(getClient GetClientFn, t translations.TranslationHelperFunc) (tool mcp.Tool, handler server.ToolHandlerFunc) {
	return mcp.NewTool("fork_repository",
			mcp.WithDescription(t("TOOL_FORK_REPOSITORY_DESCRIPTION", "Fork a GitHub repository")),
			mcp.WithString("owner",
				mcp.Required(),
				mcp.Description("Repository owner"),
			),
			mcp.WithString("repo",
				mcp.Required(),
				mcp.Description("Repository name"),
			),
			mcp.WithString("organization",
				mcp.Description("Organization to fork into (optional)"),
			),
		),
		func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			owner, err := requiredParam[string](request, "owner")
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			repo, err := requiredParam[string](request, "repo")
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			organization, err := OptionalParam[string](request, "organization")
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}

			opts := &github.RepositoryCreateForkOptions{}
			if organization != "" {
				opts.Organization = organization
			}

			client, err := getClient(ctx)
			if err != nil {
				return nil, fmt.Errorf("failed to get GitHub client: %w", err)
			}
			forkedRepo, resp, err := client.Repositories.CreateFork(ctx, owner, repo, opts)
			if err != nil {
				// GitHub returns a 202 Accepted, which causes the go-github library to return an AcceptedError
				// We need to check if the error is of type AcceptedError and handle it accordingly
				if resp != nil && resp.StatusCode == http.StatusAccepted && isAcceptedError(err) {
					// This is not really an error, the fork operation is in progress
					return mcp.NewToolResultText("Fork is in progress. It may take a few moments to complete."), nil
				}
				return nil, fmt.Errorf("failed to fork repository: %w", err)
			}
			defer func() { _ = resp.Body.Close() }()

			r, err := json.Marshal(forkedRepo)
			if err != nil {
				return nil, fmt.Errorf("failed to marshal response: %w", err)
			}

			return mcp.NewToolResultText(string(r)), nil
		}
}

// Note: isAcceptedError is defined in common.go

// OptionalParamWithDefault gets an optional parameter with a default value
func OptionalParamWithDefault[T any](request mcp.CallToolRequest, name string, defaultValue T) (T, error) {
	if val, ok := request.Params.Arguments[name]; ok && val != nil {
		// Try to convert the value to the requested type
		if converted, ok := val.(T); ok {
			return converted, nil
		}
		return defaultValue, fmt.Errorf("parameter %s has an invalid type", name)
	}
	return defaultValue, nil
}

// CreateBranch creates a tool to create a new branch.
func CreateBranch(getClient GetClientFn, t translations.TranslationHelperFunc) (tool mcp.Tool, handler server.ToolHandlerFunc) {
	return mcp.NewTool("create_branch",
			mcp.WithDescription(t("TOOL_CREATE_BRANCH_DESCRIPTION", "Create a new branch in a GitHub repository")),
			mcp.WithString("owner",
				mcp.Required(),
				mcp.Description("Repository owner"),
			),
			mcp.WithString("repo",
				mcp.Required(),
				mcp.Description("Repository name"),
			),
			mcp.WithString("branch",
				mcp.Required(),
				mcp.Description("New branch name"),
			),
			mcp.WithString("from_branch",
				mcp.Description("Source branch to create from (default: repository's default branch)"),
			),
		),
		func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			owner, err := requiredParam[string](request, "owner")
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			repo, err := requiredParam[string](request, "repo")
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			newBranch, err := requiredParam[string](request, "branch")
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			fromBranch, err := OptionalParam[string](request, "from_branch")
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}

			client, err := getClient(ctx)
			if err != nil {
				return nil, fmt.Errorf("failed to get GitHub client: %w", err)
			}

			// If source branch not provided, get default branch
			if fromBranch == "" {
				repository, resp, err := client.Repositories.Get(ctx, owner, repo)
				if err != nil {
					return nil, fmt.Errorf("failed to get repository: %w", err)
				}
				defer func() { _ = resp.Body.Close() }()
				if resp.StatusCode != http.StatusOK {
					body, err := io.ReadAll(resp.Body)
					if err != nil {
						return nil, fmt.Errorf("failed to read response body: %w", err)
					}
					return mcp.NewToolResultError(fmt.Sprintf("failed to get repository: %s", string(body))), nil
				}
				fromBranch = *repository.DefaultBranch
			}

			// Get the SHA of the source branch's HEAD
			sourceBranchRef := fmt.Sprintf("refs/heads/%s", fromBranch)
			sourceRef, resp, err := client.Git.GetRef(ctx, owner, repo, sourceBranchRef)
			if err != nil {
				return nil, fmt.Errorf("failed to get reference: %w", err)
			}
			defer func() { _ = resp.Body.Close() }()
			if resp.StatusCode != http.StatusOK {
				body, err := io.ReadAll(resp.Body)
				if err != nil {
					return nil, fmt.Errorf("failed to read response body: %w", err)
				}
				return mcp.NewToolResultError(fmt.Sprintf("failed to get reference: %s", string(body))), nil
			}

			// Create the new branch
			newRef := &github.Reference{
				Ref:    Ptr(fmt.Sprintf("refs/heads/%s", newBranch)),
				Object: &github.GitObject{SHA: sourceRef.Object.SHA},
			}
			createdRef, resp, err := client.Git.CreateRef(ctx, owner, repo, newRef)
			if err != nil {
				return nil, fmt.Errorf("failed to create branch: %w", err)
			}
			defer func() { _ = resp.Body.Close() }()
			if resp.StatusCode != http.StatusCreated {
				body, err := io.ReadAll(resp.Body)
				if err != nil {
					return nil, fmt.Errorf("failed to read response body: %w", err)
				}
				return mcp.NewToolResultError(fmt.Sprintf("failed to create branch: %s", string(body))), nil
			}

			r, err := json.Marshal(createdRef)
			if err != nil {
				return nil, fmt.Errorf("failed to marshal response: %w", err)
			}

			return mcp.NewToolResultText(string(r)), nil
		}
}

// FileEntry represents a file to be committed
type FileEntry struct {
	Path    string
	Content string
}

// PushFilesParams holds all parameters for pushing files
type PushFilesParams struct {
	Owner   string
	Repo    string
	Branch  string
	Message string
	Files   []FileEntry
}

// extractPushFilesParams extracts and validates parameters for pushing files
func extractPushFilesParams(request mcp.CallToolRequest) (*PushFilesParams, error) {
	owner, err := requiredParam[string](request, "owner")
	if err != nil {
		return nil, err
	}
	repo, err := requiredParam[string](request, "repo")
	if err != nil {
		return nil, err
	}
	branch, err := requiredParam[string](request, "branch")
	if err != nil {
		return nil, err
	}
	message, err := requiredParam[string](request, "message")
	if err != nil {
		return nil, err
	}

	// Check that files is an array
	filesObj, ok := request.Params.Arguments["files"].([]interface{})
	if !ok {
		return nil, fmt.Errorf("files parameter must be an array")
	}

	// Parse files
	files, err := parseFileEntries(filesObj)
	if err != nil {
		return nil, err
	}

	return &PushFilesParams{
		Owner:   owner,
		Repo:    repo,
		Branch:  branch,
		Message: message,
		Files:   files,
	}, nil
}

// parseFileEntries converts file objects to FileEntry structs
func parseFileEntries(filesObj []interface{}) ([]FileEntry, error) {
	files := make([]FileEntry, 0, len(filesObj))

	for _, fileObj := range filesObj {
		fileMap, ok := fileObj.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("each file must be an object")
		}

		path, ok := fileMap["path"].(string)
		if !ok || path == "" {
			return nil, fmt.Errorf("each file must have a path")
		}

		content, ok := fileMap["content"].(string)
		if !ok {
			return nil, fmt.Errorf("each file must have content")
		}

		files = append(files, FileEntry{
			Path:    path,
			Content: content,
		})
	}

	return files, nil
}

// createTreeEntries converts FileEntry objects to GitHub TreeEntry objects
func createTreeEntries(files []FileEntry) []*github.TreeEntry {
	entries := make([]*github.TreeEntry, 0, len(files))

	for _, file := range files {
		entries = append(entries, &github.TreeEntry{
			Path:    Ptr(file.Path),
			Mode:    Ptr("100644"), // File mode
			Type:    Ptr("blob"),
			Content: Ptr(file.Content),
		})
	}

	return entries
}

// commitFiles performs the Git operations to commit files
func commitFiles(ctx context.Context, client *github.Client, params *PushFilesParams, entries []*github.TreeEntry) (*github.Reference, error) {
	// Get the reference to the branch
	refName := fmt.Sprintf("refs/heads/%s", params.Branch)
	ref, resp, err := client.Git.GetRef(ctx, params.Owner, params.Repo, refName)
	if err != nil {
		return nil, fmt.Errorf("failed to get branch reference: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	// Get the commit that the branch points to
	baseCommit, resp, err := client.Git.GetCommit(ctx, params.Owner, params.Repo, *ref.Object.SHA)
	if err != nil {
		return nil, fmt.Errorf("failed to get base commit: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	// Create a tree with all the files
	tree, resp, err := client.Git.CreateTree(ctx, params.Owner, params.Repo, *baseCommit.Tree.SHA, entries)
	if err != nil {
		return nil, fmt.Errorf("failed to create tree: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	// Create a commit with the new tree
	newCommit, resp, err := client.Git.CreateCommit(ctx, params.Owner, params.Repo, &github.Commit{
		Message: Ptr(params.Message),
		Tree:    tree,
		Parents: []*github.Commit{{SHA: baseCommit.SHA}},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create commit: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	// Update the branch to point to the new commit
	ref.Object.SHA = newCommit.SHA
	updatedRef, resp, err := client.Git.UpdateRef(ctx, params.Owner, params.Repo, ref, false)
	if err != nil {
		return nil, fmt.Errorf("failed to update branch: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	return updatedRef, nil
}

// PushFiles creates a tool to push multiple files in a single commit to a GitHub repository.
func PushFiles(getClient GetClientFn, t translations.TranslationHelperFunc) (tool mcp.Tool, handler server.ToolHandlerFunc) {
	return mcp.NewTool("push_files",
			mcp.WithDescription(t("TOOL_PUSH_FILES_DESCRIPTION", "Push multiple files in a single commit to a GitHub repository")),
			mcp.WithString("owner",
				mcp.Required(),
				mcp.Description("Repository owner"),
			),
			mcp.WithString("repo",
				mcp.Required(),
				mcp.Description("Repository name"),
			),
			mcp.WithString("branch",
				mcp.Required(),
				mcp.Description("Branch name"),
			),
			mcp.WithObject("files",
				mcp.Required(),
				mcp.Description("Array of file objects with path and content properties"),
			),
			mcp.WithString("message",
				mcp.Required(),
				mcp.Description("Commit message"),
			),
		),
		func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			// Extract and validate parameters
			params, err := extractPushFilesParams(request)
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}

			// Get GitHub client
			client, err := getClient(ctx)
			if err != nil {
				return nil, fmt.Errorf("failed to get GitHub client: %w", err)
			}

			// Create tree entries from files
			entries := createTreeEntries(params.Files)

			// Commit the files
			updatedRef, err := commitFiles(ctx, client, params, entries)
			if err != nil {
				return nil, err
			}

			// Format the response
			r, err := json.Marshal(updatedRef)
			if err != nil {
				return nil, fmt.Errorf("failed to marshal response: %w", err)
			}

			return mcp.NewToolResultText(string(r)), nil
		}
}
