package model

type Message struct {
	Id       string
	UserId   string
	Title    string
	PostedAt string
	UserName string
	Members  []MessageMember
}
