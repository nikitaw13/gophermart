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

func TestAPIHandler_loadOrder(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(svc *mocks.MockService)
		body     string
		wantCode int
	}{
		{
			name: "success",
			setup: func(svc *mocks.MockService) {
				svc.EXPECT().LoadOrder(gomock.Any(), testValidOrderNumber, testUsername).Return(nil)
			},
			body:     testValidOrderNumber,
			wantCode: http.StatusAccepted,
		},
		{
			name: "order number with new line → ok",
			setup: func(svc *mocks.MockService) {
				svc.EXPECT().LoadOrder(gomock.Any(), testValidOrderNumber, testUsername).Return(nil)
			},
			body:     testValidOrderNumber + "\n",
			wantCode: http.StatusAccepted,
		},
		{
			name:     "content too large → error",
			setup:    func(svc *mocks.MockService) {},
			body:     strings.Repeat("0", maxRequestBodySize+1),
			wantCode: http.StatusRequestEntityTooLarge,
		},
		{
			name: "invalid order number → error",
			setup: func(svc *mocks.MockService) {
				svc.EXPECT().LoadOrder(gomock.Any(), testInvalidOrderNumber, testUsername).Return(model.ErrInvalidOrderNumber)
			},
			body:     testInvalidOrderNumber,
			wantCode: http.StatusUnprocessableEntity,
		},
		{
			name: "missing order number → error",
			setup: func(svc *mocks.MockService) {
				svc.EXPECT().LoadOrder(gomock.Any(), "", testUsername).Return(model.ErrMissingOrderNumber)
			},
			body:     "",
			wantCode: http.StatusUnprocessableEntity,
		},
		{
			name: "order already loaded by another user → error",
			setup: func(svc *mocks.MockService) {
				svc.EXPECT().LoadOrder(gomock.Any(), testValidOrderNumber, testUsername).Return(model.ErrOrderExistsOtherUser)
			},
			body:     testValidOrderNumber,
			wantCode: http.StatusConflict,
		},
		{
			name: "order already loaded by the same user → ok",
			setup: func(svc *mocks.MockService) {
				svc.EXPECT().LoadOrder(gomock.Any(), testValidOrderNumber, testUsername).Return(model.ErrOrderExistsSameUser)
			},
			body:     testValidOrderNumber,
			wantCode: http.StatusOK,
		},
		{
			name: "failed to load order → error",
			setup: func(svc *mocks.MockService) {
				svc.EXPECT().LoadOrder(gomock.Any(), testValidOrderNumber, testUsername).Return(errLoadOrderFailed)
			},
			body:     testValidOrderNumber,
			wantCode: http.StatusInternalServerError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			svc := mocks.NewMockService(ctrl)
			tt.setup(svc)

			req := httptest.NewRequest(http.MethodPost, "/api/user/orders", strings.NewReader(tt.body))
			req.Header.Set("Content-Type", "text/plain")
			w := httptest.NewRecorder()
			req = req.WithContext(context.WithValue(req.Context(), usernameKey, testUsername))

			handler := NewAPIHandler(svc, zap.NewNop())
			handler.loadOrder(w, req)

			if w.Code != tt.wantCode {
				t.Errorf("code = %d, want %d", w.Code, tt.wantCode)
			}
		})
	}
}

func TestAPIHandler_listOrders(t *testing.T) {
	tests := []struct {
		name       string
		setup      func(svc *mocks.MockService)
		wantCode   int
		wantBody   bool
		wantOrders []model.OrderResponse
	}{
		{
			name: "many orders → ok",
			setup: func(svc *mocks.MockService) {
				svc.EXPECT().ListOrders(gomock.Any(), testUsername).Return(testListOrdersResponse, nil)
			},
			wantCode:   http.StatusOK,
			wantBody:   true,
			wantOrders: testWantListOrdersResponse,
		},
		{
			name: "failed to list orders → error",
			setup: func(svc *mocks.MockService) {
				svc.EXPECT().ListOrders(gomock.Any(), testUsername).Return(nil, errListOrdersFailed)
			},
			wantCode:   http.StatusInternalServerError,
			wantBody:   false,
			wantOrders: []model.OrderResponse{},
		},
		{
			name: "no orders → ok",
			setup: func(svc *mocks.MockService) {
				svc.EXPECT().ListOrders(gomock.Any(), testUsername).Return(nil, nil)
			},
			wantCode: http.StatusNoContent,
			wantBody: false,
		},
		{
			name: "marshal failed → ok",
			setup: func(svc *mocks.MockService) {
				svc.EXPECT().ListOrders(gomock.Any(), testUsername).Return(testInvalidListOrdersResponse, nil)
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

			req := httptest.NewRequest(http.MethodGet, "/api/user/orders", nil)
			w := httptest.NewRecorder()
			req = req.WithContext(context.WithValue(req.Context(), usernameKey, testUsername))

			handler := NewAPIHandler(svc, zap.NewNop())
			handler.listOrders(w, req)

			if w.Code != tt.wantCode {
				t.Errorf("code = %d, want %d", w.Code, tt.wantCode)
			}

			gotBody := strings.TrimSuffix(w.Body.String(), "\n")
			wantJSON, err := json.Marshal(tt.wantOrders)
			if err != nil {
				t.Errorf("marshaling want response failed: %v", err)
			}

			if tt.wantBody && !reflect.DeepEqual(gotBody, string(wantJSON)) {
				t.Errorf("listOrders() = %v, want %v", gotBody, string(wantJSON))
			}
		})
	}
}
