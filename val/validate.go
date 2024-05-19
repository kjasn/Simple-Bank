package val

import (
	"fmt"
	"net/mail"
	"regexp"
)

var (
	isValidUsername = regexp.MustCompile(`^[a-z_]+[a-z0-9_]{5,49}$`).MatchString
	isValidFullName= regexp.MustCompile(`^[a-zA-Z]+\s+[a-zA-Z]+$`).MatchString
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
		return fmt.Errorf("username must contains lower letters, digitals, underscores, and can not start with digital")
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