package database

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

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
	GetUsers(s string) ([]model.User, error)
	GetThread(s string) (model.Thread, error)
	GetMessage(s string) (model.Message, error)
	GetMessages(i int, u string) ([]model.Message, error)
	GetMessagePosts(s string) ([]model.MessagePost, error)
	GetPost(s string) (model.Post, error)
	GetPosts(s string, i int, since string, userID string) ([]model.Post, error)
	GetThreads(i int, since string) ([]model.Thread, error)
	GetUserInfo(userID string) (model.UserInfo, error)
	PostThread(t *model.NewThread) (model.NewThread, error)
	PostPost(p *model.Post) (model.Post, error)
	PostMessage(t *model.NewMessage) (model.NewMessage, error)
	PostMessagePost(p *model.MessagePost) (model.MessagePost, error)
	DeleteThread(s string) error
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

// GetMessage will get a message with the given ID.
func (d *Database) GetMessage(messageID string) (model.Message, error) {
	message := model.Message{}
	err := DB.QueryRow(`SELECT bm.Id, bm.UserId, bm.Title, bm.PostedAt, bu.Username
			FROM board.message bm
			INNER JOIN board.user bu ON bm.UserId = bu.Id
			WHERE bm.Id = $1 AND bm.Deleted != true
			ORDER BY PostedAt DESC limit 20`, messageID).
		Scan(&message.Id, &message.UserId, &message.Title, &message.PostedAt, &message.UserName)
	if err != nil {
		return message, err
	}

	rows, err := DB.Query(`SELECT bmm.UserId, bu.Username
			FROM board.message_member bmm
			INNER JOIN board.user bu ON bmm.UserId = bu.Id
			WHERE bmm.MessageId = $1`, messageID)
	if err != nil {
		return message, err
	}
	defer rows.Close()

	for rows.Next() {
		mm := model.MessageMember{}
		if err := rows.Scan(&mm.UserId, &mm.UserName); err != nil {
			return message, err
		}
		message.Members = append(message.Members, mm)
	}
	if rows.Err() != nil {
		panic(rows.Err())
	}

	return message, nil
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

// GetUsers will return a list of users whose username matches a search string.
func (d *Database) GetUsers(search string) ([]model.User, error) {
	users := []model.User{}
	sqlStatement := `
		SELECT Id, Username
		FROM (SELECT Id, to_tsvector(Username) as lex, UserRole, Username
				FROM board.user) doc
		WHERE doc.lex @@ to_tsquery($1) AND UserRole != $2`

	rows, err := DB.Query(sqlStatement, search + ":*", constants.Banned)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		u := model.User{}
		if err := rows.Scan(&u.ID, &u.Username); err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	if rows.Err() != nil {
		panic(rows.Err())
	}

	return users, nil
}

// GetUserInfo retrieves metadata about a user.
func (d *Database) GetUserInfo(userID string) (userInfo model.UserInfo, err error) {
	userInfo = model.UserInfo{}

	sqlStatement := `
		SELECT COUNT(t), COUNT(tp), MAX(tp.postedat), u.Username
			FROM board.thread t
				INNER JOIN board.thread_post tp ON t.id = tp.threadid
				INNER JOIN board.user u on tp.userid = u.id
		WHERE u.id = $1
		GROUP BY u.Username`

	err = DB.QueryRow(sqlStatement, userID).
		Scan(&userInfo.TotalThreads, &userInfo.TotalPosts, &userInfo.LastPosted, &userInfo.Username)
	if err != nil {
		return userInfo, err
	}

	return userInfo, nil
}

// GetThreads retrieves a given number of threads.
func (d *Database) GetThreads(num int, since string) ([]model.Thread, error) {
	var threads []model.Thread

	i := d.stringToInt64(since)
	t := d.int64ToTime(i)

	rows, err := DB.Query(`SELECT bt.Id, bt.UserId, bt.Title, bt.PostedAt, bu.Username
		FROM board.thread bt
		INNER JOIN board.user bu ON bt.UserId = bu.Id
		WHERE Deleted != true AND PostedAt < $1
		ORDER BY PostedAt DESC LIMIT $2`, t, num)

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
func (d *Database) GetPosts(threadID string, num int, since string, userID string) ([]model.Post, error) {
	var posts []model.Post
	var i int64
	var t time.Time

	// if this is the first time a user is (re)visiting the thread, try to get their last location in it
	// otherwise, use the passed paramater
	if (since != "") {
		i = d.stringToInt64(since)
		t = d.int64ToTime(i)
	} else {
		i, _ = d.getLastViewedTime(userID, threadID)
		t = d.int64ToTime(i)
	}

	rows, err := DB.Query(`SELECT tp.Id, tp.ThreadId, tp.UserId, tp.Body, tp.PostedAt, bu.Username
			FROM board.thread_post tp
			INNER JOIN board.user bu ON tp.UserId = bu.Id
			WHERE ThreadId = $1 AND PostedAt < $2
			ORDER BY PostedAt DESC LIMIT $2`, threadID, t, num)

	if err != nil {
		return nil, err
	}

	if since != "" {
		d.setLastViewedTime(userID, threadID, since)
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

	if count == 0 {
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

// GetMessages retrieves a given number of messages.
func (d *Database) GetMessages(num int, userid string) ([]model.Message, error) {
	var messages []model.Message
	rows, err := DB.Query(`SELECT bm.Id, bm.UserId, bm.Title, bm.PostedAt, bu.Username
		FROM board.message_member bmm
		INNER JOIN board.message bm ON bmm.MessageId = bm.Id
		INNER JOIN board.user bu ON bm.UserId = bu.Id
		WHERE bmm.UserId = $1 AND bm.Deleted != true`, userid)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		m := model.Message{}
		if err := rows.Scan(&m.Id, &m.UserId, &m.Title, &m.PostedAt, &m.UserName); err != nil {
			return nil, err
		}
		messages = append(messages, m)
	}
	if rows.Err() != nil {
		panic(rows.Err())
	}

	return messages, nil
}

// GetMessagePosts will return all posts under a given thread.
func (d *Database) GetMessagePosts(messageID string) ([]model.MessagePost, error) {
	var messageposts []model.MessagePost
	rows, err := DB.Query(`SELECT mp.Id, mp.MessageId, mp.UserId, mp.Body, mp.PostedAt, bu.Username
			FROM board.message_post mp
			INNER JOIN board.user bu ON mp.UserId = bu.Id
			WHERE mp.MessageId = $1 ORDER BY mp.PostedAt`, messageID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		mp := model.MessagePost{}
		if err := rows.Scan(&mp.Id, &mp.MessageId, &mp.UserId, &mp.Body, &mp.PostedAt, &mp.UserName); err != nil {
			return nil, err
		}
		messageposts = append(messageposts, mp)
	}
	if rows.Err() != nil {
		panic(rows.Err())
	}

	return messageposts, nil
}

// PostMessage will create a new message "thread".
func (d *Database) PostMessage(newMessage *model.NewMessage) (message model.NewMessage, err error) {
	sqlStatement := `
		INSERT INTO board.message
		(UserId, Title)
		VALUES ($1, $2)
		RETURNING Id, UserId, Title, PostedAt, (SELECT Username FROM board.user WHERE Id = $1)`
	err = DB.QueryRow(sqlStatement,
		newMessage.T.UserId,
		newMessage.T.Title).
		Scan(&message.T.Id, &message.T.UserId, &message.T.Title, &message.T.PostedAt, &message.T.UserName)
	if err != nil {
		return message, err
	}

	for _, mm := range newMessage.M {
		sqlStatement = `
		INSERT INTO board.message_member
		(UserId, MessageId)
		VALUES ($1, $2)`
		_, err = DB.Exec(sqlStatement,
			mm.UserId,
			message.T.Id)
		if err != nil {
			return message, err
		}
	}

	sqlStatement = `
		INSERT INTO board.message_post
		(MessageId, UserId, Body)
		VALUES ($1, $2, $3)
		RETURNING Id, MessageId, UserId, Body, PostedAt, (SELECT Username FROM board.user WHERE Id = $2)`
	err = DB.QueryRow(sqlStatement,
		message.T.Id,
		newMessage.T.UserId,
		newMessage.P.Body).
		Scan(&message.P.Id, &message.P.MessageId, &message.P.UserId, &message.P.Body, &message.P.PostedAt, &message.P.UserName)
	if err != nil {
		return message, err
	}

	return message, nil
}

// PostMessagePost will create a new message_post in an existing message.
func (d *Database) PostMessagePost(message *model.MessagePost) (newMessage model.MessagePost, err error) {
	sqlStatement := `
		INSERT INTO board.message_post
		(MessageId, UserId, Body)
		VALUES ($1, $2, $3)
		RETURNING Id, MessageId, UserId, Body, PostedAt, (SELECT Username FROM board.user WHERE Id = $2)`
	err = DB.QueryRow(sqlStatement,
		message.MessageId,
		message.UserId,
		message.Body).
		Scan(&newMessage.Id, &newMessage.MessageId, &newMessage.UserId,
			&newMessage.Body, &newMessage.PostedAt, &newMessage.UserName)
	if err != nil {
		return newMessage, err
	}

	return newMessage, nil
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

func (d *Database) getLastViewedTime(userID string, threadID string) (lastViewed int64, err error) {
	sqlStatement := `
		SELECT LastViewedPostUnixTime
		FROM board.message_post
		WHERE UserId = $1 AND threadID = $2`

	err = DB.QueryRow(sqlStatement,
		userID,
		threadID).
		Scan(&lastViewed)

	if err != nil {
		return -1, err
	}

	return lastViewed, nil
}

func (d *Database) setLastViewedTime(userID string, threadID string, since string) error {
	sqlStatement := `
		UPDATE board.message_post
		SET LastViewedPostUnixTime = $1
		WHERE UserId = $2 AND threadID = $3`

	res, _ := DB.Exec(sqlStatement,
		userID,
		threadID,
		since)

	rows, err := res.RowsAffected()

	if rows > 0 {
		return nil
	}

	return err
}

func (d *Database) stringToInt64(since string) int64 {
	i, _ := strconv.ParseInt(since, 10, 64)
	return i;
}

func (d *Database) int64ToTime(i int64) time.Time {
	return time.Unix(0, i*int64(time.Millisecond))
}
