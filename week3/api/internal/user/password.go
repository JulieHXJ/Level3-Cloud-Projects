package user

import (
	"errors"
	"unicode"
	"unicode/utf8"
)

const (
	MinPasswordLength = 12
	MaxPasswordLength = 64
)

var (
	ErrPasswordTooShort = errors.New("password must be at least 12 characters")
	ErrPasswordTooLong  = errors.New("password must not exceed 64 characters")
	ErrPasswordUpper    = errors.New("password must contain at least one uppercase letter")
	ErrPasswordLower    = errors.New("password must contain at least one lowercase letter")
	ErrPasswordNumber   = errors.New("password must contain at least one number")
	ErrPasswordSymbol   = errors.New("password must contain at least one special character")
)

func ValidatePassword(password string) error {
	length := utf8.RuneCountInString(password)

	if length < MinPasswordLength {
		return ErrPasswordTooShort
	}

	if length > MaxPasswordLength {
		return ErrPasswordTooLong
	}

	var (
		hasUpper  bool
		hasLower  bool
		hasNumber bool
		hasSymbol bool
	)

	for _, r := range password {
		switch {
		case unicode.IsUpper(r):
			hasUpper = true
		case unicode.IsLower(r):
			hasLower = true
		case unicode.IsDigit(r):
			hasNumber = true
		case unicode.IsPunct(r) || unicode.IsSymbol(r):
			hasSymbol = true
		}
	}

	if !hasUpper {
		return ErrPasswordUpper
	}

	if !hasLower {
		return ErrPasswordLower
	}

	if !hasNumber {
		return ErrPasswordNumber
	}

	if !hasSymbol {
		return ErrPasswordSymbol
	}

	return nil
}
