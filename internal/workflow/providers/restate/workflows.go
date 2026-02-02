package restate

import (
	"fmt"

	"github.com/jaxxstorm/landlord/internal/config"
)

const tenantProvisioningWorkflowID = "tenant-provisioning"

func workflowServiceName(cfg config.RestateConfig, workflowID string) string {
	if cfg.ServiceName != "" {
		return cfg.ServiceName
	}

	return normalizeServiceName(workflowID)
}

func defaultWorkflowIDs() []string {
	return []string{tenantProvisioningWorkflowID}
}

// WorkflowServiceName returns the default Restate service name for tenant provisioning.
func WorkflowServiceName(cfg config.RestateConfig) string {
	return workflowServiceName(cfg, tenantProvisioningWorkflowID)
}

func WorkerServiceName(cfg config.RestateConfig) string {
	name := WorkflowServiceName(cfg)
	if cfg.WorkerServicePrefix != "" {
		name = fmt.Sprintf("%s-%s", cfg.WorkerServicePrefix, name)
	}
	if cfg.WorkerNamespace != "" {
		name = fmt.Sprintf("%s-%s", cfg.WorkerNamespace, name)
	}
	return name
}
