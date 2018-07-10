package model

type Post struct {
	Id       string
	ThreadId string
	UserId   string
	Body     string
	PostedAt string
	UpdatedAt
	UserName string
}
