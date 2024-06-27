package token

import (
	"errors"
	"fmt"
	"time"

	"github.com/dgrijalva/jwt-go"
	db "github.com/kjasn/simple-bank/db/sqlc"
)


const minSecretKeySize = 32

type JWTMaker struct {
	secretKey string
}


func NewJWTMaker(secretKey string) (Maker, error) {
	if len(secretKey) < minSecretKeySize {
		return nil, fmt.Errorf("invalid key size: secret key must be at least %d characters", minSecretKeySize)
	}

	return &JWTMaker{secretKey}, nil
}



func (maker *JWTMaker) CreateToken(username string, role db.UserRole, duration time.Duration) (string, *Payload, error) {
	payload, err := NewPayload(username, role, duration)
	if err != nil {
		return "", payload, err	
	}

	jwtToken := jwt.NewWithClaims(jwt.SigningMethodHS256, payload)

	token, err := jwtToken.SignedString([]byte(maker.secretKey))
	return token, payload, err
}

func (maker *JWTMaker) VerifyToken(token string) (*Payload, error) {
	keyFunc := func(token *jwt.Token) (interface{}, error) {
		_, ok := token.Method.(*jwt.SigningMethodHMAC)
		if !ok {
			return nil, ErrInvalidToken
		}
		return []byte(maker.secretKey), nil
	}

	// check and parse token
	jwtToken, err := jwt.ParseWithClaims(token, &Payload{}, keyFunc) 
	if err != nil {
		// two cases of this error
		verr, ok := err.(*jwt.ValidationError)
		if ok && errors.Is(verr.Inner, ErrExpiredToken) {// token expired
			return nil, ErrExpiredToken
		}

		// else, invalid token
		return nil, ErrInvalidToken
	}

	payload, ok := jwtToken.Claims.(*Payload)
	if !ok {
		return nil, ErrInvalidToken
	}
	return payload, nil
}