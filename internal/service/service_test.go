package service

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/nikitaw13/gophermart/internal/auth"
	"github.com/nikitaw13/gophermart/internal/mocks"
	"github.com/nikitaw13/gophermart/internal/model"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

func TestLoyaltyService_RegisterUser(t *testing.T) {
	tests := []struct {
		name    string
		user    model.User
		setup   func(t *testing.T, repo *mocks.MockRepository, authMgr *mocks.MockAuthManager)
		want    string
		wantErr error
	}{
		{
			name: "New user → ok",
			user: model.User{Username: testUsername, Password: testPassword72Symbols},
			setup: func(t *testing.T, repo *mocks.MockRepository, authMgr *mocks.MockAuthManager) {
				repo.EXPECT().CreateUser(gomock.Any(), testUsername, gomock.Any()).DoAndReturn(func(ctx context.Context, username, hash string) error {
					err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(testPassword72Symbols))
					if err != nil {
						t.Errorf("CompareHashAndPassword() error = %v, wantErr %v", err, nil)
					}
					return nil
				})
				authMgr.EXPECT().IssueToken(testUsername).Return(testToken, nil)
			},
			want:    testToken,
			wantErr: nil,
		},
		{
			name: "User already exists → error",
			user: model.User{Username: testExistingUsername, Password: testPassword72Symbols},
			setup: func(t *testing.T, repo *mocks.MockRepository, authMgr *mocks.MockAuthManager) {
				repo.EXPECT().CreateUser(gomock.Any(), testExistingUsername, gomock.Any()).Return(model.ErrUserAlreadyExists)
			},
			want:    testEmptyToken,
			wantErr: model.ErrUserAlreadyExists,
		},
		{
			name:    "Password too long → error",
			user:    model.User{Username: testUsername, Password: testPassword73Symbols},
			setup:   func(t *testing.T, repo *mocks.MockRepository, authMgr *mocks.MockAuthManager) {},
			want:    testEmptyToken,
			wantErr: model.ErrPasswordTooLong,
		},
		{
			name: "Issuing token failed → error",
			user: model.User{Username: testUsername, Password: testPassword72Symbols},
			setup: func(t *testing.T, repo *mocks.MockRepository, authMgr *mocks.MockAuthManager) {
				repo.EXPECT().CreateUser(gomock.Any(), testUsername, gomock.Any()).Return(nil)
				authMgr.EXPECT().IssueToken(testUsername).Return(testEmptyToken, errTestIssueToken)
			},
			want:    testEmptyToken,
			wantErr: errTestIssueToken,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			authMgr := mocks.NewMockAuthManager(ctrl)
			repo := mocks.NewMockRepository(ctrl)
			tt.setup(t, repo, authMgr)

			svc := NewLoyaltyService(repo, authMgr, zap.NewNop())
			gotToken, err := svc.RegisterUser(t.Context(), tt.user)

			if tt.wantErr != nil && !errors.Is(err, tt.wantErr) {
				t.Errorf("svc.RegisterUser() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr == nil && !errors.Is(err, tt.wantErr) {
				t.Errorf("svc.RegisterUser() unexpected error = %v, wantErr %v", err, tt.wantErr)
			}
			if gotToken != tt.want {
				t.Errorf("svc.RegisterUser() = %v, want %v", gotToken, tt.want)
			}
		})
	}
}

func bcryptHash(t *testing.T, password string) string {
	t.Helper()
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.MinCost)
	if err != nil {
		t.Errorf("bcrypt.GenerateFromPassword() error = %v, wantErr %v", err, nil)
	}
	return string(hashedPassword)
}

