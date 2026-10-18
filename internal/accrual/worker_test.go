package accrual

import (
	"context"
	"fmt"
	"net/http"
	"reflect"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/nikitaw13/gophermart/internal/mocks"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"
)

func Test_NewAccrualWorker(t *testing.T) {
	tests := []struct {
		name        string
		workerCount int
		want        int
	}{
		{name: "Zero worker count → one worker", workerCount: 0, want: 1},
		{name: "Negative worker count → one worker", workerCount: -1, want: 1},
		{name: "Positive worker count → unchanged", workerCount: 3, want: 3},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ac := NewAccrualClient(testFakeAccrualAddress, zap.NewNop())
			aw := NewAccrualWorker(ac, nil, zap.NewNop(), tt.workerCount, testAccrualPollInterval)
			if aw.workerCount != tt.want {
				t.Errorf("NewAccrualWorker() workerCount = %v, want %v", aw.workerCount, tt.want)
			}
		})
	}
}

func Test_accrualWorker_ProcessNewOrders(t *testing.T) {
	tests := []struct {
		name             string
		setup            func(repo *mocks.MockAccrualRepository, cancel context.CancelFunc)
		accrualResponses map[string]accrualStub
		wantErr          error
	}{
		{
			name: "Fetch new orders failed → keep polling",
			setup: func(repo *mocks.MockAccrualRepository, cancel context.CancelFunc) {
				fetchCount := 0
				repo.EXPECT().FetchNewOrders(gomock.Any()).Times(3).DoAndReturn(func(ctx context.Context) ([]string, error) {
					if fetchCount == 2 {
						cancel()
						return []string{}, ctx.Err()
					}
					fetchCount++
					return []string{}, testErrFetchNewOrders
				})
			},
			wantErr: nil,
		},
		{
			name: "Two orders processed → ok",
			setup: func(repo *mocks.MockAccrualRepository, cancel context.CancelFunc) {
				fetchCount := 0
				repo.EXPECT().FetchNewOrders(gomock.Any()).AnyTimes().DoAndReturn(func(ctx context.Context) ([]string, error) {
					if fetchCount == 0 {
						fetchCount++
						return []string{testValidOrderNumber, testAnotherValidOrderNumber}, nil
					}
					fetchCount++
					return []string{}, nil
				})
				expectUpdateOrder(repo, testValidOrderNumber).Return(nil)

				expectUpdateOrder(repo, testAnotherValidOrderNumber).DoAndReturn(func(ctx any, orderNumber any, orderStatus any, accrual any) error {
					cancel()
					return nil
				})
			},
			accrualResponses: testOKAccrualResponses,
			wantErr:          nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			ctrl := gomock.NewController(t)
			repo := mocks.NewMockAccrualRepository(ctrl)
			tt.setup(repo, cancel)

			ac := NewAccrualClient(testFakeAccrualAddress, zap.NewNop())
			ac.client.Transport = roundTripperFunc(func(r *http.Request) (*http.Response, error) {
				orderNumber := strings.TrimPrefix(r.URL.Path, "/api/orders/")
				stub, ok := tt.accrualResponses[orderNumber]
				if !ok {
					t.Errorf("asked unknown order number: %s", orderNumber)
					return nil, fmt.Errorf("asked unknown order number: %s", orderNumber)
				}
				return newOrderResp(stub), stub.transportErr
			})

			aw := NewAccrualWorker(ac, repo, zap.NewNop(), testAccrualWorkerCount, testAccrualFastPollInterval)

			err := aw.ProcessNewOrders(ctx)
			select {
			case <-aw.Done():
			case <-time.After(testWorkerDoneTimeout):
				t.Errorf("accrualWorker.ProcessNewOrders() channel is not closed")
			}

			compareErrors(t, "ProcessNewOrders()", err, tt.wantErr)
		})
	}
}

