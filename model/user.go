package model

import (
	"database/sql"
	"golang.org/x/crypto/bcrypt"
	"crypto/md5"
)

type User struct {
	ID           string
	Username     string
	EmailAddress string
	Password []byte
	UserRole      int
	UserPasswordMd5	sql.NullString
}

type Registration struct {
	Username     string
	EmailAddress string
	Password string
}

func (u *User) HashPassword(password string) (err error) {
	u.Password, err = bcrypt.GenerateFromPassword([]byte(password), 8)
	if err != nil {
		return err
	}

	return nil
}

func (u *User) HashPasswordMd5(password string) (hash [16]byte) {
	data := []byte(password)
	return md5.Sum(data)
}
