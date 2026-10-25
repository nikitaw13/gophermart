package storage

import (
	"context"
	"errors"
	"testing"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"go.uber.org/zap"
)

func TestWithRetries(t *testing.T) {
	ps := NewPostgresStorage(
		nil,
		testDatabaseTimeouts,
		NewPostgresErrorClassifier(),
		zap.NewNop(),
	)

	t.Run("success on first attempt", func(t *testing.T) {
		calls := 0
		err := ps.withRetries(t.Context(), func() error {
			calls++
			return nil
		})
		if err != nil {
			t.Fatalf("withRetries() error = %v, want nil", err)
		}
		if calls != 1 {
			t.Fatalf("operation called %d times, want 1", calls)
		}
	})

	t.Run("non-retriable fails fast", func(t *testing.T) {
		sentinel := errors.New("boom")
		calls := 0
		err := ps.withRetries(t.Context(), func() error {
			calls++
			return sentinel
		})
		if !errors.Is(err, sentinel) {
			t.Fatalf("withRetries() error = %v, want the original error", err)
		}
		if calls != 1 {
			t.Fatalf("operation called %d times, want 1", calls)
		}
	})

	t.Run("retriable then success", func(t *testing.T) {
		calls := 0
		err := ps.withRetries(t.Context(), func() error {
			calls++
			if calls == 1 {
				return &pgconn.PgError{Code: pgerrcode.ConnectionException}
			}
			return nil
		})
		if err != nil {
			t.Fatalf("withRetries() error = %v, want nil", err)
		}
		if calls != 2 {
			t.Fatalf("operation called %d times, want 2", calls)
		}
	})

	t.Run("retriable until exhaustion", func(t *testing.T) {
		pgErr := &pgconn.PgError{Code: pgerrcode.ConnectionException}
		calls := 0
		err := ps.withRetries(t.Context(), func() error {
			calls++
			return pgErr
		})
		if !errors.Is(err, pgErr) {
			t.Fatalf("withRetries() error = %v, want wrapped pgErr", err)
		}
		wantCalls := len(ps.timeouts) + 1
		if calls != wantCalls {
			t.Fatalf("operation called %d times, want %d", calls, wantCalls)
		}
	})

	t.Run("context cancelled during backoff", func(t *testing.T) {
		ctx, cancel := context.WithCancel(t.Context())
		defer cancel()

		calls := 0
		err := ps.withRetries(ctx, func() error {
			calls++
			cancel()
			return &pgconn.PgError{Code: pgerrcode.ConnectionException}
		})
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("withRetries() error = %v, want context.Canceled", err)
		}
		if calls != 1 {
			t.Fatalf("operation called %d times, want 1", calls)
		}
	})
}
