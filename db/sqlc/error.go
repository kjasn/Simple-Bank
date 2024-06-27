package db

import (
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

var ErrRecordNotFound = pgx.ErrNoRows

const (
	UniqueViolation = "23505"
	ForeignKeyViolation = "23503"
)

var ErrUniqueViolation = &pgconn.PgError{
	Code: UniqueViolation,
}

var ErrForeignKeyViolation= &pgconn.PgError{
	Code: ForeignKeyViolation,
}

func ErrorCode(err error) string {
	var pgErr *pgconn.PgError

	if errors.As(err, &pgErr) {
		return pgErr.Code
	}
	return ""
}