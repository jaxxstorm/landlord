package stepfunctions

import (
    "testing"

    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func TestBuildStateMachineARN(t *testing.T) {
    arn := buildStateMachineARN("my-workflow", "us-east-1", "123456789012")
    expectedARN := "arn:aws:states:us-east-1:123456789012:stateMachine:my-workflow"
    assert.Equal(t, expectedARN, arn)
}

func TestBuildStateMachineARNWithDifferentRegion(t *testing.T) {
    arn := buildStateMachineARN("my-workflow", "eu-west-1", "987654321098")
    expectedARN := "arn:aws:states:eu-west-1:987654321098:stateMachine:my-workflow"
    assert.Equal(t, expectedARN, arn)
}

func TestBuildExecutionARN(t *testing.T) {
    arn := buildExecutionARN("my-workflow", "exec-123", "us-east-1", "123456789012")
    expectedARN := "arn:aws:states:us-east-1:123456789012:execution:my-workflow:exec-123"
    assert.Equal(t, expectedARN, arn)
}

func TestBuildExecutionARNWithSpecialChars(t *testing.T) {
    arn := buildExecutionARN("my-workflow", "exec-2024-01-15", "us-west-2", "111111111111")
    expectedARN := "arn:aws:states:us-west-2:111111111111:execution:my-workflow:exec-2024-01-15"
    assert.Equal(t, expectedARN, arn)
}

func TestParseStateMachineARN(t *testing.T) {
    arn := "arn:aws:states:us-east-1:123456789012:stateMachine:my-workflow"
    workflowID, err := parseStateMachineARN(arn)
    require.NoError(t, err)
    assert.Equal(t, "my-workflow", workflowID)
}

func TestParseStateMachineARNInvalid(t *testing.T) {
    arn := "invalid-arn"
    _, err := parseStateMachineARN(arn)
    assert.Error(t, err)
}

func TestParseStateMachineARNWrongFormat(t *testing.T) {
    arn := "arn:aws:states:us-east-1:123456789012:execution:my-workflow"
    _, err := parseStateMachineARN(arn)
    assert.Error(t, err)
}

func TestParseExecutionARN(t *testing.T) {
    arn := "arn:aws:states:us-east-1:123456789012:execution:my-workflow:exec-123"
    workflowID, executionName, err := parseExecutionARN(arn)
    require.NoError(t, err)
    assert.Equal(t, "my-workflow", workflowID)
    assert.Equal(t, "exec-123", executionName)
}

func TestParseExecutionARNInvalid(t *testing.T) {
    arn := "invalid-arn"
    _, _, err := parseExecutionARN(arn)
    assert.Error(t, err)
}

func TestParseExecutionARNWrongFormat(t *testing.T) {
    arn := "arn:aws:states:us-east-1:123456789012:execution:my-workflow"
    _, _, err := parseExecutionARN(arn)
    assert.Error(t, err)
}

func TestParseExecutionARNWithSpecialChars(t *testing.T) {
    arn := "arn:aws:states:us-west-2:987654321098:execution:my-workflow:exec-2024-01-15-12-30"
    workflowID, executionName, err := parseExecutionARN(arn)
    require.NoError(t, err)
    assert.Equal(t, "my-workflow", workflowID)
    assert.Equal(t, "exec-2024-01-15-12-30", executionName)
}
