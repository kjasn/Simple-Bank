package db

import (
	"context"
	"testing"
	"time"

	"github.com/kjasn/simple-bank/utils"
	"github.com/stretchr/testify/require"
)


func createRandomTransfer(t *testing.T) Transfer {
	account1 := createRandomAccount(t)
	account2 := createRandomAccount(t)
	// generate a transfer From account1 to account2
	arg := CreateTransferParams{
		FromAccountID: account1.ID,
		ToAccountID: account2.ID,
		Amount: utils.RandomInt(1, account1.Balance),		// no more than account1's balance
	}

	transfer, err := testQueries.CreateTransfer(context.Background(), arg)

	require.NoError(t, err)
	require.NotEmpty(t, transfer)
	
	require.Equal(t, arg.FromAccountID, transfer.FromAccountID)
	require.Equal(t, arg.ToAccountID, transfer.ToAccountID)
	require.Equal(t, arg.Amount, transfer.Amount)
	require.NotZero(t, transfer.ID)
	require.NotZero(t, transfer.CreatedAt)
	return transfer
}

func TestCreateTransfer(t *testing.T) {
	createRandomTransfer(t)
}


func TestGetTransfer(t *testing.T) {
	transfer1 := createRandomTransfer(t)
	arg := GetTransferParams{
		FromAccountID: transfer1.FromAccountID,
		ToAccountID: transfer1.ToAccountID,
	}

	transfer2, err := testQueries.GetTransfer(context.Background(), arg)

	require.NoError(t, err)
	require.NotEmpty(t, transfer2)

	require.Equal(t, transfer1.ID, transfer2.ID)
	require.Equal(t, transfer1.FromAccountID, transfer2.FromAccountID)
	require.Equal(t, transfer1.ToAccountID, transfer2.ToAccountID)
	require.Equal(t, transfer1.Amount, transfer2.Amount)
	require.WithinDuration(t, transfer1.CreatedAt, transfer2.CreatedAt, time.Second)
}


func TestListTransfers(t *testing.T) {
	for i := 0; i < 10; i++ {
		createRandomTransfer(t)
	}

	arg := ListTransfersParams{
		Limit: 5,
		Offset: 5,
	}

	transfers, err := testQueries.ListTransfers(context.Background(), arg)

	require.NoError(t, err)
	require.Len(t, transfers, 5)

	for _, transfer := range transfers {
		require.NotEmpty(t, transfer)
	}
}


func TestUpdateTransfer(t *testing.T) {
	oldTransfer:= createRandomTransfer(t)

	arg := UpdateTransferParams{
		FromAccountID: oldTransfer.FromAccountID,
		ToAccountID: oldTransfer.ToAccountID,
		Amount: oldTransfer.Amount - 1,
	}

	newTransfer, err := testQueries.UpdateTransfer(context.Background(), arg)

	require.NoError(t, err)
	require.NotEmpty(t, newTransfer)

	require.Equal(t, oldTransfer.ID, newTransfer.ID)
	require.Equal(t, oldTransfer.FromAccountID, newTransfer.FromAccountID)
	require.Equal(t, oldTransfer.ToAccountID, newTransfer.ToAccountID)
	require.Equal(t, arg.Amount, newTransfer.Amount)
	require.WithinDuration(t, oldTransfer.CreatedAt, newTransfer.CreatedAt, time.Second)
}