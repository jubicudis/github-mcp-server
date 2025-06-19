// trust_table.go
// WHO: TrustTableManager
// WHAT: In-memory trust table for QHP handshake metadata
// WHEN: On every successful QHP handshake
// WHERE: System Layer 6 (Integration)
// WHY: To track trusted peers and handshake metadata
// HOW: Map of fingerprints to trust metadata
// EXTENT: All QHP-secured connections

package bridge

import (
	"sync"
	"time"
)

type TrustMetadata struct {
	Fingerprint string
	Timestamp  int64
	SessionKey string
	Meta       map[string]interface{}
}

type TrustTable struct {
	mu    sync.RWMutex
	peers map[string]TrustMetadata
}

var trustTableInstance *TrustTable
var trustTableOnce sync.Once

func GetTrustTable() *TrustTable {
	trustTableOnce.Do(func() {
		trustTableInstance = &TrustTable{
			peers: make(map[string]TrustMetadata),
		}
	})
	return trustTableInstance
}

func (t *TrustTable) Update(fingerprint, sessionKey string, meta map[string]interface{}) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.peers[fingerprint] = TrustMetadata{
		Fingerprint: fingerprint,
		Timestamp:  time.Now().Unix(),
		SessionKey: sessionKey,
		Meta:       meta,
	}
}

func (t *TrustTable) Get(fingerprint string) (TrustMetadata, bool) {
	t.mu.RLock()
	defer t.mu.RUnlock()
	meta, ok := t.peers[fingerprint]
	return meta, ok
}

func (t *TrustTable) All() map[string]TrustMetadata {
	t.mu.RLock()
	defer t.mu.RUnlock()
	copy := make(map[string]TrustMetadata)
	for k, v := range t.peers {
		copy[k] = v
	}
	return copy
}
