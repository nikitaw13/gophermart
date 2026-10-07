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
			name: "Class 08 — connection exception → Retriable",
			args: args{err: &pgconn.PgError{Code: pgerrcode.ConnectionException}},
			want: Retriable,
		},
		{
			name: "Wrapped Class 08 → Retriable",
			args: args{err: fmt.Errorf("query failed: %w", &pgconn.PgError{Code: pgerrcode.ConnectionException})},
			want: Retriable,
		},
		{
			name: "Class 40 — transaction rollback → Retriable",
			args: args{err: &pgconn.PgError{Code: pgerrcode.TransactionRollback}},
			want: Retriable,
		},
		{
			name: "Class 57 — operator intervention → Retriable",
			args: args{err: &pgconn.PgError{Code: pgerrcode.CannotConnectNow}},
			want: Retriable,
		},
		{
			name: "Class 22 — data exception → NonRetriable",
			args: args{err: &pgconn.PgError{Code: pgerrcode.DataException}},
			want: NonRetriable,
		},
		{
			name: "Class 23 — integrity constraint violation → NonRetriable",
			args: args{err: &pgconn.PgError{Code: pgerrcode.IntegrityConstraintViolation}},
			want: NonRetriable,
		},
		{
			name: "Class 42 — syntax error or access rule violation → NonRetriable",
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
			classifier := NewPostgresErrorClassifier()
			if got := classifier.Classify(tt.args.err); got != tt.want {
				t.Errorf("PostgresErrorClassifier.Classify() = %v, want %v", got, tt.want)
			}
		})
	}
}
