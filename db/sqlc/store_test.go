package db

import (
	"context"
	"testing"

	"github.com/kjasn/simple-bank/utils"
	"github.com/stretchr/testify/require"
)

func TestTransferTx(t *testing.T) {
	store := NewStore(testDB)

	account1 := createRandomAccount(t)
	account2 := createRandomAccount(t)

	// create transfer transaction by concurrency
	cc := 5		// set the number of concurrency 
	errs := make(chan error)
	results := make(chan TransferTxResult)

	for i := 0; i < cc; i++ {
		go func() {
			arg := TransferTxParams {
				FromAccountID: account1.ID,
				ToAccountID: account2.ID,
				Amount: utils.RandomInt(0, account1.Balance),	// no more than its balance
			}
			res, err := store.TransferTx(context.Background(), arg)
			errs <- err
			results <- res
		}()
	}

	for i := 0; i < cc; i++ {
		require.NoError(t, <-errs)	
		require.NotEmpty(t, <-results)
	}

	// maybe don't to check transfers and entries

	// TODO: check accounts' balance
		
}