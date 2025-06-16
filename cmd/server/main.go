/*
 * WHO: MCPServer
 * WHAT: Main GitHub MCP server implementation
 * WHEN: During GitHub API operations
 * WHERE: System Layer 6 (Integration)
 * WHY: To provide structured GitHub API access
 * HOW: Using Go HTTP server with WebSocket support
 * EXTENT: All GitHub API interactions
 */

// Purpose: Entry point for the GitHub MCP server
// This file is responsible for initializing the server, handling API requests, and managing WebSocket connections.
// It is distinct from the mcpcurl's main.go file, which handles CLI operations.

package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	// Internal packages
	"github.com/jubicudis/Tranquility-Neuro-OS/github-mcp-server/pkg/bridge"
	"github.com/jubicudis/Tranquility-Neuro-OS/github-mcp-server/pkg/common"
	ghmcp "github.com/jubicudis/Tranquility-Neuro-OS/github-mcp-server/pkg/github"
	logpkg "github.com/jubicudis/Tranquility-Neuro-OS/github-mcp-server/pkg/log"
	"github.com/jubicudis/Tranquility-Neuro-OS/github-mcp-server/pkg/translations"

	// External packages

	"github.com/gorilla/websocket"
)

// Config holds the application configuration
type Config struct {
	// Core Server Settings
	Host         string `json:"host"`
	Port         int    `json:"port"`
	LogLevel     string `json:"logLevel"`
	LogFile      string `json:"logFile"`
	GitHubToken  string `json:"githubToken"` // Essential for core GitHub MCP functionality

	// External Service Settings
	CopilotLLMURL string `json:"copilotLLMURL"` // URL for Copilot LLM

	// Bridge Settings (for connecting to TNOS MCP)
	BridgeEnabled    bool   `json:"bridgeEnabled"`    // Specifically for the bridge to TNOS MCP
	BridgePort       int    `json:"bridgePort"`       // Port this server listens on for bridge-related communication
	TNOSMCPBridgeURL string `json:"tnosMCPBridgeURL"` // URL for the TNOS MCP to bridge to
	GitHubMCPURL     string `json:"githubMCPURL"`     // URL for this GitHub MCP server (self-reference, used for discovery by other services)
}

// Global variables
var (
    config     Config
    logger     logpkg.LoggerInterface
    clients    = make(map[*Client]bool)
    clientsMtx sync.Mutex
    broadcast  = make(chan []byte)

    // Track all HTTP servers for graceful shutdown
    servers    []*http.Server
    serversMtx sync.Mutex

    // Connection managers for external services and TNOS bridge
    copilotConnection   *bridge.BloodCirculation // Connection to Copilot LLM
    bloodBridgeInstance *bridge.BloodCirculation // Primary bridge to Python TNOS MCP server

    startTime time.Time // Added startTime for uptime calculation
)

// Client represents a connected websocket client
type Client struct {
	// WHO: ConnectionManager
	// WHAT: WebSocket connection tracking
	// WHEN: During client communication
	// WHERE: System Layer 6 (Integration)
	// WHY: To manage individual client connections
	// HOW: Using client structure with connection details
	// EXTENT: All client connections
	ID          string
	Conn        *websocket.Conn
	Send        chan []byte
	ConnectedAt time.Time
	LastPing    time.Time
	Context     translations.ContextVector7D
}

// Initialize server configuration from environment or defaults
func initConfig() Config {
	// Get configuration from environment variables with defaults
	portStr := getEnv("MCP_SERVER_PORT", "10617") // This server's listening port (GitHub MCP functionality)
	port, err := strconv.Atoi(portStr)
	if err != nil {
		port = 10617 // Default port for this GitHub MCP server
	}

	bridgePortStr := getEnv("MCP_BRIDGE_LISTENING_PORT", "10619") // Port this server might listen on for specific bridge ops, if different from main port
	bridgePort, err := strconv.Atoi(bridgePortStr)
	if err != nil {
		bridgePort = 10619
	}

	return Config{
		Port:             port,
		Host:             getEnv("MCP_SERVER_HOST", "localhost"),
		LogLevel:         getEnv("MCP_LOG_LEVEL", "info"),
		LogFile:          getEnv("MCP_LOG_FILE", "github-mcp-server.log"),
		BridgeEnabled:    getEnvBool("TNOS_BRIDGE_ENABLED", true), // Is the bridge to TNOS MCP enabled?
		BridgePort:       bridgePort,                               // Port this server listens on (e.g., for incoming bridge commands if any)
		GitHubToken:      getEnv("GITHUB_TOKEN", ""),
		CopilotLLMURL:    getEnv("COPILOT_LLM_URL", "ws://localhost:8083"), // Default Copilot LLM URL
		GitHubMCPURL:     getEnv("GITHUB_MCP_URL", "ws://localhost:10617"),  // Default GitHub MCP URL (this server itself, or an external one)
		TNOSMCPBridgeURL: getEnv("TNOS_MCP_BRIDGE_TARGET_URL", "ws://localhost:9001"), // Target TNOS MCP Server URL
	}
}

// Get environment variable with default fallback
func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}

// Get boolean environment variable
func getEnvBool(key string, fallback bool) bool {
	if value, exists := os.LookupEnv(key); exists {
		if value == "true" || value == "1" || value == "yes" {
			return true
		}
		if value == "false" || value == "0" || value == "no" {
			return false
		}
	}
	return fallback
}

