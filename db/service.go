package db

import (
	"fmt"

	"sort"
	"strings"

	"github.com/Nikita-Boyarskikh/DB/models"
	"github.com/jackc/pgx"
)

var (
	status      = models.Status{}
	initialized = false
)

func GetStatus() (models.Status, error) {
	if !initialized {
		sql := `SELECT COUNT(nickname) FROM users`
		if err := conn.QueryRow(sql).Scan(&status.User); err != nil {
			return status, err
		}

		sql = `SELECT COUNT(ID) FROM forums`
		if err := conn.QueryRow(sql).Scan(&status.Forum); err != nil {
			return status, err
		}

		sql = `SELECT COUNT(ID) FROM threads`
		if err := conn.QueryRow(sql).Scan(&status.Thread); err != nil {
			return status, err
		}

		sql = `SELECT COUNT(ID) FROM posts`
		if err := conn.QueryRow(sql).Scan(&status.Post); err != nil {
			return status, err
		}
	}

	initialized = true
	return status, nil
}

func Clear() error {
	tx, err := conn.Begin()
	if err != nil {
		return err
	}

	if !initialized || status.User > 0 {
		truncate(tx, "users")
	}
	if !initialized || status.Forum > 0 {
		truncate(tx, "forums")
	}
	if !initialized || status.Thread > 0 {
		truncate(tx, "threads")
	}
	if !initialized || status.Post > 0 {
		truncate(tx, "posts")
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	status.User = 0
	status.Forum = 0
	status.Thread = 0
	status.Post = 0

	return nil
}

func NewUser() {
	status.User++
}

func NewForum() {
	status.Forum++
}

func NewThread(forumID string, userID string) error {
	if _, err := conn.Exec(`UPDATE forums SET threads = threads + 1 WHERE slug = $1`, forumID); err != nil {
		return err
	}

	if _, err := conn.Exec(`INSERT INTO forum_users (forumID, userID) VALUES ($1, $2) ON CONFLICT DO NOTHING`, forumID, userID); err != nil {
		return err
	}

	status.Thread++
	return nil
}

func NewPosts(posts models.Posts) error {
	postsCount := make(map[string]int)
	args := make([]interface{}, len(posts)*2)
	sqlArgs := make([]string, len(posts))
	sortedPosts := make(models.Posts, len(posts))
	copy(sortedPosts, posts)
	sort.Sort(sortedPosts)
	for i, post := range sortedPosts {
		postsCount[post.Forum.V]++
		args[2*i] = post.Forum.V
		args[2*i+1] = post.Author
		sqlArgs[i] = fmt.Sprintf(`($%d, $%d)`, 2*i+1, 2*i+2)
	}

	for slug, n := range postsCount {
		sql := `UPDATE forums SET posts = posts + $1 WHERE slug = $2`
		if _, err := conn.Exec(sql, n, slug); err != nil {
			return err
		}
	}

	sql := `INSERT INTO forum_users (forumID, userID) VALUES %s ON CONFLICT DO NOTHING`
	if _, err := conn.Exec(fmt.Sprintf(sql, strings.Join(sqlArgs, ", ")), args...); err != nil {
		return err
	}

	status.Post += int64(len(posts))
	return nil
}

func truncate(tx *pgx.Tx, table string) error {
	sql := fmt.Sprintf(`TRUNCATE TABLE %s`, table)
	if _, err := conn.Exec(sql); err != nil {
		if err := tx.Rollback(); err != nil {
			return err
		}
		return err
	}
	return nil
}
