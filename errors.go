package stridge

import (
	"errors"
	"fmt"
)

// ErrNotImplemented marks endpoint and request behavior left as learning work.
var ErrNotImplemented = errors.New("not implemented")

// APIError represents an error response returned by the Stridge API.
// Its fields may be expanded when the documented error response is implemented.
type APIError struct {
	StatusCode int
	Code       string
	Message    string
}

// Error implements the error interface.
func (e *APIError) Error() string {
	if e == nil {
		return "<nil>"
	}

	return fmt.Sprintf("stridge API error: status=%d code=%q message=%q", e.StatusCode, e.Code, e.Message)
}
