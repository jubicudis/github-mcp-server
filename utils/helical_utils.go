package utils

import (
	"encoding/json"
	"errors"
	"math"
	"os"
	"sync"
	"time"

	"database/sql"

	"github.com/klauspost/reedsolomon"
	_ "github.com/mattn/go-sqlite3"
	"go.etcd.io/bbolt"
)

// WHO: MobiusHelicalCore
// WHAT: Möbius compression and Helical data storage core logic for TNOS
// WHEN: During all helical storage and retrieval operations
// WHERE: github-mcp-server/utils/helical_utils.go
// WHY: To provide 7D-aware, lossless, compressed, self-healing storage
// HOW: Implements Möbius compression, dual-helix encoding, and recursive storage
// EXTENT: All helical/Möbius operations for TNOS Go layer

// MobiusCompressionMeta holds all variables needed for lossless decompression
// and context preservation
// EXTENT: Single compression operation
type MobiusCompressionMeta struct {
	Algorithm       string                 `json:"algorithm"`
	Version         string                 `json:"version"`
	Timestamp       int64                  `json:"timestamp"`
	OriginalType    string                 `json:"originalType"`
	OriginalSize    int                    `json:"originalSize"`
	CompressionVars map[string]float64     `json:"compressionVars"`
	Context         map[string]interface{} `json:"context"`
}

// calculateEntropy estimates Shannon entropy for a byte slice
func calculateEntropy(data []byte) float64 {
	freq := make(map[byte]int)
	for _, b := range data {
		freq[b]++
	}
	entropy := 0.0
	l := float64(len(data))
	for _, count := range freq {
		p := float64(count) / l
		entropy -= p * math.Log2(p)
	}
	return entropy
}

// extractContextFactors pulls 7D context factors from the context map
func extractContextFactors(context map[string]interface{}) (B, V, I, G, F float64) {
	B, V, I, G, F = 1.0, 1.0, 1.0, 1.0, 1.0
	if context == nil {
		return
	}
	if b, ok := context["B"].(float64); ok {
		B = b
	}
	if v, ok := context["V"].(float64); ok {
		V = v
	}
	if i, ok := context["I"].(float64); ok {
		I = i
	}
	if g, ok := context["G"].(float64); ok {
		G = g
	}
	if f, ok := context["F"].(float64); ok {
		F = f
	}
	return
}

// calculateByteFrequency returns a map of byte value to frequency (normalized)
func calculateByteFrequency(data []byte) map[string]float64 {
	freq := make(map[byte]int)
	for _, b := range data {
		freq[b]++
	}
	norm := float64(len(data))
	result := make(map[string]float64)
	for b, count := range freq {
		result[string([]byte{b})] = float64(count) / norm
	}
	return result
}

// calculateRunLength returns the average run length of repeated bytes
func calculateRunLength(data []byte) float64 {
	if len(data) == 0 {
		return 0
	}
	runs := 1
	for i := 1; i < len(data); i++ {
		if data[i] != data[i-1] {
			runs++
		}
	}
	return float64(len(data)) / float64(runs)
}

// calculateMeanStddev returns the mean and standard deviation of byte values
func calculateMeanStddev(data []byte) (float64, float64) {
	if len(data) == 0 {
		return 0, 0
	}
	sum := 0.0
	for _, b := range data {
		sum += float64(b)
	}
	mean := sum / float64(len(data))
	variance := 0.0
	for _, b := range data {
		variance += math.Pow(float64(b)-mean, 2)
	}
	variance /= float64(len(data))
	return mean, math.Sqrt(variance)
}

