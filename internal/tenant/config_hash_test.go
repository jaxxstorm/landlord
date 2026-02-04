package tenant

import (
	"testing"
)

func TestComputeConfigHash_IdenticalConfigs(t *testing.T) {
	config1 := map[string]interface{}{
		"image": "nginx:1.25",
		"env": map[string]string{
			"FOO": "bar",
			"DEBUG": "true",
		},
	}

	config2 := map[string]interface{}{
		"image": "nginx:1.25",
		"env": map[string]string{
			"FOO": "bar",
			"DEBUG": "true",
		},
	}

	hash1, err := ComputeConfigHash(config1)
	if err != nil {
		t.Fatalf("ComputeConfigHash failed: %v", err)
	}

	hash2, err := ComputeConfigHash(config2)
	if err != nil {
		t.Fatalf("ComputeConfigHash failed: %v", err)
	}

	if hash1 != hash2 {
		t.Errorf("Expected identical hashes for identical configs, got %s and %s", hash1, hash2)
	}

	if hash1 == "" {
		t.Error("Expected non-empty hash for non-empty config")
	}
}

func TestComputeConfigHash_DifferentConfigs(t *testing.T) {
	config1 := map[string]interface{}{
		"image": "nginx:1.25",
	}

	config2 := map[string]interface{}{
		"image": "nginx:1.26",
	}

	hash1, err := ComputeConfigHash(config1)
	if err != nil {
		t.Fatalf("ComputeConfigHash failed: %v", err)
	}

	hash2, err := ComputeConfigHash(config2)
	if err != nil {
		t.Fatalf("ComputeConfigHash failed: %v", err)
	}

	if hash1 == hash2 {
		t.Errorf("Expected different hashes for different configs, got %s", hash1)
	}
}

func TestComputeConfigHash_MinorChanges(t *testing.T) {
	config1 := map[string]interface{}{
		"env": map[string]string{
			"FOO": "bar",
		},
	}

	config2 := map[string]interface{}{
		"env": map[string]string{
			"FOO": "baz",
		},
	}

	hash1, err := ComputeConfigHash(config1)
	if err != nil {
		t.Fatalf("ComputeConfigHash failed: %v", err)
	}

	hash2, err := ComputeConfigHash(config2)
	if err != nil {
		t.Fatalf("ComputeConfigHash failed: %v", err)
	}

	if hash1 == hash2 {
		t.Errorf("Expected different hashes for configs with different env values, got %s", hash1)
	}
}

func TestComputeConfigHash_NullAndEmpty(t *testing.T) {
	tests := []struct {
		name   string
		config interface{}
	}{
		{"nil config", nil},
		{"empty map", map[string]interface{}{}},
		{"empty slice", []interface{}{}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hash, err := ComputeConfigHash(tt.config)
			if err != nil {
				t.Fatalf("ComputeConfigHash failed: %v", err)
			}

			if hash != "" {
				t.Errorf("Expected empty hash for %s, got %s", tt.name, hash)
			}
		})
	}
}

func TestComputeConfigHash_NullConsistency(t *testing.T) {
	// All nil/empty configs should produce the same (empty) hash
	hash1, _ := ComputeConfigHash(nil)
	hash2, _ := ComputeConfigHash(map[string]interface{}{})
	hash3, _ := ComputeConfigHash([]interface{}{})

	if hash1 != hash2 || hash2 != hash3 {
		t.Errorf("Expected consistent empty hashes for nil/empty configs, got %s, %s, %s", hash1, hash2, hash3)
	}
}
