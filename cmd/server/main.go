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
	"strconv"
	"sync"
	"syscall"
	"time"

	// Import internal packages with proper module paths

	"github-mcp-server/pkg/bridge"
	"github-mcp-server/pkg/common"
	"github-mcp-server/pkg/github"
	"github-mcp-server/pkg/log"
	"github-mcp-server/pkg/translations"

	// Import external packages
	logpkg "github-mcp-server/pkg/log"

	"github.com/gorilla/websocket"
)

// Server configuration
type Config struct {
	// WHO: ConfigManager
	// WHAT: Server configuration settings
	// WHEN: During server initialization
	// WHERE: System Layer 6 (Integration)
	// WHY: To centralize configuration parameters
	// HOW: Using structured configuration with env fallbacks
	// EXTENT: All server parameters
	Port          int    `json:"port"`
	Host          string `json:"host"`
	LogLevel      string `json:"logLevel"`
	LogFile       string `json:"logFile"`
	BridgeEnabled bool   `json:"bridgeEnabled"`
	BridgePort    int    `json:"bridgePort"`
	GitHubToken   string `json:"githubToken"`
}

// Global variables
var (
	config            Config
	logger            *log.Logger
	clients           = make(map[*Client]bool)
	clientsMtx        sync.Mutex
	broadcast         = make(chan []byte)
	gitHubClient      github.Client // Use canonical github.Client
	startTime         = time.Now()  // Start time for uptime calculation
	compressionBridge interface{}   // Remove type reference, not used in main.go

	// Track all HTTP servers for graceful shutdown
	servers    []*http.Server
	serversMtx sync.Mutex
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
	portStr := getEnv("MCP_SERVER_PORT", "10617")
	port, err := strconv.Atoi(portStr)
	if err != nil {
		port = 10617 // Default port
	}

	bridgePortStr := getEnv("MCP_BRIDGE_PORT", "10619")
	bridgePort, err := strconv.Atoi(bridgePortStr)
	if err != nil {
		bridgePort = 10619 // Default bridge port
	}

	return Config{
		Port:          port,
		Host:          getEnv("MCP_SERVER_HOST", "localhost"),
		LogLevel:      getEnv("MCP_LOG_LEVEL", "info"),
		LogFile:       getEnv("MCP_LOG_FILE", "github-mcp-server.log"),
		BridgeEnabled: getEnvBool("MCP_BRIDGE_ENABLED", true),
		BridgePort:    bridgePort,
		GitHubToken:   getEnv("GITHUB_TOKEN", ""),
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
	// Parse command line flags
	var configFile string
	flag.StringVar(&configFile, "config", "", "Path to configuration file")
	flag.Parse()

	// Initialize configuration
	config = initConfig()

	// If config file is provided, load from it
	if configFile != "" {
		if data, err := os.ReadFile(configFile); err == nil {
			if err := json.Unmarshal(data, &config); err != nil {
				fmt.Printf("Warning: Failed to parse config file: %v\n", err)
			}
		} else {
			fmt.Printf("Warning: Failed to read config file: %v\n", err)
		}
	}

	// Initialize logger
	logger = initLogger(config)
	// Log server startup to helical memory
	_ = logpkg.LogHelicalEvent(logpkg.NewHelicalEvent(
		"MCPServer", "startup", "github-mcp-server", "init", "GoMain", "full", "GitHub MCP Server starting up"))
	logger.Info("Starting GitHub MCP Server", "version", "1.0")
	logger.Info("Configuration loaded", "host", config.Host, "port", config.Port)

	// Initialize GitHub client using common ConnectionOptions
	connOpts := common.ConnectionOptions{
		Credentials: map[string]string{"token": config.GitHubToken},
		Logger:      logger,
	}
	gitHubClient := struct{}{} // TODO: replace with actual client initialization
	logger.Info("GitHub client initialized")

	// --- Formula Registry Initialization ---
	formulaPath := "config/formulas.json"
	if err := bridge.LoadBridgeFormulaRegistry(formulaPath); err != nil {
		logger.Error("Failed to load formula registry", "error", err.Error(), "path", formulaPath)
	} else {
		reg := bridge.GetBridgeFormulaRegistry()
		count := 0
		if reg != nil {
			count = len(reg.ListFormulas())
		}
		logger.Info("Formula registry loaded", "count", count, "path", formulaPath)
	}

	// Initialize Blood System
	ctx := context.Background()
	github.InitializeBloodSystem(ctx)

	// Start HTTP servers on all canonical ports
	startAllServers()

	// Start WebSocket broadcaster
	go handleBroadcasts()

	// Connect to bridge if enabled
	if config.BridgeEnabled {
		go connectToBridge()
	}

	// Wait for termination signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	sig := <-quit
	logger.Info("Received termination signal", "signal", sig.String())
	_ = logpkg.LogHelicalEvent(logpkg.NewHelicalEvent(
		"MCPServer", "shutdown", "github-mcp-server", "signal", "GoMain", "full", "Received termination signal: "+sig.String()))

	// Create context with timeout for graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second) // Increased timeout for proper shutdown
	defer cancel()

	// Close all client connections
	closeAllClients()

	// Gracefully shutdown all HTTP servers
	serversMtx.Lock()
	for _, srv := range servers {
		if err := srv.Shutdown(ctx); err != nil {
			logger.Error("Server shutdown failed", "address", srv.Addr, "error", err.Error())
			_ = logpkg.LogHelicalEvent(logpkg.NewHelicalEvent(
				"MCPServer", "shutdown_error", "github-mcp-server", "shutdown", "GoMain", "full", "Server shutdown failed: "+err.Error()))
		}
	}
	serversMtx.Unlock()

	logger.Info("Server shutdown complete")
	_ = logpkg.LogHelicalEvent(logpkg.NewHelicalEvent(
		"MCPServer", "shutdown_complete", "github-mcp-server", "shutdown", "GoMain", "full", "Server shutdown complete"))
}

// Initialize logger
func initLogger(config Config) *log.Logger {
	// Use TNOS-compliant log path
	logFilePath := "/systems/memory/github/short_term.log"
	if err := os.MkdirAll("/systems/memory/github", 0755); err != nil {
		fmt.Printf("Failed to create log directory: %v\n", err)
	}

	logger := log.NewLogger()

	// Configure the logger based on the config
	if config.LogLevel == "debug" {
		logger = logger.WithLevel(log.LevelDebug)
	} else if config.LogLevel == "info" {
		logger = logger.WithLevel(log.LevelInfo)
	} else if config.LogLevel == "warn" {
		logger = logger.WithLevel(log.LevelWarn)
	} else if config.LogLevel == "error" {
		logger = logger.WithLevel(log.LevelError)
	} else {
		logger = logger.WithLevel(log.LevelInfo) // Default level
	}

	logFile, err := os.OpenFile(logFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err == nil {
		logger = logger.WithOutput(logFile)
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
		if gitHubClient != nil {
			copilotStatus = "healthy"
		}
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

// Server status handler
func handleStatus(w http.ResponseWriter, r *http.Request) {
	clientsMtx.Lock()
	clientCount := len(clients)
	clientsMtx.Unlock()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":        "running",
		"uptime":        int64(time.Since(startTime).Seconds()),
		"clientCount":   clientCount,
		"host":          config.Host,
		"port":          config.Port,
		"bridgeEnabled": config.BridgeEnabled,
		"timestamp":     time.Now().Unix(),
	})
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
	// repository, err := gitHubClient.GetRepository(owner, repo)
	repository, err := nil, fmt.Errorf("TODO: implement GetRepository")

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
	// issues, err := gitHubClient.GetIssues(owner, repo)
	issues, err := nil, fmt.Errorf("TODO: implement GetIssues")

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

	// Use the GitHub client to get repository information
	// prs, err := gitHubClient.GetPullRequests(owner, repo)
	prs, err := nil, fmt.Errorf("TODO: implement GetPullRequests")

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
func handleGitHubSearch(w http.ResponseWriter, r *http.Request) {
	// WHO: SearchHandler
	// WHAT: Handle GitHub code search
	// WHEN: During API request
	// WHERE: System Layer 6 (Integration)
	// WHY: To provide search results
	// HOW: Using GitHub API client
	// EXTENT: Single API request

	logger.Info("Search API called", "method", r.Method, "path", r.URL.Path)

	// Extract search query from query parameters
	query := r.URL.Query().Get("q")

	if query == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"status": "error",
			"error":  "Missing q parameter",
		})
		return
	}

	// Use the GitHub client to search code
	// results, err := gitHubClient.SearchCode(query)
	results, err := nil, fmt.Errorf("TODO: implement SearchCode")

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
func handleGitHubCodeScanning(w http.ResponseWriter, r *http.Request) {
	// WHO: SecurityHandler
	// WHAT: Handle GitHub security alerts
	// WHEN: During API request
	// WHERE: System Layer 6 (Integration)
	// WHY: To provide security data
	// HOW: Using GitHub API client
	// EXTENT: Single API request

	logger.Info("Code Scanning API called", "method", r.Method, "path", r.URL.Path)

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
	// alerts, err := gitHubClient.GetCodeScanningAlerts(owner, repo)
	alerts, err := nil, fmt.Errorf("TODO: implement GetCodeScanningAlerts")

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
		context := translations.NewContextVector7D(inputContext)

		// Apply compression
		compressed := translations.CompressContext(context, logger)

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

			// Merge with client's existing context
			mergedContext := translations.MergeContexts(newContext, client.Context, logger)

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
		compressedContext := translations.CompressContext(client.Context, logger)

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
			decompressedContext := translations.DecompressContext(contextToDecompress, logger)

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
		TLSEnabled:  false,
		Credentials: map[string]string{"source": "github-mcp-server"},
		Headers:     map[string]string{"X-QHP-Version": "1.0"},
	}

	bloodBridge, err := bridge.NewBloodCirculation(ctx, bloodOpts)
	if err != nil {
		logger.Error("Failed to connect to Blood Bridge", "error", err.Error())
		logger.Info("Operating in standalone mode")
		return
	}

	logger.Info("Successfully connected to Blood Bridge")
	go processBloodCirculationMessages(bloodBridge)
}

