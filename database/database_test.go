package database

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/DATA-DOG/go-sqlmock.v1"
)

func TestInitDb(t *testing.T) {
	m := MockDatabase{}
	err := m.InitDb("development", ".environment")
	assert.Equal(t, err, nil)
}

func TestConnectionString(t *testing.T) {
	d := Database{}
	assert.Equal(t, 
		"postgres://admin:admin123@db:5432/db?sslmode=disable", 
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

	expected := Thread{ Id: "", UserId: "admin", Title: "What the heck", PostedAt: "A time" }
	
	assert.Equal(t, result, expected)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expections: %s", err)
	}
}
