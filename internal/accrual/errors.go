package accrual

import (
	"errors"
	"fmt"
	"time"
)

var (
	// ErrOrderNotRegistered is returned when the accrual system does not know the order.
	ErrOrderNotRegistered = errors.New("order is not registered")

	// ErrTooManyRequests is returned when the accrual system responds with HTTP 429.
	ErrTooManyRequests = errors.New("too many requests")

	// ErrAccrualServer is returned when the accrual system responds with HTTP 5xx.
	ErrAccrualServer = errors.New("the accrual server failed to fulfill an apparently valid request")

	// ErrAccrualClient is returned when the accrual system responds with HTTP 4xx.
	ErrAccrualClient = errors.New("the accrual server rejected the request")

	// ErrAccrualUnexpectedResponse is returned when the accrual system responds with an unexpected status code.
	ErrAccrualUnexpectedResponse = errors.New("unexpected response from the accrual server")
)

// RetryAfterError wraps an error and indicates how long the caller should wait before retrying.
type RetryAfterError struct {
	After time.Duration
	Err   error
}

func (e *RetryAfterError) Error() string {
	return fmt.Sprintf("%s (retry after %s)", e.Err, e.After)
}

func (e *RetryAfterError) Unwrap() error {
	return e.Err
}

// RetryClassification tells whether a failed operation can be retried.
type RetryClassification int

const (
	// NonRetriable means the operation must not be retried.
	NonRetriable RetryClassification = iota

	// Retriable means the operation can be retried.
	Retriable
)

// Classify returns the classification of the given error.
func Classify(err error) RetryClassification {
	switch {
	case errors.Is(err, ErrTooManyRequests),
		errors.Is(err, ErrAccrualServer):
		return Retriable
	default:
		return NonRetriable
	}
}
