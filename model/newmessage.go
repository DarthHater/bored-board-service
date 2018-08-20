package model

type NewMessage struct {
	T Message         `json:"Message"`
	M []MessageMember `json:"MessageMember"`
	P MessagePost     `json:"MessagePost"`
}
