package restate

import (
	"errors"
	"strings"

	"github.com/jaxxstorm/landlord/internal/workflow"
)

// isNotFoundError checks if an error indicates a resource was not found
func isNotFoundError(err error) bool {
	if err == nil {
		return false
	}

	// Check for common not found error indicators
	errMsg := err.Error()
	return strings.Contains(errMsg, "not found") ||
		strings.Contains(errMsg, "404") ||
		errors.Is(err, workflow.ErrWorkflowNotFound)
}

var ErrServiceAlreadyExists = errors.New("restate service already exists")

// isAlreadyExistsError checks if an error indicates a resource already exists
func isAlreadyExistsError(err error) bool {
	if err == nil {
		return false
	}

	if errors.Is(err, ErrServiceAlreadyExists) {
		return true
	}

	errMsg := err.Error()
	return strings.Contains(errMsg, "already exists") ||
		strings.Contains(errMsg, "409")
}

// wrapError wraps a Restate error with context
func wrapError(err error, context string) error {
	if err == nil {
		return nil
	}

	// Map common Restate errors to Provider interface errors
	if isNotFoundError(err) {
		return errors.Join(workflow.ErrWorkflowNotFound, err)
	}

	if strings.Contains(err.Error(), "invalid") || strings.Contains(err.Error(), "Invalid") {
		return errors.Join(workflow.ErrInvalidSpec, err)
	}

	// Return wrapped error with context
	return errors.New(context + ": " + err.Error())
}
