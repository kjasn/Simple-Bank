package token

import (
	"time"

	db "github.com/kjasn/simple-bank/db/sqlc"
)

// Maker is an interface for managing tokens.
type Maker interface {
	// CreateToken creates a signed token for a specific username and duration. 
	CreateToken(username string, role db.UserRole, duration time.Duration) (string, *Payload, error)
	// VerifyToken checks if the token is valid or not.
	VerifyToken(token string) (*Payload, error)
}