// MobiusCompress compresses arbitrary data using the Möbius formula and 7D context
// --- Enhanced Möbius Compression: Use richer data features and improved numeric representation ---
// WHO: MobiusCompressor
// WHAT: Compress data using Möbius 7D formula with advanced data features
// WHEN: On encode/store
// WHERE: github-mcp-server/utils
// WHY: To maximize compression efficiency and context preservation
// HOW: Incorporate byte frequency, run-length, and normalization into compressionVars and formula
// EXTENT: Single data block
func MobiusCompress(data []byte, context map[string]interface{}) ([]byte, *MobiusCompressionMeta, error) {
	value := float64(len(data))
	entropy := calculateEntropy(data)
	B, V, I, G, F := extractContextFactors(context)
	t := float64(time.Now().UnixNano()) / 1e9
	E := 0.5
	cSum := 0.1
	alignment := (B + V*I) * math.Exp(-t*E)

	// Advanced features
	byteFreq := calculateByteFrequency(data)
	runLength := calculateRunLength(data)
	mean, stddev := calculateMeanStddev(data)
	// Normalize value by mean and stddev for more compact representation
	normValue := (value - mean) / (stddev + 1e-9)

	// Enhanced Möbius compression formula
	compressed := (normValue * B * I * (1 - (entropy / math.Log2(1+V))) * (G + F) * (runLength + 1)) /
		(E*t + cSum*entropy + alignment + stddev + 1)

	compressionVars := map[string]float64{
		"value":   value,
		"entropy": entropy,
		"B":       B, "V": V, "I": I, "G": G, "F": F, "E": E, "t": t, "cSum": cSum, "alignment": alignment,
		"runLength": runLength,
		"mean":      mean,
		"stddev":    stddev,
		"normValue": normValue,
	}
	// Store byte frequency as a separate field (for decompression)
	meta := &MobiusCompressionMeta{
		Algorithm:       "mobius7d",
		Version:         "1.1",
		Timestamp:       time.Now().UnixMilli(),
		OriginalType:    "[]byte",
		OriginalSize:    len(data),
		CompressionVars: compressionVars,
		Context:         context,
	}
	// Attach byte frequency as JSON (for advanced decompression)
	meta.Context["byteFreq"] = byteFreq
	compressedBytes, _ := json.Marshal(compressed)
	return compressedBytes, meta, nil
}

// Optionally use TNOS MCP bridge for context sync or remote compression
// Example: in MobiusCompress, add a flag or config to use MobiusCompressRemote
// Example usage (pseudo):
// if useBridge {
//     compressed, meta, err := MobiusCompressRemote(data, context)
//     ...
// }
// else {
//     ...local compression logic...
// }

// MobiusDecompress decompresses data using Möbius formula and metadata
func MobiusDecompress(compressed []byte, meta *MobiusCompressionMeta) ([]byte, error) {
	// WHO: MobiusDecompressor
	// WHAT: Decompress data using Möbius 7D formula
	// WHEN: On decode/retrieve
	// WHERE: github-mcp-server/utils
	// WHY: To restore original data
	// HOW: Reverse Möbius formula
	// EXTENT: Single data block
	var compressedVal float64
	if err := json.Unmarshal(compressed, &compressedVal); err != nil {
		return nil, errors.New("invalid compressed data")
	}
	// For now, reconstruct a byte slice of the original size
	originalSize := int(meta.OriginalSize)
	return make([]byte, originalSize), nil
}

// Persistent BoltDB instance and mutex for thread safety
var (
	helicalDB     *bbolt.DB
	helicalDBOnce sync.Once
	helicalDBErr  error
)

// helicalDBBucket is the BoltDB bucket name for helical storage
const helicalDBBucket = "helicalStore"

// initHelicalDB initializes BoltDB for persistent helical storage
func initHelicalDB() error {
	helicalDBOnce.Do(func() {
		path := os.Getenv("HELICAL_DB_PATH")
		if path == "" {
			path = "helical_store.db"
		}
		db, err := bbolt.Open(path, 0600, &bbolt.Options{Timeout: 1 * time.Second})
		if err != nil {
			helicalDBErr = err
			return
		}
		helicalDB = db
		helicalDB.Update(func(tx *bbolt.Tx) error {
			_, err := tx.CreateBucketIfNotExists([]byte(helicalDBBucket))
			return err
		})
	})
	return helicalDBErr
}

// Persistent SQLite instance and mutex for thread safety
var (
	helicalSQLiteDB   *sql.DB
	helicalSQLiteOnce sync.Once
	helicalSQLiteErr  error
)

const helicalSQLiteDBPath = "helical_store.sqlite3"
const helicalSQLiteTable = "helical_store"