// Canonical MCP/QHP ports
var canonicalPorts = []int{9001, 10617, 10619, 8083}

func startAllServers() {
	for _, port := range canonicalPorts {
		go func(p int) {
			mux := http.NewServeMux()
			mux.HandleFunc("/", handleRoot)
			mux.HandleFunc("/api/health", handleHealth)
			mux.HandleFunc("/api/status", handleStatus)
			mux.HandleFunc("/ws", handleWebSocket)

			// GitHub API routes
			mux.HandleFunc("/api/github/repositories", handleGitHubRepositories)
			mux.HandleFunc("/api/github/issues", handleGitHubIssues)
			mux.HandleFunc("/api/github/pullrequests", handleGitHubPullRequests)
			mux.HandleFunc("/api/github/search", handleGitHubSearch)
			mux.HandleFunc("/api/github/code-scanning", handleGitHubCodeScanning)
			mux.HandleFunc("/api/context", handleContext)

			// Add the /bridge handler
			mux.HandleFunc("/bridge", handleBridge)

			addr := fmt.Sprintf("%s:%d", config.Host, p)
			server := &http.Server{Addr: addr, Handler: mux}

			// Track server for shutdown
			serversMtx.Lock()
			servers = append(servers, server)
			serversMtx.Unlock()

			// Start server in a goroutine
			go func() {
				logger.Info("Server listening", "address", addr)
				if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
					logger.Error("Server failed to start", "error", err.Error())
					_ = logpkg.LogHelicalEvent(logpkg.NewHelicalEvent(
						"MCPServer", "fatal_error", "github-mcp-server", "crash", "GoMain", "full", "Server failed to start: "+err.Error()))
					os.Exit(1)
				}
			}()
		}(port)
	}
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
		Headers:    nil,
		Context:    map[string]interface{}{ // Use map for context
			"who":    "BridgeServer",
			"what":   "Session",
			"when":   time.Now().Unix(),
			"where":  "Layer6",
			"why":    "WebSocket",
			"how":    "Upgrade",
			"extent": 1.0,
		},
		Logger:     logger,
	}
	bridgeClient, err := bridge.NewClient(ctx, options)
	if err != nil {
		logger.Error("Failed to create bridge client: %v", err)
		return
	}
	// Set the WebSocket connection on the bridge client if supported (optional, else use conn directly)
	// Forward messages from WebSocket to bridge
	go func() {
		for {
			_, msg, err := conn.ReadMessage()
			if err != nil {
				logger.Error("WebSocket read error: %v", err)
				conn.Close()
				return
			}
			var bridgeMsg common.Message
			if err := json.Unmarshal(msg, &bridgeMsg); err == nil {
				_ = bridgeClient.Send(bridgeMsg)
			}
		}
	}()
	// Forward messages from bridge to WebSocket
	go func() {
		for {
			msg, err := bridgeClient.Receive()
			if err != nil {
				return
			}
			msgBytes, err := json.Marshal(msg)
			if err == nil {
				_ = conn.WriteMessage(websocket.TextMessage, msgBytes)
			}
		}
	}()
	// Wait for connection close
	for {
		if _, _, err := conn.NextReader(); err != nil {
			conn.Close()
			return
		}
	}
}

// Helper stubs for missing functions and context extraction
func shouldBroadcast(msgType string) bool {
    // TODO: implement broadcast filtering logic
    return true
}

func getStringValue(m map[string]interface{}, key, defaultValue string) string {
    return common.GetString(m, key, defaultValue)
}

func getInt64Value(m map[string]interface{}, key string, defaultValue int64) int64 {
    return common.GetInt64(m, key, defaultValue)
}

func getFloat64Value(m map[string]interface{}, key string, defaultValue float64) float64 {
    return common.GetFloat64(m, key, defaultValue)
}

func getMapValue(m map[string]interface{}, key string) map[string]interface{} {
    if v, ok := m[key].(map[string]interface{}); ok {
        return v
    }
    return nil
}
