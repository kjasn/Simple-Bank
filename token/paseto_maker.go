package token

import (
	"fmt"
	"time"

	"github.com/aead/chacha20poly1305"
	"github.com/o1egl/paseto"
)

type PasetoMaker struct {
	paseto *paseto.V2
	symmetricKey []byte
}


// NewPasetoMaker create a 
func NewPasetoMaker(symmetricKey string) (Maker, error) {
	if len(symmetricKey) < chacha20poly1305.KeySize{
		return nil, fmt.Errorf("invalid key size: secret key must be at least %d characters", chacha20poly1305.KeySize)
	}

	return &PasetoMaker {
		paseto: paseto.NewV2(),
		symmetricKey: []byte(symmetricKey),
	}, nil
}



	// CreateToken creates a signed token for a specific username and duration. 
func (maker *PasetoMaker)CreateToken(username string, duration time.Duration) (string, error) {
	payload, err := NewPayload(username, duration)

	if err != nil {
		return "", err
	}

	return maker.paseto.Encrypt(maker.symmetricKey, payload, "I am a footer~")	// footer is nil
}
	// VerifyToken checks if the token is valid or not.
func(maker *PasetoMaker)VerifyToken(token string) (*Payload, error) {
	payload := &Payload{}

	err := maker.paseto.Decrypt(token, maker.symmetricKey, payload, nil)

	if err != nil {
		return nil, ErrInvalidToken
	}

	err = payload.Valid()
	if err != nil {
		return nil, err
	}

	return payload, nil

}