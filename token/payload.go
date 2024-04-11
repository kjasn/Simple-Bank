package token

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

// Error types returned by the VerifyToken func
var (
	ErrInvalidToken = errors.New("TOKEN IS INVALID")
	ErrExpiredToken = errors.New("TOKEN IS EXPIRED")
)

type Payload struct {
	ID uuid.UUID `json:"id"`	// make each payload unique
	Username string `json:"username"`
	IssuedAt time.Time `json:"issued_at"`
	ExpiredAt time.Time `json:"expired_at"`
}

// NewPayload creates a new token for the specific username and duration
func NewPayload(username string, duration time.Duration) (*Payload, error) {
	tokenID, err := uuid.NewRandom()
	if err != nil {
		return nil, err
	}

	return &Payload{
		ID: tokenID,
		Username: username,
		IssuedAt: time.Now(),
		ExpiredAt: time.Now().Add(duration),
	}, nil
}


// Valid checks if the token payload is valid or not
func (payload *Payload) Valid() error{
	if time.Now().After(payload.ExpiredAt) {
		return ErrExpiredToken
	}	

	return nil
}