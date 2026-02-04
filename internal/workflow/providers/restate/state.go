package restate

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/jaxxstorm/landlord/internal/workflow"
)

// mapInvocationSubState maps Restate status strings to workflow.WorkflowSubState
// mapInvocationSubState maps Restate status strings to workflow.WorkflowSubState
func mapInvocationSubState(status string) workflow.WorkflowSubState {
	normalized := strings.ToLower(strings.TrimSpace(status))
	
	// Use exact matching for known Restate status values
	switch normalized {
	case "backing-off", "backoff", "retrying", "retry":
		return workflow.SubStateBackingOff
	case "suspended":
		return workflow.SubStateWaiting
	case "completed", "succeeded", "success":
		return workflow.SubStateSucceeded
	case "failed", "error":
		return workflow.SubStateFailed
	case "pending":
		return workflow.SubStateWaiting
	case "running", "active":
		return workflow.SubStateRunning
	default:
		// Default to running for unknown states
		return workflow.SubStateRunning
	}
}

// extractInvocationMetadata pulls retry/backoff metadata from Restate responses.
// Also extracts user-provided metadata like config_hash from the original input.
func extractInvocationMetadata(payload map[string]interface{}) map[string]string {
	if len(payload) == 0 {
		return nil
	}
	metadata := map[string]string{}

	if status, ok := payload["status"].(string); ok {
		metadata["restate_status"] = status
	}

	// Extract user-provided metadata (e.g., config_hash) from original input if available
	if input, ok := payload["input"].(map[string]interface{}); ok {
		if metadataMap, ok := input["metadata"].(map[string]interface{}); ok {
			for key, value := range metadataMap {
				if strValue, ok := value.(string); ok {
					metadata[key] = strValue
				} else {
					metadata[key] = fmt.Sprintf("%v", value)
				}
			}
		}
	}

	for _, key := range []string{"retry_count", "retryCount", "retry_attempts", "retryAttempts", "attempts"} {
		if value, ok := payload[key]; ok {
			metadata["retry_count"] = stringifyNumeric(value)
			break
		}
	}

	for _, key := range []string{"retry_state", "retryState", "backoff", "backoff_state", "next_retry_at"} {
		if value, ok := payload[key]; ok {
			metadata["retry_state"] = stringifyNumeric(value)
			break
		}
	}

	if len(metadata) == 0 {
		return nil
	}
	return metadata
}

func stringifyNumeric(value interface{}) string {
	switch typed := value.(type) {
	case string:
		return typed
	case float64:
		return strings.TrimRight(strings.TrimRight(fmt.Sprintf("%f", typed), "0"), ".")
	case int:
		return fmt.Sprintf("%d", typed)
	case int64:
		return fmt.Sprintf("%d", typed)
	case json.Number:
		return typed.String()
	default:
		return fmt.Sprintf("%v", typed)
	}
}
