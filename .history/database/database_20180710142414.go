package database

import (
	"database/sql"
	"fmt"
	"time"
	"errors"
	"log"
	"os"

	"github.com/DarthHater/bored-board-service/constants"
	"github.com/DarthHater/bored-board-service/model"
	_ "github.com/lib/pq"
	"github.com/spf13/viper"
)

type IDatabase interface {
	InitDb(s string, e string) error
	CreateUser(u *model.User) (string, error)
	GetUser(s string) (model.User, error)
	GetThread(s string) (model.Thread, error)
	GetPost(s string) (model.Post, error)
	GetPosts(s string) ([]model.Post, error)
	GetThreads(i int) ([]model.Thread, error)
	PostThread(t *model.NewThread) (string, error)
	PostPost(p *model.Post) (string, error)
	DeleteThread(s string) (error)
	EditPost(p *model.Post) (error)
	GetUserRole(s string) (constants.Role, error)
}

type Database struct {
}

var DB *sql.DB

// Public methods

func (d *Database) InitDb(environment string, configPath string) error {
	d.setupViper()
	psqlInfo := d.connectionString()
	err := d.openConnection(psqlInfo)
	if err != nil {
		log.Print(err)
		return err
	}

	fmt.Println("Successfully connected")
	return nil
}

func (d *Database) GetThread(threadId string) (model.Thread, error) {
	thread := model.Thread{}
	err := DB.QueryRow(`SELECT bt.Id, bt.UserId, bt.Title, bt.PostedAt, bu.Username
			FROM board.thread bt
			INNER JOIN board.user bu ON bt.UserId = bu.Id
			WHERE bt.Id = $1 && bt.Deleted != true
			ORDER BY PostedAt DESC limit 20`, threadId).
		Scan(&thread.Id, &thread.UserId, &thread.Title, &thread.PostedAt, &thread.UserName)
	if err != nil {
		return thread, err
	}
	return thread, nil
}

func (d *Database) GetPost(postId string) (post model.Post, err error) {
	post = model.Post{}
	err = DB.QueryRow("SELECT Id, ThreadId, UserId, Body, PostedAt FROM board.thread_post WHERE Id = $1 && Deleted != true", postId).
		Scan(&post.Id, &post.ThreadId, &post.UserId, &post.Body, &post.PostedAt)
	if err != nil {
		return post, err
	}
	return post, nil
}

func (d *Database) GetUser(username string) (user model.User, err error) {
	user = model.User{}
	err = DB.QueryRow("SELECT Id, Username, EmailAddress, UserPassword, UserRole FROM board.user WHERE Username = $1", username).
		Scan(&user.ID, &user.Username, &user.EmailAddress, &user.UserPassword, &user.UserRole)
	if err != nil {
		return user, err
	}

	return user, nil
}

