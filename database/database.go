package database

import (
	"database/sql"
	"fmt"
	"errors"
	"log"
	"os"

	"github.com/DarthHater/bored-board-service/constants"
	"github.com/DarthHater/bored-board-service/model"
	"github.com/DavidHuie/gomigrate"
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
	GetThreads(i int, userID string) ([]model.Thread, error)
	GetUserInfo(userID string) (model.UserInfo, error)
	PostThread(t *model.NewThread) (model.NewThread, error)
	PostPost(p *model.Post) (model.Post, error)
	DeleteThread(s string) (error)
	EditPost(i string, b string) (model.Post, error)
}

type Database struct {
}

var DB *sql.DB

// Public methods

// InitDb will initalize the database by setting and using environmental
func (d *Database) InitDb(environment string, configPath string) error {
	d.setupViper()
	psqlInfo := d.connectionString()
	err := d.openConnection(psqlInfo)
	if err != nil {
		log.Print(err)
		return err
	}

	fmt.Println("Successfully connected")

	migrator, _ := gomigrate.NewMigrator(DB, gomigrate.Postgres{}, "./migrations")
	err = migrator.Migrate()

	if err != nil {
		panic(err)
	}

	return nil
}

// GetThread will get a thread with the given ID.
func (d *Database) GetThread(threadID string) (model.Thread, error) {
	thread := model.Thread{}
	err := DB.QueryRow(`SELECT bt.Id, bt.UserId, bt.Title, bt.PostedAt, bu.Username
			FROM board.thread bt
			INNER JOIN board.user bu ON bt.UserId = bu.Id
			WHERE bt.Id = $1 AND bt.Deleted != true
			ORDER BY PostedAt DESC limit 20`, threadID).
		Scan(&thread.Id, &thread.UserId, &thread.Title, &thread.PostedAt, &thread.UserName)
	if err != nil {
		return thread, err
	}
	return thread, nil
}

// GetPost retrieves a single post.
func (d *Database) GetPost(postID string) (post model.Post, err error) {
	post = model.Post{}
	err = DB.QueryRow(`SELECT tp.Id, tp.ThreadId, tp.UserId, tp.Body, tp.PostedAt, bu.Username
		FROM board.thread_post tp
		INNER JOIN board.user bu ON tp.UserId = bu.Id
		WHERE Id = $1 AND Deleted != true`, postID).
		Scan(&post.Id, &post.ThreadId, &post.UserId, &post.Body, &post.PostedAt, &post.UserName)
	if err != nil {
		return post, err
	}
	return post, nil
}

// GetUser retrieves a given user.
func (d *Database) GetUser(username string) (user model.User, err error) {
	user = model.User{}
	err = DB.QueryRow("SELECT Id, Username, EmailAddress, UserPassword, UserRole FROM board.user WHERE Username = $1", username).
		Scan(&user.ID, &user.Username, &user.EmailAddress, &user.Password, &user.UserRole)
	if err != nil {
		return user, err
	}

	return user, nil
}

// GetUserInfo retrieves metadata about a user.
func (d *Database) GetUserInfo(userID string) (userInfo model.UserInfo, err error) {
	userInfo = model.UserInfo{}

	sqlStatement := `
		SELECT COUNT(t), COUNT(tp), MAX(tp.postedat)
			FROM board.thread t
				INNER JOIN board.thread_post tp ON t.id = tp.threadid
				INNER JOIN board.user u on tp.userid = u.id
		WHERE u.id = $1`

	err = DB.QueryRow(sqlStatement, userID).
		Scan(&userInfo.NumberThreads, &userInfo.NumberPosts, &userInfo.LastPosted)
	if err != nil {
		return userInfo, err
	}

	return userInfo, nil
}

