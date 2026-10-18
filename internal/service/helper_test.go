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
	testPositiveOrderAmount        = 500
	testNegativeOrderAmount        = -500
	testZeroOrderAmount            = 0
)

var (
	errTestIssueToken        = errors.New("issue token failed")
	errTestLoadOrder         = errors.New("load order failed")
	errTestListOrders        = errors.New("list orders failed")
	errTestGetBalance        = errors.New("get balance failed")
	errTestListWithdrawals   = errors.New("list withdrawals failed")
	errTestWithdraw          = errors.New("withdraw failed")
	testFixedTimestamp       = time.Date(1993, time.August, 22, 0, 0, 0, 0, time.Local)
	testEmptyOrderResponses  = []model.OrderResponse{}
	testSingleOrderResponses = []model.OrderResponse{
		{
			OrderNumber:  testValidOrderNumber,
			Status:       model.StatusProcessed,
			Amount:       testPositiveOrderAmount,
			UploadedTime: testFixedTimestamp,
		},
	}
	testThreeOrderResponses = []model.OrderResponse{
		{
			OrderNumber:  testValidOrderNumber,
			Status:       model.StatusProcessed,
			Amount:       testPositiveOrderAmount,
			UploadedTime: testFixedTimestamp,
		},
		{
			OrderNumber:  testAnotherValidOrderNumber,
			Status:       model.StatusProcessing,
			Amount:       testPositiveOrderAmount,
			UploadedTime: testFixedTimestamp,
		},
		{
			OrderNumber:  testYetAnotherValidOrderNumber,
			Status:       model.StatusNew,
			Amount:       testPositiveOrderAmount,
			UploadedTime: testFixedTimestamp,
		},
	}
	expectedThreeOrderResponses = []model.OrderResponse{
		{
			OrderNumber:  testValidOrderNumber,
			Status:       model.StatusProcessed,
			Amount:       testPositiveOrderAmount,
			UploadedTime: testFixedTimestamp,
		},
		{
			OrderNumber:  testAnotherValidOrderNumber,
			Status:       model.StatusProcessing,
			Amount:       testPositiveOrderAmount,
			UploadedTime: testFixedTimestamp,
		},
		{
			OrderNumber:  testYetAnotherValidOrderNumber,
			Status:       model.StatusNew,
			Amount:       testPositiveOrderAmount,
			UploadedTime: testFixedTimestamp,
		},
	}
	testEmptyBalance   = model.BalanceResponse{}
	testNonZeroBalance = model.BalanceResponse{
		Current:   500.00,
		Withdrawn: 1500.00,
	}
	expectedNonZeroBalance = model.BalanceResponse{
		Current:   500.00,
		Withdrawn: 1500.00,
	}

	testEmptyWithdrawals    = []model.WithdrawalResponse{}
	testNonEmptyWithdrawals = []model.WithdrawalResponse{
		{
			OrderNumber:   testValidOrderNumber,
			Amount:        testPositiveOrderAmount,
			ProcessedTime: testFixedTimestamp,
		},
		{
			OrderNumber:   testAnotherValidOrderNumber,
			Amount:        testPositiveOrderAmount,
			ProcessedTime: testFixedTimestamp,
		},
		{
			OrderNumber:   testYetAnotherValidOrderNumber,
			Amount:        testPositiveOrderAmount,
			ProcessedTime: testFixedTimestamp,
		},
	}
	expectedNonEmptyWithdrawals = []model.WithdrawalResponse{
		{
			OrderNumber:   testValidOrderNumber,
			Amount:        testPositiveOrderAmount,
			ProcessedTime: testFixedTimestamp,
		},
		{
			OrderNumber:   testAnotherValidOrderNumber,
			Amount:        testPositiveOrderAmount,
			ProcessedTime: testFixedTimestamp,
		},
		{
			OrderNumber:   testYetAnotherValidOrderNumber,
			Amount:        testPositiveOrderAmount,
			ProcessedTime: testFixedTimestamp,
		},
	}
)
