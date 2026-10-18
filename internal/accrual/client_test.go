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
	"time"

	"go.uber.org/zap"
)

func Test_accrualClient_GetStatus(t *testing.T) {
	tests := []struct {
		name           string
		resp           *http.Response
		respErr        error
		orderNumber    string
		want           Order
		wantRetryAfter time.Duration
		wantErr        error
	}{
		{
			name:        "Successful request → ok",
			resp:        newOrderResp(accrualStub{order: testSuccessOrderResponse, httpStatus: http.StatusOK}),
			respErr:     nil,
			orderNumber: testValidOrderNumber,
			want:        testSuccessOrderResponse,
			wantErr:     nil,
		},
		{
			name:        "Order is not registered → ok",
			resp:        newResp(http.StatusNoContent, ""),
			respErr:     nil,
			orderNumber: testValidOrderNumber,
			want:        Order{},
			wantErr:     ErrOrderNotRegistered,
		},
		{
			name:        "Empty body → error",
			resp:        newResp(http.StatusOK, ""),
			respErr:     nil,
			orderNumber: testValidOrderNumber,
			want:        Order{},
			wantErr:     io.EOF,
		},

		// Transport errors from http.Client.Do.
		{
			name:        "Closed network connection → error",
			resp:        nil,
			respErr:     net.ErrClosed,
			orderNumber: testValidOrderNumber,
			want:        Order{},
			wantErr:     net.ErrClosed,
		},
		{
			name:        "Server has been closed → error",
			resp:        nil,
			respErr:     http.ErrServerClosed,
			orderNumber: testValidOrderNumber,
			want:        Order{},
			wantErr:     http.ErrServerClosed,
		},
		// Errors mapped from HTTP response status codes.
		{
			name:           "Too many requests → error",
			resp:           newResp(http.StatusTooManyRequests, ""),
			respErr:        nil,
			want:           Order{},
			wantRetryAfter: testFallbackRetryAfter * time.Second,
			wantErr:        ErrTooManyRequests,
		},
		{
			name:    "Get status client failed → error",
			resp:    newResp(http.StatusBadRequest, ""),
			respErr: nil,
			want:    Order{},
			wantErr: ErrAccrualClient,
		},
		{
			name:    "Get status server failed → error",
			resp:    newResp(http.StatusInternalServerError, ""),
			respErr: nil,
			want:    Order{},
			wantErr: ErrAccrualServer,
		},
		{
			name:    "Get status unexpected server response → error",
			resp:    newResp(http.StatusUseProxy, ""),
			respErr: nil,
			want:    Order{},
			wantErr: ErrAccrualUnexpectedResponse,
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
				return tt.resp, tt.respErr
			})

			retryAfterErr := &RetryAfterError{}

			gotOrder, err := ac.GetStatus(t.Context(), tt.orderNumber)

			compareErrors(t, "GetStatus()", err, tt.wantErr)

			if tt.wantRetryAfter != 0 {
				ok := errors.As(err, &retryAfterErr)
				if ok && retryAfterErr.After != tt.wantRetryAfter {
					t.Errorf("got = %v, want = %v", retryAfterErr.After, tt.wantRetryAfter)
				}
				if !ok {
					t.Errorf("got = %v, want = *RetryAfterError", err)
				}
			}
			if !reflect.DeepEqual(gotOrder, tt.want) {
				t.Errorf("accrualClient.GetStatus() = %v, want %v", gotOrder, tt.want)
			}
		})
	}
}

func Test_accrualClient_GetStatus_canceledContext(t *testing.T) {
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
		want     time.Duration
		wantErr  error
	}{
		{
			name:     "Positive Retry-After header → ok",
			response: testResponseWithPositiveRetryHeader,
			want:     testPositiveRetryAfter * time.Second,
			wantErr:  nil,
		},
		{
			name:     "Zero Retry-After header → ok",
			response: testResponseWithZeroRetryHeader,
			want:     testZeroRetryAfter * time.Second,
			wantErr:  nil,
		},
		{
			name:     "Negative Retry-After header → ok",
			response: testResponseWithNegativeRetryHeader,
			want:     testFallbackRetryAfter * time.Second,
			wantErr:  nil,
		},
		{
			name:     "Double max Retry-After header → ok",
			response: testResponseWithDoubleMaxRetryHeader,
			want:     maxRetryWaitSeconds * time.Second,
			wantErr:  nil,
		},
		{
			name:     "Empty Retry-After header → error",
			response: testResponseWithEmptyRetryHeader,
			want:     testFallbackRetryAfter * time.Second,
			wantErr:  strconv.ErrSyntax,
		},
		{
			name:     "Invalid Retry-After header → error",
			response: testResponseWithInvalidRetryHeader,
			want:     testFallbackRetryAfter * time.Second,
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
