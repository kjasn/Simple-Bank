package db

import (
	"context"
	"database/sql"
)

type VerifyEmailTxParams struct {
	VerifyEmailId int64
	SecretCode string
}

type VerifyEmailTxResult struct {
	User User
	VerifyEmail VerifyEmail
}

func (SQLStore *SQLStore) VerifyEmailTx(ctx context.Context, arg VerifyEmailTxParams) (VerifyEmailTxResult, error) {
	var result VerifyEmailTxResult

	err := SQLStore.execTx(ctx, func(q *Queries) error {
		var err error
		result.VerifyEmail, err = q.UpdateVerifyEmail(ctx, UpdateVerifyEmailParams{
			ID: arg.VerifyEmailId,
			SecretCode: arg.SecretCode,
		})
		if err != nil {
			return err
		}

		// both secret code and verify email id match
		result.User, err = q.UpdateUser(ctx, UpdateUserParams{
			Username: result.VerifyEmail.Username,
			IsEmailVerified: sql.NullBool{
				Bool: true,
				Valid: true,
			},
		})
		if err != nil {
			return err
		}

		return nil
	})
	return result, err
}

