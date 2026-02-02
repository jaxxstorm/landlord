package workflow

import "errors"

var (
	ErrProviderNotFound  = errors.New("workflow provider not found")
	ErrProviderConflict  = errors.New("workflow provider already registered")
	ErrInvalidSpec       = errors.New("invalid workflow specification")
	ErrWorkflowNotFound  = errors.New("workflow not found")
	ErrExecutionNotFound = errors.New("execution not found")
	ErrExecutionFailed   = errors.New("workflow execution failed")
)