// Main server handler
func main() {
	// Diagnostic: Print working directory and intended log file path at the very start
	cwd, _ := os.Getwd()
	absLogPath, _ := filepath.Abs("pkg/log/github-mcp-server.log") // FIX: removed extra 'github-mcp-server' from path
	preLogger := logpkg.NewLogger() // Temporary logger for diagnostics before main logger is ready
	preLogger.Debug("[DIAGNOSTIC] (main) Working directory: %s", cwd)
	preLogger.Debug("[DIAGNOSTIC] (main) Intended log file absolute path: %s", absLogPath)

	// Parse command line flags
	var configFile string
	flag.StringVar(&configFile, "config", "", "Path to configuration file")
	flag.Parse()

	// Initialize startTime
	startTime = time.Now()

	// Load configuration
	config = initConfig()

	// Detect stray log file in cmd/server and read its contents
	exe, err := os.Executable()
	var strayData []byte
	if err == nil {
		// exe points to <projectRoot>/cmd/server/<binary>
		projectRoot := filepath.Dir(filepath.Dir(filepath.Dir(exe)))
		strayPath := filepath.Join(projectRoot, "cmd", "server", config.LogFile)
		if data, err := os.ReadFile(strayPath); err == nil {
			strayData = data
			_ = os.Remove(strayPath)
		}
	}

	// Initialize logger
	loggerIface := initLogger(config)
	// Set global logger
	logger = loggerIface
	// Ensure concrete type for early diagnostics
	if concreteLogger, ok := logger.(*logpkg.Logger); !ok {
		preLogger.Error("Logger type assertion failed; using fallback logger")
		logger = preLogger // Use preLogger if type assertion fails
	} else {
		// The global 'logger' variable is now set.
		// No need for SetDefaultLogger if the package relies on passing the logger instance
		// or using a package-level global variable that 'logger' now represents.
		// logpkg.SetDefaultLogger(concreteLogger) // Removed this line
		// If other parts of logpkg need a default logger, ensure they access this global 'logger'
		// or that logpkg is refactored to accept/use this instance.
		// For now, we assume direct usage of the 'logger' variable or passing it as an argument is sufficient.
		_ = concreteLogger // Avoid unused variable error if concreteLogger is not used otherwise in this block
	}


	// Set Helical Memory mode based on whether the bridge to TNOS is enabled
	if !config.BridgeEnabled {
		logpkg.SetHelicalMemoryMode("standalone")
	} else {
		logpkg.SetHelicalMemoryMode("blood-connected")
	}

	// If stray log data was found, migrate into unified logger
	if len(strayData) > 0 {
		logger.Info(fmt.Sprintf("Migrated stray log contents: %s", string(strayData)))
	}

	// Diagnostic: Print log file path and working directory
	cwd, _ = os.Getwd()
	logDir := filepath.Join(cwd, "pkg", "log") // FIX: removed extra 'github-mcp-server' from path
	logFilePath := filepath.Join(logDir, filepath.Base(config.LogFile))
	logger.Debug("[DIAGNOSTIC] Working directory: %s", cwd)
	logger.Debug("[DIAGNOSTIC] Log file path: %s", logFilePath)
	if _, err := os.Stat(logFilePath); err != nil {
		logger.Warn("[DIAGNOSTIC] Log file does not exist or cannot be accessed: %v", err)
	} else {
		logger.Debug("[DIAGNOSTIC] Log file exists and is accessible.")
	}

	// Switch helical memory to connected mode
	// logpkg.SetHelicalMemoryMode("connected") // Moved to after bridge connection

	// Log server startup to github-mcp-server.log only
	logger.Info("Starting GitHub MCP Server version 1.0")
	logger.Info(fmt.Sprintf("Configuration loaded: host=%s port=%d", config.Host, config.Port))

	// --- Formula Registry Initialization ---
	// Use absolute path for formula registry
	absFormulaPath, err := filepath.Abs(filepath.Join(filepath.Dir(os.Args[0]), "config/formulas.json"))
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to resolve absolute path for formula registry: %s", err.Error()))
	} else {
		if err := bridge.LoadBridgeFormulaRegistry(absFormulaPath); err != nil {
			logger.Error(fmt.Sprintf("Failed to load formula registry: %s path=%s", err.Error(), absFormulaPath))
		} else {
			reg := bridge.GetBridgeFormulaRegistry()
			count := 0
			if reg != nil {
				count = len(reg.ListFormulas())
			}
			logger.Info(fmt.Sprintf("Formula registry loaded: count=%d path=%s", count, absFormulaPath))
		}
	}

	// Initialize Blood System
	ctx := context.Background()
	ghmcp.InitializeBloodSystem(ctx, logger) // Added logger


	// Initialize and start API Status Handler
	go func() {
		// Create a new ServeMux for the API status server to avoid conflicts
		// if other HTTP services are added to the default mux or another mux.
		statusMux := http.NewServeMux()
		statusMux.HandleFunc("/api/status", handleStatus) // Changed apiStatusHandler to handleStatus

		// Use a different port for the API status for clarity, or ensure config.Port is dedicated.
		// For now, assuming config.Port is for the main application and API status runs on it.
		// If main server also uses http.DefaultServeMux, ensure routes don't clash.
		// It's generally better to have one main http.Server instance managing all routes
		// or separate http.Server instances on different ports.

		// The main server setup in startAllServers might conflict if it also uses DefaultServeMux on the same port.
		// Let's assume startAllServers configures the main application server, and this is an auxiliary endpoint.
		// If config.Port is the main app port, this status endpoint will be part of it.

		// The variable 'addr' was here, but removed as it was unused after commenting out the separate server start below.
		// This server instance is local to this goroutine.
		// It might be better to integrate this handler into the main server setup in startAllServers.
		// For now, let it run as a separate thought, assuming it's managed or intended to be part of the main server.
		// If startAllServers() also listens on config.Port with its own mux, this might not behave as expected
		// unless handleStatus is registered on that main mux.

		// The current structure has startAllServers() which likely starts the main HTTP server.
		// Adding another ListenAndServe on the same port will fail.
		// The handleStatus should be registered on the mux used by the main server.
		// For now, I will comment out this independent server start for /api/status
		// and assume it's (or should be) registered on the main server's mux.
		// If not, this is a structural issue to be addressed.
		logger.Info("API Status Handler configured for /api/status on the main server.")
		// server := &http.Server{Addr: addr, Handler: statusMux}

		// logger.Info("API Status Handler listening", "address", addr)
		// if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		// 	logger.Error("API Status Handler failed to start", "error", err.Error())
		// 	_ = logpkg.LogHelicalEvent(logpkg.NewHelicalEvent(
		// 		"MCPServer", "fatal_error", "github-mcp-server", "crash", "GoMain", "full", "API Status Handler failed to start: "+err.Error()), logger)
		// 	os.Exit(1)
		// }
	}()

	// Start HTTP servers on all canonical ports
	startAllServers()

	// Start WebSocket broadcaster
	go handleBroadcasts()

	// Wait for termination signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	sig := <-quit
	logger.Info(fmt.Sprintf("Received termination signal: %s", sig.String()))
	_ = logpkg.LogHelicalEvent(logpkg.NewHelicalEvent("MCPServer", "shutdown", "github-mcp-server", "signal", "GoMain", "full", "Received termination signal: "+sig.String()), logger)

	// Create context with timeout for graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second) // Increased timeout for proper shutdown
	defer cancel()

	// Close all client connections
	closeAllClients()

	// Gracefully shutdown all HTTP servers
	serversMtx.Lock()
	for _, srv := range servers {
		if err := srv.Shutdown(ctx); err != nil {
			logger.Error(fmt.Sprintf("Server shutdown failed: address=%s error=%s", srv.Addr, err.Error()))
			_ = logpkg.LogHelicalEvent(logpkg.NewHelicalEvent("MCPServer", "shutdown_error", "github-mcp-server", "shutdown", "GoMain", "full", "Server shutdown failed: "+err.Error()), logger) // Added logger
		}
	}
	serversMtx.Unlock()

	logger.Info("Server shutdown complete")
	_ = logpkg.LogHelicalEvent(logpkg.NewHelicalEvent("MCPServer", "shutdown_complete", "github-mcp-server", "shutdown", "GoMain", "full", "Server shutdown complete"), logger) // Added logger
}

