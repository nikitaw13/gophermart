package handler

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/nikitaw13/gophermart/internal/mocks"
	"github.com/nikitaw13/gophermart/internal/model"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"
)

func TestAPIHandler_getBalance(t *testing.T) {
	tests := []struct {
		name  string
		setup func(svc *mocks.MockService)
	}{
		{
			name: "success",
			setup: func(svc *mocks.MockService) {
				svc.EXPECT().GetBalance(gomock.Any(), testUsername).Return(testBalanceResponse, nil)
			},
		},
		{
			name: "get balance failed → error",
			setup: func(svc *mocks.MockService) {
				svc.EXPECT().GetBalance(gomock.Any(), testUsername).Return(model.BalanceResponse{}, errGetBalanceFailed)
			},
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
		})
	}
}
