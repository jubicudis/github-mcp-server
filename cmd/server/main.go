/*
 * WHO: MCPServer
 * WHAT: Main GitHub MCP server implementation
 * WHEN: During GitHub API operations
 * WHERE: System Layer 6 (Integration)
 * WHY: To provide structured GitHub API access
 * HOW: Using Go HTTP server with WebSocket support
 * EXTENT: All GitHub API interactions
 */

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
	"sync"
	"syscall"
	"time"

	// Import internal packages with proper module paths

	"github.com/jubicudis/github-mcp-server/pkg/bridge"
	"github.com/jubicudis/github-mcp-server/pkg/log"
	"github.com/jubicudis/github-mcp-server/pkg/translations"

	// Import external packages
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
	config       Config
	logger       *log.Logger
	clients      = make(map[*Client]bool)
	clientsMtx   sync.Mutex
	broadcast    = make(chan []byte)
	gitHubClient GitHubService
	startTime    = time.Now() // Start time for uptime calculation
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
	logger.Info("Starting GitHub MCP Server", "version", "1.0")
	logger.Info("Configuration loaded", "host", config.Host, "port", config.Port)

	// Initialize GitHub client
	gitHubClientAdapter := NewClient(config.GitHubToken, logger)
	gitHubClient = gitHubClientAdapter
	logger.Info("GitHub client initialized")

	// Define HTTP routes
	http.HandleFunc("/", handleRoot)
	http.HandleFunc("/api/health", handleHealth)
	http.HandleFunc("/api/status", handleStatus)
	http.HandleFunc("/ws", handleWebSocket)

	// GitHub API routes
	http.HandleFunc("/api/github/repositories", handleGitHubRepositories)
	http.HandleFunc("/api/github/issues", handleGitHubIssues)
	http.HandleFunc("/api/github/pullrequests", handleGitHubPullRequests)
	http.HandleFunc("/api/github/search", handleGitHubSearch)
	http.HandleFunc("/api/github/code-scanning", handleGitHubCodeScanning)
	http.HandleFunc("/api/context", handleContext)

	// Start HTTP server
	addr := fmt.Sprintf("%s:%d", config.Host, config.Port)
	server := &http.Server{
		Addr: addr,
	}

	// Start server in a goroutine
	go func() {
		logger.Info("Server listening", "address", addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("Server failed to start", "error", err.Error())
			// Use os.Exit for fatal errors since we don't have a Fatal method
			os.Exit(1)
		}
	}()

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

	// Create context with timeout for graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second) // Increased timeout for proper shutdown
	defer cancel()

	// Close all client connections
	closeAllClients()

	// Shutdown server gracefully
	if err := server.Shutdown(ctx); err != nil {
		logger.Error("Server shutdown failed", "error", err.Error())
	}

	logger.Info("Server shutdown complete")
}

