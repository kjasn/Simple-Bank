package db

import (
	"context"
	"fmt"
)

// execTx executes a function within a database transaction.
func (store *SQLStore) execTx(ctx context.Context, fn func(*Queries) error) error {
	// begin transaction
	tx, err := store.connPool.Begin(ctx)
	if err != nil {
		return err
	}

	// execute the operations
	if err = fn(New(tx)); err != nil {
		// rollback transaction
		if rbErr := tx.Rollback(ctx); rbErr != nil {
			// rollback failed
			return fmt.Errorf("transaction err: %v, rollback err: %v", err, rbErr)
		}
		return err
	}
		
	// commit transaction 
	return tx.Commit(ctx)
}