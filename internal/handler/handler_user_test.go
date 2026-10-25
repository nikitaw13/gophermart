package handler

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/nikitaw13/gophermart/internal/mocks"
	"github.com/nikitaw13/gophermart/internal/model"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"
)

func TestAPIHandler_registerUser(t *testing.T) {
	tests := []struct {
		name       string
		setup      func(svc *mocks.MockService)
		body       string
		wantCode   int
		wantHeader bool
		wantToken  string
	}{
		{
			name: "success",
			setup: func(svc *mocks.MockService) {
				svc.EXPECT().RegisterUser(gomock.Any(), model.User{Username: testUsername, Password: testPassword}).Return(testToken, nil)
			},
			body:       fmt.Sprintf(`{"login":%q, "password":%q}`, testUsername, testPassword),
			wantCode:   http.StatusOK,
			wantHeader: true,
			wantToken:  bearerPrefix + testToken,
		},
		{
			name:       "bad json → error",
			setup:      func(svc *mocks.MockService) {},
			body:       `{""}`,
			wantCode:   http.StatusBadRequest,
			wantHeader: false,
		},
		{
			name: "user already exists → error",
			setup: func(svc *mocks.MockService) {
				svc.EXPECT().RegisterUser(gomock.Any(), model.User{Username: testUsername, Password: testPassword}).Return("", model.ErrUserAlreadyExists)
			},
			body:       fmt.Sprintf(`{"login":%q, "password":%q}`, testUsername, testPassword),
			wantCode:   http.StatusConflict,
			wantHeader: false,
		},
		{
			name:       "password too long → error",
			setup:      func(svc *mocks.MockService) {},
			body:       fmt.Sprintf(`{"login":%q, "password":%q}`, testUsername, testPassword73Chars),
			wantCode:   http.StatusBadRequest,
			wantHeader: false,
		},
		{
			name: "defensive password too long → error",
			setup: func(svc *mocks.MockService) {
				svc.EXPECT().RegisterUser(gomock.Any(), model.User{Username: testUsername, Password: testPassword}).Return("", model.ErrPasswordTooLong)
			},
			body:       fmt.Sprintf(`{"login":%q, "password":%q}`, testUsername, testPassword),
			wantCode:   http.StatusBadRequest,
			wantHeader: false,
		},
		{
			name: "empty token → error",
			setup: func(svc *mocks.MockService) {
				svc.EXPECT().RegisterUser(gomock.Any(), model.User{Username: testUsername, Password: testPassword}).Return("", nil)
			},
			body:       fmt.Sprintf(`{"login":%q, "password":%q}`, testUsername, testPassword),
			wantCode:   http.StatusInternalServerError,
			wantHeader: false,
		},
		{
			name: "register user failed → error",
			setup: func(svc *mocks.MockService) {
				svc.EXPECT().RegisterUser(gomock.Any(), model.User{Username: testUsername, Password: testPassword}).Return("", errRegisterUserFailed)
			},
			body:       fmt.Sprintf(`{"login":%q, "password":%q}`, testUsername, testPassword),
			wantCode:   http.StatusInternalServerError,
			wantHeader: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			svc := mocks.NewMockService(ctrl)
			tt.setup(svc)

			req := httptest.NewRequest(http.MethodPost, "/api/user/register", strings.NewReader(tt.body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			handler := NewAPIHandler(svc, zap.NewNop())
			handler.registerUser(w, req)

			if w.Code != tt.wantCode {
				t.Errorf("code = %d, want %d", w.Code, tt.wantCode)
			}
			gotToken := w.Header().Get("Authorization")

			if tt.wantHeader && gotToken != tt.wantToken {
				t.Errorf("Authorization = %q, want %q", gotToken, tt.wantToken)
			}

			if !tt.wantHeader && gotToken != "" {
				t.Errorf("unexpected Authorization header = %q, body %q", gotToken, tt.body)
			}
		})
	}
}

func TestAPIHandler_loginUser(t *testing.T) {
	tests := []struct {
		name       string
		setup      func(svc *mocks.MockService)
		body       string
		wantCode   int
		wantHeader bool
		wantToken  string
	}{
		{
			name: "success",
			setup: func(svc *mocks.MockService) {
				svc.EXPECT().LoginUser(gomock.Any(), model.User{Username: testUsername, Password: testPassword}).Return(testToken, nil)
			},
			body:       fmt.Sprintf(`{"login":%q, "password":%q}`, testUsername, testPassword),
			wantCode:   http.StatusOK,
			wantHeader: true,
			wantToken:  bearerPrefix + testToken,
		},
		{
			name:       "bad json → error",
			setup:      func(svc *mocks.MockService) {},
			body:       `{""}`,
			wantCode:   http.StatusBadRequest,
			wantHeader: false,
		},
		{
			name: "user does not exist → error",
			setup: func(svc *mocks.MockService) {
				svc.EXPECT().LoginUser(gomock.Any(), model.User{Username: testUsername, Password: testPassword}).Return("", model.ErrUserNotFound)
			},
			body:       fmt.Sprintf(`{"login":%q, "password":%q}`, testUsername, testPassword),
			wantCode:   http.StatusUnauthorized,
			wantHeader: false,
		},
		{
			name:       "password too long → error",
			setup:      func(svc *mocks.MockService) {},
			body:       fmt.Sprintf(`{"login":%q, "password":%q}`, testUsername, testPassword73Chars),
			wantCode:   http.StatusBadRequest,
			wantHeader: false,
		},
		{
			name: "mismatched password → error",
			setup: func(svc *mocks.MockService) {
				svc.EXPECT().LoginUser(gomock.Any(), model.User{Username: testUsername, Password: testPassword}).Return("", model.ErrMismatchedPassword)
			},
			body:       fmt.Sprintf(`{"login":%q, "password":%q}`, testUsername, testPassword),
			wantCode:   http.StatusUnauthorized,
			wantHeader: false,
		},
		{
			name: "defensive password too long → error",
			setup: func(svc *mocks.MockService) {
				svc.EXPECT().LoginUser(gomock.Any(), model.User{Username: testUsername, Password: testPassword}).Return("", model.ErrPasswordTooLong)
			},
			body:       fmt.Sprintf(`{"login":%q, "password":%q}`, testUsername, testPassword),
			wantCode:   http.StatusUnauthorized,
			wantHeader: false,
		},
		{
			name: "empty token → error",
			setup: func(svc *mocks.MockService) {
				svc.EXPECT().LoginUser(gomock.Any(), model.User{Username: testUsername, Password: testPassword}).Return("", nil)
			},
			body:       fmt.Sprintf(`{"login":%q, "password":%q}`, testUsername, testPassword),
			wantCode:   http.StatusInternalServerError,
			wantHeader: false,
		},
		{
			name: "login user failed → error",
			setup: func(svc *mocks.MockService) {
				svc.EXPECT().LoginUser(gomock.Any(), model.User{Username: testUsername, Password: testPassword}).Return("", errLoginUserFailed)
			},
			body:       fmt.Sprintf(`{"login":%q, "password":%q}`, testUsername, testPassword),
			wantCode:   http.StatusInternalServerError,
			wantHeader: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			svc := mocks.NewMockService(ctrl)
			tt.setup(svc)

			req := httptest.NewRequest(http.MethodPost, "/api/user/login", strings.NewReader(tt.body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			handler := NewAPIHandler(svc, zap.NewNop())
			handler.loginUser(w, req)

			if w.Code != tt.wantCode {
				t.Errorf("code = %d, want %d", w.Code, tt.wantCode)
			}
			gotToken := w.Header().Get("Authorization")

			if tt.wantHeader && gotToken != tt.wantToken {
				t.Errorf("Authorization = %q, want %q", gotToken, tt.wantToken)
			}

			if !tt.wantHeader && gotToken != "" {
				t.Errorf("unexpected Authorization header = %q, body %q", gotToken, tt.body)
			}
		})
	}
}
