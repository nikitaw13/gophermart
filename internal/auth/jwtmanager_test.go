// Package auth issues and verifies JWT-based authentication tokens.
// Tokens are consumed by the service (issued on register) and the middleware (verified per request).
// It uses only a secret key and HMAC signing, with no database or HTTP dependencies.
package auth

import (
	"testing"

	"go.uber.org/zap"
)

const secretKey = "TEST"

func TestJWTManager_IssueToken(t *testing.T) {
	type fields struct {
		secretKey string
		logger    *zap.Logger
	}
	type args struct {
		username string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    string
		wantErr bool
	}{
		{
			name:    "",
			fields:  fields{},
			args:    args{},
			want:    "",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewJWTManager(tt.fields.secretKey, tt.fields.logger)
			got, err := m.IssueToken(tt.args.username)
			if (err != nil) != tt.wantErr {
				t.Fatalf("JWTManager.IssueToken() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr {
				return
			}
			if got != tt.want {
				t.Errorf("JWTManager.IssueToken() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestJWTManager_VerifyToken(t *testing.T) {
	type fields struct {
		secretKey string
		logger    *zap.Logger
	}
	type args struct {
		tokenString string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    string
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewJWTManager(tt.fields.secretKey, tt.fields.logger)
			got, err := m.VerifyToken(tt.args.tokenString)
			if (err != nil) != tt.wantErr {
				t.Fatalf("JWTManager.VerifyToken() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr {
				return
			}
			if got != tt.want {
				t.Errorf("JWTManager.VerifyToken() = %v, want %v", got, tt.want)
			}
		})
	}
}
