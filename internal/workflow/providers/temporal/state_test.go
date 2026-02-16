package temporal

import (
	"testing"
	"time"

	"github.com/jaxxstorm/landlord/internal/workflow"
	"github.com/stretchr/testify/assert"
	enumspb "go.temporal.io/api/enums/v1"
	workflowservice "go.temporal.io/api/workflowservice/v1"
	workflowpb "go.temporal.io/api/workflow/v1"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestMapExecutionState(t *testing.T) {
	tests := []struct {
		name     string
		input    enumspb.WorkflowExecutionStatus
		expected workflow.ExecutionState
	}{
		{name: "running", input: enumspb.WORKFLOW_EXECUTION_STATUS_RUNNING, expected: workflow.StateRunning},
		{name: "continued as new", input: enumspb.WORKFLOW_EXECUTION_STATUS_CONTINUED_AS_NEW, expected: workflow.StateRunning},
		{name: "completed", input: enumspb.WORKFLOW_EXECUTION_STATUS_COMPLETED, expected: workflow.StateSucceeded},
		{name: "failed", input: enumspb.WORKFLOW_EXECUTION_STATUS_FAILED, expected: workflow.StateFailed},
		{name: "timed out", input: enumspb.WORKFLOW_EXECUTION_STATUS_TIMED_OUT, expected: workflow.StateTimedOut},
		{name: "canceled", input: enumspb.WORKFLOW_EXECUTION_STATUS_CANCELED, expected: workflow.StateCancelled},
		{name: "terminated", input: enumspb.WORKFLOW_EXECUTION_STATUS_TERMINATED, expected: workflow.StateFailed},
		{name: "unknown defaults to running", input: enumspb.WorkflowExecutionStatus(9999), expected: workflow.StateRunning},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, mapExecutionState(tt.input))
		})
	}
}

func TestMapSubStateUnknownDefaultsToRunning(t *testing.T) {
	assert.Equal(t, workflow.SubStateRunning, mapSubState(enumspb.WorkflowExecutionStatus(9999)))
}

func TestBuildStatusMetadataBackoff(t *testing.T) {
	resp := &workflowservice.DescribeWorkflowExecutionResponse{
		WorkflowExecutionInfo: &workflowpb.WorkflowExecutionInfo{
			Status: enumspb.WORKFLOW_EXECUTION_STATUS_RUNNING,
		},
		PendingActivities: []*workflowpb.PendingActivityInfo{
			{
				Attempt:                 3,
				CurrentRetryInterval:    durationpb.New(5 * time.Second),
				NextAttemptScheduleTime: timestamppb.Now(),
			},
		},
	}
	meta := buildStatusMetadata(resp, nil)
	assert.Equal(t, "3", meta["retry_count"])
	assert.Equal(t, "backing-off", meta["retry_state"])
	assert.Equal(t, string(workflow.SubStateBackingOff), meta["workflow_sub_state"])
}
