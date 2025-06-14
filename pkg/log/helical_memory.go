// filepath: github-mcp-server/pkg/log/helical_memory.go
// WHO: HelicalMemoryLogger
// WHAT: Dual-helix (short/long-term) 7D context event logger for TNOS MCP
// WHEN: During all major MCP/bridge/server events
// WHERE: System Layer 6 (Integration)
// WHY: To provide HDSA-compliant, 7D-aware, self-healing memory logging
// HOW: Writes JSON lines to short_term.log and long_term.log in memory/
// EXTENT: All MCP/bridge/server event logging

package log

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

var (
	helicalMode = "standalone" // default mode

	// TNOS-compliant root memory directory for blood-connected mode
	helicalMemoryDir = "/Users/Jubicudis/Tranquility-Neuro-OS/systems/memory/short_term/circulatory/github"
	helicalLogFile  = "github-mcp-server.log"
)

// HelicalEvent represents a 7D context event for helical memory
// EXTENT: Single event
// Compatible with HDSA and 7D context
// Add additional fields as needed for quantum/AI context
// Meta can hold entropy, compression, or parity info

type HelicalEvent struct {
	Who    string                 `json:"who"`
	What   string                 `json:"what"`
	When   int64                  `json:"when"`
	Where  string                 `json:"where"`
	Why    string                 `json:"why"`
	How    string                 `json:"how"`
	Extent string                 `json:"extent"`
	TS     int64                  `json:"ts"`
	Msg    string                 `json:"msg"`
	Meta   map[string]interface{} `json:"meta,omitempty"`
}

// LogHelicalEvent logs a 7D context event to both short_term.log and long_term.log (only in blood-connected mode)
func LogHelicalEvent(event HelicalEvent, logger LoggerInterface) error {
	if helicalMode == "standalone" {
		// Log directly to the main server log in standalone mode
		b, err := json.Marshal(event)
		if err != nil {
			logger.Info("Failed to marshal helical event: %v", err)
			return err
		}
		logger.Info("[HELICAL] %s", string(b))
		return nil
	}

	ensureHelicalMemoryDir()
	b, err := json.Marshal(event)
	if err != nil {
		logger.Info("Failed to marshal helical event: %v", err)
		return err
	}
	line := string(b) + "\n"
	fpath := filepath.Join(helicalMemoryDir, helicalLogFile)
	f, err := os.OpenFile(fpath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		logger.Info("helical memory open %s: %v", helicalLogFile, err)
		return fmt.Errorf("helical memory open %s: %w", helicalLogFile, err)
	}
	if _, err := f.WriteString(line); err != nil {
		f.Close()
		logger.Info("helical memory write %s: %v", helicalLogFile, err)
		return fmt.Errorf("helical memory write %s: %w", helicalLogFile, err)
	}
	f.Close()
	return nil
}

// NewHelicalEvent creates a HelicalEvent from 7D context fields
func NewHelicalEvent(who, what, where, why, how, extent, msg string) HelicalEvent {
	ts := time.Now().Unix()
	return HelicalEvent{
		Who:    who,
		What:   what,
		When:   ts,
		Where:  where,
		Why:    why,
		How:    how,
		Extent: extent,
		TS:     ts,
		Msg:    msg,
	}
}

func ensureHelicalMemoryDir() {
	_ = os.MkdirAll(helicalMemoryDir, 0755)
}

// SetHelicalMemoryMode sets the operational mode for helical memory logging
func SetHelicalMemoryMode(mode string) {
	helicalMode = mode
}
