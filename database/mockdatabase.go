package database

import (
	"github.com/darthhater/bored-board-service/model"
)

type MockDatabase struct {
}

func (m *MockDatabase) InitDb(s string, e string) error {
	return nil
}

func (m *MockDatabase) GetThread(threadId string) (model.Thread, error) {
	result := model.Thread{ Id: "", UserId: "admin", Title: "What the heck", PostedAt: "A time" }

	return result, nil
}
