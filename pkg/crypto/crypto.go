// WHO: CryptoUtilities
// WHAT: Shared cryptographic utilities for TNOS
// WHEN: During security operations
// WHERE: System Layer 6 (Integration)
// WHY: To provide consistent cryptographic operations
// HOW: Using Go's crypto packages
// EXTENT: All cryptographic needs

package crypto

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"
)

// GenerateQuantumFingerprint creates a unique cryptographic identity for a node
// WHO: CryptoIDGenerator
// WHAT: Generate cryptographic identity
// WHEN: During QHP handshake
// WHERE: System Layer 6 (Integration)
// WHY: For secure peer identification
// HOW: Using SHA-256 with entropy
// EXTENT: All QHP operations
func GenerateQuantumFingerprint(nodeName string) string {
	entropy := make([]byte, 32)
	_, err := rand.Read(entropy)
	if err != nil {
		// Fall back to timestamp-based ID if entropy generation fails
		timestamp := []byte(fmt.Sprintf("%d", time.Now().UnixNano()))
		h := sha256.New()
		h.Write([]byte(nodeName))
		h.Write(timestamp)
		return hex.EncodeToString(h.Sum(nil))
	}
	
	h := sha256.New()
	h.Write([]byte(nodeName))
	h.Write(entropy)
	return hex.EncodeToString(h.Sum(nil))
}

// GenerateChallengeResponse creates a challenge response for QHP handshake
// WHO: ChallengeResponseGenerator
// WHAT: Generate secure challenge response
// WHEN: During QHP handshake
// WHERE: System Layer 6 (Integration)
// WHY: For secure verification
// HOW: Using SHA-256 with challenge and fingerprint
// EXTENT: All QHP operations
func GenerateChallengeResponse(challenge, fingerprint string) string {
	h := sha256.New()
	h.Write([]byte(challenge + fingerprint))
	return hex.EncodeToString(h.Sum(nil))
}

// VerifyChallengeResponse verifies a challenge response
// WHO: ChallengeResponseVerifier
// WHAT: Verify secure challenge response
// WHEN: During QHP handshake
// WHERE: System Layer 6 (Integration)
// WHY: For secure verification
// HOW: Using SHA-256 with challenge and fingerprint
// EXTENT: All QHP operations
func VerifyChallengeResponse(challenge, response, peerFingerprint string) bool {
	h := sha256.New()
	h.Write([]byte(challenge + peerFingerprint))
	expected := hex.EncodeToString(h.Sum(nil))
	return expected == response
}

// GenerateSessionKey creates a secure session key
// WHO: SessionKeyGenerator
// WHAT: Generate secure session key
// WHEN: During QHP handshake
// WHERE: System Layer 6 (Integration)
// WHY: For secure communication
// HOW: Using crypto/rand
// EXTENT: All QHP operations
func GenerateSessionKey(length int) (string, error) {
	if length <= 0 {
		length = 32 // Default length
	}
	
	keyBytes := make([]byte, length)
	_, err := rand.Read(keyBytes)
	if err != nil {
		return "", fmt.Errorf("failed to generate session key: %w", err)
	}
	
	return hex.EncodeToString(keyBytes), nil
}
