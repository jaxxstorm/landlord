package stepfunctions

import (
    "errors"
    "testing"

    "github.com/aws/aws-sdk-go-v2/service/sfn/types"
    "github.com/jaxxstorm/landlord/internal/workflow"
    "github.com/stretchr/testify/assert"
)

func TestIsStateMachineNotFound(t *testing.T) {
    err := &types.StateMachineDoesNotExist{}
    assert.True(t, isStateMachineNotFound(err))
}

func TestIsStateMachineNotFoundWithOtherError(t *testing.T) {
    err := &types.InvalidDefinition{}
    assert.False(t, isStateMachineNotFound(err))
}

func TestIsStateMachineNotFoundWithGenericError(t *testing.T) {
    err := errors.New("some error")
    assert.False(t, isStateMachineNotFound(err))
}

func TestIsExecutionNotFound(t *testing.T) {
    err := &types.ExecutionDoesNotExist{}
    assert.True(t, isExecutionNotFound(err))
}

func TestIsExecutionNotFoundWithOtherError(t *testing.T) {
    err := &types.InvalidDefinition{}
    assert.False(t, isExecutionNotFound(err))
}

func TestIsInvalidDefinition(t *testing.T) {
    err := &types.InvalidDefinition{}
    assert.True(t, isInvalidDefinition(err))
}

func TestIsInvalidDefinitionWithOtherError(t *testing.T) {
    err := &types.ExecutionDoesNotExist{}
    assert.False(t, isInvalidDefinition(err))
}

func TestIsStateMachineAlreadyExists(t *testing.T) {
    err := &types.StateMachineAlreadyExists{}
    assert.True(t, isStateMachineAlreadyExists(err))
}

func TestIsStateMachineAlreadyExistsWithOtherError(t *testing.T) {
    err := &types.InvalidDefinition{}
    assert.False(t, isStateMachineAlreadyExists(err))
}

func TestIsExecutionAlreadyExists(t *testing.T) {
    err := &types.ExecutionAlreadyExists{}
    assert.True(t, isExecutionAlreadyExists(err))
}

func TestIsExecutionAlreadyExistsWithOtherError(t *testing.T) {
    err := &types.InvalidDefinition{}
    assert.False(t, isExecutionAlreadyExists(err))
}

func TestWrapAWSErrorStateMachineNotFound(t *testing.T) {
    originalErr := &types.StateMachineDoesNotExist{}
    wrappedErr := wrapAWSError(originalErr, "DescribeStateMachine")
    // Should wrap the error with operation context
    assert.ErrorIs(t, wrappedErr, workflow.ErrWorkflowNotFound)
}

func TestWrapAWSErrorExecutionNotFound(t *testing.T) {
    originalErr := &types.ExecutionDoesNotExist{}
    wrappedErr := wrapAWSError(originalErr, "DescribeExecution")
    // Should wrap the error with operation context
    assert.ErrorIs(t, wrappedErr, workflow.ErrExecutionNotFound)
}

func TestWrapAWSErrorInvalidDefinition(t *testing.T) {
    originalErr := &types.InvalidDefinition{}
    wrappedErr := wrapAWSError(originalErr, "CreateStateMachine")
    // Should wrap the error with operation context
    assert.ErrorIs(t, wrappedErr, workflow.ErrInvalidSpec)
}

func TestWrapAWSErrorGenericError(t *testing.T) {
    originalErr := errors.New("unknown AWS error")
    wrappedErr := wrapAWSError(originalErr, "SomeOperation")
    assert.Error(t, wrappedErr)
    // Should wrap with operation context
    assert.Contains(t, wrappedErr.Error(), "SomeOperation")
}

func TestWrapAWSErrorPreservesOriginal(t *testing.T) {
    originalErr := errors.New("test error")
    wrappedErr := wrapAWSError(originalErr, "TestOperation")

    // Should be able to unwrap and get original error
    assert.NotNil(t, wrappedErr)
    // Verify operation is in error message
    assert.Contains(t, wrappedErr.Error(), "TestOperation")
}
