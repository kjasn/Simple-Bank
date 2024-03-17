// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.25.0

package db

import (
	"context"
)

type Querier interface {
	// NO NEED to get account then update
	AddAccountBalance(ctx context.Context, arg AddAccountBalanceParams) (Account, error)
	CreateAccount(ctx context.Context, arg CreateAccountParams) (Account, error)
	CreateEntry(ctx context.Context, arg CreateEntryParams) (Entry, error)
	CreateTransfer(ctx context.Context, arg CreateTransferParams) (Transfer, error)
	// maybe we don't want get return after delete
	DeleteAccount(ctx context.Context, id int64) error
	GetAccount(ctx context.Context, id int64) (Account, error)
	GetAccountForUpdate(ctx context.Context, id int64) (Account, error)
	GetEntry(ctx context.Context, id int64) (Entry, error)
	GetTransfer(ctx context.Context, id int64) (Transfer, error)
	ListAccounts(ctx context.Context, arg ListAccountsParams) ([]Account, error)
	ListEntries(ctx context.Context, arg ListEntriesParams) ([]Entry, error)
	ListTransfers(ctx context.Context, arg ListTransfersParams) ([]Transfer, error)
	// ONLY update the balance of an account by id, and we want to get return after update
	UpdateAccount(ctx context.Context, arg UpdateAccountParams) (Account, error)
	// ONLY update the amount of an entry by account_id
	UpdateEntry(ctx context.Context, arg UpdateEntryParams) (Entry, error)
	// ONLY update the amount of an entry by from_account_id AND to_account_id
	UpdateTransfer(ctx context.Context, arg UpdateTransferParams) (Transfer, error)
}

var _ Querier = (*Queries)(nil)
