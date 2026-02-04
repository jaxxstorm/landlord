package tenant

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
)

// ComputeConfigHash computes a SHA256 hash of the compute configuration.
// This hash is used to detect when a tenant's config has changed, enabling
// the reconciler to restart workflows with updated configuration.
//
// The hash is deterministic - identical configs produce identical hashes.
// Returns empty string if config is nil or empty (for consistency).
func ComputeConfigHash(config interface{}) (string, error) {
	// Handle nil/empty config consistently
	if config == nil {
		return "", nil
	}

	// Marshal to JSON for deterministic representation
	configJSON, err := json.Marshal(config)
	if err != nil {
		return "", fmt.Errorf("failed to marshal config: %w", err)
	}

	// Return empty hash for empty JSON objects/arrays
	if len(configJSON) == 0 || string(configJSON) == "{}" || string(configJSON) == "[]" {
		return "", nil
	}

	// Compute SHA256 hash
	hash := sha256.Sum256(configJSON)
	return hex.EncodeToString(hash[:]), nil
}
