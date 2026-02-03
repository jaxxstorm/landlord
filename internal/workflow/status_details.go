package workflow

import (
	"strconv"
	"strings"
)

// ExtractWorkflowDetails derives canonical sub-state, retry count, and error message from execution status.
func ExtractWorkflowDetails(status *ExecutionStatus) (WorkflowSubState, *int, *string) {
	if status == nil {
		return SubStateRunning, nil, nil
	}

	subState := MapExecutionStateToSubState(status.State)
	
	// Check if metadata contains a pre-computed sub-state (from provider-specific mapping)
	if len(status.Metadata) > 0 {
		if providerSubState, ok := status.Metadata["workflow_sub_state"]; ok {
			// Validate and use the provider-specific sub-state
			switch WorkflowSubState(providerSubState) {
			case SubStateRunning, SubStateWaiting, SubStateBackingOff, SubStateError, SubStateSucceeded, SubStateFailed:
				subState = WorkflowSubState(providerSubState)
			}
		}
	}

	if hasBackoffMetadata(status.Metadata) {
		subState = SubStateBackingOff
	}

	retryCount := retryCountFromMetadata(status.Metadata)
	if retryCount == nil {
		if count := retryCountFromHistory(status.History); count > 0 {
			retryCount = &count
		}
	}

	var errMsg *string
	if status.Error != nil && status.Error.Message != "" {
		errMsg = &status.Error.Message
		if (status.State == StateRunning || status.State == StatePending) && subState != SubStateBackingOff {
			subState = SubStateError
		}
	}

	return subState, retryCount, errMsg
}

func hasBackoffMetadata(metadata map[string]string) bool {
	if len(metadata) == 0 {
		return false
	}
	for key, value := range metadata {
		k := strings.ToLower(key)
		v := strings.ToLower(value)
		if strings.Contains(k, "backoff") || strings.Contains(k, "retry_state") {
			if strings.Contains(v, "backoff") || strings.Contains(v, "backing") || v == "true" {
				return true
			}
		}
	}
	return false
}

func retryCountFromMetadata(metadata map[string]string) *int {
	if len(metadata) == 0 {
		return nil
	}
	for key, value := range metadata {
		k := strings.ToLower(key)
		switch k {
		case "retry_count", "retrycount", "retry_attempts", "retryattempts", "attempts":
			if value == "" {
				return nil
			}
			if count, err := strconv.Atoi(value); err == nil {
				return &count
			}
		}
	}
	return nil
}

func retryCountFromHistory(history []ExecutionEvent) int {
	if len(history) == 0 {
		return 0
	}
	count := 0
	for _, event := range history {
		typeLower := strings.ToLower(event.Type)
		if strings.Contains(typeLower, "retry") {
			count++
		}
	}
	return count
}
