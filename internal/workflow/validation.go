package workflow

import (
	"encoding/json"
	"fmt"
	"regexp"
)

var (
	workflowIDPattern    = regexp.MustCompile(`^[a-z0-9-]{1,128}$`)
	executionNamePattern = regexp.MustCompile(`^[a-zA-Z0-9-_]{1,128}$`)
)

// ValidateWorkflowSpec validates a workflow specification
func ValidateWorkflowSpec(spec *WorkflowSpec) error {
	if spec == nil {
		return fmt.Errorf("spec cannot be nil")
	}

	// Validate WorkflowID
	if spec.WorkflowID == "" {
		return fmt.Errorf("workflow_id is required")
	}
	if !workflowIDPattern.MatchString(spec.WorkflowID) {
		return fmt.Errorf("workflow_id must match pattern ^[a-z0-9-]{1,128}$")
	}

	// Validate ProviderType
	if spec.ProviderType == "" {
		return fmt.Errorf("provider_type is required")
	}

	// Validate Name
	if spec.Name == "" {
		return fmt.Errorf("name is required")
	}
	if len(spec.Name) > 256 {
		return fmt.Errorf("name cannot exceed 256 characters")
	}

	// Validate Definition
	if len(spec.Definition) == 0 {
		return fmt.Errorf("definition is required")
	}
	if !json.Valid(spec.Definition) {
		return fmt.Errorf("definition must be valid JSON")
	}

	// Validate Timeout if set
	if spec.Timeout != 0 && spec.Timeout <= 0 {
		return fmt.Errorf("timeout must be positive if set")
	}

	// Validate RetryPolicy if set
	if spec.RetryPolicy != nil {
		if spec.RetryPolicy.BackoffRate < 1.0 {
			return fmt.Errorf("retry_policy.backoff_rate must be >= 1.0")
		}
	}

	return nil
}

// ValidateExecutionInput validates execution input
func ValidateExecutionInput(input *ExecutionInput) error {
	if input == nil {
		return fmt.Errorf("input cannot be nil")
	}

	// Validate ExecutionName if set
	if input.ExecutionName != "" {
		if !executionNamePattern.MatchString(input.ExecutionName) {
			return fmt.Errorf("execution_name must match pattern ^[a-zA-Z0-9-_]{1,128}$")
		}
	}

	// Validate Input JSON
	if len(input.Input) == 0 {
		return fmt.Errorf("input is required")
	}
	if !json.Valid(input.Input) {
		return fmt.Errorf("input must be valid JSON")
	}

	return nil
}
