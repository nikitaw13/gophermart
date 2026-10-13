package service

import (
	"context"
	"reflect"
	"testing"

	"github.com/nikitaw13/gophermart/internal/model"
	"go.uber.org/zap"
)

const (
	testUsername       = "GopherUsername"
	testPassword       = "GopherPassword"
	testSpecialSymbols = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!@#$%^&*()_"
)

func TestNewLoyaltyService(t *testing.T) {
	type args struct {
		repository  Repository
		authManager AuthManager
		logger      *zap.Logger
	}
	tests := []struct {
		name string
		args args
		want *LoyaltyService
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewLoyaltyService(tt.args.repository, tt.args.authManager, tt.args.logger); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewLoyaltyService() = %v, want %v", got, tt.want)
			}
		})
	}
}

// func TestLoyaltyService_RegisterUser(t *testing.T) {
// 	type args struct {
// 		ctx  context.Context
// 		user model.User
// 	}
// 	tests := []struct {
// 		name    string
// 		args    args
// 		want    string
// 		wantErr error
// 	}{
// 		{
// 			name:    "User not exist → ok",
// 			args:    args{user: model.User{Username: testUsername, Password: strings.Repeat("A", 72)}},
// 			want:    "",
// 			wantErr: nil,
// 		},
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			ctrl := gomock.NewController(t)
// 			auth := mocks.NewMockAuthManager(ctrl)
// 			repo := mocks.NewMockRepository(ctrl)
// 			var gotHash string

// 			auth.EXPECT().IssueToken(tt.args.user).Return(gomock.Any())
// 			repo.EXPECT().UserExists(gomock.Any(), tt.args.user.Username).Return(false, nil)
// 			repo.EXPECT().CreateUser(gomock.Any(), tt.args.user.Username, gomock.Any()).DoAndReturn(func(ctx context.Context, username, hash string) error {
// 				gotHash = hash
// 				return nil
// 			})

// 			svc := NewLoyaltyService(repo, auth, zap.NewNop())

// 			gotToken, err := svc.RegisterUser(tt.args.ctx, tt.args.user)

// 			if !errors.Is(err, tt.wantErr) {
// 				t.Fatalf("LoyaltyService.RegisterUser() error = %v, wantErr %v", err, tt.wantErr)
// 			}
// 			if tt.wantErr {
// 				return
// 			}
// 			if gotToken != tt.want {
// 				t.Errorf("LoyaltyService.RegisterUser() = %v, want %v", gotToken, tt.want)
// 			}
// 		})
// 	}
// }

func TestLoyaltyService_LoginUser(t *testing.T) {
	type fields struct {
		repository  Repository
		authManager AuthManager
		logger      *zap.Logger
	}
	type args struct {
		ctx  context.Context
		user model.User
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
			svc := NewLoyaltyService(tt.fields.repository, tt.fields.authManager, tt.fields.logger)

			got, err := svc.LoginUser(tt.args.ctx, tt.args.user)
			if (err != nil) != tt.wantErr {
				t.Fatalf("LoyaltyService.LoginUser() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr {
				return
			}
			if got != tt.want {
				t.Errorf("LoyaltyService.LoginUser() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestLoyaltyService_IdentifyUser(t *testing.T) {
	type fields struct {
		repository  Repository
		authManager AuthManager
		logger      *zap.Logger
	}
	type args struct {
		token string
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
			svc := NewLoyaltyService(tt.fields.repository, tt.fields.authManager, tt.fields.logger)

			got, err := svc.IdentifyUser(tt.args.token)
			if (err != nil) != tt.wantErr {
				t.Fatalf("LoyaltyService.IdentifyUser() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr {
				return
			}
			if got != tt.want {
				t.Errorf("LoyaltyService.IdentifyUser() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestLoyaltyService_LoadOrder(t *testing.T) {
	type fields struct {
		repository  Repository
		authManager AuthManager
		logger      *zap.Logger
	}
	type args struct {
		ctx         context.Context
		orderNumber string
		username    string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := NewLoyaltyService(tt.fields.repository, tt.fields.authManager, tt.fields.logger)
			if err := svc.LoadOrder(tt.args.ctx, tt.args.orderNumber, tt.args.username); (err != nil) != tt.wantErr {
				t.Errorf("LoyaltyService.LoadOrder() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestLoyaltyService_ListOrders(t *testing.T) {
	type fields struct {
		repository  Repository
		authManager AuthManager
		logger      *zap.Logger
	}
	type args struct {
		ctx      context.Context
		username string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []model.OrderResponse
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := NewLoyaltyService(tt.fields.repository, tt.fields.authManager, tt.fields.logger)

			got, err := svc.ListOrders(tt.args.ctx, tt.args.username)
			if (err != nil) != tt.wantErr {
				t.Fatalf("LoyaltyService.ListOrders() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr {
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("LoyaltyService.ListOrders() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestLoyaltyService_GetBalance(t *testing.T) {
	type fields struct {
		repository  Repository
		authManager AuthManager
		logger      *zap.Logger
	}
	type args struct {
		ctx      context.Context
		username string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    model.BalanceResponse
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := NewLoyaltyService(tt.fields.repository, tt.fields.authManager, tt.fields.logger)

			got, err := svc.GetBalance(tt.args.ctx, tt.args.username)
			if (err != nil) != tt.wantErr {
				t.Fatalf("LoyaltyService.GetBalance() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr {
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("LoyaltyService.GetBalance() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestLoyaltyService_Withdraw(t *testing.T) {
	type fields struct {
		repository  Repository
		authManager AuthManager
		logger      *zap.Logger
	}
	type args struct {
		ctx        context.Context
		username   string
		withdrawal model.WithdrawalRequest
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := NewLoyaltyService(tt.fields.repository, tt.fields.authManager, tt.fields.logger)

			if err := svc.Withdraw(tt.args.ctx, tt.args.username, tt.args.withdrawal); (err != nil) != tt.wantErr {
				t.Errorf("LoyaltyService.Withdraw() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestLoyaltyService_ListWithdrawals(t *testing.T) {
	type fields struct {
		repository  Repository
		authManager AuthManager
		logger      *zap.Logger
	}
	type args struct {
		ctx      context.Context
		username string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []model.WithdrawalResponse
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := NewLoyaltyService(tt.fields.repository, tt.fields.authManager, tt.fields.logger)

			got, err := svc.ListWithdrawals(tt.args.ctx, tt.args.username)
			if (err != nil) != tt.wantErr {
				t.Fatalf("LoyaltyService.ListWithdrawals() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr {
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("LoyaltyService.ListWithdrawals() = %v, want %v", got, tt.want)
			}
		})
	}
}
