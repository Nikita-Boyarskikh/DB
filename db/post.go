package db

import (
	"time"

	"errors"

	//"github.com/jackc/pgx"
	//"github.com/jackc/pgx/pgtype"
	"github.com/Nikita-Boyarskikh/DB/config"
	"github.com/mailru/easyjson/opt"
)

//easyjson:json
type Post struct {
	ID       opt.Int64
	Author   string
	Created  opt.String
	Forum    opt.String
	IsEdited opt.Bool
	Message  string
	Parent   opt.Int64
	Thread   opt.Int32
}

//easyjson:json
type Posts []Post

//easyjson:json
type PostFull struct {
	Post   Post
	Author User
	Thread Thread
	Forum  Forum
}

func CreatePostsInThread(forum string, thread int32, posts Posts) (Posts, error) {
	//batch := conn.BeginBatch()
	//defer batch.Close()
	//
	//for _, post := range posts {
	//	log.Printf(`INSERT INTO posts(authorID, forumID, message, threadID) VALUES (%s, %s, %s, %d)`,
	//		post.Author, forum, post.Message, thread)
	//	batch.Queue(`INSERT INTO posts(authorID, forumID, message, threadID) VALUES ($1, $2, $3, $4)`,
	//		[]interface{}{post.Author, forum, post.Message, thread},
	//		[]pgtype.OID{pgtype.VarcharOID, pgtype.VarcharOID, pgtype.TextOID, pgtype.Int4OID},
	//		[]int16{pgx.BinaryFormatCode},
	//	)
	//}
	//
	//log.Println(batch.ExecResults())
	return posts, nil
}

func UpdatePostMessage(id int64, message string) (Post, error) {
	var (
		post    Post
		created time.Time
		forum   string
		parent  int64
		thread  int32
	)

	log.Printf(`UPDATE posts SET message=%s, isEdited=TRUE WHERE ID=%d`, message, id)
	if err := conn.QueryRow(`UPDATE posts SET message=$1, isEdited=TRUE WHERE ID=$2
			RETURNING authorID, created AT TIME ZONE 'UTC', forumID, parentID, threadID`, message, id).
		Scan(&post.Author, &created, &forum, &parent, &thread); err != nil {
		return Post{}, err
	}

	post.ID = opt.OInt64(id)
	post.Created = opt.OString(created.Format(config.TimestampOutLayout))
	post.Forum = opt.OString(forum)
	post.Message = message
	post.IsEdited = opt.OBool(true)
	post.Parent = opt.OInt64(parent)
	post.Thread = opt.OInt32(thread)

	return post, nil
}

func GetPost(id int64) (Post, error) {
	var (
		post    Post
		created time.Time
		forum   string
		edited  bool
		parent  int64
		thread  int32
	)

	log.Printf(`SELECT authorID, created AT TIME ZONE 'UTC', forumID, isEdited, parentID, threadID FROM posts
		WHERE id=%d`, id)
	if err := conn.QueryRow(`SELECT authorID, created AT TIME ZONE 'UTC', forumID, isEdited, message, parentID, threadID FROM posts
		WHERE id=$1`, id).
		Scan(&post.Author, &created, &forum, &edited, &post.Message, &parent, &thread); err != nil {
		return Post{}, err
	}

	post.ID = opt.OInt64(id)
	post.Created = opt.OString(created.Format(config.TimestampOutLayout))
	post.Forum = opt.OString(forum)
	post.IsEdited = opt.OBool(edited)
	post.Parent = opt.OInt64(parent)
	post.Thread = opt.OInt32(thread)

	return post, nil
}

func GetPosts(id int32, limit int, since time.Time, sort string, desc bool) (Posts, error) {
	switch sort {
	case "flat":
		break
	case "tree":
		break
	case "parent_tree":
		break
	default:
		return Posts{}, errors.New("unknown sort type")
	}
}
