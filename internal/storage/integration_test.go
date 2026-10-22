package storage

import (
	"errors"
	"testing"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/nikitaw13/gophermart/internal/model"
	"go.uber.org/zap"
)

func TestUserFlow(t *testing.T) {
	resetDB(t)
	ps := NewPostgresStorage(
		testPool,
		testDatabaseTimeouts,
		NewPostgresErrorClassifier(),
		zap.NewNop(),
	)

	t.Run("create user", func(t *testing.T) {
		err := ps.CreateUser(t.Context(), testUsername, testPasswordHash)
		if err != nil {
			t.Errorf("CreateUser() error = %v", err)
		}

		err = ps.CreateUser(t.Context(), testUsername, testPasswordHash)
		if !errors.Is(err, model.ErrUserAlreadyExists) {
			t.Errorf("CreateUser() error = %v, want ErrUserAlreadyExists", err)
		}
	})
	t.Run("get password hash", func(t *testing.T) {
		_, err := ps.GetPasswordHash(t.Context(), testAnotherUsername)
		if !errors.Is(err, model.ErrUserNotFound) {
			t.Errorf("GetPasswordHash() error = %v, want ErrUserNotFound", err)
		}

		gotHash, err := ps.GetPasswordHash(t.Context(), testUsername)
		if err != nil {
			t.Errorf("GetPasswordHash() error = %v", err)
		}
		if gotHash != testPasswordHash {
			t.Errorf("GetPasswordHash() = %s, want %s", gotHash, testPasswordHash)
		}
	})
}

func TestOrderFlow(t *testing.T) {
	resetDB(t)
	ps := NewPostgresStorage(
		testPool,
		testDatabaseTimeouts,
		NewPostgresErrorClassifier(),
		zap.NewNop(),
	)

	tag, err := ps.db.Exec(t.Context(), prepareOrderFlow,
		testUsername, testPassword,
		testYetAnotherUsername, testPassword,
	)
	if err != nil {
		t.Fatalf("seed users: %v", err)
	}
	if n := tag.RowsAffected(); n != 2 {
		t.Fatalf("seed users: affected %d rows, want 2", n)
	}

	t.Run("register order", func(t *testing.T) {
		_, err := ps.RegisterOrder(t.Context(), "", testAnotherUsername)

		var pgErr *pgconn.PgError
		if !errors.As(err, &pgErr) || pgErr.Code != pgerrcode.NotNullViolation {
			t.Fatalf("RegisterOrder() error = %v, want NotNullViolation", err)
		}

		if err == nil {
			t.Fatalf("RegisterOrder() error = %v, want NotNullViolation", err)
		}

		result, err := ps.RegisterOrder(t.Context(), testOrderNumber, testUsername)
		if err != nil {
			t.Fatalf("RegisterOrder() error = %v", err)
		}

		if !result.Inserted {
			t.Fatalf("RegisterOrder().Inserted = %t, want true", result.Inserted)
		}
		if result.Owner != testUsername {
			t.Fatalf("RegisterOrder().Owner = %s, want %s", result.Owner, testUsername)
		}

		// The same user
		result, err = ps.RegisterOrder(t.Context(), testOrderNumber, testUsername)
		if err != nil {
			t.Fatalf("RegisterOrder() error = %v", err)
		}
		if result.Inserted {
			t.Fatalf("RegisterOrder().Inserted = %t, want false", result.Inserted)
		}
		if result.Owner != testUsername {
			t.Fatalf("RegisterOrder().Owner = %s, want %s", result.Owner, testUsername)
		}

		// Another user
		result, err = ps.RegisterOrder(t.Context(), testOrderNumber, testYetAnotherUsername)
		if err != nil {
			t.Fatalf("RegisterOrder() error = %v", err)
		}
		if result.Inserted {
			t.Fatalf("RegisterOrder().Inserted = %t, want false", result.Inserted)
		}
		if result.Owner != testUsername {
			t.Fatalf("RegisterOrder().Owner = %s, want %s", result.Owner, testUsername)
		}

		// List orders
		gotOrders, err := ps.ListOrders(t.Context(), testUsername)
		if err != nil {
			t.Fatalf("ListOrders() error = %v", err)
		}
		if len(gotOrders) != 1 {
			t.Fatalf("ListOrders() got %d orders, want %d", len(gotOrders), 1)
		}

		if gotOrders[0].OrderNumber != testWantOrderResponse.OrderNumber ||
			gotOrders[0].Amount != testWantOrderResponse.Amount ||
			gotOrders[0].Status != testWantOrderResponse.Status {
			t.Errorf("ListOrders() = %v, want %v", gotOrders[0], testWantOrderResponse)
		}

		gotOrders, err = ps.ListOrders(t.Context(), testYetAnotherUsername)
		if err != nil {
			t.Errorf("ListOrders() error = %v", err)
		}
		if len(gotOrders) != 0 {
			t.Errorf("ListOrders() got %d orders, want %d", len(gotOrders), 0)
		}
	})
}

