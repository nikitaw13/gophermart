package service

import (
	"errors"
	"time"

	"github.com/nikitaw13/gophermart/internal/model"
)

const (
	testUsername                   = "GopherUsername"
	testExistingUsername           = "GopherExistingUsername"
	testEmptyUsername              = ""
	testAnotherUsername            = "GopherAnotherUsername"
	testToken                      = "GopherToken"
	testEmptyToken                 = ""
	testPassword                   = "GopherPassword"
	testPassword72Symbols          = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!@#$%^&*()"
	testPassword73Symbols          = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!@#$%^&*()_"
	testInvalidHash                = "12345"
	testEmptyOrderNumber           = ""
	testValidOrderNumber           = "79927398713"
	testAnotherValidOrderNumber    = "5412751234123456"
	testYetAnotherValidOrderNumber = "4000123456789010"
	testInvalidOrderNumber         = "123"
	testPositiveAmount             = 500
	testNegativeAmount             = -500
	testZeroAmount                 = 0
)

var (
	errIssueToken            = errors.New("issue token failed")
	errLoadOrder             = errors.New("load order failed")
	errListOrders            = errors.New("list orders failed")
	errGetBalance            = errors.New("get balance failed")
	errListWithdrawals       = errors.New("list withdrawals failed")
	errWithdraw              = errors.New("withdraw failed")
	testFixedTimestamp       = time.Date(1993, time.August, 22, 0, 0, 0, 0, time.Local)
	testEmptyOrderResponses  = []model.OrderResponse{}
	testSingleOrderResponses = []model.OrderResponse{
		{
			OrderNumber:  testValidOrderNumber,
			Status:       model.StatusProcessed,
			Amount:       testPositiveAmount,
			UploadedTime: testFixedTimestamp,
		},
	}
	testThreeOrderResponses = []model.OrderResponse{
		{
			OrderNumber:  testValidOrderNumber,
			Status:       model.StatusProcessed,
			Amount:       testPositiveAmount,
			UploadedTime: testFixedTimestamp,
		},
		{
			OrderNumber:  testAnotherValidOrderNumber,
			Status:       model.StatusProcessing,
			Amount:       testPositiveAmount,
			UploadedTime: testFixedTimestamp,
		},
		{
			OrderNumber:  testYetAnotherValidOrderNumber,
			Status:       model.StatusNew,
			Amount:       testPositiveAmount,
			UploadedTime: testFixedTimestamp,
		},
	}
	testWantThreeOrderResponses = []model.OrderResponse{
		{
			OrderNumber:  testValidOrderNumber,
			Status:       model.StatusProcessed,
			Amount:       testPositiveAmount,
			UploadedTime: testFixedTimestamp,
		},
		{
			OrderNumber:  testAnotherValidOrderNumber,
			Status:       model.StatusProcessing,
			Amount:       testPositiveAmount,
			UploadedTime: testFixedTimestamp,
		},
		{
			OrderNumber:  testYetAnotherValidOrderNumber,
			Status:       model.StatusNew,
			Amount:       testPositiveAmount,
			UploadedTime: testFixedTimestamp,
		},
	}
	testEmptyBalance   = model.BalanceResponse{}
	testNonZeroBalance = model.BalanceResponse{
		Current:   500.00,
		Withdrawn: 1500.00,
	}
	testWantNonZeroBalance = model.BalanceResponse{
		Current:   500.00,
		Withdrawn: 1500.00,
	}

	testEmptyWithdrawals    = []model.WithdrawalResponse{}
	testNonEmptyWithdrawals = []model.WithdrawalResponse{
		{
			OrderNumber:   testValidOrderNumber,
			Amount:        testPositiveAmount,
			ProcessedTime: testFixedTimestamp,
		},
		{
			OrderNumber:   testAnotherValidOrderNumber,
			Amount:        testPositiveAmount,
			ProcessedTime: testFixedTimestamp,
		},
		{
			OrderNumber:   testYetAnotherValidOrderNumber,
			Amount:        testPositiveAmount,
			ProcessedTime: testFixedTimestamp,
		},
	}
	testWantNonEmptyWithdrawals = []model.WithdrawalResponse{
		{
			OrderNumber:   testValidOrderNumber,
			Amount:        testPositiveAmount,
			ProcessedTime: testFixedTimestamp,
		},
		{
			OrderNumber:   testAnotherValidOrderNumber,
			Amount:        testPositiveAmount,
			ProcessedTime: testFixedTimestamp,
		},
		{
			OrderNumber:   testYetAnotherValidOrderNumber,
			Amount:        testPositiveAmount,
			ProcessedTime: testFixedTimestamp,
		},
	}
)
