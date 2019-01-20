package model

import (
	"crypto/md5"
	"database/sql"

	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID              string
	Username        string
	EmailAddress    string
	Password        []byte
	UserRole        int
	UserPasswordMd5 sql.NullString
	ConfirmCode     string
	Active          bool
}

type Registration struct {
	Username     string
	EmailAddress string
	Password     string
}

// HashPassword with return a bcrypt hash of a string.
func (u *User) HashPassword(password string) (err error) {
	u.Password, err = bcrypt.GenerateFromPassword([]byte(password), 8)
	if err != nil {
		return err
	}

	return nil
}

// HashPasswordMd5 with return an MD5 hash of a string.
func (u *User) HashPasswordMd5(password string) (hash [16]byte) {
	data := []byte(password)
	return md5.Sum(data)
}