func TestWithdrawalFlow(t *testing.T) {
	resetDB(t)
	ps := NewPostgresStorage(
		testPool,
		testDatabaseTimeouts,
		NewPostgresErrorClassifier(),
		zap.NewNop(),
	)

	var userID int64
	err := ps.db.QueryRow(t.Context(), prepareWithdrawalFlow,
		testUsername,
		testPassword,
		testOrderNumber,
		testAccrualAmount,
	).Scan(&userID)

	if err != nil {
		t.Fatalf("prepareWithdrawalFlow() error = %v", err)
	}

	t.Run("withdraw", func(t *testing.T) {
		_, err := ps.GetBalance(t.Context(), testAnotherUsername)
		if !errors.Is(err, pgx.ErrNoRows) {
			t.Fatalf("GetBalance() error = %v, want ErrNoRows", err)
		}

		balanceBefore, err := ps.GetBalance(t.Context(), testUsername)
		if err != nil {
			t.Fatalf("GetBalance() error = %v, want nil", err)
		}

		// Success
		result, err := ps.Withdraw(t.Context(), testUsername, testWithdrawalRequest)
		if err != nil {
			t.Fatalf("Withdraw() error = %v", err)
		}
		if !result.Inserted {
			t.Fatalf("Withdraw().Inserted = %t, want true", result.Inserted)
		}
		if result.Owner != testUsername {
			t.Fatalf("Withdraw().Owner = %s, want %s", result.Owner, testUsername)
		}

		// The real transaction check: the balance dropped by exactly the withdrawn amount.
		balanceAfter, err := ps.GetBalance(t.Context(), testUsername)
		if err != nil {
			t.Fatalf("GetBalance() error = %v, want nil", err)
		}
		if balanceAfter.Withdrawn != testWithdrawalRequest.Amount {
			t.Fatalf("GetBalance().Withdrawn = %v, want %v", balanceAfter.Withdrawn, testWithdrawalRequest.Amount)
		}
		if got := balanceBefore.Current - balanceAfter.Current; got != testWithdrawalRequest.Amount {
			t.Fatalf("balance decreased by %v, want %v", got, testWithdrawalRequest.Amount)
		}

		_, err = ps.Withdraw(t.Context(), testUsername, testWithdrawalRequestInsufficientFunds)
		if !errors.Is(err, model.ErrInsufficientFunds) {
			t.Fatalf("Withdraw() error = %v, want ErrInsufficientFunds", err)
		}

		// The failed withdrawal must not have touched the balance.
		balanceAfterFailed, err := ps.GetBalance(t.Context(), testUsername)
		if err != nil {
			t.Fatalf("GetBalance() error = %v, want nil", err)
		}
		if balanceAfterFailed != balanceAfter {
			t.Fatalf("balance changed after a failed withdrawal: got %v, want %v", balanceAfterFailed, balanceAfter)
		}

		// Already withdrawn
		result, err = ps.Withdraw(t.Context(), testUsername, testWithdrawalRequest)
		if err != nil {
			t.Fatalf("Withdraw() error = %v", err)
		}
		if result.Inserted {
			t.Fatalf("Withdraw().Inserted = %t, want false", result.Inserted)
		}
		if result.Owner != testUsername {
			t.Fatalf("Withdraw().Owner = %s, want %s", result.Owner, testUsername)
		}

		gotWithdrawals, err := ps.ListWithdrawals(t.Context(), testUsername)
		if err != nil {
			t.Errorf("ListWithdrawals() error = %v", err)
		}
		if len(gotWithdrawals) != 1 {
			t.Errorf("ListWithdrawals() got %d withdrawals, want %d", len(gotWithdrawals), 1)
		}
		if gotWithdrawals[0].OrderNumber != testWantListWithdrawalsResponse.OrderNumber ||
			gotWithdrawals[0].Amount != testWantListWithdrawalsResponse.Amount {
			t.Errorf("ListWithdrawals() = %v, want %v", gotWithdrawals[0], testWantListWithdrawalsResponse)
		}
	})
}

