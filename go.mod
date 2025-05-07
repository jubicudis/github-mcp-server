// WHO: DependencyManager
// WHAT: Module definition for GitHub MCP server
// WHEN: Build time and dependency resolution
// WHERE: System Layer 6 (Integration)
// WHY: To support proper Go module structure
// HOW: Using Go module system
// EXTENT: All GitHub MCP functionality
module github.com/jubicudis/Tranquility-Neuro-OS/github-mcp-server

go 1.23.7

toolchain go1.24.2

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
	github.com/mark3labs/mcp-go v0.20.1
	github.com/sirupsen/logrus v1.9.3
	github.com/spf13/cobra v1.9.1
	github.com/spf13/viper v1.20.1
	github.com/stretchr/testify v1.10.0
)

require (
	github.com/jubicudis/Tranquility-Neuro-OS/github-mcp-server/models v0.0.0-00010101000000-000000000000
	github.com/jubicudis/Tranquility-Neuro-OS/github-mcp-server/pkg/github v0.0.0-00010101000000-000000000000
	github.com/jubicudis/Tranquility-Neuro-OS/github-mcp-server/pkg/log v0.0.0-00010101000000-000000000000
	github.com/jubicudis/Tranquility-Neuro-OS/github-mcp-server/pkg/translations v0.0.0-00010101000000-000000000000
	github.com/jubicudis/Tranquility-Neuro-OS/github-mcp-server/utils v0.0.0-00010101000000-000000000000
)

require (
	github.com/Azure/go-ansiterm v0.0.0-20250102033503-faa5f7b0171c // indirect
	github.com/Microsoft/go-winio v0.6.2 // indirect
	github.com/cenkalti/backoff/v4 v4.3.0 // indirect
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc // indirect
	github.com/distribution/reference v0.6.0 // indirect
	github.com/docker/go-connections v0.5.0 // indirect
	github.com/docker/go-units v0.5.0 // indirect
	github.com/felixge/httpsnoop v1.0.4 // indirect
	github.com/fsnotify/fsnotify v1.8.0 // indirect
	github.com/go-logr/logr v1.4.2 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/go-viper/mapstructure/v2 v2.2.1 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/google/go-querystring v1.1.0 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.26.3 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/moby/docker-image-spec v1.3.1 // indirect
	github.com/opencontainers/go-digest v1.0.0 // indirect
	github.com/opencontainers/image-spec v1.1.1 // indirect
	github.com/pelletier/go-toml/v2 v2.2.3 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/pmezard/go-difflib v1.0.1-0.20181226105442-5d4384ee4fb2 // indirect
	github.com/sagikazarmark/locafero v0.9.0 // indirect
	github.com/shurcooL/graphql v0.0.0-20230722043721-ed46e5a46466 // indirect
	github.com/sourcegraph/conc v0.3.0 // indirect
	github.com/spf13/afero v1.14.0 // indirect
	github.com/spf13/cast v1.7.1 // indirect
	github.com/spf13/pflag v1.0.6 // indirect
	github.com/subosito/gotenv v1.6.0 // indirect
	github.com/yosida95/uritemplate/v3 v3.0.2 // indirect
	go.opentelemetry.io/auto/sdk v1.1.0 // indirect
	go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.49.0 // indirect
	go.opentelemetry.io/otel v1.35.0 // indirect
	go.opentelemetry.io/otel/metric v1.35.0 // indirect
	go.opentelemetry.io/otel/trace v1.35.0 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	golang.org/x/oauth2 v0.30.0 // indirect
	golang.org/x/sys v0.31.0 // indirect
	golang.org/x/text v0.23.0 // indirect
	google.golang.org/grpc v1.72.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

// WHO: ImportResolver
// WHAT: Local path mappings
// WHEN: During module resolution
// WHERE: System Layer 6 (Integration)
// WHY: To handle internal package references
// HOW: Using Go module replace directives
// EXTENT: Internal packages only
replace (
	github.com/jubicudis/Tranquility-Neuro-OS/github-mcp-server => ./
	github.com/jubicudis/Tranquility-Neuro-OS/github-mcp-server/models => ./models
	github.com/jubicudis/Tranquility-Neuro-OS/github-mcp-server/pkg/github => ./pkg/github
	github.com/jubicudis/Tranquility-Neuro-OS/github-mcp-server/pkg/log => ./pkg/log
	github.com/jubicudis/Tranquility-Neuro-OS/github-mcp-server/pkg/translations => ./pkg/translations
	github.com/jubicudis/Tranquility-Neuro-OS/github-mcp-server/utils => ./utils
)
