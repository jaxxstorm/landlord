package temporal

import (
	"fmt"

	"github.com/jaxxstorm/landlord/internal/workflow"
	enumspb "go.temporal.io/api/enums/v1"
	workflowservice "go.temporal.io/api/workflowservice/v1"
)

func mapExecutionState(status enumspb.WorkflowExecutionStatus) workflow.ExecutionState {
	switch status {
	case enumspb.WORKFLOW_EXECUTION_STATUS_COMPLETED:
		return workflow.StateSucceeded
	case enumspb.WORKFLOW_EXECUTION_STATUS_CANCELED:
		return workflow.StateCancelled
	case enumspb.WORKFLOW_EXECUTION_STATUS_FAILED:
		return workflow.StateFailed
	case enumspb.WORKFLOW_EXECUTION_STATUS_TERMINATED:
		return workflow.StateFailed
	case enumspb.WORKFLOW_EXECUTION_STATUS_TIMED_OUT:
		return workflow.StateTimedOut
	case enumspb.WORKFLOW_EXECUTION_STATUS_CONTINUED_AS_NEW:
		return workflow.StateRunning
	case enumspb.WORKFLOW_EXECUTION_STATUS_RUNNING:
		return workflow.StateRunning
	default:
		return workflow.StateRunning
	}
}

func mapSubState(status enumspb.WorkflowExecutionStatus) workflow.WorkflowSubState {
	switch status {
	case enumspb.WORKFLOW_EXECUTION_STATUS_COMPLETED:
		return workflow.SubStateSucceeded
	case enumspb.WORKFLOW_EXECUTION_STATUS_FAILED,
		enumspb.WORKFLOW_EXECUTION_STATUS_TERMINATED,
		enumspb.WORKFLOW_EXECUTION_STATUS_CANCELED,
		enumspb.WORKFLOW_EXECUTION_STATUS_TIMED_OUT:
		return workflow.SubStateFailed
	case enumspb.WORKFLOW_EXECUTION_STATUS_RUNNING,
		enumspb.WORKFLOW_EXECUTION_STATUS_CONTINUED_AS_NEW:
		return workflow.SubStateRunning
	default:
		return workflow.SubStateRunning
	}
}

func buildStatusMetadata(resp *workflowservice.DescribeWorkflowExecutionResponse, rec *executionRecord) map[string]string {
	meta := map[string]string{}

	if rec != nil && rec.input != nil {
		for k, v := range rec.input.Metadata {
			meta[k] = v
		}
	}

	if resp == nil || resp.WorkflowExecutionInfo == nil {
		if len(meta) == 0 {
			return nil
		}
		return meta
	}

	status := resp.WorkflowExecutionInfo.GetStatus()
	meta["temporal_status"] = status.String()
	meta["workflow_sub_state"] = string(mapSubState(status))

	highestAttempt := int32(0)
	backingOff := false
	for _, activity := range resp.GetPendingActivities() {
		if activity.GetAttempt() > highestAttempt {
			highestAttempt = activity.GetAttempt()
		}
		if activity.GetCurrentRetryInterval() != nil || activity.GetNextAttemptScheduleTime() != nil {
			backingOff = true
		}
	}
	if highestAttempt > 0 {
		meta["retry_count"] = fmt.Sprintf("%d", highestAttempt)
	}
	if backingOff {
		meta["retry_state"] = "backing-off"
		meta["workflow_sub_state"] = string(workflow.SubStateBackingOff)
	}

	if len(meta) == 0 {
		return nil
	}
	return meta
}
