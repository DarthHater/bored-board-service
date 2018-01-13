package model

type User struct {
	ID           string
	Username     string
	EmailAddress string
	UserPassword []byte
}
