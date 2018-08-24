package database

import (
	"os"
	"testing"

	"github.com/DarthHater/bored-board-service/constants"
	"github.com/DarthHater/bored-board-service/model"

	"github.com/stretchr/testify/assert"
	sqlmock "gopkg.in/DATA-DOG/go-sqlmock.v1"
)

func TestInitDb(t *testing.T) {
	m := MockDatabase{}
	err := m.InitDb("development", "../.environment")
	assert.Equal(t, err, nil)
}

func TestConnectionStringDevelopment(t *testing.T) {
	d := Database{}
	os.Setenv("ENVIRONMENT", "development")
	d.setupViper()
	assert.Equal(t,
		"postgres://admin:admin123@database:5432/db?sslmode=disable",
		d.connectionString(),
	)
}

func TestConnectionStringProduction(t *testing.T) {
	d := Database{}
	os.Setenv("ENVIRONMENT", "production")
	d.setupViper()
	assert.Equal(t,
		"postgres://admin:admin123@database:5432/db",
		d.connectionString(),
	)
}

func TestGetThread(t *testing.T) {
	d := Database{}
	var mock sqlmock.Sqlmock
	var err error
	DB, mock, err = sqlmock.New()
	if err != nil {
		t.Fatalf("An error %s occurred when opening stub database connection", err)
	}
	defer DB.Close()

	row := sqlmock.NewRows([]string{"id", "userId", "title", "postedat", "username"}).
		AddRow("", "admin", "What the heck", "A time", "admin")

	mock.ExpectQuery("SELECT (.+) FROM board.thread").WillReturnRows(row)

	result, err := d.GetThread("a thread")

	expected := model.Thread{Id: "", UserId: "admin", Title: "What the heck", PostedAt: "A time", UserName: "admin"}

	assert.Equal(t, result, expected)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expections: %s", err)
	}
}

func TestGetMessage(t *testing.T) {
	d := Database{}
	var mock sqlmock.Sqlmock
	var err error
	DB, mock, err = sqlmock.New()
	if err != nil {
		t.Fatalf("An error %s occurred when opening stub database connection", err)
	}
	defer DB.Close()

	row := sqlmock.NewRows([]string{"id", "userId", "title", "postedat", "username"}).
		AddRow("", "admin", "What the heck", "A time", "admin")

	mock.ExpectQuery("SELECT (.+) FROM board.message").WillReturnRows(row)

	result, err := d.GetMessage("a message")

	expected := model.Message{Id: "", UserId: "admin", Title: "What the heck", PostedAt: "A time", UserName: "admin"}

	assert.Equal(t, result, expected)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expections: %s", err)
	}
}

func TestGetPost(t *testing.T) {
	d := Database{}
	var mock sqlmock.Sqlmock
	var err error
	DB, mock, err = sqlmock.New()
	if err != nil {
		t.Fatalf("An error %s occurred when opening stub database connection", err)
	}
	defer DB.Close()

	row := sqlmock.NewRows([]string{"id", "threadid", "userid", "body", "postedat", "username"}).
		AddRow("", "", "", "Post Body", "A time", "admin")

	mock.ExpectQuery("SELECT (.+) FROM board.thread_post").WillReturnRows(row)

	result, err := d.GetPost("a thread")

	expected := model.Post{Id: "", ThreadId: "", UserId: "", Body: "Post Body", PostedAt: "A time", UserName: "admin"}

	assert.Equal(t, result, expected)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expections: %s", err)
	}
}

func TestGetThreads(t *testing.T) {
	d := Database{}
	var mock sqlmock.Sqlmock
	var err error
	DB, mock, err = sqlmock.New()
	if err != nil {
		t.Fatalf("An error %s occurred when opening stub database connection", err)
	}
	defer DB.Close()

	row := sqlmock.NewRows([]string{"id", "userId", "title", "postedat", "username"}).
		AddRow("", "admin", "What the heck", "A time", "admin").
		AddRow("", "admin", "DJ Khaled", "A time", "admin")

	mock.ExpectQuery("SELECT (.+) FROM board.thread").WillReturnRows(row)

	result, err := d.GetThreads(20, "")

	expected := []model.Thread{
		{Id: "", UserId: "admin", Title: "What the heck", PostedAt: "A time", UserName: "admin"},
		{Id: "", UserId: "admin", Title: "DJ Khaled", PostedAt: "A time", UserName: "admin"},
	}

	assert.Equal(t, result, expected)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expections: %s", err)
	}
}

