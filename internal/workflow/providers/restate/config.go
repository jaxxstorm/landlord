package restate

import "fmt"

// normalizeServiceName converts a workflow ID (kebab-case) to a PascalCase service name
// Restate services typically use PascalCase naming convention
func normalizeServiceName(workflowID string) string {
	if workflowID == "" {
		return ""
	}

	// Convert kebab-case to PascalCase
	// Example: "tenant-provisioning" -> "TenantProvisioning"
	var result string
	capitalizeNext := true

	for _, ch := range workflowID {
		if ch == '-' {
			capitalizeNext = true
		} else if capitalizeNext {
			result += string(ch - 32) // Convert to uppercase
			capitalizeNext = false
		} else {
			result += string(ch)
		}
	}

	return result
}

// generateExecutionID generates a unique execution ID
func generateExecutionID(workflowID string) string {
	// In production, this would generate a truly unique ID
	// For now, use a simple format: workflow-id-timestamp-based
	return fmt.Sprintf("%s-%x", workflowID, getRandomBytes())
}

// getRandomBytes returns a simple random value (in production, use crypto/rand)
func getRandomBytes() uint32 {
	// Placeholder - in production use crypto/rand
	return 12345
}
