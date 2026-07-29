package ec2runner

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/jaxxstorm/landlord/internal/compute"
)

const launchFingerprintTag = "landlord.ec2_runner.launch_fingerprint"

// LaunchConfig is the tenant-facing EC2 runner configuration.
type LaunchConfig struct {
	AMI      string `json:"ami"`
	UserData string `json:"user_data"`
}

var launchConfigSchema = json.RawMessage(`{
  "$schema": "https://json-schema.org/draft/2020-12/schema",
  "type": "object",
  "properties": {
    "ami": { "type": "string", "minLength": 1 },
    "user_data": { "type": "string", "minLength": 1 }
  },
  "required": ["ami", "user_data"],
  "additionalProperties": false
}`)

func parseLaunchConfig(raw json.RawMessage) (*LaunchConfig, error) {
	if len(raw) == 0 {
		return nil, fmt.Errorf("EC2 runner config is required")
	}
	var cfg LaunchConfig
	if err := json.Unmarshal(raw, &cfg); err != nil {
		return nil, fmt.Errorf("invalid EC2 runner config: %w", err)
	}
	if strings.TrimSpace(cfg.AMI) == "" {
		return nil, fmt.Errorf("ami is required")
	}
	if strings.TrimSpace(cfg.UserData) == "" {
		return nil, fmt.Errorf("user_data is required")
	}
	return &cfg, nil
}

func launchFingerprint(cfg *LaunchConfig) string {
	if cfg == nil {
		return ""
	}
	sum := sha256.Sum256([]byte(cfg.AMI + "\x00" + cfg.UserData))
	return hex.EncodeToString(sum[:])
}

func managedTags(spec *compute.TenantComputeSpec, cfg *LaunchConfig) []types.Tag {
	labels := compute.MergeLabels(spec.Labels, compute.DefaultMetadata(spec))
	labels[launchFingerprintTag] = launchFingerprint(cfg)
	tags := make([]types.Tag, 0, len(labels))
	for key, value := range labels {
		key, value := key, value
		tags = append(tags, types.Tag{Key: &key, Value: &value})
	}
	return tags
}
