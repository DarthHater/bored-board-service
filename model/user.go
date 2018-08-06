package model

import (
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID           string
	Username     string
	EmailAddress string
	UserPassword []byte
	IsAdmin      bool
}

type Registration struct {
	Username     string
	EmailAddress string
	UserPassword string
}

func (u *User) HashPassword(password string) (err error) {
	u.UserPassword, err = bcrypt.GenerateFromPassword([]byte(password), 8)
	if err != nil {
		return err
	}

	return nil
}
