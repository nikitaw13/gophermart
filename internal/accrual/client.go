package accrual

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"go.uber.org/zap"
)

const maxRetryWaitSeconds = 60

// accrualClient talks to the loyalty accrual system over HTTP.
type accrualClient struct {
	address string
	client  *http.Client
	logger  *zap.Logger
}

// NewAccrualClient creates an HTTP client for the loyalty accrual system at the given address.
func NewAccrualClient(address string, logger *zap.Logger) *accrualClient {
	return &accrualClient{
		address: address,
		client: &http.Client{
			Timeout: 5 * time.Second,
		},
		logger: logger,
	}
}

// GetStatus fetches the current accrual status of the order with the given number.
func (ac *accrualClient) GetStatus(ctx context.Context, orderNumber string) (Order, error) {
	var result Order
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("%s/api/orders/%s", ac.address, orderNumber), nil)

	if err != nil {
		return result, err
	}

	response, err := ac.client.Do(request)
	if err != nil {
		return result, err
	}
	defer response.Body.Close()

	if response.StatusCode == http.StatusOK {
		err = json.NewDecoder(response.Body).Decode(&result)
		if err != nil {
			return result, err
		}
		return result, nil
	}
	if response.StatusCode == http.StatusNoContent {
		return result, ErrOrderNotRegistered
	}
	if response.StatusCode == http.StatusTooManyRequests {
		waitTime, err := parseRetryAfterHeader(response)
		if err != nil {
			ac.logger.Debug("invalid Retry-After header, using default delay",
				zap.Error(err), zap.String("orderNumber", orderNumber))
		}
		return result, &RetryAfterError{
			After: waitTime,
			Err:   ErrTooManyRequests,
		}
	}
	if response.StatusCode >= http.StatusBadRequest && response.StatusCode < http.StatusInternalServerError {
		return result, ErrAccrualClient
	}
	if response.StatusCode >= http.StatusInternalServerError {
		return result, ErrAccrualServer
	}

	return result, ErrAccrualUnexpectedResponse
}

func parseRetryAfterHeader(response *http.Response) (time.Duration, error) {
	seconds, err := strconv.Atoi(response.Header.Get("Retry-After"))
	if err != nil {
		seconds = 1
	}
	if seconds < 0 {
		seconds = 1
	}
	return time.Duration(min(maxRetryWaitSeconds, seconds)) * time.Second, err
}