func Test_accrualWorker_consumeJobs(t *testing.T) {
	tests := []struct {
		name             string
		accrualResponses map[string]accrualStub
		setup            func(repo *mocks.MockAccrualRepository)
		orders           []string
	}{
		{
			name:             "Two orders processed → ok",
			accrualResponses: testOKAccrualResponses,
			setup: func(repo *mocks.MockAccrualRepository) {
				expectUpdateOrder(repo, testValidOrderNumber).Return(nil)
				expectUpdateOrder(repo, testAnotherValidOrderNumber).Return(nil)
			},
			orders: []string{testValidOrderNumber, testAnotherValidOrderNumber},
		},
		{
			name: "First order server error, second processed → ok",
			accrualResponses: map[string]accrualStub{
				testValidOrderNumber: {
					order:        testOrderResponses[testValidOrderNumber],
					httpStatus:   http.StatusInternalServerError,
					transportErr: nil,
				},
				testAnotherValidOrderNumber: {
					order:        testOrderResponses[testAnotherValidOrderNumber],
					httpStatus:   http.StatusOK,
					transportErr: nil,
				},
			},
			setup: func(repo *mocks.MockAccrualRepository) {
				expectUpdateOrder(repo, testAnotherValidOrderNumber).Return(nil)
			},
			orders: []string{testValidOrderNumber, testAnotherValidOrderNumber},
		},
		{
			name: "First order client error, second processed → ok",
			accrualResponses: map[string]accrualStub{
				testValidOrderNumber: {
					order:        testOrderResponses[testValidOrderNumber],
					httpStatus:   http.StatusBadRequest,
					transportErr: nil,
				},
				testAnotherValidOrderNumber: {
					order:        testOrderResponses[testAnotherValidOrderNumber],
					httpStatus:   http.StatusOK,
					transportErr: nil,
				},
			},
			setup: func(repo *mocks.MockAccrualRepository) {
				expectUpdateOrder(repo, testAnotherValidOrderNumber).Return(nil)
			},
			orders: []string{testValidOrderNumber, testAnotherValidOrderNumber},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var wg sync.WaitGroup
			ctrl := gomock.NewController(t)
			repo := mocks.NewMockAccrualRepository(ctrl)
			tt.setup(repo)
			ac := NewAccrualClient(testFakeAccrualAddress, zap.NewNop())

			ac.client.Transport = roundTripperFunc(func(r *http.Request) (*http.Response, error) {
				orderNumber := strings.TrimPrefix(r.URL.Path, "/api/orders/")
				stub, ok := tt.accrualResponses[orderNumber]
				if !ok {
					t.Errorf("asked unknown order number: %s", orderNumber)
					return nil, fmt.Errorf("asked unknown order number: %s", orderNumber)
				}
				return newOrderResp(stub), stub.transportErr
			})

			aw := NewAccrualWorker(ac, repo, zap.NewNop(), testAccrualWorkerCount, testAccrualPollInterval)

			stream := make(chan string)
			wg.Go(func() {
				aw.consumeJobs(t.Context(), stream)
			})
			for _, orderNumber := range tt.orders {
				stream <- orderNumber
			}
			close(stream)
			wg.Wait()
		})
	}
}

func Test_accrualWorker_processOrder(t *testing.T) {
	tests := []struct {
		name            string
		accrualResponse accrualStub
		orderNumber     string
		setup           func(repo *mocks.MockAccrualRepository)
		wantErr         error
	}{
		{
			// No UpdateOrder expectation: if the worker updates the order, the test fails.
			name: "Status registered → no update",
			accrualResponse: accrualStub{
				order:        testOrderResponses[testYetAnotherValidOrderNumber],
				httpStatus:   http.StatusOK,
				transportErr: nil,
			},
			orderNumber: testYetAnotherValidOrderNumber,
			setup:       func(repo *mocks.MockAccrualRepository) {},
			wantErr:     nil,
		},
		{
			name: "Update order failed → error",
			accrualResponse: accrualStub{
				order:        testOrderResponses[testValidOrderNumber],
				httpStatus:   http.StatusOK,
				transportErr: nil,
			},
			orderNumber: testValidOrderNumber,
			setup: func(repo *mocks.MockAccrualRepository) {
				expectUpdateOrder(repo, testValidOrderNumber).Return(testErrUpdateOrder)
			},
			wantErr: testErrUpdateOrder,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			repo := mocks.NewMockAccrualRepository(ctrl)
			tt.setup(repo)

			ac := NewAccrualClient(testFakeAccrualAddress, zap.NewNop())

			ac.client.Transport = roundTripperFunc(func(r *http.Request) (*http.Response, error) {
				return newOrderResp(tt.accrualResponse), tt.accrualResponse.transportErr
			})

			aw := NewAccrualWorker(ac, repo, zap.NewNop(), testAccrualWorkerCount, testAccrualPollInterval)

			err := aw.processOrder(t.Context(), tt.orderNumber)

			compareErrors(t, "processOrder()", err, tt.wantErr)
		})
	}
}

