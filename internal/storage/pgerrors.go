package storage

import (
	"errors"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
)

// PostgresErrorClassification tells whether a failed operation can be retried.
type PostgresErrorClassification int

const (
	// NonRetriable means the operation must not be retried.
	NonRetriable PostgresErrorClassification = iota

	// Retriable means the operation can be retried.
	Retriable
)

// PostgresErrorClassifier classifies PostgreSQL errors.
type PostgresErrorClassifier struct{}

// NewPostgresErrorClassifier returns a classifier for PostgreSQL errors.
func NewPostgresErrorClassifier() *PostgresErrorClassifier {
	return &PostgresErrorClassifier{}
}

// Classify returns the classification of the given error.
func (c *PostgresErrorClassifier) Classify(err error) PostgresErrorClassification {
	if err == nil {
		return NonRetriable
	}

	// Try to extract pgconn.PgError from the error chain.
	if pgErr, ok := errors.AsType[*pgconn.PgError](err); ok {
		return ClassifyPgError(pgErr)
	}

	// Unknown errors are treated as non-retriable.
	return NonRetriable
}

// ClassifyPgError classifies a PostgreSQL error code as retriable or non-retriable.
func ClassifyPgError(pgErr *pgconn.PgError) PostgresErrorClassification {
	// PostgreSQL error codes: https://www.postgresql.org/docs/current/errcodes-appendix.html

	switch pgErr.Code {
	// Class 08 — connection exception (Retriable).
	case pgerrcode.ConnectionException,
		pgerrcode.ConnectionDoesNotExist,
		pgerrcode.ConnectionFailure:
		return Retriable

	// Class 40 — transaction rollback.
	case pgerrcode.TransactionRollback, // 40000
		pgerrcode.SerializationFailure, // 40001
		pgerrcode.DeadlockDetected:     // 40P01
		return Retriable

	// Class 57 — operator intervention.
	case pgerrcode.CannotConnectNow: // 57P03
		return Retriable
	}

	// Class 22 — data exception.
	if pgerrcode.IsDataException(pgErr.Code) {
		return NonRetriable
	}
	// Class 23 — integrity constraint violation.
	if pgerrcode.IsIntegrityConstraintViolation(pgErr.Code) {
		return NonRetriable
	}
	// Class 42 — syntax error or access rule violation
	if pgerrcode.IsSyntaxErrororAccessRuleViolation(pgErr.Code) {
		return NonRetriable
	}

	// Unknown error codes are non-retriable.
	return NonRetriable
}