func TestLoyaltyService_LoginUser(t *testing.T) {
	tests := []struct {
		name    string
		user    model.User
		setup   func(t *testing.T, repo *mocks.MockRepository, authMgr *mocks.MockAuthManager)
		want    string
		wantErr error
	}{
		{
			name: "User exists → ok",
			user: model.User{Username: testUsername, Password: testPassword72Symbols},
			setup: func(t *testing.T, repo *mocks.MockRepository, authMgr *mocks.MockAuthManager) {
				repo.EXPECT().GetPasswordHash(gomock.Any(), testUsername).Return(bcryptHash(t, testPassword72Symbols), nil)
				authMgr.EXPECT().IssueToken(testUsername).Return(testToken, nil)
			},
			want:    testToken,
			wantErr: nil,
		},
		{
			name: "User not found → error",
			user: model.User{Username: testUsername, Password: testPassword72Symbols},
			setup: func(t *testing.T, repo *mocks.MockRepository, authMgr *mocks.MockAuthManager) {
				repo.EXPECT().GetPasswordHash(gomock.Any(), testUsername).Return("", model.ErrUserNotFound)
			},
			want:    testEmptyToken,
			wantErr: model.ErrUserNotFound,
		},
		{
			name: "Password mismatch → error",
			user: model.User{Username: testUsername, Password: testPassword},
			setup: func(t *testing.T, repo *mocks.MockRepository, authMgr *mocks.MockAuthManager) {
				repo.EXPECT().GetPasswordHash(gomock.Any(), testUsername).Return(bcryptHash(t, testPassword72Symbols), nil)
			},
			want:    testEmptyToken,
			wantErr: model.ErrMismatchedPassword,
		},
		{
			name: "Invalid hash → error",
			user: model.User{Username: testUsername, Password: testPassword},
			setup: func(t *testing.T, repo *mocks.MockRepository, authMgr *mocks.MockAuthManager) {
				repo.EXPECT().GetPasswordHash(gomock.Any(), testUsername).Return(testInvalidHash, nil)
			},
			want:    testEmptyToken,
			wantErr: bcrypt.ErrHashTooShort,
		},
		{
			name: "Issuing token failed → error",
			user: model.User{Username: testUsername, Password: testPassword72Symbols},
			setup: func(t *testing.T, repo *mocks.MockRepository, authMgr *mocks.MockAuthManager) {
				repo.EXPECT().GetPasswordHash(gomock.Any(), testUsername).Return(bcryptHash(t, testPassword72Symbols), nil)
				authMgr.EXPECT().IssueToken(testUsername).Return(testEmptyToken, errTestIssueToken)
			},
			want:    testEmptyToken,
			wantErr: errTestIssueToken,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			authMgr := mocks.NewMockAuthManager(ctrl)
			repo := mocks.NewMockRepository(ctrl)
			tt.setup(t, repo, authMgr)

			svc := NewLoyaltyService(repo, authMgr, zap.NewNop())
			gotToken, err := svc.LoginUser(t.Context(), tt.user)

			if tt.wantErr != nil && !errors.Is(err, tt.wantErr) {
				t.Errorf("LoginUser() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr == nil && !errors.Is(err, tt.wantErr) {
				t.Errorf("LoginUser() unexpected error = %v, wantErr %v", err, tt.wantErr)
			}
			if gotToken != tt.want {
				t.Errorf("LoginUser() = %v, want %v", gotToken, tt.want)
			}
		})
	}
}

