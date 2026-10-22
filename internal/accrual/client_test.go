package accrual

import (
	"context"
	"errors"
	"io"
	"net"
	"net/http"
	"reflect"
	"strconv"
	"testing"

	"go.uber.org/zap"
)

func Test_AccrualClient_GetStatus(t *testing.T) {
	tests := []struct {
		name           string
		httpStatus     int
		body           string
		respErr        error
		orderNumber    string
		want           Order
		wantRetryAfter int
		wantErr        error
	}{
		{
			name:        "Successful request → ok",
			httpStatus:  http.StatusOK,
			body:        orderJSONBody(testSuccessOrderResponse),
			respErr:     nil,
			orderNumber: testValidOrderNumber,
			want:        testSuccessOrderResponse,
			wantErr:     nil,
		},
		{
			name:        "Order is not registered → error",
			httpStatus:  http.StatusNoContent,
			respErr:     nil,
			orderNumber: testValidOrderNumber,
			want:        Order{},
			wantErr:     ErrOrderNotRegistered,
		},
		{
			name:        "Empty body → error",
			httpStatus:  http.StatusOK,
			respErr:     nil,
			orderNumber: testValidOrderNumber,
			want:        Order{},
			wantErr:     io.EOF,
		},

		// Transport errors from http.Client.Do.
		{
			name:        "Closed network connection → error",
			respErr:     net.ErrClosed,
			orderNumber: testValidOrderNumber,
			want:        Order{},
			wantErr:     net.ErrClosed,
		},
		{
			name:        "Server has been closed → error",
			respErr:     http.ErrServerClosed,
			orderNumber: testValidOrderNumber,
			want:        Order{},
			wantErr:     http.ErrServerClosed,
		},
		// Errors mapped from HTTP response status codes.
		{
			name:           "Too many requests → error",
			httpStatus:     http.StatusTooManyRequests,
			respErr:        nil,
			want:           Order{},
			wantRetryAfter: testFallbackRetryAfter,
			wantErr:        ErrTooManyRequests,
		},
		{
			name:       "Get status client failed → error",
			httpStatus: http.StatusBadRequest,
			respErr:    nil,
			want:       Order{},
			wantErr:    ErrAccrualClient,
		},
		{
			name:       "Get status server failed → error",
			httpStatus: http.StatusInternalServerError,
			respErr:    nil,
			want:       Order{},
			wantErr:    ErrAccrualServer,
		},
		{
			name:       "Get status unexpected server response → error",
			httpStatus: http.StatusUseProxy,
			respErr:    nil,
			want:       Order{},
			wantErr:    ErrAccrualUnexpectedResponse,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ac := NewAccrualClient(
				testFakeAccrualAddress,
				zap.NewNop(),
			)
			ac.client.Transport = roundTripperFunc(func(r *http.Request) (*http.Response, error) {
				if tt.respErr != nil {
					return nil, tt.respErr
				}
				return newResp(tt.httpStatus, tt.body), nil
			})

			retryAfterErr := &RetryAfterError{}

			gotOrder, err := ac.GetStatus(t.Context(), tt.orderNumber)

			compareErrors(t, "GetStatus()", err, tt.wantErr)

			if tt.wantRetryAfter != 0 {
				ok := errors.As(err, &retryAfterErr)
				if ok && retryAfterErr.After != tt.wantRetryAfter {
					t.Errorf("GetStatus().RetryAfter.After = %v, want %v", retryAfterErr.After, tt.wantRetryAfter)
				}
				if !ok {
					t.Errorf("GetStatus() error = %v, want *RetryAfterError", err)
				}
			}
			if !reflect.DeepEqual(gotOrder, tt.want) {
				t.Errorf("AccrualClient.GetStatus() = %v, want %v", gotOrder, tt.want)
			}
		})
	}
}

func Test_AccrualClient_GetStatus_canceledContext(t *testing.T) {
	tests := []struct {
		name        string
		orderNumber string
		wantErr     error
	}{
		{
			name:        "Canceled context → error",
			orderNumber: testValidOrderNumber,
			wantErr:     context.Canceled,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ac := NewAccrualClient(
				testFakeAccrualAddress,
				zap.NewNop(),
			)
			ac.client.Transport = roundTripperFunc(func(r *http.Request) (*http.Response, error) {
				if r.Context().Err() != nil {
					return nil, r.Context().Err()
				}
				return nil, nil
			})

			ctx, cancel := context.WithCancel(context.Background())
			cancel()
			_, err := ac.GetStatus(ctx, tt.orderNumber)

			compareErrors(t, "GetStatus() with canceled context", err, tt.wantErr)
		})
	}
}

func Test_parseRetryAfterHeader(t *testing.T) {
	tests := []struct {
		name     string
		response *http.Response
		want     int
		wantErr  error
	}{
		{
			name:     "Positive Retry-After header → ok",
			response: testResponseWithPositiveRetryHeader,
			want:     testPositiveRetryAfter,
			wantErr:  nil,
		},
		{
			name:     "Zero Retry-After header → ok",
			response: testResponseWithZeroRetryHeader,
			want:     testZeroRetryAfter,
			wantErr:  nil,
		},
		{
			name:     "Negative Retry-After header → ok",
			response: testResponseWithNegativeRetryHeader,
			want:     testFallbackRetryAfter,
			wantErr:  nil,
		},
		{
			name:     "Double max Retry-After header → ok",
			response: testResponseWithDoubleMaxRetryHeader,
			want:     maxRetryWaitSeconds,
			wantErr:  nil,
		},
		{
			name:     "Empty Retry-After header → error",
			response: testResponseWithEmptyRetryHeader,
			want:     testFallbackRetryAfter,
			wantErr:  strconv.ErrSyntax,
		},
		{
			name:     "Invalid Retry-After header → error",
			response: testResponseWithInvalidRetryHeader,
			want:     testFallbackRetryAfter,
			wantErr:  strconv.ErrSyntax,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := parseRetryAfterHeader(tt.response)

			compareErrors(t, "parseRetryAfterHeader()", err, tt.wantErr)

			if got != tt.want {
				t.Errorf("parseRetryAfterHeader() = %v, want %v", got, tt.want)
			}
		})
	}
}
