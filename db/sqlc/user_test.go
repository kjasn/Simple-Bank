package db

import (
	"context"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/kjasn/simple-bank/utils"
	"github.com/stretchr/testify/require"
)

// make sure each unit test independent
func createRandomUser(t *testing.T) User {
	hashedPassword, err := utils.HashPassword(utils.RandomString(6))
	require.NoError(t,err)


	arg := CreateUserParams {
		Username: utils.RandomOwner(),
		Role: UserRoleDepositor,
		HashedPassword: hashedPassword, 
		FullName: utils.RandomOwner(),
		Email: utils.RandomEmail(),
	}

	user, err := testStore.CreateUser(context.Background(), arg);

	require.NoError(t, err)
	require.NotEmpty(t, user)
	require.Equal(t, arg.Username, user.Username)
	require.Equal(t, arg.HashedPassword, user.HashedPassword)
	require.Equal(t, arg.FullName, user.FullName)
	require.Equal(t, arg.Email, user.Email)

	// first created, the field PasswordChangedAt should be default zero val
	require.True(t, user.PasswordChangedAt.IsZero())
	require.NotZero(t, user.CreatedAt)

	return user
}

func TestCreateUser(t *testing.T) {
	createRandomUser(t)
}

func TestGetUser(t *testing.T) {
	user1 := createRandomUser(t)
	user2, err := testStore.GetUser(context.Background(), user1.Username)

	require.NoError(t, err)
	require.NotEmpty(t, user2)

	require.Equal(t, user1.Username, user2.Username)
	require.Equal(t, user1.HashedPassword, user2.HashedPassword)
	require.Equal(t, user1.FullName, user2.FullName)
	require.Equal(t, user1.Email, user2.Email)

	require.WithinDuration(t, user1.CreatedAt, user2.CreatedAt, time.Second)
	require.WithinDuration(t, user1.PasswordChangedAt, user2.PasswordChangedAt, time.Second)
}


func TestUpdateUser(t *testing.T) {
	oldUser := createRandomUser(t)
	// update email
	newEmail := utils.RandomEmail()
	updatedUser, err := testStore.UpdateUser(context.Background(), UpdateUserParams{
		Email: pgtype.Text{
			String: newEmail,
			Valid: true,
		},
		Username: oldUser.Username,
	})

	require.NoError(t, err)
	require.NotEmpty(t, updatedUser)
	require.Equal(t, oldUser.Username, updatedUser.Username)
	require.Equal(t, oldUser.HashedPassword, updatedUser.HashedPassword)
	require.Equal(t, oldUser.FullName, updatedUser.FullName)
	require.Equal(t, newEmail, updatedUser.Email)


	// update full name
	newFullName := utils.RandomOwner()
	updatedUser, err = testStore.UpdateUser(context.Background(), UpdateUserParams{
		FullName: pgtype.Text{
			String: newFullName,
			Valid: true,
		},
		Username: oldUser.Username,
	})

	require.NoError(t, err)
	require.NotEmpty(t, updatedUser)
	require.Equal(t, oldUser.Username, updatedUser.Username)
	require.Equal(t, oldUser.HashedPassword, updatedUser.HashedPassword)
	require.Equal(t, newFullName, updatedUser.FullName)
	require.Equal(t, newEmail, updatedUser.Email)


	// update password
	newPassword := utils.RandomString(6)
	newHashedPassword, err := utils.HashPassword(newPassword)
	require.NoError(t, err)

	updatedUser, err = testStore.UpdateUser(context.Background(), UpdateUserParams{
		HashedPassword: pgtype.Text{
			String: newHashedPassword,
			Valid: true,
		},
		Username: oldUser.Username,
	})

	require.NoError(t, err)
	require.NotEmpty(t, updatedUser)
	require.Equal(t, oldUser.Username, updatedUser.Username)
	require.Equal(t, newHashedPassword, updatedUser.HashedPassword)
	require.Equal(t, newFullName, updatedUser.FullName)
	require.Equal(t, newEmail, updatedUser.Email)

	// update user role
	updatedUser, err = testStore.UpdateUser(context.Background(), UpdateUserParams{
		Role: NullUserRole {
			UserRole: UserRoleBanker,
			Valid: true,
		},
		Username: oldUser.Username,
	})

	require.NoError(t, err)
	require.NotEmpty(t, updatedUser)
	require.Equal(t, oldUser.Username, updatedUser.Username)
	require.Equal(t, UserRoleBanker, updatedUser.Role)
	require.Equal(t, newHashedPassword, updatedUser.HashedPassword)
	require.Equal(t, newFullName, updatedUser.FullName)
	require.Equal(t, newEmail, updatedUser.Email)
}