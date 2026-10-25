package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"github.com/nikitaw13/gophermart/internal/mocks"
	"github.com/nikitaw13/gophermart/internal/model"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"
)

func TestAPIHandler_withdraw(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(svc *mocks.MockService)
		body     string
		wantCode int
	}{
		{
			name: "success",
			setup: func(svc *mocks.MockService) {
				svc.EXPECT().Withdraw(gomock.Any(), testUsername,
					model.WithdrawalRequest{
						OrderNumber: testValidOrderNumber,
						Amount:      testValidAmount,
					}).Return(nil)
			},
			body:     fmt.Sprintf(`{"order":%q, "sum": %v}`, testValidOrderNumber, testValidAmount),
			wantCode: http.StatusOK,
		},
		{
			name: "non positive amount → error",
			setup: func(svc *mocks.MockService) {
				svc.EXPECT().Withdraw(gomock.Any(), testUsername,
					model.WithdrawalRequest{
						OrderNumber: testValidOrderNumber,
						Amount:      testInvalidAmount,
					}).Return(model.ErrNonPositiveAmount)
			},
			body:     fmt.Sprintf(`{"order":%q, "sum": %v}`, testValidOrderNumber, testInvalidAmount),
			wantCode: http.StatusUnprocessableEntity,
		},
		{
			name: "insufficient funds → error",
			setup: func(svc *mocks.MockService) {
				svc.EXPECT().Withdraw(gomock.Any(), testUsername,
					model.WithdrawalRequest{
						OrderNumber: testValidOrderNumber,
						Amount:      testValidAmount,
					}).Return(model.ErrInsufficientFunds)
			},
			body:     fmt.Sprintf(`{"order":%q, "sum": %v}`, testValidOrderNumber, testValidAmount),
			wantCode: http.StatusPaymentRequired,
		},
		{
			name: "missing order number → error",
			setup: func(svc *mocks.MockService) {
				svc.EXPECT().Withdraw(gomock.Any(), testUsername,
					model.WithdrawalRequest{
						OrderNumber: "",
						Amount:      testValidAmount,
					}).Return(model.ErrMissingOrderNumber)
			},
			body:     fmt.Sprintf(`{"order":%q, "sum": %v}`, "", testValidAmount),
			wantCode: http.StatusUnprocessableEntity,
		},
		{
			name: "invalid order number → error",
			setup: func(svc *mocks.MockService) {
				svc.EXPECT().Withdraw(gomock.Any(), testUsername,
					model.WithdrawalRequest{
						OrderNumber: testInvalidOrderNumber,
						Amount:      testValidAmount,
					}).Return(model.ErrInvalidOrderNumber)
			},
			body:     fmt.Sprintf(`{"order":%q, "sum": %v}`, testInvalidOrderNumber, testValidAmount),
			wantCode: http.StatusUnprocessableEntity,
		},
		{
			name: "withdrawal already processed by another user → error",
			setup: func(svc *mocks.MockService) {
				svc.EXPECT().Withdraw(gomock.Any(), testUsername,
					model.WithdrawalRequest{
						OrderNumber: testValidOrderNumber,
						Amount:      testValidAmount,
					}).Return(model.ErrWithdrawalExistsOtherUser)
			},
			body:     fmt.Sprintf(`{"order":%q, "sum": %v}`, testValidOrderNumber, testValidAmount),
			wantCode: http.StatusConflict,
		},
		{
			name: "withdrawal already processed by the same user → ok",
			setup: func(svc *mocks.MockService) {
				svc.EXPECT().Withdraw(gomock.Any(), testUsername,
					model.WithdrawalRequest{
						OrderNumber: testValidOrderNumber,
						Amount:      testValidAmount,
					}).Return(model.ErrWithdrawalExistsSameUser)
			},
			body:     fmt.Sprintf(`{"order":%q, "sum": %v}`, testValidOrderNumber, testValidAmount),
			wantCode: http.StatusOK,
		},
		{
			name: "withdraw failed → error",
			setup: func(svc *mocks.MockService) {
				svc.EXPECT().Withdraw(gomock.Any(), testUsername,
					model.WithdrawalRequest{
						OrderNumber: testValidOrderNumber,
						Amount:      testValidAmount,
					}).Return(errWithdrawFailed)
			},
			body:     fmt.Sprintf(`{"order":%q, "sum": %v}`, testValidOrderNumber, testValidAmount),
			wantCode: http.StatusInternalServerError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			svc := mocks.NewMockService(ctrl)
			tt.setup(svc)

			req := httptest.NewRequest(http.MethodPost, "/api/user/balance/withdraw", strings.NewReader(tt.body))
			req.Header.Set("Content-Type", "application/json")
			req = req.WithContext(context.WithValue(req.Context(), usernameKey, testUsername))
			w := httptest.NewRecorder()

			handler := NewAPIHandler(svc, zap.NewNop())
			handler.withdraw(w, req)

			if w.Code != tt.wantCode {
				t.Errorf("code = %d, want %d", w.Code, tt.wantCode)
			}
		})
	}
}

func TestAPIHandler_listWithdrawals(t *testing.T) {
	tests := []struct {
		name            string
		setup           func(svc *mocks.MockService)
		wantCode        int
		wantBody        bool
		wantWithdrawals []model.WithdrawalResponse
	}{
		{
			name: "many withdrawals → ok",
			setup: func(svc *mocks.MockService) {
				svc.EXPECT().ListWithdrawals(gomock.Any(), testUsername).Return(testListWithdrawalsResponse, nil)
			},
			wantCode:        http.StatusOK,
			wantBody:        true,
			wantWithdrawals: testWantListWithdrawalsResponse,
		},
		{
			name: "failed to list withdrawals → error",
			setup: func(svc *mocks.MockService) {
				svc.EXPECT().ListWithdrawals(gomock.Any(), testUsername).Return(nil, errListWithdrawalsFailed)
			},
			wantCode:        http.StatusInternalServerError,
			wantBody:        false,
			wantWithdrawals: []model.WithdrawalResponse{},
		},
		{
			name: "no withdrawals → ok",
			setup: func(svc *mocks.MockService) {
				svc.EXPECT().ListWithdrawals(gomock.Any(), testUsername).Return(nil, nil)
			},
			wantCode:        http.StatusNoContent,
			wantBody:        false,
			wantWithdrawals: []model.WithdrawalResponse{},
		},
		{
			name: "marshal failed → ok",
			setup: func(svc *mocks.MockService) {
				svc.EXPECT().ListWithdrawals(gomock.Any(), testUsername).Return(testInvalidListWithdrawalsResponse, nil)
			},
			wantCode: http.StatusOK,
			wantBody: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			svc := mocks.NewMockService(ctrl)
			tt.setup(svc)

			req := httptest.NewRequest(http.MethodGet, "/api/user/withdrawals", nil)
			req = req.WithContext(context.WithValue(req.Context(), usernameKey, testUsername))
			w := httptest.NewRecorder()

			handler := NewAPIHandler(svc, zap.NewNop())
			handler.listWithdrawals(w, req)

			if w.Code != tt.wantCode {
				t.Errorf("code = %d, want %d", w.Code, tt.wantCode)
			}

			gotBody := strings.TrimSuffix(w.Body.String(), "\n")
			wantJSON, err := json.Marshal(tt.wantWithdrawals)
			if err != nil {
				t.Errorf("marshaling want response failed: %v", err)
			}
			if tt.wantBody && !reflect.DeepEqual(gotBody, string(wantJSON)) {
				t.Errorf("listWithdrawals() = %v, want %v", gotBody, string(wantJSON))
			}
		})
	}
}
