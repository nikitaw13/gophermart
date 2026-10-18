package auth

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"go.uber.org/zap"
)

const (
	testSecretKey        = "GopherSecretKey"
	testAnotherSecretKey = "GopherAnotherSecretKey"
	testUsername         = "GopherUsername"
	testAnotherUsername  = "GopherAnotherUsername"
	testSpecialSymbols   = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!@#$%^&*()_"
)

func TestJWTManager_IssueToken(t *testing.T) {
	type payload struct {
		Username string `json:"username"`
	}
	type fields struct {
		secretKey string
		logger    *zap.Logger
	}
	type args struct {
		username string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		{
			name:   "valid username → ok",
			fields: fields{secretKey: testSecretKey},
			args:   args{username: testUsername},
		},
		{
			name:   "yet another valid username → ok",
			fields: fields{secretKey: testSecretKey},
			args:   args{username: testAnotherUsername},
		},
		{
			name:   "Empty username → ok",
			fields: fields{secretKey: testSecretKey},
			args:   args{username: ""},
		},
		{
			name:   "Username with special symbols → ok",
			fields: fields{secretKey: testSecretKey},
			args:   args{username: testSpecialSymbols},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			manager := NewJWTManager(tt.fields.secretKey, tt.fields.logger)
			gotToken, err := manager.IssueToken(tt.args.username)
			if err != nil {
				t.Fatalf("IssueToken(%q) unexpected error: %v", tt.args.username, err)
			}

			parts := strings.Split(gotToken, ".")
			payloadBytes, err := base64.RawURLEncoding.DecodeString(parts[1])
			if err != nil {
				t.Fatalf("decode payload segment failed: %v", err)
			}
			var gotPayload payload
			err = json.Unmarshal(payloadBytes, &gotPayload)
			if err != nil {
				t.Fatalf("unmarshal payload failed: %v", err)
			}

			if gotPayload.Username != tt.args.username {
				t.Errorf("JWTManager.IssueToken() = %v, want %v", gotPayload.Username, tt.args.username)
			}
		})
	}

	// Tokens signed with different secrets must produce different signatures.
	tests = []struct {
		name   string
		fields fields
		args   args
	}{
		{
			name:   "same username, different secrets → different signatures",
			fields: fields{secretKey: testSecretKey},
			args:   args{username: testUsername},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			originalKeyManager := NewJWTManager(tt.fields.secretKey, tt.fields.logger)
			differentKeyManager := NewJWTManager(testAnotherSecretKey, tt.fields.logger)

			tokenFromOriginalKey, err := originalKeyManager.IssueToken(tt.args.username)
			if err != nil {
				t.Fatalf("IssueToken with original secret failed: %v", err)
			}
			tokenFromDifferentKey, err := differentKeyManager.IssueToken(tt.args.username)
			if err != nil {
				t.Fatalf("IssueToken with different secret failed: %v", err)
			}

			originalTokenParts := strings.Split(tokenFromOriginalKey, ".")
			signatureOriginalKey, err := base64.RawURLEncoding.DecodeString(originalTokenParts[2])
			if err != nil {
				t.Fatalf("decode signature from original secret token failed: %v", err)
			}

			differentTokenParts := strings.Split(tokenFromDifferentKey, ".")
			signatureDifferentKey, err := base64.RawURLEncoding.DecodeString(differentTokenParts[2])
			if err != nil {
				t.Fatalf("decode signature from different secret token failed: %v", err)
			}
			if bytes.Equal(signatureOriginalKey, signatureDifferentKey) {
				t.Errorf("tokens signed with different secrets have identical signatures: got %q, want different from %q", signatureOriginalKey, signatureDifferentKey)
			}
		})
	}
}

func ecdsaPrivateKey() *ecdsa.PrivateKey {
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		panic(err)
	}
	return privateKey
}

func signToken(alg jwt.SigningMethod, key any, username string, ttl time.Duration) func(t *testing.T) string {
	return func(t *testing.T) string {
		t.Helper()
		token := jwt.NewWithClaims(alg, Claims{
			RegisteredClaims: jwt.RegisteredClaims{
				ExpiresAt: jwt.NewNumericDate(time.Now().Add(ttl)),
			},
			Username: username,
		})
		signed, err := token.SignedString(key)
		if err != nil {
			t.Fatalf("sign %v: %v", alg, err)
		}
		return signed
	}
}

func TestJWTManager_VerifyToken(t *testing.T) {
	type fields struct {
		secretKey string
		logger    *zap.Logger
	}
	tests := []struct {
		name    string
		fields  fields
		token   func(t *testing.T) string
		want    string
		wantErr bool
	}{
		{
			name:    "HMAC token → ok",
			fields:  fields{secretKey: testSecretKey},
			token:   signToken(jwt.SigningMethodHS256, []byte(testSecretKey), testUsername, tokenTTL),
			want:    testUsername,
			wantErr: false,
		},
		{
			name:    "ES256 token → error",
			fields:  fields{secretKey: testSecretKey},
			token:   signToken(jwt.SigningMethodES256, ecdsaPrivateKey(), testUsername, tokenTTL),
			want:    "",
			wantErr: true,
		},
		{
			name:    "wrong secret HMAC token → error",
			fields:  fields{secretKey: testSecretKey},
			token:   signToken(jwt.SigningMethodHS256, []byte(testAnotherSecretKey), testUsername, tokenTTL),
			want:    "",
			wantErr: true,
		},
		{
			name:    "expired token → error",
			fields:  fields{secretKey: testSecretKey},
			token:   signToken(jwt.SigningMethodHS256, []byte(testSecretKey), testAnotherUsername, -time.Hour),
			want:    "",
			wantErr: true,
		},
		{
			name:   "IssueToken → VerifyToken roundtrip → ok",
			fields: fields{secretKey: testSecretKey},
			token: func(t *testing.T) string {
				t.Helper()
				tok, err := NewJWTManager(testSecretKey, nil).IssueToken(testUsername)
				if err != nil {
					t.Fatalf("IssueToken: %v", err)
				}
				return tok
			},
			want:    testUsername,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			manager := NewJWTManager(tt.fields.secretKey, tt.fields.logger)
			gotUsername, err := manager.VerifyToken(tt.token(t))

			if (err != nil) != tt.wantErr {
				t.Fatalf("VerifyToken() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.wantErr {
				if gotUsername != "" {
					t.Errorf("VerifyToken() = %q on error, want empty username", gotUsername)
				}
				return
			}

			if gotUsername != tt.want {
				t.Fatalf("different users: got %q, want %q", gotUsername, tt.want)
			}
		})
	}
}
