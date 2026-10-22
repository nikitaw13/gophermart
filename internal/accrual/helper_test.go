package accrual

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/nikitaw13/gophermart/internal/mocks"
	"github.com/nikitaw13/gophermart/internal/model"
	"go.uber.org/mock/gomock"
)

const (
	testPositiveOrderAmount        = 500.00
	testFakeAccrualAddress         = "http://localhost"
	testPositiveRetryAfter         = 30
	testZeroRetryAfter             = 0
	testNegativeRetryAfter         = -30
	testFallbackRetryAfter         = 1
	testShortRetryAfter            = 1
	testInvalidRetryAfter          = "100.00"
	testValidOrderNumber           = "79927398713"
	testAnotherValidOrderNumber    = "5412751234123452"
	testYetAnotherValidOrderNumber = "4000123456789017"
	testAccrualWorkerCount         = 1
	testAccrualPollInterval        = 3 * time.Second
	testAccrualFastPollInterval    = time.Millisecond
	testWorkerDoneTimeout          = 3 * time.Second
)

type accrualStub struct {
	order Order
	// The request reached the accrual system and was answered with this status code.
	httpStatus int
	// The accrual system was unreachable.
	transportErr error
}

var (
	errUpdateOrder    = errors.New("update order failed")
	errFetchNewOrders = errors.New("fetch new orders failed")

	testSuccessOrderResponse = Order{
		Order:   testValidOrderNumber,
		Status:  model.StatusProcessing,
		Accrual: testPositiveOrderAmount,
	}

	testOrderResponses = map[string]Order{
		testValidOrderNumber:           testSuccessOrderResponse,
		testAnotherValidOrderNumber:    {Order: testAnotherValidOrderNumber, Status: model.StatusProcessed, Accrual: 400.00},
		testYetAnotherValidOrderNumber: {Order: testYetAnotherValidOrderNumber, Status: model.StatusRegistered, Accrual: 0},
	}

	testOKAccrualResponses = map[string]accrualStub{
		testValidOrderNumber:        {order: testOrderResponses[testValidOrderNumber], httpStatus: http.StatusOK, transportErr: nil},
		testAnotherValidOrderNumber: {order: testOrderResponses[testAnotherValidOrderNumber], httpStatus: http.StatusOK, transportErr: nil},
	}

	testResponseWithPositiveRetryHeader  = &http.Response{Header: http.Header{"Retry-After": {fmt.Sprint(testPositiveRetryAfter)}}}
	testResponseWithZeroRetryHeader      = &http.Response{Header: http.Header{"Retry-After": {fmt.Sprint(testZeroRetryAfter)}}}
	testResponseWithNegativeRetryHeader  = &http.Response{Header: http.Header{"Retry-After": {fmt.Sprint(testNegativeRetryAfter)}}}
	testResponseWithDoubleMaxRetryHeader = &http.Response{Header: http.Header{"Retry-After": {fmt.Sprint(maxRetryWaitSeconds * 2)}}}
	testResponseWithEmptyRetryHeader     = &http.Response{Header: http.Header{"Retry-After": {""}}}
	testResponseWithInvalidRetryHeader   = &http.Response{Header: http.Header{"Retry-After": {testInvalidRetryAfter}}}
)

type roundTripperFunc func(*http.Request) (*http.Response, error)

func (f roundTripperFunc) RoundTrip(request *http.Request) (*http.Response, error) {
	return f(request)
}

func newResp(status int, body string) *http.Response {
	return &http.Response{
		StatusCode: status,
		Header:     http.Header{"Content-Type": []string{"application/json"}},
		Body:       io.NopCloser(strings.NewReader(body)),
	}
}

// orderJSONBody renders the order as an accrual system response body.
func orderJSONBody(order Order) string {
	return fmt.Sprintf(`{"order":"%s","status":"%v","accrual":%v}`,
		order.Order,
		order.Status,
		order.Accrual)
}

// newOrderResp builds an accrual system JSON response from the stub.
func newOrderResp(stub accrualStub) *http.Response {
	return newResp(stub.httpStatus, orderJSONBody(stub.order))
}

// expectUpdateOrder sets up an UpdateOrder expectation for the given order,
// taking the order number, status and accrual from testOrderResponses.
func expectUpdateOrder(repo *mocks.MockAccrualRepository, orderNumber string) *gomock.Call {
	return repo.EXPECT().UpdateOrder(gomock.Any(),
		testOrderResponses[orderNumber].Order,
		testOrderResponses[orderNumber].Status,
		testOrderResponses[orderNumber].Accrual)
}

func compareErrors(t *testing.T, method string, got error, wantErr error) {
	t.Helper()
	if wantErr != nil && !errors.Is(got, wantErr) {
		t.Errorf("%s error = %v, want %v", method, got, wantErr)
	}
	if wantErr == nil && !errors.Is(got, wantErr) {
		t.Errorf("%s unexpected error = %v, want %v", method, got, wantErr)
	}
}
