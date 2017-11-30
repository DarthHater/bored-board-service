package database

type MockDatabase struct {
}

func (m *MockDatabase) InitDb(s string, e string) error {
	return nil
}

func (m *MockDatabase) GetThread(threadId string) (Thread, error) {
	result := Thread{ Id: "", UserId: "admin", Title: "What the heck", PostedAt: "A time" }

	return result, nil
}