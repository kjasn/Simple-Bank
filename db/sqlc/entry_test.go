package db

import (
	"context"
	"testing"
	"time"

	"github.com/kjasn/simple-bank/utils"
	"github.com/stretchr/testify/require"
)

// make sure each unit test independent
func createRandomEntry(t *testing.T, account Account) Entry {
	arg := CreateEntryParams{
		AccountID: account.ID,
		Amount:    utils.RandomMoney(),
	}

	entry, err := testStore.CreateEntry(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, entry)

	require.Equal(t, arg.AccountID, entry.AccountID)
	require.Equal(t, arg.Amount, entry.Amount)

	require.NotZero(t, entry.ID)
	require.NotZero(t, entry.CreatedAt)

	return entry
}

func TestCreateEntry(t *testing.T) {
	account := createRandomAccount(t)
	createRandomEntry(t, account)
}

func TestGetEntry(t *testing.T) {
	account := createRandomAccount(t)

	entry1 := createRandomEntry(t, account)
	entry2, err := testStore.GetEntry(context.Background(), entry1.ID)

	require.NoError(t, err)
	require.NotEmpty(t, entry2)

	require.Equal(t, entry1.ID, entry2.ID)
	require.Equal(t, entry1.Amount, entry2.Amount)
	require.WithinDuration(t, entry1.CreatedAt, entry2.CreatedAt, time.Second)	// set the delta time 1s
}

func TestUpdateEntry(t *testing.T) {
	account := createRandomAccount(t)

	oldEntry := createRandomEntry(t, account)
	arg := UpdateEntryParams {
		ID: oldEntry.ID,
		Amount: utils.RandomMoney(),
	}

	newEntry, err := testStore.UpdateEntry(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, newEntry)

	require.Equal(t, oldEntry.ID, newEntry.ID)
	require.Equal(t, arg.Amount, newEntry.Amount)
	require.WithinDuration(t, oldEntry.CreatedAt, newEntry.CreatedAt, time.Second)	// set the delta time 1s
}



func TestListEntries(t *testing.T) {
	account := createRandomAccount(t)
	for i := 0; i < 20; i++ {
		createRandomEntry(t, account)
	}

	arg := ListEntriesParams {
		AccountID: account.ID,
		Limit: int32(utils.RandomInt(5, 10)),
		Offset: int32(utils.RandomInt(5, 10)),
	}

	entries, err := testStore.ListEntries(context.Background(), arg)
	require.NoError(t, err)
	require.Len(t, entries, int(arg.Limit))
	for i := 0; i < len(entries); i++ {
		require.NotEmpty(t, entries[i])
	}
}
