package handler

import (
	"errors"
	"math"
	"slices"
	"time"

	"github.com/nikitaw13/gophermart/internal/model"
)

const (
	testUsername           = "GopherUsername"
	testPassword           = "GopherPassword"
	testToken              = "GopherToken"
	testInvalidToken       = "GopherInvalidToken"
	testPassword73Chars    = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!@#$%^&*()+"
	testValidOrderNumber   = "4368910403255712"
	testInvalidOrderNumber = "123"
	testValidAmount        = 500.00
	testInvalidAmount      = -500.00
)

var (
	errGetBalanceFailed      = errors.New("get balance failed")
	errRegisterUserFailed    = errors.New("register user failed")
	errLoginUserFailed       = errors.New("login user failed")
	errLoadOrderFailed       = errors.New("load order failed")
	errListOrdersFailed      = errors.New("list orders failed")
	errWithdrawFailed        = errors.New("withdraw failed")
	errListWithdrawalsFailed = errors.New("list withdrawals failed")
	errBadToken              = errors.New("bad token")
)

var (
	testFixedTime = time.Now()

	testListOrdersResponse = []model.OrderResponse{
		{
			OrderNumber:  testValidOrderNumber,
			Status:       model.StatusProcessed,
			Amount:       100.00,
			UploadedTime: testFixedTime,
		},
		{
			OrderNumber:  testValidOrderNumber,
			Status:       model.StatusProcessing,
			Amount:       200.00,
			UploadedTime: testFixedTime,
		},
		{
			OrderNumber:  testValidOrderNumber,
			Status:       model.StatusNew,
			Amount:       300.00,
			UploadedTime: testFixedTime,
		},
	}
	testWantListOrdersResponse = slices.Clone(testListOrdersResponse)

	testInvalidListOrdersResponse = []model.OrderResponse{
		{
			Amount: math.NaN(),
		},
	}

	testListWithdrawalsResponse = []model.WithdrawalResponse{
		{
			OrderNumber:   testValidOrderNumber,
			Amount:        100.00,
			ProcessedTime: testFixedTime,
		},
		{
			OrderNumber:   testValidOrderNumber,
			Amount:        200.00,
			ProcessedTime: testFixedTime,
		},
		{
			OrderNumber:   testValidOrderNumber,
			Amount:        300.00,
			ProcessedTime: testFixedTime,
		},
	}

	testWantListWithdrawalsResponse = slices.Clone(testListWithdrawalsResponse)

	testInvalidListWithdrawalsResponse = []model.WithdrawalResponse{
		{
			Amount: math.NaN(),
		},
	}

	testGetBalanceResponse = model.BalanceResponse{
		Current:   500.00,
		Withdrawn: 1000.00,
	}

	testWantGetBalanceResponse = testGetBalanceResponse

	testInvalidGetBalanceResponse = model.BalanceResponse{
		Current: math.NaN(),
	}
)
