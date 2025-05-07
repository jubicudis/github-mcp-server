// WHO: GithubPackage
// WHAT: Module definition for GitHub client implementation
// WHEN: Build time and dependency resolution
// WHERE: Layer 6 (System Integration)
// WHY: To provide GitHub API integration
// HOW: Using Go modules system
// EXTENT: All GitHub client functionality
module github.com/jubicudis/github-mcp-server/pkg/github

go 1.22.0

require (
	github.com/google/go-github/v69 v69.2.0
	github.com/mark3labs/mcp-go v0.18.0
	github.com/sirupsen/logrus v1.9.3
	github.com/jubicudis/github-mcp-server/pkg/log v0.0.0-00010101000000-000000000000
	github.com/jubicudis/github-mcp-server/pkg/translations v0.0.0-00010101000000-000000000000
	golang.org/x/oauth2 v0.30.0
)

// Local module references
replace (
	github.com/jubicudis/github-mcp-server/pkg/log => ../log
	github.com/jubicudis/github-mcp-server/pkg/translations => ../translations
	github.com/github/github-mcp-server/pkg/log => ../log
	github.com/github/github-mcp-server/pkg/translations => ../translations
)
