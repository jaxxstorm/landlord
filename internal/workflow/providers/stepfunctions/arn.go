package stepfunctions

import (
	"fmt"
	"strings"
)

// buildStateMachineARN constructs a state machine ARN from components
// Format: arn:aws:states:{region}:{accountID}:stateMachine:{workflowID}
func buildStateMachineARN(workflowID, region, accountID string) string {
	return fmt.Sprintf("arn:aws:states:%s:%s:stateMachine:%s",
		region, accountID, workflowID)
}

// buildExecutionARN constructs an execution ARN from components
// Format: arn:aws:states:{region}:{accountID}:execution:{workflowID}:{executionName}
func buildExecutionARN(workflowID, executionName, region, accountID string) string {
	return fmt.Sprintf("arn:aws:states:%s:%s:execution:%s:%s",
		region, accountID, workflowID, executionName)
}

// parseExecutionARN extracts components from an execution ARN
// Returns workflowID, executionName, or error if invalid
func parseExecutionARN(arn string) (workflowID, executionName string, err error) {
	// arn:aws:states:region:account:execution:workflow:name
	parts := strings.Split(arn, ":")
	if len(parts) != 8 {
		return "", "", fmt.Errorf("invalid execution ARN format: %s", arn)
	}
	if parts[0] != "arn" || parts[1] != "aws" || parts[2] != "states" || parts[5] != "execution" {
		return "", "", fmt.Errorf("invalid execution ARN format: %s", arn)
	}
	return parts[6], parts[7], nil
}

// parseStateMachineARN extracts workflow ID from a state machine ARN
// Returns workflowID or error if invalid
func parseStateMachineARN(arn string) (workflowID string, err error) {
	// arn:aws:states:region:account:stateMachine:workflow
	parts := strings.Split(arn, ":")
	if len(parts) != 7 {
		return "", fmt.Errorf("invalid state machine ARN format: %s", arn)
	}
	if parts[0] != "arn" || parts[1] != "aws" || parts[2] != "states" || parts[5] != "stateMachine" {
		return "", fmt.Errorf("invalid state machine ARN format: %s", arn)
	}
	return parts[6], nil
}