func (d *Database) GetThreads(num int) ([]model.Thread, error) {
	var threads []model.Thread
	rows, err := DB.Query(`SELECT bt.Id, bt.UserId, bt.Title, bt.PostedAt, bu.Username
		FROM board.thread bt
		INNER JOIN board.user bu ON bt.UserId = bu.Id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		t := model.Thread{}
		if err := rows.Scan(&t.Id, &t.UserId, &t.Title, &t.PostedAt, &t.UserName); err != nil {
			return nil, err
		}
		threads = append(threads, t)
	}
	if rows.Err() != nil {
		panic(rows.Err())
	}

	return threads, nil
}

func (d *Database) GetPosts(threadId string) ([]model.Post, error) {
	var posts []model.Post
	rows, err := DB.Query(`SELECT tp.Id, tp.ThreadId, tp.UserId, tp.Body, tp.PostedAt, bu.Username
			FROM board.thread_post tp
			INNER JOIN board.user bu ON tp.UserId = bu.Id
			WHERE ThreadId = $1`, threadId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		p := model.Post{}
		if err := rows.Scan(&p.Id, &p.ThreadId, &p.UserId, &p.Body, &p.PostedAt, &p.UserName); err != nil {
			return nil, err
		}
		posts = append(posts, p)
	}
	if rows.Err() != nil {
		panic(rows.Err())
	}

	return posts, nil
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

func (d *Database) PostPost(post *model.Post) (postid string, err error) {
	var id string
	sqlStatement := `
		INSERT INTO board.thread_post
		(ThreadId, UserId, Body)
		VALUES ($1, $2, $3)
		RETURNING Id`
	err = DB.QueryRow(sqlStatement,
		post.ThreadId,
		post.UserId,
		post.Body).Scan(&id)
	if err != nil {
		return "", err
	}

	return id, nil
}

func (d *Database) DeleteThread(threadId string) (err error) {
	sqlStatement := `
		UPDATE board.thread
		SET Deleted = true
		WHERE Id = $1`
	res, err := DB.Exec(sqlStatement, threadId)
	if err != nil {
		panic(err)
	}
	count, err := res.RowsAffected()
	if err != nil {
		panic(err)
	}

	if (count > 0) {
		return
	}

	return errors.New("Couldn't find that thread")
}

func (d *Database) EditPost(editPost *model.Post) (err error) {
	tx, err := DB.Begin()

    if err != nil {
        return
	}
	defer func() {
        if err != nil {
            tx.Rollback()
            return
        }
        err = tx.Commit()
	}()

	sqlStatement := `
		UPDATE board.thread_post
		SET Body = $1
		WHERE Id = $3 && Deleted != true
		AND PostedAt::date + '10 minutes'::interval > now()`
	res, err := DB.Exec(sqlStatement, editPost.Body, time.Now().Local(), editPost.Id)
	if err != nil {
		panic(err)
	}
	count, err := res.RowsAffected()
	if err != nil {
		panic(err)
	}

	if (count > 0) {
		return
	}

	return errors.New("Posts can only be edited for 10 minutes")
}

func (d *Database) GetUserRole(userId string) (constants.Role, error) {
	var userRoleId int
	sqlStatement := `
		SELECT UserRole
		FROM board.user
		WHERE Id = $1`
	err := DB.QueryRow(sqlStatement, userId).Scan(&userRoleId)
	if err != nil {
		return -1, err
	}

	return constants.Role(userRoleId), nil
}

func (d *Database) CreateUser(user *model.User) (userid string, err error) {
	var id string
	sqlStatement := `
		INSERT INTO board.user
		(Username, EmailAddress, UserPassword)
		VALUES ($1, $2, $3)
		RETURNING Id`
	err = DB.QueryRow(sqlStatement,
		user.Username,
		user.EmailAddress,
		user.UserPassword).Scan(&id)
	if err != nil {
		return "", err
	}

	return id, nil

}
// Internal methods

func (d *Database) setupViper() {
	viper.SetEnvPrefix("BBS")
	viper.SetDefault("DATABASE", "database")
	viper.SetDefault("DATABASE_PORT", 5432)
	viper.SetDefault("DATABASE_USER", "admin")
	viper.SetDefault("DATABASE_PASSWORD", "admin123")
	viper.SetDefault("DATABASE_DATABASE", "db")
	viper.BindEnv("DATABASE")
	viper.BindEnv("DATABASE_PORT")
	viper.BindEnv("DATABASE_USER")
	viper.BindEnv("DATABASE_PASSWORD")
	viper.BindEnv("DATABASE_DATABASE")
}

func (d *Database) connectionString() (connectionString string) {
	var environment = os.Getenv("ENVIRONMENT")
	var string = fmt.Sprintf("postgres://%s:%s@%s:%d/%s",
		viper.GetString("DATABASE_USER"),
		viper.GetString("DATABASE_PASSWORD"),
		viper.GetString("DATABASE"),
		viper.GetInt("DATABASE_PORT"),
		viper.GetString("DATABASE_DATABASE"))
	if environment == "development" {
		return string + "?sslmode=disable"
	} else {
		return string
	}
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
