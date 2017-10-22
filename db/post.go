package db

import (
	"time"

	"github.com/Nikita-Boyarskikh/DB/config"
	"github.com/jackc/pgx"

	"strconv"
	"strings"

	"fmt"

	"github.com/Nikita-Boyarskikh/DB/models"
	"github.com/mailru/easyjson/opt"
)

func CreatePostsInThread(forum string, thread int32, posts models.Posts) (models.Posts, error) {
	sql := `INSERT INTO posts(ID, authorID, forumID, message, parentID, threadID, parents) VALUES `

	var (
		args            []interface{}
		ID              int64
		sqlPlaceholders []string
		err             error
	)
	for count, post := range posts {
		if err := conn.QueryRow(`SELECT nextval('posts_id_seq')`).Scan(&ID); err != nil {
			return models.Posts{}, nil
		}

		numargs := 7
		sqlPlaceholders = append(sqlPlaceholders, "($"+strings.Join([]string{
			strconv.Itoa(count*numargs + 1), strconv.Itoa(count*numargs + 2),
			strconv.Itoa(count*numargs + 3), strconv.Itoa(count*numargs + 4),
			strconv.Itoa(count*numargs + 5), strconv.Itoa(count*numargs + 6),
		}, ", $")+
			", (SELECT parents FROM posts WHERE ID = $"+strconv.Itoa(count*numargs+5)+
			") || $"+strconv.Itoa(count*numargs+7)+")")

		args = append(args, ID, post.Author, forum, post.Message, post.Parent.V, thread, []int64{ID})
	}

	if len(args) == 0 {
		return models.Posts{}, nil
	}

	rows, err := conn.Query(sql+strings.Join(sqlPlaceholders, ", ")+
		` RETURNING ID, created AT TIME ZONE 'UTC'`, args...)
	if err != nil {
		return models.Posts{}, err
	}

	i := 0
	var result models.Posts
	for rows.Next() {
		values, err := rows.Values()
		if err != nil {
			return models.Posts{}, err
		}

		post := models.Post{
			ID:       opt.OInt64(values[0].(int64)),
			Author:   posts[i].Author,
			Created:  opt.OString(values[1].(time.Time).Format(config.TimestampOutLayout)),
			Forum:    opt.OString(forum),
			IsEdited: opt.OBool(false),
			Message:  posts[i].Message,
			Parent:   posts[i].Parent,
			Thread:   opt.OInt32(thread),
		}

		result = append(result, post)
		i++
	}

	if err := NewPosts(result); err != nil {
		return models.Posts{}, err
	}
	return result, nil
}

func UpdatePostMessage(id int64, message string, edited bool) (models.Post, error) {
	var (
		post    models.Post
		created time.Time
		forum   string
		parent  int64
		thread  int32
	)

	if err := conn.QueryRow(`UPDATE posts SET message=$1, isEdited=$2 WHERE ID=$3
			RETURNING authorID, created AT TIME ZONE 'UTC', forumID, parentID, threadID`, message, edited, id).
		Scan(&post.Author, &created, &forum, &parent, &thread); err != nil {
		return models.Post{}, err
	}

	post.ID = opt.OInt64(id)
	post.Created = opt.OString(created.Format(config.TimestampOutLayout))
	post.Forum = opt.OString(forum)
	post.Message = message
	post.IsEdited = opt.OBool(edited)
	post.Parent = opt.OInt64(parent)
	post.Thread = opt.OInt32(thread)

	return post, nil
}

func GetPost(id int64) (models.Post, error) {
	var (
		post    models.Post
		created time.Time
		forum   string
		edited  bool
		parent  int64
		thread  int32
	)

	if err := conn.QueryRow(`SELECT authorID, created AT TIME ZONE 'UTC', forumID, isEdited, message, parentID, threadID FROM posts
		WHERE id=$1`, id).
		Scan(&post.Author, &created, &forum, &edited, &post.Message, &parent, &thread); err != nil {
		return models.Post{}, err
	}

	post.ID = opt.OInt64(id)
	post.Created = opt.OString(created.Format(config.TimestampOutLayout))
	post.Forum = opt.OString(forum)
	post.IsEdited = opt.OBool(edited)
	post.Parent = opt.OInt64(parent)
	post.Thread = opt.OInt32(thread)

	return post, nil
}

