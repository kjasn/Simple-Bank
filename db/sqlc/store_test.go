package db

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)



func TestTransferTx(t *testing.T) {
	store := NewStore(testDB)

	account1 := createRandomAccount(t)
	account2 := createRandomAccount(t)

	fmt.Println(">>start: ", account1.Balance, " ", account2.Balance)

	// create transfer transaction by concurrency
	cc := 5		// set the number of concurrent goroutines
	errs := make(chan error, cc)
	results := make(chan TransferTxResult, cc)

	var amount int64 = 5

	for i := 0; i < cc; i++ { 
		txName := fmt.Sprintf("tx %d", i + 1)
		
		go func() {
			arg := TransferTxParams {
				FromAccountID: account1.ID,
				ToAccountID: account2.ID,
				Amount: amount,
			}
			// note each transaction
			ctx := context.WithValue(context.Background(), txKey, txName)
			res, err := store.TransferTx(ctx, arg)
			errs <- err
			results <- res
		}()
	}

	// check result
	existed := make(map[int]bool)

	for i := 0; i < cc; i++ {
		require.NoError(t, <-errs)	

		res := <-results
		require.NotEmpty(t, res)

		// check transfer
		transfer := res.Transfer

		require.NotEmpty(t, transfer)
		require.Equal(t, account1.ID, transfer.FromAccountID)
		require.Equal(t, account2.ID, transfer.ToAccountID)
		require.NotZero(t, transfer.ID)
		require.NotZero(t, transfer.CreatedAt)

		_, err := store.GetTransfer(context.Background(), transfer.ID)
		require.NoError(t, err)


		// check fromEntry
		fromEntry := res.FromEntry

		require.NotEmpty(t, fromEntry)
		require.Equal(t, account1.ID, fromEntry.AccountID)
		require.Equal(t, -amount, fromEntry.Amount)
		require.NotZero(t, fromEntry.ID)
		require.NotZero(t, fromEntry.CreatedAt)


		// check toEntry
		toEntry := res.ToEntry

		require.NotEmpty(t, toEntry)
		require.Equal(t, account2.ID, toEntry.AccountID)
		require.Equal(t, amount, toEntry.Amount)
		require.NotZero(t, toEntry.ID)
		require.NotZero(t, toEntry.CreatedAt)

		// check account
		fromAccount := res.FromAccount
		toAccount := res.ToAccount

		require.NotEmpty(t, fromAccount)
		require.NotEmpty(t, toAccount)
		require.Equal(t, account1.ID, fromAccount.ID)
		require.Equal(t, account2.ID, toAccount.ID)


		// check balance
		diff1 := account1.Balance - res.FromAccount.Balance
		diff2 := res.ToAccount.Balance - account2.Balance

		require.Equal(t, diff1, diff2)
		require.True(t, diff1 > 0)
		require.True(t, diff1 % amount == 0)
		k := int(diff1 / amount)
		require.True(t, k >= 1 && k <= cc)
		require.NotContains(t, existed, k)
		existed[k] = true


		fmt.Printf("transaction %d: %d, %d\n", k, fromAccount.Balance, toAccount.Balance)
	}


	// check the final updated balance

	// updatedAccount1, err := testQueries.GetAccount(context.Background(), account1.ID)
	updatedAccount1, err := store.GetAccount(context.Background(), account1.ID)
	require.NoError(t, err)
	require.Equal(t, account1.Balance - int64(cc) * amount, updatedAccount1.Balance)

	// updatedAccount2, err := testQueries.GetAccount(context.Background(), account2.ID)
	updatedAccount2, err := store.GetAccount(context.Background(), account2.ID)
	require.NoError(t, err)
	require.Equal(t, account2.Balance + int64(cc) * amount, updatedAccount2.Balance)

	// logs
	fmt.Println(">>end: ", updatedAccount1.Balance, " ", updatedAccount2.Balance)
}