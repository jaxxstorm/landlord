package ecs

import (
	"encoding/json"
	"testing"
)

func TestValidateConfigRequiresFields(t *testing.T) {
	p := New(nil, nil)

	if err := p.ValidateConfig(json.RawMessage(`{"task_definition_arn":"arn"}`)); err == nil {
		t.Fatalf("expected error for missing cluster_arn")
	}

	if err := p.ValidateConfig(json.RawMessage(`{"cluster_arn":"arn"}`)); err == nil {
		t.Fatalf("expected error for missing task_definition_arn")
	}

	if err := p.ValidateConfig(json.RawMessage(`{"cluster_arn":"arn","task_definition_arn":"arn"}`)); err == nil {
		t.Fatalf("expected error for missing service_name or service_name_prefix")
	}
}

func TestValidateConfigAcceptsValidConfig(t *testing.T) {
	p := New(nil, nil)

	cfg := json.RawMessage(`{
		"cluster_arn":"arn:aws:ecs:us-west-2:123456789012:cluster/example",
		"task_definition_arn":"arn:aws:ecs:us-west-2:123456789012:task-definition/example:1",
		"service_name_prefix": "landlord-tenant-",
		"desired_count": 1,
		"launch_type": "FARGATE"
	}`)

	if err := p.ValidateConfig(cfg); err != nil {
		t.Fatalf("expected config to be valid: %v", err)
	}
}

func TestConfigSchemaIncludesRequiredFields(t *testing.T) {
	p := New(nil, nil)

	var schema map[string]interface{}
	if err := json.Unmarshal(p.ConfigSchema(), &schema); err != nil {
		t.Fatalf("expected schema to be valid JSON: %v", err)
	}

	required, ok := schema["required"].([]interface{})
	if !ok {
		t.Fatalf("expected required to be an array")
	}

	seen := map[string]bool{}
	for _, item := range required {
		if value, ok := item.(string); ok {
			seen[value] = true
		}
	}

	if !seen["cluster_arn"] || !seen["task_definition_arn"] {
		t.Fatalf("expected required fields to include cluster_arn and task_definition_arn")
	}

	anyOf, ok := schema["anyOf"].([]interface{})
	if !ok {
		t.Fatalf("expected anyOf to be an array")
	}
	if len(anyOf) == 0 {
		t.Fatalf("expected anyOf to include service name requirements")
	}
	foundServiceName := false
	foundServiceNamePrefix := false
	for _, entry := range anyOf {
		entryMap, ok := entry.(map[string]interface{})
		if !ok {
			continue
		}
		requiredEntry, ok := entryMap["required"].([]interface{})
		if !ok {
			continue
		}
		for _, item := range requiredEntry {
			if value, ok := item.(string); ok {
				if value == "service_name" {
					foundServiceName = true
				}
				if value == "service_name_prefix" {
					foundServiceNamePrefix = true
				}
			}
		}
	}
	if !foundServiceName || !foundServiceNamePrefix {
		t.Fatalf("expected anyOf to require service_name or service_name_prefix")
	}
}
