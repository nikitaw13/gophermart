package handler

import (
	"context"

	"github.com/nikitaw13/gophermart/internal/model"
)

// Service is the business-logic interface consumed by APIHandler.
type Service interface {
	IdentifyUser(token string) (string, error)
	RegisterUser(ctx context.Context, user model.User) (string, error)
	LoginUser(ctx context.Context, user model.User) (string, error)
	LoadOrder(ctx context.Context, orderNumber string, username string) error
	ListOrders(ctx context.Context, username string) ([]model.OrderResponse, error)
	Withdraw(ctx context.Context, username string, withdrawal model.WithdrawalRequest) error
	ListWithdrawals(ctx context.Context, username string) ([]model.WithdrawalResponse, error)
	GetBalance(ctx context.Context, username string) (model.BalanceResponse, error)
}
