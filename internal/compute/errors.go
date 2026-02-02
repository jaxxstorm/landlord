package compute

import "errors"

var (
	// ErrProviderNotFound is returned when a provider is not registered
	ErrProviderNotFound = errors.New("compute provider not found")

	// ErrProviderConflict is returned when attempting to register a provider with a duplicate name
	ErrProviderConflict = errors.New("provider already registered")

	// ErrInvalidSpec is returned when a compute specification is invalid
	ErrInvalidSpec = errors.New("invalid compute specification")

	// ErrProvisionFailed is returned when provisioning fails
	ErrProvisionFailed = errors.New("provisioning failed")

	// ErrUpdateFailed is returned when update fails
	ErrUpdateFailed = errors.New("update failed")

	// ErrTenantNotFound is returned when tenant compute resources don't exist
	ErrTenantNotFound = errors.New("tenant compute resources not found")
)