func TestGetPosts(t *testing.T) {
	d := Database{}
	var mock sqlmock.Sqlmock
	var err error
	DB, mock, err = sqlmock.New()
	if err != nil {
		t.Fatalf("An error %s occurred when opening stub database connection", err)
	}
	defer DB.Close()

	row := sqlmock.NewRows([]string{"id", "threadid", "userid", "body", "postedat", "username"}).
		AddRow("", "", "", "Post Body", "A time", "admin").
		AddRow("", "", "", "Post Body 2", "A time", "admin")

	mock.ExpectQuery("SELECT (.+) FROM board.thread_post").WillReturnRows(row)

	result, err := d.GetPosts("A thread")

	expected := []model.Post{
		{Id: "", ThreadId: "", UserId: "", Body: "Post Body", PostedAt: "A time", UserName: "admin"},
		{Id: "", ThreadId: "", UserId: "", Body: "Post Body 2", PostedAt: "A time", UserName: "admin"},
	}

	assert.Equal(t, result, expected)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expections: %s", err)
	}
}

func TestGetMessages(t *testing.T) {
	d := Database{}
	var mock sqlmock.Sqlmock
	var err error
	DB, mock, err = sqlmock.New()
	if err != nil {
		t.Fatalf("An error %s occurred when opening stub database connection", err)
	}
	defer DB.Close()

	row := sqlmock.NewRows([]string{"id", "userId", "title", "postedat", "username"}).
		AddRow("", "admin", "What the heck", "A time", "admin").
		AddRow("", "admin", "DJ Khaled", "A time", "admin")

	mock.ExpectQuery("SELECT (.+) FROM board.message").WillReturnRows(row)

	result, err := d.GetMessages(20, "")

	expected := []model.Message{
		{Id: "", UserId: "admin", Title: "What the heck", PostedAt: "A time", UserName: "admin"},
		{Id: "", UserId: "admin", Title: "DJ Khaled", PostedAt: "A time", UserName: "admin"},
	}

	assert.Equal(t, result, expected)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expections: %s", err)
	}
}

func TestGetMessagePosts(t *testing.T) {
	d := Database{}
	var mock sqlmock.Sqlmock
	var err error
	DB, mock, err = sqlmock.New()
	if err != nil {
		t.Fatalf("An error %s occurred when opening stub database connection", err)
	}
	defer DB.Close()

	row := sqlmock.NewRows([]string{"id", "threadid", "userid", "body", "postedat", "username"}).
		AddRow("", "", "", "Post Body", "A time", "admin").
		AddRow("", "", "", "Post Body 2", "A time", "admin")

	mock.ExpectQuery("SELECT (.+) FROM board.message_post").WillReturnRows(row)

	result, err := d.GetMessagePosts("A thread")

	expected := []model.MessagePost{
		{Id: "", MessageId: "", UserId: "", Body: "Post Body", PostedAt: "A time", UserName: "admin"},
		{Id: "", MessageId: "", UserId: "", Body: "Post Body 2", PostedAt: "A time", UserName: "admin"},
	}

	assert.Equal(t, result, expected)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expections: %s", err)
	}
}

