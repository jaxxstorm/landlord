package ecs

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/jaxxstorm/landlord/internal/compute"
)

// ComputeConfig represents ECS provider configuration stored in compute_config.
type ComputeConfig struct {
	Region          string            `json:"region,omitempty"`
	ClusterARN      string            `json:"cluster_arn"`
	TaskDefinition  string            `json:"task_definition_arn"`
	ServiceName     string            `json:"service_name,omitempty"`
	ServiceNamePref string            `json:"service_name_prefix,omitempty"`
	DesiredCount    *int32            `json:"desired_count,omitempty"`
	LaunchType      string            `json:"launch_type,omitempty"`
	Subnets         []string          `json:"subnets,omitempty"`
	SecurityGroups  []string          `json:"security_groups,omitempty"`
	AssignPublicIP  *bool             `json:"assign_public_ip,omitempty"`
	Tags            map[string]string `json:"tags,omitempty"`
	AssumeRole      *AssumeRoleConfig `json:"assume_role,omitempty"`
}

// AssumeRoleConfig configures optional assume-role behavior for ECS calls.
type AssumeRoleConfig struct {
	RoleARN     string `json:"role_arn"`
	ExternalID  string `json:"external_id,omitempty"`
	SessionName string `json:"session_name,omitempty"`
}

var ecsConfigSchema = json.RawMessage(`{
  "$schema": "https://json-schema.org/draft/2020-12/schema",
  "type": "object",
  "properties": {
    "region": { "type": "string" },
    "cluster_arn": { "type": "string" },
    "task_definition_arn": { "type": "string" },
    "service_name": { "type": "string" },
    "service_name_prefix": { "type": "string" },
    "desired_count": { "type": "integer", "minimum": 0 },
    "launch_type": { "type": "string", "enum": ["EC2", "FARGATE", "EXTERNAL"] },
    "subnets": {
      "type": "array",
      "items": { "type": "string" }
    },
    "security_groups": {
      "type": "array",
      "items": { "type": "string" }
    },
    "assign_public_ip": { "type": "boolean" },
    "tags": {
      "type": "object",
      "additionalProperties": { "type": "string" }
    },
    "assume_role": {
      "type": "object",
      "properties": {
        "role_arn": { "type": "string" },
        "external_id": { "type": "string" },
        "session_name": { "type": "string" }
      },
      "required": ["role_arn"],
      "additionalProperties": false
    }
  },
  "required": ["cluster_arn", "task_definition_arn"],
  "additionalProperties": true
}`)

func parseComputeConfig(raw json.RawMessage, defaults map[string]interface{}) (*ComputeConfig, error) {
	if len(raw) == 0 {
		if len(defaults) == 0 {
			return nil, fmt.Errorf("ecs config is required")
		}
	}

	mergedRaw, err := compute.MergeConfigJSON(defaults, raw)
	if err != nil {
		return nil, fmt.Errorf("merge ecs config: %w", err)
	}

	var cfg ComputeConfig
	if err := json.Unmarshal(mergedRaw, &cfg); err != nil {
		return nil, fmt.Errorf("invalid JSON structure: %w", err)
	}

	if err := validateComputeConfig(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

func validateComputeConfig(cfg *ComputeConfig) error {
	if cfg.ClusterARN == "" {
		return fmt.Errorf("cluster_arn is required")
	}
	if cfg.TaskDefinition == "" {
		return fmt.Errorf("task_definition_arn is required")
	}
	if cfg.DesiredCount != nil && *cfg.DesiredCount < 0 {
		return fmt.Errorf("desired_count must be >= 0")
	}
	if cfg.LaunchType != "" {
		cfg.LaunchType = strings.ToUpper(cfg.LaunchType)
		switch cfg.LaunchType {
		case "EC2", "FARGATE", "EXTERNAL":
			// ok
		default:
			return fmt.Errorf("launch_type must be EC2, FARGATE, or EXTERNAL")
		}
	}
	if cfg.AssumeRole != nil && cfg.AssumeRole.RoleARN == "" {
		return fmt.Errorf("assume_role.role_arn is required")
	}
	return nil
}

func resolveServiceName(cfg *ComputeConfig, tenantID string) string {
	if cfg == nil {
		return ""
	}
	if cfg.ServiceName != "" {
		return cfg.ServiceName
	}
	prefix := cfg.ServiceNamePref
	if prefix == "" {
		prefix = "landlord-tenant-"
	}
	return fmt.Sprintf("%s%s", prefix, tenantID)
}

func resolveRegion(cfg *ComputeConfig) string {
	if cfg == nil {
		return ""
	}
	if cfg.Region != "" {
		return cfg.Region
	}
	if region := regionFromARN(cfg.ClusterARN); region != "" {
		return region
	}
	return regionFromARN(cfg.TaskDefinition)
}

func regionFromARN(arn string) string {
	parts := strings.Split(arn, ":")
	if len(parts) < 4 {
		return ""
	}
	return parts[3]
}