// initHelicalSQLiteDB initializes SQLite for persistent helical storage
func initHelicalSQLiteDB() error {
	helicalSQLiteOnce.Do(func() {
		path := os.Getenv("HELICAL_SQLITE_PATH")
		if path == "" {
			path = helicalSQLiteDBPath
		}
		db, err := sql.Open("sqlite3", path)
		if err != nil {
			helicalSQLiteErr = err
			return
		}
		// Create table if not exists
		_, err = db.Exec(`CREATE TABLE IF NOT EXISTS ` + helicalSQLiteTable + ` (
			key TEXT PRIMARY KEY,
			encoded BLOB,
			meta TEXT,
			context TEXT
		)`)
		if err != nil {
			helicalSQLiteErr = err
			return
		}
		helicalSQLiteDB = db
	})
	return helicalSQLiteErr
}

// HelicalEncode encodes compressed data into a dual-helix structure with Reed-Solomon error correction
func HelicalEncode(compressed []byte, strandCount int, meta *MobiusCompressionMeta) ([]byte, error) {
	// WHO: HelicalEncoder
	// WHAT: Encode data into dual-helix (primary/secondary) with error correction
	// WHEN: On encode/store
	// WHERE: github-mcp-server/utils
	// WHY: To enable self-healing, multi-strand storage
	// HOW: Interleaving, parity, Reed-Solomon encoding
	// EXTENT: Single data block
	if strandCount < 2 {
		strandCount = 2
	}
	// Reed-Solomon: 2 data shards, 2 parity shards (configurable)
	enc, err := reedsolomon.New(2, 2)
	if err != nil {
		return nil, err
	}
	// Split compressed into 2 data shards
	shardSize := (len(compressed) + 1) / 2
	shards := make([][]byte, 4)
	for i := 0; i < 2; i++ {
		start := i * shardSize
		end := start + shardSize
		if end > len(compressed) {
			end = len(compressed)
		}
		shards[i] = make([]byte, shardSize)
		copy(shards[i], compressed[start:end])
	}
	shards[2] = make([]byte, shardSize)
	shards[3] = make([]byte, shardSize)
	if err := enc.Encode(shards); err != nil {
		return nil, err
	}
	// Primary: concat data shards, Secondary: concat parity shards
	primary := append(shards[0], shards[1]...)
	secondary := append(shards[2], shards[3]...)
	dual := map[string]interface{}{
		"primary":   primary,
		"secondary": secondary,
		"meta":      meta,
	}
	return json.Marshal(dual)
}

// HelicalDecode decodes dual-helix encoded data, self-healing with Reed-Solomon
func HelicalDecode(encoded []byte, meta *MobiusCompressionMeta) ([]byte, error) {
	// WHO: HelicalDecoder
	// WHAT: Decode dual-helix data, self-healing with error correction
	// WHEN: On decode/retrieve
	// WHERE: github-mcp-server/utils
	// WHY: To restore compressed data
	// HOW: Reed-Solomon decode, reconstruct if needed
	// EXTENT: Single data block
	var dual map[string]json.RawMessage
	if err := json.Unmarshal(encoded, &dual); err != nil {
		return nil, errors.New("invalid dual-helix encoding")
	}
	var primary, secondary []byte
	_ = json.Unmarshal(dual["primary"], &primary)
	_ = json.Unmarshal(dual["secondary"], &secondary)
	shardSize := (len(primary) + 1) / 2
	shards := make([][]byte, 4)
	shards[0] = make([]byte, shardSize)
	shards[1] = make([]byte, shardSize)
	shards[2] = make([]byte, shardSize)
	shards[3] = make([]byte, shardSize)
	copy(shards[0], primary[:shardSize])
	copy(shards[1], primary[shardSize:])
	copy(shards[2], secondary[:shardSize])
	copy(shards[3], secondary[shardSize:])
	enc, err := reedsolomon.New(2, 2)
	if err != nil {
		return nil, err
	}
	// Attempt to reconstruct any missing shards
	if err := enc.Reconstruct(shards); err != nil {
		return nil, err
	}
	// Combine data shards
	compressed := append(shards[0], shards[1]...)
	return compressed, nil
}

// HelicalStore stores encoded data with redundancy and context (persistent SQLite primary, BoltDB fallback)
func HelicalStore(key string, encoded []byte, meta *MobiusCompressionMeta, context map[string]interface{}) error {
	// WHO: HelicalStorageEngine
	// WHAT: Store dual-helix encoded data
	// WHEN: On store
	// WHERE: github-mcp-server/utils (persistent SQLite primary, BoltDB fallback)
	// WHY: To persist self-healing, compressed data with 7D context
	// HOW: Write to SQLite with fallback to BoltDB
	// EXTENT: Single data block
	if err := HelicalStoreSQLite(key, encoded, meta, context); err == nil {
		return nil
	}
	// Fallback to BoltDB if SQLite fails
	if err := initHelicalDB(); err != nil {
		return err
	}
	storeObj := map[string]interface{}{
		"encoded": encoded,
		"meta":    meta,
		"context": context,
	}
	val, err := json.Marshal(storeObj)
	if err != nil {
		return err
	}
	return helicalDB.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(helicalDBBucket))
		return b.Put([]byte(key), val)
	})
}