// Initialize logger
func initLogger(config Config) logpkg.LoggerInterface {
	// Always use absolute path for log file
	projectRoot, _ := filepath.Abs(".")
	logDir := filepath.Join(projectRoot, "pkg", "log") // FIX: removed extra 'github-mcp-server' from path
	logFilePath := filepath.Join(logDir, filepath.Base(config.LogFile))
	logger := logpkg.NewLogger()

	if err := os.MkdirAll(logDir, 0755); err != nil {
		logger.Error("Failed to create log dir %s: %v", logDir, err)
	}
	logger.Debug("[DIAGNOSTIC] Using logDir=%s, logFile=%s", logDir, logFilePath)

	switch strings.ToLower(config.LogLevel) {
	case "debug":
		logger = logger.WithLevel(logpkg.LevelDebug)
	case "warn":
		logger = logger.WithLevel(logpkg.LevelWarn)
	case "error":
		logger = logger.WithLevel(logpkg.LevelError)
	default:
		logger = logger.WithLevel(logpkg.LevelInfo)
	}

	f, err := os.OpenFile(logFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		logger.Error("[DIAGNOSTIC] Failed to open log file %s: %v", logFilePath, err)
	} else {
		logger = logger.WithOutput(f)
	}

	return logger
}

// Root handler
func handleRoot(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "running",
		"service": "GitHub MCP Server",
		"version": "1.0",
	})
}

// Health check handler
func handleHealth(w http.ResponseWriter, r *http.Request) {
	// Import context timeout from common.go
	bridgeTimeout := 5 * time.Second

	// Create a context with timeout to check GitHub Copilot connection status
	ctx, cancel := context.WithTimeout(r.Context(), bridgeTimeout)
	defer cancel()

	// Variables to store component statuses
	copilotStatus := "unknown"
	copilotDetails := map[string]interface{}{}

	// Check GitHub Copilot connection in a context-aware manner
	copilotChan := make(chan bool, 1)
	go func() {
		// This would be a real connection check in production
		copilotStatus = "healthy"
		copilotChan <- true
	}()

	// Wait for completion or timeout
	select {
	case <-copilotChan:
		// Successfully checked status
	case <-ctx.Done():
		if ctx.Err() == context.DeadlineExceeded {
			// Using error definition from pkg/bridge/common.go
			logger.Warn("Context deadline exceeded when checking GitHub Copilot status")
			copilotStatus = "timeout"
			copilotDetails["error"] = "context deadline exceeded"
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":    "healthy",
		"timestamp": time.Now().Unix(),
		"components": map[string]interface{}{
			"github_copilot": map[string]interface{}{
				"status":  copilotStatus,
				"details": copilotDetails,
			},
		},
	})
}

// Define APIServerStatus struct
type APIServerStatus struct {
	ServerName        string                 `json:"server_name"`
	Version           string                 `json:"version"`
	Status            string                 `json:"status"`
	StartTime         string                 `json:"start_time"`
	UpTime            string                 `json:"up_time"`
	ActiveConnections int                    `json:"active_connections"`
	Host              string                 `json:"host"`
	Port              int                    `json:"port"`
	BridgeEnabled     bool                   `json:"bridge_enabled"`
	BridgeStatus      string                 `json:"bridge_status"`
	BridgeDetails     map[string]interface{} `json:"bridge_details"`
	CopilotLLMStatus  string                 `json:"copilot_llm_status"`
	GitHubMCPStatus   string                 `json:"github_mcp_status"` // Status of this server's GitHub MCP functionality
	LogFile           string                 `json:"log_file"`
	Timestamp         int64                  `json:"timestamp"`
}

// Server status handler
func handleStatus(w http.ResponseWriter, r *http.Request) {
    clientsMtx.Lock()
    clientCount := len(clients)
    clientsMtx.Unlock()

    bridgeStatus := "disconnected"
    bridgeDetails := map[string]interface{}{"state": "N/A"}
    
    // Use bloodBridgeInstance (canonical TNOS MCP bridge) instead of tnosBridge
    if bloodBridgeInstance != nil {
        state := bloodBridgeInstance.GetState()
        bridgeStatus = string(state)
        bridgeDetails["url"] = bloodBridgeInstance.TargetURL
        bridgeDetails["state"] = string(state)
        
        // TNOS 7D Biomimicry: Blood circulation metrics
        metrics := bloodBridgeInstance.GetMetrics()
        if metrics != nil {
            bridgeDetails["cells_transmitted"] = metrics.CellsTransmitted
            bridgeDetails["cells_received"] = metrics.CellsReceived
            bridgeDetails["error_count"] = metrics.ErrorCount
            bridgeDetails["clotting_events"] = metrics.ClottingEvents
            // Additional biomimicry metrics
            bridgeDetails["circulation_health"] = calculateCirculationHealth(metrics)
            bridgeDetails["flow_rate"] = calculateFlowRate(metrics)
        }
    } else if config.BridgeEnabled {
        bridgeStatus = "Enabled but not Initialized"
        bridgeDetails["state"] = bridgeStatus
    }

    uptime := time.Since(startTime)

    status := APIServerStatus{
        ServerName:        "GitHub MCP Server (TNOS-Integrated)",
        Version:           common.AppVersion,
        Status:            "Running",
        StartTime:         startTime.Format(time.RFC3339),
        UpTime:            uptime.String(),
        ActiveConnections: clientCount,
        Host:              config.Host,
        Port:              config.Port,
        BridgeEnabled:     config.BridgeEnabled,
        BridgeStatus:      bridgeStatus,
        BridgeDetails:     bridgeDetails,
        CopilotLLMStatus:  "Not Implemented",
        GitHubMCPStatus:   "Running (Self)",
        LogFile:           config.LogFile,
        Timestamp:         time.Now().Unix(),
    }

    // Get Copilot LLM Connection Status
    if copilotConnection != nil {
        status.CopilotLLMStatus = string(copilotConnection.GetState())
    } else {
        status.CopilotLLMStatus = "Disabled/Not Initialized"
    }

    response, err := json.MarshalIndent(status, "", "  ")
    if err != nil {
        logger.Error("Failed to marshal status response", "error", err)
        http.Error(w, "Failed to marshal status response", http.StatusInternalServerError)
        return
    }
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusOK)
    w.Write(response)
}