func GetPosts(id int32, limit int, since int64, sort string, desc bool) (models.Posts, error) {
	var (
		rows  *pgx.Rows
		posts models.Posts
	)

	switch sort {
	case "tree":
		sqlPattern := `SELECT ID, authorID, created AT TIME ZONE 'UTC', forumID, isEdited, message, parentID FROM posts
		WHERE threadID=$1 %s ORDER BY parents %s`

		var (
			parents string
			sql     string
		)
		if since < 1 {
			parents = ""
		} else {
			if desc {
				parents = ` AND parents < (SELECT parents FROM posts WHERE ID=` + strconv.FormatInt(since, 10) + `)`
			} else {
				parents = ` AND parents > (SELECT parents FROM posts WHERE ID=` + strconv.FormatInt(since, 10) + `)`
			}
		}

		if desc {
			sql = fmt.Sprintf(sqlPattern, parents, ` DESC LIMIT $2`)
		} else {
			sql = fmt.Sprintf(sqlPattern, parents, ` LIMIT $2`)
		}

		var err error
		if rows, err = conn.Query(sql, id, limit); err != nil {
			return models.Posts{}, err
		}
		break

	case "parent_tree":
		args := make([]string, 5)
		args[0] = "$1"
		args[1] = "$2"

		var (
			sqlPattern = `SELECT ID, authorID, created AT TIME ZONE 'UTC', forumID, isEdited, message, parentID FROM posts
			WHERE threadID = %s AND parents[1] IN (
				SELECT ID FROM posts WHERE threadID = %s AND parentID = 0 %s ORDER BY ID %s
			) ORDER BY parents %s`
			count = 3
		)

		if desc {
			if since > 0 {
				args[2] = "AND parents[1] <"
			}
			args[3] = `DESC`
			args[4] = `DESC`
		} else {
			if since > 0 {
				args[2] = "AND parents[1] >"
			}
			args[3] = `ASC`
			args[4] = `ASC`
		}

		if since > 0 {
			args[2] += ` (SELECT parents[1] FROM posts WHERE ID = $` + strconv.Itoa(count) + `)`
			count++
		} else {
			args[2] = ""
		}

		if limit != -1 {
			args[3] += ` LIMIT $` + strconv.Itoa(count)
		}

		sql := fmt.Sprintf(sqlPattern, args[0], args[1], args[2], args[3], args[4])

		if since > 0 {
			var err error
			if rows, err = conn.Query(sql, id, id, since, limit); err != nil {
				return models.Posts{}, err
			}
		} else {
			var err error
			if rows, err = conn.Query(sql, id, id, limit); err != nil {
				return models.Posts{}, err
			}
		}
		break

	default: // flat
		var (
			sinceSql    string
			count       = 2
		)
		if since > 0 {
			if desc {
				sinceSql = " AND ID < $" + strconv.Itoa(count)
			} else {
				sinceSql = " AND ID > $" + strconv.Itoa(count)
			}
			count++
		}

		sql := `SELECT ID, authorID, created AT TIME ZONE 'UTC', forumID, isEdited, message, parentID FROM posts
		WHERE threadID=$1 ` + sinceSql + ` ORDER BY (created, ID)`

		if desc {
			sql += ` DESC LIMIT $` + strconv.Itoa(count)
		} else {
			sql += ` LIMIT $` + strconv.Itoa(count)
		}

		if since > 0 {
			var err error
			if rows, err = conn.Query(sql, id, since, limit); err != nil {
				return models.Posts{}, err
			}
		} else {
			var err error
			if rows, err = conn.Query(sql, id, limit); err != nil {
				return models.Posts{}, err
			}
		}
	}

	for rows.Next() {
		values, err := rows.Values()
		if err != nil {
			return models.Posts{}, err
		}

		posts = append(posts, models.Post{
			ID:       opt.OInt64(values[0].(int64)),
			Author:   values[1].(string),
			Created:  opt.OString(values[2].(time.Time).Format(config.TimestampOutLayout)),
			Forum:    opt.OString(values[3].(string)),
			IsEdited: opt.OBool(values[4].(bool)),
			Message:  values[5].(string),
			Parent:   opt.OInt64(values[6].(int64)),
			Thread:   opt.OInt32(id),
		})
	}

	return posts, nil
}

func CheckAllPostsInOneThread(id int32, posts models.Posts) (bool, error) {
	postParents := make(map[int64]bool)
	for _, post := range posts {
		if post.Parent.Defined && post.Parent.V != 0 {
			postParents[post.Parent.V] = true
		}
	}

	var parentIDs []string
	for id := range postParents {
		parentIDs = append(parentIDs, strconv.FormatInt(id, 10))
	}

	if len(parentIDs) == 0 {
		return true, nil
	}

	format := `SELECT threadID FROM posts WHERE ID IN (%s)`
	sql := fmt.Sprintf(format, strings.Join(parentIDs, ", "))
	rows, err := conn.Query(sql)
	if err != nil {
		return false, err
	}

	counter := 0
	for rows.Next() {
		v, err := rows.Values()
		if err != nil {
			return false, err
		}

		if v[0].(int32) != id {
			return false, nil
		}
		counter++
	}

	return counter > 0, nil
}
