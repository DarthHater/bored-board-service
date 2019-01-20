package model

import (
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID           string
	Username     string
	EmailAddress string
	Password     []byte
	ConfirmCode  string
	UserRole     int
	Active       bool
}

type Registration struct {
	Username     string
	EmailAddress string
	Password     string
}

func (u *User) HashPassword(password string) (err error) {
	u.Password, err = bcrypt.GenerateFromPassword([]byte(password), 8)
	if err != nil {
		return err
	}

	return nil
}