func TestPostMessage(t *testing.T) {
	d := Database{}
	var mock sqlmock.Sqlmock
	var err error
	var newMessage model.NewMessage
	DB, mock, err = sqlmock.New()
	if err != nil {
		t.Fatalf("An error %s occurred when opening stub database connection", err)
	}
	defer DB.Close()

	messageMock := sqlmock.NewRows([]string{"id", "userId", "title", "postedat", "username"}).AddRow("", "", "Ok", "", "andy")
	messagePostMock := sqlmock.NewRows([]string{"id", "messageid", "userid", "body", "postedat", "username"}).AddRow("", "", "", "I'm Posting", "datetime", "andy")

	mock.ExpectQuery("INSERT INTO board.message").WithArgs(
		newMessage.T.Title,
		newMessage.T.UserId).
		WillReturnRows(messageMock)

	mock.ExpectQuery("INSERT INTO board.message_post").WithArgs(
		newMessage.T.Id,
		newMessage.T.UserId,
		newMessage.P.Body).
		WillReturnRows(messagePostMock)

	if id, err := d.PostMessage(&newMessage); err != nil {
		t.Errorf("Error was not expected while inserting thread: %s", err)
	} else {
		t.Logf("Thread inserted with id: %s", id)
	}

	if err = mock.ExpectationsWereMet(); err != nil {
		t.Errorf("There were unfulfilled expectations: %s", err)
	}
}

func TestPostMessagePost(t *testing.T) {
	d := Database{}
	var mock sqlmock.Sqlmock
	var err error
	var message model.MessagePost
	DB, mock, err = sqlmock.New()
	if err != nil {
		t.Fatalf("An error %s occurred when opening stub database connection", err)
	}
	defer DB.Close()

	mock.ExpectQuery("INSERT INTO board.message_post").WithArgs(
		message.MessageId,
		message.UserId,
		message.Body).
		WillReturnRows(sqlmock.NewRows([]string{"id", "messageid", "userid", "body", "postedat", "username"}).AddRow("1", "3", "4", "I'm Posting", "datetime", "andy"))

	if message, err := d.PostMessagePost(&message); err != nil {
		t.Errorf("Error was not expected while inserting message: %s", err)
	} else {
		t.Logf("Message post inserted with id: %s", message.Id)
	}

	if err = mock.ExpectationsWereMet(); err != nil {
		t.Errorf("There were unfulfilled expectations: %s", err)
	}
}

func TestPostThread(t *testing.T) {
	d := Database{}
	var mock sqlmock.Sqlmock
	var err error
	var newThread model.NewThread
	DB, mock, err = sqlmock.New()
	if err != nil {
		t.Fatalf("An error %s occurred when opening stub database connection", err)
	}
	defer DB.Close()

	threadMock := sqlmock.NewRows([]string{"id", "userId", "title", "postedat", "username"}).AddRow("", "", "Ok", "", "andy")
	postMock := sqlmock.NewRows([]string{"id", "threadid", "userid", "body", "postedat", "username"}).AddRow("", "", "", "I'm Posting", "datetime", "andy")

	mock.ExpectQuery("INSERT INTO board.thread").WithArgs(
		newThread.T.Title,
		newThread.T.UserId).
		WillReturnRows(threadMock)

	mock.ExpectQuery("INSERT INTO board.thread_post").WithArgs(
		newThread.T.Id,
		newThread.T.UserId,
		newThread.P.Body).
		WillReturnRows(postMock)

	if id, err := d.PostThread(&newThread); err != nil {
		t.Errorf("Error was not expected while inserting thread: %s", err)
	} else {
		t.Logf("Thread inserted with id: %s", id)
	}

	if err = mock.ExpectationsWereMet(); err != nil {
		t.Errorf("There were unfulfilled expectations: %s", err)
	}
}

func TestPostPost(t *testing.T) {
	d := Database{}
	var mock sqlmock.Sqlmock
	var err error
	var post model.Post
	DB, mock, err = sqlmock.New()
	if err != nil {
		t.Fatalf("An error %s occurred when opening stub database connection", err)
	}
	defer DB.Close()

	mock.ExpectQuery("INSERT INTO board.thread_post").WithArgs(
		post.ThreadId,
		post.UserId,
		post.Body).
		WillReturnRows(sqlmock.NewRows([]string{"id", "threadid", "userid", "body", "postedat", "username"}).AddRow("1", "3", "4", "I'm Posting", "datetime", "andy"))

	if post, err := d.PostPost(&post); err != nil {
		t.Errorf("Error was not expected while inserting post: %s", err)
	} else {
		t.Logf("Post inserted with id: %s", post.Id)
	}

	if err = mock.ExpectationsWereMet(); err != nil {
		t.Errorf("There were unfulfilled expectations: %s", err)
	}
}