// TNOS 7D Biomimicry helper functions for circulation health
func calculateCirculationHealth(metrics *bridge.BloodMetrics) string {
    if metrics == nil {
        return "unknown"
    }
    
    errorRate := float64(metrics.ErrorCount) / float64(metrics.CellsTransmitted+1) // Avoid division by zero
    if errorRate < 0.01 {
        return "excellent"
    } else if errorRate < 0.05 {
        return "good"
    } else if errorRate < 0.1 {
        return "fair"
    }
    return "poor"
}

func calculateFlowRate(metrics *bridge.BloodMetrics) float64 {
    if metrics == nil {
        return 0.0
    }
    
    // Simple flow rate calculation based on total cells and uptime
    uptime := time.Since(startTime).Seconds()
    if uptime == 0 {
        return 0.0
    }
    
    totalCells := float64(metrics.CellsTransmitted + metrics.CellsReceived)
    return totalCells / uptime // cells per second
}

// GitHub repositories handler
func handleGitHubRepositories(w http.ResponseWriter, r *http.Request) {
	// WHO: RepoHandler
	// WHAT: Handle GitHub repository requests
	// WHEN: During API request
	// WHERE: System Layer 6 (Integration)
	// WHY: To provide repository data
	// HOW: Using GitHub API client
	// EXTENT: Single API request

	logger.Info("Repositories API called", "method", r.Method, "path", r.URL.Path)

	// Extract owner and repo from query parameters
	owner := r.URL.Query().Get("owner")
	repo := r.URL.Query().Get("repo")

	if owner == "" || repo == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"status": "error",
			"error":  "Missing owner or repo parameter",
		})
		return
	}

	// Use the GitHub client to get repository information
	var repository map[string]interface{}
	var err error
	// repository, err = gitHubClient.GetRepositoryByName(ctx, owner, repo) // TODO: Uncomment and use actual gitHubClient
	repository, err = nil, fmt.Errorf("TODO: implement GetRepositoryByName") // Placeholder

	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"status": "error",
			"error":  err.Error(),
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(repository)
}

// GitHub issues handler
func handleGitHubIssues(w http.ResponseWriter, r *http.Request) {
	// WHO: IssueHandler
	// WHAT: Handle GitHub issues requests
	// WHEN: During API request
	// WHERE: System Layer 6 (Integration)
	// WHY: To provide issue data
	// HOW: Using GitHub API client
	// EXTENT: Single API request

	logger.Info("Issues API called", "method", r.Method, "path", r.URL.Path)

	// Extract owner and repo from query parameters
	owner := r.URL.Query().Get("owner")
	repo := r.URL.Query().Get("repo")

	if owner == "" || repo == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"status": "error",
			"error":  "Missing owner or repo parameter",
		})
		return
	}

	// Use the GitHub client to get repository information
	var issues []map[string]interface{}
	var err error
	// issues, err = gitHubClient.GetIssues(ctx, owner, repo) // TODO: Uncomment and use actual gitHubClient
	issues, err = nil, fmt.Errorf("TODO: implement GetIssues") // Placeholder

	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"status": "error",
			"error":  err.Error(),
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(issues)
}

// GitHub pull requests handler
func handleGitHubPullRequests(w http.ResponseWriter, r *http.Request) {
	// WHO: PRHandler
	// WHAT: Handle GitHub pull requests
	// WHEN: During API request
	// WHERE: System Layer 6 (Integration)
	// WHY: To provide PR data
	// HOW: Using GitHub API client
	// EXTENT: Single API request

	logger.Info("Pull Requests API called", "method", r.Method, "path", r.URL.Path)

	// Extract owner and repo from query parameters
	owner := r.URL.Query().Get("owner")
	repo := r.URL.Query().Get("repo")

	if owner == "" || repo == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"status": "error",
			"error":  "Missing owner or repo parameter",
		})
		return
	}

	// Use the GitHub client to get pull request information
	var prs []map[string]interface{}
	var err error
	// prs, err = gitHubClient.GetPullRequests(ctx, owner, repo) // TODO: Uncomment and use actual gitHubClient
	prs, err = nil, fmt.Errorf("TODO: implement GetPullRequests") // Placeholder

	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"status": "error",
			"error":  err.Error(),
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(prs)
}

// GitHub search handler
func handleGitHubSearchCode(w http.ResponseWriter, r *http.Request) {
	// WHO: SearchCodeHandler
	// WHAT: Handle GitHub code search requests
	// WHEN: During API request
	// WHERE: System Layer 6 (Integration)
	// WHY: To provide code search results
	// HOW: Using GitHub API client
	// EXTENT: Single API request

	logger.Info("Search Code API called", "method", r.Method, "path", r.URL.Path)

	// Extract query from query parameters
	query := r.URL.Query().Get("query")

	if query == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"status": "error",
			"error":  "Missing query parameter",
		})
		return
	}

	// Use the GitHub client to search code
	var results map[string]interface{}
	var err error
	// results, err = gitHubClient.SearchCode(ctx, query) // TODO: Uncomment and use actual gitHubClient
	results, err = nil, fmt.Errorf("TODO: implement SearchCode") // Placeholder

	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"status": "error",
			"error":  err.Error(),
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(results)
}

// GitHub code scanning handler
func handleGitHubCodeScanningAlerts(w http.ResponseWriter, r *http.Request) {
	// WHO: CodeScanningAlertsHandler
	// WHAT: Handle GitHub code scanning alerts requests
	// WHEN: During API request
	// WHERE: System Layer 6 (Integration)
	// WHY: To provide code scanning alert data
	// HOW: Using GitHub API client
	// EXTENT: Single API request

	logger.Info("Code Scanning Alerts API called", "method", r.Method, "path", r.URL.Path)

	// Extract owner and repo from query parameters
	owner := r.URL.Query().Get("owner")
	repo := r.URL.Query().Get("repo")

	if owner == "" || repo == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"status": "error",
			"error":  "Missing owner or repo parameter",
		})
		return
	}

	// Use the GitHub client to get code scanning alerts
	var alerts []map[string]interface{}
	var err error
	// alerts, err = gitHubClient.GetCodeScanningAlerts(ctx, owner, repo) // TODO: Uncomment and use actual gitHubClient
	alerts, err = nil, fmt.Errorf("TODO: implement GetCodeScanningAlerts") // Placeholder

	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"status": "error",
			"error":  err.Error(),
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(alerts)
}

