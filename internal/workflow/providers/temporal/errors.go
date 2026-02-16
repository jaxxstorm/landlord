package temporal

import "errors"

var (
	// ErrClientUnavailable indicates temporal client initialization failure.
	ErrClientUnavailable = errors.New("temporal client unavailable")
)