// GetThreads retrieves a given number of threads.
func (d *Database) GetThreads(num int, userID string) ([]model.Thread, error) {
	var threads []model.Thread

	sqlStatement := `SELECT bt.Id, bt.UserId, bt.Title, bt.PostedAt, bu.Username
		FROM board.thread bt
		INNER JOIN board.user bu ON bt.UserId = bu.Id
		WHERE Deleted != true`

	if userID != "" {
		sqlStatement += " AND board.user = 1$"
	}

	rows, err := DB.Query(sqlStatement)

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

// GetPosts will return all posts under a given thread.
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

// PostThread creates a new thread.
func (d *Database) PostThread(newThread *model.NewThread) (thread model.NewThread, err error) {
	sqlStatement := `
		INSERT INTO board.thread
		(UserId, Title)
		VALUES ($1, $2)
		RETURNING Id, UserId, Title, PostedAt, (SELECT Username FROM board.user WHERE Id = $1)`
	err = DB.QueryRow(sqlStatement,
		newThread.T.UserId,
		newThread.T.Title).
		Scan(&thread.T.Id, &thread.T.UserId, &thread.T.Title, &thread.T.PostedAt, &thread.T.UserName)
	if err != nil {
		return thread, err
	}

	sqlStatement = `
		INSERT INTO board.thread_post
		(ThreadId, UserId, Body)
		VALUES ($1, $2, $3)
		RETURNING Id, ThreadId, UserId, Body, PostedAt, (SELECT Username FROM board.user WHERE Id = $2)`
	err = DB.QueryRow(sqlStatement,
		thread.T.Id,
		newThread.T.UserId,
		newThread.P.Body).
		Scan(&thread.P.Id, &thread.P.ThreadId, &thread.P.UserId, &thread.P.Body, &thread.P.PostedAt, &thread.P.UserName)
	if err != nil {
		return thread, err
	}

	return thread, nil
}

// PostPost will create a new post.
func (d *Database) PostPost(post *model.Post) (newPost model.Post, err error) {
	sqlStatement := `
		INSERT INTO board.thread_post
		(ThreadId, UserId, Body)
		VALUES ($1, $2, $3)
		RETURNING Id, ThreadId, UserId, Body, PostedAt, (SELECT Username FROM board.user WHERE Id = $2)`
	err = DB.QueryRow(sqlStatement,
		post.ThreadId,
		post.UserId,
		post.Body).
		Scan(&newPost.Id, &newPost.ThreadId, &newPost.UserId,
			 &newPost.Body, &newPost.PostedAt, &newPost.UserName)
	if err != nil {
		return newPost, err
	}

	return newPost, nil
}

// DeleteThread will do a soft delete on a thread and all of its corresponding posts.
func (d *Database) DeleteThread(threadID string) (err error) {
	sqlStatement := `
		UPDATE board.thread
		SET Deleted = true
		WHERE Id = $1`

	res, err := DB.Exec(sqlStatement, threadID)

	if err != nil {
		panic(err)
	}

	count, err := res.RowsAffected()

	if (count == 0) {
		return errors.New("Couldn't find that thread")
	}

	sqlStatement = `
		UPDATE board.thread_post
		SET Deleted = true
		WHERE ThreadId = $1`

	res, err = DB.Exec(sqlStatement, threadID)

	if err != nil {
		panic(err)
	}

	return
}

// EditPost allows a user to edit a post within 10 minutes of posting it.
func (d *Database) EditPost(id string, body string) (post model.Post, err error) {
	sqlStatement := `
		UPDATE board.thread_post
		SET Body = $1
		WHERE Id = $2 AND PostedAt + '10 minutes'::interval > localtimestamp
		RETURNING Id, ThreadId, UserId, Body, PostedAt, (SELECT Username FROM board.user WHERE Id = UserId)`
	err = DB.QueryRow(sqlStatement, body, id).
		Scan(&post.Id, &post.ThreadId, &post.UserId, &post.Body,
			&post.PostedAt, &post.UserName)

	if err != nil {
		if err == sql.ErrNoRows {
			return post, ErrEditPost
		}
		return post, err
	}

	return post, nil
}

// CreateUser creates a new user.
func (d *Database) CreateUser(user *model.User) (userid string, err error) {
	var id string
	sqlStatement := `
		INSERT INTO board.user
		(Username, EmailAddress, UserPassword, UserRole)
		VALUES ($1, $2, $3, $4)
		RETURNING Id`
	err = DB.QueryRow(sqlStatement,
		user.Username,
		user.EmailAddress,
		user.Password,
		constants.User).Scan(&id)
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
