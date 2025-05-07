module github.com/jubicudis/github-mcp-server

go 1.22.0

// WHO: DependencyManager
// WHAT: Direct project dependencies
// WHEN: Build and runtime
// WHERE: System Layer 6 (Integration)
// WHY: Core functionality requirements
// HOW: Using semantic versioning
// EXTENT: All required external libraries
require (
	github.com/docker/docker v28.0.4+incompatible
	github.com/google/go-cmp v0.7.0
	github.com/google/go-github/v69 v69.2.0
	github.com/gorilla/websocket v1.5.3
	github.com/mark3labs/mcp-go v0.18.0
	github.com/migueleliasweb/go-github-mock v1.1.0
	github.com/shurcooL/graphql v0.0.0-20230722043721-ed46e5a46466
	github.com/sirupsen/logrus v1.9.3
	github.com/spf13/cobra v1.9.1
	github.com/spf13/viper v1.20.1
	github.com/stretchr/testify v1.10.0
	golang.org/x/oauth2 v0.30.0
)

// WHO: ImportResolver
// WHAT: Local path mappings
// WHEN: During module resolution
// WHERE: System Layer 6 (Integration)
// WHY: To handle internal package references
// HOW: Using Go module replace directives
// EXTENT: Internal packages only
replace (
	github.com/jubicudis/github-mcp-server/pkg/github => ./pkg/github
	github.com/jubicudis/github-mcp-server/pkg/log => ./pkg/log
	github.com/jubicudis/github-mcp-server/pkg/translations => ./pkg/translations
)
