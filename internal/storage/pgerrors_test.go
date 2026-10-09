package storage

import (
	"errors"
	"fmt"
	"testing"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
)

var errUnknownTest = errors.New("GopherTestUnknownError")

const unknownTestPgError = "99999"

func TestPostgresErrorClassifier_Classify(t *testing.T) {
	type args struct {
		err error
	}
	tests := []struct {
		name string
		args args
		want PostgresErrorClassification
	}{
		{
			name: "nil → NonRetriable",
			args: args{err: nil},
			want: NonRetriable,
		},
		{
			name: "Unknown error → NonRetriable",
			args: args{err: errUnknownTest},
			want: NonRetriable,
		},
		{
			name: "Unknown PgError → NonRetriable",
			args: args{err: &pgconn.PgError{Code: unknownTestPgError}},
			want: NonRetriable,
		},
		{
			name: "Class 08 — ConnectionFailure → Retriable",
			args: args{err: &pgconn.PgError{Code: pgerrcode.ConnectionFailure}},
			want: Retriable,
		},
		{
			name: "Class 08 — ConnectionDoesNotExist → Retriable",
			args: args{err: &pgconn.PgError{Code: pgerrcode.ConnectionDoesNotExist}},
			want: Retriable,
		},
		{
			name: "Class 08 — SQLClientUnableToEstablishSQLConnection → NonRetriable",
			args: args{err: &pgconn.PgError{Code: pgerrcode.SQLClientUnableToEstablishSQLConnection}},
			want: NonRetriable,
		},
		{
			name: "Class 08 — ConnectionException → Retriable",
			args: args{err: &pgconn.PgError{Code: pgerrcode.ConnectionException}},
			want: Retriable,
		},
		{
			name: "Wrapped Class 08 → Retriable",
			args: args{err: fmt.Errorf("query failed: %w", &pgconn.PgError{Code: pgerrcode.ConnectionException})},
			want: Retriable,
		},
		{
			name: "Class 40 — SerializationFailure → Retriable",
			args: args{err: &pgconn.PgError{Code: pgerrcode.SerializationFailure}},
			want: Retriable,
		},
		{
			name: "Class 40 — DeadlockDetected → Retriable",
			args: args{err: &pgconn.PgError{Code: pgerrcode.DeadlockDetected}},
			want: Retriable,
		},
		{
			name: "Class 40 — TransactionRollback → Retriable",
			args: args{err: &pgconn.PgError{Code: pgerrcode.TransactionRollback}},
			want: Retriable,
		},
		{
			name: "Class 57 — CannotConnectNow → Retriable",
			args: args{err: &pgconn.PgError{Code: pgerrcode.CannotConnectNow}},
			want: Retriable,
		},
		{
			name: "Class 22 — DataException → NonRetriable",
			args: args{err: &pgconn.PgError{Code: pgerrcode.DataException}},
			want: NonRetriable,
		},
		{
			name: "Class 22 — InvalidTextRepresentation → NonRetriable",
			args: args{err: &pgconn.PgError{Code: pgerrcode.InvalidTextRepresentation}},
			want: NonRetriable,
		},
		{
			name: "Class 23 — IntegrityConstraintViolation → NonRetriable",
			args: args{err: &pgconn.PgError{Code: pgerrcode.IntegrityConstraintViolation}},
			want: NonRetriable,
		},
		{
			name: "Class 42 — SyntaxErrorOrAccessRuleViolation → NonRetriable",
			args: args{err: &pgconn.PgError{Code: pgerrcode.SyntaxErrorOrAccessRuleViolation}},
			want: NonRetriable,
		},
		{
			name: "Wrapped Class 42 → NonRetriable",
			args: args{err: fmt.Errorf("query failed: %w", &pgconn.PgError{Code: pgerrcode.SyntaxErrorOrAccessRuleViolation})},
			want: NonRetriable,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			classifier := NewPostgresErrorClassifier()
			if got := classifier.Classify(tt.args.err); got != tt.want {
				t.Errorf("PostgresErrorClassifier.Classify() = %v, want %v", got, tt.want)
			}
		})
	}
}
