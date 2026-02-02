package models

import "encoding/json"

// ComputeConfigDiscoveryResponse represents the compute config discovery response.
type ComputeConfigDiscoveryResponse struct {
	// Provider is the active compute provider identifier (e.g., "docker").
	Provider string `json:"provider"`

	// Schema is the JSON Schema (draft 2020-12) for compute_config.
	Schema json.RawMessage `json:"schema"`

	// Defaults is an optional defaults object for compute_config.
	Defaults json.RawMessage `json:"defaults,omitempty"`
}