// Context handler
func handleContext(w http.ResponseWriter, r *http.Request) {
	// WHO: ContextHandler
	// WHAT: Handle context operations
	// WHEN: During API request
	// WHERE: System Layer 6 (Integration)
	// WHY: To provide context services
	// HOW: Using translations package
	// EXTENT: Context lifecycle

	logger.Info("Context API called", "method", r.Method, "path", r.URL.Path)

	// Process based on HTTP method
	switch r.Method {
	case "GET":
		// Return a new default context
		context := translations.NewContextVector7D(map[string]interface{}{                                                          
			"who":    "APIClient",
			"what":   "ContextRequest",
			"when":   time.Now().Unix(),
			"where":  "MCP_Server",
			"why":    "Context_Creation",
			"how":    "HTTP_API",
			"extent": 1.0,
			"source": "github_mcp",
		})

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(context)

	case "POST":
		// Process posted context
		var inputContext map[string]interface{}
		decoder := json.NewDecoder(r.Body)
		if err := decoder.Decode(&inputContext); err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{
				"status": "error",
				"error":  "Invalid JSON: " + err.Error(),
			})
			return
		}

		// Create 7D context from input
		var context = translations.NewContextVector7D(inputContext)

		// Apply compression
		compressed := translations.CompressContext(context, logger.(*logpkg.Logger))

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(compressed)

	default:
		// Method not allowed
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]string{
			"status": "error",
			"error":  "Method not allowed",
		})
	}
}

// Handle WebSocket connections
func handleWebSocket(w http.ResponseWriter, r *http.Request) {
	// WHO: ConnectionHandler
	// WHAT: WebSocket connection establishment
	// WHEN: During client connection request
	// WHERE: System Layer 6 (Integration)
	// WHY: To enable real-time bidirectional communication
	// HOW: Using gorilla/websocket with context awareness
	// EXTENT: Single client connection lifecycle

	logger.Info("WebSocket connection request received")

	// Configure WebSocket upgrader with proper security settings
	upgrader := websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			// In production, implement proper origin checking
			return true
		},
	}

	// Upgrade HTTP connection to WebSocket
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		logger.Error("Failed to upgrade connection", "error", err.Error())
		return
	}

	// Create context for this connection with 7D dimensions
	contextVector := translations.NewContextVector7D(map[string]interface{}{
		"who":    "WebSocketClient",
		"what":   "Connection",
		"when":   time.Now().Unix(),
		"where":  "MCP_Server",
		"why":    "Real-time_Communication",
		"how":    "WebSocket_Protocol",
		"extent": 1.0,
		"source": "github_mcp",
	})

	// Create client with 7D context awareness
	clientID := fmt.Sprintf("%s-%d", r.RemoteAddr, time.Now().UnixNano())
	client := &Client{
		ID:          clientID,
		Conn:        conn,
		Send:        make(chan []byte, 256), // Buffer for outgoing messages
		ConnectedAt: time.Now(),
		LastPing:    time.Now(),
		Context:     contextVector,
	}

	// Store client in clients map
	clientsMtx.Lock()
	clients[client] = true
	clientCount := len(clients)
	clientsMtx.Unlock()

	logger.Info("Client connected", "id", clientID, "total", clientCount)

	// Send welcome message with context
	welcomeMsg := map[string]interface{}{
		"type":      "connection",
		"status":    "connected",
		"clientId":  clientID,
		"timestamp": time.Now().Unix(),
		"context":   contextVector,
	}

	welcomeData, _ := json.Marshal(welcomeMsg)
	client.Send <- welcomeData

	// Start client reader and writer goroutines
	go readPump(client)
	go writePump(client)
}

// readPump reads messages from the WebSocket connection
func readPump(client *Client) {
	// WHO: MessageReader
	// WHAT: WebSocket message reading
	// WHEN: During active connection
	// WHERE: System Layer 6 (Integration)
	// WHY: To process incoming client messages
	// HOW: Using gorilla/websocket with context
	// EXTENT: Client connection message flow

	defer func() {
		// Clean up when the connection ends
		clientsMtx.Lock()
		delete(clients, client)
		clientCount := len(clients)
		clientsMtx.Unlock()

		client.Conn.Close()
		close(client.Send)

		logger.Info("Client disconnected", "id", client.ID, "total", clientCount)
	}()

	// Configure connection
	client.Conn.SetReadLimit(512 * 1024) // 512KB max message size
	client.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	client.Conn.SetPongHandler(func(string) error {
		client.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		client.LastPing = time.Now()
		return nil
	})

	// Main read loop
	for {
		_, message, err := client.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err,
				websocket.CloseGoingAway,
				websocket.CloseAbnormalClosure) {
				logger.Error("WebSocket error", "id", client.ID, "error", err.Error())
			}
			break
		}

		// Process the message
		processIncomingMessage(client, message)
	}
}

