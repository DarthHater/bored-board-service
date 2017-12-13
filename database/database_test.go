package database

import (
	"github.com/darthhater/bored-board-service/model"
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/DATA-DOG/go-sqlmock.v1"
)

func TestInitDb(t *testing.T) {
	m := MockDatabase{}
	err := m.InitDb("development", "../.environment")
	assert.Equal(t, err, nil)
}

func TestConnectionString(t *testing.T) {
	d := Database{}
	d.setupViper("development", "../.environment")
	assert.Equal(t, 
		"postgres://admin:admin123@database:5432/db?sslmode=disable", 
		d.connectionString("development"),
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

	row := sqlmock.NewRows([]string{"id", "userId", "title", "postedat"}).
		AddRow("", "admin", "What the heck", "A time")

	mock.ExpectQuery("SELECT (.+) FROM board.thread WHERE (.+)").WillReturnRows(row)

	result, err := d.GetThread("a thread")

	expected := model.Thread{ Id: "", UserId: "admin", Title: "What the heck", PostedAt: "A time" }
	
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

	row := sqlmock.NewRows([]string{"id", "userId", "title", "postedat"}).
		AddRow("", "admin", "What the heck", "A time").
		AddRow("", "admin", "DJ Khaled", "A time")

	mock.ExpectQuery("SELECT (.+) FROM board.thread").WillReturnRows(row)

	result, err := d.GetThreads(20)

	expected := []model.Thread{
		{ Id: "", UserId: "admin", Title: "What the heck", PostedAt: "A time" },
		{ Id: "", UserId: "admin", Title: "DJ Khaled", PostedAt: "A time"},
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

	row := sqlmock.NewRows([]string{"id", "threadid", "userid", "body", "postedat"}).
		AddRow("", "", "", "Post Body", "A time").
		AddRow("", "", "", "Post Body 2", "A time")

	mock.ExpectQuery("SELECT (.+) FROM board.thread_post").WillReturnRows(row)

	result, err := d.GetPosts("A thread")

	expected := []model.Post{
		{ Id: "", ThreadId: "", UserId: "", Body: "Post Body", PostedAt: "A time" },
		{ Id: "", ThreadId: "", UserId: "", Body: "Post Body 2", PostedAt: "A time"},
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
			WillReturnResult(sqlmock.NewResult(1,1))
	
	if id, err := d.PostThread(&newThread); err != nil {
		t.Errorf("Error was not expected while inserting thread: %s", err)
	} else {
		t.Logf("Thread inserted with id: %s", id)
	}

	if err = mock.ExpectationsWereMet(); err != nil {
		t.Errorf("There were unfulfilled expectations: %s", err)
	}
}
