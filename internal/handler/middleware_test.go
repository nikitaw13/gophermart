package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/nikitaw13/gophermart/internal/mocks"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"
)

func Test_requireJSONContent(t *testing.T) {
	tests := []struct {
		name        string
		contentType string
		wantCode    int
	}{
		{
			name:        "valid json → ok",
			contentType: "application/json",
			wantCode:    http.StatusOK,
		},
		{
			name:        "missing content-type → error",
			contentType: "",
			wantCode:    http.StatusBadRequest,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusOK)
			})

			req := httptest.NewRequest(http.MethodPost, "/", nil)
			req.Header.Set("Content-Type", tt.contentType)

			w := httptest.NewRecorder()
			requireJSONContent(next).ServeHTTP(w, req)

			if w.Code != tt.wantCode {
				t.Errorf("code = %d, want %d", w.Code, tt.wantCode)
			}
		})
	}
}

func TestAPIHandler_validateToken(t *testing.T) {
	tests := []struct {
		name       string
		authHeader string
		setup      func(svc *mocks.MockService)
		wantCode   int
		wantUser   string
	}{
		{
			name:       "valid token → ok",
			authHeader: bearerPrefix + testToken,
			setup: func(svc *mocks.MockService) {
				svc.EXPECT().IdentifyUser(testToken).Return(testUsername, nil)
			},
			wantCode: http.StatusOK,
			wantUser: testUsername,
		},
		{
			name:     "missing authorization header → error",
			setup:    func(svc *mocks.MockService) {},
			wantCode: http.StatusUnauthorized,
		},
		{
			name:       "invalid token → error",
			authHeader: bearerPrefix + testInvalidToken,
			setup: func(svc *mocks.MockService) {
				svc.EXPECT().IdentifyUser(testInvalidToken).Return("", errBadToken)
			},
			wantCode: http.StatusUnauthorized,
		},
		{
			name:       "empty username → error",
			authHeader: bearerPrefix + testToken,
			setup: func(svc *mocks.MockService) {
				svc.EXPECT().IdentifyUser(testToken).Return("", nil)
			},
			wantCode: http.StatusUnauthorized,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			svc := mocks.NewMockService(ctrl)
			tt.setup(svc)

			var gotUser string
			next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				gotUser = usernameFromContext(r.Context())
				w.WriteHeader(http.StatusOK)
			})

			req := httptest.NewRequest(http.MethodGet, "/api/user/orders", nil)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}

			w := httptest.NewRecorder()
			NewAPIHandler(svc, zap.NewNop()).validateToken(next).ServeHTTP(w, req)
			if w.Code != tt.wantCode {
				t.Errorf("code = %d, want %d", w.Code, tt.wantCode)
			}
			if gotUser != tt.wantUser {
				t.Errorf("gotUser = %s, want %s", gotUser, tt.wantUser)
			}
		})
	}
}