// Initialize logger
func initLogger(config Config) *log.Logger {
	// Get log directory from environment
	logDir := getEnv("MCP_LOG_DIR", filepath.Join(os.Getenv("TNOS_ROOT"), "logs"))
	if logDir == "" {
		logDir = filepath.Join(os.Getenv("HOME"), "logs")
	}

	// Create log directory if it doesn't exist
	if err := os.MkdirAll(logDir, 0755); err != nil {
		fmt.Printf("Failed to create log directory: %v\n", err)
	}

	// Initialize logger
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

	// Add file output if needed
	logFilePath := filepath.Join(logDir, config.LogFile)
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
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":    "healthy",
		"timestamp": time.Now().Unix(),
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
	repository, err := gitHubClient.GetRepository(owner, repo)

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
	issues, err := gitHubClient.GetIssues(owner, repo)

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
	prs, err := gitHubClient.GetPullRequests(owner, repo)

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
	results, err := gitHubClient.SearchCode(query)

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
	alerts, err := gitHubClient.GetCodeScanningAlerts(owner, repo)

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
			result, err = gitHubClient.GetRepository(owner, repo)
		} else {
			err = fmt.Errorf("missing owner or repo parameters")
		}

	case "get_issues":
		if owner != "" && repo != "" {
			result, err = gitHubClient.GetIssues(owner, repo)
		} else {
			err = fmt.Errorf("missing owner or repo parameters")
		}

	case "get_pulls":
		if owner != "" && repo != "" {
			result, err = gitHubClient.GetPullRequests(owner, repo)
		} else {
			err = fmt.Errorf("missing owner or repo parameters")
		}

	case "search_code":
		query, _ := data["query"].(string)
		if query != "" {
			result, err = gitHubClient.SearchCode(query)
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

// Helper functions for type conversion
func getStringValue(data map[string]interface{}, key, defaultValue string) string {
	if val, ok := data[key]; ok {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return defaultValue
}

func getFloat64Value(data map[string]interface{}, key string, defaultValue float64) float64 {
	if val, ok := data[key]; ok {
		switch v := val.(type) {
		case int:
			return float64(v)
		case int32:
			return float64(v)
		case int64:
			return float64(v)
		case float32:
			return float64(v)
		case float64:
			return v
		}
	}
	return defaultValue
}

func getInt64Value(data map[string]interface{}, key string, defaultValue int64) int64 {
	if val, ok := data[key]; ok {
		switch v := val.(type) {
		case int:
			return int64(v)
		case int32:
			return int64(v)
		case int64:
			return v
		case float32:
			return int64(v)
		case float64:
			return int64(v)
		}
	}
	return defaultValue
}

func getMapValue(data map[string]interface{}, key string) map[string]interface{} {
	if val, ok := data[key]; ok {
		if m, ok := val.(map[string]interface{}); ok {
			return m
		}
	}
	return make(map[string]interface{})
}

func shouldBroadcast(msgType string) bool {
	// Only broadcast certain message types
	// This is a simple implementation - customize as needed
	broadcastTypes := map[string]bool{
		"github":  true,
		"context": true,
	}

	return broadcastTypes[msgType]
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
	// WHAT: Connect to TNOS MCP bridge
	// WHEN: During server startup
	// WHERE: System Layer 6 (Integration)
	// WHY: To integrate with TNOS system
	// HOW: Using WebSocket connection with retry
	// EXTENT: Bridge connection lifecycle with error handling

	// Log connection attempt
	logger.Info("Connecting to MCP bridge", "port", config.BridgePort)

	// Create a context for the bridge connection with extended timeout (30s instead of 5s)
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second) // Increased timeout for bridge connection
	defer cancel()

	// Create a context vector for the bridge connection
	bridgeContext := translations.NewContextVector7D(map[string]interface{}{
		"who":    "BridgeConnector",
		"what":   "Bridge_Connection",
		"when":   time.Now().Unix(),
		"where":  "MCP_Server",
		"why":    "TNOS_Integration",
		"how":    "WebSocket_Protocol",
		"extent": 1.0,
		"source": "github_mcp",
	})

	// Set up bridge connection options
	opts := bridge.ConnectionOptions{
		ServerURL:  fmt.Sprintf("ws://localhost:%d/bridge", config.BridgePort),
		ServerPort: config.BridgePort,
		Context:    bridgeContext,
		Logger:     logger,
		Timeout:    60 * time.Second, // Increased timeout to match handshake timeout
		MaxRetries: 5,
		RetryDelay: 2 * time.Second,
	}

	// Connect to the bridge with retry
	var bridgeClient *bridge.Client
	var err error

	for i := 0; i < 5; i++ { // Try 5 times
		bridgeClient, err = bridge.NewClient(ctx, opts)
		if err == nil {
			break
		}

		logger.Warn("Failed to connect to MCP bridge, retrying...", "attempt", i+1, "error", err.Error())
		time.Sleep(2 * time.Second) // Wait before retry
	}

	if err != nil {
		logger.Error("Failed to connect to MCP bridge after retries", "error", err.Error())
		return
	}

	logger.Info("Successfully connected to MCP bridge")

	// Handle bridge messages in a goroutine
	go func() {
		for {
			message, err := bridgeClient.Receive()
			if err != nil {
				logger.Error("Error receiving bridge message", "error", err.Error())

				// Attempt to reconnect if disconnected
				time.Sleep(5 * time.Second)
				newCtx, newCancel := context.WithTimeout(context.Background(), 60*time.Second) // Increased timeout for retry attempts
				bridgeClient, err = bridge.NewClient(newCtx, opts)
				newCancel()

				if err != nil {
					logger.Error("Failed to reconnect to bridge", "error", err.Error())
					return
				}

				continue
			}

			// Process bridge message
			logger.Debug("Received bridge message", "type", message.Type)
			processTNOSBridgeMessage(bridgeClient, message)
		}
	}()
}

// processMCPBridgeMessage handles messages received from the main MCP bridge
func processMCPBridgeMessage(client *bridge.Client, message bridge.Message) {
	// WHO: MessageProcessor
	// WHAT: Process MCP bridge messages
	// WHEN: When messages are received
	// WHERE: System Layer 6 (Integration)
	// WHY: To handle bridge communications
	// HOW: Using message type dispatch
	// EXTENT: All bridge messages

	// Log message receipt
	logger.Debug("Processing bridge message", "type", message.Type)

	// Handle message based on type
	switch message.Type {
	case "context_sync":
		handleContextSync(message)
	case "formula_execution":
		handleFormulaExecution(message)
	case "visualization_request":
		handleVisualizationRequest(message, client)
	case "compression_request":
		handleCompressionRequest(message, client)
	case "error":
		handleBridgeError(message)
	default:
		logger.Warn("Unhandled bridge message type", "type", message.Type)
	}
}

// handleContextSync processes context synchronization messages
func handleContextSync(message bridge.Message) {
	// WHO: ContextSynchronizer
	// WHAT: Synchronize context between MCP systems
	// WHEN: During context sync operations
	// WHERE: System Layer 6 (Integration)
	// WHY: To maintain context consistency
	// HOW: Using context translation
	// EXTENT: All context dimensions

	logger.Debug("Processing context sync message")

	// Check if we have payload data
	if message.Payload == nil {
		logger.Error("Invalid context sync message format: missing payload")
		return
	}

	// Process context sync (implementation depends on specific needs)
	// ...

	logger.Info("Context synchronized with TNOS MCP")
}

// handleFormulaExecution processes formula execution messages
func handleFormulaExecution(message bridge.Message) {
	// WHO: FormulaExecutor
	// WHAT: Execute formulas from TNOS
	// WHEN: During formula execution requests
	// WHERE: System Layer 6 (Integration)
	// WHY: To perform calculations
	// HOW: Using formula registry
	// EXTENT: Single formula execution

	logger.Debug("Processing formula execution message")

	// Check if we have payload data
	if message.Payload == nil {
		logger.Error("Invalid formula execution message format: missing payload")
		return
	}

	// Process formula execution (implementation depends on specific needs)
	// ...

	logger.Info("Formula executed successfully")
}

// processTNOSBridgeMessage handles messages from the TNOS MCP bridge
func processTNOSBridgeMessage(bridgeClient *bridge.Client, message bridge.Message) {
	// WHO: BridgeMessageProcessor
	// WHAT: Process incoming bridge messages
	// WHEN: During bridge communication
	// WHERE: System Layer 6 (Integration)
	// WHY: To handle TNOS system messages
	// HOW: Using message type dispatch
	// EXTENT: Single message lifecycle

	// Process message based on type
	switch message.Type {
	case "context":
		// Handle context update
		// Make sure Payload exists
		if message.Payload == nil {
			logger.Error("Failed to parse context message: payload is missing")
			return
		}

		// Create 7D context from payload map
		updatedContext := translations.NewContextVector7D(message.Payload)

		// Broadcast context update to WebSocket clients
		broadcastContextUpdate(updatedContext)

	case "event":
		// Handle TNOS event
		// Make sure Payload exists
		if message.Payload == nil {
			logger.Error("Failed to parse event message: payload is missing")
			return
		}

		// Log event
		eventType, _ := message.Payload["type"].(string)
		logger.Info("Received TNOS event", "type", eventType)

		// Broadcast event to WebSocket clients
		broadcastEvent(message.Payload)

	case "formula":
		// Handle formula registration or update
		// Make sure Payload exists
		if message.Payload == nil {
			logger.Error("Failed to parse formula message: payload is missing")
			return
		}

		// Log formula update
		formulaName, _ := message.Payload["name"].(string)
		logger.Info("Received formula update", "name", formulaName)

		// Acknowledge formula update
		ackMessage := bridge.Message{
			Type:    "formula_ack",
			Payload: map[string]interface{}{
				"status": "received",
				"name":   formulaName,
			},
		}

		if err := bridgeClient.Send(ackMessage); err != nil {
			logger.Error("Failed to send formula acknowledgement", "error", err.Error())
		}

	case "ping":
		// Respond to ping
		pongMessage := bridge.Message{
			Type: "pong",
			Payload: map[string]interface{}{
				"timestamp": time.Now().Unix(),
			},
		}

		if err := bridgeClient.Send(pongMessage); err != nil {
			logger.Error("Failed to send pong", "error", err.Error())
		}

	default:
		logger.Warn("Unknown bridge message type", "type", message.Type)
	}
}

// broadcastContextUpdate broadcasts a context update to all WebSocket clients
func broadcastContextUpdate(context translations.ContextVector7D) {
	updateMsg := map[string]interface{}{
		"type":      "context_update",
		"source":    "tnos_bridge",
		"timestamp": time.Now().Unix(),
		"context":   context,
	}

	broadcastData, _ := json.Marshal(updateMsg)
	broadcast <- broadcastData
}

// broadcastEvent broadcasts a TNOS event to all WebSocket clients
func broadcastEvent(eventData map[string]interface{}) {
	eventData["type"] = "tnos_event"
	eventData["source"] = "tnos_bridge"
	eventData["timestamp"] = time.Now().Unix()

	broadcastData, _ := json.Marshal(eventData)
	broadcast <- broadcastData
}

// handleVisualizationRequest processes visualization requests
func handleVisualizationRequest(message bridge.Message, client *bridge.Client) {
	// WHO: VisualizationManager
	// WHAT: Handle visualization requests
	// WHEN: During visualization operations
	// WHERE: System Layer 6 (Integration)
	// WHY: To generate visual representations
	// HOW: Using visualization service
	// EXTENT: Single visualization request

	logger.Debug("Processing visualization request")

	// Check if we have payload data
	if message.Payload == nil {
		logger.Error("Invalid visualization request format: missing payload")
		return
	}

	// Process visualization request (implementation depends on specific needs)
	// ...

	// Extract request ID if available
	requestID, _ := message.Payload["requestId"].(string)

	// Send response back to bridge
	response := bridge.Message{
		Type:      "visualization_response",
		Timestamp: time.Now().Unix(),
		Payload: map[string]interface{}{
			"status":    "success",
			"requestId": requestID,
			// Add visualization data here
		},
	}

	if err := client.Send(response); err != nil {
		logger.Error("Failed to send visualization response", "error", err.Error())
	}
}

// handleCompressionRequest processes compression requests
func handleCompressionRequest(message bridge.Message, client *bridge.Client) {
	// WHO: CompressionManager
	// WHAT: Handle compression requests
	// WHEN: During compression operations
	// WHERE: System Layer 6 (Integration)
	// WHY: To compress/decompress data
	// HOW: Using Möbius compression
	// EXTENT: Single compression request

	logger.Debug("Processing compression request")

	// Check if we have payload data
	if message.Payload == nil {
		logger.Error("Invalid compression request format: missing payload")
		return
	}

	// Process compression request (implementation depends on specific needs)
	// ...

	// Extract request ID if available
	requestID, _ := message.Payload["requestId"].(string)

	// Send response back to bridge
	response := bridge.Message{
		Type:      "compression_response",
		Timestamp: time.Now().Unix(),
		Payload: map[string]interface{}{
			"status":    "success",
			"requestId": requestID,
			// Add compression result here
		},
	}

	if err := client.Send(response); err != nil {
		logger.Error("Failed to send compression response", "error", err.Error())
	}
}

// handleBridgeError processes error messages from the bridge
func handleBridgeError(message bridge.Message) {
	// WHO: ErrorHandler
	// WHAT: Process bridge errors
	// WHEN: During error conditions
	// WHERE: System Layer 6 (Integration)
	// WHY: For error handling
	// HOW: Using error processing
	// EXTENT: Single error message

	// Check if we have payload data
	if message.Payload == nil {
		logger.Error("Invalid error message format: missing payload")
		return
	}

	errorMsg, _ := message.Payload["message"].(string)
	errorCode, _ := message.Payload["code"].(float64)

	logger.Error("Bridge error received",
		"code", errorCode,
		"message", errorMsg)
}
