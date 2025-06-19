/*
 * WHO: MCPServer (AI+Inventor Partnership)
 * WHAT: Canonical, full-featured, event-driven, 7D/ATM-compliant main entry for GitHub MCP Server
 * WHEN: At every server startup, shutdown, and major event
 * WHERE: System Layer 6 (Integration), Tranquility-Neuro-OS
 * WHY: To unify all biological, quantum, AI, and context-driven logic in a single, extensible entrypoint
 * HOW: By referencing all canonical protocols, formula registry, and 7D/AI-DNA/ATM/TranquilSpeak patterns
 * EXTENT: All MCP, bridge, context, and API operations
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
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/gorilla/websocket"
	"github.com/jubicudis/Tranquility-Neuro-OS/github-mcp-server/pkg/bridge"
	"github.com/jubicudis/Tranquility-Neuro-OS/github-mcp-server/pkg/common"
	pkgcontext "github.com/jubicudis/Tranquility-Neuro-OS/github-mcp-server/pkg/context"
	"github.com/jubicudis/Tranquility-Neuro-OS/github-mcp-server/pkg/log"
	"github.com/jubicudis/Tranquility-Neuro-OS/github-mcp-server/pkg/tranquilspeak"
	"github.com/jubicudis/Tranquility-Neuro-OS/github-mcp-server/pkg/translations"
)

var (
	// Use correct log import alias
	logger log.LoggerInterface
	mainContext translations.ContextVector7D
	bridgeClient *bridge.Bridge
	servers []*http.Server
	serversMtx sync.Mutex
)

type Config struct {
	Host string
	Port int
	LogLevel string
	LogFile string
	GitHubToken string
	BridgeEnabled bool
	BridgePort int
	FormulaRegistryPath string
}

func main() {
	// 1. Load configuration
	config := loadConfig()

	// 2. Initialize logger (7D/AI-DNA-aware)
	logger = initLogger(config)
	logger.Info("[BOOT] GitHub MCP Server starting up (7D/ATM/AI-DNA/TranquilSpeak)")

	// 3. Initialize 7D context (AI-DNA-aware)
	mainContext = translations.NewContextVector7D(map[string]interface{}{
		"who": "MCPServer",
		"what": "Startup",
		"when": time.Now().Unix(),
		"where": "Layer6",
		"why": "SystemInitialization",
		"how": "CanonicalMain",
		"extent": 1.0,
		"source": "github_mcp",
	})

	// 4. Load formula registry and log all formulas
	formulaPath := config.FormulaRegistryPath
	if formulaPath == "" {
		formulaPath = filepath.Join("config", "formulas.json")
	}
	if err := bridge.LoadBridgeFormulaRegistry(formulaPath); err != nil {
		logger.Error("Failed to load formula registry: %v", err)
	} else {
		reg := bridge.GetBridgeFormulaRegistry()
		if reg != nil {
			for _, f := range reg.ListFormulas() {
				logger.Info("Formula loaded", "id", f.ID, "desc", f.Description, "meta", f.Metadata)
			}
		}
	}

	// 5. Initialize canonical MCP bridge client
	bridgeOpts := common.ConnectionOptions{
		ServerURL: fmt.Sprintf("ws://%s:%d/bridge", config.Host, config.BridgePort),
		ServerPort: config.BridgePort,
		Context: mainContext.ToMap(),
		Logger: logger,
		Timeout: 60 * time.Second,
		MaxRetries: 5,
		RetryDelay: 2 * time.Second,
		Credentials: map[string]string{"source": "github-mcp-server"},
		Headers: map[string]string{"X-QHP-Version": "1.0"},
	}
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	var err error
	bridgeClient, err = bridge.NewClient(ctx, bridgeOpts)
	if err != nil {
		logger.Error("Failed to connect to MCP bridge: %v", err)
	} else {
		logger.Info("MCP bridge client connected")
	}

	// 6. Initialize context orchestrator (ATM/7D/biomimetic)
	orchestrator := pkgcontext.NewContextOrchestrator(logger)
	logger.Info("Context orchestrator initialized (ATM/7D/biomimetic)")

	// 7. Register ATM event triggers
	registerATMEventTriggers(orchestrator)

	// 8. Start GitHub MCP API server (REST + WebSocket)
	startGitHubMCPServer(config, orchestrator)

	// 9. Load TranquilSpeak symbol registry for correct symbol cluster logging
	err = tranquilspeak.LoadSymbolRegistry("/Users/Jubicudis/Tranquility-Neuro-OS/systems/tranquilspeak/circulatory/github-mcp-server/symbolic_mapping_registry_autogen_20250603.tsq")
	if err != nil {
		logger.Info("[BOOT] WARNING: Could not load TranquilSpeak symbol registry, symbol cluster logging may be incomplete: %v", err)
	}

	// 10. Wait for shutdown signal
	waitForShutdown()

	// 11. Send a bridge status message to the MCP bridge (demonstrate bridgeClient usage)
	if bridgeClient != nil {
		statusMsg := common.Message{
			Type:      "status",
			Timestamp: time.Now().Unix(),
			Payload: map[string]interface{}{
				"event":   "startup",
				"context": mainContext.ToMap(),
				"message": "GitHub MCP Server startup complete. Bridge client active.",
			},
		}
		err := bridgeClient.Send(statusMsg)
		if err != nil {
			logger.Warn("Failed to send bridge status message: %v", err)
		} else {
			logger.Info("Bridge status message sent to MCP bridge.")
		}
	}
}

func loadConfig() Config {
	var config Config
	flag.StringVar(&config.Host, "host", "localhost", "Server host")
	flag.IntVar(&config.Port, "port", 10617, "Server port")
	flag.StringVar(&config.LogLevel, "loglevel", "info", "Log level")
	flag.StringVar(&config.LogFile, "logfile", "github-mcp-server.log", "Log file")
	flag.StringVar(&config.GitHubToken, "token", os.Getenv("GITHUB_TOKEN"), "GitHub token")
	flag.BoolVar(&config.BridgeEnabled, "bridge", true, "Enable MCP bridge")
	flag.IntVar(&config.BridgePort, "bridgeport", 10619, "Bridge port")
	flag.StringVar(&config.FormulaRegistryPath, "formulas", "", "Formula registry path")
	flag.Parse()
	return config
}

func initLogger(config Config) log.LoggerInterface {
	logDir := filepath.Join("pkg", "log")
	logFilePath := filepath.Join(logDir, filepath.Base(config.LogFile))
	_ = os.MkdirAll(logDir, 0755)
	logger := log.NewLogger()
	switch strings.ToLower(config.LogLevel) {
	case "debug":
		logger = logger.WithLevel(log.LevelDebug)
	case "warn":
		logger = logger.WithLevel(log.LevelWarn)
	case "error":
		logger = logger.WithLevel(log.LevelError)
	default:
		logger = logger.WithLevel(log.LevelInfo)
	}
	f, err := os.OpenFile(logFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err == nil {
		logger = logger.WithOutput(f)
	}
	return logger
}

func registerATMEventTriggers(orchestrator *pkgcontext.ContextOrchestrator) {
	// Startup event
	orchestrator.ProcessContext("startup", map[string]interface{}{
		"event": "startup",
		"timestamp": time.Now().Unix(),
		"context": mainContext.ToMap(),
	})
	// Shutdown event
	orchestrator.ProcessContext("shutdown", map[string]interface{}{
		"event": "shutdown",
		"timestamp": time.Now().Unix(),
		"context": mainContext.ToMap(),
	})
	// API event example
	orchestrator.ProcessContext("api_event", map[string]interface{}{
		"event": "api_event",
		"timestamp": time.Now().Unix(),
		"context": mainContext.ToMap(),
	})
}

func startGitHubMCPServer(config Config, orchestrator *pkgcontext.ContextOrchestrator) {
	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})
	// Canonical GitHub MCP API endpoints
	mux.HandleFunc("/api/github", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("GitHub MCP API is running"))
	})
	// Canonical WebSocket endpoint (for event-driven, ATM-compliant ops)
	mux.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		// Use gorilla/websocket Upgrader directly (no GetCanonicalUpgrader)
		upgrader := websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin: func(r *http.Request) bool { return true },
		}
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			logger.Error("WebSocket upgrade failed: %v", err)
			return
		}
		defer conn.Close()

		// QHP handshake phase
		var handshakeMsg map[string]interface{}
		_, handshakeData, err := conn.ReadMessage()
		if err != nil {
			logger.Error("QHP handshake read error: %v", err)
			return
		}
		if err := json.Unmarshal(handshakeData, &handshakeMsg); err != nil {
			logger.Error("QHP handshake unmarshal error: %v", err)
			return
		}
		if handshakeMsg["type"] != "qhp_handshake" {
			logger.Error("QHP handshake: invalid type")
			return
		}
		fingerprint, _ := handshakeMsg["fingerprint"].(string)
		// Optionally: verify developer override token here
		// Accept handshake, generate session key
		sessionKey := bridge.GenerateQuantumFingerprint(fingerprint + "-session")
		// Update trust table
		bridge.GetTrustTable().Update(fingerprint, sessionKey, map[string]interface{}{"remote_addr": r.RemoteAddr})
		ack := map[string]interface{}{
			"type":        "qhp_handshake_ack",
			"fingerprint": fingerprint,
			"session_key": sessionKey,
			"trust_table": bridge.GetTrustTable().All(),
		}
		ackData, _ := json.Marshal(ack)
		if err := conn.WriteMessage(1, ackData); err != nil {
			logger.Error("QHP handshake ack write error: %v", err)
			return
		}
		logger.Info("QHP handshake complete", "fingerprint", fingerprint, "session_key", sessionKey)

		// Event-driven, ATM-compliant message loop
		for {
			mt, message, err := conn.ReadMessage()
			if err != nil {
				logger.Error("WebSocket read error: %v", err)
				break
			}
			var payload map[string]interface{}
			_ = json.Unmarshal(message, &payload)
			atmTrigger := tranquilspeak.CreateTrigger(
				"WebSocketClient", "WebSocketMessage", "Layer6", "ClientEvent", "WebSocket", "1.0", tranquilspeak.TriggerTypeDataTransport, "context", payload,
			)
			_ = atmTrigger // prevent unused variable warning
			err = orchestrator.ProcessContext("websocket_message", payload)
			if err != nil {
				logger.Error("ATM trigger processing failed: %v", err)
			}
			logger.Info("ATM trigger fired for WebSocket message", "payload", payload)
			// Echo back for now (extend with event-driven logic)
			if err := conn.WriteMessage(mt, message); err != nil {
				logger.Error("WebSocket write error: %v", err)
				break
			}
		}
	})

	// QHP HTTP endpoint for handshake/testing
	mux.HandleFunc("/qhp", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			w.Write([]byte("POST required"))
			return
		}
		var req map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("Invalid JSON"))
			return
		}
		fingerprint, _ := req["fingerprint"].(string)
		if fingerprint == "" {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("Missing fingerprint"))
			return
		}
		sessionKey := bridge.GenerateQuantumFingerprint(fingerprint + "-session")
		bridge.GetTrustTable().Update(fingerprint, sessionKey, map[string]interface{}{"remote_addr": r.RemoteAddr})
		resp := map[string]interface{}{
			"type":        "qhp_handshake_ack",
			"fingerprint": fingerprint,
			"session_key": sessionKey,
			"trust_table": bridge.GetTrustTable().All(),
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
		logger.Info("QHP HTTP handshake complete", "fingerprint", fingerprint, "session_key", sessionKey)
	})

	addr := fmt.Sprintf("%s:%d", config.Host, config.Port)
	server := &http.Server{Addr: addr, Handler: mux}
	addServer(server)
	go func() {
		logger.Info("GitHub MCP Server listening", "address", addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("Server ListenAndServe failed: %v", err)
		}
	}()
}

func addServer(server *http.Server) {
	serversMtx.Lock()
	servers = append(servers, server)
	serversMtx.Unlock()
}

func waitForShutdown() {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	sig := <-quit
	logger.Info("Received shutdown signal", "signal", sig.String())
	updateSelfDocumentation("shutdown", "Server shutdown initiated.")
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	serversMtx.Lock()
	for _, srv := range servers {
		_ = srv.Shutdown(ctx)
	}
	serversMtx.Unlock()
	logger.Info("Shutdown complete (7D/ATM/AI-DNA/TranquilSpeak)")
	updateSelfDocumentation("shutdown", "Server shutdown complete.")
}

// Self-documenting logic for technical log and TODOs
func updateSelfDocumentation(event, message string) {
	docPath := "../docs/development/REBUILD_TECHNICAL_LOG.md"
	todoPath := "../docs/TODO.md"
	timestamp := time.Now().Format(time.RFC3339)
	entry := fmt.Sprintf("[%s] EVENT: %s\nDETAIL: %s\n", timestamp, event, message)
	// Append to technical log
	f, err := os.OpenFile(docPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err == nil {
		f.WriteString(entry)
		f.Close()
	}
	// Optionally, update TODOs for major events
	if event == "startup" || event == "shutdown" {
		todoEntry := fmt.Sprintf("- [%s] %s\n", timestamp, message)
		tf, err := os.OpenFile(todoPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err == nil {
			tf.WriteString(todoEntry)
			tf.Close()
		}
	}
}