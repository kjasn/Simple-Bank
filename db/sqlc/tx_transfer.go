package db

import (
	"context"
	"fmt"
)

type TransferTxParams struct {	// same as CreateTransferParams
	FromAccountID int64 `json:"from_account_id"`
	ToAccountID int64 `json:"to_account_id"`
	Amount int64 `json:"amount"`
}

type TransferTxResult struct { 
	Transfer Transfer `json:"transfer"`
	FromAccount Account `json:"from_account"`
	FromEntry Entry `json:"from_entry"`
	ToAccount Account `json:"to_account"`
	ToEntry Entry `json:"to_entry"`
}


var txKey = struct{}{}

// TransferTx create transfer, entries between 2 accounts, update accounts balance 
func (SQLStore *SQLStore) TransferTx(ctx context.Context, arg TransferTxParams) (TransferTxResult, error) {
	var result TransferTxResult

	err := SQLStore.execTx(ctx, func(q *Queries) error {
		var err error
		// use the Queries q within a transaction to access db 

		// create transfer
		result.Transfer, err = q.CreateTransfer(ctx, CreateTransferParams(arg))

		if err != nil {
			return err
		}

		// create fromAccount entry
		result.FromEntry, err = q.CreateEntry(ctx, CreateEntryParams{
			AccountID: arg.FromAccountID,
			Amount: -arg.Amount,
		})
		
		if err != nil {
			return err
		}

		// create toAccount entry
		result.ToEntry, err = q.CreateEntry(ctx, CreateEntryParams{
			AccountID: arg.ToAccountID,
			Amount: arg.Amount,
		})
		if err != nil {
			return err
		}


		account1, err := q.GetAccountForUpdate(ctx, arg.FromAccountID)
		if err != nil {
			return err
		}
		

		if account1.Balance < arg.Amount {
			return fmt.Errorf("insufficient balance")
		}

		// update accounts' balance
		if arg.FromAccountID < arg.ToAccountID {
			result.FromAccount, result.ToAccount, err = addMoney(ctx, q, 
				arg.FromAccountID, -arg.Amount, arg.ToAccountID, arg.Amount)

		} else {
			result.ToAccount, result.FromAccount, err = addMoney(ctx, q, 
				arg.ToAccountID, arg.Amount, arg.FromAccountID, -arg.Amount)
		}

		if err != nil {
			return err
		}

		return nil
	})

	return result, err
}


// subtract from account1 and add to account2
func addMoney(
	ctx context.Context,
	q *Queries, 
	accountID1,
	amount1,
	accountID2,
	amount2 int64,
) (account1, account2 Account, err error) {

	account1, err = q.AddAccountBalance(ctx, AddAccountBalanceParams{
		Amount: amount1,
		ID: accountID1,
	})

	if err != nil {
		return
	}

	account2, err = q.AddAccountBalance(ctx, AddAccountBalanceParams{
		Amount: amount2,
		ID: accountID2,
	})

	return	// anyway, return err with accounts
}