func TestLoyaltyService_IdentifyUser(t *testing.T) {
	tests := []struct {
		name    string
		token   string
		setup   func(t *testing.T, authMgr *mocks.MockAuthManager)
		want    string
		wantErr error
	}{
		{
			name:  "Valid token → ok",
			token: testToken,
			setup: func(t *testing.T, authMgr *mocks.MockAuthManager) {
				authMgr.EXPECT().VerifyToken(testToken).Return(testUsername, nil)
			},
			want:    testUsername,
			wantErr: nil,
		},
		{
			name:  "Invalid token → error",
			token: testToken,
			setup: func(t *testing.T, authMgr *mocks.MockAuthManager) {
				authMgr.EXPECT().VerifyToken(testToken).Return(testAnotherUsername, auth.ErrTokenNotValid)
			},
			want:    testEmptyUsername,
			wantErr: auth.ErrTokenNotValid,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			authMgr := mocks.NewMockAuthManager(ctrl)
			repo := mocks.NewMockRepository(ctrl)
			svc := NewLoyaltyService(repo, authMgr, zap.NewNop())
			tt.setup(t, authMgr)

			gotUsername, err := svc.IdentifyUser(tt.token)

			if tt.wantErr != nil && !errors.Is(err, tt.wantErr) {
				t.Errorf("IdentifyUser() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr == nil && !errors.Is(err, tt.wantErr) {
				t.Errorf("IdentifyUser() unexpected error = %v, wantErr %v", err, tt.wantErr)
			}
			if gotUsername != tt.want {
				t.Errorf("IdentifyUser() = %v, want %v", gotUsername, tt.want)
			}
		})
	}
}