func TestAccrualFlow(t *testing.T) {
	resetDB(t)
	ps := NewPostgresStorage(
		testPool,
		testDatabaseTimeouts,
		NewPostgresErrorClassifier(),
		zap.NewNop(),
	)

	_, err := ps.db.Exec(t.Context(), prepareAccrualFlow, testUsername, testPassword)
	if err != nil {
		t.Fatalf("seed user: %v", err)
	}

	t.Run("fetch new orders on empty db", func(t *testing.T) {
		got, err := ps.FetchNewOrders(t.Context())
		if err != nil {
			t.Fatalf("FetchNewOrders() error = %v", err)
		}
		if len(got) != 0 {
			t.Fatalf("FetchNewOrders() got %d orders, want %d", len(got), 0)
		}
	})

	t.Run("register order", func(t *testing.T) {
		result, err := ps.RegisterOrder(t.Context(), testOrderNumber, testUsername)
		if err != nil {
			t.Fatalf("RegisterOrder() error = %v", err)
		}
		if !result.Inserted {
			t.Fatalf("RegisterOrder().Inserted = %t, want true", result.Inserted)
		}
		if result.Owner != testUsername {
			t.Fatalf("RegisterOrder().Owner = %s, want %s", result.Owner, testUsername)
		}

		gotOrders, err := ps.ListOrders(t.Context(), testUsername)
		if err != nil {
			t.Fatalf("ListOrders() error = %v", err)
		}
		if len(gotOrders) != 1 {
			t.Fatalf("ListOrders() got %d orders, want %d", len(gotOrders), 1)
		}
		if gotOrders[0].OrderNumber != testWantOrderResponse.OrderNumber ||
			gotOrders[0].Status != testWantOrderResponse.Status ||
			gotOrders[0].Amount != testWantOrderResponse.Amount {
			t.Errorf("ListOrders() = %v, want %v", gotOrders[0], testWantOrderResponse)
		}
	})

	t.Run("fetch new orders", func(t *testing.T) {
		got, err := ps.FetchNewOrders(t.Context())
		if err != nil {
			t.Fatalf("FetchNewOrders() error = %v", err)
		}
		if len(got) != 1 {
			t.Fatalf("FetchNewOrders() got %d orders, want %d", len(got), 1)
		}
		if got[0] != testOrderNumber {
			t.Fatalf("FetchNewOrders()[0] = %s, want %s", got[0], testOrderNumber)
		}
	})

	t.Run("update order", func(t *testing.T) {
		err := ps.UpdateOrder(t.Context(), testOrderNumber, model.StatusProcessed, testAccrualAmount)
		if err != nil {
			t.Fatalf("UpdateOrder() error = %v", err)
		}

		gotOrders, err := ps.ListOrders(t.Context(), testUsername)
		if err != nil {
			t.Fatalf("ListOrders() error = %v", err)
		}
		if len(gotOrders) != 1 {
			t.Fatalf("ListOrders() got %d orders, want %d", len(gotOrders), 1)
		}
		if gotOrders[0].Status != model.StatusProcessed {
			t.Errorf("ListOrders().Status = %s, want %s", gotOrders[0].Status, model.StatusProcessed)
		}
		if gotOrders[0].Amount != testAccrualAmount {
			t.Errorf("ListOrders().Amount = %v, want %v", gotOrders[0].Amount, testAccrualAmount)
		}

		balance, err := ps.GetBalance(t.Context(), testUsername)
		if err != nil {
			t.Fatalf("GetBalance() error = %v", err)
		}
		if balance.Current != testAccrualAmount {
			t.Errorf("GetBalance().Current = %v, want %v", balance.Current, testAccrualAmount)
		}
		if balance.Withdrawn != 0 {
			t.Errorf("GetBalance().Withdrawn = %v, want %v", balance.Withdrawn, 0)
		}
	})

	t.Run("fetch new orders after processing", func(t *testing.T) {
		got, err := ps.FetchNewOrders(t.Context())
		if err != nil {
			t.Fatalf("FetchNewOrders() error = %v", err)
		}
		if len(got) != 0 {
			t.Fatalf("FetchNewOrders() got %d orders, want %d", len(got), 0)
		}
	})
}