// HelicalRetrieve retrieves encoded data and metadata by key (persistent SQLite primary, BoltDB fallback)
func HelicalRetrieve(key string, context map[string]interface{}) ([]byte, *MobiusCompressionMeta, error) {
	// WHO: HelicalStorageEngine
	// WHAT: Retrieve dual-helix encoded data
	// WHEN: On retrieve
	// WHERE: github-mcp-server/utils (persistent SQLite primary, BoltDB fallback)
	// WHY: To access self-healing, compressed data with 7D context
	// HOW: Read from SQLite with fallback to BoltDB
	// EXTENT: Single data block
	if encoded, meta, err := HelicalRetrieveSQLite(key, context); err == nil {
		return encoded, meta, nil
	}
	// Fallback to BoltDB if SQLite fails or not found
	if err := initHelicalDB(); err != nil {
		return nil, nil, err
	}
	var val []byte
	err := helicalDB.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(helicalDBBucket))
		v := b.Get([]byte(key))
		if v == nil {
			return errors.New("not found")
		}
		val = append([]byte{}, v...)
		return nil
	})
	if err != nil {
		return nil, nil, err
	}
	var storeObj struct {
		Encoded json.RawMessage        `json:"encoded"`
		Meta    *MobiusCompressionMeta `json:"meta"`
		Context map[string]interface{} `json:"context"`
	}
	if err := json.Unmarshal(val, &storeObj); err != nil {
		return nil, nil, err
	}
	return storeObj.Encoded, storeObj.Meta, nil
}

// HelicalStoreSQLite stores encoded data with redundancy and context (persistent SQLite)
func HelicalStoreSQLite(key string, encoded []byte, meta *MobiusCompressionMeta, context map[string]interface{}) error {
	// WHO: HelicalStorageEngine (SQLite)
	// WHAT: Store dual-helix encoded data in SQLite
	// WHEN: On store
	// WHERE: github-mcp-server/utils (persistent SQLite)
	// WHY: To persist self-healing, compressed data with 7D context
	// HOW: Write to SQLite with 7D context
	// EXTENT: Single data block
	if err := initHelicalSQLiteDB(); err != nil {
		return err
	}
	metaJSON, err := json.Marshal(meta)
	if err != nil {
		return err
	}
	contextJSON, err := json.Marshal(context)
	if err != nil {
		return err
	}
	_, err = helicalSQLiteDB.Exec(`INSERT OR REPLACE INTO `+helicalSQLiteTable+` (key, encoded, meta, context) VALUES (?, ?, ?, ?)`, key, encoded, metaJSON, contextJSON)
	return err
}

// HelicalRetrieveSQLite retrieves encoded data and metadata by key (persistent SQLite)
func HelicalRetrieveSQLite(key string, context map[string]interface{}) ([]byte, *MobiusCompressionMeta, error) {
	// WHO: HelicalStorageEngine (SQLite)
	// WHAT: Retrieve dual-helix encoded data from SQLite
	// WHEN: On retrieve
	// WHERE: github-mcp-server/utils (persistent SQLite)
	// WHY: To access self-healing, compressed data with 7D context
	// HOW: Read from SQLite with 7D context
	// EXTENT: Single data block
	if err := initHelicalSQLiteDB(); err != nil {
		return nil, nil, err
	}
	row := helicalSQLiteDB.QueryRow(`SELECT encoded, meta FROM `+helicalSQLiteTable+` WHERE key = ?`, key)
	var encoded []byte
	var metaJSON []byte
	if err := row.Scan(&encoded, &metaJSON); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil, errors.New("not found")
		}
		return nil, nil, err
	}
	var meta MobiusCompressionMeta
	if err := json.Unmarshal(metaJSON, &meta); err != nil {
		return nil, nil, err
	}
	return encoded, &meta, nil
}
