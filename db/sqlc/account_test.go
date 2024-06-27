package db

import (
	"context"
	"testing"
	"time"

	"github.com/kjasn/simple-bank/utils"
	"github.com/stretchr/testify/require"
)

// make sure each unit test independent
func createRandomAccount(t *testing.T) Account {
	// create a user first
	user := createRandomUser(t)

	arg := CreateAccountParams {
		Owner: user.Username,	// set foreign key 
		Balance: utils.RandomMoney(),
		Currency: utils.RandomCurrency(),
	}

	account, err := testStore.CreateAccount(context.Background(), arg);

	// check the return with testify/require
	require.NoError(t, err)
	require.NotEmpty(t, account)
	require.Equal(t, arg.Owner, account.Owner)
	require.Equal(t, arg.Balance, account.Balance)
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
	account2, err := testStore.GetAccount(context.Background(), account1.ID)

	require.NoError(t, err)
	require.NotEmpty(t, account2)

	require.Equal(t, account1.ID, account2.ID)
	require.Equal(t, account1.Owner, account2.Owner)
	require.Equal(t, account1.Balance, account2.Balance)
	require.Equal(t, account1.Currency, account2.Currency)
	require.WithinDuration(t, account1.CreatedAt, account2.CreatedAt, time.Second)	// set the delta time 1s
}

func TestUpdateAccount(t *testing.T) {
	oldAccount := createRandomAccount(t)
	arg := UpdateAccountParams {
		ID: oldAccount.ID,
		Balance: utils.RandomMoney(),
	}

	newAccount, err := testStore.UpdateAccount(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, newAccount)
	require.Equal(t, oldAccount.ID, newAccount.ID)
	require.Equal(t, arg.Balance, newAccount.Balance)
	require.Equal(t, oldAccount.Currency, newAccount.Currency)
	require.WithinDuration(t, oldAccount.CreatedAt, newAccount.CreatedAt, time.Second)	// set the delta time 1s
}

func TestDeleteAccount(t *testing.T) {
	account1 := createRandomAccount(t)	
	err := testStore.DeleteAccount(context.Background(), account1.ID)
	require.NoError(t, err)

	account2, err := testStore.GetAccount(context.Background(), account1.ID)
	require.Error(t, err)
	require.EqualError(t, err, ErrRecordNotFound.Error())
	// require.EqualError(t, err, pgx.ErrNoRows.Error())
	require.Empty(t, account2)
}


func TestListAccounts(t *testing.T) {
	var lastAccount Account
	for i := 0; i < 20; i++ {
		lastAccount = createRandomAccount(t)
	}

	arg := ListAccountsParams {
		Owner: lastAccount.Owner,
		Limit: int32(utils.RandomInt(5, 10)),
		Offset: 0,
	}	

	accounts, err := testStore.ListAccounts(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, accounts)

	for _, account := range accounts {
		require.NotEmpty(t, account)
		require.Equal(t, lastAccount.Owner, account.Owner)
	}
}