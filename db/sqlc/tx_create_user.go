package db

import (
	"context"
)

type CreateUserTxParams struct {
	CreateUserParams
	AfterCreated func(user User) error
}

type CreateUserTxResult struct {
	User User
}

func (SQLStore *SQLStore) CreateUserTx(ctx context.Context, arg CreateUserTxParams) (CreateUserTxResult, error) {
	var result CreateUserTxResult

	err := SQLStore.execTx(ctx, func(q *Queries) error {
		var err error
		result.User, err = q.CreateUser(ctx, arg.CreateUserParams)
		if err != nil {
			return err
		}

		return arg.AfterCreated(result.User)
	})

	return result, err
}

