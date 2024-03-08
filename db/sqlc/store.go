package db

import (
	"context"
	"database/sql"
	"fmt"
)

// provide all functions to execute db queries and transactions
type Store struct {
	*Queries
	db *sql.DB
}


func NewStore(db *sql.DB) *Store {
	return &Store{
		db: db,
		Queries: New(db),
	}
}

// execTx executes a function within a database transaction.
func (store *Store) execTx(ctx context.Context, fn func(*Queries) error) error {
	// begin transaction
	tx, err := store.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	// execute the operations
	if err = fn(New(tx)); err != nil {
		// rollback transaction
		if rbErr := tx.Rollback(); rbErr != nil {
			// rollback failed
			return fmt.Errorf("transaction err: %v, rollback err: %v", err, rbErr)
		}
		return err
	}
		
	// commit transaction 
	return tx.Commit()
}


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
func (Store *Store) TransferTx(ctx context.Context, arg TransferTxParams) (TransferTxResult, error) {
	var result TransferTxResult

	err := Store.execTx(ctx, func(q *Queries) error {
		var err error
		// use the Queries q within a transaction to access db 

		// create transfer
		txName := ctx.Value(txKey)
		fmt.Println(txName, "create transfer")
		result.Transfer, err = q.CreateTransfer(ctx, CreateTransferParams(arg))

		if err != nil {
			return err
		}

		// create fromAccount entry
		fmt.Println(txName, "create FromAccount entry")
		result.FromEntry, err = q.CreateEntry(ctx, CreateEntryParams{
			AccountID: arg.FromAccountID,
			Amount: -arg.Amount,
		})
		
		if err != nil {
			return err
		}

		// create toAccount entry
		fmt.Println(txName, "create ToAccount entry")
		result.ToEntry, err = q.CreateEntry(ctx, CreateEntryParams{
			AccountID: arg.ToAccountID,
			Amount: arg.Amount,
		})
		if err != nil {
			return err
		}

		// TODO: update accounts' balance
		// var wg sync.WaitGroup
		// wg.Add(1)
		// go func() error{
		// 	defer wg.Done()
			
		// 	var updatedAccount1, updatedAccount2 Account

		// 	// update accounts' balance
		// 	if updatedAccount1, err = q.UpdateAccount(ctx, UpdateAccountParams{
		// 		ID: arg.FromAccountID,
		// 		Balance: account1.Balance - arg.Amount,
		// 	}); err != nil {
		// 		return err
		// 	}
		// 	if updatedAccount2, err = q.UpdateAccount(ctx, UpdateAccountParams{
		// 		ID: arg.ToAccountID,
		// 		Balance: account2.Balance + arg.Amount,
		// 	}); err != nil {
		// 		return err
		// 	}

		// 	result.FromAccount = updatedAccount1
		// 	result.ToAccount = updatedAccount2
		// 	return nil
		// }()
		// wg.Wait()
		// return nil

		fmt.Println(txName, "get FromAccount")
		account1, err := q.GetAccountForUpdate(ctx, arg.FromAccountID)
		if err != nil {
			return err
		}

		fmt.Println(txName, "get ToAccount")
		account2, err := q.GetAccountForUpdate(ctx, arg.ToAccountID)
		if err != nil {
			return err
		}

		if account1.Balance < arg.Amount {
			return fmt.Errorf("insufficient balance")
		}

		// update accounts' balance
		fmt.Println(txName, "update FromAccount's balance")	
		result.FromAccount, err = q.UpdateAccount(context.Background(), UpdateAccountParams{
			ID: arg.FromAccountID,
			Balance: account1.Balance - arg.Amount,
		})
		if err != nil {
			return err 
		}

		fmt.Println(txName, "update ToAccount's balance")	
		result.ToAccount, err = q.UpdateAccount(context.Background(), UpdateAccountParams{
			ID: arg.ToAccountID,
			Balance: account2.Balance + arg.Amount,
		})
		if err != nil {
			return err
		}

		return nil
	})

	return result, err
}