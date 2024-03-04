package db

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/kjasn/simple-bank/utils"
	"github.com/stretchr/testify/require"
)

// make sure each unite test independent
func createRandomAccount(t *testing.T) Account {
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

	return account
}

func TestCreateAccount(t *testing.T) {
	createRandomAccount(t)
}

func TestGetAccount(t *testing.T) {
	account1 := createRandomAccount(t)
	account2, err := testQueries.GetAccount(context.Background(), account1.ID)

	require.NoError(t, err)
	require.NotEmpty(t, account2)

	require.Equal(t, account1.ID, account2.ID)
	require.Equal(t, account1.Owner, account2.Owner)
	require.Equal(t, account1.Banlance, account2.Banlance)
	require.Equal(t, account1.Currency, account2.Currency)
	require.WithinDuration(t, account1.CreatedAt, account2.CreatedAt, time.Second)	// set the delta time 1s
}

func TestUpdateAccount(t *testing.T) {
	oldAccount := createRandomAccount(t)
	arg := UpdateAccountParams {
		ID: oldAccount.ID,
		Banlance: utils.RandomMoney(),
	}

	newAccount, err := testQueries.UpdateAccount(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, newAccount)
	require.Equal(t, oldAccount.ID, newAccount.ID)
	require.Equal(t, arg.Banlance, newAccount.Banlance)
	require.Equal(t, oldAccount.Currency, newAccount.Currency)
	require.WithinDuration(t, oldAccount.CreatedAt, newAccount.CreatedAt, time.Second)	// set the delta time 1s
}

func TestDeleteAccount(t *testing.T) {
	account1 := createRandomAccount(t)	
	err := testQueries.DeleteAccount(context.Background(), account1.ID)
	require.NoError(t, err)

	account2, err := testQueries.GetAccount(context.Background(), account1.ID)
	require.Error(t, err)
	require.EqualError(t, err, sql.ErrNoRows.Error())
	require.Empty(t, account2)
}


func TestListAccounts(t *testing.T) {
	for i := 0; i < 20; i++ {
		createRandomAccount(t)
	}

	arg := ListAccountsParams {
		Limit: int32(utils.RandomInt(5, 10)),
		Offset: int32(utils.RandomInt(5, 10)),
	}	

	accounts, err := testQueries.ListAccounts(context.Background(), arg)
	require.NoError(t, err)
	require.Len(t, accounts, int(arg.Limit))
	for i := 0; i < len(accounts); i++ {
		require.NotEmpty(t, accounts[i])
	}
}