package stepfunctions

import (
	sfnTypes "github.com/aws/aws-sdk-go-v2/service/sfn/types"

	"github.com/jaxxstorm/landlord/internal/workflow"
)

// mapExecutionState maps AWS Step Functions execution status to workflow.ExecutionState
func mapExecutionState(status sfnTypes.ExecutionStatus) workflow.ExecutionState {
	switch status {
	case sfnTypes.ExecutionStatusRunning:
		return workflow.StateRunning
	case sfnTypes.ExecutionStatusSucceeded:
		return workflow.StateSucceeded
	case sfnTypes.ExecutionStatusFailed, sfnTypes.ExecutionStatusTimedOut, sfnTypes.ExecutionStatusAborted:
		return workflow.StateFailed
	default:
		return workflow.StatePending
	}
}
