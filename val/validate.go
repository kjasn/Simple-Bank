package val

import (
	"fmt"
	"net/mail"
	"regexp"

	db "github.com/kjasn/simple-bank/db/sqlc"
)

var (
	isValidUsername = regexp.MustCompile(`^[a-zA-Z]+[a-zA-Z0-9_]{5,49}$`).MatchString
	isValidFullName= regexp.MustCompile(`^[a-zA-Z]+$`).MatchString
	// isValidFullName= regexp.MustCompile(`^[a-zA-Z]+\s+[a-zA-Z]+$`).MatchString
)

func ValidateString(value string, minLength int, maxLength int) error {
	n := len(value)
	if n < minLength || n > maxLength {
		return fmt.Errorf("illegal string length, it must contains %d-%d characters", minLength, maxLength)
	}

	return nil
}


func ValidateUsername(name string) error {
	// check length
	if err := ValidateString(name, 6, 50); err != nil {
		return err
	}

	// check characters - only allow  lower letters, digitals, underscores
	if !isValidUsername(name) {
		return fmt.Errorf("username can only contains lower letters, digitals, underscores, and can not start with digital and underscores")
	}
	return nil
}

func ValidUserRole(role string) error {
	if role != string(db.UserRoleDepositor) && role != string(db.UserRoleBanker) {
		return fmt.Errorf("user role is illegal")
	}
	return nil
}

func ValidateFullName(name string) error {
	// check length
	if err := ValidateString(name, 6, 50); err != nil {
		return err
	}

	// check characters - only allow  letters and spaces
	if !isValidFullName(name) {
		return fmt.Errorf("username must contains letters and spaces")
	}
	return nil
}
func ValidateEmail(email string) error {
	if _, err := mail.ParseAddress(email); err != nil {
		return err
	}
	return nil
}

func ValidatePassword(password string) error {
	return ValidateString(password, 6, 50)
}

func ValidateSecretCode(code string) error {
	return ValidateString(code, 32, 32)
}

func ValidateVerifyEmailId(id int64) error {
	if id <= 0 {
		return fmt.Errorf("verify email id must be a positive number, %d is a negative number", id)
	}
	return nil
}