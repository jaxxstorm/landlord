package workflow

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExtractWorkflowDetails_Defaults(t *testing.T) {
	status := &ExecutionStatus{State: StateRunning}
	subState, retryCount, errMsg := ExtractWorkflowDetails(status)
	assert.Equal(t, SubStateRunning, subState)
	assert.Nil(t, retryCount)
	assert.Nil(t, errMsg)
}

func TestExtractWorkflowDetails_BackoffOverridesError(t *testing.T) {
	status := &ExecutionStatus{
		State: StateRunning,
		Error: &ExecutionError{Message: "transient"},
		Metadata: map[string]string{
			"retry_state": "backoff",
			"retry_count": "3",
		},
	}
	subState, retryCount, errMsg := ExtractWorkflowDetails(status)
	assert.Equal(t, SubStateBackingOff, subState)
	if assert.NotNil(t, retryCount) {
		assert.Equal(t, 3, *retryCount)
	}
	if assert.NotNil(t, errMsg) {
		assert.Equal(t, "transient", *errMsg)
	}
}

func TestExtractWorkflowDetails_ErrorSubState(t *testing.T) {
	status := &ExecutionStatus{
		State: StateRunning,
		Error: &ExecutionError{Message: "boom"},
	}
	subState, _, _ := ExtractWorkflowDetails(status)
	assert.Equal(t, SubStateError, subState)
}

func TestExtractWorkflowDetails_PendingMapsToWaiting(t *testing.T) {
	status := &ExecutionStatus{State: StatePending}
	subState, _, _ := ExtractWorkflowDetails(status)
	assert.Equal(t, SubStateWaiting, subState)
}
