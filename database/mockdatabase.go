package database

import (
	"github.com/DarthHater/bored-board-service/model"
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
