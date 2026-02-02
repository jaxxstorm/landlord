package tenant

import "fmt"

// NextStatus determines the next status in the tenant lifecycle based on current status
func NextStatus(current Status) (Status, error) {
	switch current {
	case StatusRequested:
		return StatusProvisioning, nil
	case StatusPlanning:
		return StatusProvisioning, nil
	case StatusProvisioning:
		return StatusReady, nil
	case StatusUpdating:
		return StatusReady, nil
	case StatusDeleting:
		return StatusArchived, nil
	case StatusArchiving:
		return StatusArchived, nil
	case StatusArchived, StatusReady, StatusFailed:
		// Already in terminal state
		return "", fmt.Errorf("%s is a terminal state", current)
	default:
		return "", fmt.Errorf("unknown status: %s", current)
	}
}

// ShouldReconcile determines if a tenant status requires reconciliation action
func ShouldReconcile(status Status) bool {
	switch status {
	case StatusRequested,
		StatusPlanning,
		StatusProvisioning,
		StatusUpdating,
		StatusDeleting,
		StatusArchiving:
		return true
	case StatusReady,
		StatusArchived,
		StatusFailed:
		return false
	default:
		return false
	}
}

// IsTerminalStatus checks if a status is terminal (no further transitions)
func IsTerminalStatus(status Status) bool {
	return status == StatusReady || status == StatusArchived || status == StatusFailed
}

// ValidateTransition checks if a status transition is valid
func ValidateTransition(from, to Status) error {
	validTransitions := map[Status][]Status{
		StatusRequested:    {StatusProvisioning, StatusFailed},
		StatusPlanning:     {StatusProvisioning, StatusFailed},
		StatusProvisioning: {StatusReady, StatusFailed},
		StatusReady:        {StatusUpdating, StatusDeleting, StatusArchiving},
		StatusUpdating:     {StatusReady, StatusFailed},
		StatusDeleting:     {StatusArchived, StatusFailed},
		StatusArchiving:    {StatusArchived, StatusFailed},
		StatusArchived:     {},                                // Terminal, no transitions
		StatusFailed:       {StatusDeleting, StatusArchiving}, // Allow archive/delete after failure
	}

	allowed, ok := validTransitions[from]
	if !ok {
		return fmt.Errorf("unknown source status: %s", from)
	}

	for _, valid := range allowed {
		if to == valid {
			return nil
		}
	}

	return fmt.Errorf("invalid transition from %s to %s", from, to)
}
