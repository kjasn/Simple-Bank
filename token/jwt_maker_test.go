package token

import (
	"testing"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/kjasn/simple-bank/utils"
	"github.com/stretchr/testify/require"
)


func TestJWTMaker(t *testing.T) {
	// first create a JWTMaker 
	maker, err := NewJWTMaker(utils.RandomString(minSecretKeySize))
	require.NoError(t, err)

	
	username := utils.RandomOwner()
	duration := time.Minute
	issuedAt := time.Now()
	expiredAt := issuedAt.Add(duration)

	token, payload, err := maker.CreateToken(username, duration)
	require.NoError(t, err)
	require.NotEmpty(t, token)
	require.NotEmpty(t, payload)

	payload, err = maker.VerifyToken(token)
	require.NoError(t, err)
	require.NotEmpty(t, payload)

	require.NotZero(t, payload.ID)
	require.Equal(t, username, payload.Username)
	require.WithinDuration(t, issuedAt, payload.IssuedAt, time.Second)
	require.WithinDuration(t, expiredAt, payload.ExpiredAt, time.Second)
}


func TestExpiredJWTToken(t *testing.T) {
	maker, err := NewJWTMaker(utils.RandomString(minSecretKeySize))
	require.NoError(t, err)

	// create a expired token
	token, payload, err := maker.CreateToken(utils.RandomOwner(), -time.Minute)
	require.NoError(t, err)
	require.NotEmpty(t, token)

	// verify token
	payload, err = maker.VerifyToken(token)
	require.Error(t, err)	// expected error
	require.EqualError(t, err, ErrExpiredToken.Error())

	require.Nil(t, payload)

}


func TestInvalidJWTTokenAlgNone(t *testing.T) {
	payload, err := NewPayload(utils.RandomOwner(), time.Minute)
	require.NoError(t, err)

	jwtToken := jwt.NewWithClaims(jwt.SigningMethodNone, payload)
	token, err := jwtToken.SignedString(jwt.UnsafeAllowNoneSignatureType)
	require.NoError(t, err)
	

	maker, err := NewJWTMaker(utils.RandomString(minSecretKeySize))
	require.NoError(t, err)

	// verify token
	payload, err = maker.VerifyToken(token)
	require.Error(t, err)	// expected error
	require.EqualError(t, err, ErrInvalidToken.Error())
	require.Nil(t, payload)
}