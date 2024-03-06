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
	FromEntry Entry `json:"from_entry"`
	ToEntry Entry `json:"to_entry"`
}

// TransferTx create transfer, entries between 2 accounts, update accounts balance 
func (Store *Store) TransferTx(ctx context.Context, arg TransferTxParams) (TransferTxResult, error) {
	var result TransferTxResult

	err := Store.execTx(ctx, func(q *Queries) error {
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

		// TODO: update accounts' balance
		// ...
		return nil
	})

	return result, err
}