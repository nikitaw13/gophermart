package handler

import (
	"net/http"
	"testing"

	"github.com/go-chi/chi"
	"github.com/nikitaw13/gophermart/internal/mocks"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"
)

func TestAPIHandler_Routes(t *testing.T) {
	tests := []struct {
		name      string
		method    string
		route     string
		wantFound bool
	}{
		{name: "register user", method: http.MethodPost, route: "/api/user/register", wantFound: true},
		{name: "login user", method: http.MethodPost, route: "/api/user/login", wantFound: true},
		{name: "load order", method: http.MethodPost, route: "/api/user/orders", wantFound: true},
		{name: "list orders", method: http.MethodGet, route: "/api/user/orders", wantFound: true},
		{name: "get balance", method: http.MethodGet, route: "/api/user/balance", wantFound: true},
		{name: "withdraw", method: http.MethodPost, route: "/api/user/balance/withdraw", wantFound: true},
		{name: "list withdrawals", method: http.MethodGet, route: "/api/user/withdrawals", wantFound: true},
		{name: "unknown route", method: http.MethodGet, route: "/api/user/unknown", wantFound: false},
	}

	ctrl := gomock.NewController(t)
	svc := mocks.NewMockService(ctrl)
	handler := NewAPIHandler(svc, zap.NewNop())

	registered := make(map[string]bool)
	err := chi.Walk(handler.Routes(), func(method, route string, _ http.Handler, _ ...func(http.Handler) http.Handler) error {
		registered[method+" "+route] = true
		return nil
	})
	if err != nil {
		t.Fatalf("chi.Walk() error = %v", err)
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			key := tt.method + " " + tt.route
			if got := registered[key]; got != tt.wantFound {
				t.Errorf("route %q registered = %v, want %v", key, got, tt.wantFound)
			}
		})
	}
}
