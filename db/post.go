package db

import (
	"time"

	"github.com/Nikita-Boyarskikh/DB/config"
	"github.com/jackc/pgx"

	"strconv"
	"strings"

	"fmt"

	"errors"

	"github.com/Nikita-Boyarskikh/DB/models"
	"github.com/jackc/pgx/pgtype"
	"github.com/mailru/easyjson/opt"
)

func CreatePostsInThread(tx *pgx.Tx, forum string, thread int32,
	posts models.Posts, parents []int64, parentsWithoutZeros []string) (models.Posts, error) {
	sql := `INSERT INTO posts(ID, authorID, forumID, message, parentID, threadID, parents) VALUES `

	var (
		args            []interface{}
		sqlPlaceholders []string
		err             error
		parentsMap      = make(map[int64][]int64)
	)

	if len(parentsWithoutZeros) > 0 {
		rows, err := tx.Query("SELECT ID, parents FROM posts WHERE ID IN (" +
			strings.Join(parentsWithoutZeros, ", ") + ")")
		if err != nil {
			return models.Posts{}, err
		}

		for rows.Next() {
			vals, err := rows.Values()
			if err != nil {
				return models.Posts{}, err
			}

			elems := make([]int64, len(vals[1].(*pgtype.Int8Array).Elements))
			for i, el := range vals[1].(*pgtype.Int8Array).Elements {
				if el.Status != pgtype.Present {
					return models.Posts{}, errors.New("wrong parents elements status")
				}

				elems[i] = el.Int
			}

			parentsMap[vals[0].(int64)] = elems
		}
	}

	rows, err := tx.Query("SELECT nextval('posts_id_seq') FROM generate_series(0, $1)", len(posts)-1)
	if err != nil {
		return models.Posts{}, err
	}

	var ids []int64
	for rows.Next() {
		vals, err := rows.Values()
		if err != nil {
			return models.Posts{}, err
		}
		ids = append(ids, vals[0].(int64))
	}

	for count, post := range posts {
		numargs := 7
		sqlPlaceholders = append(sqlPlaceholders, "($"+strings.Join([]string{
			strconv.Itoa(count*numargs + 1), strconv.Itoa(count*numargs + 2),
			strconv.Itoa(count*numargs + 3), strconv.Itoa(count*numargs + 4),
			strconv.Itoa(count*numargs + 5), strconv.Itoa(count*numargs + 6),
			strconv.Itoa(count*numargs + 7),
		}, ", $")+")")

		resParents := append(parentsMap[parents[count]], ids[count])
		args = append(args, ids[count], post.Author, forum, post.Message, post.Parent.V, thread, resParents)
	}

	if len(args) == 0 {
		return models.Posts{}, nil
	}

	rows, err = tx.Query(sql+strings.Join(sqlPlaceholders, ", ")+
		` RETURNING created AT TIME ZONE 'UTC'`, args...)
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
			ID:       opt.OInt64(ids[i]),
			Author:   posts[i].Author,
			Created:  opt.OString(values[0].(time.Time).Format(config.TimestampOutLayout)),
			Forum:    opt.OString(forum),
			IsEdited: opt.OBool(false),
			Message:  posts[i].Message,
			Parent:   posts[i].Parent,
			Thread:   opt.OInt32(thread),
		}

		result = append(result, post)
		i++
	}

	if err := NewPosts(tx, result); err != nil {
		return models.Posts{}, err
	}

	return result, nil
}

func UpdatePostMessage(tx *pgx.Tx, id int64, message string, edited bool) (models.Post, error) {
	var (
		post    models.Post
		created time.Time
		forum   string
		parent  int64
		thread  int32
	)

	if err := tx.QueryRow(`UPDATE posts SET message=$1, isEdited=$2 WHERE ID=$3
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

func GetPost(tx *pgx.Tx, id int64) (models.Post, error) {
	var (
		post    models.Post
		created time.Time
		forum   string
		edited  bool
		parent  int64
		thread  int32
	)

	if err := tx.QueryRow(`SELECT authorID, created AT TIME ZONE 'UTC', forumID, isEdited, message, parentID, threadID FROM posts
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

func GetPosts(tx *pgx.Tx, id int32, limit int, since int64, sort string, desc bool) (models.Posts, error) {
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
		if rows, err = tx.Query(sql, id, limit); err != nil {
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
			if rows, err = tx.Query(sql, id, id, since, limit); err != nil {
				return models.Posts{}, err
			}
		} else {
			var err error
			if rows, err = tx.Query(sql, id, id, limit); err != nil {
				return models.Posts{}, err
			}
		}
		break

	default: // flat
		var (
			sinceSql string
			count    = 2
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
			if rows, err = tx.Query(sql, id, since, limit); err != nil {
				return models.Posts{}, err
			}
		} else {
			var err error
			if rows, err = tx.Query(sql, id, limit); err != nil {
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

func CheckAllPostsInOneThread(tx *pgx.Tx, id int32, posts models.Posts) (bool, error) {
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
	rows, err := tx.Query(sql)
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
