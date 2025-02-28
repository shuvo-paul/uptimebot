package model

import (
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID       int
	Name     string
	Email    string
	Password string
	Verified bool
}

func (u *User) HashPassword() error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u.Password = string(hashedPassword)
	return nil
}

func (u *User) VerifyPassword(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password))
	return err == nil
}

func (u *User) ValidatePassword() error {
	passwordLength := len(u.Password)
	if passwordLength < 6 || passwordLength > 12 {
		return fmt.Errorf("password must be between 6 and 12 characters")
	}

	var hasNumber, hasSymbol, hasUpper, hasLower bool

	for _, char := range u.Password {
		switch {
		case char >= '0' && char <= '9':
			hasNumber = true
		case (char >= '!' && char <= '/') || (char >= ':' && char <= '@'):
			hasSymbol = true
		case char >= 'A' && char <= 'Z':
			hasUpper = true
		case char >= 'a' && char <= 'z':
			hasLower = true
		}
	}

	if !hasNumber {
		return fmt.Errorf("password must contain at least one number")
	}

	if !hasSymbol {
		return fmt.Errorf("password must contain at least one symbol")
	}

	if !hasUpper {
		return fmt.Errorf("password must contain at least one uppercase letter")
	}

	if !hasLower {
		return fmt.Errorf("password must contain at least one lowercase letter")
	}

	return nil
}