func TestEditPost(t *testing.T) {
	d := Database{}
	var mock sqlmock.Sqlmock
	var err error
	var body, userID string
	DB, mock, err = sqlmock.New()
	if err != nil {
		t.Fatalf("An error %s occurred when opening stub database connection", err)
	}
	defer DB.Close()

	mock.ExpectQuery("UPDATE board.thread_post").
		WillReturnRows(sqlmock.NewRows([]string{"id", "threadid", "userid", "body", "postedat", "username"}).
			AddRow("1", "2", "3", ":)", "datetime", "andy"))

	post, err := d.EditPost(userID, body)

	if err != nil {
		t.Errorf("Error was not expected while updating post: %s", err)
	} else {
		t.Log("Post updated")
	}

	expected := model.Post{Id: "1", ThreadId: "2", UserId: "3", Body: ":)", PostedAt: "datetime", UserName: "andy"}

	assert.Equal(t, post, expected)

	if err = mock.ExpectationsWereMet(); err != nil {
		t.Errorf("There were unfulfilled expectations: %s", err)
	}
}

func TestTooLateToEditPost(t *testing.T) {
	d := Database{}
	var mock sqlmock.Sqlmock
	var err error
	var body, userID string

	DB, mock, err = sqlmock.New()
	if err != nil {
		t.Fatalf("An error %s occurred when opening stub database connection", err)
	}
	defer DB.Close()

	mock.ExpectQuery("UPDATE board.thread_post").
		WillReturnRows(sqlmock.NewRows([]string{"id", "threadid", "userid", "body", "postedat", "username"}))

	if _, err := d.EditPost(userID, body); err != nil {
		if err == ErrEditPost {
			t.Log("Correct error returned")
		} else {
			t.Errorf("Unexpected error")
		}
	} else {
		t.Errorf("Expecting an error but there was none")
	}

	if err = mock.ExpectationsWereMet(); err != nil {
		t.Errorf("There were unfulfilled expectations: %s", err)
	}
}

func TestGetUser(t *testing.T) {
	d := Database{}
	var mock sqlmock.Sqlmock
	var err error
	DB, mock, err = sqlmock.New()
	if err != nil {
		t.Fatalf("An error %s occurred when opening stub database connection", err)
	}
	defer DB.Close()

	row := sqlmock.NewRows([]string{"id", "username", "emailaddress", "userpassword", "isadmin"}).
		AddRow("1", "CoolGuy420", "hsimpson@springfield.org", []byte("fake password"), false)

	mock.ExpectQuery("SELECT (.+) FROM board.user").WillReturnRows(row)

	result, err := d.GetUser("CoolGuy420")

	expected := model.User{ID: "1", Username: "CoolGuy420", EmailAddress: "hsimpson@springfield.org", Password: []byte("fake password")}

	assert.Equal(t, result, expected)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expections: %s", err)
	}
}

func TestCreateUser(t *testing.T) {
	d := Database{}
	var mock sqlmock.Sqlmock
	var err error
	var user model.User
	DB, mock, err = sqlmock.New()
	if err != nil {
		t.Fatalf("An error %s occurred when opening stub database connection", err)
	}
	defer DB.Close()

	mock.ExpectQuery("INSERT INTO board.user").WithArgs(
		user.Username,
		user.EmailAddress,
		user.Password,
		constants.User).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("1"))

	if id, err := d.CreateUser(&user); err != nil {
		t.Errorf("Error was not expected while inserting user: %s", err)
	} else {
		t.Logf("User inserted with id: %s", id)
	}

	if err = mock.ExpectationsWereMet(); err != nil {
		t.Errorf("There were unfulfilled expectations: %s", err)
	}
}
