package stepfunctions

import (
    "testing"

    "github.com/aws/aws-sdk-go-v2/service/sfn/types"
    "github.com/jaxxstorm/landlord/internal/workflow"
    "github.com/stretchr/testify/assert"
)

func TestMapExecutionStateRunning(t *testing.T) {
    state := mapExecutionState(types.ExecutionStatusRunning)
    assert.Equal(t, workflow.StateRunning, state)
}

func TestMapExecutionStateSucceeded(t *testing.T) {
    state := mapExecutionState(types.ExecutionStatusSucceeded)
    assert.Equal(t, workflow.StateSucceeded, state)
}

func TestMapExecutionStateFailed(t *testing.T) {
    state := mapExecutionState(types.ExecutionStatusFailed)
    assert.Equal(t, workflow.StateFailed, state)
}

func TestMapExecutionStateTimedOut(t *testing.T) {
    state := mapExecutionState(types.ExecutionStatusTimedOut)
    assert.Equal(t, workflow.StateFailed, state)
}

func TestMapExecutionStateAborted(t *testing.T) {
    state := mapExecutionState(types.ExecutionStatusAborted)
    assert.Equal(t, workflow.StateFailed, state)
}

func TestMapExecutionStateDefault(t *testing.T) {
    // Test an unknown state type
    state := mapExecutionState(types.ExecutionStatus("UNKNOWN"))
    assert.Equal(t, workflow.StatePending, state)
}
