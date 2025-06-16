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

// LogHelicalEvent logs a 7D context event.
// In "standalone" mode, it marshals the event and logs it as a JSON string using the provided logger's Info method.
// In "blood-connected" mode, it marshals the event and writes it to the dedicated helical memory log file.
func LogHelicalEvent(event HelicalEvent, logger LoggerInterface) error {
	fmt.Fprintf(os.Stderr, "[DIAGNOSTIC] LogHelicalEvent called with helicalMode=%s\n", helicalMode)
	b, err := json.Marshal(event)
	if err != nil {
		if logger != nil {
			logger.Error("Failed to marshal helical event: %v", err)
		}
		return fmt.Errorf("failed to marshal helical event: %w", err)
	}
	jsonEvent := string(b)

	if helicalMode == "standalone" {
		fpath := "/Users/Jubicudis/Tranquility-Neuro-OS/github-mcp-server/pkg/log/github-mcp-server.log"
		fmt.Fprintf(os.Stderr, "[DIAGNOSTIC] Writing helical event to standalone log: %s\n", fpath)
		f, err := os.OpenFile(fpath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			if logger != nil {
				logger.Error("Failed to open standalone helical memory file %s: %v", fpath, err)
			}
			return fmt.Errorf("standalone helical memory open %s: %w", fpath, err)
		}
		defer f.Close()
		if _, err := f.WriteString(jsonEvent + "\n"); err != nil {
			if logger != nil {
				logger.Error("Failed to write to standalone helical memory file %s: %v", fpath, err)
			}
			return fmt.Errorf("standalone helical memory write %s: %w", fpath, err)
		}
		return nil
	}

	fmt.Fprintf(os.Stderr, "[DIAGNOSTIC] Writing helical event to blood-connected log: %s\n", filepath.Join(helicalMemoryDir, helicalLogFile))
	if err := ensureHelicalMemoryDir(); err != nil {
		if logger != nil {
			logger.Error("Failed to ensure helical memory directory %s: %v", helicalMemoryDir, err)
		}
		return fmt.Errorf("failed to ensure helical memory directory %s: %w", helicalMemoryDir, err)
	}

	fpath := filepath.Join(helicalMemoryDir, helicalLogFile)
	f, err := os.OpenFile(fpath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		if logger != nil {
			logger.Error("Failed to open helical memory file %s: %v", fpath, err)
		}
		return fmt.Errorf("helical memory open %s: %w", fpath, err)
	}
	defer f.Close()

	if _, err := f.WriteString(jsonEvent + "\n"); err != nil {
		if logger != nil {
			logger.Error("Failed to write to helical memory file %s: %v", fpath, err)
		}
		return fmt.Errorf("helical memory write %s: %w", fpath, err)
	}
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

func ensureHelicalMemoryDir() error { // Changed to return an error
	if err := os.MkdirAll(helicalMemoryDir, 0755); err != nil {
		return fmt.Errorf("failed to create helical memory directory %s: %w", helicalMemoryDir, err)
	}
	return nil
}

// SetHelicalMemoryMode sets the operational mode for helical memory logging.
// Valid modes are "standalone" and "blood-connected".
func SetHelicalMemoryMode(mode string) {
	fmt.Fprintf(os.Stderr, "[DIAGNOSTIC] SetHelicalMemoryMode called with mode=%s\n", mode)
	if mode == "standalone" || mode == "blood-connected" {
		helicalMode = mode
		fmt.Fprintf(os.Stderr, "[DIAGNOSTIC] Helical memory mode set to %s\n", helicalMode)
	} else {
		// Optionally log an error if an invalid mode is set, using a default/global logger if available,
		// or print to stderr if no logger is accessible here.
		fmt.Fprintf(os.Stderr, "Warning: Invalid helical memory mode specified: %s. Defaulting to 'standalone'.\n", mode)
		helicalMode = "standalone" // Default to standalone if an invalid mode is provided.
	}
}
