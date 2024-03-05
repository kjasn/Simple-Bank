package db

import (
	"context"
	"testing"
	"time"

	"github.com/kjasn/simple-bank/utils"
	"github.com/stretchr/testify/require"
)

// make sure each unite test independent
func createRandomEntry(t *testing.T) Entry {
	account := createRandomAccount(t)
	arg := CreateEntryParams {
		AccountID: account.ID,
		Amount: utils.RandomMoney(),
	}

	entry, err := testQueries.CreateEntry(context.Background(), arg);

	// check the return with testify/require
	require.NoError(t, err)
	require.NotEmpty(t, entry)
	require.Equal(t, arg.Amount, entry.Amount)
	// assert not zero value of its type
	require.NotZero(t, entry.ID)
	require.NotZero(t, entry.CreatedAt)

	return entry
}

func TestCreateEntry(t *testing.T) {
	createRandomEntry(t)
}

func TestGetEntry(t *testing.T) {
	entry1 := createRandomEntry(t)
	entry2, err := testQueries.GetEntry(context.Background(), entry1.ID)

	require.NoError(t, err)
	require.NotEmpty(t, entry2)

	require.Equal(t, entry1.ID, entry2.ID)
	require.Equal(t, entry1.Amount, entry2.Amount)
	require.WithinDuration(t, entry1.CreatedAt, entry2.CreatedAt, time.Second)	// set the delta time 1s
}

func TestUpdateEntry(t *testing.T) {
	oldEntry := createRandomEntry(t)
	arg := UpdateEntryParams {
		ID: oldEntry.ID,
		Amount: utils.RandomMoney(),
	}

	newEntry, err := testQueries.UpdateEntry(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, newEntry)

	require.Equal(t, oldEntry.ID, newEntry.ID)
	require.Equal(t, arg.Amount, newEntry.Amount)
	require.WithinDuration(t, oldEntry.CreatedAt, newEntry.CreatedAt, time.Second)	// set the delta time 1s
}

// func TestDeleteEntry(t *testing.T) {
// 	entry1 := createRandomEntry(t)	
// 	err := testQueries.DeleteEntry(context.Background(), entry1.ID)
// 	require.NoError(t, err)

// 	entry2, err := testQueries.GetEntry(context.Background(), entry1.ID)
// 	require.Error(t, err)
// 	require.EqualError(t, err, sql.ErrNoRows.Error())
// 	require.Empty(t, entry2)
// }


func TestListEntries(t *testing.T) {
	for i := 0; i < 20; i++ {
		createRandomEntry(t)
	}

	arg := ListEntriesParams {
		Limit: int32(utils.RandomInt(5, 10)),
		Offset: int32(utils.RandomInt(5, 10)),
	}

	entries, err := testQueries.ListEntries(context.Background(), arg)
	require.NoError(t, err)
	require.Len(t, entries, int(arg.Limit))
	for i := 0; i < len(entries); i++ {
		require.NotEmpty(t, entries[i])
	}
}
