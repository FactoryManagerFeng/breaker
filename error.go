package breaker

import "errors"

var (
	ErrTooManyRequests = errors.New("too many requests")
	ErrOpenState       = errors.New("breaker is open")
)
