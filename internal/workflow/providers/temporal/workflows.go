package temporal

const (
	tenantProvisioningWorkflowID = "tenant-provisioning"
	tenantProvisioningActivityID = "tenant-provisioning-activity"
)

func defaultWorkflowIDs() []string {
	return []string{tenantProvisioningWorkflowID}
}
