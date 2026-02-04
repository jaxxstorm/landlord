package controller

import (
	"github.com/jaxxstorm/landlord/internal/tenant"
	"github.com/jaxxstorm/landlord/internal/workflow"
)

// isDegradedWorkflow checks if a workflow is in a degraded state that warrants restart.
// Degraded states are:
// - SubStateBackingOff: workflow is backing off due to failures
//
// These states indicate provisioning issues that may be resolved by restarting
// the workflow with updated configuration.
//
// Workflows in other states are NOT considered degraded:
// - SubStateRunning: actively provisioning, should not interrupt
// - SubStateSucceeded: completed successfully
// - SubStateFailed: terminal failure, handled separately
// - SubStateWaiting: waiting for external event, not an error state
func isDegradedWorkflow(execStatus *workflow.ExecutionStatus) bool {
	if execStatus == nil {
		return false
	}

	// Only consider running workflows (not terminal states)
	if execStatus.State == workflow.StateSucceeded ||
		execStatus.State == workflow.StateFailed ||
		execStatus.State == workflow.StateTimedOut ||
		execStatus.State == workflow.StateCancelled {
		return false
	}

	// Extract sub-state using workflow package's logic
	subState, _, _ := workflow.ExtractWorkflowDetails(execStatus)

	// Degraded state that warrants restart on config change
	return subState == workflow.SubStateBackingOff
}

// hasConfigChanged checks if tenant's current config differs from workflow's stored config.
// Returns true if:
// - Tenant has a non-empty WorkflowConfigHash AND
// - Current config hash differs from stored hash
// Returns false if:
// - Tenant has no WorkflowConfigHash (old workflow, no hash stored)
// - Current config hash matches stored hash
// - Config hash computation fails
func hasConfigChanged(t *tenant.Tenant) bool {
	if t.WorkflowConfigHash == nil || *t.WorkflowConfigHash == "" {
		// No stored hash - can't detect changes (backward compatibility)
		return false
	}

	currentHash, err := tenant.ComputeConfigHash(t.DesiredConfig)
	if err != nil {
		// Can't compute hash, assume no change to avoid false restarts
		return false
	}

	// Compare stored hash with current hash
	return currentHash != *t.WorkflowConfigHash
}