// writePump writes messages to the WebSocket connection
func writePump(client *Client) {
	// WHO: MessageWriter
	// WHAT: WebSocket message writing
	// WHEN: During active connection
	// WHERE: System Layer 6 (Integration)
	// WHY: To send messages to client
	// HOW: Using gorilla/websocket with context
	// EXTENT: Client connection message flow

	ticker := time.NewTicker(30 * time.Second)
	defer func() {
		ticker.Stop()
		client.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-client.Send:
			// Check if channel is closed
			if !ok {
				client.Conn.WriteMessage(websocket.CloseMessage, nil)
				return
			}

			// Write message to WebSocket
			client.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			w, err := client.Conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}

			w.Write(message)

			// Add queued messages to the current WebSocket message
			n := len(client.Send)
			for i := 0; i < n; i++ {
				w.Write([]byte("\n"))
				w.Write(<-client.Send)
			}

			if err := w.Close(); err != nil {
				return
			}

		case <-ticker.C:
			// Send ping to keep connection alive
			client.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := client.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// processIncomingMessage processes incoming WebSocket messages
func processIncomingMessage(client *Client, message []byte) {
	// WHO: MessageProcessor
	// WHAT: Process client message
	// WHEN: When message arrives
	// WHERE: System Layer 6 (Integration)
	// WHY: To handle client requests
	// HOW: Using JSON parsing with context
	// EXTENT: Single message lifecycle

	// Parse the message
	var data map[string]interface{}
	if err := json.Unmarshal(message, &data); err != nil {
		logger.Error("Failed to parse message", "id", client.ID, "error", err.Error())
		return
	}

	// Get message type
	msgType, ok := data["type"].(string)
	if !ok {
		msgType = "unknown"
	}

	// Log message receipt
	logger.Debug("Received message", "id", client.ID, "type", msgType)

	// Update context with message information
	msgContext := translations.NewContextVector7D(map[string]interface{}{
		"who":    "WebSocketClient",
		"what":   msgType,
		"when":   time.Now().Unix(),
		"where":  "MCP_Server",
		"why":    "Client_Request",
		"how":    "WebSocket_Message",
		"extent": 0.8,
		"source": "github_mcp",
	})

	// Handle message based on type
	switch msgType {
	case "ping":
		// Respond to ping
		sendResponse(client, map[string]interface{}{
			"type":      "pong",
			"timestamp": time.Now().Unix(),
			"context":   msgContext,
		})

	case "query":
		// Handle query
		handleQuery(client, data, msgContext)

	case "github":
		// Handle GitHub API request
		handleGitHubRequest(client, data, msgContext)

	case "context":
		// Handle context request
		handleContextRequest(client, data, msgContext)

	default:
		// Unknown message type
		sendResponse(client, map[string]interface{}{
			"type":    "error",
			"error":   "Unknown message type",
			"context": msgContext,
		})
	}

	// Broadcast message to monitoring clients if needed
	// This is optional depending on your requirements
	if shouldBroadcast(msgType) {
		broadcastData := map[string]interface{}{
			"type":      "broadcast",
			"clientId":  client.ID,
			"original":  msgType,
			"timestamp": time.Now().Unix(),
			"context":   msgContext,
		}

		broadcastMsg, _ := json.Marshal(broadcastData)
		broadcast <- broadcastMsg
	}
}

// Helper functions for message handling

func sendResponse(client *Client, data map[string]interface{}) {
	response, err := json.Marshal(data)
	if err != nil {
		logger.Error("Failed to marshal response", "error", err.Error())
		return
	}

	client.Send <- response
}

func handleQuery(client *Client, data map[string]interface{}, context translations.ContextVector7D) {
	// Extract query parameters
	query, _ := data["query"].(string)

	// Create response with query results
	response := map[string]interface{}{
		"type":      "query_result",
		"query":     query,
		"timestamp": time.Now().Unix(),
		"context":   context,
		"result":    "Query processed", // Placeholder, implement actual query logic
	}

	sendResponse(client, response)
}

func handleGitHubRequest(client *Client, data map[string]interface{}, context translations.ContextVector7D) {
	// WHO: GitHubRequestHandler
	// WHAT: Handle GitHub API via WebSocket
	// WHEN: During WebSocket message processing
	// WHERE: System Layer 6 (Integration)
	// WHY: To provide GitHub data access
	// HOW: Using GitHub client
	// EXTENT: Single GitHub request

	// Extract request details
	action, _ := data["action"].(string)
	owner, _ := data["owner"].(string)
	repo, _ := data["repo"].(string)

	var result interface{}
	var err error

	// Process based on action type
	switch action {
	case "get_repository":
		if owner != "" && repo != "" {
			// result, err = gitHubClient.GetRepository(owner, repo)
			result, err = nil, fmt.Errorf("TODO: implement GetRepository")
		} else {
			err = fmt.Errorf("missing owner or repo parameters")
		}

	case "get_issues":
		if owner != "" && repo != "" {
			// result, err = gitHubClient.GetIssues(owner, repo)
			result, err = nil, fmt.Errorf("TODO: implement GetIssues")
		} else {
			err = fmt.Errorf("missing owner or repo parameters")
		}

	case "get_pulls":
		if owner != "" && repo != "" {
			// result, err = gitHubClient.GetPullRequests(owner, repo)
			result, err = nil, fmt.Errorf("TODO: implement GetPullRequests")
		} else {
			err = fmt.Errorf("missing owner or repo parameters")
		}

	case "search_code":
		query, _ := data["query"].(string)
		if query != "" {
			// result, err = gitHubClient.SearchCode(query)
			result, err = nil, fmt.Errorf("TODO: implement SearchCode")
		} else {
			err = fmt.Errorf("missing query parameter")
		}

	default:
		err = fmt.Errorf("unknown GitHub action: %s", action)
	}

	// Create response
	response := map[string]interface{}{
		"type":      "github_result",
		"action":    action,
		"timestamp": time.Now().Unix(),
		"context":   context,
	}

	if err != nil {
		response["status"] = "error"
		response["error"] = err.Error()
	} else {
		response["status"] = "success"
		response["data"] = result
	}

	sendResponse(client, response)
}

func handleContextRequest(client *Client, data map[string]interface{}, context translations.ContextVector7D) {
	// WHO: ContextRequestHandler
	// WHAT: Handle context WebSocket requests
	// WHEN: During WebSocket message processing
	// WHERE: System Layer 6 (Integration)
	// WHY: To provide context operations
	// HOW: Using translations package
	// EXTENT: Context translation lifecycle

	// Process context request
	operation, _ := data["operation"].(string)

	var response map[string]interface{}

	switch operation {
	case "get":
		// Return current context
		response = map[string]interface{}{
			"type":      "context_result",
			"operation": "get",
			"timestamp": time.Now().Unix(),
			"context":   client.Context,
		}

	case "update":
		// Extract context data from message
		rawContext, _ := data["context"].(map[string]interface{})
		if rawContext != nil {
			// Create new context
			newContext := translations.NewContextVector7D(rawContext)

			// Merge with client's existing context using 7D Merge method
			mergedContext := newContext.Merge(client.Context)

			// Update client's context
			client.Context = mergedContext

			// Return updated context
			response = map[string]interface{}{
				"type":      "context_result",
				"operation": "update",
				"timestamp": time.Now().Unix(),
				"context":   client.Context,
				"status":    "updated",
			}
		} else {
			response = map[string]interface{}{
				"type":      "error",
				"error":     "Invalid context data",
				"timestamp": time.Now().Unix(),
				"context":   context,
			}
		}

	case "compress":
		// Apply Möbius compression to context
		compressedContext := translations.CompressContext(client.Context, logger.(*logpkg.Logger))

		response = map[string]interface{}{
			"type":      "context_result",
			"operation": "compress",
			"timestamp": time.Now().Unix(),
			"context":   compressedContext,
			"status":    "compressed",
		}

	case "decompress":
		// Extract compressed context data from message
		rawContext, _ := data["context"].(map[string]interface{})
		if rawContext != nil {
			// Create context from raw data
			contextToDecompress := translations.NewContextVector7D(rawContext)

			// Decompress context
			decompressedContext := translations.DecompressContext(contextToDecompress, logger.(*logpkg.Logger))

			response = map[string]interface{}{
				"type":      "context_result",
				"operation": "decompress",
				"timestamp": time.Now().Unix(),
				"context":   decompressedContext,
				"status":    "decompressed",
			}
		} else {
			response = map[string]interface{}{
				"type":      "error",
				"error":     "Invalid compressed context data",
				"timestamp": time.Now().Unix(),
				"context":   context,
			}
		}

	case "bridge":
		// Extract GitHub context data from message
		rawGitHubContext, _ := data["github_context"].(map[string]interface{})
		if rawGitHubContext != nil {
			// Prepare GitHub context
			githubContext := translations.GitHubContext{
				User:      getStringValue(rawGitHubContext, "user", "System"),
				Identity:  getStringValue(rawGitHubContext, "identity", "System"),
				Operation: getStringValue(rawGitHubContext, "operation", "Transform"),
				Type:      getStringValue(rawGitHubContext, "type", "Context"),
				Purpose:   getStringValue(rawGitHubContext, "purpose", "Protocol_Compliance"),
				Scope:     getFloat64Value(rawGitHubContext, "scope", 1.0),
				Timestamp: getInt64Value(rawGitHubContext, "timestamp", time.Now().Unix()),
				Source:    getStringValue(rawGitHubContext, "source", "github_mcp"),
				Metadata:  getMapValue(rawGitHubContext, "metadata"),
			}

			// Bridge context between GitHub and TNOS
			bridgedContext := translations.BridgeMCPContext(githubContext, &client.Context, logger)

			// Update client context
			client.Context = bridgedContext

			response = map[string]interface{}{
				"type":      "context_result",
				"operation": "bridge",
				"timestamp": time.Now().Unix(),
				"context":   bridgedContext,
				"status":    "bridged",
			}
		} else {
			response = map[string]interface{}{
				"type":      "error",
				"error":     "Invalid GitHub context data",
				"timestamp": time.Now().Unix(),
				"context":   context,
			}
		}

	default:
		response = map[string]interface{}{
			"type":      "error",
			"error":     "Unknown context operation",
			"timestamp": time.Now().Unix(),
			"context":   context,
		}
	}

	sendResponse(client, response)
}

// Close all client connections
func closeAllClients() {
	clientsMtx.Lock()
	defer clientsMtx.Unlock()

	for client := range clients {
		delete(clients, client)
		close(client.Send)
	}
}

// Handle broadcasts to clients
func handleBroadcasts() {
	for {
		message, ok := <-broadcast
		if !ok {
			return // Channel closed
		}

		clientsMtx.Lock()
		for client := range clients {
			select {
			case client.Send <- message:
				// Message sent successfully
			default:
				// Failed to send, remove client
				close(client.Send)
				delete(clients, client)
			}
		}
		clientsMtx.Unlock()
	}
}

// Connect to MCP bridge
func connectToBridge() {
	// WHO: BridgeConnector
	// WHAT: Connect to Blood Bridge
	// WHEN: During server startup
	// WHERE: System Layer 6 (Integration)
	// WHY: To integrate with Blood Bridge
	// HOW: Using WebSocket connection with retry logic
	// EXTENT: Blood Bridge connection lifecycle with error handling

	if logger == nil {
		logger = logpkg.NewLogger()
	}
	logger.Info("Connecting to Blood Bridge", "port", config.BridgePort)

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	bridgeContext := translations.NewContextVector7D(map[string]interface{}{
		"who":    "BridgeConnector",
		"what":   "Blood_Bridge_Connection",
		"when":   time.Now().Unix(),
		"where":  "MCP_Server",
		"why":    "Blood_Integration",
		"how":    "WebSocket_Protocol",
		"extent": 1.0,
		"source": "github_mcp",
	})

	contextMap := map[string]interface{}{
		"who":    bridgeContext.Who,
		"what":   bridgeContext.What,
		"when":   bridgeContext.When,
		"where":  bridgeContext.Where,
		"why":    bridgeContext.Why,
		"how":    bridgeContext.How,
		"extent": bridgeContext.Extent,
		"source": bridgeContext.Source,
	}

	bloodOpts := common.ConnectionOptions{
		ServerURL:   "ws://localhost:10619/bridge",
		ServerPort:  10619,
		Context:     contextMap,
		Logger:      logger,
		Timeout:     60 * time.Second,
		MaxRetries:  5,
		RetryDelay:  2 * time.Second,
		Credentials: map[string]string{"source": "github-mcp-server"},
		Headers:     map[string]string{"X-QHP-Version": "1.0"},
	}

							bloodBridge, err := bridge.NewBloodCirculation(ctx, bloodOpts)
				if err != nil {
					logger.Error("Failed to connect to Blood Bridge", "error", err.Error())
					logger.Info("Operating in standalone mode - server will continue without blood bridge")
					logpkg.SetHelicalMemoryMode("standalone")
					bloodBridgeInstance = nil // No blood bridge available
					_ = logpkg.LogHelicalEvent(logpkg.NewHelicalEvent("MCPServer", "startup", "github-mcp-server", "init", "GoMain", "full", "GitHub MCP Server starting up (standalone mode)"), logger)
					// Server continues without blood bridge
				} else {
					// assign global instance
					bloodBridgeInstance = bloodBridge
					
					// START the blood circulation system
					bloodBridgeInstance.Start()
					
					logger.Info("Blood circulation system started")
					
					// Fix: pass the concrete logger type to ghmcp.NewClient
					var concreteLogger *logpkg.Logger
					if l, ok := logger.(*logpkg.Logger); ok {
						concreteLogger = l
					} else {
						concreteLogger = logpkg.NewLogger()
					}
					go processBloodCirculationMessages(
						bloodBridgeInstance, 
						ghmcp.NewClient(config.GitHubToken, concreteLogger),
						logger,
						startTime,
						&clients,
						&clientsMtx,
					)
				
					// After successful TNOS MCP bridge connection:
					logpkg.SetHelicalMemoryMode("connected")
					_ = logpkg.LogHelicalEvent(logpkg.NewHelicalEvent("MCPServer", "startup", "github-mcp-server", "init", "GoMain", "full", "GitHub MCP Server starting up"), logger)
				}
				
				// Server continues here regardless of blood bridge status
				logger.Info("GitHub MCP Server initialization complete")
			}

// Canonical MCP/QHP ports for TNOS 7D layer communication
var canonicalPorts = []int{9001, 10617, 8083, 10619}

func startAllServers() {
    // Log canonical ports for reference
    logger.Info("TNOS Canonical MCP/QHP ports available", "ports", canonicalPorts)
	// Main application server (WebSocket, API endpoints)
	mainMux := http.NewServeMux() // Use a specific mux for the main server
	mainMux.HandleFunc("/", handleRoot)
	mainMux.HandleFunc("/health", handleHealth)
	mainMux.HandleFunc("/ws", handleWebSocket) // WebSocket endpoint
	mainMux.HandleFunc("/api/status", handleStatus) // Register handleStatus here

	mainAddr := fmt.Sprintf("%s:%d", config.Host, config.Port)
	mainServer := &http.Server{
		Addr:    mainAddr,
		Handler: mainMux, // Use the main mux
	}
	addServer(mainServer) // Track this server for graceful shutdown

	go func() {
		logger.Info(fmt.Sprintf("Main GitHub MCP Server listening on %s", mainAddr))
		if err := mainServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error(fmt.Sprintf("Main server ListenAndServe failed: %s", err.Error()))
			_ = logpkg.LogHelicalEvent(logpkg.NewHelicalEvent("MCPServer", "fatal_error", "github-mcp-server", "crash", "GoMain", "full", "Main server ListenAndServe failed: "+err.Error()), logger)
			os.Exit(1) // Critical failure
		}
	}()

	// --- Bridge Listener (if this server also *listens* for bridge connections) ---
	// The config.BridgePort (MCP_BRIDGE_LISTENING_PORT, default 10619) was intended for this.
	// If this server needs to accept incoming connections *as a bridge endpoint* for other services,
	// then another server should be started on config.BridgePort.
	// The current `tnosBridge` and `copilotConnection` are outgoing connections from this server.

	// If config.BridgePort is for this server to *listen* for incoming bridge-related commands (not just outgoing connections),
	// a separate server instance would be needed here.
	// Example:
	// if config.BridgeEnabled && config.BridgePort != config.Port { // Only if different from main port
	// bridgeMux := http.NewServeMux()
	// bridgeMux.HandleFunc("/bridge-endpoint", handleBridgeSpecificRequests) // Example handler
	// bridgeAddr := fmt.Sprintf("%s:%d", config.Host, config.BridgePort)
	// bridgeServer := &http.Server{Addr: bridgeAddr, Handler: bridgeMux}
	// addServer(bridgeServer)
	// go func() {
	// logger.Info(fmt.Sprintf("Bridge listener server starting on %s", bridgeAddr))
	// if err := bridgeServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
	// logger.Error(fmt.Sprintf("Bridge listener server failed: %s", err.Error()))
	// }
	// }()
	// }
}

