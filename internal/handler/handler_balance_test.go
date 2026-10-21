package handler

import (
	"context"
	"encoding/json"
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

func TestAPIHandler_getBalance(t *testing.T) {
	tests := []struct {
		name        string
		setup       func(svc *mocks.MockService)
		wantCode    int
		wantBody    bool
		wantBalance model.BalanceResponse
	}{
		{
			name: "success",
			setup: func(svc *mocks.MockService) {
				svc.EXPECT().GetBalance(gomock.Any(), testUsername).Return(testGetBalanceResponse, nil)
			},
			wantCode:    http.StatusOK,
			wantBody:    true,
			wantBalance: testWantGetBalanceResponse,
		},
		{
			name: "get balance failed → error",
			setup: func(svc *mocks.MockService) {
				svc.EXPECT().GetBalance(gomock.Any(), testUsername).Return(model.BalanceResponse{}, errGetBalanceFailed)
			},
			wantCode: http.StatusInternalServerError,
			wantBody: false,
		},
		{
			name: "empty balance → ok",
			setup: func(svc *mocks.MockService) {
				svc.EXPECT().GetBalance(gomock.Any(), testUsername).Return(model.BalanceResponse{}, nil)
			},
			wantCode: http.StatusOK,
			wantBody: false,
		},
		{
			name: "marshal failed → ok",
			setup: func(svc *mocks.MockService) {
				svc.EXPECT().GetBalance(gomock.Any(), testUsername).Return(testInvalidGetBalanceResponse, nil)
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

			req := httptest.NewRequest(http.MethodGet, "/api/user/balance", nil)
			req = req.WithContext(context.WithValue(req.Context(), usernameKey, testUsername))
			w := httptest.NewRecorder()

			handler := NewAPIHandler(svc, zap.NewNop())
			handler.getBalance(w, req)

			if w.Code != tt.wantCode {
				t.Errorf("code = %d, want %d", w.Code, tt.wantCode)
			}

			gotBody := strings.TrimSuffix(w.Body.String(), "\n")
			wantJSON, err := json.Marshal(tt.wantBalance)
			if err != nil {
				t.Errorf("marshaling want response failed: %v", err)
			}

			if tt.wantBody && !reflect.DeepEqual(gotBody, string(wantJSON)) {
				t.Errorf("getBalance() = %v, want %v", gotBody, string(wantJSON))
			}
		})
	}
}
