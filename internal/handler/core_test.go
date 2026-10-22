package handler

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"github.com/nikitaw13/gophermart/internal/model"
	"go.uber.org/zap"
)

func Test_decodeJSON(t *testing.T) {
	tests := []struct {
		name        string
		body        string
		wantDecoded bool
		wantBody    any
		wantCode    int
		wantOK      bool
	}{
		{
			name:        "success",
			body:        fmt.Sprintf(`{"login":"%s","password":"%s"}`, testUsername, testPassword),
			wantDecoded: true,
			wantBody:    model.User{Username: testUsername, Password: testPassword},
			wantCode:    http.StatusOK,
			wantOK:      true,
		},
		{
			name:        "content too large → error",
			body:        strings.Repeat("0", maxRequestBodySize+1),
			wantCode:    http.StatusRequestEntityTooLarge,
			wantDecoded: false,
			wantOK:      false,
		},
		{
			name:        "json syntax error → error",
			body:        `{'name':'Ana'}`,
			wantCode:    http.StatusBadRequest,
			wantDecoded: false,
			wantOK:      false,
		},
		{
			name:        "json unmarshal type error → error",
			body:        `{"login":false}`,
			wantCode:    http.StatusBadRequest,
			wantDecoded: false,
			wantOK:      false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			req := httptest.NewRequest(http.MethodPost, "/api/user/orders", strings.NewReader(tt.body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			gotBody, gotOK := decodeJSON[model.User](w, req, zap.NewNop())

			if tt.wantDecoded && !reflect.DeepEqual(gotBody, tt.wantBody) {
				t.Errorf("decodeJSON() got = %v, want %v", gotBody, tt.wantBody)
			}
			if w.Code != tt.wantCode {
				t.Errorf("code = %d, want %d", w.Code, tt.wantCode)
			}
			if gotOK != tt.wantOK {
				t.Errorf("decodeJSON() got = %v, want %v", gotOK, tt.wantOK)
			}
		})
	}
}