func Test_accrualWorker_getStatusWithRetries(t *testing.T) {
	tests := []struct {
		name                string
		accrualResponses    []accrualStub
		orderNumber         string
		want                Order
		wantErr             error
		cancelAfterAttempts int
		retryAfterHeader    http.Header
	}{
		{
			name: "First attempt got server error, second processed → ok",
			accrualResponses: []accrualStub{
				{
					order:        testOrderResponses[testValidOrderNumber],
					httpStatus:   http.StatusInternalServerError,
					transportErr: nil,
				},
				{
					order:        testOrderResponses[testValidOrderNumber],
					httpStatus:   http.StatusOK,
					transportErr: nil,
				},
			},
			orderNumber:         testValidOrderNumber,
			want:                testOrderResponses[testValidOrderNumber],
			wantErr:             nil,
			cancelAfterAttempts: 0,
		},
		{
			name: "Cancel context → error",
			accrualResponses: []accrualStub{
				{
					order:        testOrderResponses[testValidOrderNumber],
					httpStatus:   http.StatusInternalServerError,
					transportErr: nil,
				},
				{
					order:        testOrderResponses[testValidOrderNumber],
					httpStatus:   http.StatusOK,
					transportErr: nil,
				},
			},
			orderNumber:         testValidOrderNumber,
			want:                Order{},
			wantErr:             context.Canceled,
			cancelAfterAttempts: 1,
		},
		{
			name: "First attempt got 429, second processed → ok",
			accrualResponses: []accrualStub{
				{
					order:        testOrderResponses[testValidOrderNumber],
					httpStatus:   http.StatusTooManyRequests,
					transportErr: nil,
				},
				{
					order:        testOrderResponses[testValidOrderNumber],
					httpStatus:   http.StatusOK,
					transportErr: nil,
				},
			},
			retryAfterHeader:    http.Header{"Retry-After": []string{fmt.Sprint(testShortRetryAfter)}},
			orderNumber:         testValidOrderNumber,
			want:                testOrderResponses[testValidOrderNumber],
			wantErr:             nil,
			cancelAfterAttempts: 0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			ctrl := gomock.NewController(t)
			repo := mocks.NewMockAccrualRepository(ctrl)

			ac := NewAccrualClient(testFakeAccrualAddress, zap.NewNop())

			attempt := 0
			ac.client.Transport = roundTripperFunc(func(r *http.Request) (*http.Response, error) {
				stub := tt.accrualResponses[attempt]
				attempt++
				if tt.cancelAfterAttempts == attempt {
					cancel()
				}

				resp := newOrderResp(stub)
				if tt.retryAfterHeader.Get("Retry-After") != "" {
					resp.Header.Set("Retry-After", tt.retryAfterHeader.Get("Retry-After"))
				}
				return resp, stub.transportErr
			})

			aw := NewAccrualWorker(ac, repo, zap.NewNop(), testAccrualWorkerCount, testAccrualPollInterval)

			gotOrder, err := aw.getStatusWithRetries(ctx, tt.orderNumber)

			compareErrors(t, "getStatusWithRetries()", err, tt.wantErr)

			if !reflect.DeepEqual(gotOrder, tt.want) {
				t.Errorf("getStatusWithRetries() = %v, want %v", gotOrder, tt.want)
			}
		})
	}
}
