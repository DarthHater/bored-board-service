package database

import (
	"fmt"
	"database/sql"

	"github.com/darthhater/bored-board-service/model"
	"github.com/spf13/viper"
	_ "github.com/lib/pq"
)

type IDatabase interface {
	InitDb(s string, e string) error
	GetThread(s string) (model.Thread, error)
	GetThreads(i int) ([]model.Thread, error)
	PostThread(t *model.NewThread) (string, error)
}

type Database struct {
}
var DB *sql.DB

// Public methods

func (d *Database) InitDb(environment string, configPath string) error {
	d.setupViper(environment, configPath)
	psqlInfo := d.connectionString(environment)
	err := d.openConnection(psqlInfo)
	if err != nil {
		return err
	}

	fmt.Println("Successfully connected")
	return nil
}

func (d *Database) GetThread(threadId string) (model.Thread, error) {
	thread := model.Thread{}
	err := DB.QueryRow("SELECT Id, UserId, Title, PostedAt FROM board.thread WHERE Id = $1", threadId).
		Scan(&thread.Id, &thread.UserId, &thread.Title, &thread.PostedAt)
	if err != nil {
		return thread, err
	}
	return thread, nil 
}

func (d *Database) GetThreads(num int) ([]model.Thread, error) {
	var threads []model.Thread
	rows, err := DB.Query("SELECT Id, UserId, Title, PostedAt FROM board.thread")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		t := model.Thread{}
        if err := rows.Scan(&t.Id, &t.UserId, &t.Title, &t.PostedAt); err != nil {
                return nil, err
		}
		threads = append(threads, t)
	}
	if rows.Err() != nil {
		panic(rows.Err())
	}

	return threads, nil 
}

func (d *Database) PostThread(newThread *model.NewThread) (threadid string, err error) {
	var id string
	sqlStatement := `
		INSERT INTO board.thread
		(UserId, Title, PostedAt)
		VALUES ($1, $2, $3)
		RETURNING Id`
	err = DB.QueryRow(sqlStatement, 
		newThread.T.UserId, 
		newThread.T.Title,
		newThread.T.PostedAt).Scan(&id)
	if err != nil {
		return "", err
	}

	sqlStatement = `
		INSERT INTO board.thread_post
		(ThreadId, UserId, Body, PostedAt)
		VALUES ($1, $2, $3, $4)`
	_, err = DB.Exec(sqlStatement,
		id,
		newThread.T.UserId,
		newThread.P.Body,
		newThread.P.PostedAt)
	if err != nil {
		return "", err
	}

	return id, nil
}

// Internal methods

func (d *Database) setupViper(environment string, configPath string) {
	viper.SetConfigName(environment)
	viper.AddConfigPath(configPath)

	err := viper.ReadInConfig()
	if err != nil {
		panic(err)
	}
}

func (d *Database) connectionString(env string) (connectionString string) {
	return fmt.Sprintf("postgres://%s:%s@database:%d/%s?sslmode=disable",
		viper.GetString("database.User"), 
		viper.GetString("database.Password"), 
		viper.GetInt("database.Port"), 
		viper.GetString("database.Database"))
}

func (d *Database) openConnection(psqlInfo string) error {
	var err error
	DB, err = sql.Open("postgres", psqlInfo)
	if err != nil {
		return err
	}
	err = DB.Ping()
	if err != nil {
		return err
	}
	return nil
}
