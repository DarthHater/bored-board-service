package database

import (
	"testing"

	"github.com/DarthHater/bored-board-service/model"

	"github.com/stretchr/testify/assert"
	sqlmock "gopkg.in/DATA-DOG/go-sqlmock.v1"
)

func TestInitDb(t *testing.T) {
	m := MockDatabase{}
	err := m.InitDb("development", "../.environment")
	assert.Equal(t, err, nil)
}

func TestConnectionString(t *testing.T) {
	d := Database{}
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

func TestGetPost(t *testing.T) {
	d := Database{}
	var mock sqlmock.Sqlmock
	var err error
	DB, mock, err = sqlmock.New()
	if err != nil {
		t.Fatalf("An error %s occurred when opening stub database connection", err)
	}
	defer DB.Close()

	row := sqlmock.NewRows([]string{"id", "threadid", "userid", "body", "postedat"}).
		AddRow("", "", "", "Post Body", "A time")

	mock.ExpectQuery("SELECT (.+) FROM board.thread_post WHERE (.+)").WillReturnRows(row)

	result, err := d.GetPost("a thread")

	expected := model.Post{Id: "", ThreadId: "", UserId: "", Body: "Post Body", PostedAt: "A time"}

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

	result, err := d.GetThreads(20)

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

	mock.ExpectQuery("INSERT INTO board.thread").WithArgs(
		newThread.T.Title,
		newThread.T.UserId,
		newThread.T.PostedAt).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("1"))

	mock.ExpectExec("INSERT INTO board.thread_post").WithArgs(
		"1",
		newThread.T.UserId,
		newThread.P.Body,
		newThread.P.PostedAt).
		WillReturnResult(sqlmock.NewResult(1, 1))

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
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("1"))

	if id, err := d.PostPost(&post); err != nil {
		t.Errorf("Error was not expected while inserting post: %s", err)
	} else {
		t.Logf("Post inserted with id: %s", id)
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

	expected := model.User{ID: "1", Username: "CoolGuy420", EmailAddress: "hsimpson@springfield.org", UserPassword: []byte("fake password")}

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
		user.UserPassword).
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
