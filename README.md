# GitHub MCP Server

GitHub MCP Server implemented in Go.

## Setup

Create a GitHub Personal Access Token with the appropriate permissions
and set it as the GITHUB_PERSONAL_ACCESS_TOKEN environment variable.

## Testing on VS Code Insiders

### Requirements

You can either use a Docker image or build the binary from the repo.

#### Docker image

As of now, this repo is private, and hence the docker image is not available publicly. To pull it,
you need to make sure you can access the GitHub docker registry. See [this](https://docs.github.com/en/packages/working-with-a-github-packages-registry/working-with-the-container-registry#authenticating-with-a-personal-access-token-classic)
for more details.

To make sure you can access the GitHub docker registry, run the following command:

```bash
docker pull ghcr.io/github/github-mcp-server:main
```

If the above command works, you are good to go.

#### Build from repo
First, install `github-mcp-server` by cloning the repo and running the following command:

```bash
go install ./cmd/github-mcp-server
```

If you don't want to clone the repo, you can run:

```bash
GOPRIVATE=github.com/github go install github.com/github/github-mcp-server/cmd/github-mcp-server@latest
```

This will install the `github-mcp-server` binary in your `$GOPATH/bin` directory.

Find where the binary is installed by running:

```bash
# note this assumes $GOPATH/bin is in your $PATH
which github-mcp-server
```

### Start VS Code Insiders

Start VS Code Insiders and make sure you pass the `GITHUB_PERSONAL_ACCESS_TOKEN` environment variable to the process.

One way to do this is to make sure that [you can run VS code from your terminal](https://code.visualstudio.com/docs/setup/mac#_launch-vs-code-from-the-command-line) and
start it with the following command:

```bash
export GITHUB_PERSONAL_ACCESS_TOKEN=your-token-here
code-insiders
```

Another way is to set the environment variable in your shell configuration file (e.g., `.bashrc`, `.zshrc`, etc.).

Run **Preferences: Open User Settings (JSON)**, and create or append to the `mcp` setting:

If you are using the docker image, use this configuration:

```json
{
  "mcp": {
    "inputs": [],
    "servers": {
      "github-mcp-server": {
        "type": "stdio",
        "command": "docker",
        "args": [
          "run",
          "-i",
          "--rm",
          "-e",
          "GITHUB_PERSONAL_ACCESS_TOKEN",
          "ghcr.io/github/github-mcp-server:main"
        ],
        "env": {}
      }
    }
  }
}
```

If you built the binary from the repo use this configuration:

```json
{
  "mcp": {
    "inputs": [ ],
    "servers": {
      "mcp-github-server": {
        "command": "path-to-your/github-mcp-server",
        "args": ["stdio"],
        "env": {   }
      }
    }
  }
}
```

Right on top of `servers`, you should see a `Start` link to start the server.


Try something like the following prompt to verify that it works:

```
I'd like to know more about my GitHub profile.
```

## GitHub Enterprise Server

The flag `--gh-host` and the environment variable `GH_HOST` can be used to set the GitHub Enterprise Server hostname.


## Tools

### Users

- **get_me** - Get details of the authenticated user
  - No parameters required

### Issues

- **get_issue** - Gets the contents of an issue within a repository

  - `owner`: Repository owner (string, required)
  - `repo`: Repository name (string, required)
  - `issue_number`: Issue number (number, required)

- **create_issue** - Create a new issue in a GitHub repository

  - `owner`: Repository owner (string, required)
  - `repo`: Repository name (string, required)
  - `title`: Issue title (string, required)
  - `body`: Issue body content (string, optional)
  - `assignees`: Comma-separated list of usernames to assign to this issue (string, optional)
  - `labels`: Comma-separated list of labels to apply to this issue (string, optional)

- **add_issue_comment** - Add a comment to an issue

  - `owner`: Repository owner (string, required)
  - `repo`: Repository name (string, required)
  - `issue_number`: Issue number (number, required)
  - `body`: Comment text (string, required)

- **list_issues** - List and filter repository issues

  - `owner`: Repository owner (string, required)
  - `repo`: Repository name (string, required)
  - `state`: Filter by state ('open', 'closed', 'all') (string, optional)
  - `labels`: Comma-separated list of labels to filter by (string, optional)
  - `sort`: Sort by ('created', 'updated', 'comments') (string, optional)
  - `direction`: Sort direction ('asc', 'desc') (string, optional)
  - `since`: Filter by date (ISO 8601 timestamp) (string, optional)
  - `page`: Page number (number, optional)
  - `per_page`: Results per page (number, optional)

- **update_issue** - Update an existing issue in a GitHub repository

  - `owner`: Repository owner (string, required)
  - `repo`: Repository name (string, required)
  - `issue_number`: Issue number to update (number, required)
  - `title`: New title (string, optional)
  - `body`: New description (string, optional)
  - `state`: New state ('open' or 'closed') (string, optional)
  - `labels`: Comma-separated list of new labels (string, optional)
  - `assignees`: Comma-separated list of new assignees (string, optional)
  - `milestone`: New milestone number (number, optional)

- **search_issues** - Search for issues and pull requests
  - `query`: Search query (string, required)
  - `sort`: Sort field (string, optional)
  - `order`: Sort order (string, optional)
  - `page`: Page number (number, optional)
  - `per_page`: Results per page (number, optional)

### Pull Requests

- **get_pull_request** - Get details of a specific pull request

  - `owner`: Repository owner (string, required)
  - `repo`: Repository name (string, required)
  - `pull_number`: Pull request number (number, required)

- **list_pull_requests** - List and filter repository pull requests

  - `owner`: Repository owner (string, required)
  - `repo`: Repository name (string, required)
  - `state`: PR state (string, optional)
  - `sort`: Sort field (string, optional)
  - `direction`: Sort direction (string, optional)
  - `per_page`: Results per page (number, optional)
  - `page`: Page number (number, optional)

- **merge_pull_request** - Merge a pull request

  - `owner`: Repository owner (string, required)
  - `repo`: Repository name (string, required)
  - `pull_number`: Pull request number (number, required)
  - `commit_title`: Title for the merge commit (string, optional)
  - `commit_message`: Message for the merge commit (string, optional)
  - `merge_method`: Merge method (string, optional)

- **get_pull_request_files** - Get the list of files changed in a pull request

  - `owner`: Repository owner (string, required)
  - `repo`: Repository name (string, required)
  - `pull_number`: Pull request number (number, required)

- **get_pull_request_status** - Get the combined status of all status checks for a pull request

  - `owner`: Repository owner (string, required)
  - `repo`: Repository name (string, required)
  - `pull_number`: Pull request number (number, required)

- **update_pull_request_branch** - Update a pull request branch with the latest changes from the base branch

  - `owner`: Repository owner (string, required)
  - `repo`: Repository name (string, required)
  - `pull_number`: Pull request number (number, required)
  - `expected_head_sha`: The expected SHA of the pull request's HEAD ref (string, optional)

- **get_pull_request_comments** - Get the review comments on a pull request

  - `owner`: Repository owner (string, required)
  - `repo`: Repository name (string, required)
  - `pull_number`: Pull request number (number, required)

- **get_pull_request_reviews** - Get the reviews on a pull request

  - `owner`: Repository owner (string, required)
  - `repo`: Repository name (string, required)
  - `pull_number`: Pull request number (number, required)

- **create_pull_request_review** - Create a review on a pull request review

  - `owner`: Repository owner (string, required)
  - `repo`: Repository name (string, required)
  - `pull_number`: Pull request number (number, required)
  - `body`: Review comment text (string, optional)
  - `event`: Review action ('APPROVE', 'REQUEST_CHANGES', 'COMMENT') (string, required)
  - `commit_id`: SHA of commit to review (string, optional)
  - `comments`: Line-specific comments array of objects, each object with path (string), position (number), and body (string) (array, optional)

### Repositories

- **create_or_update_file** - Create or update a single file in a repository

  - `owner`: Repository owner (string, required)
  - `repo`: Repository name (string, required)
  - `path`: File path (string, required)
  - `message`: Commit message (string, required)
  - `content`: File content (string, required)
  - `branch`: Branch name (string, optional)
  - `sha`: File SHA if updating (string, optional)

- **push_files** - Push multiple files in a single commit

  - `owner`: Repository owner (string, required)
  - `repo`: Repository name (string, required)
  - `branch`: Branch to push to (string, required)
  - `files`: Files to push, each with path and content (array, required)
  - `message`: Commit message (string, required)

- **search_repositories** - Search for GitHub repositories

  - `query`: Search query (string, required)
  - `sort`: Sort field (string, optional)
  - `order`: Sort order (string, optional)
  - `page`: Page number (number, optional)
  - `per_page`: Results per page (number, optional)

- **create_repository** - Create a new GitHub repository

  - `name`: Repository name (string, required)
  - `description`: Repository description (string, optional)
  - `private`: Whether the repository is private (boolean, optional)
  - `auto_init`: Auto-initialize with README (boolean, optional)
  - `gitignore_template`: Gitignore template name (string, optional)

- **get_file_contents** - Get contents of a file or directory

  - `owner`: Repository owner (string, required)
  - `repo`: Repository name (string, required)
  - `path`: File path (string, required)
  - `ref`: Git reference (string, optional)

- **fork_repository** - Fork a repository

  - `owner`: Repository owner (string, required)
  - `repo`: Repository name (string, required)
  - `organization`: Target organization name (string, optional)

- **create_branch** - Create a new branch

  - `owner`: Repository owner (string, required)
  - `repo`: Repository name (string, required)
  - `branch`: New branch name (string, required)
  - `sha`: SHA to create branch from (string, required)

- **list_commits** - Gets commits of a branch in a repository
  - `owner`: Repository owner (string, required)
  - `repo`: Repository name (string, required)
  - `sha`: Branch name, tag, or commit SHA (string, optional)
  - `path`: Only commits containing this file path (string, optional)
  - `page`: Page number (number, optional)
  - `per_page`: Results per page (number, optional)

### Search

- **search_code** - Search for code across GitHub repositories

  - `query`: Search query (string, required)
  - `sort`: Sort field (string, optional)
  - `order`: Sort order (string, optional)
  - `page`: Page number (number, optional)
  - `per_page`: Results per page (number, optional)

- **search_users** - Search for GitHub users
  - `query`: Search query (string, required)
  - `sort`: Sort field (string, optional)
  - `order`: Sort order (string, optional)
  - `page`: Page number (number, optional)
  - `per_page`: Results per page (number, optional)

### Code Scanning

- **get_code_scanning_alert** - Get a code scanning alert

  - `owner`: Repository owner (string, required)
  - `repo`: Repository name (string, required)
  - `alert_number`: Alert number (number, required)

- **list_code_scanning_alerts** - List code scanning alerts for a repository
  - `owner`: Repository owner (string, required)
  - `repo`: Repository name (string, required)
  - `ref`: Git reference (string, optional)
  - `state`: Alert state (string, optional)
  - `severity`: Alert severity (string, optional)

## Resources

### Repository Content

- **Get Repository Content**  
  Retrieves the content of a repository at a specific path.

  - **Template**: `repo://{owner}/{repo}/contents{/path*}`
  - **Parameters**:
    - `owner`: Repository owner (string, required)
    - `repo`: Repository name (string, required)
    - `path`: File or directory path (string, optional)

- **Get Repository Content for a Specific Branch**  
  Retrieves the content of a repository at a specific path for a given branch.

  - **Template**: `repo://{owner}/{repo}/refs/heads/{branch}/contents{/path*}`
  - **Parameters**:
    - `owner`: Repository owner (string, required)
    - `repo`: Repository name (string, required)
    - `branch`: Branch name (string, required)
    - `path`: File or directory path (string, optional)

- **Get Repository Content for a Specific Commit**  
  Retrieves the content of a repository at a specific path for a given commit.

  - **Template**: `repo://{owner}/{repo}/sha/{sha}/contents{/path*}`
  - **Parameters**:
    - `owner`: Repository owner (string, required)
    - `repo`: Repository name (string, required)
    - `sha`: Commit SHA (string, required)
    - `path`: File or directory path (string, optional)

- **Get Repository Content for a Specific Tag**  
  Retrieves the content of a repository at a specific path for a given tag.

  - **Template**: `repo://{owner}/{repo}/refs/tags/{tag}/contents{/path*}`
  - **Parameters**:
    - `owner`: Repository owner (string, required)
    - `repo`: Repository name (string, required)
    - `tag`: Tag name (string, required)
    - `path`: File or directory path (string, optional)

- **Get Repository Content for a Specific Pull Request**  
  Retrieves the content of a repository at a specific path for a given pull request.

  - **Template**: `repo://{owner}/{repo}/refs/pull/{pr_number}/head/contents{/path*}`
  - **Parameters**:
    - `owner`: Repository owner (string, required)
    - `repo`: Repository name (string, required)
    - `pr_number`: Pull request number (string, required)
    - `path`: File or directory path (string, optional)

## Standard input/output server

```sh
go run cmd/github-mcp-server/main.go stdio
```

E.g:

Set the PAT token in the environment variable and run:

```sh
script/get-me
```

And you should see the output of the GitHub MCP server responding with the user information.

```sh
GitHub MCP Server running on stdio
{
  "jsonrpc": "2.0",
  "id": 3,
  "result": {
    "content": [
      {
        "type": "text",
        "text": "{\"login\":\"juruen\",\"id\" ... }
      }
    ]
  }
}

```

## i18n / Overriding descriptions

The descriptions of the tools can be overridden by creating a github-mcp-server.json file in the same directory as the binary.
The file should contain a JSON object with the tool names as keys and the new descriptions as values.
For example:

```json
{
  "TOOL_ADD_ISSUE_COMMENT_DESCRIPTION": "an alternative description",
  "TOOL_CREATE_BRANCH_DESCRIPTION": "Create a new branch in a GitHub repository"
}
```

You can create an export of the current translations by running the binary with the `--export-translations` flag.
This flag will preserve any translations/overrides you have made, while adding any new translations that have been added to the binary since the last time you exported.

```sh
./github-mcp-server --export-translations
cat github-mcp-server.json
```

You can also use ENV vars to override the descriptions. The environment variable names are the same as the keys in the JSON file,
prefixed with `GITHUB_MCP_` and all uppercase.

For example, to override the `TOOL_ADD_ISSUE_COMMENT_DESCRIPTION` tool, you can set the following environment variable:

```sh
export GITHUB_MCP_TOOL_ADD_ISSUE_COMMENT_DESCRIPTION="an alternative description"
```

## TODO

Testing

- Integration tests
- Blackbox testing: ideally comparing output to Anthropic's server to make sure that this is a fully compatible drop-in replacement.
