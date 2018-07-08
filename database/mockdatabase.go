package database

import (
	"github.com/DarthHater/bored-board-service/model"
	"github.com/DarthHater/bored-board-service/constants"
)

type MockDatabase struct {
}

func (m *MockDatabase) InitDb(s string, e string) error {
	return nil
}

func (m *MockDatabase) GetThread(threadId string) (model.Thread, error) {
	result := model.Thread{Id: "", UserId: "admin", Title: "What the heck", PostedAt: "A time"}

	return result, nil
}

func (m *MockDatabase) GetPosts(threadId string) ([]model.Post, error) {
	result := []model.Post{
		{Id: "", ThreadId: "", UserId: "", Body: "Post Body", PostedAt: "A time", EditedAt: "Rite now"},
		{Id: "", ThreadId: "", UserId: "", Body: "Post Body 2", PostedAt: "A time", EditedAt: "Rite now"},
	}

	return result, nil
}

func (m *MockDatabase) GetUser(username string) (result model.User, err error) {
	result = model.User{ID: "1", Username: "CoolGuy420", EmailAddress: "hsimpson@springfield.org", UserPassword: []byte("fake password"), UserRole: int(constants.Admin)}

	return result, nil
}
