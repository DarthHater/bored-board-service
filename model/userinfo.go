package model

// UserInfo is for holding information about the user.
type UserInfo struct {
	Username		string 	`json:"username"`
	TotalPosts		string	`json:"totalPosts"`
	TotalThreads		string	`json:"totalThreads"`
	DateJoined		string	`json:"dateJoined"`
	LastPosted		string	`json:"lastPosted"`
}