func TestLoyaltyService_LoadOrder(t *testing.T) {
	type args struct {
		orderNumber string
		username    string
	}
	tests := []struct {
		name    string
		args    args
		setup   func(repo *mocks.MockRepository)
		wantErr error
	}{
		{
			name: "Free valid order number → ok",
			args: args{orderNumber: testValidOrderNumber, username: testUsername},
			setup: func(repo *mocks.MockRepository) {
				result := model.OrderRegistrationResult{Inserted: true, Owner: testUsername}
				repo.EXPECT().RegisterOrder(gomock.Any(), testValidOrderNumber, testUsername).Return(result, nil)
			},
			wantErr: nil,
		},
		{
			name:    "Empty order number → error",
			args:    args{orderNumber: testEmptyOrderNumber, username: testUsername},
			setup:   func(repo *mocks.MockRepository) {},
			wantErr: model.ErrMissingOrderNumber,
		},
		{
			name:    "Invalid order number → error",
			args:    args{orderNumber: testInvalidOrderNumber, username: testUsername},
			setup:   func(repo *mocks.MockRepository) {},
			wantErr: model.ErrInvalidOrderNumber,
		},
		{
			name: "Order already exists, same owner → error",
			args: args{orderNumber: testValidOrderNumber, username: testUsername},
			setup: func(repo *mocks.MockRepository) {
				result := model.OrderRegistrationResult{Inserted: false, Owner: testUsername}
				repo.EXPECT().RegisterOrder(gomock.Any(), testValidOrderNumber, testUsername).Return(result, nil)
			},
			wantErr: model.ErrOrderExistsSameUser,
		},
		{
			name: "Order already exists, different owner → error",
			args: args{orderNumber: testValidOrderNumber, username: testUsername},
			setup: func(repo *mocks.MockRepository) {
				result := model.OrderRegistrationResult{Inserted: false, Owner: testAnotherUsername}
				repo.EXPECT().RegisterOrder(gomock.Any(), testValidOrderNumber, testUsername).Return(result, nil)
			},
			wantErr: model.ErrOrderExistsOtherUser,
		},
		{
			name: "Loading order failed → error",
			args: args{orderNumber: testValidOrderNumber, username: testUsername},
			setup: func(repo *mocks.MockRepository) {
				result := model.OrderRegistrationResult{Inserted: false, Owner: ""}
				repo.EXPECT().RegisterOrder(gomock.Any(), testValidOrderNumber, testUsername).Return(result, errTestLoadOrder)
			},
			wantErr: errTestLoadOrder,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			authMgr := mocks.NewMockAuthManager(ctrl)
			repo := mocks.NewMockRepository(ctrl)
			tt.setup(repo)

			svc := NewLoyaltyService(repo, authMgr, zap.NewNop())
			err := svc.LoadOrder(t.Context(), tt.args.orderNumber, tt.args.username)

			if tt.wantErr != nil && !errors.Is(err, tt.wantErr) {
				t.Errorf("LoadOrder() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr == nil && !errors.Is(err, tt.wantErr) {
				t.Errorf("LoadOrder() unexpected error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestLoyaltyService_ListOrders(t *testing.T) {
	tests := []struct {
		name     string
		username string
		setup    func(repo *mocks.MockRepository)
		want     []model.OrderResponse
		wantErr  error
	}{
		{
			name:     "User has three orders → ok",
			username: testUsername,
			setup: func(repo *mocks.MockRepository) {
				repo.EXPECT().ListOrders(gomock.Any(), testUsername).Return(testThreeOrderResponses, nil)
			},
			want:    expectedThreeOrderResponses,
			wantErr: nil,
		},
		{
			name:     "User has zero orders → ok",
			username: testUsername,
			setup: func(repo *mocks.MockRepository) {
				repo.EXPECT().ListOrders(gomock.Any(), testUsername).Return(testEmptyOrderResponses, nil)
			},
			want:    testEmptyOrderResponses,
			wantErr: nil,
		},
		{
			name:     "List orders failed → error",
			username: testUsername,
			setup: func(repo *mocks.MockRepository) {
				repo.EXPECT().ListOrders(gomock.Any(), testUsername).Return(testSingleOrderResponses, errTestListOrders)
			},
			want:    testSingleOrderResponses,
			wantErr: errTestListOrders,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			authMgr := mocks.NewMockAuthManager(ctrl)
			repo := mocks.NewMockRepository(ctrl)
			tt.setup(repo)
			svc := NewLoyaltyService(repo, authMgr, zap.NewNop())

			gotOrders, err := svc.ListOrders(t.Context(), tt.username)
			if tt.wantErr != nil && !errors.Is(err, tt.wantErr) {
				t.Errorf("ListOrders() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr == nil && !errors.Is(err, tt.wantErr) {
				t.Errorf("ListOrders() unexpected error = %v, wantErr %v", err, tt.wantErr)
			}
			if !reflect.DeepEqual(gotOrders, tt.want) {
				t.Errorf("ListOrders() = %v, want %v", gotOrders, tt.want)
			}
		})
	}
}

func TestLoyaltyService_GetBalance(t *testing.T) {
	tests := []struct {
		name     string
		username string
		setup    func(repo *mocks.MockRepository)
		want     model.BalanceResponse
		wantErr  error
	}{
		{
			name:     "User has non-zero balance → ok",
			username: testUsername,
			setup: func(repo *mocks.MockRepository) {
				repo.EXPECT().GetBalance(gomock.Any(), testUsername).Return(testNonZeroBalance, nil)
			},
			want:    expectedNonZeroBalance,
			wantErr: nil,
		},
		{
			name:     "Get balance failed → error",
			username: testUsername,
			setup: func(repo *mocks.MockRepository) {
				repo.EXPECT().GetBalance(gomock.Any(), testUsername).Return(testEmptyBalance, errTestGetBalance)
			},
			want:    testEmptyBalance,
			wantErr: errTestGetBalance,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			authMgr := mocks.NewMockAuthManager(ctrl)
			repo := mocks.NewMockRepository(ctrl)
			tt.setup(repo)
			svc := NewLoyaltyService(repo, authMgr, zap.NewNop())

			gotBalance, err := svc.GetBalance(t.Context(), tt.username)
			if tt.wantErr != nil && !errors.Is(err, tt.wantErr) {
				t.Errorf("GetBalance() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr == nil && !errors.Is(err, tt.wantErr) {
				t.Errorf("GetBalance() unexpected error = %v, wantErr %v", err, tt.wantErr)
			}
			if !reflect.DeepEqual(gotBalance, tt.want) {
				t.Errorf("GetBalance() = %v, want %v", gotBalance, tt.want)
			}
		})
	}
}

func TestLoyaltyService_Withdraw(t *testing.T) {
	type args struct {
		username   string
		withdrawal model.WithdrawalRequest
	}
	tests := []struct {
		name    string
		args    args
		setup   func(repo *mocks.MockRepository)
		wantErr error
	}{
		{
			name: "Valid order number, positive balance → ok",
			args: args{username: testUsername, withdrawal: model.WithdrawalRequest{
				OrderNumber: testValidOrderNumber,
				Amount:      testPositiveOrderAmount,
			}},
			setup: func(repo *mocks.MockRepository) {
				request := model.WithdrawalRequest{OrderNumber: testValidOrderNumber, Amount: testPositiveOrderAmount}
				result := model.WithdrawalRegistrationResult{Inserted: true, Owner: testUsername}
				repo.EXPECT().Withdraw(gomock.Any(), testUsername, request).Return(result, nil)
			},
			wantErr: nil,
		},
		{
			name: "Invalid order number → error",
			args: args{username: testUsername, withdrawal: model.WithdrawalRequest{
				OrderNumber: testInvalidOrderNumber,
				Amount:      testPositiveOrderAmount,
			}},
			setup:   func(repo *mocks.MockRepository) {},
			wantErr: model.ErrInvalidOrderNumber,
		},
		{
			name: "Empty order number → error",
			args: args{username: testUsername, withdrawal: model.WithdrawalRequest{
				OrderNumber: testEmptyOrderNumber,
				Amount:      testPositiveOrderAmount,
			}},
			setup:   func(repo *mocks.MockRepository) {},
			wantErr: model.ErrMissingOrderNumber,
		},
		{
			name: "Valid order number, insufficient funds → error",
			args: args{username: testUsername, withdrawal: model.WithdrawalRequest{
				OrderNumber: testValidOrderNumber,
				Amount:      testPositiveOrderAmount,
			}},
			setup: func(repo *mocks.MockRepository) {
				request := model.WithdrawalRequest{OrderNumber: testValidOrderNumber, Amount: testPositiveOrderAmount}
				result := model.WithdrawalRegistrationResult{Inserted: false, Owner: testEmptyUsername}
				repo.EXPECT().Withdraw(gomock.Any(), testUsername, request).Return(result, model.ErrInsufficientFunds)
			},
			wantErr: model.ErrInsufficientFunds,
		},
		{
			name: "Valid order number, negative amount → error",
			args: args{username: testUsername, withdrawal: model.WithdrawalRequest{
				OrderNumber: testValidOrderNumber,
				Amount:      testNegativeOrderAmount,
			}},
			setup:   func(repo *mocks.MockRepository) {},
			wantErr: model.ErrNonPositiveAmount,
		},
		{
			name: "Valid order number, zero amount → error",
			args: args{username: testUsername, withdrawal: model.WithdrawalRequest{
				OrderNumber: testValidOrderNumber,
				Amount:      testZeroOrderAmount,
			}},
			setup:   func(repo *mocks.MockRepository) {},
			wantErr: model.ErrNonPositiveAmount,
		},
		{
			name: "Withdrawal already exists, same owner → error",
			args: args{username: testUsername, withdrawal: model.WithdrawalRequest{
				OrderNumber: testValidOrderNumber,
				Amount:      testPositiveOrderAmount,
			}},
			setup: func(repo *mocks.MockRepository) {
				request := model.WithdrawalRequest{OrderNumber: testValidOrderNumber, Amount: testPositiveOrderAmount}
				result := model.WithdrawalRegistrationResult{Inserted: false, Owner: testUsername}
				repo.EXPECT().Withdraw(gomock.Any(), testUsername, request).Return(result, nil)
			},
			wantErr: model.ErrWithdrawalExistsSameUser,
		},
		{
			name: "Withdrawal already exists, different owner → error",
			args: args{username: testUsername, withdrawal: model.WithdrawalRequest{
				OrderNumber: testValidOrderNumber,
				Amount:      testPositiveOrderAmount,
			}},
			setup: func(repo *mocks.MockRepository) {
				request := model.WithdrawalRequest{OrderNumber: testValidOrderNumber, Amount: testPositiveOrderAmount}
				result := model.WithdrawalRegistrationResult{Inserted: false, Owner: testAnotherUsername}
				repo.EXPECT().Withdraw(gomock.Any(), testUsername, request).Return(result, nil)
			},
			wantErr: model.ErrWithdrawalExistsOtherUser,
		},
		{
			name: "Withdrawal failed → error",
			args: args{username: testUsername, withdrawal: model.WithdrawalRequest{
				OrderNumber: testValidOrderNumber,
				Amount:      testPositiveOrderAmount,
			}},
			setup: func(repo *mocks.MockRepository) {
				request := model.WithdrawalRequest{OrderNumber: testValidOrderNumber, Amount: testPositiveOrderAmount}
				result := model.WithdrawalRegistrationResult{Inserted: false, Owner: testEmptyUsername}
				repo.EXPECT().Withdraw(gomock.Any(), testUsername, request).Return(result, errTestWithdraw)
			},
			wantErr: errTestWithdraw,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			authMgr := mocks.NewMockAuthManager(ctrl)
			repo := mocks.NewMockRepository(ctrl)
			tt.setup(repo)

			svc := NewLoyaltyService(repo, authMgr, zap.NewNop())
			err := svc.Withdraw(t.Context(), tt.args.username, tt.args.withdrawal)

			if tt.wantErr != nil && !errors.Is(err, tt.wantErr) {
				t.Errorf("Withdraw() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr == nil && !errors.Is(err, tt.wantErr) {
				t.Errorf("Withdraw() unexpected error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestLoyaltyService_ListWithdrawals(t *testing.T) {
	tests := []struct {
		name     string
		username string
		setup    func(repo *mocks.MockRepository)
		want     []model.WithdrawalResponse
		wantErr  error
	}{
		{
			name:     "User has three withdrawals → ok",
			username: testUsername,
			setup: func(repo *mocks.MockRepository) {
				repo.EXPECT().ListWithdrawals(gomock.Any(), testUsername).Return(testNonEmptyWithdrawals, nil)
			},
			want:    expectedNonEmptyWithdrawals,
			wantErr: nil,
		},
		{
			name:     "User has zero withdrawals → ok",
			username: testUsername,
			setup: func(repo *mocks.MockRepository) {
				repo.EXPECT().ListWithdrawals(gomock.Any(), testUsername).Return(testEmptyWithdrawals, nil)
			},
			want:    testEmptyWithdrawals,
			wantErr: nil,
		},
		{
			name:     "List withdrawals failed → error",
			username: testUsername,
			setup: func(repo *mocks.MockRepository) {
				repo.EXPECT().ListWithdrawals(gomock.Any(), testUsername).Return(testEmptyWithdrawals, errTestListWithdrawals)
			},
			want:    testEmptyWithdrawals,
			wantErr: errTestListWithdrawals,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			authMgr := mocks.NewMockAuthManager(ctrl)
			repo := mocks.NewMockRepository(ctrl)
			tt.setup(repo)

			svc := NewLoyaltyService(repo, authMgr, zap.NewNop())
			gotWithdrawals, err := svc.ListWithdrawals(t.Context(), tt.username)

			if tt.wantErr != nil && !errors.Is(err, tt.wantErr) {
				t.Errorf("ListWithdrawals() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr == nil && !errors.Is(err, tt.wantErr) {
				t.Errorf("ListWithdrawals() unexpected error = %v, wantErr %v", err, tt.wantErr)
			}
			if !reflect.DeepEqual(gotWithdrawals, tt.want) {
				t.Errorf("ListWithdrawals() = %v, want %v", gotWithdrawals, tt.want)
			}
		})
	}
}
