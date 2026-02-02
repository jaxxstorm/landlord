package controller

import (
	"github.com/jaxxstorm/landlord/internal/tenant"
)

// nextStatus wraps tenant.NextStatus for backward compatibility
func nextStatus(current tenant.Status) (tenant.Status, error) {
	return tenant.NextStatus(current)
}

// shouldReconcile wraps tenant.ShouldReconcile for backward compatibility
func shouldReconcile(status tenant.Status) bool {
	return tenant.ShouldReconcile(status)
}

// isTerminalStatus wraps tenant.IsTerminalStatus for backward compatibility
func isTerminalStatus(status tenant.Status) bool {
	return tenant.IsTerminalStatus(status)
}

// validateTransition wraps tenant.ValidateTransition for backward compatibility
func validateTransition(from, to tenant.Status) error {
	return tenant.ValidateTransition(from, to)
}
