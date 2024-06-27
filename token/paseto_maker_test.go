package token

import (
	"testing"
	"time"

	db "github.com/kjasn/simple-bank/db/sqlc"
	"github.com/kjasn/simple-bank/utils"
	"github.com/stretchr/testify/require"
)


func TestPasetoMaker (t *testing.T) {
	maker, err := NewPasetoMaker(utils.RandomString(minSecretKeySize))
	require.NoError(t, err)

	username := utils.RandomOwner()
	role := db.UserRoleDepositor
	duration := time.Minute
	issuedAt := time.Now()
	expiredAt := issuedAt.Add(duration)

	token, payload, err := maker.CreateToken(username, role, duration)
	require.NoError(t, err)
	require.NotEmpty(t, token)
	require.NotEmpty(t, payload)

	payload, err = maker.VerifyToken(token)
	require.NoError(t, err)
	require.NotEmpty(t, payload)

	require.NotZero(t, payload.ID)
	require.Equal(t, username, payload.Username)
	require.Equal(t, role, payload.Role)
	require.WithinDuration(t, issuedAt, payload.IssuedAt, time.Second)
	require.WithinDuration(t, expiredAt, payload.ExpiredAt, time.Second)
}



func TestExpiredPasetoToken(t *testing.T) {
	maker, err := NewJWTMaker(utils.RandomString(minSecretKeySize))
	require.NoError(t, err)
	role := db.UserRoleDepositor

	// create a expired token
	token, payload, err := maker.CreateToken(utils.RandomOwner(), role, -time.Minute)
	require.NoError(t, err)
	require.NotEmpty(t, token)

	// verify token
	payload, err = maker.VerifyToken(token)
	require.Error(t, err)	// expected error
	require.EqualError(t, err, ErrExpiredToken.Error())

	require.Nil(t, payload)
}