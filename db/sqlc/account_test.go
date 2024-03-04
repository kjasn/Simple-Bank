package db

import (
	"context"
	"testing"

	"github.com/kjasn/simple-bank/utils"
	"github.com/stretchr/testify/require"
)


func TestCreateAccount(t *testing.T) {
	arg := CreateAccountParams {
		Owner: utils.RandomOwner(),
		Banlance: utils.RandomMoney(),
		Currency: utils.RandomCurrency(),
	}

	account, err := testQueries.CreateAccount(context.Background(), arg);

	// check the return with testify/require
	require.NoError(t, err)
	require.NotEmpty(t, account)
	require.Equal(t, arg.Owner, account.Owner)
	require.Equal(t, arg.Banlance, account.Banlance)
	require.Equal(t, arg.Currency, account.Currency)
	// assert not zero value of its type
	require.NotZero(t, account.ID)
	require.NotZero(t, account.CreatedAt)
}