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
)

var (
	// Use correct log import alias
	logger log.LoggerInterface
	mainContext pkgcontext.ContextVector7D
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
	var err error

	// 1. Load Symbol Registry (TranquilSpeak, event/trigger definitions)
	err = tranquilspeak.LoadSymbolRegistry("/Users/Jubicudis/Tranquility-Neuro-OS/systems/tranquilspeak/circulatory/github-mcp-server/symbolic_mapping_registry_autogen_20250603.tsq")
	if err != nil {
		panic("[FATAL] Could not load TranquilSpeak symbol registry: " + err.Error())
	}

	// 2. Load Formula Registry (Mobius, HemoFlux, etc.)
	config := loadConfig()
	formulaPath := config.FormulaRegistryPath
	if formulaPath == "" {
		formulaPath = filepath.Join("config", "formulas.json")
	}
	if err := bridge.LoadBridgeFormulaRegistry(formulaPath); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load formula registry: %v\n", err)
	}

	// 3. Initialize 7D Context and Memory (ContextVector7D, Helical Memory, etc.)
		mainContext = pkgcontext.ContextVector7D{
		Who:    "MCPServer",
		What:   "Startup",
		When:   time.Now().Unix(),
		Where:  "Layer6",
		Why:    "SystemInitialization",
		How:    "CanonicalMain",
		Extent: 1.0,
		Source: "github_mcp",
	}

	// 4. Configure Canonical Logging and Event Routing (TriggerMatrix, logger)
	triggerMatrix := tranquilspeak.NewTriggerMatrix()
	logger = initLogger(config, triggerMatrix)
	logger.Info("[BOOT] GitHub MCP Server starting up (7D/ATM/AI-DNA/TranquilSpeak)")

	// 5. Register All Canonical Tools and Bridges (Context orchestrator, event triggers)
	orchestrator := pkgcontext.NewContextOrchestrator(logger)
	logger.Info("Context orchestrator initialized (ATM/7D/biomimetic)")
	registerATMEventTriggers(orchestrator)

	// 6. Initialize Port Assignment AI (Mobius Collapse, formula registry-driven)
	// (If there is a Mobius/port assignment module, initialize it here. Otherwise, ensure formula registry is loaded.)
	// Example: bridge.GetBridgeFormulaRegistry().EnsureHemofluxFormulas()
	if reg := bridge.GetBridgeFormulaRegistry(); reg != nil {
		reg.EnsureHemofluxFormulas()
		logger.Info("Mobius Collapse/Port Assignment AI initialized (formula registry-driven)")
	}

	// 7. Start MCP Server (REST + WebSocket)
	startGitHubMCPServer(config, orchestrator)
	logger.Info("GitHub MCP Server started (REST + WebSocket)")

	// 8. Start Bridge to TNOS MCP Server (after MCP server is running and ready)
	if config.BridgeEnabled {
		bridgeOpts := common.ConnectionOptions{
			ServerURL:   fmt.Sprintf("ws://%s:%d/bridge", config.Host, config.BridgePort),
			ServerPort:  config.BridgePort,
			Context:     log.ToMap(mainContext),
			Logger:      logger,
			Timeout:     60 * time.Second,
			MaxRetries:  5,
			RetryDelay:  2 * time.Second,
			Credentials: map[string]string{"source": "github-mcp-server"},
			Headers:     map[string]string{"X-QHP-Version": "1.0"},
		}
		ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
		defer cancel()
		bridgeClient, err = bridge.NewClient(ctx, bridgeOpts, triggerMatrix)
		if err != nil {
			logger.Error("Failed to connect to MCP bridge: %v", err)
		} else {
			logger.Info("MCP bridge client connected to TNOS MCP server")
			// Optionally send a status message
			statusMsg := common.Message{
				Type:      "status",
				Timestamp: time.Now().Unix(),
				Payload: map[string]interface{}{
					"event":   "startup",
					"context": log.ToMap(mainContext),
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

	// 9. Wait for shutdown signal
	waitForShutdown()
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

func initLogger(config Config, triggerMatrix *tranquilspeak.TriggerMatrix) log.LoggerInterface {
	logger := log.NewLogger(triggerMatrix)
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
	return logger
}

func registerATMEventTriggers(orchestrator *pkgcontext.ContextOrchestrator) {
	// Startup event
	orchestrator.ProcessContext("startup", map[string]interface{}{
		"event": "startup",
		"timestamp": time.Now().Unix(),
		"context": log.ToMap(mainContext),
	})
	// Shutdown event
	orchestrator.ProcessContext("shutdown", map[string]interface{}{
		"event": "shutdown",
		"timestamp": time.Now().Unix(),
		"context": log.ToMap(mainContext),
	})
	// API event example
	orchestrator.ProcessContext("api_event", map[string]interface{}{
		"event": "api_event",
		"timestamp": time.Now().Unix(),
		"context": log.ToMap(mainContext),
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