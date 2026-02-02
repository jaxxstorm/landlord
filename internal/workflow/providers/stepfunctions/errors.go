package stepfunctions

import (
	"errors"
	"fmt"

	sfnTypes "github.com/aws/aws-sdk-go-v2/service/sfn/types"

	"github.com/jaxxstorm/landlord/internal/workflow"
)

// ProviderError represents errors returned by the provider
type ProviderError struct {
	Op  string
	Err error
}

func (e *ProviderError) Error() string {
	return fmt.Sprintf("stepfunctions: %s: %v", e.Op, e.Err)
}

func (e *ProviderError) Unwrap() error { return e.Err }

// helper constructors
func wrapErr(op string, err error) error {
	if err == nil {
		return nil
	}
	return &ProviderError{Op: op, Err: err}
}

// isStateMachineNotFound returns true if error indicates state machine missing
func isStateMachineNotFound(err error) bool {
	var target *sfnTypes.StateMachineDoesNotExist
	return errors.As(err, &target)
}

func isExecutionNotFound(err error) bool {
	var target *sfnTypes.ExecutionDoesNotExist
	return errors.As(err, &target)
}

func isInvalidDefinition(err error) bool {
	var target *sfnTypes.InvalidDefinition
	return errors.As(err, &target)
}

func isStateMachineAlreadyExists(err error) bool {
	var target *sfnTypes.StateMachineAlreadyExists
	return errors.As(err, &target)
}

func isExecutionAlreadyExists(err error) bool {
	var target *sfnTypes.ExecutionAlreadyExists
	return errors.As(err, &target)
}

// wrapAWSError maps AWS SDK errors to workflow package sentinel errors
func wrapAWSError(err error, operation string) error {
	if isStateMachineNotFound(err) {
		return fmt.Errorf("%s: %w", operation, workflow.ErrWorkflowNotFound)
	}
	if isExecutionNotFound(err) {
		return fmt.Errorf("%s: %w", operation, workflow.ErrExecutionNotFound)
	}
	if isInvalidDefinition(err) {
		return fmt.Errorf("%s: %w", operation, workflow.ErrInvalidSpec)
	}
	return fmt.Errorf("%s: %w", operation, err)
}