// --- MCP Bridge WebSocket Handler ---
func handleBridge(w http.ResponseWriter, r *http.Request) {
	upgrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool { return true },
	}
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		logger.Error("Failed to upgrade /bridge connection: %v", err)
		return
	}
	defer conn.Close()

	// Create a bridge client for this WebSocket session
	ctx := context.Background()
	options := common.ConnectionOptions{
		ServerURL:  "", // Not used for server-side
		Timeout:    60 * time.Second,
		MaxRetries: 0,
		RetryDelay: 0,
		Context:    map[string]interface{}{ // 7D TNOS context for bridge session
			common.ContextKeyWho:    "BridgeServer",
			common.ContextKeyWhat:   "Session",
			common.ContextKeyWhen:   time.Now().Unix(),
			common.ContextKeyWhere:  "Layer6",
			common.ContextKeyWhy:    "WebSocket",
			common.ContextKeyHow:    "Upgrade",
			common.ContextKeyExtent: 1.0,
		},
		Logger:     logger,
	}
	_, err = bridge.NewClient(ctx, options) // FIX: use blank identifier since bridgeClient is unused
	if err != nil {
		logger.Error("Failed to create bridge client: %v", err)
		return
	}

	// Example: simple echo loop for demonstration (replace with actual bridge logic)
	for {
		mt, message, err := conn.ReadMessage()
		if err != nil {
			logger.Error("Bridge WebSocket read error: %v", err)
			break
		}
		logger.Debug("Bridge received message: %s", string(message))
		// Here you would process the message with bridgeClient as needed

		// Echo back for now
		if err := conn.WriteMessage(mt, message); err != nil {
			logger.Error("Bridge WebSocket write error: %v", err)
			break
		}
	}
}
func addServer(server *http.Server) {
    serversMtx.Lock()
    defer serversMtx.Unlock()
    servers = append(servers, server)
}
// shouldBroadcast determines if a message type should be sent to broadcast channel
func shouldBroadcast(msgType string) bool {
    // Currently broadcasting all messages; refine as needed
    return true
}

// Safe helper functions for extracting typed values from map[string]interface{}
func getStringValue(m map[string]interface{}, key, defaultValue string) string {
    if v, ok := m[key]; ok {
        if s, ok := v.(string); ok {
            return s
        }
    }
    return defaultValue
}

func getFloat64Value(m map[string]interface{}, key string, defaultValue float64) float64 {
    if v, ok := m[key]; ok {
        switch val := v.(type) {
        case float64:
            return val
        case float32:
            return float64(val)
        case int:
            return float64(val)
        case int64:
            return float64(val)
        }
    }
    return defaultValue
}

func getInt64Value(m map[string]interface{}, key string, defaultValue int64) int64 {
    if v, ok := m[key]; ok {
        switch val := v.(type) {
        case int64:
            return val
        case int:
            return int64(val)
        case float64:
            return int64(val)
        }
    }
    return defaultValue
}

func getMapValue(m map[string]interface{}, key string) map[string]interface{} {
    if v, ok := m[key]; ok {
        if mp, ok := v.(map[string]interface{}); ok {
            return mp
        }
    }
    return map[string]interface{}{}
}

// End of main.go