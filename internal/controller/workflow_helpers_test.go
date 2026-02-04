package controller

import (
	"testing"

	"github.com/jaxxstorm/landlord/internal/tenant"
	"github.com/jaxxstorm/landlord/internal/workflow"
)

func TestIsDegradedWorkflow_BackingOff(t *testing.T) {
	status := &workflow.ExecutionStatus{
		ExecutionID: "test-exec",
		State:       workflow.StateRunning,
		Metadata: map[string]string{
			"workflow_sub_state": string(workflow.SubStateBackingOff),
		},
	}

	if !isDegradedWorkflow(status) {
		t.Error("Expected backing-off workflow to be degraded")
	}
}

func TestIsDegradedWorkflow_Retrying(t *testing.T) {
	// Note: SubStateRetrying doesn't exist, but backing-off covers retry scenarios
	t.Skip("SubStateRetrying not defined in workflow package")
}

func TestIsDegradedWorkflow_Running_NotDegraded(t *testing.T) {
	status := &workflow.ExecutionStatus{
		ExecutionID: "test-exec",
		State:       workflow.StateRunning,
		Metadata: map[string]string{
			"workflow_sub_state": string(workflow.SubStateRunning),
		},
	}

	if isDegradedWorkflow(status) {
		t.Error("Expected running workflow to NOT be degraded")
	}
}

func TestIsDegradedWorkflow_Succeeded_NotDegraded(t *testing.T) {
	status := &workflow.ExecutionStatus{
		ExecutionID: "test-exec",
		State:       workflow.StateSucceeded,
		Metadata: map[string]string{
			"workflow_sub_state": string(workflow.SubStateSucceeded),
		},
	}

	if isDegradedWorkflow(status) {
		t.Error("Expected succeeded workflow to NOT be degraded")
	}
}

func TestIsDegradedWorkflow_Failed_NotDegraded(t *testing.T) {
	status := &workflow.ExecutionStatus{
		ExecutionID: "test-exec",
		State:       workflow.StateFailed,
		Metadata: map[string]string{
			"workflow_sub_state": string(workflow.SubStateFailed),
		},
	}

	if isDegradedWorkflow(status) {
		t.Error("Expected failed workflow to NOT be degraded (terminal state)")
	}
}

func TestIsDegradedWorkflow_NilStatus(t *testing.T) {
	if isDegradedWorkflow(nil) {
		t.Error("Expected nil status to NOT be degraded")
	}
}

func TestIsDegradedWorkflow_NoMetadata(t *testing.T) {
	status := &workflow.ExecutionStatus{
		ExecutionID: "test-exec",
		State:       workflow.StateRunning,
		Metadata:    nil,
	}

	if isDegradedWorkflow(status) {
		t.Error("Expected workflow without metadata to NOT be degraded")
	}
}

func TestIsDegradedWorkflow_NoSubState(t *testing.T) {
	status := &workflow.ExecutionStatus{
		ExecutionID: "test-exec",
		State:       workflow.StateRunning,
		Metadata: map[string]string{
			"other_field": "value",
		},
	}

	if isDegradedWorkflow(status) {
		t.Error("Expected workflow without sub_state to NOT be degraded")
	}
}

func TestHasConfigChanged_HashChanged(t *testing.T) {
	config1 := map[string]interface{}{"image": "nginx:1.25"}
	config2 := map[string]interface{}{"image": "nginx:1.26"}

	hash1, _ := tenant.ComputeConfigHash(config1)
	
	tn := &tenant.Tenant{
		Name:               "test",
		DesiredConfig:      config2,
		WorkflowConfigHash: &hash1,
	}

	if !hasConfigChanged(tn) {
		t.Error("Expected config change to be detected when hashes differ")
	}
}

func TestHasConfigChanged_HashUnchanged(t *testing.T) {
	config := map[string]interface{}{"image": "nginx:1.25"}
	hash, _ := tenant.ComputeConfigHash(config)
	
	tn := &tenant.Tenant{
		Name:               "test",
		DesiredConfig:      config,
		WorkflowConfigHash: &hash,
	}

	if hasConfigChanged(tn) {
		t.Error("Expected no config change when hashes match")
	}
}

func TestHasConfigChanged_NoStoredHash(t *testing.T) {
	tn := &tenant.Tenant{
		Name:               "test",
		DesiredConfig:      map[string]interface{}{"image": "nginx:1.25"},
		WorkflowConfigHash: nil,
	}

	if hasConfigChanged(tn) {
		t.Error("Expected no config change when no hash stored (backward compatibility)")
	}
}

func TestHasConfigChanged_EmptyStoredHash(t *testing.T) {
	emptyHash := ""
	tn := &tenant.Tenant{
		Name:               "test",
		DesiredConfig:      map[string]interface{}{"image": "nginx:1.25"},
		WorkflowConfigHash: &emptyHash,
	}

	if hasConfigChanged(tn) {
		t.Error("Expected no config change when stored hash is empty")
	}
}
