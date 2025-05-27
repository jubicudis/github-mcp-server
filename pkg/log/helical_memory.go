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
	"sync"
	"time"
)

var (
	helicalShortTerm  = "short_term.log"
	helicalLongTerm   = "long_term.log"
	helicalOnce       sync.Once
	resolvedMemoryDir string
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

// getMemoryDir returns the absolute path to the memory directory
func getMemoryDir() string {
	helicalOnce.Do(func() {
		// Try to resolve project root (github-mcp-server)
		execPath, err := os.Executable()
		if err == nil {
			root := filepath.Dir(execPath)
			// Look for github-mcp-server/memory
			for i := 0; i < 4; i++ {
				candidate := filepath.Join(root, "memory")
				if fi, err := os.Stat(candidate); err == nil && fi.IsDir() {
					resolvedMemoryDir = candidate
					return
				}
				root = filepath.Dir(root)
			}
		}
		// Fallback: try CWD/memory
		cwd, _ := os.Getwd()
		candidate := filepath.Join(cwd, "memory")
		if fi, err := os.Stat(candidate); err == nil && fi.IsDir() {
			resolvedMemoryDir = candidate
			return
		}
		// Last resort: use relative path
		resolvedMemoryDir = filepath.Join("..", "memory")
	})
	return resolvedMemoryDir
}

// ensureHelicalMemoryDir ensures the memory/ directory exists
func ensureHelicalMemoryDir() {
	dir := getMemoryDir()
	_ = os.MkdirAll(dir, 0755)
}

// LogHelicalEvent logs a 7D context event to both short_term.log and long_term.log
func LogHelicalEvent(event HelicalEvent) error {
	ensureHelicalMemoryDir()
	b, err := json.Marshal(event)
	if err != nil {
		return err
	}
	line := string(b) + "\n"
	for _, fname := range []string{helicalShortTerm, helicalLongTerm} {
		fpath := filepath.Join(getMemoryDir(), fname)
		f, err := os.OpenFile(fpath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			return fmt.Errorf("helical memory open %s: %w", fname, err)
		}
		if _, err := f.WriteString(line); err != nil {
			f.Close()
			return fmt.Errorf("helical memory write %s: %w", fname, err)
		}
		f.Close()
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
